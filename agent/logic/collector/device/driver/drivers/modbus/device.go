package modbus

import (
	"agent/utils"
	ubytes "agent/utils/bytes"
	"agent/utils/encoding"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goburrow/modbus"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/consts"
	"agent/entity/definition/datatype"
	model2 "agent/entity/model"
	"agent/logic/collector/device/model"
)

const (
	DefaultTimeout                  = 2 * time.Second
	DefaultSlaveId                  = 1
	defaultWriteRetryInterval       = 100 * time.Millisecond
	defaultAllowModbusReadFailTimes = 3
)

const (
	CodeReadCoils              = "01"
	CodeReadDiscreteInputs     = "02"
	CodeReadHoldingRegisters   = "03"
	CodeReadInputRegisters     = "04"
	CodeWriteMultipleRegisters = "10"
	CodeWriteSingleRegister    = "06"
	CodeWriteSingleCoil        = "05"
)

const (
	extFunNot = "not" // bool量取反扩展函数
)

const (
	InvalidValueUint8  = 0x80
	InvalidValueUint16 = 0x8000
	InvalidValueUint32 = 0x80000000
	InvalidValueUint64 = 0x8000000000000000
)

var transportMap sync.Map

// ModbusDevice Modbus设备
type ModbusDevice struct {
	clientHandler modbus.ClientHandler
	client        modbus.Client
	option        Option
	invalidValue  InvalidValue

	chanInfo         model.ChannelInfo
	packets          model.ListCollectPackets
	requestFailTimes int
}

// InvalidValue 无效值定义
type InvalidValue struct {
	Bit8  map[uint8]struct{}
	Bit16 map[uint16]struct{}
	Bit32 map[uint32]struct{}
	Bit64 map[uint64]struct{}
}

// NewModbusDevice 创建Modbus设备
func NewModbusDevice() *ModbusDevice {
	return &ModbusDevice{
		clientHandler: nil,
		client:        nil,
		invalidValue: InvalidValue{
			Bit8:  map[uint8]struct{}{InvalidValueUint8: {}},
			Bit16: map[uint16]struct{}{InvalidValueUint16: {}},
			Bit32: map[uint32]struct{}{InvalidValueUint32: {}},
			Bit64: map[uint64]struct{}{InvalidValueUint64: {}},
		},
	}
}

// Open 建立连接
func (d *ModbusDevice) Open(chanInfo model.ChannelInfo, packets model.ListCollectPackets) consts.Quality {
	d.chanInfo = chanInfo
	d.packets = packets
	d.requestFailTimes = 0
	d.option.Load(chanInfo, packets)

	var err error
	d.clientHandler, err = d.getTransportAndConnect(chanInfo)
	if d.clientHandler == nil {
		return consts.QualityCommDisconnected
	}
	d.client = modbus.NewClient(d.clientHandler)
	d.ParseChanExtend(chanInfo)

	if err != nil {
		return consts.QualityConfigError
	}

	return consts.QualityOk
}

// ParseChanExtend 解析通道扩展参数
func (d *ModbusDevice) ParseChanExtend(chanInfo model.ChannelInfo) {
	if len(chanInfo.DriverExtend) == 0 {
		return
	}

	var result map[string]string
	err := json.Unmarshal([]byte(chanInfo.DriverExtend), &result)
	if err != nil {
		return
	}
	invalidValuesStr, ok := result["invalid_values"]
	if !ok {
		return
	}
	if !d.ParseInvalidValueDefine(invalidValuesStr) {
		log.Warnf("ParseInvalidValueDefine fail")
	}
}

