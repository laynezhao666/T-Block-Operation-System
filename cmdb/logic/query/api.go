// Package query 采集相关配置查询
package query

import (
	"archive/zip"
	"bytes"
	"cmdb/entity/cond"
	"cmdb/repo/db"
	"cmdb/util/convutil"
	"common/entity/model"
	"context"
	"encoding/json"
	"etrpc-go/log"
	"etrpc-go/util/copyutil"
	"etrpc-go/util/httputil"
	"fmt"
	"sync"
	"time"
	"trpcprotocol/cmdb"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/http"
)

// IConfigQueryApi 采集配置查询相关接口
type IConfigQueryApi interface {
	// GetCollectorDevice 获取采集器下的子设备
	GetCollectorDevice(ctx context.Context, deviceNumbers []string) (*cmdb.RspGetCollectorDevice, error)
	// GetCollectorPoint  获取采集器下的标准测点
	GetCollectorPoint(ctx context.Context, deviceNumbers []string) (*cmdb.RspGetCollectorPoint, error)
	// GetCollectorTemplate 获取采集模版配置信息
	GetCollectorTemplate(ctx context.Context, templateNames []string) (*cmdb.RspGetCollectorTemplate, error)
	// ExportCollectorConfig 导出所有的采集器配置为文件
	ExportCollectorConfig(ctx context.Context, req *cmdb.ReqExportCollectorConfig) (*emptypb.Empty, error)
	// GetCollectorDataVer 获取采集器数据版本
	GetCollectorDataVer(ctx context.Context, deviceNumbers []string) (*cmdb.RspGetConfigModifyTime, error)

	// GetDeviceEntity 获取设备列表
	GetDeviceEntity(ctx context.Context, req *cmdb.ReqGetDeviceEntity) (*cmdb.RspGetDeviceEntity, error)
	// GetDevicePoint 获取标准测点列表
	GetDevicePoint(ctx context.Context, req *cmdb.ReqGetDevicePoint) (*cmdb.RspGetDevicePoint, error)
	// ListCollectorDevice 获取采集器列表
	ListCollectorDevice(ctx context.Context, req *cmdb.ReqListCollectorDevice) (*cmdb.RspListCollectorDevice, error)

	// GetMozuInfo 获取模组信息列表
	GetMozuInfo(ctx context.Context, req *cmdb.ReqGetMozuInfo) (*cmdb.RspGetMozuInfo, error)
}

var (
	collectorApi IConfigQueryApi
	initOnce     sync.Once
)

// GetConfigQueryApi 创建采集配置查询相关实现类
func GetConfigQueryApi() IConfigQueryApi {
	initOnce.Do(func() {
		obj := &configQueryApiImpl{
			deviceEntityDao:           db.NewDeviceEntityDao(),
			devicePointDao:            db.NewDevicePointDao(),
			collectorDeviceDao:        db.NewCollectorDeviceDao(),
			collectorTemplateDao:      db.NewCollectorTemplateDao(),
			collectorTemplatePointDao: db.NewCollectorTemplatePointDao(),
			mozuInfoDao:               db.NewMozuInfoDao(),

			mozuVerCache: make(map[int32]*model.MozuInfo),
		}
		collectorApi = obj
		// 定时刷新数据是否变化
		go func() {
			obj.watchCollectorVer()
			timer := time.NewTimer(time.Second * 5)
			for {
				select {
				case <-timer.C:
					obj.watchCollectorVer()
					timer.Reset(time.Second * 5)
				}
			}
		}()
	})
	return collectorApi
}

type configQueryApiImpl struct {
	deviceEntityDao           db.IDeviceEntityDao
	devicePointDao            db.IDevicePointDao
	collectorDeviceDao        db.ICollectorDeviceDao
	collectorTemplateDao      db.ICollectorTemplateDao
	collectorTemplatePointDao db.ICollectorTemplatePointDao
	mozuInfoDao               db.IMozuInfoDao

	mozuVerCache       map[int32]*model.MozuInfo // 模组版本缓存信息
	collectorVerChange sync.Map                  // 每个采集器版本是否变化
	collectorVerCache  sync.Map                  // 模组数据版本cache信息
}

