// Package tbox 提供TBox设备相关的数据模型。
package tbox

// Position 设备物理位置信息
type Position struct {
	Room  string `json:"room"`  // 机房
	Block string `json:"block"` // 区域
	No    string `json:"no"`    // 编号
	Mark  string `json:"mark"`  // 标记
	Desc  string `json:"desc"`  // 描述
}
