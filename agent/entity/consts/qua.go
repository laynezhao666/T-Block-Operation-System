package consts

// Quality 测点值质量类型
type Quality int

const (
	QualityOk Quality = 0 // 正常

	QualityUncertain Quality = -100 // 原0x80

	QualityCommDisconnected Quality = -200 // 通信中断
	QualityCmdSendError     Quality = -201 // 报文发送失败
	QualityCmdRespTimeout   Quality = -202 // 报文响应超时
	QualityCmdRespError     Quality = -203 // 报文响应错误
	QualityConfigError      Quality = -204 // 配置错误
	QualityDriverOpenFailed Quality = -205 // 驱动打开失败
	QualityCannotOpen       Quality = -206 // 通道无法打开
	QualityHeartbeatFail    Quality = -207 // 订阅模式心跳连接失败
	QualitySubscribeFail    Quality = -208 // 订阅失败

	QualityCommAbnormal      Quality = -300 // 通讯异常
	QualityRespCRCError      Quality = -301 // 校验错误
	QualityRespHaveErrorCode Quality = -302 // 报文指示有错误码
	QualityRespTransIDError  Quality = -303 // 请求报文与响应报文序列号不一致
	QualityRespLenError      Quality = -304 // 响应报文长度校验失败

	QualityUncollected       Quality = -403 // 未采集 (未初始化)
	QualityValueOutOfRange   Quality = -406 // 值越界
	QualityValueTypeError    Quality = -408 // 值格式转换错误
	QualityValueInvalidError Quality = -409 // 设备返回数据为无效值（协议约定当值为XX时表示无效，如0x8000）
	QualityTooManyZero               = -411 // 零值过多
	QualityValueAbnormal     Quality = -421 // 值异常 原0x40
	QualityValueTmsExpired   Quality = -422 // 值的时间戳过期（订阅推送的数据无最新值）
	QualityValueOverflow     Quality = -423 // 值溢出
	QualityAddrError         Quality = -424 // 地址格式解析错误
	// 下面的值定义在弱电巡检仪里会返回
	QualityMissExpVal Quality = -431 // 参与表达式计算的数据当前不可用
	QualityValIdle    Quality = -432 // 该测点当前不在计算范围内，处于空闲状态，此时时间戳不再更新。例如：SOC及剩余容量仅在放电时计算，在非放电时就是空闲

	QualityUnderBoxNorthErr Quality = -503 //子设备（非直采）异常，即下层通讯管理机放在value中的北向错误类型（包含-99990--99999等）

	QualityStdExprErr     = -900 // 标准化表达式错误
	QualityStdEvalErr     = -901 // 标准化计算错误
	QualityStdParamErr    = -902 // 标准化参数格式错误
	QualityStdParamQuaErr = -903 // 标准化参数数据质量异常
	QualityStdValTypeErr  = -904 // 标准点数据类型非法
	QualityStdValNilErr   = -905 // 标准点值为空（nil）
	QualityStdNaNInfErr   = -906 // 标准点值为NaN或Inf
)
