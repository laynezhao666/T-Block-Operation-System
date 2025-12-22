package repo

// DbPoint 代表存入数据库的单个测点的数据结构
type DbPoint struct {
	I string  `json:"i"` // 测点名称
	V float64 `json:"v"` // 测点值
	Q string  `json:"q"` // 质量？
	T int64   `json:"t"` // 时间戳
	D int64   `json:"d"` // 上报数据类型，变化1，离线2，周期60
}
