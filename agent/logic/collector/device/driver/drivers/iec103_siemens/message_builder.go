package iec103_siemens

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/logic/collector/device/model"
)

// msgBuilder 报文构造器
type msgBuilder struct {
	srcDeviceAddr  uint16
	dstDeviceAddr  uint16
	srcStationAddr uint16
	dstStationAddr uint16
	sequenceNum    uint32
}

// GetNextSequence 获取下一个序列号
func (m *msgBuilder) GetNextSequence() uint16 {
	num := atomic.AddUint32(&m.sequenceNum, 1)
	if num > 65535 {
		atomic.CompareAndSwapUint32(&m.sequenceNum, num, 0)
		num = atomic.AddUint32(&m.sequenceNum, 1)
	}
	return uint16(num)
}

// BuildHeartbeat 构建心跳报文
func (m *msgBuilder) BuildHeartbeat() []byte {
	data := make([]byte, IEC103_HeadLen)

	// 固定格式
	binary.LittleEndian.PutUint16(data[0:2], 0xEB90)
	binary.LittleEndian.PutUint32(data[2:6], IEC103_LenHeadVal)
	binary.LittleEndian.PutUint16(data[6:8], 0xEB90)

	// 源地址
	binary.LittleEndian.PutUint16(data[8:10], m.srcStationAddr)
	binary.LittleEndian.PutUint16(data[10:12], m.srcDeviceAddr)

	// 目标地址
	binary.LittleEndian.PutUint16(data[12:14], m.dstStationAddr)
	binary.LittleEndian.PutUint16(data[14:16], m.dstDeviceAddr)

	// 数据编号
	binary.LittleEndian.PutUint16(data[16:18], m.GetNextSequence())

	// 设备类型
	binary.LittleEndian.PutUint16(data[18:20], 0x0000)

	// 网络状态
	binary.LittleEndian.PutUint16(data[20:22], 0x1000)

	// 路由地址
	binary.LittleEndian.PutUint16(data[22:24], 0x0000)
	binary.LittleEndian.PutUint16(data[24:26], 0x0000)

	// 结束标记
	binary.LittleEndian.PutUint16(data[26:28], 0xFFFF)

	return data
}

// BuildTotalCall 构建总召唤报文
func (m *msgBuilder) BuildTotalCall() []byte {
	header := m.BuildCommonHeader(IEC103_HeadLen + 8)

	// ASDU部分
	asdu := make([]byte, 8)
	asdu[0] = ASDU_Type_CtrlSvc     // 类型标识
	asdu[1] = 0x81                  // VSQ
	asdu[2] = 0x09                  // 传送原因
	asdu[3] = byte(m.dstDeviceAddr) // 应用服务数据单元地址
	asdu[4] = 0xFE                  // 功能类型
	asdu[5] = 0xF5                  // 信息序号
	asdu[6] = 0x00                  // 返回信息标识符
	asdu[7] = 0x01                  // 通用分类数据集数目

	return append(header, asdu...)
}

// BuildMeasurementRequest 构建遥测请求
func (m *msgBuilder) BuildMeasurementRequest(stationAddr, deviceAddr, sequence uint16) []byte {
	// 遥测组号是2，请求整个组的数据
	header := m.BuildCommonHeader(IEC103_HeadLen + 11)

	asdu := make([]byte, 11)
	asdu[0] = ASDU_Type_CtrlSvc // 类型标识
	asdu[1] = 0x81              // VSQ
	asdu[2] = 0x2A              // 传送原因: 通用分类读命令
	asdu[3] = 0x0B              // 应用服务数据单元地址
	asdu[4] = 0xFE              // 功能类型
	asdu[5] = 0xF1              // 信息序号: 读一个组的全部标题的属性或值
	asdu[6] = 0x00              // 返回信息标识符
	asdu[7] = 0x01              // 通用分类数据集数目
	asdu[8] = 0x02              // 组号: 2
	asdu[9] = 0x00              // 条目号: 0
	asdu[10] = 0x01             // 描述类别: 实际值

	return append(header, asdu...)
}

// BuildEnergyCall 构建电度请求
func (m *msgBuilder) BuildEnergyCall() []byte {
	// 电度组号是3
	header := m.BuildCommonHeader(IEC103_HeadLen + 11)

	asdu := make([]byte, 11)
	asdu[0] = ASDU_Type_CtrlSvc     // 类型标识
	asdu[1] = 0x81                  // VSQ
	asdu[2] = COT_Read              // 传送原因: 通用分类读命令
	asdu[3] = byte(m.dstDeviceAddr) // 应用服务数据单元地址
	asdu[4] = 0xFE                  // 功能类型
	asdu[5] = 0xF1                  // 信息序号: 读一个组的全部标题的属性或值
	asdu[6] = 0x00                  // 返回信息标识符
	asdu[7] = 0x01                  // 通用分类数据集数目
	asdu[8] = 0x03                  // 组号: 3
	asdu[9] = 0x00                  // 条目号: 0
	asdu[10] = 0x01                 // 描述类别: 实际值

	return append(header, asdu...)
}

