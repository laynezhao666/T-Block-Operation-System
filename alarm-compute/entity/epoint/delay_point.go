package epoint

// DelayPointMap DelayPointMap
type DelayPointMap struct {
	// 测点 <interval, pointList> ，按 interval 时间点获取数据
	HPointMap map[int][]string

	// duration 测点 <duration, pointList> ，获取从当前时间开始，一段时间内的测点 [t0+duration, t0)
	HDPointMap map[int][]string

	// range 测点 < delay, < range, pointList > > ，选取一段时间获取测点数据。目前用于跳变运算符，需要选去中间一段时间获取数据
	// 注意，HDPointMap 是从当前时间开始计算获取一段时间
	HRPointMap map[int]map[int][]string
}

// NewDelayPointMap NewDelayPointMap
func NewDelayPointMap() *DelayPointMap {
	return &DelayPointMap{
		HPointMap:  make(map[int][]string),
		HDPointMap: make(map[int][]string),
		HRPointMap: make(map[int]map[int][]string),
	}
}
