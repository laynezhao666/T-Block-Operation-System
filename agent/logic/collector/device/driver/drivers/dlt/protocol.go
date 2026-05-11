// Package dlt DL/T 645-2007 多功能电能表通信协议驱动
// 本文件定义了协议相关的常量和数据标识
package dlt

// 数据标识定义 (DI3-DI2-DI1-DI0)
// 根据DL/T 645-2007附录A
const (
	// ========== 电能量数据标识 (DI3=00) ==========

	// 正向有功总电能
	DI_PosActiveTotal uint32 = 0x00010000
	// 正向有功费率1电能
	DI_PosActiveRate1 uint32 = 0x00010001
	// 正向有功费率2电能
	DI_PosActiveRate2 uint32 = 0x00010002
	// 正向有功费率3电能
	DI_PosActiveRate3 uint32 = 0x00010003
	// 正向有功费率4电能
	DI_PosActiveRate4 uint32 = 0x00010004

	// 反向有功总电能
	DI_NegActiveTotal uint32 = 0x00020000
	// 反向有功费率1电能
	DI_NegActiveRate1 uint32 = 0x00020001
	// 反向有功费率2电能
	DI_NegActiveRate2 uint32 = 0x00020002
	// 反向有功费率3电能
	DI_NegActiveRate3 uint32 = 0x00020003
	// 反向有功费率4电能
	DI_NegActiveRate4 uint32 = 0x00020004

	// 正向无功总电能
	DI_PosReactiveTotal uint32 = 0x00030000
	// 反向无功总电能
	DI_NegReactiveTotal uint32 = 0x00040000

	// 一象限无功总电能
	DI_Q1ReactiveTotal uint32 = 0x00050000
	// 二象限无功总电能
	DI_Q2ReactiveTotal uint32 = 0x00060000
	// 三象限无功总电能
	DI_Q3ReactiveTotal uint32 = 0x00070000
	// 四象限无功总电能
	DI_Q4ReactiveTotal uint32 = 0x00080000

	// 正向视在总电能
	DI_PosApparentTotal uint32 = 0x00090000
	// 反向视在总电能
	DI_NegApparentTotal uint32 = 0x000A0000

	// ========== 变量数据标识 (DI3=02) ==========

	// A相电压
	DI_VoltageA uint32 = 0x02010100
	// B相电压
	DI_VoltageB uint32 = 0x02010200
	// C相电压
	DI_VoltageC uint32 = 0x02010300

	// A相电流
	DI_CurrentA uint32 = 0x02020100
	// B相电流
	DI_CurrentB uint32 = 0x02020200
	// C相电流
	DI_CurrentC uint32 = 0x02020300

	// 瞬时总有功功率
	DI_ActivePowerTotal uint32 = 0x02030000
	// A相有功功率
	DI_ActivePowerA uint32 = 0x02030100
	// B相有功功率
	DI_ActivePowerB uint32 = 0x02030200
	// C相有功功率
	DI_ActivePowerC uint32 = 0x02030300

	// 瞬时总无功功率
	DI_ReactivePowerTotal uint32 = 0x02040000
	// A相无功功率
	DI_ReactivePowerA uint32 = 0x02040100
	// B相无功功率
	DI_ReactivePowerB uint32 = 0x02040200
	// C相无功功率
	DI_ReactivePowerC uint32 = 0x02040300

	// 瞬时总视在功率
	DI_ApparentPowerTotal uint32 = 0x02050000
	// A相视在功率
	DI_ApparentPowerA uint32 = 0x02050100
	// B相视在功率
	DI_ApparentPowerB uint32 = 0x02050200
	// C相视在功率
	DI_ApparentPowerC uint32 = 0x02050300

	// 总功率因数
	DI_PowerFactorTotal uint32 = 0x02060000
	// A相功率因数
	DI_PowerFactorA uint32 = 0x02060100
	// B相功率因数
	DI_PowerFactorB uint32 = 0x02060200
	// C相功率因数
	DI_PowerFactorC uint32 = 0x02060300

	// 电网频率
	DI_Frequency uint32 = 0x02800002

	// ========== 参变量数据标识 (DI3=04) ==========

	// 日期及星期
	DI_DateAndWeek uint32 = 0x04000101
	// 时间
	DI_Time uint32 = 0x04000102
	// 电表运行状态字
	DI_MeterStatus uint32 = 0x04000301
	// 电表常数(有功)
	DI_MeterConstActive uint32 = 0x04000401
	// 电表常数(无功)
	DI_MeterConstReactive uint32 = 0x04000402
	// 电表资产号
	DI_AssetNumber uint32 = 0x04000403
	// 电表额定电压
	DI_RatedVoltage uint32 = 0x04000404
	// 电表额定电流
	DI_RatedCurrent uint32 = 0x04000405
	// 电表表号
	DI_MeterNumber uint32 = 0x04000406
)

