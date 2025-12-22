package cm

import (
	"fmt"
	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/entity/model"
	model3 "agent/logic/collector/device/model"
	model2 "agent/logic/collector/rtdb/model"
	"agent/repo/cm"
	cmutils "agent/repo/cm/utils"
	"agent/utils"
	"sort"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/samber/lo"

	pb "trpcprotocol/agent"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	TemplateInfoSep string = "/"
)

var (
	work     *worker
	muWorker sync.Mutex
)

type worker struct {
	tboxDevices           []model.Device
	tboxDevicesMutex      sync.RWMutex
	devices               map[definition.DeviceGidType]model.Device
	deviceMutex           sync.RWMutex
	instanceTemplates     map[definition.DeviceGidType]*model.TemplateData
	instanceTemplateMutex sync.RWMutex
	templates             map[string]*model.TemplateData
	templateMutex         sync.RWMutex
	stdData               *model.StdData // 全局标准化，无需区分采集设备
	stdDataMutex          sync.RWMutex
	mozuIDs               map[definition.DeviceGidType]string
	mozuIDMutex           sync.RWMutex
	reader                cm.Reader
	lastReaderName        string
	cron                  *cron.Cron
	configVersion         map[string]*model.ConfigVersion // 当前生效版本
	configVersionMutex    sync.RWMutex
	targetVersion         map[string]*model.ConfigVersion // 初始化过程中的目标版本
	targetVersionMutex    sync.RWMutex
	idToGid               map[string]definition.DeviceGidType
	stdDevice             *model.StdDeviceData
	stdDeviceMutex        sync.RWMutex
	isReloading           bool // 标记是否正在重新加载配置
	reloadMutex           sync.Mutex
}

// StdDeviceTreeNode 标准设备树节点
type StdDeviceTreeNode struct {
	Device   *model.StdDevice
	Children []*StdDeviceTreeNode
}

func newWorker() *worker {
	return &worker{
		devices:           make(map[definition.DeviceGidType]model.Device),
		instanceTemplates: make(map[definition.DeviceGidType]*model.TemplateData),
		templates:         make(map[string]*model.TemplateData),
		mozuIDs:           make(map[definition.DeviceGidType]string),
		idToGid:           make(map[string]definition.DeviceGidType),
	}
}

// Worker 获得默认配置管理worker
func Worker() *worker {
	muWorker.Lock()
	defer muWorker.Unlock()

	if work == nil {
		work = newWorker()
	}
	return work
}

// ReInitWorker 重新初始化worker
func ReInitWorker() error {
	log.Info("start to ReInit Worker")
	work.StopCron()
	newWork := newWorker()
	err := newWork.Init(Worker().lastReaderName)
	if err != nil {
		work.StartCron()
		return err
	}

	muWorker.Lock()
	defer muWorker.Unlock()
	work.UnInit()
	work = newWork

	log.Warn("ReInit Worker done")
	return nil
}

// Init 设备、目标、标准点配置初始化
func (w *worker) Init(configReadName string) error {
	w.StartReload()
	defer w.FinishReload()

	w.lastReaderName = configReadName
	w.reader = cm.NewReader(configReadName)
	// 初始化目标版本
	err := w.InitVersion()
	if err != nil {
		log.Warnf("Init Version err:%s", err.Error())
	}
	err = w.InitCron(configReadName)
	if err != nil {
		// return err
		log.Warnf("Init Cron err:%s", err.Error())
	}
	// 版本完全更新标记
	versionUpdated := true
	err = w.InitStd()
	if err != nil {
		versionUpdated = false
		log.Warnf("Init std err:%s, try backup", err.Error())
		w.reader = cm.NewReader(cm.BackupModName)
		err = w.InitStd()
		if err != nil {
			log.Warnf("backup init std err:%s", err.Error())
		}
		w.reader = cm.NewReader(configReadName)
	}
	err = w.InitDevicesAndTemplates()
	if err != nil {
		versionUpdated = false
		log.Warnf("init devices and templates err: %v, try backup", err)
		err = w.BackupInitDevicesAndTemplates()
		if err != nil {
			return err
		}
	}
	if configReadName == cm.LocalFileConfigModName {
		err = w.SaveCurrentDevicesConfig(definition.CollectorDeviceTypeTBox)
		if err != nil {
			versionUpdated = false
			log.Warnf("save devices config fail: %s", err.Error())
		}
	}
	// 所有步骤成功，提交版本更新
	if versionUpdated {
		w.SetConfigVersion(w.GetTargetVersion())
	} else {
		log.Warnf("init config failed, abandon set version")
	}
	return nil
}

