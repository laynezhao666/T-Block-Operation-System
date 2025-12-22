package message

// PointFetchInfo PointFetchInfo  delay point time info
type PointFetchInfo struct {
	// 获取测点的时长
	Duration int `json:"duration,omitempty"`
	// 获取测点的间隔
	Interval int `json:"interval,omitempty"`

	// 跳变延迟信息
	RangeDelay int `json:"jpDelay,omitempty"`
}

// PointTypeMap PointTypeMap
type PointTypeMap struct {
	// 采集测点类型表达式(A+B)
	Express string `json:"express,omitempty"`
	// 变量映射 例如：{"A":["1153029394454814720.GecFireAlarm"],"B":["1153029394454814720.中压市电供电中断"]}
	// 一个变量映射为 gid.pointName列表
	// gid 在不同模组中是全局唯一的
	PMap map[string][]string `json:"pMap,omitempty"`
	// 计算表达式使用的引擎
	Engine string `json:"engine,omitempty"`
	// 计算时间
	EvalTime int64 `json:"evalTime,omitempty"`
	// 拉取测点信息
	PointFetchList map[string][]PointFetchInfo `json:"-"`
	// 跳变函数拉取一段时间的配置
	JPRangeSec int `json:"-"`
}

// AlarmTaskRet 告警计算传入的分析结果
type AlarmTaskRet struct {
	PointValueMap        map[string]float64         `json:"pointValueMap,omitempty"`
	HistoryPointValueMap map[string]map[int]float64 `json:"historyPointValueMap,omitempty"`
	PointMap             map[string][]string        `json:"pointMap,omitempty"`
	StartRunAt           int64                      `json:"startRunAt,omitempty"`
	ExpMap               *PointTypeMap              `json:"expMap,omitempty"`
}