func (c *configQueryApiImpl) watchCollectorVer() {
	ctx := trpc.BackgroundContext()
	allMozu, _, err := c.mozuInfoDao.List(ctx, &cond.ListMozuInfoCond{})
	if err != nil {
		log.AlarmContextf(ctx, "watch collector ver: get all mozu info fail, err:%v", err)
		return
	}
	// 检查是否有模组版本变化
	for _, mozu := range allMozu {
		if last, ok := c.mozuVerCache[mozu.MozuId]; ok && last.PublishVersion == mozu.PublishVersion {
			continue
		}
		log.InfoContextf(ctx, "watch collector ver: mozu_id:[%d] ver change", mozu.MozuId)
		c.mozuVerCache[mozu.MozuId] = mozu
		collectors, _, err := c.collectorDeviceDao.GetList(ctx, &cond.ListCollectorDeviceCond{
			CollectorType: []int32{model.CollectorTypeTbox, model.CollectorTypeVendorBox,
				model.CollectorTypeDoor, model.CollectorTypeTone},
			MozuId: []int32{mozu.MozuId},
		})
		if err != nil {
			log.AlarmContextf(ctx, "watch collector ver: get collector of mozu_id:[%d] fail, err:%v",
				mozu.MozuId, err)
			continue
		}
		pointCollectors, err := c.devicePointDao.GetDistinctCollector(ctx, mozu.MozuId)
		if err != nil {
			log.AlarmContextf(ctx, "watch collector ver: get point collector of mozu_id:[%d] fail, err:%v",
				mozu.MozuId, err)
			continue
		}
		for _, collector := range collectors {
			c.collectorVerChange.Store(collector.DeviceNumber, mozu.MozuId)
		}
		for _, belongCollector := range pointCollectors {
			c.collectorVerChange.Store(belongCollector, mozu.MozuId)
		}
	}
}