// UnInit 退出时清理
func (w *worker) UnInit() {
	if w.cron == nil {
		return
	}
	w.cron.Stop()
}

// StopCron 停止定时任务
func (w *worker) StopCron() {
	if w.cron == nil {
		return
	}
	w.cron.Stop()
}

func (w *worker) StartCron() {
	if w.cron == nil {
		return
	}
	w.cron.Start()
}

// InitCron 设置定时任务
func (w *worker) InitCron(configReadName string) error {
	// 只有tlink和taskserver模式支持查询版本号
	if configReadName != cm.TLinkModName && configReadName != cm.TaskServerModName {
		return nil
	}
	// 定时任务检测版本更新
	if w.cron == nil {
		w.cron = cron.New(cron.WithSeconds())
	}
	s := config.GetRB().GetCheckTime()
	spec := fmt.Sprintf("*/%s * * * * *", s)
	log.Info("check cron add :%v", spec)
	_, err := w.cron.AddFunc(
		spec, func() {
			w.CheckVersionAndUpdate()
		})
	if err != nil {
		log.Warnf("cron add error:%v", err)
	}
	w.cron.Start()
	return nil
}

// InitVersion 初始化目标版本
func (w *worker) InitVersion() error {
	var err error
	versionMap, err := w.reader.GetCmdbVersion()
	if err != nil {
		return err
	}
	w.SetTargetVersion(versionMap) // 存入临时版本
	return nil
}

// InitStd 设置标准层映射
func (w *worker) InitStd() error {
	if config.GetRB().IsStdCalEnable() {
		std, err := w.reader.GetStdData(w.GetTargetVersion())
		if err != nil {
			return err
		}
		// 更新内存
		w.SetStdData(std)
		// 获取标准设备列表
		stdDeviceMap := make(map[string]bool, len(std.StdPointsInfo))
		for _, d := range std.StdPointsInfo {
			stdDeviceMap[d.StdDevice] = true
		}
		// 标准层设备
		stdDevice, err := w.reader.GetStdDevice(stdDeviceMap)
		if err != nil {
			return err
		}
		// 标准测点绑定
		stdDevice.StdPoints = w.getStdPointMap()
		w.SetStdDeviceData(stdDevice)
	}
	return nil
}

func (w *worker) getStdPointMap() map[string]model3.StdInstancePointsInfo {
	stdPoints := w.GetStdData().StdPointsInfo
	pointMap := make(map[string]model3.StdInstancePointsInfo)
	for _, p := range stdPoints {
		ps, ok := pointMap[p.StdDevice]
		if !ok {
			ps = make(model3.StdInstancePointsInfo, 0)
		}
		ps = append(ps, p)
		pointMap[p.StdDevice] = ps
	}
	return pointMap
}

func (w *worker) InitDevicesAndTemplates() error {
	return w.initDevicesAndTemplatesCore(
		w.reader,
		true,
		true,
	)
}

// BackupInitDevicesAndTemplates 备份逻辑
func (w *worker) BackupInitDevicesAndTemplates() error {
	return w.initDevicesAndTemplatesCore(
		cm.NewReader(cm.BackupModName),
		false,
		false,
	)
}

