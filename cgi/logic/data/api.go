package data

import (
	"cgi/entity/dto"
	"cgi/entity/errcode"
	"cgi/repo/db"
	"common/entity/consts"
	"common/entity/model"
	"context"
	"etrpc-go/log"
	"etrpc-go/util/copyutil"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	pb "trpcprotocol/cgi"
	"trpcprotocol/data-cache"
	dataPb "trpcprotocol/data-query"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-go/errs"
)

// IDataApi 数据查询接口
type IDataApi interface {
	Query(ctx context.Context, req *pb.ReqDataQuery) (*pb.RspDataQuery, error)
	TracePoint(ctx context.Context, mozuId int32, key string) (*pb.RspTracePoint, error)
	QueryLatest(ctx context.Context, req *pb.ReqQueryLatest) (*pb.RspQueryLatest, error)
}

var (
	dataApi     IDataApi
	dataAPiOnce sync.Once
)

// GetDataApi 创建Data相关逻辑接口实现类
func GetDataApi() IDataApi {
	dataAPiOnce.Do(func() {
		dataApi = &dataApiImpl{
			dataCacheProxy:  data_cache.NewPointClientProxy(),
			dataProxy:       dataPb.NewDataClientProxy(),
			deviceEntityDao: db.NewDeviceEntityDao(),
			devicePointDao:  db.NewDevicePointDao(),
			collectorDao:    db.NewCollectorDao(),
		}
	})
	return dataApi
}

type dataApiImpl struct {
	dataCacheProxy  data_cache.PointClientProxy
	dataProxy       dataPb.DataClientProxy
	deviceEntityDao db.IDeviceEntityDao
	devicePointDao  db.IDevicePointDao
	collectorDao    db.ICollectorDao
}

func (d dataApiImpl) QueryLatest(ctx context.Context, req *pb.ReqQueryLatest) (*pb.RspQueryLatest, error) {
	cacheRsp, err := d.dataCacheProxy.QueryData(ctx, &data_cache.ReqQueryData{
		PointList:   req.PointKeys,
		End:         time.Now().Unix(),
		QualityType: data_cache.QualityType_QUALITY_TYPE_ALL,
	})
	if err != nil {
		return nil, errors.Wrap(err, "query data cache failed")
	}
	rsp := &pb.RspQueryLatest{
		Maps: make(map[string]*pb.PointValue),
	}
	for item, data := range cacheRsp.PointMap {
		rsp.Maps[item] = &pb.PointValue{
			Val:     data.Values[0].V,
			Ts:      data.Values[0].T,
			Quality: int32(data.Values[0].Q),
		}
	}
	return rsp, nil
}

type collectorPointInfo struct {
	parentCollector *model.CollectorDevice        // 父级采集设备
	collector       *model.CollectorDevice        // 采集设备信息
	point           *model.CollectorTemplatePoint // 测点信息
}

func (d dataApiImpl) TracePoint(ctx context.Context, mozuId int32, key string) (*pb.RspTracePoint, error) {
	rootPoints, count := d.devicePointDao.GetList(mozuId, &dto.CondGetDevicePointList{PointKey: []string{key}})
	if count != 1 {
		return nil, errs.Newf(errcode.RequestParamError, "bad request param, [%d] record found", count)
	}
	// 保存根测点,即查询的测点
	rootPoint := rootPoints[0]
	// 保存所有的测点标识,包含标准点和采集点
	allPointKeys := []string{rootPoint.PointKey}
	// 保存标准点的map
	stdPointMap := map[string]*model.DevicePoint{rootPoint.PointKey: rootPoint}
	// 下一层的所有测点标识
	subPointKeys := getRefPoint(rootPoint.ExpressionMap)
	// 只要下一层存在,一直往下查找
	for len(subPointKeys) > 0 {
		allPointKeys = append(allPointKeys, subPointKeys...)
		// 查询所有下一层标准点, 查不到则说明是采集点，后面可以过滤出来
		subDevicePoints, _ := d.devicePointDao.GetList(mozuId, &dto.CondGetDevicePointList{PointKey: subPointKeys})
		subPointKeys = lo.Uniq(lo.FlatMap(subDevicePoints, func(item *model.DevicePoint, index int) []string {
			stdPointMap[item.PointKey] = item
			return getRefPoint(item.ExpressionMap)
		}))
	}
	// 过滤出所有的采集测点
	collectorPointKeys := lo.Filter(allPointKeys, func(item string, index int) bool {
		_, ok := stdPointMap[item]
		return !ok
	})
	collectorPointInfoMap, err := d.GetCollectorPointInfo(ctx, mozuId, collectorPointKeys)
	if err != nil {
		return nil, err
	}
	dataRsp, err := d.dataCacheProxy.QueryData(ctx, &data_cache.ReqQueryData{
		PointList:   allPointKeys,
		End:         time.Now().Unix(),
		QualityType: data_cache.QualityType_QUALITY_TYPE_ALL})
	if err != nil {
		return nil, errors.Wrapf(err, "query cache api fail")
	}
	return buildTracePointResult(rootPoint, stdPointMap, collectorPointInfoMap, dataRsp.PointMap), nil
}

