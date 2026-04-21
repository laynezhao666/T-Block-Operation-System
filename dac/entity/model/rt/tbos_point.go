package rt

const (
	PointTypeCollect    = 1 //采集点
	PointTypeStd        = 2 //标准点
	PointIntervalPeriod = 60
	PointIntervalChange = 1
)

// MsgPoint 测点信息
type MsgPoint struct {
	I string `json:"i"` // 测点名称
	V string `json:"v"` // 测点值
	Q string `json:"q"` // 质量
	T string `json:"t"` // 时间戳
}

// MsgKey 测点key
type MsgKey struct {
	MozuID      string `json:"mID"`   // 模组ID
	DeviceGiD   string `json:"dID"`   // 设备GID
	WorkerID    string `json:"wID"`   // WorkerID
	Seq         int    `json:"seq"`   // 序号
	Timestamp   int64  `json:"t"`     // 推送时间S
	Interval    int32  `json:"d"`     // 测点周期
	BalancerKey string `json:"bKey"`  // 分区hash Key
	PubMs       int64  `json:"pubMs"` // 推送毫秒时间
	Type        int32  `json:"type"`  // 测点类型（采集、标准）
}

// MsgValue 测点value
type MsgValue struct {
	Interval      int64       `json:"interval"`
	BoxID         string      `json:"box_id"` // TBox ID
	Points        []*MsgPoint `json:"points"` // 测点数据组
	VirtualPoints []*MsgPoint `json:"virtual_points"`
}

// StdInstancePointInfo 标准化测点
type StdInstancePointInfo struct {
	// 标准设备
	StdDevice string `json:"device_gid"`
	// 设备编号
	DeviceNumber string `json:"device_number"`
	// 标准测点
	StdPoint string `json:"point_name_en"`
	// 标准测点中文名
	StdPointZh string `json:"point_name_zh"`
	// 测点key
	PointKey string `json:"point_key"`
	// 变化阈值（绝对值）
	Threshold string `json:"threshold"`
	// 映射表达式
	Expr string `json:"expression"`
	// 映射
	Mapping string `json:"expression_map"`
	// 可读的映射
	MappingZh string `json:"expression_map_zh"`
	// 映射参数
	Param map[string]string
	// 测点值类型(数据类型：模拟量、状态量)
	ValueType string `json:"value_type"`
	// 是否启用(0:禁用, 1:启用)
	Enable int32 `json:"point_kpi"`
	// 测点值有效范围
	ValueValidRange string `json:"value_valid_range"`
	// 测点值单位
	ValueUnit string `json:"value_unit"`
	// 测点值精度
	ValuePrecision string `json:"value_precision"`
	// 值枚举映射
	ValueEnum string `json:"value_enum"`
	// 读写
	PointRw string `json:"point_rw"`
	// 等级
	PointLevel string `json:"point_level"`
	// 类型
	PointType int32 `json:"point_type"`
	// 特征
	PointFeature string `json:"point_feature"`
	// 分组
	PointGroup string `json:"point_group"`
	// 是否标准点
	PointStandard int32 `json:"point_standard"`
	// 模组id
	MozuId int32 `json:"mozu_id"`
}

// StdInstancePointsInfo 标准化测点信息
type StdInstancePointsInfo []StdInstancePointInfo

// PointConfig 标准点配置
type PointConfig struct {
	DevicePoints []any `json:"device_points"`
}

// ConfigVersion 标准点配置版本
type ConfigVersion struct {
	Collector string `json:"collector"`
	Point     string `json:"point"`
}