// 统一逻辑：设备设置、模板获取、模板处理
func (w *worker) initDevicesAndTemplatesCore(reader cm.Reader, saveDeviceConfig bool, saveTemplatesToFile bool) error {
	// 设置采集层
	collectDevices, tboxDevices, deviceMap, err := reader.GetDevices()
	if err != nil {
		log.Warnf("get devices failed: %s", err.Error())
		return err
	}
	w.SetTboxDevice(tboxDevices)

	if saveDeviceConfig && w.lastReaderName != cm.LocalFileConfigModName {
		// 采集设备写本地文件
		if err := cmutils.SaveConfigMapToDirFileWithVersion(
			deviceMap, consts.DeviceTag, w.GetTargetVersion()); err != nil {
			log.Warnf("save device to file failed, %v", err)
		}
	}

	templateList, tpl2DeviceName := w.setDevicesAndPrepareTemplate(collectDevices)
	dtList := utils.RemoveDuplicates(templateList)

	rawTemplateMap, err := reader.GetTemplates(dtList)
	if err != nil {
		log.Warnf("GetTemplates failed:%s", err.Error())
		return err
	}
	for name, info := range rawTemplateMap {
		// 写本地文件,每个device文件夹下都写一份
		if saveTemplatesToFile {
			deviceNums := lo.Uniq(tpl2DeviceName[name])
			for _, device := range deviceNums {
				err := cmutils.SaveConfigMapToMultipleFile(
					consts.ProjectPath+"/"+device+"/"+consts.RelativeTemplateDir+"/",
					map[string]any{name: info},
				)
				if err != nil {
					log.Warnf("save templates to multiple file failed, %v", err)
				}
			}
		}
		// 对原始的模版数据进行解析,并更新内存
		m, err := cmutils.ParseCollectTemplateInfo(info)
		if err != nil {
			log.Errorf("parse template %s failed: %v", name, err)
			continue
		}
		w.SetTemplate(name, m)
	}

	w.processDeviceTemplates(collectDevices)
	return nil
}

func (w *worker) setDevicesAndPrepareTemplate(collectDevices []model.Device) ([]string, map[string][]string) {
	var templateList []string
	// 驱动模版对应的采集器
	tpl2DeviceName := make(map[string][]string)
	for i := range collectDevices {
		d := &collectDevices[i]
		w.SetDevice(d.Gid, collectDevices[i])
		w.SetGidById(d.ID, d.Gid)
		mozuStr := fmt.Sprintf("%v", d.MozuID)
		w.SetDeviceMozuID(d.Gid, mozuStr)

		name := d.TemplateData.TemplateName
		_, ok := tpl2DeviceName[name]
		if !ok {
			tpl2DeviceName[name] = []string{d.BelongCollectorDevice}
		} else {
			tpl2DeviceName[name] = append(tpl2DeviceName[name], d.BelongCollectorDevice)
		}
		templateList = append(templateList, name)
	}

	return templateList, tpl2DeviceName
}

// 配置驱动模版
func (w *worker) processDeviceTemplates(collectDevices []model.Device) {
	for i := range collectDevices {
		d := &collectDevices[i]
		t, ok := w.GetTemplateByName(d.TemplateData.TemplateName)
		if !ok {
			log.Errorf("template %s not found for device %v", d.TemplateData.TemplateName, d.Gid)
			continue
		}

		copied := t.Copy()
		deviceGid := string(d.Gid)

		// 将模版实例化到设备
		for i := range copied.PointsInfo {
			p := &copied.PointsInfo[i]
			p.ID = definition.GenerateDataPointID(deviceGid, definition.PointIDType(p.ID))
		}
		// 将模版实例化到子设备(非直采情况)
		sub2data := map[string]model.Device{}
		for _, sub := range d.SubDevices {
			sub2data[sub.ID] = sub
		}

		for i := range copied.SubDevices {
			subTemplate := &copied.SubDevices[i]
			if device, ok := sub2data[string(subTemplate.InstanceDeviceGid)]; ok {
				subTemplate.DeviceGiD = device.Gid
				subTemplate.InstanceDeviceGid = device.Gid // 南向plugin使用，后续统一
				for j := range subTemplate.PointsInfo {
					p := &subTemplate.PointsInfo[j]
					p.ID = definition.GenerateDataPointID(device.Gid, definition.PointIDType(p.ID))
				}
			}
		}
		w.SetDeviceTemplate(d.Gid, copied)
	}
}

// CopyAllDevices 复制所有设备
func (w *worker) CopyAllDevices() map[definition.DeviceGidType]*model.Device {
	data := make(map[definition.DeviceGidType]*model.Device, 0)
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()
	for k, v := range w.devices {
		data[k] = v.Copy()
	}
	return data
}

