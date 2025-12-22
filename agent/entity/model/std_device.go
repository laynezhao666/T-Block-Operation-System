package model

import (
	"errors"
	"fmt"
	"agent/entity/consts"
	"agent/logic/collector/device/model"
	"strings"
	"sync"
	pb "trpcprotocol/agent"
)

// StdDeviceData 标准设备数据
type StdDeviceData struct {
	// 标准设备列表
	StdDevices []StdDevice `json:"std_devices"`
	// [gid]设备
	StdDeviceMap      map[string]StdDevice
	StdDeviceMapMutex sync.RWMutex
	// [gid]测点
	StdPoints      map[string]model.StdInstancePointsInfo
	StdPointsMutex sync.RWMutex
	// [短名称]
	ConciseCodeMap      map[string]string
	ConciseCodeMapMutex sync.RWMutex
}

// Copy 复制
func (s *StdDeviceData) Copy() *StdDeviceData {
	if s == nil {
		return nil
	}
	newStdDeviceData := StdDeviceData{
		StdDevices: make([]StdDevice, len(s.StdDevices)),
		StdPoints:  make(map[string]model.StdInstancePointsInfo),
	}
	copy(newStdDeviceData.StdDevices, s.StdDevices)
	for k, v := range s.StdPoints {
		newPointsInfo := make(model.StdInstancePointsInfo, len(v))
		copy(newPointsInfo, v)
		newStdDeviceData.StdPoints[k] = newPointsInfo
	}
	return &newStdDeviceData
}

// GetDeviceByGid 通过gid获取设备信息
func (s *StdDeviceData) GetDeviceByGid(gid string) (StdDevice, bool) {
	if s == nil {
		return StdDevice{}, false
	}
	s.StdDeviceMapMutex.RLock()
	defer s.StdDeviceMapMutex.RUnlock()
	d, ok := s.StdDeviceMap[gid]
	return d, ok
}

// GetPointsByGid 通过gid获取测点信息
func (s *StdDeviceData) GetPointsByGid(gid string) (model.StdInstancePointsInfo, bool) {
	if s == nil {
		return model.StdInstancePointsInfo{}, false
	}
	s.StdPointsMutex.RLock()
	defer s.StdPointsMutex.RUnlock()
	ps, ok := s.StdPoints[gid]
	return ps, ok
}

// GetPointsByPointKey 通过gid.测点英文获取测点信息
func (s *StdDeviceData) GetPointsByPointKey(pointKey string) model.StdInstancePointsInfo {
	if s == nil {
		return model.StdInstancePointsInfo{}
	}
	s.StdPointsMutex.RLock()
	defer s.StdPointsMutex.RUnlock()
	parts := strings.Split(pointKey, consts.MpExpressionRefSplitChar)
	if len(parts) < 2 {
		return nil
	}
	gid := parts[0]
	pointEn := parts[1]
	ps, ok := s.StdPoints[gid]
	if !ok {
		return nil
	}
	res := make(model.StdInstancePointsInfo, 0)
	for _, p := range ps {
		if p.StdPoint == pointEn {
			res = append(res, p)
		}
	}
	return res
}

// GetStdPoints 获取所有标准测点列表
func (s *StdDeviceData) GetStdPoints() []model.StdInstancePointInfo {
	if s == nil {
		return model.StdInstancePointsInfo{}
	}
	s.StdPointsMutex.RLock()
	defer s.StdPointsMutex.RUnlock()
	res := make([]model.StdInstancePointInfo, 0)
	for _, ps := range s.StdPoints {
		res = append(res, ps...)
	}
	return res
}

