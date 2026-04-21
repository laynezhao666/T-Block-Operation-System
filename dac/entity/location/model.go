// Package location 提供边缘节点位置信息的获取和管理。
package location

// l 全局边缘信息实例
var (
	l EdgeInfo = &edgeInfo{}
)

// EdgeInfo 边缘节点信息接口
type EdgeInfo interface {
	Park() string   // 获取园区名称
	Mozu() string   // 获取模组名称
	MozuID() int    // 获取模组ID
	ParkID() int    // 获取园区ID
	MozuIDs() []int // 获取所有模组ID列表
}

// edgeInfo 边缘节点信息实现
type edgeInfo struct {
	ParkName   string `json:"park"`       // 园区名称
	ParkId     int64  `json:"parkId"`     // 园区ID
	Building   string `json:"building"`   // 楼栋名称
	BuildingID int64  `json:"buildingId"` // 楼栋ID
	MozuName   string `json:"mozu"`       // 模组名称
	MozuId     int64  `json:"mozuId"`     // 模组ID
}

// Mozu 返回模组名称
func (e *edgeInfo) Mozu() string {
	return e.MozuName
}

// Park 返回园区名称
func (e *edgeInfo) Park() string {
	return e.ParkName
}

// MozuID 返回模组ID
func (e *edgeInfo) MozuID() int {
	return int(e.MozuId)
}

// ParkID 返回园区ID
func (e *edgeInfo) ParkID() int {
	return int(e.ParkId)
}

// MozuIDs 返回模组ID列表（单节点模式只有一个）
func (e *edgeInfo) MozuIDs() []int {
	return []int{e.MozuID()}
}

// edgeInfoPooling 池化模式的边缘信息（支持多模组）
type edgeInfoPooling []edgeInfo

// Mozu 返回第一个模组名称
func (e edgeInfoPooling) Mozu() string {
	if len(e) == 0 {
		return ""
	}
	return e[0].Mozu()
}

// Park 返回第一个园区名称
func (e edgeInfoPooling) Park() string {
	if len(e) == 0 {
		return ""
	}
	return e[0].Park()
}

// MozuID 返回第一个模组ID
func (e edgeInfoPooling) MozuID() int {
	if len(e) == 0 {
		return 0
	}
	return e[0].MozuID()
}

// ParkID 返回第一个园区ID
func (e edgeInfoPooling) ParkID() int {
	if len(e) == 0 {
		return 0
	}
	return e[0].ParkID()
}

// MozuIDs 返回所有模组ID列表
func (e edgeInfoPooling) MozuIDs() []int {
	ids := make([]int, 0, len(e))
	for i := range e {
		ids = append(ids, e[i].MozuID())
	}
	return ids
}

// Info 获取全局边缘节点信息
func Info() EdgeInfo {
	return l
}

// getEdgeLocationResp 获取边缘位置信息的响应结构
type getEdgeLocationResp struct {
	Code    int      `json:"code"`    // 响应码
	Message string   `json:"message"` // 响应消息
	Data    edgeInfo `json:"data"`    // 边缘信息
}

// getPoolingEdgeLocationResp 获取池化边缘位置信息的响应结构
type getPoolingEdgeLocationResp struct {
	Code    int             `json:"code"`    // 响应码
	Message string          `json:"message"` // 响应消息
	Data    edgeInfoPooling `json:"data"`    // 池化边缘信息
}