// GetAllDevices 获得所有设备
func (w *worker) GetAllDevices() []model.Device {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()
	devices := make([]model.Device, 0, len(w.devices))
	for _, d := range w.devices {
		devices = append(devices, d)
	}
	return devices
}

// GetDeviceByGid  通过采集设备gid查找设备
func (w *worker) GetDeviceByGid(gid definition.DeviceGidType) (model.Device, bool) {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()

	d, ok := w.devices[gid]
	return d, ok
}

// GetDeviceById  通过采集设备id查找设备
func (w *worker) GetDeviceById(id string) (model.Device, bool) {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()
	gid, ok := w.idToGid[id]
	if !ok {
		return model.Device{}, false
	}
	d, ok := w.devices[gid]
	return d, ok
}

// SetGidById 通过采集设备id设置采集设备gid
func (w *worker) SetGidById(id string, gid definition.DeviceGidType) {
	w.deviceMutex.Lock()
	defer w.deviceMutex.Unlock()

	w.idToGid[id] = gid
}

// GetDeviceGidById 通过采集设备id查找采集设备gid
func (w *worker) GetDeviceGidById(id string) (definition.DeviceGidType, bool) {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()

	gid, ok := w.idToGid[id]
	return gid, ok
}

// GetDeviceIdByGid 通过采集设备gid查找采集设备id
func (w *worker) GetDeviceIdByGid(gid definition.DeviceGidType) (string, bool) {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()

	d, ok := w.devices[gid]
	return d.ID, ok
}

// GetDevicesByGids 通过设备gid批量查找设备
func (w *worker) GetDevicesByGids(gids definition.DeviceGidArrType) ([]model.Device, bool, definition.DeviceGidArrType) {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()

	devices := make([]model.Device, 0, len(gids))
	notFoundIds := make(definition.DeviceGidArrType, 0)
	for _, id := range gids {
		d, ok := w.devices[id]
		if !ok {
			notFoundIds = append(notFoundIds, id)
		} else {
			devices = append(devices, d)
		}
	}
	if len(notFoundIds) != 0 {
		return nil, false, notFoundIds
	}
	return devices, true, nil
}

// GetDevicesByIds  通过设备id批量查找设备
func (w *worker) GetDevicesByIds(ids []string) ([]model.Device, bool, []string) {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()
	devices := make([]model.Device, 0, len(ids))
	notFoundIds := make([]string, 0)
	for _, id := range ids {
		gid, ok := w.idToGid[id]
		if !ok {
			notFoundIds = append(notFoundIds, id)
			continue
		}
		d, ok := w.devices[gid]
		if !ok {
			notFoundIds = append(notFoundIds, id)
		} else {
			devices = append(devices, d)
		}
	}
	if len(notFoundIds) != 0 {
		return nil, false, notFoundIds
	}
	return devices, true, nil
}

// SetDevice 设置设备
func (w *worker) SetDevice(gid definition.DeviceGidType, d model.Device) {
	w.deviceMutex.Lock()
	defer w.deviceMutex.Unlock()

	w.devices[gid] = d
}

// DeleteDevicesByGids 根据设备gid删除设备
func (w *worker) DeleteDevicesByGids(gids ...definition.DeviceGidType) {
	w.deviceMutex.Lock()
	defer w.deviceMutex.Unlock()
	log.Infof("DeleteDevicesByGids: %v", gids)
	for _, gid := range gids {
		delete(w.devices, gid)
	}
}

// DeleteDevicesByIds 根据设备id删除设备
func (w *worker) DeleteDevicesByIds(ids ...string) {
	w.deviceMutex.Lock()
	defer w.deviceMutex.Unlock()
	log.Infof("DeleteDevicesByIds: %v", ids)
	for _, id := range ids {
		gid, _ := w.idToGid[id]
		delete(w.devices, gid)
	}
}

// GetDeviceTemplateByGid 根据设备id查找设备模板
func (w *worker) GetDeviceTemplateByGid(gid definition.DeviceGidType) (*model.TemplateData, bool) {
	w.instanceTemplateMutex.RLock()
	defer w.instanceTemplateMutex.RUnlock()

	t, ok := w.instanceTemplates[gid]
	return t, ok
}