// BuildFaultWaveRequest 构建故障录波请求
func (m *msgBuilder) BuildFaultWaveRequest(stationAddr, deviceAddr, sequence uint16) []byte {
	// 故障录波使用传统IEC103方式
	header := m.BuildCommonHeader(IEC103_HeadLen + 10)

	asdu := make([]byte, 10)
	asdu[0] = 0x18 // 类型标识: 扰动数据传输的命令
	asdu[1] = 0x81 // VSQ
	asdu[2] = 0x1F // 传送原因: 扰动数据的传输
	asdu[3] = 0x0B // 应用服务数据单元地址
	asdu[4] = 0xA0 // 功能类型
	asdu[5] = 0x00 // 信息序号
	asdu[6] = 0x01 // 命令类型: 故障的选择
	asdu[7] = 0x01 // 扰动值的类型: 瞬间值
	asdu[8] = 0x01 // 故障序号: 1
	asdu[9] = 0x00 // 实际通道序号: 全局

	return append(header, asdu...)
}

// BuildControlCommand 构建控制命令
func (m *msgBuilder) BuildControlCommand(packet *model.ControlProtocolPacket, value string) (*ControlData, error) {

	// 解析控制参数
	groupNum, entryNum, err := m.ParseControlExtend(packet.Command)
	if err != nil {
		return nil, err
	}

	// 构建控制值
	controlValue := byte(0x01) // 默认分
	if value == "1" || strings.ToLower(value) == "on" || strings.ToLower(value) == "合" {
		controlValue = 0x02 // 合
	}

	// 构建控制选择命令
	selectData := m.BuildControlSelect(groupNum, entryNum, controlValue)

	// 构建控制执行命令
	executeData := m.BuildControlExecute(groupNum, entryNum, controlValue)

	return &ControlData{
		SelectData:  selectData,
		ExecuteData: executeData,
	}, nil
}

// BuildControlSelect 构建控制选择命令
func (m *msgBuilder) BuildControlSelect(groupNum, entryNum, value byte) []byte {
	header := m.BuildCommonHeader(IEC103_HeadLen + 15)

	asdu := make([]byte, 15)
	asdu[0] = 0x0A                  // 类型标识: 通用分类数据控制方向
	asdu[1] = 0x81                  // VSQ
	asdu[2] = COT_Write             // 传送原因: 通用分类写命令
	asdu[3] = byte(m.dstDeviceAddr) // 应用服务数据单元地址
	asdu[4] = 0xFE                  // 功能类型
	asdu[5] = INF_WriteWithConfirm  // 信息序号: 带确认的写条目
	asdu[6] = 0x00                  // 返回信息标识符
	asdu[7] = 0x01                  // 通用分类数据集数目
	asdu[8] = groupNum              // 组号
	asdu[9] = entryNum              // 条目号
	asdu[10] = 0x01                 // 描述类别: 实际值
	asdu[11] = 0x09                 // 数据类型: DIP（双点信息）
	asdu[12] = 0x01                 // 数据宽度: 1
	asdu[13] = 0x01                 // 数据数目: 1
	asdu[14] = value                // 控制值

	return append(header, asdu...)
}

// BuildControlSelectWithInfoNum 构建控制命令（支持自定义信息序号）
// infoNum: 信息序号，0xF9=带确认的写条目（选择），0xFA=带执行的写条目（执行）
func (m *msgBuilder) BuildControlSelectWithInfoNum(groupNum, entryNum, value, infoNum byte) []byte {
	header := m.BuildCommonHeader(IEC103_HeadLen + 15)

	asdu := make([]byte, 15)
	asdu[0] = 0x0A                  // 类型标识: 通用分类数据控制方向
	asdu[1] = 0x81                  // VSQ
	asdu[2] = COT_Write             // 传送原因: 通用分类写命令
	asdu[3] = byte(m.dstDeviceAddr) // 应用服务数据单元地址
	asdu[4] = 0xFE                  // 功能类型
	asdu[5] = infoNum               // 信息序号: 由参数指定
	asdu[6] = 0x00                  // 返回信息标识符
	asdu[7] = 0x01                  // 通用分类数据集数目
	asdu[8] = groupNum              // 组号
	asdu[9] = entryNum              // 条目号
	asdu[10] = 0x01                 // 描述类别: 实际值
	asdu[11] = 0x09                 // 数据类型: DIP（双点信息）
	asdu[12] = 0x01                 // 数据宽度: 1
	asdu[13] = 0x01                 // 数据数目: 1
	asdu[14] = value                // 控制值: 0x01=分, 0x02=合

	return append(header, asdu...)
}