// GetCollectorDevice 获取采集设备
func (c *configQueryApiImpl) GetCollectorDevice(ctx context.Context, deviceNumbers []string) (
	*cmdb.RspGetCollectorDevice, error) {
	// 1、初始化返回信息
	majorDevices, _, err := c.collectorDeviceDao.GetList(ctx, &cond.ListCollectorDeviceCond{
		DeviceNumber: deviceNumbers,
		CollectorType: []int32{model.CollectorTypeTbox, model.CollectorTypeVendorBox,
			model.CollectorTypeDoor, model.CollectorTypeTone},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get collector box list fail")
	}
	rspConfigMap := make(map[string]*cmdb.RspGetCollectorDevice_CollectorDeviceInfo)
	for _, majorDevice := range majorDevices {
		rspConfig := &cmdb.RspGetCollectorDevice_CollectorDeviceInfo{
			DeviceGid:     majorDevice.DeviceGid,
			DeviceNumber:  majorDevice.DeviceNumber,
			DeviceTypeEn:  majorDevice.DeviceTypeEn,
			DeviceTypeZh:  majorDevice.DeviceTypeZh,
			CollectorType: majorDevice.CollectorType,
			DeviceCode:    majorDevice.DeviceCode,
			DeviceName:    majorDevice.DeviceName,
			Channel:       convutil.JsonStrToPbStruct(majorDevice.ChannelLink),
			Tpl:           convutil.JsonStrToPbStruct(majorDevice.TemplateInfo),
			MozuId:        majorDevice.MozuId,
			SubDevices:    make([]*cmdb.RspGetCollectorDevice_CollectorDeviceInfo_SubDevice, 0),
		}
		rspConfigMap[majorDevice.DeviceNumber] = rspConfig
	}
	// 2、获取所有的Box子设备列表
	subDevices, _, err := c.collectorDeviceDao.GetList(ctx, &cond.ListCollectorDeviceCond{
		ParentDeviceNumber: deviceNumbers,
	})
	if err != nil {
		return nil, err
	}
	// 3、将每个子设备写到对应的采集器列表中
	for _, subDevice := range subDevices {
		rspSubDevice := &cmdb.RspGetCollectorDevice_CollectorDeviceInfo_SubDevice{
			DeviceGid:     subDevice.DeviceGid,
			DeviceNumber:  subDevice.DeviceNumber,
			DeviceTypeEn:  subDevice.DeviceTypeEn,
			DeviceTypeZh:  subDevice.DeviceTypeZh,
			CollectorType: subDevice.CollectorType,
			DeviceCode:    subDevice.DeviceCode,
			DeviceName:    subDevice.DeviceName,
			Channel:       convutil.JsonStrToPbStruct(subDevice.ChannelLink),
			Tpl:           convutil.JsonStrToPbStruct(subDevice.TemplateInfo),
			MozuId:        subDevice.MozuId,
		}
		if rspConfig, ok := rspConfigMap[subDevice.ParentDeviceNumber]; ok {
			rspConfig.SubDevices = append(rspConfig.SubDevices, rspSubDevice)
		}
	}
	return &cmdb.RspGetCollectorDevice{
		ConfigMap: rspConfigMap,
	}, nil
}

// GetCollectorPoint 获取采集测点
func (c *configQueryApiImpl) GetCollectorPoint(ctx context.Context, deviceNumbers []string) (
	*cmdb.RspGetCollectorPoint, error) {
	// 1、初始化返回信息
	rspConfigMap := make(map[string]*cmdb.RspGetCollectorPoint_CollectorTemplateInfo)
	for _, deviceNumber := range deviceNumbers {
		rspConfig := &cmdb.RspGetCollectorPoint_CollectorTemplateInfo{
			DevicePoints: make([]*cmdb.RspGetCollectorPoint_CollectorTemplateInfo_DevicePoint, 0),
		}
		rspConfigMap[deviceNumber] = rspConfig
	}
	// 2、获取所有设备测点数据
	devicePoints, _, err := c.devicePointDao.GetList(ctx, &cond.ListDevicePointCond{
		BelongCollector: deviceNumbers,
	})
	if err != nil {
		return nil, err
	}
	pointMap := lo.SliceToMap(devicePoints, func(item *model.DevicePoint) (string, *model.DevicePoint) {
		return item.PointKey, item
	})
	// 部分标准到标准的点涉及常量点,故取出所有的常量测点
	if len(devicePoints) > 0 {
		constDevicePoints, _, err := c.devicePointDao.GetList(ctx, &cond.ListDevicePointCond{
			MozuId:    []int32{devicePoints[0].MozuId},
			PointType: []int32{model.PointTypeConstant},
		})
		if err != nil {
			return nil, err
		}
		for _, item := range constDevicePoints {
			pointMap[item.PointKey] = item
		}
	}
	// 3、将所有的设备测点和采集设备进行关联
	for _, item := range devicePoints {
		if err = item.StdToCollectorPoint(pointMap); err != nil {
			log.ErrorContextf(ctx, "point:[%s], convert to collector point fail, err: %v", item.PointKey, err)
			continue
		}
		rspPoint := &cmdb.RspGetCollectorPoint_CollectorTemplateInfo_DevicePoint{}
		_ = copyutil.Copy(item, rspPoint)
		rspConfigMap[item.BelongCollector].DevicePoints = append(rspConfigMap[item.BelongCollector].DevicePoints, rspPoint)
	}
	return &cmdb.RspGetCollectorPoint{
		ConfigMap: rspConfigMap,
	}, nil
}

// GetCollectorTemplate 获取采集模版
func (c *configQueryApiImpl) GetCollectorTemplate(ctx context.Context, templateNames []string) (
	*cmdb.RspGetCollectorTemplate, error) {
	// 1、初始化返回数据结构
	rspTemplateMap := make(map[string]*cmdb.RspGetCollectorTemplate_CollectorTemplateInfo)
	for _, templateName := range templateNames {
		rspTemplate := &cmdb.RspGetCollectorTemplate_CollectorTemplateInfo{
			Points: make([]*cmdb.RspGetCollectorTemplate_CollectorTemplateInfo_TemplatePoint, 0),
		}
		rspTemplateMap[templateName] = rspTemplate
	}
	// 2.1 获取所有的采集模版信息
	templates, err := c.collectorTemplateDao.QueryCollectorTemplate(ctx, templateNames)
	if err != nil {
		return nil, err
	}
	// 2.2 将采集模版信息和响应体进行关联
	for _, template := range templates {
		rspTemplateMap[template.TemplateName].Drvinfo =
			&cmdb.RspGetCollectorTemplate_CollectorTemplateInfo_CollectorTemplate{
				Cls:     template.DeviceTypeEn,
				Drvlib:  template.ProtocolType,
				Protver: template.ProtocolVersion,
				Vendor:  template.Manufacturer,
				Extend:  template.ProtocolExtend,
			}
	}
	// 3.1 获取所有模版对应的测点数据
	templatePoints, err := c.collectorTemplatePointDao.QueryCollectorTemplatePoint(ctx, templateNames)
	if err != nil {
		return nil, err
	}
	for _, point := range templatePoints {
		rspTemplatePoint := &cmdb.RspGetCollectorTemplate_CollectorTemplateInfo_TemplatePoint{
			PointNameEn: point.PointNameEn,
			PointNameZh: point.PointNameZh,
			PointType:   point.PointType,
			PointRw:     point.PointRw,
			SubDevice:   point.SubDevice,
			Deltadef:    convutil.JsonStrToPbStruct(point.DeltaDef),
			Verifydef:   convutil.JsonStrToPbStruct(point.VerifyDef),
			Expdef:      convutil.JsonStrToPbStruct(point.ExpDef),
			Protdef:     convutil.JsonStrToPbStruct(point.ProtDef),
			Valdef:      convutil.JsonStrToPbStruct(point.ValDef),
			Simulator:   convutil.JsonStrToPbStruct(point.Simulator),
		}
		if rspTemplate, ok := rspTemplateMap[point.TemplateName]; ok {
			rspTemplate.Points = append(rspTemplate.Points, rspTemplatePoint)
		}
	}
	return &cmdb.RspGetCollectorTemplate{
		ConfigMap: rspTemplateMap,
	}, nil
}

// ExportCollectorConfig 批量导出采集器的所有配置
func (c *configQueryApiImpl) ExportCollectorConfig(ctx context.Context, req *cmdb.ReqExportCollectorConfig) (
	*emptypb.Empty, error) {
	dtoReq := &cond.ListCollectorDeviceCond{
		DeviceNumber: req.DeviceNumber,
	}
	if req.MozuId > 0 {
		dtoReq.MozuId = []int32{req.MozuId}
	}
	if req.CollectorType > 0 && req.CollectorType <= 4 {
		dtoReq.CollectorType = []int32{req.CollectorType}
	}
	// 取出所有的采集设备
	collectorDevices, _, err := c.collectorDeviceDao.GetList(ctx, dtoReq)
	if err != nil {
		return nil, errors.Wrapf(err, "get collector device list fail")
	}
	if len(collectorDevices) == 0 {
		return nil, fmt.Errorf("bad req condition, no relate collector devices found")
	}
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)
	// 依次处理每个采集器
	for _, device := range collectorDevices {
		// 获取采集设备
		collectorDeviceRsp, err := c.GetCollectorDevice(ctx, []string{device.DeviceNumber})
		if err != nil {
			return nil, errors.Wrapf(err, "get collector device info of [%s] fail", device.DeviceNumber)
		}
		// 拿到采集设备关联的采集模版
		templateNames := make([]string, 0)
		if collectorInfo, ok := collectorDeviceRsp.ConfigMap[device.DeviceNumber]; ok {
			if collectorInfo.CollectorType == model.CollectorTypeVendorBox {
				templateNames = append(templateNames, (collectorInfo.Tpl.AsMap()["tplnm"]).(string))
			} else {
				for _, subDevice := range collectorInfo.SubDevices {
					templateNames = append(templateNames, (subDevice.Tpl.AsMap()["tplnm"]).(string))
				}
			}
		}
		stdDeviceRsp, err := c.GetDeviceEntity(ctx, &cmdb.ReqGetDeviceEntity{BelongCollector: device.DeviceNumber})
		if err != nil {
			return nil, errors.Wrapf(err, "get std device list of [%s] fail", device.DeviceNumber)
		}
		// 获取标准测点
		collectorPointRsp, err := c.GetCollectorPoint(ctx, []string{device.DeviceNumber})
		if err != nil {
			return nil, errors.Wrapf(err, "get device point list of [%s] fail", device.DeviceNumber)
		}
		// 获取相关采集模版
		templatesRsp, err := c.GetCollectorTemplate(ctx, templateNames)
		if err != nil {
			return nil, errors.Wrapf(err, "get template info of [%s] fail", device.DeviceNumber)
		}
		// 写入数据到对应的配置文件
		_ = writeSingleFile(collectorDeviceRsp, zipWriter, device.DeviceNumber, "devices", true)
		_ = writeSingleFile(collectorPointRsp, zipWriter, device.DeviceNumber, "std", true)
		_ = writeSingleFile(stdDeviceRsp, zipWriter, device.DeviceNumber, "std_device", true)
		for templateName, templateContent := range templatesRsp.ConfigMap {
			_ = writeSingleFile(templateContent, zipWriter,
				fmt.Sprintf("%s/%s", device.DeviceNumber, "templates"), templateName, false)
		}
	}
	// 设置为导出文件
	msg := trpc.Message(ctx)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	head := http.Head(ctx)
	head.Response.Header().Set("Content-Type", "application/zip")
	head.Response.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=data_%s.zip", time.Now().Format("2006_01_02_15_04_05")))
	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}
	_, err = zipBuffer.WriteTo(head.Response)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// writeSingleFile 写入单个配置文件