func buildTracePointResult(rootPoint *model.DevicePoint, stdPointMap map[string]*model.DevicePoint,
	collectorPointInfoMap map[string]*collectorPointInfo,
	dataMap map[string]*data_cache.RspQueryData_PointValueList) *pb.RspTracePoint {
	// 标准测点，构建响应参数
	rsp := &pb.RspTracePoint{
		StdInfo: &pb.RspTracePoint_StdPoint{},
	}
	_ = copyutil.Copy(rootPoint, rsp)
	_ = copyutil.Copy(rootPoint, rsp.StdInfo)
	rsp.PointType = consts.PointCategoryStd
	if val, ok := dataMap[rootPoint.PointKey]; ok {
		rsp.Value = &pb.PointValue{
			Quality: int32(val.Values[0].Q),
			Val:     val.Values[0].V,
			Ts:      val.Values[0].T,
		}
	}
	varRefPoints := getVarRefPoint(rootPoint.ExpressionMap)
	children := make([]*pb.RspTracePoint, 0, len(varRefPoints))
	for _, varRefPoint := range varRefPoints {
		varName, refPoint := varRefPoint[0], varRefPoint[1]
		if subPoint, ok := stdPointMap[refPoint]; ok {
			rspItem := buildTracePointResult(subPoint, stdPointMap, collectorPointInfoMap, dataMap)
			rspItem.VarName = varName
			children = append(children, rspItem)
		} else if subCollectorPoint, ok := collectorPointInfoMap[refPoint]; ok {
			// 采集测点，已经是叶子结点，直接加入父节点
			rspItem := &pb.RspTracePoint{
				VarName:      varName,
				DeviceGid:    subCollectorPoint.collector.DeviceGid,
				DeviceNumber: subCollectorPoint.collector.DeviceNumber,
				PointKey:     refPoint,
				CollectInfo:  &pb.RspTracePoint_CollectPoint{},
				PointType:    consts.PointCategoryCollect,
			}
			_ = copyutil.Copy(subCollectorPoint.collector, rspItem.CollectInfo)
			if subCollectorPoint.point != nil {
				rspItem.PointNameEn = subCollectorPoint.point.PointNameEn
				rspItem.PointNameZh = subCollectorPoint.point.PointNameZh
				_ = copyutil.Copy(subCollectorPoint.point, rspItem.CollectInfo)
			}
			// 厂商采集器下的设备,模版信息和连接通道信息从厂商采集器信息上取
			if subCollectorPoint.parentCollector != nil {
				rspItem.CollectInfo.TemplateInfo = subCollectorPoint.parentCollector.TemplateInfo
				rspItem.CollectInfo.ChannelLink = subCollectorPoint.parentCollector.ChannelLink
			}
			if val, ok := dataMap[refPoint]; ok {
				rspItem.Value = &pb.PointValue{
					Quality: int32(val.Values[0].Q),
					Val:     val.Values[0].V,
					Ts:      val.Values[0].T,
				}
			}
			children = append(children, rspItem)
		} else {
			// 查询不到的测点
			rspItem := &pb.RspTracePoint{
				VarName:      varName,
				DeviceGid:    "not found",
				DeviceNumber: "not found",
				PointKey:     refPoint,
			}
			children = append(children, rspItem)
		}
	}
	rsp.Children = children
	return rsp
}