// BuildControlExecute 构建控制执行命令
func (m *msgBuilder) BuildControlExecute(groupNum, entryNum, value byte) []byte {
	header := m.BuildCommonHeader(IEC103_HeadLen + 15)

	asdu := make([]byte, 15)
	asdu[0] = 0x0A                  // 类型标识: 通用分类数据控制方向
	asdu[1] = 0x81                  // VSQ
	asdu[2] = COT_Write             // 传送原因: 通用分类写命令
	asdu[3] = byte(m.dstDeviceAddr) // 应用服务数据单元地址
	asdu[4] = 0xFE                  // 功能类型
	asdu[5] = INF_WriteWithExecute  // 信息序号: 带执行的写条目
	asdu[6] = 0x00                  // 返回信息标识符
	asdu[7] = 0x01                  // 通用分类数据集数目
	asdu[8] = groupNum              // 组号
	asdu[9] = entryNum              // 条目号
	asdu[10] = 0x01                 // 描述类别: 实际值
	asdu[11] = 0x09                 // 数据类型: DIP（双点信息）
	asdu[12] = 0x01                 // 数据宽度: 1
	asdu[13] = 0x01                 // 数据数目: 1
	asdu[14] = value                // 控制值

	return append(header, asdu...)
}

// BuildCommonHeader 构建通用报文头
func (m *msgBuilder) BuildCommonHeader(length uint32) []byte {
	data := make([]byte, IEC103_HeadLen)

	binary.LittleEndian.PutUint16(data[0:2], 0xEB90)
	binary.LittleEndian.PutUint32(data[2:6], length)
	binary.LittleEndian.PutUint16(data[6:8], 0xEB90)
	binary.LittleEndian.PutUint16(data[8:10], m.srcStationAddr)
	binary.LittleEndian.PutUint16(data[10:12], m.srcDeviceAddr)
	binary.LittleEndian.PutUint16(data[12:14], m.dstStationAddr)
	binary.LittleEndian.PutUint16(data[14:16], m.dstDeviceAddr)
	binary.LittleEndian.PutUint16(data[16:18], m.GetNextSequence())
	binary.LittleEndian.PutUint16(data[18:20], 0x0000)
	binary.LittleEndian.PutUint16(data[20:22], 0x1000)
	binary.LittleEndian.PutUint16(data[22:24], 0x0000)
	binary.LittleEndian.PutUint16(data[24:26], 0x0000)
	binary.LittleEndian.PutUint16(data[26:28], 0xFFFF)

	return data
}

// ParseSeqNum 解析序列号
func (m *msgBuilder) ParseSeqNum(data []byte) uint16 {
	if len(data) < IEC103_HeadLen {
		return 0
	}
	return binary.LittleEndian.Uint16(data[16:18])
}

// ParseControlExtend 解析控制扩展字段
func (m *msgBuilder) ParseControlExtend(extend string) (uint8, uint8, error) {
	parts := strings.Split(extend, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid control extend format: %s", extend)
	}

	group, err := strconv.ParseUint(parts[0], 10, 8)
	if err != nil {
		return 0, 0, err
	}

	entry, err := strconv.ParseUint(parts[1], 10, 8)
	if err != nil {
		return 0, 0, err
	}

	return uint8(group), uint8(entry), nil
}

// BuildClockSync 构建时钟同步请求
func (m *msgBuilder) BuildClockSync(t time.Time, summerTime bool) []byte {
	header := m.BuildCommonHeader(IEC103_HeadLen + 13)
	ms := t.Nanosecond()/1000000 + t.Second()*1000
	minute := t.Minute()
	hour := t.Hour()
	day := t.Day()
	week := t.Weekday()
	month := t.Month()
	year := t.Year()

	asdu := make([]byte, 13)
	asdu[0] = ASDU_Type_ClockSync   // 类型标识
	asdu[1] = 0x81                  // VSQ
	asdu[2] = COT_ClockSync         // 传送原因: 时钟同步
	asdu[3] = byte(m.dstDeviceAddr) // 应用服务数据单元地址
	asdu[4] = 0xFF                  // 功能类型
	asdu[5] = 0x00
	binary.LittleEndian.PutUint16(asdu[6:8], uint16(ms))
	asdu[8] = byte(minute)
	hourByte := byte(hour & 0x1F)
	if summerTime {
		hourByte |= 0x80 // 设置夏时制位
	}
	asdu[9] = hourByte
	weekAndDay := byte(week)<<5 | byte(day)
	asdu[10] = weekAndDay
	asdu[11] = byte(month)
	asdu[12] = byte(year - 2000)
	// 打印报文内容和具体的设置进去的值
	log.Infof("set clock sync: ms=%d minute=%d hour=%d day=%d week=%d month=%d year=%d，summer:%v, pak:% X",
		ms, minute, hour, day, week, month, year, summerTime, asdu)

	return append(header, asdu...)
}