// ParseInvalidValueDefine 解析无效值定义
func (d *ModbusDevice) ParseInvalidValueDefine(invalidValues string) bool {
	items := strings.Split(invalidValues, ",")
	tempBuf := make([]byte, 8)
	tmpInvalidValue := InvalidValue{
		Bit8:  make(map[uint8]struct{}),
		Bit16: make(map[uint16]struct{}),
		Bit32: make(map[uint32]struct{}),
		Bit64: make(map[uint64]struct{}),
	}

	for _, item := range items {
		if !strings.Contains(item, "0x") {
			return false
		}

		item = strings.TrimSpace(item)
		item = strings.ReplaceAll(item, "0x", "")
		var tempLen int
		if !utils.LoadHex(tempBuf, &tempLen, item) {
			log.Warnf("Failed to load hex values:%v\n", item)
			return false
		}

		switch tempLen {
		case 1: // 1 字节
			tmpInvalidValue.Bit8[tempBuf[0]] = struct{}{}
		case 2: // 2 字节， 并且统一以网络字节序解析
			val := (uint16(tempBuf[0]) << 8) | uint16(tempBuf[1])
			tmpInvalidValue.Bit16[val] = struct{}{}
		case 4:
			val := uint32(tempBuf[0])<<24 | uint32(tempBuf[1])<<16 |
				uint32(tempBuf[2])<<8 | uint32(tempBuf[3])
			tmpInvalidValue.Bit32[val] = struct{}{}
		case 8:
			val := uint64(tempBuf[0])<<56 | uint64(tempBuf[1])<<48 |
				uint64(tempBuf[2])<<40 | uint64(tempBuf[3])<<32 |
				uint64(tempBuf[4])<<24 | uint64(tempBuf[5])<<16 |
				uint64(tempBuf[6])<<8 | uint64(tempBuf[7])
			tmpInvalidValue.Bit64[val] = struct{}{}
		default:
			return false
		}
	}
	// 替换原有的无效值配置
	d.invalidValue = tmpInvalidValue
	return true
}

func (d *ModbusDevice) getTransportAndConnect(chanInfo model.ChannelInfo) (modbus.ClientHandler, error) {
	switch chanInfo.ProtocolVer {
	case "RTU":
		return d.getRTUTransportAndConnect(chanInfo)
	case "TCP":
		return d.getTCPTransportAndConnect(chanInfo)
	default:
		log.Errorf("chanInfo.ProtocolVer not support:%s", chanInfo.ProtocolVer)
	}

	return nil, nil

}

func parseRTUParam(params string) (int, string, int, int, error) {
	parts := strings.Split(params, consts.CollectParamSep)
	if len(parts) != 4 {
		return 0, "", 0, 0, fmt.Errorf("param [%v] len err", params)
	}

	// 解析波特率
	baudRate, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", 0, 0, errors.New("baudRate parse err")
	}

	// 解析奇偶校验
	parity := parts[1]

	// 解析数据位
	dataBits, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, "", 0, 0, errors.New("dataBits parse err")
	}

	// 解析停止位
	stopBits, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, "", 0, 0, errors.New("stopBits parse err")
	}

	return baudRate, parity, dataBits, stopBits, nil
}

func (d *ModbusDevice) getRTUTransportAndConnect(chanInfo model.ChannelInfo) (modbus.ClientHandler, error) {
	var rtuHandle *modbus.RTUClientHandler

	if d.option.shareTransport {
		v, ok := transportMap.Load(chanInfo.Name + chanInfo.Address)
		if !ok {
			v, ok := transportMap.Load(chanInfo.Name + chanInfo.Address)
			if !ok {
				rtuHandle = modbus.NewRTUClientHandler(chanInfo.Name)
				transportMap.Store(chanInfo.Name+chanInfo.Address, rtuHandle)
			} else {
				rtuHandle = v.(*modbus.RTUClientHandler)
			}
		} else {
			rtuHandle = v.(*modbus.RTUClientHandler)
		}
	} else {
		rtuHandle = modbus.NewRTUClientHandler(chanInfo.Name)
	}

	// 有参数则赋值，没有使用默认值
	if baudRate, parity, dataBits, stopBits, err := parseRTUParam(chanInfo.Params); err == nil {
		rtuHandle.BaudRate = baudRate
		rtuHandle.Parity = parity
		rtuHandle.DataBits = dataBits
		rtuHandle.StopBits = stopBits
	} else {
		log.Errorf("param parse err:%v", err)
	}

	slaveID := DefaultSlaveId
	if len(chanInfo.Address) > 0 {
		slaverId, err := strconv.Atoi(chanInfo.Address)
		if err != nil {
			return nil, err
		}
		slaveID = slaverId
	}
	rtuHandle.SlaveId = byte(slaveID)
	rtuHandle.Timeout = time.Duration(d.option.ReadTimeOut) * time.Millisecond

	return rtuHandle, rtuHandle.Connect()
}