// GetCollectorPointInfo 查询采集测点信息
func (d dataApiImpl) GetCollectorPointInfo(ctx context.Context, mozuId int32, collectorPointKeys []string) (map[string]*collectorPointInfo, error) {
	// 取出所有的采集点的gid，并查询出所有设备
	collectorGids := lo.Uniq(lo.Map(collectorPointKeys, func(item string, index int) string {
		return item[0:strings.Index(item, ".")]
	}))
	collectorDevices, _ := d.collectorDao.GetDeviceList(mozuId, &dto.CondCollectorGetDeviceList{DeviceGid: collectorGids})
	collectorDeviceMap := lo.SliceToMap(collectorDevices, func(item *model.CollectorDevice) (string, *model.CollectorDevice) {
		return item.DeviceGid, item
	})
	// 取出所有采集器对应的采集模版
	collectorTemplateMap := make(map[string]*model.CollectorDevice)
	for _, collector := range collectorDevices {
		// 厂商采集器下的子设备需要特殊处理，模版名称在父级设备采集器上
		if collector.CollectorType == model.CollectorTypeVendorSubDevice {
			pList, _ := d.collectorDao.GetDeviceList(mozuId, &dto.CondCollectorGetDeviceList{
				DeviceNumber: []string{collector.ParentDeviceNumber}})
			if len(pList) > 0 {
				collectorTemplateMap[collector.DeviceGid] = pList[0]
			}
		} else {
			collectorTemplateMap[collector.DeviceGid] = collector
		}
	}
	// 取出所有相关的测点查询条件,用于查询采集测点
	collectorConds := make([][]string, 0)
	for _, pointKey := range collectorPointKeys {
		dotPos := strings.Index(pointKey, ".")
		collectorGid, pointName := pointKey[0:dotPos], pointKey[dotPos+1:]
		collectorDevice, ok := collectorDeviceMap[collectorGid]
		if !ok {
			continue
		}
		templateName := collectorTemplateMap[collectorGid].TemplateName
		switch collectorDevice.CollectorType {
		case model.CollectorTypeTbox, model.CollectorTypeVendorBox:
			continue
		case model.CollectorTypeTboxSubDevice:
			collectorConds = append(collectorConds, []string{templateName, "", pointName})
		case model.CollectorTypeVendorSubDevice:
			collectorConds = append(collectorConds, []string{templateName, collectorDevice.DeviceCode, pointName})
		}
	}
	// 查询所有的采集测点信息
	templatePoints, err := d.collectorDao.GetTemplatePointByTriple(ctx, collectorConds)
	if err != nil {
		return nil, errors.Wrapf(err, "fetch template points info from mysql fail")
	}
	templatePointMap := lo.SliceToMap(templatePoints, func(item *model.CollectorTemplatePoint) (string, *model.CollectorTemplatePoint) {
		return item.CalcUniqueKey(), item
	})
	// 生成最终的结果
	res := make(map[string]*collectorPointInfo)
	for _, pointKey := range collectorPointKeys {
		dotPos := strings.Index(pointKey, ".")
		collectorGid, pointName := pointKey[0:dotPos], pointKey[dotPos+1:]
		collectorDevice, ok := collectorDeviceMap[collectorGid]
		if !ok {
			continue
		}
		templateName := collectorTemplateMap[collectorGid].TemplateName
		switch collectorDevice.CollectorType {
		case model.CollectorTypeTbox, model.CollectorTypeVendorBox:
			res[pointKey] = &collectorPointInfo{collector: collectorDevice}
			continue
		case model.CollectorTypeTboxSubDevice:
			infoKey := strings.Join([]string{templateName, "", pointName}, "|")
			res[pointKey] = &collectorPointInfo{collector: collectorDevice, point: templatePointMap[infoKey]}
		case model.CollectorTypeVendorSubDevice:
			infoKey := strings.Join([]string{templateName, collectorDevice.DeviceCode, pointName}, "|")
			res[pointKey] = &collectorPointInfo{parentCollector: collectorTemplateMap[collectorGid],
				collector: collectorDevice, point: templatePointMap[infoKey]}
		}
	}
	return res, nil
}

func getRefPoint(refMap string) []string {
	refPointExp := lo.Filter(strings.Split(strings.Trim(strings.TrimSpace(refMap), ";"), ";"), func(item string, index int) bool {
		return len(item) > 0
	})
	return lo.Uniq(lo.Map(refPointExp, func(item string, index int) string {
		return item[strings.Index(item, "=")+1:]
	}))
}

func getVarRefPoint(refMap string) [][]string {
	refPointExp := lo.Filter(strings.Split(strings.Trim(strings.TrimSpace(refMap), ";"), ";"), func(item string, index int) bool {
		return len(item) > 0
	})
	return lo.Map(refPointExp, func(item string, index int) []string {
		eqPos := strings.Index(item, "=")
		return []string{item[0:eqPos], item[eqPos+1:]}
	})
}

