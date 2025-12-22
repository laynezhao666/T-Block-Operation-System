package consts

// tbos client name
const (
	TbosMysqlName       = "trpc.mysql.tbos"        // MySQL名称
	TbosRedisName       = "trpc.redis.tbos"        // Redis名称
	TbosInfluxName      = "trpc.influx.tbos"       // InfluxDB名称
	TbosMajorKafkaName  = "trpc.kafka.tbos.major"  // 主用Kafka名称
	TbosBackupKafkaName = "trpc.kafka.tbos.backup" // 备用Kafka名称

	MySQLReadBatchSize = 5000 // 一次从mysql读取多少条记录
)

// point relate consts
const (
	PointConcatSep = "."

	CalcTypePeriod  = "period"  // 周期性计算
	CalcTypeChanged = "changed" // 变化驱动计算

	PointIntervalPeriod  int32 = 60 // 周期测点类型
	PointIntervalChanged int32 = 1  // 变化测点类型

	PointCategoryStd     = 1 // 标准测点
	PointCategoryCollect = 2 // 采集测点

	PointTypeStdStr   = "std"                // 标准测点类型字符串
	PointTypeCollect  = 1                    // 采集测点类型
	PointTypeStd      = 2                    // 标准测点类型
	PointTypeVirtual  = 3                    // 虚拟测点类型
	PointTypeAlarmStr = "alarmVirtualPoints" // 告警测点类型字符串
	PointTypeAlarm    = 4                    // 告警测点类型

	CommonStatusPointName        = "Comm"    // 通用状态测点
	TboxSubDeviceStatusPointName = "commste" // tbox子设备状态测点
)