// SavePointInfo 保存测点信息
func (s *StdDeviceData) SavePointInfo(pointKey string, express string, paramMap map[string]*pb.RefObject) (
	[]model.StdInstancePointInfo, error) {
	if s == nil {
		return nil, errors.New("StdDeviceData is empty")
	}
	s.StdPointsMutex.RLock()
	defer s.StdPointsMutex.RUnlock()
	parts := strings.Split(pointKey, consts.MpExpressionRefSplitChar)
	if len(parts) < 2 {
		return nil, nil
	}
	gid := parts[0]
	pointEn := parts[1]
	ps, ok := s.StdPoints[gid]
	if !ok {
		return nil, nil
	}
	res := make(model.StdInstancePointsInfo, 0)
	for _, p := range ps {
		if p.StdPoint == pointEn {
			// 将map反写为string
			// todo 校验
			// 1. 采集设备必须存在
			// 2. 对应的采集点必须存在
			mapping := ConvertMappingObjectToString(paramMap)
			p.Expr = express
			p.Mapping = mapping
		}
		res = append(res, p)
	}
	// 更新内存
	s.StdPoints[gid] = res
	// 将更新后的列表返回
	newPointList := s.GetStdPoints()
	return newPointList, nil
}

// ConvertMappingObjectToString 将map[string]*pb.RefObject 转换为 string
func ConvertMappingObjectToString(data map[string]*pb.RefObject) string {
	var result []string
	for key, value := range data {
		objectGID := value.Gid
		pointNameEN := value.PointNameEn
		if objectGID != "" && pointNameEN != "" {
			result = append(result, fmt.Sprintf("%s=%s.%s", key, objectGID, pointNameEN))
		}
	}
	return strings.Join(result, ";") + ";"
}

// GetPointsByConciseCode 通过短编号获取设备gid
func (s *StdDeviceData) GetPointsByConciseCode(gid string) string {
	if s == nil {
		return ""
	}
	s.ConciseCodeMapMutex.RLock()
	defer s.ConciseCodeMapMutex.RUnlock()
	return s.ConciseCodeMap[gid]
}

// StdDeviceTree 设备树
type StdDeviceTree struct {
	List []*TreeNode
}

// TreeNode 设备树节点
type TreeNode struct {
	DeviceGid         string      `json:"device_gid"`
	DeviceNumber      string      `json:"device_number"`
	DeviceNumberShow  string      `json:"device_number_show"`
	DeviceNo          string      `json:"device_no"`
	DeviceTypeEn      string      `json:"device_type_en"`
	DeviceTypeZh      string      `json:"device_type_zh"`
	ApplicationTypeEn string      `json:"application_type_en"`
	ApplicationTypeZh string      `json:"application_type_zh"`
	Children          []*TreeNode `json:"children"`
	DeviceCount       int32       `json:"device_count"`
}

// StdDevice 标准设备结构体
type StdDevice struct {
	DeviceGid               string `json:"device_gid"`                 // 设备GID
	DeviceNumber            string `json:"device_number"`              // 设备编码
	DeviceNumberShow        string `json:"device_number_show"`         // 展示设备编码
	ConciseCode             string `json:"concise_code"`               // 简短设备编码
	DeviceNo                string `json:"device_no"`                  // 设备编号,同层同类型设备序号
	DeviceName              string `json:"device_name"`                // 设备名称
	MozuId                  int32  `json:"mozu_id"`                    // 所属模组ID
	MozuName                string `json:"mozu_name"`                  // 模组名称
	IdcArea                 string `json:"idc_area"`                   // 机房区域
	FuncRoom                string `json:"func_room"`                  // 方仓/功能间
	ParentDeviceNumber      string `json:"parent_device_number"`       // 父级设备编码
	EnableStatus            int32  `json:"enable_status"`              // 可用状态
	DeviceTypeEn            string `json:"device_type_en"`             // 设备种类英文
	DeviceTypeZh            string `json:"device_type_zh"`             // 设备种类中文
	ApplicationTypeEn       string `json:"application_type_en"`        // 应用类型英文
	ApplicationTypeZh       string `json:"application_type_zh"`        // 应用类型中文
	BelongApplicationTypeEn string `json:"belong_application_type_en"` // 所属应用类型
}
