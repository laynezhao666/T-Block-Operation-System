package entity

// Point 测点信息
type Point struct {
	Name    string  `json:"i"` // 测点名称
	Quality int64   `json:"q"` // 测点质量
	Value   float64 `json:"v"` // 测点值
	Time    int64   `json:"t"` // 时间戳
}

// CachePoint 用于Redis
type CachePoint struct {
	Quality int64   `json:"q"` // 测点质量
	Value   float64 `json:"v"` // 测点值
	Time    int64   `json:"t"` // 时间戳
}
