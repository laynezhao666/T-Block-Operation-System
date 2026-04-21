package consts

// Quality 测点值质量类型
type Quality int

const (
	QualityOK        Quality = 0    // 正常
	QualityUncertain Quality = -100 // 未知
)