func (d dataApiImpl) Query(ctx context.Context, req *pb.ReqDataQuery) (*pb.RspDataQuery, error) {
	finalRsp := &pb.RspDataQuery{}
	if len(req.Conditions) == 0 {
		return finalRsp, nil
	}
	mozuId := req.MozuId
	// 提取参数
	pointKeyList, pointIdList, deviceGidList, deviceNumberList, pointNameZhs, deviceTypeZhs, applicationTypeZh,
		pointIdEnList, needChill := extractParameters(req)
	if (len(pointKeyList) + len(pointIdList) + len(deviceGidList) + len(deviceNumberList) +
		len(pointNameZhs) + len(deviceTypeZhs) + len(applicationTypeZh) + len(pointIdEnList)) == 0 {
		return finalRsp, nil
	}
	if needChill && len(deviceGidList) != 0 {
		// 将gid转换为设备编号，因为查子设备只支持用设备编号查
		deviceList, total := d.deviceEntityDao.GetList(req.MozuId, &dto.CondDeviceEntityGetList{
			DeviceGid: deviceGidList,
		})
		if total > 0 {
			for _, item := range deviceList {
				deviceNumberList = append(deviceNumberList, item.DeviceNumber)
			}
		}
		deviceNumberList = lo.Uniq(deviceNumberList)
	}
	// 查询设备信息
	deviceMap := make(map[string]*pb.DeviceEntityObject)
	dataQueryRsp := &dataPb.QueryResponse{}
	var allGids []string
	devices, err := queryDeviceInfo(d, ctx, mozuId, deviceGidList, deviceNumberList, deviceTypeZhs,
		applicationTypeZh, needChill)
	if err != nil {
		log.Error(err)
	} else {
		for k, v := range devices {
			deviceMap[k] = v
			allGids = append(allGids, v.DeviceGid)
		}
	}
	if len(devices) == 0 {
		log.Infof("data cgi device not found,%v", req.Conditions)
		return finalRsp, nil
	}
	// 查询测点信息
	pointInfoMap, dataQueryPointKeys, err := queryPointInfo(
		d, mozuId, ctx, pointKeyList, pointIdList, pointIdEnList, allGids, deviceNumberList, pointNameZhs)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if len(pointInfoMap) == 0 && len(dataQueryPointKeys) == 0 {
		return finalRsp, nil
	}
	totalKey := len(dataQueryPointKeys)
	sort.Slice(dataQueryPointKeys, func(i, j int) bool {
		return dataQueryPointKeys[i] < dataQueryPointKeys[j]
	})
	if req.Page > 0 && req.Size > 0 {
		chunkRes := lo.Chunk(dataQueryPointKeys, int(req.Size))
		if int(req.Page) > len(chunkRes) {
			dataQueryPointKeys = []string{}
		} else {
			dataQueryPointKeys = chunkRes[req.Page-1]
		}
	}
	// 查询测点值
	dataQueryRsp, err = queryDataValues(d, ctx, req, dataQueryPointKeys)
	if err != nil {
		// 查不到值，依然进行基础信息的填充
		log.Error(err)
		dataQueryRsp = &dataPb.QueryResponse{
			GetPointData: make(map[string]*dataPb.InnerMap, 0),
		}
	}
	// 数据填充
	begin, _ := stringToDate(req.StartTime)
	end, _ := stringToDate(req.EndTime)
	// todo：空值填补
	if err := d.encodeResponse(dataQueryPointKeys, pointInfoMap, deviceMap, *dataQueryRsp, finalRsp,
		begin, end, int(req.Interval)); err != nil {
		return nil, err
	}
	finalRsp.Total = int32(totalKey)
	return finalRsp, nil
}

func extractParameters(req *pb.ReqDataQuery) ([]string, []string, []string, []string, []string, []string, []string,
	[]string, bool) {
	var pointKeyList []string
	var pointIdList []string
	var deviceGidList []string
	var deviceNumberList []string
	var pointNameZhs []string
	var deviceTypeZhs []string
	var applicationTypeZh []string
	var pointIdEnList []string
	for _, condition := range req.Conditions {
		switch condition.Name {
		case PointKeyConditionName:
			pointKeyList = condition.Value
			for _, pointKey := range pointKeyList {
				gid, err := pointKeyToGid(pointKey)
				if err != nil {
					continue
				}
				deviceGidList = append(deviceGidList, gid)
			}
		case PointIdConditionName:
			pointIdList = condition.Value
			for _, pointId := range pointIdList {
				deviceNumber, err := pointKeyToGid(pointId)
				if err != nil {
					continue
				}
				deviceNumberList = append(deviceNumberList, deviceNumber)
			}
		case PointIdEnConditionName:
			// 编号+测点英文
			pointIdEnList = condition.Value
			for _, pointIdEn := range pointIdEnList {
				deviceNumber, err := pointKeyToGid(pointIdEn)
				if err != nil {
					continue
				}
				deviceNumberList = append(deviceNumberList, deviceNumber)
			}
		case DeviceGidConditionName:
			deviceGidList = append(deviceGidList, condition.Value...)
		case DeviceTypeZhConditionName:
			deviceTypeZhs = condition.Value
		case ApplicationTypeZhConditionName:
			applicationTypeZh = condition.Value
		}
	}
	needChill := req.Cascade
	return pointKeyList, pointIdList, deviceGidList, deviceNumberList, pointNameZhs,
		deviceTypeZhs, applicationTypeZh, pointIdEnList, needChill
}