// DataIdInfo 数据标识信息
type DataIdInfo struct {
	ID       uint32 // 数据标识
	Name     string // 名称
	Unit     string // 单位
	Decimals int    // 小数位数
	Length   int    // 数据长度(字节)
}

// DataIdMap 数据标识映射表
var DataIdMap = map[uint32]DataIdInfo{
	// 电能量
	DI_PosActiveTotal:   {DI_PosActiveTotal, "正向有功总电能", "kWh", 2, 4},
	DI_PosActiveRate1:   {DI_PosActiveRate1, "正向有功费率1电能", "kWh", 2, 4},
	DI_PosActiveRate2:   {DI_PosActiveRate2, "正向有功费率2电能", "kWh", 2, 4},
	DI_NegActiveTotal:   {DI_NegActiveTotal, "反向有功总电能", "kWh", 2, 4},
	DI_PosReactiveTotal: {DI_PosReactiveTotal, "正向无功总电能", "kvarh", 2, 4},
	DI_NegReactiveTotal: {DI_NegReactiveTotal, "反向无功总电能", "kvarh", 2, 4},
	DI_PosApparentTotal: {DI_PosApparentTotal, "正向视在总电能", "kVAh", 2, 4},
	DI_NegApparentTotal: {DI_NegApparentTotal, "反向视在总电能", "kVAh", 2, 4},

	// 电压
	DI_VoltageA: {DI_VoltageA, "A相电压", "V", 1, 2},
	DI_VoltageB: {DI_VoltageB, "B相电压", "V", 1, 2},
	DI_VoltageC: {DI_VoltageC, "C相电压", "V", 1, 2},

	// 电流
	DI_CurrentA: {DI_CurrentA, "A相电流", "A", 3, 3},
	DI_CurrentB: {DI_CurrentB, "B相电流", "A", 3, 3},
	DI_CurrentC: {DI_CurrentC, "C相电流", "A", 3, 3},

	// 有功功率
	DI_ActivePowerTotal: {DI_ActivePowerTotal, "总有功功率", "kW", 4, 3},
	DI_ActivePowerA:     {DI_ActivePowerA, "A相有功功率", "kW", 4, 3},
	DI_ActivePowerB:     {DI_ActivePowerB, "B相有功功率", "kW", 4, 3},
	DI_ActivePowerC:     {DI_ActivePowerC, "C相有功功率", "kW", 4, 3},

	// 无功功率
	DI_ReactivePowerTotal: {DI_ReactivePowerTotal, "总无功功率", "kvar", 4, 3},
	DI_ReactivePowerA:     {DI_ReactivePowerA, "A相无功功率", "kvar", 4, 3},
	DI_ReactivePowerB:     {DI_ReactivePowerB, "B相无功功率", "kvar", 4, 3},
	DI_ReactivePowerC:     {DI_ReactivePowerC, "C相无功功率", "kvar", 4, 3},

	// 视在功率
	DI_ApparentPowerTotal: {DI_ApparentPowerTotal, "总视在功率", "kVA", 4, 3},
	DI_ApparentPowerA:     {DI_ApparentPowerA, "A相视在功率", "kVA", 4, 3},
	DI_ApparentPowerB:     {DI_ApparentPowerB, "B相视在功率", "kVA", 4, 3},
	DI_ApparentPowerC:     {DI_ApparentPowerC, "C相视在功率", "kVA", 4, 3},

	// 功率因数
	DI_PowerFactorTotal: {DI_PowerFactorTotal, "总功率因数", "", 3, 2},
	DI_PowerFactorA:     {DI_PowerFactorA, "A相功率因数", "", 3, 2},
	DI_PowerFactorB:     {DI_PowerFactorB, "B相功率因数", "", 3, 2},
	DI_PowerFactorC:     {DI_PowerFactorC, "C相功率因数", "", 3, 2},

	// 频率
	DI_Frequency: {DI_Frequency, "电网频率", "Hz", 2, 2},

	// 电表参数
	DI_MeterStatus:       {DI_MeterStatus, "电表运行状态", "", 0, 2},
	DI_MeterConstActive:  {DI_MeterConstActive, "电表常数(有功)", "imp/kWh", 0, 3},
	DI_MeterConstReactive:{DI_MeterConstReactive, "电表常数(无功)", "imp/kvarh", 0, 3},
}

// GetDataIdInfo 获取数据标识信息
func GetDataIdInfo(dataId uint32) (DataIdInfo, bool) {
	info, ok := DataIdMap[dataId]
	return info, ok
}