// CopyAllTemplateData 复制所有模板数据
func (w *worker) CopyAllTemplateData() map[string]*model.TemplateData {
	data := make(map[string]*model.TemplateData, 0)
	w.templateMutex.RLock()
	defer w.templateMutex.RUnlock()
	for k, v := range w.templates {
		data[k] = v.Copy()
	}
	return data
}

// SetDeviceTemplate 根据设备gid设置设备模板
func (w *worker) SetDeviceTemplate(gid definition.DeviceGidType, t *model.TemplateData) {
	w.instanceTemplateMutex.Lock()
	defer w.instanceTemplateMutex.Unlock()

	w.instanceTemplates[gid] = t
}

// GetTemplateByName 根据模板名称查找模板
func (w *worker) GetTemplateByName(name string) (*model.TemplateData, bool) {
	w.templateMutex.RLock()
	defer w.templateMutex.RUnlock()

	t, ok := w.templates[name]
	return t, ok
}

// DeleteTemplateByName 根据模板名称删除模板
func (w *worker) DeleteTemplateByName(names ...string) {
	w.templateMutex.RLock()
	defer w.templateMutex.RUnlock()
	log.Infof("DeleteTemplatesByNames: %v", names)
	for _, name := range names {
		delete(w.templates, name)
	}
}

// SetTemplate 根据模板名称设置模板
func (w *worker) SetTemplate(name string, t *model.TemplateData) {
	w.templateMutex.Lock()
	defer w.templateMutex.Unlock()

	w.templates[name] = t
}

// CopyStdData 复制标准测点数据
func (w *worker) CopyStdData() *model.StdData {
	w.stdDataMutex.RLock()
	defer w.stdDataMutex.RUnlock()
	return w.stdData.Copy()
}

// GetStdData 获取标准测点
func (w *worker) GetStdData() *model.StdData {
	w.stdDataMutex.RLock()
	defer w.stdDataMutex.RUnlock()

	return w.stdData
}

// GetStdDeviceData 获取标准设备
func (w *worker) GetStdDeviceData() *model.StdDeviceData {
	w.stdDeviceMutex.RLock()
	defer w.stdDeviceMutex.RUnlock()
	return w.stdDevice
}

// GetStdDeviceTree 获取标准设备树
func (w *worker) GetStdDeviceTree() []*StdDeviceTreeNode {
	stdDevicesData := w.GetStdDeviceData().Copy()
	if stdDevicesData == nil {
		return nil
	}
	stdDevices := stdDevicesData.StdDevices
	roots := []*StdDeviceTreeNode{}
	deviceMap := make(map[string]*StdDeviceTreeNode)
	// 1: 构建节点和map
	for i := range stdDevices {
		dev := stdDevices[i]
		node := &StdDeviceTreeNode{
			Device:   &dev,
			Children: make([]*StdDeviceTreeNode, 0),
		}
		deviceMap[dev.DeviceNumber] = node
	}
	// 2: 构建父子关系
	for _, dev := range stdDevices {
		node := deviceMap[dev.DeviceNumber]
		parentNumber := dev.ParentDeviceNumber

		if parentNumber == "" {
			roots = append(roots, node)
		} else {
			if parentNode, exists := deviceMap[parentNumber]; exists {
				parentNode.Children = append(parentNode.Children, node)
			} else {
				roots = append(roots, node)
			}
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Device.DeviceNumber < roots[j].Device.DeviceNumber
	})
	return roots
}

// SetStdDeviceData 设置标准设备
func (w *worker) SetStdDeviceData(stdDevice *model.StdDeviceData) {
	w.stdDeviceMutex.Lock()
	defer w.stdDeviceMutex.Unlock()
	w.stdDevice = stdDevice
}

// GetCollectData 获取采集测点
func (w *worker) GetCollectData() model2.DataPoints {
	collectDevices := w.GetAllDevices()
	if len(collectDevices) == 0 {
		return nil
	}

	allPoints := make(model2.DataPoints, 0, 1000)
	for _, device := range collectDevices {
		templateData, ok := w.GetDeviceTemplateByGid(device.Gid)
		if !ok {
			log.Errorf("not find device %v templatte", device.Gid)
			continue
		}

		pointsInfo := templateData.GetPoints()
		points := make(model2.DataPoints, 0, len(pointsInfo))
		for i := range pointsInfo {
			point := &pointsInfo[i]
			points = append(points, model2.DataPoint{
				ID:             point.ID,
				Rtd:            model2.NewRTData(),
				IsValueChanged: false,
			})
		}

		allPoints = append(allPoints, points...)
	}
	return allPoints
}