func queryPointInfo(d dataApiImpl, mozuId int32, ctx context.Context, pointKeyList, pointIdList, pointIdEnList,
	deviceGidList, deviceNumberList, pointNameZhs []string) (map[string]*model.DevicePoint, []string, error) {
	// 查询测点信息
	pointReq, virPointKey := prepareCmdbPointRequest(d, mozuId, pointKeyList, deviceGidList,
		deviceNumberList, pointIdList, pointIdEnList, pointNameZhs)
	rawPointList, _ := d.devicePointDao.GetList(mozuId, pointReq)
	var pointList []*model.DevicePoint
	// 因为查cmdb时，gid和测点中文这两个条件是对结果取并集，所以需要再根据传入的pointid作校验
	if len(pointIdList) > 0 {
		for _, p := range rawPointList {
			if InArray(fmt.Sprintf("%s.%s", p.DeviceNumber, p.PointNameZh), pointIdList) {
				pointList = append(pointList, p)
			}
		}
	} else {
		pointList = rawPointList
	}
	// 处理查询结果
	pointInfoMap := make(map[string]*model.DevicePoint)
	// gid.测点中文
	pointZhMap := make(map[string]interface{})

	// 后续查测点数据时的传参
	var dataQueryPointKeys []string
	for _, pointInfo := range pointList {
		pointInfoMap[pointInfo.PointKey] = pointInfo
		dataQueryPointKeys = append(dataQueryPointKeys, pointInfo.PointKey)
		pointZhMap[fmt.Sprintf("%s.%s", pointInfo.DeviceGid, pointInfo.PointNameZh)] = 1
	}
	// 兼容虚拟点
	for _, vk := range virPointKey {
		if _, ok := pointZhMap[vk]; !ok {
			// 查不到测点配置则为虚拟点
			dataQueryPointKeys = append(dataQueryPointKeys, vk)
		}
	}
	dataQueryPointKeys = lo.Uniq(dataQueryPointKeys)
	return pointInfoMap, dataQueryPointKeys, nil
}

func prepareCmdbPointRequest(d dataApiImpl, mozuId int32, pointKeyList, deviceGidList, deviceNumberList, pointIdList,
	pointIdEnList, pointNameZhs []string) (*dto.CondGetDevicePointList, []string) {
	cmdbPointReq := &dto.CondGetDevicePointList{}
	// 兼容虚拟点: 测点中文在配置侧是查不到的，需要对pointIdList再处理，将其中的编号直接换成gid
	var virPointKey []string
	// pointIdList：JXB-JS-BD02-M401-EVS-CTHS05#BC01.冷通道温度过高
	// pointkey：3689574269601187646.冷通道温度过高
	for _, pointId := range pointIdList {
		list := strings.Split(pointId, GidPointConnectionChar)
		if len(list) != 2 {
			continue
		}
		deviceNumber := list[0]
		pointZh := list[1]
		deviceNumberList = append(deviceNumberList, deviceNumber)
		pointNameZhs = append(pointNameZhs, pointZh)
		// 将编号换为gid
		deviceList, total := d.deviceEntityDao.GetList(mozuId, &dto.CondDeviceEntityGetList{
			DeviceNumber: []string{deviceNumber},
		})
		if total > 0 {
			// 有实际存在的设备，则认为这个目标为虚拟点
			gid := deviceList[0].DeviceGid
			virPointKey = append(virPointKey, fmt.Sprintf("%s.%s", gid, pointZh))
		}
	}
	// 通过pointKey来查的也有可能是虚拟点
	for _, pointKey := range pointKeyList {
		list := strings.Split(pointKey, GidPointConnectionChar)
		if len(list) != 2 {
			continue
		}
		maybePointZh := list[1]
		// todo：临时逻辑如果maybePointZh是中文字符，则认为是虚拟点
		if isAllChinese(maybePointZh) {
			virPointKey = append(virPointKey, pointKey)
		}
	}
	// pointIdEn 设备编号.测点英文转换为pointKey
	for _, pointIdEn := range pointIdEnList {
		list := strings.Split(pointIdEn, GidPointConnectionChar)
		if len(list) != 2 {
			continue
		}
		deviceNumber := list[0]
		pointEn := list[1]
		// 将编号换为gid
		deviceList, total := d.deviceEntityDao.GetList(mozuId, &dto.CondDeviceEntityGetList{
			DeviceNumber: []string{deviceNumber},
		})
		if total > 0 {
			gid := deviceList[0].DeviceGid
			pointKeyList = append(pointKeyList, fmt.Sprintf("%s.%s", gid, pointEn))
		}
	}
	cmdbPointReq.PointKey = pointKeyList
	cmdbPointReq.DeviceGid = deviceGidList
	// 两个条件不能共存，因为是and查询
	if len(deviceGidList) == 0 {
		cmdbPointReq.DeviceNumber = deviceNumberList
	}
	cmdbPointReq.PointNameZh = pointNameZhs

	return cmdbPointReq, virPointKey
}

