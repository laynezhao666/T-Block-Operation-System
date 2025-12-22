// Package model 全局使用到的实体
package model

// StdPoint 字段类型标准化后的测点
type StdPoint struct {
	Name    string  // 测点名称
	Quality int16   // 测点质量
	Value   float64 // 测点值
	Time    uint32  // 时间戳

	Type     int32 // 测点类型
	Interval int32 // 是否变化点
	MozuId   int32 // 模组ID
}

// CachePoint 用于本地缓存的测点对象
type CachePoint struct {
	Quality int16   // 测点质量
	Value   float64 // 测点值
}