func writeSingleFile(data any, writer *zip.Writer, deviceNumber, dataType string, wrap bool) error {
	fileWriter, err := writer.Create(fmt.Sprintf("%s/%s.json", deviceNumber, dataType))
	if err != nil {
		return errors.Wrapf(err, "create %s file of [%s] fail", dataType, deviceNumber)
	}
	wrapData := data
	if wrap {
		wrapData = httputil.ResponseEntity[any]{
			Data: data,
		}
	}
	marshal, _ := json.Marshal(wrapData)
	_, err = fileWriter.Write(marshal)
	if err != nil {
		return errors.Wrapf(err, "write %s file of [%s] fail", dataType, deviceNumber)
	}
	return nil
}

// GetDeviceEntity 获取设备信息
func (c *configQueryApiImpl) GetDeviceEntity(ctx context.Context, req *cmdb.ReqGetDeviceEntity) (
	*cmdb.RspGetDeviceEntity, error) {
	dtoReq := &cond.ListDeviceEntityCond{}
	rspData := make([]*cmdb.RspGetDeviceEntity_DeviceEntity, 0)
	_ = copyutil.Copy(req, dtoReq)
	// 处理所属采集器参数
	if req.BelongCollector != "" {
		deviceNumbers, err := c.devicePointDao.GetDeviceNumberByCollector(ctx, req.BelongCollector)
		if err != nil {
			return nil, err
		}
		if len(deviceNumbers) == 0 {
			return &cmdb.RspGetDeviceEntity{
				List:  rspData,
				Total: 0,
			}, err
		}
		if len(dtoReq.DeviceNumber) == 0 {
			dtoReq.DeviceNumber = deviceNumbers
		} else {
			deviceNumberMap := lo.SliceToMap(deviceNumbers, func(item string) (string, struct{}) {
				return item, struct{}{}
			})
			dtoReq.DeviceNumber = lo.Filter(dtoReq.DeviceNumber, func(item string, index int) bool {
				_, ok := deviceNumberMap[item]
				return ok
			})
			if len(dtoReq.DeviceNumber) == 0 {
				return &cmdb.RspGetDeviceEntity{
					List:  rspData,
					Total: 0,
				}, nil
			}
		}
	}
	devices, total, err := c.deviceEntityDao.GetList(ctx, dtoReq)
	if err != nil {
		return nil, err
	}
	for _, item := range devices {
		rspItem := &cmdb.RspGetDeviceEntity_DeviceEntity{}
		if err = copyutil.Copy(item, rspItem); err != nil {
			return nil, errors.Wrapf(err, "copy db field to rsp field fail")
		}
		rspItem.CreateAt = item.CreateAt.Format(time.DateTime)
		rspItem.UpdateAt = item.UpdateAt.Format(time.DateTime)
		rspData = append(rspData, rspItem)
	}
	return &cmdb.RspGetDeviceEntity{
		List:  rspData,
		Total: int32(total),
	}, nil
}

