package consts

// Quality 测点值质量类型
type Quality int32

const (
	QualityOk Quality = 0 // 正常

	QualityPushKafkaErr     Quality = -600 // 测点推送kafka失败
	QualityCalcLessPointErr Quality = -601 // 测点计算缺少测点错误
	QualityQueryCacheApiErr Quality = -602 // 查询缓存api错误

	QualityStdExprErr    Quality = -900 // 标准化表达式错误
	QualityStdEvalErr    Quality = -901 // 标准化计算错误
	QualityStdValTypeErr Quality = -904 // 标准点数据类型非法
	QualityStdValNilErr  Quality = -905 // 标准点值为空（nil）
	QualityStdNaNInfErr  Quality = -906 // 标准点值为NaN或Inf

)
