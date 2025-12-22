package epoint

// ValueMap { 实例测点: 测点值 }, { point: value }
type ValueMap map[string]float64

// IntervalMap { 间隔1: 测点值 }, { interval: value }
type IntervalMap map[int]float64

// HistoryUnixTimeValueMap { 实例测点: { Unix时间戳: 测点值 } }
type HistoryTimeValueMap map[string]map[int64]float64

// HistoryValueMap { 实例测点: { 间隔1: 测点值 } }, { pointName: { interval: value } }
type HistoryValueMap map[string]IntervalMap

// SymValueMapList {变量名称：[{间隔1:测点值, ...}, {间隔1:测点值, ...}] }
// 适配单变量映射为多设备测点的应用场景
// eg: {"A": [{0: 1.0, 5:2.0}, {0: 2.0, 5: 1.0}]}
type SymValueMapList map[string][]IntervalMap