// SetStdData 设置标准数据
func (w *worker) SetStdData(std *model.StdData) {
	w.stdDataMutex.Lock()
	defer w.stdDataMutex.Unlock()

	w.stdData = std
}

// SetStdPointData 设置标准测点数据
func (w *worker) SetStdPointData(ps []model3.StdInstancePointInfo) {
	w.stdDataMutex.Lock()
	defer w.stdDataMutex.Unlock()

	w.stdData.StdPointsInfo = ps
}

// GetDeviceMozuID 获取模组id
func (w *worker) GetDeviceMozuID(deviceGiD definition.DeviceGidType) string {
	w.mozuIDMutex.RLock()
	defer w.mozuIDMutex.RUnlock()

	return w.mozuIDs[deviceGiD]
}

// SetDeviceMozuID 设置模组id
func (w *worker) SetDeviceMozuID(deviceGiD definition.DeviceGidType, mozu string) {
	w.mozuIDMutex.Lock()
	defer w.mozuIDMutex.Unlock()

	w.mozuIDs[deviceGiD] = mozu
}

// GetDeviceInfo 获取设备信息
func (w *worker) GetDeviceInfo(deviceGiD definition.DeviceGidType) (model.DeviceInfo, bool) {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()

	d, ok := w.devices[deviceGiD]
	return d.GetDeviceInfo(), ok
}

// WatchDeviceChanged 监控设备变化
func WatchDeviceChanged() <-chan bool {
	return cm.ConfigChangedChan()
}

// WatchStdConfigChangedChan 监控标准配置变化
func WatchStdConfigChangedChan() <-chan bool {
	return cm.StdConfigChangedChan()
}

// WatchDeviceConfigChangedChan 监控设备配置变化
func WatchDeviceConfigChangedChan() <-chan bool {
	return cm.DeviceConfigChangedChan()
}

// GetCachedTemplateData 获取缓存模板数据
func GetCachedTemplateData(deviceGid definition.DeviceGidType) model.TemplateData {
	t, ok := Worker().GetDeviceTemplateByGid(deviceGid)
	if !ok {
		return model.TemplateData{}
	}
	return *t
}

// GetTemplatesInfoNodes 获取模板信息节点
func (w *worker) GetTemplatesInfoNodes() []*pb.ProcessTemplatesRsp_RspGet_TemplateNode {
	templatesMap := w.CopyAllTemplateData()
	rootNode := &pb.ProcessTemplatesRsp_RspGet_TemplateNode{
		Children: make([]*pb.ProcessTemplatesRsp_RspGet_TemplateNode, 0),
	}
	var found bool
	var currentNode *pb.ProcessTemplatesRsp_RspGet_TemplateNode
	for k, v := range templatesMap {
		currentNode = rootNode
		found = false
		for _, node := range currentNode.Children {
			if node.Name == v.DrvInfo.Class {
				found = true
				currentNode = node
				break
			}
		}
		if !found {
			newNode := &pb.ProcessTemplatesRsp_RspGet_TemplateNode{
				Name:     v.DrvInfo.Class,
				Path:     "",
				IsDir:    1,
				Children: make([]*pb.ProcessTemplatesRsp_RspGet_TemplateNode, 0),
			}
			currentNode.Children = append(currentNode.Children, newNode)
			currentNode = newNode
		}
		found = false
		for _, node := range currentNode.Children {
			if node.Name == v.DrvInfo.Vendor {
				found = true
				currentNode = node
				break
			}
		}
		if !found {
			newNode := &pb.ProcessTemplatesRsp_RspGet_TemplateNode{
				Name:     v.DrvInfo.Vendor,
				Path:     v.DrvInfo.Class,
				IsDir:    1,
				Children: make([]*pb.ProcessTemplatesRsp_RspGet_TemplateNode, 0),
			}
			currentNode.Children = append(currentNode.Children, newNode)
			currentNode = newNode
		}
		found = false
		for _, node := range currentNode.Children {
			if node.Name == k {
				found = true
				break
			}
		}
		if !found {
			newNode := &pb.ProcessTemplatesRsp_RspGet_TemplateNode{
				Name:  k,
				Path:  v.DrvInfo.Class + "/" + v.DrvInfo.Vendor,
				IsDir: 0,
			}
			currentNode.Children = append(currentNode.Children, newNode)
		}
	}
	return rootNode.Children
}