// GetDevicePoint 获取设备测点信息
func (c *configQueryApiImpl) GetDevicePoint(ctx context.Context, req *cmdb.ReqGetDevicePoint) (
	*cmdb.RspGetDevicePoint, error) {
	dtoReq := &cond.ListDevicePointCond{}
	_ = copyutil.Copy(req, dtoReq)
	points, total, err := c.devicePointDao.GetList(ctx, dtoReq)
	if err != nil {
		return nil, err
	}
	rspData := make([]*cmdb.RspGetDevicePoint_DevicePoint, 0)
	for _, item := range points {
		rspItem := &cmdb.RspGetDevicePoint_DevicePoint{}
		if err = copyutil.Copy(item, rspItem); err != nil {
			return nil, errors.Wrapf(err, "copy db field to rsp field fail")
		}
		rspItem.CreateAt = item.CreateAt.Format(time.DateTime)
		rspItem.UpdateAt = item.UpdateAt.Format(time.DateTime)
		rspData = append(rspData, rspItem)
	}
	return &cmdb.RspGetDevicePoint{
		List:  rspData,
		Total: int32(total),
	}, nil
}

// GetMozuInfo 获取模组信息
func (c *configQueryApiImpl) GetMozuInfo(ctx context.Context, req *cmdb.ReqGetMozuInfo) (
	*cmdb.RspGetMozuInfo, error) {
	list, _, err := c.mozuInfoDao.List(ctx, &cond.ListMozuInfoCond{
		MozuId: req.MozuId,
	})
	if err != nil {
		return nil, err
	}
	rspList := lo.Map(list, func(item *model.MozuInfo, index int) *cmdb.RspGetMozuInfo_MozuInfo {
		rspItem := &cmdb.RspGetMozuInfo_MozuInfo{}
		_ = copyutil.Copy(item, rspItem)
		return rspItem
	})
	return &cmdb.RspGetMozuInfo{
		List:  rspList,
		Total: int32(len(rspList)),
	}, nil
}

