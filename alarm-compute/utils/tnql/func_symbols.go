package tnql

const (
	/*
		函数中的参数字段说明

		TOKEN - 测点 token，即 A、B、C 这种已经讲测点转换后的字符串
		value - 测点值
		duration - 持续时间
		interval - 采样间隔（可选），默认为 1，即每秒的数据都获取。获取的测点不可超过 1W 个，超过会自动调整 interval = 测点数量 / 10000

		示例：
		1. 恒等于函数 AEQ(TOKEN, value, duration, [interval]), All Equal
			a. AEQ(整流模块.整流模块工作状态, 3, 110)
				- 说明：【整流模块.整流模块工作状态】测点等于3，
				- 且 110 秒内的测点数据（默认每秒都采样）都满足该条件，则返回 true，否则返回 false
			b. AEQ(门禁控制器.门常开告警状态_1, 1, 43200, 60)
				- 说明：【门禁控制器.门常开告警状态_1】测点等于1，
				- 且 43200 秒内的测点数据（每隔 60 秒采样）都满足该条件，则返回 true，否则返回 false
				- 由于判断的持续时间 duration 为 43200，默认按 interval 1 的话需要获取 4W+测点。
				- 设置 interval 为 60，即表示每隔 60 秒采样，获取测点数据为 43200/60 = 720，减少获取的测点数据量，提高性能
		2. 最小值不等于函数 MinNEQ(TOKEN, value, duration, [interval]) Min No Equal
			- MinNEQ(整流模块.整流模块工作状态, 1, 300)
			- 说明：使用 300 秒内的【整流模块.整流模块工作状态】测点进行计算，最小值不等于1，则返回 true，否则返回 false
	*/

	// 以下为 aggregation functions 聚合函数，返回 float64 值

	// ExprFuncSum 求和 Sum(TOKEN, duration, [interval]) Sum Equal
	ExprFuncSum = "Sum"
	// ExprFuncAvg TODO
	ExprFuncAvg = "Avg"
	// ExprFuncMin TODO
	ExprFuncMin = "Min"
	// ExprFuncMax TODO
	ExprFuncMax = "Max"

	// ExprFuncDeltaGT Delta(TOKEN, pntValue, delay, [interval])
	ExprFuncDeltaGT = "DeltaGT"
	// ExprFuncDeltaGTE ExprFuncDeltaGTE
	ExprFuncDeltaGTE = "DeltaGTE"
	// ExprFuncDeltaEQ ExprFuncDeltaEQ
	ExprFuncDeltaEQ = "DeltaEQ"
	// ExprFuncDeltaNEQ ExprFuncDeltaNEQ
	ExprFuncDeltaNEQ = "DeltaNEQ"
	// ExprFuncDeltaLT ExprFuncDeltaLT
	ExprFuncDeltaLT = "DeltaLT"
	// ExprFuncDeltaLTE ExprFuncDeltaLTE
	ExprFuncDeltaLTE = "DeltaLTE"

	// 以下为 compartor functions 比较函数，返回 bool 值

	// ExprFuncDelayEQ 延迟 DelayEQ(TOKEN, value, duration)
	// 取 (T0 - duration) 与 T0 秒 的两个时刻的测点值运行表达式（T0为当前时刻）
	// 示例： DelayEQ(A, 5, 60)，A测点 (T0 - 60) 的值为 6，T0 秒 的值为 7，分别计算是否等于 5，都满足则返回 true，否则返回 false
	// 表达式 - 示例 A==5
	// duration - 持续时间
	ExprFuncDelayEQ  = "DelayEQ"
	ExprFuncDelayNEQ = "DelayNEQ"
	ExprFuncDelayLT  = "DelayLT"
	ExprFuncDelayLTE = "DelayLTE"
	ExprFuncDelayGT  = "DelayGT"
	ExprFuncDelayGTE = "DelayGTE"

	// ExprFuncJP 跳变 JP(TOKEN, value, duration)
	// 取T0与Td秒前、Td+1秒前3个时刻的数据判断
	ExprFuncJP = "JP"
	// ExprFuncNJP 非跳变 NJP(TOKEN, value, duration)，等同于 `恒等于` AEQ 函数
	ExprFuncNJP = "NJP"

	// 以下为 compartor all check functions 比较函数，返回 bool 值，使用时间段内的所有测点分别进行计算，全都满足则返回 true，否则返回 false

	// ExprFuncAEQ 恒等于函数 AEQ(TOKEN, value, duration, [interval]) 即所有的测点都等于给定的数值 All Equal
	ExprFuncAEQ = "AEQ"
	// ExprFuncANEQ 恒不等于 ANEQ(TOKEN, value, duration, [interval]) All not Equal
	ExprFuncANEQ = "ANEQ"
	// ExprFuncAGT 恒大于 AGT(TOKEN, value, duration, [interval]) All Great Than
	ExprFuncAGT = "AGT"
	// ExprFuncAGTE 恒大于等于 AGTE(TOKEN, value, duration, [interval]) All Great Than Equal
	ExprFuncAGTE = "AGTE"
	// ExprFuncALT 恒小于 ALT(TOKEN, value, duration, [interval]) All Less Than
	ExprFuncALT = "ALT"
	// ExprFuncALTE 恒小于等于 ANEQ(TOKEN, value, duration, [interval]) All Less Than Equal
	ExprFuncALTE = "ALTE"

	// 以下为 compartor aggregation functions 比较函数，返回 bool 值，使用聚合函数比较大小

	// ExprFuncSumEQ 求和等于 SumEQ(TOKEN, value, duration, [interval]) Sum Equal
	ExprFuncSumEQ = "SumEQ"
	// ExprFuncSumNEQ TODO
	ExprFuncSumNEQ = "SumNEQ"
	// ExprFuncSumGT TODO
	ExprFuncSumGT = "SumGT"
	// ExprFuncSumGTE TODO
	ExprFuncSumGTE = "SumGTE"
	// ExprFuncSumLT TODO
	ExprFuncSumLT = "SumLT"
	// ExprFuncSumLTE TODO
	ExprFuncSumLTE = "SumLTE"
	// ExprFuncAvgEQ TODO
	ExprFuncAvgEQ = "AvgEQ"
	// ExprFuncAvgNEQ TODO
	ExprFuncAvgNEQ = "AvgNEQ"
	// ExprFuncAvgGT TODO
	ExprFuncAvgGT = "AvgGT"
	// ExprFuncAvgGTE TODO
	ExprFuncAvgGTE = "AvgGTE"
	// ExprFuncAvgLT TODO
	ExprFuncAvgLT = "AvgLT"
	// ExprFuncAvgLTE TODO
	ExprFuncAvgLTE = "AvgLTE"
	// ExprFuncMinEQ TODO
	ExprFuncMinEQ = "MinEQ"
	// ExprFuncMinNEQ TODO
	ExprFuncMinNEQ = "MinNEQ"
	// ExprFuncMinGT TODO
	ExprFuncMinGT = "MinGT"
	// ExprFuncMinGTE TODO
	ExprFuncMinGTE = "MinGTE"
	// ExprFuncMinLT TODO
	ExprFuncMinLT = "MinLT"
	// ExprFuncMinLTE TODO
	ExprFuncMinLTE = "MinLTE"
	// ExprFuncMaxEQ TODO
	ExprFuncMaxEQ = "MaxEQ"
	// ExprFuncMaxNEQ TODO
	ExprFuncMaxNEQ = "MaxNEQ"
	// ExprFuncMaxGT TODO
	ExprFuncMaxGT = "MaxGT"
	// ExprFuncMaxGTE TODO
	ExprFuncMaxGTE = "MaxGTE"
	// ExprFuncMaxLT TODO
	ExprFuncMaxLT = "MaxLT"
	// ExprFuncMaxLTE TODO
	ExprFuncMaxLTE = "MaxLTE"

	// ExprFuncCountRateGT TODO
	ExprFuncCountRateGT = "CountRateGT"
	// ExprFuncCountRateGTE TODO
	ExprFuncCountRateGTE = "CountRateGTE"
	// ExprFuncCountRateLT TODO
	ExprFuncCountRateLT = "CountRateLT"
	// ExprFuncCountRateLTE TODO
	ExprFuncCountRateLTE = "CountRateLTE"
	// ExprFuncCountRateEQ TODO
	ExprFuncCountRateEQ = "CountRateEQ"
	// ExprFuncCountRateNEQ TODO
	ExprFuncCountRateNEQ = "CountRateNEQ"

	// ExprFuncCountGT TODO
	ExprFuncCountGT = "CountGT"
	// ExprFuncCountGTE TODO
	ExprFuncCountGTE = "CountGTE"
	// ExprFuncCountLT TODO
	ExprFuncCountLT = "CountLT"
	// ExprFuncCountLTE TODO
	ExprFuncCountLTE = "CountLTE"
	// ExprFuncCountLTE TODO
	ExprFuncCountEQ = "CountEQ"
	// ExprFuncCountLTE TODO
	ExprFuncCountNEQ = "CountNEQ"

	// ExprFuncAbsValue TODO
	ExprFuncAbsValue = "AbsValue"
	// ExprFuncAvgValue TODO
	ExprFuncAvgValue = "AvgValue"
	// ExprFuncMaxValue TODO
	ExprFuncMaxValue = "MaxValue"
	// ExprFuncMinValue TODO
	ExprFuncMinValue = "MinValue"
	// ExprFuncSumValue TODO
	ExprFuncSumValue = "SumValue"
)