// GetNextDeviceGid 获取下一个设备GID
func (w *worker) GetNextDeviceGid() definition.DeviceGidType {
	w.deviceMutex.Lock()
	defer w.deviceMutex.Unlock()

	var gid definition.DeviceGidType
	const maxTryCount = 10000
	var i int = 0
	timestamp := time.Now().Unix()
	gid = cmutils.GetAlternativeDeviceGid()
	if !gid.IsInt() {
		cmutils.SetAlternativeDeviceGid("0")
		gidStr := fmt.Sprintf("%v%v", timestamp, timestamp)
		gid = definition.DeviceGidType(gidStr)
		log.Errorf("generate device gid failed: not an int type, use <timestamp,timestamp> as gid: %v", gid)
		// return gid
	}
	for ; i < maxTryCount; i++ {
		if _, ok := w.devices[gid]; ok {
			gid, _ = gid.AddOne()
		} else {
			break
		}
	}
	if i == maxTryCount {
		gidStr := fmt.Sprintf("%v%v", timestamp, timestamp)
		gid = definition.DeviceGidType(gidStr)
		log.Errorf("generate device gid failed: exceed max try count, use <timestamp,timestamp> as gid: %v", gid)
	}
	retGid := gid
	gid, _ = gid.AddOne()
	cmutils.SetAlternativeDeviceGid(gid)
	return retGid
}

// GetTargetVersion 获得目标版本
func (w *worker) GetTargetVersion() map[string]*model.ConfigVersion {
	data := make(map[string]*model.ConfigVersion, 0)
	w.targetVersionMutex.RLock()
	defer w.targetVersionMutex.RUnlock()
	for k, v := range w.targetVersion {
		data[k] = v.Copy()
	}
	return data
}

// GetConfigVersion 获得当前版本
func (w *worker) GetConfigVersion() map[string]*model.ConfigVersion {
	data := make(map[string]*model.ConfigVersion, 0)
	w.configVersionMutex.RLock()
	defer w.configVersionMutex.RUnlock()
	for k, v := range w.configVersion {
		data[k] = v.Copy()
	}
	return data
}

// SetConfigVersion 设置当前生效版本
func (w *worker) SetConfigVersion(v map[string]*model.ConfigVersion) {
	w.configVersionMutex.Lock()
	defer w.configVersionMutex.Unlock()
	for k, c := range v {
		log.Warnf("set version:deviceNumber:%s,std version:%s,device version:%s", k, c.Point, c.Collector)
	}
	w.configVersion = v
}

// SetTargetVersion 设置临时目标版本
func (w *worker) SetTargetVersion(v map[string]*model.ConfigVersion) {
	w.targetVersionMutex.Lock()
	defer w.targetVersionMutex.Unlock()
	for k, c := range v {
		log.Warnf("target version:deviceNumber:%s,std version:%s,device version:%s", k, c.Point, c.Collector)
	}
	w.targetVersion = v
}

// CmdbVersion 临时mock
type CmdbVersion struct {
	Collector string
	Point     string
}

// VersionResult 对比结果
type VersionResult struct {
	Collector bool
	Point     bool
}