// ListCollectorDevice 获取采集设备列表
func (c *configQueryApiImpl) ListCollectorDevice(ctx context.Context, req *cmdb.ReqListCollectorDevice) (*cmdb.RspListCollectorDevice, error) {
	collectorDeviceCond := &cond.ListCollectorDeviceCond{}
	_ = copyutil.Copy(req, collectorDeviceCond)
	// 处理协议类型参数，根据条件查找出符合条件的模版
	if len(req.ProtocolType) > 0 {
		templates, _, err := c.collectorTemplateDao.GetList(ctx, &cond.ListCollectorTemplateCond{
			TemplateName: req.TemplateName,
			ProtocolType: req.ProtocolType,
			MozuId:       req.MozuId,
		})
		if err != nil {
			return nil, err
		}
		if len(templates) == 0 {
			return &cmdb.RspListCollectorDevice{
				List:  make([]*cmdb.CollectorDevice, 0),
				Total: 0,
			}, nil
		}
		collectorDeviceCond.TemplateName = lo.Map(templates, func(item *model.CollectorTemplate, index int) string {
			return item.TemplateName
		})
	}
	// 获取采集设备列表
	devices, total, err := c.collectorDeviceDao.GetList(ctx, collectorDeviceCond)
	if err != nil {
		return nil, err
	}
	return &cmdb.RspListCollectorDevice{
		List: lo.Map(devices, func(item *model.CollectorDevice, index int) *cmdb.CollectorDevice {
			rspItem := &cmdb.CollectorDevice{}
			_ = copyutil.Copy(item, rspItem)
			return rspItem
		}),
		Total: int32(total),
	}, nil
}

