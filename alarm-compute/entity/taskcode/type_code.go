package taskcode

// RuleTaskTimeType 任务类型 实时/延时/虚拟
type RuleTaskTimeType int64

const (
	RuleTaskRealtime  = RuleTaskTimeType(0)
	RuleTaskDelaytime = RuleTaskTimeType(1)
	RuleTaskVirtual   = RuleTaskTimeType(2)
)
