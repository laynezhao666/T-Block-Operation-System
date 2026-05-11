package cm

import (
	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/entity/model"
	model3 "agent/logic/collector/device/model"
	model2 "agent/logic/collector/rtdb/model"
	"agent/logic/tbox/hotstandby"
	"agent/repo/cm"
	cmutils "agent/repo/cm/utils"
	"agent/utils"
	"fmt"
	"sort"
	"sync"
	"time"

	pb "trpcprotocol/agent"

	"github.com/robfig/cron/v3"

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
	lastMozu              string // 最后设置的有效 mozu 值
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
	gid2CollectDev        map[definition.DeviceGidType]string // gid和对应的采集器编号
	gid2CollectDevMutex   sync.RWMutex
	std2CollectDev        map[definition.DeviceGidType]string // 标准层gid和对应的采集器编号
	std2CollectDevMutex   sync.RWMutex
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
func ReInitWorker(event *model.ConfigChangeEvent) error {
	log.Info("start to ReInit Worker")
	work.StopCron()
	newWork := newWorker()
	err := newWork.Init(Worker().lastReaderName, event)
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
func (w *worker) Init(configReadName string, changes *model.ConfigChangeEvent) error {
	w.StartReload()
	defer w.FinishReload()

	w.lastReaderName = configReadName
	w.reader = cm.NewReader(configReadName)
	// 初始化目标版本
	err := w.InitVersion()
	if err != nil {
		log.Warnf("Init Version err:%s", err.Error())
	}
	if changes == nil {
		// 首次加载，从本地获取已有的版本
		changes, err = w.GetChangesFromLocal()
		if err != nil {
			log.Errorf("GetChangesFromLocal failed: %s", err.Error())
		}
	}
	log.Warnf("Detected version changes, collector changes: %v , std changes: %v",
		changes.CollectorChanged, changes.StdChanged)
	err = w.InitCron(configReadName)
	if err != nil {
		// return err
		log.Warnf("Init Cron err:%s", err.Error())
	}
	// 版本完全更新标记
	versionUpdated := true
	err = w.InitStd(changes)
	if err != nil {
		versionUpdated = false
		log.Warnf("Init std err:%s, try backup", err.Error())
		w.reader = cm.NewReader(cm.BackupModName)
		err = w.InitStd(changes)
		if err != nil {
			log.Warnf("backup init std err:%s", err.Error())
		}
		w.reader = cm.NewReader(configReadName)
	}
	err = w.InitDevicesAndTemplates(changes)
	if err != nil {
		versionUpdated = false
		log.Warnf("init devices and templates err: %v, try backup", err)
		err = w.BackupInitDevicesAndTemplates(changes)
		if err != nil {
			return err
		}
	}
	if configReadName == cm.LocalFileConfigModName && !config.GetRB().IsLocalDebugEnable() {
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

// GetChangesFromLocal 用本地已存版本进行对比
func (w *worker) GetChangesFromLocal() (*model.ConfigChangeEvent, error) {
	// 对比生成变更列表
	changes := &model.ConfigChangeEvent{
		CollectorChanged: make([]string, 0),
		StdChanged:       make([]string, 0),
	}
	localVersionMap, err := cm.NewReader(cm.BackupModName).GetCmdbVersion()
	// 对比获取发生了变化的信息
	if err != nil {
		log.Warnf("Failed to get backup versions: %v, performing full initialization", err)
	} else {
		// 从CMDB获取目标版本
		targetVersionMap := w.GetTargetVersion()

		for deviceNumber, targetVer := range targetVersionMap {
			backupVer, exists := localVersionMap[deviceNumber]
			if !exists {
				log.Warnf("Device %s not found in backup, will initialize", deviceNumber)
				changes.CollectorChanged = append(changes.CollectorChanged, deviceNumber)
				changes.StdChanged = append(changes.StdChanged, deviceNumber)
				continue
			}

			// 检查采集配置版本变更
			if targetVer.Collector != backupVer.Collector {
				log.Warnf("Collector version changed for %s: %s -> %s",
					deviceNumber, backupVer.Collector, targetVer.Collector)
				changes.CollectorChanged = append(changes.CollectorChanged, deviceNumber)
			}

			// 检查标准点配置版本变更
			if targetVer.Point != backupVer.Point {
				log.Warnf("Standard point version changed for %s: %s -> %s",
					deviceNumber, backupVer.Point, targetVer.Point)
				changes.StdChanged = append(changes.StdChanged, deviceNumber)
			}
		}

	}
	return changes, nil
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
		log.Infof("config read name:%v not support check version", configReadName)
		return nil
	}
	// 定时任务检测版本更新
	if w.cron == nil {
		w.cron = cron.New(cron.WithSeconds())
	}
	s := config.GetRB().GetCheckTime()
	spec := fmt.Sprintf("*/%s * * * * *", s)
	log.Infof("check cron add :%v", spec)
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
func (w *worker) InitStd(changes *model.ConfigChangeEvent) error {
	if !config.GetRB().IsStdCalEnable() {
		return nil
	}
	if config.GetRB().Project.Source == cm.LocalFileConfigModName {
		return w.InitStdByLocal()
	}
	// 记录函数总耗时
	totalStart := time.Now()
	defer func() {
		log.Warnf("[InitStd] Total time: %v", time.Since(totalStart))
	}()

	// 获取所有目标设备
	allStart := time.Now()
	allDevices := cmutils.GetTargetDevice()
	log.Warnf("[InitStd] Get all devices (%d) in: %v", len(allDevices), time.Since(allStart))

	// 确定需要从CMDB拉取的设备
	var cmdbDevices []string
	var localDevices []string
	if changes != nil && len(changes.StdChanged) > 0 {
		cmdbDevices = changes.StdChanged
		log.Infof("[InitStd] Changes detected: %d devices to update", len(cmdbDevices))
	}

	// ========== 1. 变更设备的标准点获取 ==========
	cmdbStart := time.Now()
	cmdbStd, err := w.reader.GetStdData(w.GetTargetVersion(), cmdbDevices)
	if err != nil {
		return fmt.Errorf("failed to get std data from CMDB: %w", err)
	}
	log.Warnf("[InitStd] Fetched %d std points from CMDB for %d devices in: %v",
		len(cmdbStd.StdPointsInfo), len(cmdbDevices), time.Since(cmdbStart))

	// ========== 2. 未变更设备的标准点获取 ==========
	localStart := time.Now()
	localStd := &model.StdData{}
	if len(allDevices) > len(cmdbDevices) {
		localDevices = sliceExclude(allDevices, cmdbDevices)
		log.Debugf("[InitStd] Local devices to load: %d", len(localDevices))

		backupReader := cm.NewReader(cm.BackupModName)
		if localData, err := backupReader.GetStdData(w.GetConfigVersion(), localDevices); err == nil {
			localStd = localData
			log.Infof("[InitStd] Loaded %d std points from backup in: %v",
				len(localStd.StdPointsInfo), time.Since(localStart))
		} else {
			log.Warnf("[InitStd] Backup load failed: %v (time: %v)", err, time.Since(localStart))
		}
	} else {
		log.Warnf("[InitStd] No local devices to load (time: %v)", time.Since(localStart))
	}

	// ========== 3. 标准点数据合并 ==========
	mergeStart := time.Now()
	mergedStd := &model.StdData{
		StdPointsInfo: append(cmdbStd.StdPointsInfo, localStd.StdPointsInfo...),
	}

	w.SetStdData(mergedStd)
	log.Warnf("[InitStd] Merged %d std points (%d cmdb + %d local) in: %v",
		len(mergedStd.StdPointsInfo), len(cmdbStd.StdPointsInfo), len(localStd.StdPointsInfo),
		time.Since(mergeStart))

	// ========== 4. 变更设备的标准设备获取 ==========
	stdDeviceStart := time.Now()

	// 获取变更设备的标准设备
	changeDeviceStart := time.Now()
	changeDeviceMap := make(map[string]bool, len(cmdbStd.StdPointsInfo))
	for _, d := range cmdbStd.StdPointsInfo {
		changeDeviceMap[d.StdDevice] = true
	}
	log.Debugf("[InitStd] Created change device map in: %v", time.Since(changeDeviceStart))

	changeStdStart := time.Now()
	changeStdDevice, err := w.reader.GetStdDevice(nil, cmdbDevices)
	if err != nil {
		log.Warnf("[InitStd] Failed to get changed std devices: %v (time: %v)",
			err, time.Since(changeStdStart))
		return err
	}
	log.Infof("[InitStd] Got %d changed std devices in: %v",
		len(changeStdDevice.StdDevices), time.Since(changeStdStart))

	// ========== 5. 未变更设备的标准设备获取 ==========
	localStdStart := time.Now()
	localDeviceMap := make(map[string]bool, len(localStd.StdPointsInfo))
	for _, d := range localStd.StdPointsInfo {
		localDeviceMap[d.StdDevice] = true
	}
	log.Debugf("[InitStd] Created local device map in: %v", time.Since(localStdStart))

	backupReader := cm.NewReader(cm.BackupModName)
	localStdDevice, err := backupReader.GetStdDevice(localDeviceMap, localDevices)
	if err != nil {
		log.Warnf("[InitStd] Failed to get local std devices: %v (time: %v)",
			err, time.Since(localStdStart))
		return err
	}
	log.Warnf("[InitStd] Got %d local std devices in: %v",
		len(localStdDevice.StdDevices), time.Since(localStdStart))

	// ========== 6. 标准设备数据合并 ==========
	mergeDeviceStart := time.Now()

	// 合并标准设备列表
	mergedStdDevices := append(changeStdDevice.StdDevices, localStdDevice.StdDevices...)

	// 合并设备映射
	mergedStdDeviceMap := make(map[string]model.StdDevice)
	for k, v := range localStdDevice.StdDeviceMap {
		mergedStdDeviceMap[k] = v
	}
	for k, v := range changeStdDevice.StdDeviceMap {
		mergedStdDeviceMap[k] = v
	}

	// 合并简码映射
	mergedConciseCodeMap := make(map[string]string)
	for k, v := range localStdDevice.ConciseCodeMap {
		mergedConciseCodeMap[k] = v
	}
	for k, v := range changeStdDevice.ConciseCodeMap {
		mergedConciseCodeMap[k] = v
	}

	// 合并设备编号映射
	mergedDeviceNumberMap := make(map[string]string)
	for k, v := range localStdDevice.DeviceNumberMap {
		mergedDeviceNumberMap[k] = v
	}
	for k, v := range changeStdDevice.DeviceNumberMap {
		mergedDeviceNumberMap[k] = v
	}

	log.Warnf("[InitStd] Merged device maps in: %v", time.Since(mergeDeviceStart))

	// 创建合并后的标准设备数据结构
	createDataStart := time.Now()
	mergedStdDeviceData := &model.StdDeviceData{
		StdDevices:          mergedStdDevices,
		StdDeviceMap:        mergedStdDeviceMap,
		ConciseCodeMap:      mergedConciseCodeMap,
		DeviceNumberMap:     mergedDeviceNumberMap,
		StdPoints:           make(map[string]model3.StdInstancePointsInfo),
		StdPointsMutex:      sync.RWMutex{},
		ConciseCodeMapMutex: sync.RWMutex{},
	}
	for _, dev := range mergedStdDeviceData.StdDevices {
		mozuStr := fmt.Sprintf("%v", dev.MozuId)
		w.SetDeviceMozuID(definition.DeviceGidType(dev.DeviceGid), mozuStr)
	}
	// 标准测点绑定
	pointMapStart := time.Now()
	mergedPointMap := w.getStdPointMap()
	mergedStdDeviceData.StdPoints = mergedPointMap
	log.Warnf("[InitStd] Got and set point map in: %v", time.Since(pointMapStart))

	// 设置标准设备数据
	w.SetStdDeviceData(mergedStdDeviceData)

	log.Warnf("[InitStd] Created and set std device data in: %v", time.Since(createDataStart))
	log.Warnf("[InitStd] Merged std device data: %d devices total (%d changed + %d unchanged), "+
		"%d device mappings, %d concise codes (merge time: %v)",
		len(mergedStdDevices),
		len(changeStdDevice.StdDevices),
		len(localStdDevice.StdDevices),
		len(mergedStdDeviceMap),
		len(mergedConciseCodeMap),
		time.Since(mergeDeviceStart))

	log.Warnf("[InitStd] Std devices processing completed in: %v", time.Since(stdDeviceStart))
	return nil
}

// InitStdByLocal 本地模式加载标准层
func (w *worker) InitStdByLocal() error {
	localReader := cm.NewReader(cm.LocalFileConfigModName)
	std, err := localReader.GetStdData(w.GetTargetVersion(), nil)
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
	stdDevice, err := localReader.GetStdDevice(stdDeviceMap, nil)
	if err != nil {
		return err
	}
	for _, dev := range stdDevice.StdDevices {
		mozuStr := fmt.Sprintf("%v", dev.MozuId)
		w.SetDeviceMozuID(definition.DeviceGidType(dev.DeviceGid), mozuStr)
	}
	// 标准测点绑定
	stdDevice.StdPoints = w.getStdPointMap()
	w.SetStdDeviceData(stdDevice)

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

func (w *worker) InitDevicesAndTemplates(changes *model.ConfigChangeEvent) error {
	if config.GetRB().Project.Source == cm.LocalFileConfigModName {
		return w.initDevicesAndTemplatesByLocal()
	}
	// 获取变更设备列表
	changedDevices := []string{}
	if changes != nil && len(changes.CollectorChanged) > 0 {
		changedDevices = changes.CollectorChanged
	}
	return w.initDevicesAndTemplatesCore(
		w.reader,
		cm.NewReader(cm.BackupModName),
		true,
		true,
		changedDevices,
	)
}

func (w *worker) BackupInitDevicesAndTemplates(changes *model.ConfigChangeEvent) error {
	changedDevices := []string{}
	return w.initDevicesAndTemplatesCore(
		w.reader,
		cm.NewReader(cm.BackupModName),
		false,
		false,
		changedDevices,
	)
}

// 统一逻辑：设备设置、模板获取、模板处理
func (w *worker) initDevicesAndTemplatesCore(
	primaryReader cm.Reader,
	backupReader cm.Reader,
	saveDeviceConfig bool,
	saveTemplatesToFile bool,
	changedDevices []string) error {
	// 获取所有目标设备
	allDevices := cmutils.GetTargetDevice()
	log.Debugf("Processing %d target devices (%d changed)",
		len(allDevices), len(changedDevices))

	fullCollectDevices := make([]model.Device, 0)
	fullTboxDevices := make([]model.Device, 0)
	fullDeviceMap := make(map[string]any)

	// ================== 第一步：获取设备配置（混合来源） ==================

	// 1. 从主数据源获取变更设备配置
	var changedCollect, changedTbox []model.Device
	var changedDeviceMap map[string]any
	var err error
	if len(changedDevices) > 0 {
		changedCollect, changedTbox, changedDeviceMap, err = primaryReader.GetDevices(changedDevices)
		if err != nil {
			log.Warnf("Failed to ues primaryReader get changed devices: %v. ", err)
			return err
		}
		// 合并到全量结果
		fullCollectDevices = append(fullCollectDevices, changedCollect...)
		fullTboxDevices = append(fullTboxDevices, changedTbox...)
		for device, config := range changedDeviceMap {
			fullDeviceMap[device] = config
		}
	}
	// 2. 未变更设备从备读获取配置
	unchangedDevices := sliceExclude(allDevices, changedDevices)
	if len(unchangedDevices) > 0 {
		unchangedCollect, unchangedTbox, unchangedDeviceMap, err := backupReader.GetDevices(unchangedDevices)
		if err != nil {
			log.Warnf("Failed to get unchanged devices: %v.", err)
			return err
		}
		// 合并到全量结果
		fullCollectDevices = append(fullCollectDevices, unchangedCollect...)
		fullTboxDevices = append(fullTboxDevices, unchangedTbox...)
		for device, config := range unchangedDeviceMap {
			// 优先使用变更设备配置，所以这里只添加未获取过的设备
			if _, exists := fullDeviceMap[device]; !exists {
				fullDeviceMap[device] = config
			}
		}
	}

	// 设置采集层
	w.SetTboxDevice(fullTboxDevices)
	log.Warnf("Loaded %d collector devices (%d tbox, %d normal)",
		len(fullCollectDevices), len(fullTboxDevices),
		len(fullCollectDevices)-len(fullTboxDevices))

	// ================== 第二步：准备模板 ==================
	templateList, tpl2DeviceName := w.setDevicesAndPrepareTemplate(fullCollectDevices)
	dtList := utils.RemoveDuplicates(templateList)
	log.Debugf("Found %d unique templates: %v", len(dtList), dtList)

	// ================== 第三步：获取并处理模板（按变更关联性区分数据源） ==================
	var changedTemplates []string               // 与变更设备相关的模板
	var unchangedTemplates []string             // 只与未变更设备相关的模板
	var allTemplatesSet = make(map[string]bool) // 所有模板去重
	// 1. 区分两类模板
	for template, devices := range tpl2DeviceName {
		allTemplatesSet[template] = true
		isChanged := false
		for _, device := range devices {
			if contains(changedDevices, device) {
				isChanged = true
				break
			}
		}
		if isChanged {
			changedTemplates = append(changedTemplates, template)
		} else {
			unchangedTemplates = append(unchangedTemplates, template)
		}
	}
	log.Infof("Template breakdown: %d total, %d changed, %d unchanged",
		len(allTemplatesSet), len(changedTemplates), len(unchangedTemplates))

	// 2. 分别从不同数据源获取模板
	rawTemplateMap := make(map[string]any)
	// 获取与变更设备相关的模板（主数据源）
	if len(changedTemplates) > 0 {
		changedRawMap, err := primaryReader.GetTemplates(changedTemplates)
		if err != nil {
			log.Warnf("Failed to get changed templates from primary: %v. Trying backup", err)
			return err
		}
		// 合并到结果
		for name, info := range changedRawMap {
			rawTemplateMap[name] = info
		}
	}

	// 获取与未变更设备相关的模板（备份数据源）
	if len(unchangedTemplates) > 0 {
		unchangedRawMap, err := backupReader.GetTemplates(unchangedTemplates)
		if err != nil {
			log.Warnf("Failed to get unchanged templates from backup: %v. Trying primary", err)
			return err
		}
		// 合并到结果（避免覆盖变更模板）
		for name, info := range unchangedRawMap {
			if _, exists := rawTemplateMap[name]; !exists {
				rawTemplateMap[name] = info
			}
		}
	}
	for name, info := range rawTemplateMap {
		// 解析并更新内存中的模板
		m, err := cmutils.ParseCollectTemplateInfo(info)
		if err != nil {
			log.Errorf("Parse template '%s' failed: %v", name, err)
			continue
		}
		w.SetTemplate(name, m)
		log.Debugf("Updated template '%s' in memory", name)

		// 仅保存与变更设备相关的模板
		if !saveTemplatesToFile || !contains(changedTemplates, name) {
			continue
		}

		devicesToSave := []string{}
		if deviceNames, ok := tpl2DeviceName[name]; ok {
			// 只保存变更设备关联的模板
			for _, device := range deviceNames {
				if contains(changedDevices, device) {
					devicesToSave = append(devicesToSave, device)
				}
			}
		}

		for _, device := range devicesToSave {
			targetDir := consts.ProjectPath + "/" + device + "/" + consts.RelativeTemplateDir + "/"
			err := cmutils.SaveConfigMapToMultipleFile(targetDir, map[string]any{name: info})
			if err != nil {
				log.Warnf("Save template '%s' for device %s failed: %v", name, device, err)
			} else {
				log.Debugf("Saved template '%s' for device %s", name, device)
			}
		}
	}

	// ================== 第四步：保存设备配置（在模板文件保存之后，确保模板文件先于设备配置落盘） ==================
	if saveDeviceConfig && w.lastReaderName != cm.LocalFileConfigModName {
		// 只保存变更设备的配置
		filteredMap := make(map[string]any)
		for _, device := range changedDevices {
			if config, exists := fullDeviceMap[device]; exists {
				filteredMap[device] = config
			}
		}

		if len(filteredMap) > 0 {
			if err := cmutils.SaveConfigMapToDirFileWithVersion(
				filteredMap, consts.DeviceTag, w.GetTargetVersion()); err != nil {
				log.Warnf("Partial save device config failed: %v", err)
			} else {
				log.Infof("Saved %d changed device configs", len(filteredMap))
			}
		}
	}

	// ================== 第五步：处理设备模板关联 ==================
	w.processDeviceTemplates(fullCollectDevices)
	// 如果有开启热备，记录gid和采集器编号的映射
	if config.GetRB().HotStandbyEnable() {
		gid2CollectDev := make(map[definition.DeviceGidType]string)
		for _, d := range fullCollectDevices {
			gid2CollectDev[d.Gid] = d.BelongCollectorDevice
			for _, device := range d.SubDevices {
				gid2CollectDev[device.Gid] = d.BelongCollectorDevice
			}
		}
		for _, d := range fullTboxDevices {
			gid2CollectDev[d.Gid] = d.ID
		}
		w.SetGid2CollectMap(gid2CollectDev)
	}

	log.Warnf("Device and template initialization complete！")
	return nil
}

// 统一逻辑：设备设置、模板获取、模板处理
func (w *worker) initDevicesAndTemplatesByLocal() error {
	localReader := cm.NewReader(cm.LocalFileConfigModName)
	// 设置采集层
	collectDevices, tboxDevices, _, err := localReader.GetDevices(nil)
	if err != nil {
		log.Warnf("get devices failed: %s", err.Error())
		return err
	}
	w.SetTboxDevice(tboxDevices)

	templateList, _ := w.setDevicesAndPrepareTemplate(collectDevices)
	dtList := utils.RemoveDuplicates(templateList)

	rawTemplateMap, err := localReader.GetTemplates(dtList)
	if err != nil {
		log.Warnf("GetTemplates failed:%s", err.Error())
		return err
	}
	for name, info := range rawTemplateMap {
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
			// 表达式转换
			if p.ExprDef.Expr != "" {
				p.ExprDef.Expr = utils.TransformExpression(p.ExprDef.Expr)
			}
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

// GetCollectorGidByGid 通过采集设备gid获取采集器gid
func (w *worker) GetCollectorGidByGid(gid definition.DeviceGidType) string {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()

	d, ok := w.devices[gid]
	if !ok {
		return ""
	}
	pGid := d.BelongCollectorGid

	return pGid
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
	// 记录最后一个有效的 mozu 值
	if mozu != "" && mozu != "0" {
		w.lastMozu = mozu
	}
}

// GetLastMozu 获取最后设置的有效 mozu 值
func (w *worker) GetLastMozu() string {
	w.mozuIDMutex.RLock()
	defer w.mozuIDMutex.RUnlock()

	return w.lastMozu
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

func WatchConfigVersionChange() <-chan *model.ConfigChangeEvent {
	return cm.ConfigVersionChangedChan()
}

// GetDeviceInfo 获取设备信息
func (w *worker) GetDeviceInfo(deviceGiD definition.DeviceGidType) (model.DeviceInfo, bool) {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()

	d, ok := w.devices[deviceGiD]
	return d.GetDeviceInfo(), ok
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
	//log.Debugf("check deviceNumber:%v", deviceNumbers)
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
	//collectorVersionChangedDevices := make(map[string][]string)
	//pointVersionChangedDevices := make(map[string][]string)
	//var collectorChange, stdChange bool
	var changed bool
	collectorChanged := make([]string, 0)
	stdChanged := make([]string, 0)
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
			collectorChanged = append(collectorChanged, deviceNumber)
			changed = true
		}
		pointMatch := cmdbStdVersion == workStdVersion
		if !pointMatch {
			stdChanged = append(stdChanged, deviceNumber)
			changed = true
		}
		// 填充到结果中
		result[deviceNumber] = VersionResult{
			Collector: collectorMatch,
			Point:     pointMatch,
		}
	}

	if changed {
		log.Warnf("version changed! collector changed:%v, std changed:%v", collectorChanged, stdChanged)
		cm.ConfigVersionChangedChan() <- &model.ConfigChangeEvent{
			CollectorChanged: collectorChanged,
			StdChanged:       stdChanged,
		}
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

// GetStdDeviceGidByNumber 通过标准设备编号查找设备Gid
func (w *worker) GetStdDeviceGidByNumber(deviceNumber string) (definition.DeviceGidType, bool) {
	gid, ok := w.stdDevice.GetGidByDeviceNumber(deviceNumber)
	if !ok {
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

// 辅助函数：从切片A中排除切片B的元素
func sliceExclude(a, b []string) []string {
	m := make(map[string]bool)
	for _, v := range b {
		m[v] = true
	}

	var result []string
	for _, v := range a {
		if !m[v] {
			result = append(result, v)
		}
	}
	return result
}

// 查找缺失设备
func findMissingDevices(all []string, found map[string]any) []string {
	missing := []string{}
	for _, d := range all {
		if _, exists := found[d]; !exists {
			missing = append(missing, d)
		}
	}
	return missing
}

// 判断切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetMapSubDevicesGid2Id 子设备gid 到 子设备设备编号的对应关系
func (w *worker) GetMapSubDevicesGid2Id() map[definition.DeviceGidType]string {
	w.deviceMutex.RLock()
	defer w.deviceMutex.RUnlock()
	subGid2Id := make(map[definition.DeviceGidType]string)
	for _, d := range w.devices {
		for _, sub := range d.SubDevices {
			subGid2Id[sub.Gid] = sub.ID
		}
	}
	return subGid2Id
}

// SetGid2CollectMap 设置设备gid到负责采集的设备编号的映射
func (w *worker) SetGid2CollectMap(gid2Dev map[definition.DeviceGidType]string) {
	w.gid2CollectDevMutex.Lock()
	defer w.gid2CollectDevMutex.Unlock()
	w.gid2CollectDev = gid2Dev
}

// GetGid2CollectMap 获取设备gid到负责采集的设备编号的映射，注意这里返回的map只能读不能写，否则会有多协程并发问题
func (w *worker) GetGid2CollectMap() map[definition.DeviceGidType]string {
	w.gid2CollectDevMutex.RLock()
	defer w.gid2CollectDevMutex.RUnlock()
	return w.gid2CollectDev
}

// SetStd2CollectMap 设置标准设备gid到负责采集的设备编号的映射
func (w *worker) SetStd2CollectMap(collectGid2Std map[definition.DataPointIDType][]string) {
	std2CollectDev := make(map[definition.DeviceGidType]string)
	gid2Collect := w.GetGid2CollectMap()
	if len(gid2Collect) == 0 {
		return
	}
	for point, stdIds := range collectGid2Std {
		collectDev, has := gid2Collect[definition.DeviceGidType(point.GetPointGid())]
		if !has {
			log.Warnf("GetPoint2CollectMap not found point: %v", point)
			continue
		}
		for _, stdId := range stdIds {
			std2CollectDev[definition.DeviceGidType(stdId)] = collectDev
		}
	}
	w.std2CollectDevMutex.Lock()
	defer w.std2CollectDevMutex.Unlock()
	w.std2CollectDev = std2CollectDev
}

// GetStd2CollectMap 获取标准设备gid到负责采集的设备编号的映射，注意这里返回的map只能读不能写，否则会有多协程并发问题
func (w *worker) GetStd2CollectMap() map[definition.DeviceGidType]string {
	w.std2CollectDevMutex.RLock()
	defer w.std2CollectDevMutex.RUnlock()
	return w.std2CollectDev
}

// NeedCollect 检查是否需要采集
func (w *worker) NeedCollect(deviceGid definition.DeviceGidType) bool {
	// 不采集备设备时只加载了主设备的测点，直接返回true
	if false == config.GetRB().Task.Local.CollectSlave {
		return true
	}
	gid2Dev := w.GetGid2CollectMap()
	devIsMaster := hotstandby.GetHotStandbyManager().GetDevStatusMap()
	if len(gid2Dev) == 0 || len(devIsMaster) == 0 {
		return true
	}
	// 为从设备负责采集的范围
	if devNum, ok := gid2Dev[deviceGid]; ok {
		if status, has := devIsMaster[devNum]; has && status == false {
			return config.GetRB().Task.Local.CollectSlave
		}
	}
	return true
}

// NeedDistribute 检查是否需要分发
func (w *worker) NeedDistribute(deviceGid definition.DeviceGidType, dataType int) bool {
	// 不采集备设备时只加载了主设备的测点，直接返回true
	if false == config.GetRB().Task.Local.CollectSlave {
		return true
	}
	devIsMaster := hotstandby.GetHotStandbyManager().GetDevStatusMap()
	var gid2Dev map[definition.DeviceGidType]string
	if dataType == definition.KafkaDataTypeStd {
		gid2Dev = w.GetStd2CollectMap()
	} else {
		gid2Dev = w.GetGid2CollectMap()
	}
	if len(gid2Dev) == 0 || len(devIsMaster) == 0 {
		return true
	}
	// 根据所属采集器是否为master来决定是否分发
	if devNum, ok := gid2Dev[deviceGid]; ok {
		if status, has := devIsMaster[devNum]; has {
			return status
		}
	}
	return true
}
