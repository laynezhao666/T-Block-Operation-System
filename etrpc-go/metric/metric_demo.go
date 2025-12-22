package metric

import "trpc.group/trpc-go/trpc-go/metrics"

var FailCnt = NewMetric("FailCnt", WithPolicy(metrics.PolicyMAX)) // 失败次数

// reportTest 上报测试
func reportTest() {
	// Report函数协程安全，且是异步的,智研sdk按照每分钟进行汇聚后上报
	FailCnt.Report(10)
}