func (d *ModbusDevice) getTCPTransportAndConnect(chanInfo model.ChannelInfo) (modbus.ClientHandler, error) {
	var tcpHandle *modbus.TCPClientHandler

	if d.option.shareTransport {
		v, ok := transportMap.Load(chanInfo.Name + chanInfo.Address)
		if !ok {
			tcpHandle = modbus.NewTCPClientHandler(chanInfo.Name)
			transportMap.Store(chanInfo.Name+chanInfo.Address, tcpHandle)
		} else {
			tcpHandle = v.(*modbus.TCPClientHandler)
		}
	} else {
		tcpHandle = modbus.NewTCPClientHandler(chanInfo.Name)
	}

	slaveID := DefaultSlaveId
	if len(chanInfo.Address) > 0 {
		slaverId, err := strconv.Atoi(chanInfo.Address)
		if err != nil {
			return nil, err
		}
		slaveID = slaverId
	}
	tcpHandle.SlaveId = byte(slaveID)
	tcpHandle.Timeout = time.Duration(d.option.ReadTimeOut) * time.Millisecond

	return tcpHandle, tcpHandle.Connect()
}

// Close 关闭连接
func (d *ModbusDevice) Close() consts.Quality {
	if d != nil && d.clientHandler != nil {
		var err error
		switch h := d.clientHandler.(type) {
		case *modbus.TCPClientHandler:
			err = h.Close()
		case *modbus.RTUClientHandler:
			err = h.Close()
		}
		if err != nil {
			return consts.QualityCmdRespError
		}
		d.clientHandler = nil
	}
	d.client = nil
	return consts.QualityOk
}

// Request 请求
func (d *ModbusDevice) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model2.MessageStatistics) {
	statics := model2.MessageStatistics{SendCount: 0, SuccessCount: 0}
	if packet == nil {
		return consts.QualityOk, statics
	}
	if d == nil || d.client == nil {
		return consts.QualityCommDisconnected, statics
	}

	funcCode, startAddr, quantity, err := splitCommand(packet.Command)
	if err != nil {
		statics.ErrLog = errors.New(fmt.Sprintf("modbus cmd split err, cmd:%v", packet.Command))
		return consts.QualityConfigError, statics
	}
	// 查询
	statics.SendCount++
	var results []byte
	var qua consts.Quality
	results, qua, err = d.readModbusData(funcCode, startAddr, quantity)
	statics.RecvPackets = encoding.ParseBytesToHex(results)
	if err != nil {
		statics.ErrLog = err
		return qua, statics
	}
	statics.SuccessCount++
	currentTime := utils.GetNowUTCTimeStamp()
	for i := range packet.Points {
		parser, err := GetPointValParser(packet.Points[i])
		if err != nil {
			log.Errorf("value parser error=%v, cmd=%v", err, packet.Command)
			setPointValueErrorQua(packet.Points[i], consts.QualityConfigError)
			continue
		}

		offset := int(parser.Addr - startAddr)
		if offset < 0 {
			setPointValueErrorQua(packet.Points[i], consts.QualityConfigError)
			continue
		}
		var byteIndex int
		switch funcCode {
		case CodeReadCoils, CodeReadDiscreteInputs:
			byteIndex = offset >> 3
		case CodeReadHoldingRegisters, CodeReadInputRegisters:
			byteIndex = offset << 1
		}

		dataByteSize := datatype.GetDataTypeBytes(parser.DataType)
		if dataByteSize < 0 {
			log.Errorf("point date type configured error, cmd=%v, addr=%v, data_type=%v",
				packet.Command, parser.Addr, parser.DataType)
			continue
		}

		if parser.DataType == datatype.BoolType && (funcCode == CodeReadHoldingRegisters ||
			funcCode == CodeReadInputRegisters) {
			dataByteSize++ // 特殊bool占2字节
		}

		if byteIndex < 0 || byteIndex+dataByteSize > len(results) {
			setPointValueErrorQua(packet.Points[i], consts.QualityConfigError)
			continue
		}
		value := results[byteIndex : byteIndex+dataByteSize]

		err = parseValue(value, parser, packet.Points[i], funcCode, offset&0x7, currentTime, &d.invalidValue)
		if err != nil {
			log.Errorf("parse point occur error, err=%v, cmd=%v, point_addr=%v",
				err, packet.Command, parser.Addr)
			continue
		}
		if err = doExtFun(parser, packet.Points[i]); err != nil {
			log.Errorf("point doExtFun error, err: %v, cmd: %v, point addr: %v", err, packet.Command, parser.Addr)
			continue
		}
	}
	return consts.QualityOk, statics
}