func queryDeviceInfo(d dataApiImpl, ctx context.Context, mozuId int32, gidList, deviceNumberList []string,
	deviceTypeZhs []string, applicationTypeZh []string, needChill bool) (
	map[string]*pb.DeviceEntityObject, error) {
	reqCond := &dto.CondDeviceEntityGetList{
		DeviceGid:         gidList,
		DeviceNumber:      deviceNumberList,
		ApplicationTypeZh: applicationTypeZh,
		Page:              1,
		Size:              10000,
	}
	deviceList, _ := d.deviceEntityDao.GetList(mozuId, reqCond)
	rspDeviceList := make([]*pb.DeviceEntityObject, 0, len(deviceList))
	for _, item := range deviceList {
		rspItem := &pb.DeviceEntityObject{}
		if err := copyutil.Copy(item, rspItem); err != nil {
			return nil, errors.Wrapf(err, "copy db data to rsp data fail")
		}
		rspDeviceList = append(rspDeviceList, rspItem)
	}

	deviceMap := make(map[string]*pb.DeviceEntityObject)
	for _, device := range rspDeviceList {
		deviceMap[device.DeviceGid] = device
	}
	// 若需要查询下级设备,由于设备侧不支持同时查同级和下级，因此需要再查一次.
	if needChill && len(deviceNumberList) > 0 {
		reqCond := &dto.CondDeviceEntityGetList{
			ParentDeviceNumber: deviceNumberList,
			ApplicationTypeZh:  applicationTypeZh,
			Page:               1,
			Size:               10000,
		}
		deviceList, _ := d.deviceEntityDao.GetList(mozuId, reqCond)
		rspDeviceList := make([]*pb.DeviceEntityObject, 0, len(deviceList))
		for _, item := range deviceList {
			rspItem := &pb.DeviceEntityObject{}
			if err := copyutil.Copy(item, rspItem); err != nil {
				return nil, errors.Wrapf(err, "copy db data to rsp data fail")
			}
			rspDeviceList = append(rspDeviceList, rspItem)
		}

		for _, device := range rspDeviceList {
			deviceMap[device.DeviceGid] = device
		}
	}

	return deviceMap, nil
}

func queryDataValues(d dataApiImpl, ctx context.Context, req *pb.ReqDataQuery, dataQueryPointKeys []string) (*dataPb.QueryResponse, error) {
	dataQueryRsp := &dataPb.QueryResponse{}
	switch req.DataType {
	case pb.ReqDataQuery_CURRENT:
		// 实时
		dataQueryReq := prepareDataQueryRequest(dataQueryPointKeys)
		rsp, err := d.dataProxy.DataQuery(ctx, dataQueryReq)
		if err != nil {
			return nil, fmt.Errorf("request data query api fail: %w", err)
		}
		err = copyutil.Copy(rsp, dataQueryRsp)
		if err != nil {
			return nil, errors.Wrapf(err, "copy req param to cmdb get tree req param fail")
		}
	case pb.ReqDataQuery_HISTORY:
		// 历史
		dataQueryReq := prepareHistoryDataQueryRequest(req, dataQueryPointKeys)
		rsp, err := d.dataProxy.DataQuery(ctx, dataQueryReq)
		if err != nil {
			return nil, fmt.Errorf("request data query api fail: %w", err)
		}
		err = copyutil.Copy(rsp, dataQueryRsp)
		if err != nil {
			return nil, errors.Wrapf(err, "copy req param to cmdb get tree req param fail")
		}
	default:
	}

	return dataQueryRsp, nil
}

func prepareDataQueryRequest(dataQueryPointKeys []string) *dataPb.QueryRequest {
	return &dataPb.QueryRequest{
		PointList: dataQueryPointKeys,
		Begin:     0,
		End:       time.Now().Unix(),
	}
}
func prepareHistoryDataQueryRequest(req *pb.ReqDataQuery, dataQueryPointKeys []string) *dataPb.QueryRequest {
	begin, _ := stringToDate(req.StartTime)
	end, _ := stringToDate(req.EndTime)

	return &dataPb.QueryRequest{
		PointList: dataQueryPointKeys,
		Begin:     begin,
		End:       end,
		Interval:  int64(req.Interval),
	}
}

// 数据组合
// todo：空值填补
func (d dataApiImpl) encodeResponse(pointKeys []string, pointInfoMap map[string]*model.DevicePoint,
	deviceMap map[string]*pb.DeviceEntityObject, dataQueryRsp dataPb.QueryResponse,
	rsp *pb.RspDataQuery, begin int64, end int64, interval int) error {
	for _, pointKey := range pointKeys {
		pointInfo, ok := pointInfoMap[pointKey]
		if !ok {
			// 若这个key查到了数据，但是没有测点信息，说明是虚拟点,需要构造一个测点信息
			_, isVirtual := dataQueryRsp.GetPointData[pointKey]
			if !isVirtual {
				log.Infof("data cgi pointInfo not found, pointKey:%s", pointKey)
				continue
			} else {
				list := strings.Split(pointKey, GidPointConnectionChar)
				if len(list) != 2 {
					continue
				}
				gid := list[0]
				pointZh := list[1]
				pointInfo = &model.DevicePoint{
					DeviceGid:   gid,
					PointNameEn: pointZh,
					PointNameZh: pointZh,
					PointKey:    pointKey,
				}
			}
		}
		pointObj, err := d.createPointObj(pointKey, pointInfo, deviceMap, dataQueryRsp, begin, end, interval)
		if err != nil {
			log.Errorf("create point obj fail, pointKey:%s, err:%s", pointKey, err.Error())
			continue
		}
		rsp.List = append(rsp.List, pointObj)
	}
	return nil
}

