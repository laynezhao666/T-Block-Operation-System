// Package model 主链路使用到的一些实体
package model

// OriginPointMsg kafka原始测点信息
type OriginPointMsg struct {
	KafkaKey  []byte
	KafkaVal  []byte
	StdPoints []*Point
}

// Point 字段类型标准化后的测点
type Point struct {
	Name     string  // 测点名称
	Quality  int32   // 测点质量
	Value    float64 // 测点值
	Time     int64   // 时间戳
	Type     int32   // 测点类型
	Interval int32   // 是否变化点
	MozuId   int32   // 模组ID

	CollectTs  int64 // 采集时间
	ConsumerTs int64 // 消费时间
	InfluxTs   int64 // Influx写入时间
}

// GetMozuId 获取测点所属模组
func GetMozuId(point *Point) int32 {
	return point.MozuId
}