func (d *ModbusDevice) readModbusData(funCode string, startAddr uint16, quantity uint16) (
	[]byte, consts.Quality, error) {
	var result []byte
	var err error

	for i := 0; i <= d.option.ReadRetries; i++ {
		switch funCode {
		case CodeReadCoils:
			result, err = d.client.ReadCoils(startAddr, quantity)
		case CodeReadDiscreteInputs:
			result, err = d.client.ReadDiscreteInputs(startAddr, quantity)
		case CodeReadHoldingRegisters:
			result, err = d.client.ReadHoldingRegisters(startAddr, quantity)
		case CodeReadInputRegisters:
			result, err = d.client.ReadInputRegisters(startAddr, quantity)
		default:
			return result, consts.QualityConfigError, errors.New("fun code error")
		}

		if err == nil {
			break
		}
	}

	if err != nil {
		d.requestFailTimes++
		// 连续失败时，尝试重新打开
		if d.requestFailTimes >= defaultAllowModbusReadFailTimes {
			d.Close()
			// 避免频繁重建
			time.Sleep(100 * time.Millisecond)
			d.Open(d.chanInfo, d.packets)
			time.Sleep(100 * time.Millisecond)
			log.Infof("modbus reopen channel %v; request err: %v", d.chanInfo.Name, err)
		}
		return result, consts.QualityCmdRespError, err
	}

	d.requestFailTimes = 0
	return result, consts.QualityOk, nil
}

// RequestPing 发送采集指令，最小化指令发送包
func (d *ModbusDevice) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	qua, _ := d.Request(ctx, &packet)
	return qua
}

// Control 控制
func (d *ModbusDevice) Control(packet *model.ControlProtocolPacket, val string) consts.Quality {
	if d == nil || d.client == nil {
		return consts.QualityCommDisconnected
	}
	if packet == nil || packet.Point == nil {
		return consts.QualityConfigError
	}

	// 解析值转换器
	parser, ok := packet.Point.Attr.ValParser.(*ModbusValueParser)
	if !ok || parser == nil {
		return consts.QualityConfigError
	}

	// 解析原始命令
	cmd := strings.ToUpper(packet.Command)
	if len(cmd) < 6 { // 至少包含功能码和地址
		return consts.QualityConfigError
	}

	// 获取寄存器地址
	addr, err := strconv.ParseUint(cmd[2:6], 16, 16)
	if err != nil {
		return consts.QualityConfigError
	}
	startAddr := uint16(addr)

	// 确定写入功能码
	var writeFuncCode string
	switch {
	case strings.HasPrefix(cmd, CodeReadHoldingRegisters):
		if datatype.GetDataTypeBytes(parser.DataType) > 2 {
			writeFuncCode = CodeWriteMultipleRegisters // 16进制10
		} else {
			writeFuncCode = CodeWriteSingleRegister // 06
		}
	case strings.HasPrefix(cmd, CodeReadCoils):
		writeFuncCode = CodeWriteSingleCoil // 05
	default:
		return consts.QualityConfigError
	}

	// 转换写入值到字节
	data, err := ubytes.ConvertStringToBytes(val, parser.DataType, parser.ByteOrder)
	if err != nil {
		return consts.QualityConfigError
	}

	// 执行Modbus写入操作
	var result []byte
	for i := 0; i <= d.option.WriteRetries; i++ {
		switch writeFuncCode {
		case CodeWriteSingleCoil:
			var value uint16
			if value, err = ubytes.BytesToUint16(data, parser.ByteOrder); err == nil {
				result, err = d.client.WriteSingleCoil(startAddr, value)
			}
		case CodeWriteSingleRegister,
			CodeWriteMultipleRegisters:
			quantity := uint16(len(data) / 2)
			result, err = d.client.WriteMultipleRegisters(startAddr, quantity, data)
		}

		if err == nil {
			break
		}
		time.Sleep(defaultWriteRetryInterval)
	}

	if err != nil {
		return consts.QualityCmdRespError
	}

	// 验证响应数据
	if !verifyResponse(writeFuncCode, startAddr, result) {
		return consts.QualityCmdRespError
	}

	return consts.QualityOk
}