// todo：空值填补
// 创建单个测点信息
func (d dataApiImpl) createPointObj(pointKey string, pointInfo *model.DevicePoint,
	deviceMap map[string]*pb.DeviceEntityObject, dataQueryRsp dataPb.QueryResponse,
	begin int64, end int64, interval int) (*pb.PointObj, error) {

	pointObj := &pb.PointObj{
		PointKey: pointKey,
		//DeviceNumber: pointInfo.DeviceNumber,
		DeviceGid:   pointInfo.DeviceGid,
		PointNameZh: pointInfo.PointNameZh,
		PointNameEn: pointInfo.PointNameEn,
		// todo q这里未做处理
		Q:            "0",
		Unit:         pointInfo.ValueUnit,
		Status:       d.convertValueTypeToStatus(pointInfo.ValueType),
		ReadAndWrite: d.convertPointRwToPointRW(pointInfo.PointRw),
		Simulation:   d.convertValueTypeToSimulation(pointInfo.ValueType),
		ValueType:    d.convertValueType(pointInfo.ValueType),
	}
	gid, _ := pointKeyToGid(pointKey)
	deviceInfo, ok := deviceMap[gid]
	if !ok {
		log.Warnf("device not found for pointKey: %s", pointKey)
		return nil, fmt.Errorf("device not found for pointKey: %s", pointKey)
	}
	pointObj.PointId = fmt.Sprintf("%s.%s", deviceInfo.DeviceNumber, pointInfo.PointNameZh)
	pointObj.DeviceNumber = deviceInfo.DeviceNumber
	pointObj.DeviceTypeEn = deviceInfo.DeviceTypeEn
	pointObj.DeviceTypeZh = deviceInfo.DeviceTypeZh
	pointObj.ApplicationTypeEn = deviceInfo.ApplicationTypeEn
	pointObj.ApplicationTypeZh = deviceInfo.ApplicationTypeZh
	// todo：空值填补
	pointDatas, updateTime, latestValue, statsDataMap, err := d.createPointDatas(pointKey, dataQueryRsp,
		begin, end, interval)
	if err != nil {
		return nil, err
	}
	// 枚举值替换
	var statusValue string
	if pointObj.Status == true {
		statusMap := parseEnumDefinition(pointInfo.ValueEnum)
		if statusMap != nil {
			// 对latestValue进行规整, 转换为1、0等
			v := formatFloatString(latestValue)
			statusValue, ok = statusMap[v]
			if !ok {
				statusValue = v
			}
		}
	}
	// 统计值转换
	var stats []*pb.Stats
	for k, v := range statsDataMap {
		stats = append(stats, &pb.Stats{
			Name:  k,
			Value: v,
		})
	}
	pointObj.PointData = pointDatas
	pointObj.UpdateTime = updateTime
	pointObj.LatestValue = latestValue
	pointObj.Stats = stats
	pointObj.EnumValue = statusValue
	return pointObj, nil
}

// 是否状态量
func (d dataApiImpl) convertValueTypeToStatus(valueType string) bool {
	switch valueType {
	case "enum", "bool":
		return true
	default:
		return false
	}
}

// 是否模拟量
func (d dataApiImpl) convertValueTypeToSimulation(valueType string) bool {
	if valueType == "float" {
		return true
	}
	return false
}

// 是否权限为读写
func (d dataApiImpl) convertPointRwToPointRW(pointRw string) bool {
	if pointRw == "读写" {
		return true
	}
	return false
}

// 将类型转换为之前的定义 -1:错误,-2:后台模拟值,1:double, 2:boolean, 3:long, 4:other(字符串)
func (d dataApiImpl) convertValueType(valueType string) int32 {
	switch valueType {
	case "float":
		return 1
	case "bool":
		return 2
	case "enum":
		return 4
	default:
		return 1
	}
}

// 测点值处理
func (d dataApiImpl) createPointDatas(pointKey string, dataQueryRsp dataPb.QueryResponse,
	begin int64, end int64, interval int) ([]*pb.PointData, string, string,
	map[string]float64, error) {
	var pointDatas []*pb.PointData
	var updateTime, latestValue string
	var statsDataMap map[string]float64
	if dataQueryRsp.GetPointData == nil {
		pointDatas = d.createDefaultPointData(begin, end, interval)
		updateTime = DefaultPointValue
		latestValue = DefaultPointValue
	} else {
		pointData, ok := dataQueryRsp.GetPointData[pointKey]
		if !ok {
			pointDatas = d.createDefaultPointData(begin, end, interval)
			updateTime = DefaultPointValue
			latestValue = DefaultPointValue
		} else {
			var err error
			pointDatas, updateTime, latestValue, statsDataMap, err = d.processPointData(
				pointData, begin, end, interval)
			if err != nil {
				return nil, "", "", nil, err
			}
		}
	}

	return pointDatas, updateTime, latestValue, statsDataMap, nil
}