// GetCollectorDataVer 获取采集设备配置版本信息，用于采集层判断是否需要更新配置
func (c *configQueryApiImpl) GetCollectorDataVer(ctx context.Context, deviceNumbers []string) (
	*cmdb.RspGetConfigModifyTime, error) {
	reloadDeviceNumbers := make([]string, 0, len(deviceNumbers))
	res := make(map[string]*cmdb.RspGetConfigModifyTime_ModifyTime)
	for _, item := range deviceNumbers {
		if _, ok := c.collectorVerChange.Load(item); ok {
			reloadDeviceNumbers = append(reloadDeviceNumbers, item)
			continue
		}
		if val, ok := c.collectorVerCache.Load(item); !ok {
			reloadDeviceNumbers = append(reloadDeviceNumbers, item)
		} else {
			res[item] = val.(*cmdb.RspGetConfigModifyTime_ModifyTime)
		}
	}
	if len(reloadDeviceNumbers) > 0 {
		pointVer, err := c.devicePointDao.GetCollectorDataVer(ctx, reloadDeviceNumbers)
		if err != nil {
			return nil, err
		}
		collectorVer, err := c.collectorDeviceDao.GetCollectorDataVer(ctx, reloadDeviceNumbers)
		if err != nil {
			return nil, err
		}
		for _, deviceNumber := range reloadDeviceNumbers {
			pointVerStr := fmt.Sprint(pointVer[deviceNumber])
			collectorVerStr := fmt.Sprint(collectorVer[deviceNumber])

			// 若编号在 collectorVerChange 中（模组版本变化触发），但在 pointVer/collectorVer 中不存在
			// 则使用对应模组的最近更新时间+0作为兜底版本号
			if mozuId, ok := c.collectorVerChange.Load(deviceNumber); ok {
				if _, exists := pointVer[deviceNumber]; !exists {
					if mozuInfo, mOk := c.mozuVerCache[mozuId.(int32)]; mOk {
						pointVerStr = fmt.Sprintf("%d-0", mozuInfo.UpdateAt.Unix())
					}
				}
				if _, exists := collectorVer[deviceNumber]; !exists {
					if mozuInfo, mOk := c.mozuVerCache[mozuId.(int32)]; mOk {
						collectorVerStr = fmt.Sprintf("%d-0", mozuInfo.UpdateAt.Unix())
					}
				}
			}

			item := &cmdb.RspGetConfigModifyTime_ModifyTime{
				Point:     pointVerStr,
				Collector: collectorVerStr,
			}
			res[deviceNumber] = item
			c.collectorVerCache.Store(deviceNumber, item)
			c.collectorVerChange.Delete(deviceNumber)
		}
	}
	return &cmdb.RspGetConfigModifyTime{
		ConfigMap: res,
	}, nil
}
