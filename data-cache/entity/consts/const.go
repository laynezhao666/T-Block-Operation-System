// Package consts 保存一些常量
package consts

const (
	TbosMajorKafkaName  = "trpc.kafka.tbos.major"  // 主用Kafka名称
	TbosBackupKafkaName = "trpc.kafka.tbos.backup" // 备用Kafka名称

	PointIntervalPeriod  = 60 // 周期测点
	PointIntervalChanged = 1  // 变化测点

	PointTypeStdStr   = "std"                // 标准测点类型字符串
	PointTypeCollect  = 1                    // 采集测点类型
	PointTypeStd      = 2                    // 标准测点类型
	PointTypeVirtual  = 3                    // 虚拟测点类型
	PointTypeAlarmStr = "alarmVirtualPoints" // 告警测点类型字符串
	PointTypeAlarm    = 4                    // 告警测点类型
)
