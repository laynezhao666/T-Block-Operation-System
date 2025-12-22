package model

// PointData 测点数据
type PointData struct {
	Name    string  // 测点名称
	Time    int64   // 	测点时间
	Value   float64 // 测点值
	Quality int32   // 测点质量
}

// IsValid 是否有效
func (p *PointData) IsValid(now int64) bool {
	if p.Quality < 0 {
		return false
	}
	if now > p.Time+3*60 || now < p.Time {
		return false
	}
	return true
}