// 缺省值
func (d dataApiImpl) createDefaultPointData(begin int64, end int64, interval int) []*pb.PointData {
	var pointDatas []*pb.PointData
	timeKeys := generateTimeSeries(begin, end, interval)
	for _, t := range timeKeys {
		timeStr, err := dataToString(t)
		if err != nil {
			continue
		}
		pointDatas = append(pointDatas, &pb.PointData{
			UpdateTime: timeStr,
			Value:      DefaultPointValue,
		})
	}
	return pointDatas
}

// 时序数据处理
func (d dataApiImpl) processPointData(pointData *dataPb.InnerMap, begin int64, end int64, interval int) ([]*pb.PointData, string, string, map[string]float64, error) {
	var rawkeys []int64
	// 由于查不到的值也需要填充
	timeKeys := generateTimeSeries(begin, end, interval)
	for t := range pointData.InnerMap {
		rawkeys = append(rawkeys, t)
	}
	if len(timeKeys) == 0 && len(rawkeys) > 0 {
		// 取实时数据（单个点）
		timeKeys = append(timeKeys, rawkeys[len(rawkeys)-1])
	}
	// 根据排序后的时间戳顺序处理数据
	//sort.Slice(keys, func(i, j int) bool {
	//	return keys[i] < keys[j]
	//})

	var pointDatas []*pb.PointData
	var updateTime string
	latestValue := DefaultPointValue

	// 初始化统计变量
	var sumVal, maxVal, minVal float64
	count := len(rawkeys)
	if count == 0 {
		return nil, "", "", nil, fmt.Errorf("no data points")
	}

	for i, t := range timeKeys {
		v, ok := pointData.InnerMap[t]
		if !ok {
			// 取不到值填"--'
			timeStr, err := dataToString(t)
			if err != nil {
				return nil, "", "", nil, fmt.Errorf("time conversion failed: %w", err)
			}
			data := &pb.PointData{
				UpdateTime: timeStr,
				Value:      DefaultPointValue,
			}
			pointDatas = append(pointDatas, data)
			if i == len(timeKeys)-1 {
				updateTime = timeStr
				latestValue = DefaultPointValue
			}
			continue
		}
		// 转换v为float64
		val, err := convertToFloat64(v)
		if err != nil {
			return nil, "", "", nil, fmt.Errorf("value conversion failed: %w", err)
		}
		// 更新统计值
		if i == 0 {
			maxVal = val
			minVal = val
			sumVal = val
		} else {
			sumVal += val
			if val > maxVal {
				maxVal = val
			}
			if val < minVal {
				minVal = val
			}
		}

		timeStr, err := dataToString(t)
		if err != nil {
			return nil, "", "", nil, fmt.Errorf("time conversion failed: %w", err)
		}
		data := &pb.PointData{
			UpdateTime: timeStr,
			Value:      formatValue(v),
		}
		pointDatas = append(pointDatas, data)
		if i == len(timeKeys)-1 {
			updateTime = timeStr
			latestValue = formatValue(v)
		}
	}
	// 计算平均值
	avgVal := sumVal / float64(count)

	// 创建并填充统计结果
	statsDataMap := make(map[string]float64)
	statsDataMap["count"] = float64(count)
	statsDataMap["max"] = maxVal
	statsDataMap["min"] = minVal
	statsDataMap["sum"] = sumVal
	statsDataMap["avg"] = avgVal

	return pointDatas, updateTime, latestValue, statsDataMap, nil
}

// 生成时间序列
func generateTimeSeries(begin, end int64, interval int) []int64 {
	var result []int64
	if begin >= end || interval <= 0 {
		return result
	}
	step := int64(interval)
	for current := begin; current <= end; current += step {
		result = append(result, current)
	}
	return result
}

// 辅助函数：将interface{}转换为float64
func convertToFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("unsupported value type: %T", v)
	}
}

// 测点值格式化
func formatValue(v float64) string {
	// 根据配置格式化值的逻辑
	//todo：科学计数法数字转换，保留两位小数  singleData.put("value", EEnumUtils.eEnumFormat2normal(mibValue));
	return fmt.Sprintf("%.2f", v) // 示例：固定两位小数
}

// 字符串处理：将float转为整形
func formatFloatString(s string) string {
	// 尝试将字符串解析为浮点数
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// 如果解析失败，返回原字符串
		return s
	}
	// 检查浮点数是否为整数
	if f == float64(int(f)) {
		// 如果是整数，返回整数形式的字符串
		return strconv.Itoa(int(f))
	}
	// 否则返回原字符串
	return s
}

// 解析枚举map
func parseEnumDefinition(enumDef string) map[string]string {
	statusMap := make(map[string]string)

	// 首先按逗号分割字符串
	pairs := strings.Split(enumDef, ",")

	for _, pair := range pairs {
		// 去除可能的空格
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		// 按等号分割键值对
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		// 将键值对存入map
		statusMap[key] = value
	}
	return statusMap
}