// 协议响应验证
func verifyResponse(funcCode string, addr uint16, resp []byte) bool {
	// todo: 检验回包长度是否符合请求长度
	return true
}

func parseValue(bytes []byte, valParser *ModbusValueParser, point *model.PointInfo, funcCode string, pos int,
	currentTime int64, invalidValue *InvalidValue) error {
	if bytes == nil || valParser == nil || point == nil {
		return errors.New("nil pointer")
	}
	if len(bytes) < datatype.GetDataTypeBytes(valParser.DataType) {
		return errors.New("bytes len too short")
	}

	invalid := false
	var value interface{}
	var err error

	switch valParser.DataType {
	case datatype.BoolType:
		if funcCode == CodeReadCoils || funcCode == CodeReadDiscreteInputs {
			value, err = readBool(bytes[0], pos, valParser.ByteOrder)
		} else {
			value = (valParser.ByteOrder.Uint16(bytes) & (0xffff >> (16 - valParser.BitEnd))) >> valParser.BitBegin
		}
	case datatype.Int8Type, datatype.Uint8Type:
		value = bytes[0]
		invalid = isInvalid(invalidValue.Bit8, value.(uint8))
	case datatype.Int16Type, datatype.Uint16Type:
		value = valParser.ByteOrder.Uint16(bytes)
		invalid = isInvalid(invalidValue.Bit16, value.(uint16))
	case datatype.Int32Type, datatype.Uint32Type:
		value = valParser.ByteOrder.Uint32(bytes)
		invalid = isInvalid(invalidValue.Bit32, value.(uint32))
	case datatype.Int64Type, datatype.Uint64Type:
		value = valParser.ByteOrder.Uint64(bytes)
		invalid = isInvalid(invalidValue.Bit64, value.(uint64))
	case datatype.FloatType:
		temp := valParser.ByteOrder.Uint32(bytes)
		invalid = isInvalid(invalidValue.Bit32, temp)
		value = valParser.ByteOrder.Float(bytes)
	case datatype.DoubleType:
		temp := valParser.ByteOrder.Uint64(bytes)
		invalid = isInvalid(invalidValue.Bit64, temp)
		value = valParser.ByteOrder.Double(bytes)
	default:
		return fmt.Errorf("not supported datatype: %v", valParser.DataType)
	}

	if err != nil {
		return err
	}
	if !invalid {
		point.RtVal.Pv.SetValue(value)
		point.RtVal.Qua = consts.QualityOk
		point.RtVal.Tms = currentTime
	} else {
		point.RtVal.Qua = consts.QualityValueInvalidError
	}
	return nil
}

func isInvalid[T comparable](invalidMap map[T]struct{}, value T) bool {
	_, ok := invalidMap[value]
	return ok
}

func setPointValueErrorQua(point *model.PointInfo, qua consts.Quality) {
	point.RtVal.Qua = qua
	point.RtVal.Tms = utils.GetNowUTCTimeStamp()
}

func doExtFun(valParser *ModbusValueParser, point *model.PointInfo) error {
	if valParser.Extend == extFunNot && point.Attr.Type != model.AnalogType {
		v, err := point.RtVal.Pv.AsBool()
		if err != nil {
			point.RtVal.Qua = consts.QualityConfigError
			return err
		}

		point.RtVal.Pv.SetValue(utils.Bool2Int(!v))
	}
	return nil
}
