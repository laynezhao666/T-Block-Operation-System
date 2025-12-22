// Package model 一些公共的实体
package model

// Point 字段类型标准化后的测点
type Point struct {
	Name    string  `json:"i"` // 测点名称
	Quality int32   `json:"q"` // 测点质量
	Value   float64 `json:"v"` // 测点值
	Time    int64   `json:"t"` // 时间戳
	MozuId  int32   `json:"m"`

	EvalTms int64 // 测点计算时间
	SendTms int64 // 测点发送时间
}

// IsValid 是否有效
func (p *Point) IsValid(now int64) bool {
	if p.Quality < 0 {
		return false
	}
	if now > p.Time+3*60 || now < p.Time {
		return false
	}
	return true
}
