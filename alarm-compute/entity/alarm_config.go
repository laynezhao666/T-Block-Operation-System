package entity

// VariableGidMap (告警/恢复)变量映射
type VariableGidMap struct {
	ExprMap map[string]string `json:"expr_map"`
	Engine  string            `json:"engine"`
}

// ExpressionsMap (告警/恢复)表达式映射
type ExpressionsMap struct {
	Fire    VariableGidMap `json:"fire"`
	Restore VariableGidMap `json:"restore"`
}

// AlarmConfig 告警配置
type AlarmConfig struct {
	Rid               int64
	Gid               string
	RidVersion        string
	RidType           int64
	MozuId            int64
	AlarmExpression   string
	RestoreExpression string
	ExpressionMap     ExpressionsMap
	AlarmLevel        string
	AlarmName         string
	ContentTemplate   string
}
