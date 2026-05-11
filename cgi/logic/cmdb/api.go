// Package cmdb 配置相关接口实现逻辑
package cmdb

import (
	"cgi/entity/dto"
	"cgi/logic/util"
	"cgi/repo/db"
	"common/entity/consts"
	"common/entity/model"
	"context"
	"etrpc-go/util/copyutil"
	"fmt"
	"sort"
	"strings"
	pb "trpcprotocol/cgi"
	idc_tbos_data_cache "trpcprotocol/data-cache"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// ICmdbApi Cmdb相关逻辑接口
type ICmdbApi interface {
	GetDeviceTree(ctx context.Context, req *pb.ReqGetDeviceTree) (*pb.RspGetDeviceTree, error)
	GetDevicePoint(ctx context.Context, req *pb.ReqGetDevicePoint) (*pb.RspGetDevicePoint, error)
	GetSubTree(ctx context.Context, req *pb.ReqGetSubTree) (*pb.RspGetSubTree, error)
	GetSubTreeFieldDic(ctx context.Context, req *pb.ReqGetSubTreeFieldDic) (*pb.RspCommonGetKeyDict, error)
	GetMozuInfo(ctx context.Context, req *pb.ReqGetMozuInfo) (*pb.RspGetMozuInfo, error)
	GetDeviceEntity(ctx context.Context, req *pb.ReqGetDeviceEntity) (*pb.RspGetDeviceEntity, error)
	GetCollectorStatusTree(ctx context.Context, req *pb.ReqGetCollectorStatusTree) (*pb.RspGetCollectorStatusTree, error)
	GetCollectorInfo(ctx context.Context, req *pb.ReqGetCollectorInfo) (*pb.RspGetCollectorInfo, error)
	GetCollectorPoint(ctx context.Context, req *pb.ReqGetCollectorPoint) (*pb.RspGetCollectorPoint, error)
}

// NewCmdbApi 创建Cmdb相关逻辑接口实现类
func NewCmdbApi() ICmdbApi {
	return &cmdbApi{
		deviceEntityDao:  db.NewDeviceEntityDao(),
		devicePointDao:   db.NewDevicePointDao(),
		mozuInfoDao:      db.NewMozuInfoDao(),
		collectorDao:     db.NewCollectorDao(),
		alarmStrategyDao: db.NewAlarmStrategyDao(),
		dataCacheCli:     idc_tbos_data_cache.NewPointClientProxy(),
	}
}

type cmdbApi struct {
	deviceEntityDao  db.IDeviceEntityDao
	devicePointDao   db.IDevicePointDao
	mozuInfoDao      db.IMozuInfoDao
	collectorDao     db.ICollectorDao
	alarmStrategyDao db.IAlarmStrategyDao
	dataCacheCli     idc_tbos_data_cache.PointClientProxy
}

func (obj *cmdbApi) GetDeviceEntity(ctx context.Context, req *pb.ReqGetDeviceEntity) (*pb.RspGetDeviceEntity, error) {
	reqCond := &dto.CondDeviceEntityGetList{}
	if err := copyutil.Copy(req, reqCond); err != nil {
		return nil, errors.Wrapf(err, "copy req param to db cond fail")
	}
	deviceList, total := obj.deviceEntityDao.GetList(req.MozuId, reqCond)
	rspDeviceList := make([]*pb.DeviceEntityObject, 0, len(deviceList))
	for _, item := range deviceList {
		rspItem := &pb.DeviceEntityObject{}
		if err := copyutil.Copy(item, rspItem); err != nil {
			return nil, errors.Wrapf(err, "copy db data to rsp data fail")
		}
		rspDeviceList = append(rspDeviceList, rspItem)
	}
	return &pb.RspGetDeviceEntity{
		List:  rspDeviceList,
		Total: int32(total),
	}, nil
}

func (obj *cmdbApi) GetDeviceTree(ctx context.Context, req *pb.ReqGetDeviceTree) (*pb.RspGetDeviceTree, error) {
	// 1、获取所有的设备
	deviceList, _ := obj.deviceEntityDao.GetList(req.MozuId, nil)
	// 2、组装所有树节点
	_, nodeMap := constructNodes(deviceList)
	// 3、组装树层级结构
	rootNodes := buildTreeStructure(req, deviceList, nodeMap)
	for _, item := range rootNodes {
		buildTreeSubDeviceCount(item)
	}
	return &pb.RspGetDeviceTree{List: rootNodes}, nil
}

func (obj *cmdbApi) GetSubTree(ctx context.Context, req *pb.ReqGetSubTree) (*pb.RspGetSubTree, error) {
	tree, err := obj.getSubDeviceTree(req.MozuId, req.DeviceGid, req.DeviceNumber)
	if err != nil {
		return nil, err
	}
	rsp := &pb.RspGetSubTree_TreeNode{}
	_ = copyutil.Copy(tree, rsp)
	return &pb.RspGetSubTree{
		Root: rsp,
	}, nil
}

func constructNodes(deviceList []*model.DeviceEntity) ([]*pb.RspGetDeviceTree_TreeNode,
	map[string]*pb.RspGetDeviceTree_TreeNode) {
	deviceTreeNodes := make([]*pb.RspGetDeviceTree_TreeNode, 0, len(deviceList))
	nodeMap := make(map[string]*pb.RspGetDeviceTree_TreeNode)

	for _, device := range deviceList {
		node := &pb.RspGetDeviceTree_TreeNode{}
		_ = copyutil.Copy(device, node)
		node.Children = make([]*pb.RspGetDeviceTree_TreeNode, 0)
		deviceTreeNodes = append(deviceTreeNodes, node)
		nodeMap[node.DeviceNumber] = node
	}

	return deviceTreeNodes, nodeMap
}

func buildTreeStructure(req *pb.ReqGetDeviceTree, deviceList []*model.DeviceEntity,
	nodeMap map[string]*pb.RspGetDeviceTree_TreeNode) []*pb.RspGetDeviceTree_TreeNode {
	var rootNodes []*pb.RspGetDeviceTree_TreeNode
	if req.TreeType == pb.ReqGetDeviceTree_DEVICE_TYPE {
		// 筛选出所有的zone设备并转化为map
		zoneDevices := lo.Filter(deviceList, func(item *model.DeviceEntity, index int) bool {
			return strings.EqualFold(strings.ToUpper(item.BelongApplicationTypeEn), "ZONE")
		})
		zoneDeviceMap := lo.SliceToMap(zoneDevices, func(item *model.DeviceEntity) (string, *model.DeviceEntity) {
			return item.IdcArea, item
		})
		// 按应用类型进行分组
		applicationTypeMap := lo.GroupBy(deviceList, func(item *model.DeviceEntity) string {
			return item.ApplicationTypeEn
		})
		for applicationType, leafs := range applicationTypeMap {
			// 按zone区域进行分组
			zoneApplicationTypeDeviceMap := lo.GroupBy(lo.Filter(leafs, func(item *model.DeviceEntity, index int) bool {
				return len(item.IdcArea) > 0
			}), func(item *model.DeviceEntity) string {
				return item.IdcArea
			})
			// 生成根节点
			first := leafs[0]
			root := &pb.RspGetDeviceTree_TreeNode{
				DeviceGid:               applicationType,
				DeviceNumber:            applicationType,
				DeviceName:              applicationType,
				EnableStatus:            1,
				ApplicationTypeEn:       applicationType,
				ApplicationTypeZh:       first.ApplicationTypeZh,
				BelongApplicationTypeEn: first.BelongApplicationTypeEn,
				Children:                make([]*pb.RspGetDeviceTree_TreeNode, 0, len(zoneApplicationTypeDeviceMap)),
			}
			rootNodes = append(rootNodes, root)
			// 处理各个区域的设备
			for idcArea, curZoneDevices := range zoneApplicationTypeDeviceMap {
				areaNode := &pb.RspGetDeviceTree_TreeNode{
					DeviceGid:    idcArea,
					DeviceNumber: idcArea,
					DeviceName:   idcArea,
					EnableStatus: 1,
					Children:     make([]*pb.RspGetDeviceTree_TreeNode, 0, len(curZoneDevices)),
				}
				if zoneDevice, ok := zoneDeviceMap[idcArea]; ok {
					areaNode.ApplicationTypeEn = zoneDevice.ApplicationTypeEn
					areaNode.ApplicationTypeZh = zoneDevice.ApplicationTypeZh
					areaNode.BelongApplicationTypeEn = zoneDevice.BelongApplicationTypeEn
				} else {
					areaNode.DeviceTypeEn = "Other"
					areaNode.DeviceTypeZh = "其他区域"
					areaNode.ApplicationTypeEn = "Other"
					areaNode.ApplicationTypeZh = "其他区域"
				}
				// 将区域内的设备加到区域节点
				for _, item := range curZoneDevices {
					areaNode.Children = append(areaNode.Children, nodeMap[item.DeviceNumber])
				}
				root.Children = append(root.Children, areaNode)
			}
		}
	} else {
		for _, device := range deviceList {
			// 根节点
			if device.ParentDeviceNumber == "" {
				rootNodes = append(rootNodes, nodeMap[device.DeviceNumber])
			}
			if parent, ok := nodeMap[device.ParentDeviceNumber]; ok {
				parent.Children = append(parent.Children, nodeMap[device.DeviceNumber])
			}
		}
	}
	return rootNodes
}

func buildTreeSubDeviceCount(tree *pb.RspGetDeviceTree_TreeNode) int32 {
	if len(tree.Children) == 0 {
		tree.DeviceCount = 0
		return 1
	}
	var totalCnt int32
	for _, child := range tree.Children {
		totalCnt += buildTreeSubDeviceCount(child)
	}
	tree.DeviceCount = totalCnt
	return totalCnt
}

func (obj *cmdbApi) GetDevicePoint(ctx context.Context, req *pb.ReqGetDevicePoint) (*pb.RspGetDevicePoint, error) {
	rsp := &pb.RspGetDevicePoint{}
	if len(req.DeviceTypeZh) > 0 || len(req.DeviceTypeEn) > 0 ||
		len(req.ApplicationTypeZh) > 0 || len(req.ApplicationTypeEn) > 0 {
		deviceList, total := obj.deviceEntityDao.GetList(req.MozuId, &dto.CondDeviceEntityGetList{
			DeviceGid:         req.DeviceGid,
			DeviceNumber:      req.DeviceNumber,
			ApplicationTypeEn: req.ApplicationTypeEn,
			ApplicationTypeZh: req.ApplicationTypeZh,
		})
		if total > 0 {
			gids := make([]string, 0, len(deviceList))
			for _, item := range deviceList {
				gids = append(gids, item.DeviceGid)
			}
			req.DeviceGid = gids
			req.DeviceNumber = nil
		} else {
			return rsp, nil
		}
	}
	pointReq := &dto.CondGetDevicePointList{}
	_ = copyutil.Copy(req, pointReq)
	pointList, total := obj.devicePointDao.GetList(req.MozuId, pointReq)
	rsp.List = lo.Map(pointList, func(item *model.DevicePoint, index int) *pb.DevicePointObject {
		rspItem := &pb.DevicePointObject{}
		_ = copyutil.Copy(item, rspItem)
		return rspItem
	})
	rsp.Total = int32(total)

	return rsp, nil
}

func (obj *cmdbApi) getSubDeviceTree(mozuId int32, deviceGid, deviceNUmber string) (*dto.DeviceTreeNode, error) {
	daoReq := &dto.CondDeviceEntityGetList{}
	if deviceGid != "" {
		daoReq.DeviceGid = []string{deviceGid}
	}
	if deviceNUmber != "" {
		daoReq.DeviceNumber = []string{deviceNUmber}
	}
	list, total := obj.deviceEntityDao.GetList(mozuId, daoReq)
	if total != 1 {
		return nil, fmt.Errorf("[%d] record found, bad request param", total)
	}
	rootDevice := list[0]
	allDevices := []*model.DeviceEntity{rootDevice}
	parentDeviceNumbers := []string{rootDevice.DeviceNumber}
	for len(parentDeviceNumbers) > 0 {
		subDeviceDaoReq := &dto.CondDeviceEntityGetList{
			ParentDeviceNumber: parentDeviceNumbers,
		}
		subDevices, totalSub := obj.deviceEntityDao.GetList(mozuId, subDeviceDaoReq)
		if totalSub == 0 {
			break
		}
		parentDeviceNumbers = lo.Uniq(lo.Map(subDevices, func(item *model.DeviceEntity, index int) string {
			return item.DeviceNumber
		}))
		allDevices = append(allDevices, subDevices...)
	}
	nodeMap := make(map[string]*dto.DeviceTreeNode)
	// 2、组装所有树节点
	deviceTreeNodes := make([]*dto.DeviceTreeNode, 0, len(allDevices))
	for _, device := range allDevices {
		node := &dto.DeviceTreeNode{}
		_ = copyutil.Copy(device, node)
		node.Children = make([]*dto.DeviceTreeNode, 0)
		deviceTreeNodes = append(deviceTreeNodes, node)
		nodeMap[node.DeviceNumber] = node
	}
	var rootNode *dto.DeviceTreeNode
	for _, device := range allDevices {
		if device.DeviceGid == rootDevice.DeviceGid {
			rootNode = nodeMap[device.DeviceNumber]
		}
		if parent, ok := nodeMap[device.ParentDeviceNumber]; ok {
			parent.Children = append(parent.Children, nodeMap[device.DeviceNumber])
		}
	}
	return rootNode, nil
}

type getTreeNodeFieldFunc = func(node *dto.DeviceTreeNode) string

var filedFuncMap = map[string]getTreeNodeFieldFunc{
	"application_type_en": func(node *dto.DeviceTreeNode) string { return node.ApplicationTypeEn },
	"application_type_zh": func(node *dto.DeviceTreeNode) string { return node.ApplicationTypeZh },
	"device_type_en":      func(node *dto.DeviceTreeNode) string { return node.DeviceTypeZh },
	"device_type_zh":      func(node *dto.DeviceTreeNode) string { return node.DeviceTypeZh },
	"device_gid":          func(node *dto.DeviceTreeNode) string { return node.DeviceGid },
	"device_number":       func(node *dto.DeviceTreeNode) string { return node.DeviceNumber },
}

func getSubTreeFieldDic(root *dto.DeviceTreeNode, getField func(node *dto.DeviceTreeNode) string) []string {
	res := []string{getField(root)}
	for _, item := range root.Children {
		res = append(res, getSubTreeFieldDic(item, getField)...)
	}
	return res
}

func (obj *cmdbApi) GetSubTreeFieldDic(ctx context.Context, req *pb.ReqGetSubTreeFieldDic) (
	*pb.RspCommonGetKeyDict, error) {
	fieldFunc := filedFuncMap[req.FiledType]
	if fieldFunc == nil {
		return nil, fmt.Errorf("field_tye [%s] not support", req.FiledType)
	}
	tree, err := obj.getSubDeviceTree(req.MozuId, req.DeviceGid, req.DeviceNumber)
	if err != nil {
		return nil, errors.Wrapf(err, "get sub tree fail")
	}
	dicList := getSubTreeFieldDic(tree, fieldFunc)
	dicList = lo.Uniq(lo.Filter(dicList, func(item string, index int) bool {
		return len(item) > 0
	}))
	return &pb.RspCommonGetKeyDict{
		List: dicList,
	}, nil
}

func (obj *cmdbApi) GetMozuInfo(ctx context.Context, req *pb.ReqGetMozuInfo) (*pb.RspGetMozuInfo, error) {
	condReq := &dto.CondMozuInfoGetList{
		MozuId:   req.MozuId,
		MozuName: req.MozuName,
		MozuCode: req.MozuCode,
	}
	mozuList, err := obj.mozuInfoDao.GetList(ctx, condReq)
	if err != nil {
		return nil, err
	}
	rspList := lo.Map(mozuList, func(item *model.MozuInfo, index int) *pb.RspGetMozuInfo_MozuInfo {
		rspItem := &pb.RspGetMozuInfo_MozuInfo{}
		_ = copyutil.Copy(item, rspItem)
		return rspItem
	})
	return &pb.RspGetMozuInfo{
		List:  rspList,
		Total: int32(len(rspList)),
	}, nil
}

func (obj *cmdbApi) GetCollectorStatusTree(ctx context.Context, req *pb.ReqGetCollectorStatusTree) (*pb.RspGetCollectorStatusTree, error) {
	// 构建查询条件查询所有相关的采集设备
	condDao := &dto.CondCollectorGetDeviceList{}
	switch req.CollectorType {
	case model.CollectorTypeTbox:
		condDao.CollectorType = []int32{model.CollectorTypeTbox, model.CollectorTypeTboxSubDevice}
	case model.CollectorTypeVendorBox:
		condDao.CollectorType = []int32{model.CollectorTypeVendorBox, model.CollectorTypeVendorSubDevice}
	case model.CollectorTypeDoor:
		condDao.CollectorType = []int32{model.CollectorTypeDoor, model.CollectorTypeDoorSubDevice}
	case model.CollectorTypeTone:
		condDao.CollectorType = []int32{model.CollectorTypeTone, model.CollectorTypeToneSubDevice}
	}
	collectorDevices, _ := obj.collectorDao.GetDeviceList(req.MozuId, condDao)
	// 组装设备树
	nodeMap := make(map[string]*pb.RspGetCollectorStatusTree, len(collectorDevices))
	for _, item := range collectorDevices {
		rspItem := &pb.RspGetCollectorStatusTree{}
		_ = copyutil.Copy(item, rspItem)
		rspItem.Children = make([]*pb.RspGetCollectorStatusTree, 0)
		rspItem.CommStateId = getCollectorStatusKey(item)
		// 处理下设备名称
		if item.DeviceName == "" {
			if idx := strings.LastIndex(item.DeviceCode, "_"); idx > 0 {
				rspItem.DeviceName = fmt.Sprintf("%s%s", item.DeviceTypeZh, item.DeviceCode[idx+1:])
			}
		}
		nodeMap[rspItem.DeviceNumber] = rspItem
	}
	// 组装设备树
	collectorNodes := make([]*pb.RspGetCollectorStatusTree, 0)
	for _, item := range collectorDevices {
		if item.ParentDeviceNumber == "" {
			collectorNodes = append(collectorNodes, nodeMap[item.DeviceNumber])
		} else {
			if node, ok := nodeMap[item.ParentDeviceNumber]; ok {
				node.Children = append(node.Children, nodeMap[item.DeviceNumber])
			}
		}
	}
	// 填充设备树节点的子节点数量
	for _, node := range nodeMap {
		node.DeviceCount = int32(len(node.Children))
	}
	sort.Slice(collectorNodes, func(i, j int) bool {
		return collectorNodes[i].DeviceNumber < collectorNodes[j].DeviceNumber
	})
	rsp := &pb.RspGetCollectorStatusTree{
		DeviceGid:   fmt.Sprint(req.MozuId),
		Children:    collectorNodes,
		DeviceCount: int32(len(collectorNodes)),
	}
	// 填充模组中文名称
	list, err := obj.mozuInfoDao.GetList(ctx, &dto.CondMozuInfoGetList{MozuId: []int32{req.MozuId}})
	if err == nil && len(list) == 1 {
		rsp.DeviceNumber = list[0].MozuCode
		rsp.DeviceName = list[0].MozuName
	}
	return rsp, nil
}

func (obj *cmdbApi) GetCollectorInfo(ctx context.Context, req *pb.ReqGetCollectorInfo) (*pb.RspGetCollectorInfo, error) {
	collectors, cnt := obj.collectorDao.GetDeviceList(req.MozuId, &dto.CondCollectorGetDeviceList{
		DeviceGid: []string{req.DeviceGid},
	})
	if cnt != 1 {
		return nil, fmt.Errorf("bad request param, [%d] record found", cnt)
	}
	// 将设备信息转换为rsp结构
	collector := collectors[0]
	rsp := &pb.RspGetCollectorInfo{}
	_ = copyutil.Copy(collector, rsp)
	rsp.ChannelLink = util.JsonStrToPbStruct(collector.ChannelLink)
	rsp.TemplateInfo = util.JsonStrToPbStruct(collector.TemplateInfo)
	rsp.StatusId = getCollectorStatusKey(collector)

	// 如果是Tbox采集器或者厂商采集器,则获取子设备信息
	if collector.CollectorType == model.CollectorTypeTbox ||
		collector.CollectorType == model.CollectorTypeVendorBox ||
		collector.CollectorType == model.CollectorTypeDoor ||
		collector.CollectorType == model.CollectorTypeTone {
		subDevices, _ := obj.collectorDao.GetDeviceList(req.MozuId, &dto.CondCollectorGetDeviceList{
			ParentDeviceNumber: []string{collector.DeviceNumber},
		})
		rspSubDevices := make([]*pb.RspGetCollectorInfo, 0)
		for _, item := range subDevices {
			rspItem := &pb.RspGetCollectorInfo{}
			_ = copyutil.Copy(item, rspItem)
			rspItem.StatusId = getCollectorStatusKey(item)
			rspItem.ChannelLink = util.JsonStrToPbStruct(item.ChannelLink)
			rspItem.TemplateInfo = util.JsonStrToPbStruct(item.TemplateInfo)
			if len(item.DeviceName) == 0 {
				rspItem.DeviceName = item.DeviceGid
			}
			rspSubDevices = append(rspSubDevices, rspItem)
		}
		rsp.Devices = rspSubDevices
		if collector.CollectorType == model.CollectorTypeTbox {
			// Tbox状态信息，暂时常量定义
			stateInfo := map[string][]string{
				"power": {".PowerFault_0", ".PowerFault_1"},
				"di": {".DI0", ".DI1", ".DI2", ".DI3", ".DI4", ".DI5", ".DI6",
					".DI7", ".DI8", ".DI9", ".DI10", ".DI11",
				},
				"do": {".DO0", ".DO1", ".DO2", ".DO3"},
				"com": {".COM0", ".COM1", ".COM2", ".COM3", ".COM4", ".COM5",
					".COM6", ".COM7", ".COM8", ".COM9"},
			}
			rsp.State, _ = copyutil.ConvertToStruct(stateInfo)
		}
	} else {
		metric := map[string]any{
			"total_request_id": fmt.Sprintf("%s.total_req", collector.DeviceGid),
			"metrics": []map[string]string{
				{
					"id":   fmt.Sprintf("%s.total_req", collector.DeviceGid),
					"name": "总请求数",
				},
				{
					"id":   fmt.Sprintf("%s.success_req", collector.DeviceGid),
					"name": "成功请求数",
				},
				{
					"id":   fmt.Sprintf("%s.point_throughput", collector.DeviceGid),
					"name": "每秒采集测点数",
				},
			},
		}
		if len(rsp.DeviceName) == 0 {
			rsp.DeviceName = rsp.DeviceCode
		}
		rsp.State, _ = copyutil.ConvertToStruct(metric)
	}
	return rsp, nil
}

func (obj *cmdbApi) GetCollectorPoint(ctx context.Context, req *pb.ReqGetCollectorPoint) (*pb.RspGetCollectorPoint, error) {
	collectors, cnt := obj.collectorDao.GetDeviceList(req.MozuId, &dto.CondCollectorGetDeviceList{
		DeviceGid: []string{req.DeviceGid},
	})
	if cnt != 1 {
		return nil, fmt.Errorf("bad request param, [%d] record found", cnt)
	}
	collector := collectors[0]
	rsp := &pb.RspGetCollectorPoint{
		Points: make([]*pb.RspGetCollectorPoint_CollectorPoint, 0),
	}
	// 查询出采集器涉及的采集测点
	points := make([]*model.CollectorTemplatePoint, 0)
	var err error
	switch collector.CollectorType {
	case model.CollectorTypeTboxSubDevice, model.CollectorTypeToneSubDevice, model.CollectorTypeDoorSubDevice:
		points, err = obj.collectorDao.GetTemplatePoint(ctx, collector.TemplateName, "")
		if err != nil {
			return nil, fmt.Errorf("get template point fail, err:%v", err)
		}
	case model.CollectorTypeVendorSubDevice:
		vendorCollector, cnt := obj.collectorDao.GetDeviceList(req.MozuId, &dto.CondCollectorGetDeviceList{
			DeviceNumber: []string{collector.ParentDeviceNumber},
		})
		if cnt != 1 {
			return nil, fmt.Errorf("bad request param, template not found")
		}
		points, err = obj.collectorDao.GetTemplatePoint(ctx, vendorCollector[0].TemplateName, collector.DeviceCode)
		if err != nil {
			return nil, fmt.Errorf("get template point fail, err:%v", err)
		}
	default:
		return nil, fmt.Errorf("collector type not support")
	}
	// 将采集器信息转换为rsp结构
	for _, item := range points {
		rspItem := &pb.RspGetCollectorPoint_CollectorPoint{}
		_ = copyutil.Copy(item, rspItem)
		rspItem.PointKey = fmt.Sprintf("%s.%s", collector.DeviceGid, item.PointNameEn)
		rspItem.DeltaDef = util.JsonStrToPbStruct(item.DeltaDef)
		rspItem.VerifyDef = util.JsonStrToPbStruct(item.VerifyDef)
		rspItem.ExpDef = util.JsonStrToPbStruct(item.ExpDef)
		rspItem.ProtDef = util.JsonStrToPbStruct(item.ProtDef)
		rspItem.ValDef = util.JsonStrToPbStruct(item.ValDef)
		rspItem.Simulator = util.JsonStrToPbStruct(item.Simulator)
		rsp.Points = append(rsp.Points, rspItem)
	}
	return rsp, nil
}

func getCollectorStatusKey(collector *model.CollectorDevice) string {
	if collector.CollectorType == model.CollectorTypeTboxSubDevice {
		return fmt.Sprintf("%s.%s", collector.DeviceGid, consts.TboxSubDeviceStatusPointName)
	} else {
		return fmt.Sprintf("%s.%s", collector.DeviceGid, consts.CommonStatusPointName)
	}
}