// CheckVersionAndUpdate 检测版本更新,只实现tlink模式
func (w *worker) CheckVersionAndUpdate() error {
	// 检查是否正在重新加载配置，如果是则跳过本次检查
	w.reloadMutex.Lock()
	if w.isReloading {
		w.reloadMutex.Unlock()
		log.Warn("Config reload in progress, skip version check")
		return nil
	}
	w.reloadMutex.Unlock()

	// 获取采集设备
	deviceNumbers := cmutils.GetTargetDevice()
	log.Debugf("check deviceNumber:%v", deviceNumbers)
	// 当前版本
	workerVersionMap := w.GetConfigVersion()
	// 对比结果
	result := map[string]VersionResult{}
	// 拉取CMDB版本对比
	cmdbVersionMap, err := w.reader.GetCmdbVersion()
	if err != nil {
		log.Errorf("get cmdb version failed:%s", err.Error())
		return err
	}
	collectorVersionChangedDevices := make(map[string][]string)
	pointVersionChangedDevices := make(map[string][]string)
	var collectorChange, stdChange bool
	for _, deviceNumber := range deviceNumbers {
		// cmdb
		cmdbVersion, ok := cmdbVersionMap[deviceNumber]
		if !ok {
			log.Warnf("get cmdb version failed, deviceNumber:%s", deviceNumber)
			continue
		}
		cmdbStdVersion := cmdbVersion.Point
		cmdbDeviceVersion := cmdbVersion.Collector

		// worker
		var workerVersion *model.ConfigVersion
		workerVersion, ok = workerVersionMap[deviceNumber]
		if !ok {
			log.Warnf("get work version empty, deviceNumber:%s", deviceNumber)
			workerVersion = &model.ConfigVersion{
				Collector: "",
				Point:     "",
			}
		}
		workDeviceVersion := workerVersion.Collector
		workStdVersion := workerVersion.Point
		log.Debugf("check version---device:%s---work device version:%s,cmdb device version:%s---work std version:%s,"+
			"cmdb std version:%s",
			deviceNumber, workDeviceVersion, cmdbDeviceVersion, workStdVersion, cmdbStdVersion)
		// 对比
		collectorMatch := cmdbDeviceVersion == workDeviceVersion
		if !collectorMatch {
			collectorVersionChangedDevices[deviceNumber] = []string{workDeviceVersion, cmdbDeviceVersion}
			collectorChange = true
		}
		pointMatch := cmdbStdVersion == workStdVersion
		if !pointMatch {
			pointVersionChangedDevices[deviceNumber] = []string{workStdVersion, cmdbStdVersion}
			stdChange = true
		}
		// 填充到结果中
		result[deviceNumber] = VersionResult{
			Collector: collectorMatch,
			Point:     pointMatch,
		}
	}

	if collectorChange {
		// 全量更新
		log.Warnf("devices %v collect devices version change, start init device and tpls...", collectorVersionChangedDevices)
		cm.ConfigChangedChan() <- true
		return nil
	}
	if stdChange {
		log.Warnf("devices %v std version change, start init std...", pointVersionChangedDevices)
		cm.StdConfigChangedChan() <- true
	}
	return nil
}

// FinishReload 完成重新加载后更新状态
func (w *worker) FinishReload() {
	w.reloadMutex.Lock()
	defer w.reloadMutex.Unlock()
	w.isReloading = false // 清除加载状态
}

// StartReload 进入加载后标记状态
func (w *worker) StartReload() {
	w.reloadMutex.Lock()
	defer w.reloadMutex.Unlock()
	w.isReloading = true // 标记开始加载
}

// GetStdDeviceGidById  通过标准设备英文标识查找设备
func (w *worker) GetStdDeviceGidById(id string) (definition.DeviceGidType, bool) {
	gid := w.stdDevice.GetPointsByConciseCode(id)
	if gid == "" {
		return "", false
	}
	return definition.DeviceGidType(gid), true
}

// GetStdDeviceByGid  通过标准设备gid查找设备
func (w *worker) GetStdDeviceByGid(gid definition.DeviceGidType) (model.StdDevice, bool) {
	std, ok := w.stdDevice.GetDeviceByGid(string(gid))
	return std, ok
}

// SetTboxDevice 设置Tbox设备
func (w *worker) SetTboxDevice(tboxDevices []model.Device) {
	w.tboxDevicesMutex.Lock()
	defer w.tboxDevicesMutex.Unlock()

	w.tboxDevices = tboxDevices
}

// GetTboxDeviceGids 获取Tbox设备
func (w *worker) GetTboxDeviceGids() definition.DeviceGidArrType {
	w.tboxDevicesMutex.RLock()
	defer w.tboxDevicesMutex.RUnlock()
	d := make(definition.DeviceGidArrType, len(w.tboxDevices))
	for i, v := range w.tboxDevices {
		d[i] = v.Gid
	}
	return d
}
