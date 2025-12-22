package strategy

// StrategyFilter ...
type StrategyFilter struct {
	MozuId int64
	// 是否允许传入的mozu_id为0，默认为不允许
	MozuAllowZero bool
	Rid           []int64
	Gid           []string
	RidType       []int64
	DeviceNumber  []string
	ApplyType     []string
	DeviceType    []string
	AlarmName     []string
	Level         []string
	Page          int64
	Size          int64
}
