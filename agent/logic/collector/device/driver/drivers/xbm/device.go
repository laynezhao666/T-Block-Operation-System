package xbm

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/consts"
	model2 "agent/entity/model"
	"agent/logic/collector/device/model"
	"agent/utils"
)

const (
	// 默认超时时间
	DefaultTimeout = 2 * time.Second
	// 默认重试次数
	DefaultRetries = 3
	// 默认允许连续失败次数
	DefaultAllowFailTimes = 3
	// 默认设备地址
	DefaultDeviceAddr = 1
)

// XBM协议常量
const (
	// 帧头
	FrameHeader byte = 0xAB
	// 最小帧长度: Header(1) + TransformerAddr(1) + SensorAddr(2) + Command(1) + CRC(2) = 7
	MinFrameLength = 7
	// 响应数据帧最小长度: Header(1) + TransformerAddr(1) + SensorAddr(2) + Command(1) + Data(2) + CRC(2) = 9
	MinResponseLength = 9
)

// CRC-16查找表
var crcTable = []uint16{
	0x0000, 0x1021, 0x2042, 0x3063, 0x4084, 0x50a5, 0x60c6, 0x70e7,
	0x8108, 0x9129, 0xa14a, 0xb16b, 0xc18c, 0xd1ad, 0xe1ce, 0xf1ef,
}

// 共享串口连接
var serialPortMap sync.Map

// sharedSerialPort 共享串口结构体
type sharedSerialPort struct {
	port     serial.Port
	refCount int
	mu       sync.Mutex
}

// XBMDevice XBM变压器设备
type XBMDevice struct {
	// 串口连接
	port serial.Port
	// 通道信息
	chanInfo model.ChannelInfo
	// 采集指令包列表
	packets model.ListCollectPackets
	// 配置选项
	option Option
	// 请求失败次数
	requestFailTimes int
	// 是否共享传输
	shareTransport bool
	// 互斥锁，保护串口操作
	mu         sync.Mutex
	deviceAddr uint8
}

// Option 配置选项
type Option struct {
	// 读超时时间(毫秒)
	ReadTimeOut int `json:"read_timeout"`
	// 重试次数
	ReadRetries int `json:"read_retries"`
	// 是否共享传输
	ShareTransport bool `json:"share_transport"`
}

// Load 加载配置
func (o *Option) Load(chanInfo model.ChannelInfo, packets model.ListCollectPackets) {
	// 设置默认值
	o.ReadTimeOut = int(DefaultTimeout.Milliseconds())
	o.ReadRetries = DefaultRetries
	o.ShareTransport = true

	// 从通道超时设置
	if chanInfo.TimeoutMs > 0 {
		o.ReadTimeOut = chanInfo.TimeoutMs
	}
}

// NewXBMDevice 创建XBM设备
func NewXBMDevice() *XBMDevice {
	return &XBMDevice{
		port:             nil,
		requestFailTimes: 0,
	}
}

// Open 打开通道
func (d *XBMDevice) Open(chanInfo model.ChannelInfo, packets model.ListCollectPackets) consts.Quality {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 先清理旧资源（防止多次调用Open导致资源泄露）
	d.closeInternal()

	d.chanInfo = chanInfo
	d.packets = packets
	d.requestFailTimes = 0
	d.option.Load(chanInfo, packets)
	d.shareTransport = d.option.ShareTransport
	// 获取设备地址，调用字符串转换函数，支持十六进制和十进制转换为整型
	addr, err := parseIntStr(chanInfo.Address)
	if err != nil {
		log.Errorf("FDM parse device address error: %v", err)
		return consts.QualityConfigError
	}
	d.deviceAddr = uint8(addr)
	d.port, err = d.getSerialPortAndConnect(chanInfo)
	if err != nil {
		log.Errorf("XBM open serial port error: %v", err)
		return consts.QualityCommDisconnected
	}

	if d.port == nil {
		return consts.QualityCommDisconnected
	}

	return consts.QualityOk
}

// getSerialPortAndConnect 获取串口连接
func (d *XBMDevice) getSerialPortAndConnect(chanInfo model.ChannelInfo) (serial.Port, error) {
	// 解析串口参数，XBM协议默认：9600:N:8:1
	baudRate, parity, dataBits, stopBits, err := model.ParseRTUParam(chanInfo.Params)
	if err != nil {
		// 使用XBM协议默认参数：9600:N:8:1
		baudRate = 9600
		parity = "N"
		dataBits = 8
		stopBits = 1
		log.Warnf("XBM parse params error: %v, use default 9600:N:8:1", err)
	}

	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   model.ParseParity(parity),
		DataBits: dataBits,
		StopBits: model.MapStopBits(stopBits),
	}

	if d.shareTransport {
		key := chanInfo.Name
		v, ok := serialPortMap.Load(key)
		if ok {
			shared := v.(*sharedSerialPort)
			shared.mu.Lock()
			shared.refCount++
			shared.mu.Unlock()
			return shared.port, nil
		}

		// 创建新连接
		port, err := serial.Open(chanInfo.Name, mode)
		if err != nil {
			return nil, fmt.Errorf("open serial port %s error: %w", chanInfo.Name, err)
		}

		// 设置超时
		if err := port.SetReadTimeout(time.Duration(d.option.ReadTimeOut) * time.Millisecond); err != nil {
			port.Close()
			return nil, fmt.Errorf("set read timeout error: %w", err)
		}

		serialPortMap.Store(key, &sharedSerialPort{
			port:     port,
			refCount: 1,
		})
		return port, nil
	}

	// 非共享模式，直接创建连接
	port, err := serial.Open(chanInfo.Name, mode)
	if err != nil {
		return nil, fmt.Errorf("open serial port %s error: %w", chanInfo.Name, err)
	}

	// 设置超时
	if err := port.SetReadTimeout(time.Duration(d.option.ReadTimeOut) * time.Millisecond); err != nil {
		port.Close()
		return nil, fmt.Errorf("set read timeout error: %w", err)
	}

	return port, nil
}

// closeInternal 内部关闭方法（不加锁）
func (d *XBMDevice) closeInternal() {
	if d.port == nil {
		return
	}

	if d.shareTransport {
		key := d.chanInfo.Name
		if v, ok := serialPortMap.Load(key); ok {
			shared := v.(*sharedSerialPort)
			shared.mu.Lock()
			shared.refCount--
			if shared.refCount <= 0 {
				shared.port.Close()
				serialPortMap.Delete(key)
			}
			shared.mu.Unlock()
		}
	} else {
		d.port.Close()
	}
	d.port = nil
}

// Close 关闭通道
func (d *XBMDevice) Close() consts.Quality {
	if d == nil {
		return consts.QualityOk
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.closeInternal()
	return consts.QualityOk
}

// Request 发送采集指令
func (d *XBMDevice) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model2.MessageStatistics) {
	statics := model2.MessageStatistics{SendCount: 0, SuccessCount: 0}

	if packet == nil {
		return consts.QualityOk, statics
	}
	if d == nil || d.port == nil {
		return consts.QualityCommDisconnected, statics
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	currentTime := utils.GetNowUTCTimeStamp()
	comm, err := parseIntStr(packet.Command)
	if err != nil {
		return consts.QualityConfigError, statics
	}
	// 遍历所有测点进行采集
	for i := range packet.Points {
		parser, err := GetPointValParser(packet.Points[i])
		if err != nil {
			log.Errorf("XBM get value parser error: %v", err)
			setPointValueErrorQua(packet.Points[i], consts.QualityConfigError)
			continue
		}

		statics.SendCount++

		// 构建并发送请求
		value, qua, err := d.readData(d.deviceAddr, parser.SensorAddr, uint8(comm))
		if err != nil {
			statics.ErrLog = err
			setPointValueErrorQua(packet.Points[i], qua)
			continue
		}

		statics.SuccessCount++

		// 设置测点值
		packet.Points[i].RtVal.Pv.SetValue(value)
		packet.Points[i].RtVal.Qua = consts.QualityOk
		packet.Points[i].RtVal.Tms = currentTime
	}

	// 判断整体质量
	if statics.SuccessCount == 0 && statics.SendCount > 0 {
		d.requestFailTimes++
		if d.requestFailTimes >= DefaultAllowFailTimes {
			// 连续失败，返回通信中断，上层会调用Close和Open重连
			return consts.QualityCommDisconnected, statics
		}
		return consts.QualityCmdRespError, statics
	}

	d.requestFailTimes = 0
	return consts.QualityOk, statics
}

// readData 读取数据
func (d *XBMDevice) readData(transformerAddr byte, sensorAddr uint16, command byte) (float64, consts.Quality, error) {
	var lastErr error

	for i := 0; i <= d.option.ReadRetries; i++ {
		// 构建请求包
		reqPacket := BuildRequestPacket(transformerAddr, sensorAddr, command)

		// 清空接收缓冲区
		d.port.ResetInputBuffer()

		// 发送请求
		_, err := d.port.Write(reqPacket)
		if err != nil {
			lastErr = fmt.Errorf("write error: %w", err)
			continue
		}

		// 读取响应
		respData, err := d.readResponse()
		if err != nil {
			lastErr = fmt.Errorf("read error: %w", err)
			continue
		}

		// 解析响应
		value, err := ParseResponse(respData, transformerAddr, sensorAddr, command)
		if err != nil {
			lastErr = fmt.Errorf("parse error: %w", err)
			continue
		}

		return value, consts.QualityOk, nil
	}

	return 0, consts.QualityCmdRespError, lastErr
}

// BuildRequestPacket 构建请求数据包
// 格式: Header(0xAB) + TransformerAddr(1) + SensorAddr(2, little-endian) + Command(1) + CRC(2, little-endian)
func BuildRequestPacket(transformerAddr byte, sensorAddr uint16, command byte) []byte {
	packet := make([]byte, 0, MinFrameLength)

	// 帧头 0xAB
	packet = append(packet, FrameHeader)

	// 变压器地址 (1字节)
	packet = append(packet, transformerAddr)

	// 传感器地址 (2字节，小端序)
	sensorAddrBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(sensorAddrBytes, sensorAddr)
	packet = append(packet, sensorAddrBytes...)

	// 命令码 (1字节)
	packet = append(packet, command)

	// 计算CRC (从帧头开始计算)
	crc := CalculateCRC(packet)
	crcBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(crcBytes, crc)
	packet = append(packet, crcBytes...)

	return packet
}

// readResponse 读取响应数据
func (d *XBMDevice) readResponse() ([]byte, error) {
	buf := make([]byte, 64)
	result := make([]byte, 0, 64)

	// 设置读取超时
	deadline := time.Now().Add(time.Duration(d.option.ReadTimeOut) * time.Millisecond)

	for time.Now().Before(deadline) {
		n, err := d.port.Read(buf)
		if err != nil {
			if err == io.EOF {
				continue
			}
			return nil, err
		}

		if n > 0 {
			result = append(result, buf[:n]...)

			// 检查是否收到完整帧
			if len(result) >= MinResponseLength {
				// 验证帧头
				if result[0] != FrameHeader {
					return nil, fmt.Errorf("invalid frame header: 0x%02X", result[0])
				}
				// 根据命令类型判断数据长度
				// 对于获取数据命令(0x50-0x54, 0x70-0x75)，数据区为2字节
				// 总长度 = 7 + 数据长度
				break
			}
		}
	}

	if len(result) == 0 {
		return nil, errors.New("read timeout, no data received")
	}

	if len(result) < MinFrameLength {
		return nil, fmt.Errorf("response too short: %d bytes", len(result))
	}

	// 验证帧头
	if result[0] != FrameHeader {
		return nil, fmt.Errorf("invalid frame header: 0x%02X", result[0])
	}

	return result, nil
}

// ParseResponse 解析响应数据
// 响应格式: Header(1) + TransformerAddr(1) + SensorAddr(2) + Command(1) + Data(n) + CRC(2)
func ParseResponse(data []byte, expectedTransformerAddr byte, expectedSensorAddr uint16, expectedCommand byte) (float64, error) {
	if len(data) < MinResponseLength {
		return 0, fmt.Errorf("response too short: %d bytes", len(data))
	}

	// 验证帧头
	if data[0] != FrameHeader {
		return 0, fmt.Errorf("invalid frame header: 0x%02X", data[0])
	}

	// 验证变压器地址
	transformerAddr := data[1]
	if transformerAddr != expectedTransformerAddr {
		return 0, fmt.Errorf("transformer address mismatch: expected 0x%02X, got 0x%02X", expectedTransformerAddr, transformerAddr)
	}

	// 验证传感器地址
	sensorAddr := binary.LittleEndian.Uint16(data[2:4])
	if sensorAddr != expectedSensorAddr {
		return 0, fmt.Errorf("sensor address mismatch: expected 0x%04X, got 0x%04X", expectedSensorAddr, sensorAddr)
	}

	// 检查命令码
	command := data[4]

	// 检查是否为异常响应
	if command == CmdAbnormalResponse {
		if len(data) >= 8 {
			abnormalCode := data[5]
			return 0, fmt.Errorf("abnormal response, original command: 0x%02X", abnormalCode)
		}
		return 0, errors.New("abnormal response received")
	}

	// 验证命令码
	if command != expectedCommand {
		return 0, fmt.Errorf("command mismatch: expected 0x%02X, got 0x%02X", expectedCommand, command)
	}

	// 获取数据区（2字节）
	if len(data) < 9 {
		return 0, fmt.Errorf("response data incomplete, length: %d", len(data))
	}
	dataValue := binary.LittleEndian.Uint16(data[5:7])

	// 验证CRC
	crcStart := len(data) - 2
	receivedCRC := binary.LittleEndian.Uint16(data[crcStart:])
	calculatedCRC := CalculateCRC(data[:crcStart])
	if receivedCRC != calculatedCRC {
		return 0, fmt.Errorf("CRC mismatch: expected 0x%04X, got 0x%04X", calculatedCRC, receivedCRC)
	}

	// 根据命令类型转换值
	divisor, _ := GetValueFormula(command)
	return float64(dataValue) / divisor, nil
}

// CalculateCRC 计算CRC-16校验码
// 根据协议文档中的CRC算法实现
func CalculateCRC(data []byte) uint16 {
	var crc uint16 = 0
	var da byte

	for _, b := range data {
		da = byte(crc>>8) >> 4
		crc <<= 4
		crc ^= crcTable[da^(b>>4)]

		da = byte(crc>>8) >> 4
		crc <<= 4
		crc ^= crcTable[da^(b&0x0F)]
	}

	return crc
}

// RequestPing 发送采集指令，最小化指令发送包
func (d *XBMDevice) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	qua, _ := d.Request(ctx, &packet)
	return qua
}

// Control 发送控制指令
func (d *XBMDevice) Control(packet *model.ControlProtocolPacket, val string) consts.Quality {
	// XBM协议控制指令暂不支持
	log.Warnf("XBM Control not supported yet")
	return consts.QualityConfigError
}

// setPointValueErrorQua 设置测点错误质量
func setPointValueErrorQua(point *model.PointInfo, qua consts.Quality) {
	point.RtVal.Qua = qua
	point.RtVal.Tms = utils.GetNowUTCTimeStamp()
}

// parseIntStr 解析设备地址字符串，支持十六进制和十进制格式
// 支持格式：
// - 十进制： "16"
// - 十六进制： "0x10"
func parseIntStr(addrStr string) (uint64, error) {
	if addrStr == "" {
		return DefaultDeviceAddr, nil
	}

	// 去除前后空格
	addrStr = strings.TrimSpace(addrStr)

	// 检查十六进制格式
	if strings.HasPrefix(addrStr, "0x") || strings.HasPrefix(addrStr, "0X") {
		// 0x前缀的十六进制
		hexStr := addrStr[2:]
		val, err := strconv.ParseUint(hexStr, 16, 8)
		if err != nil {
			return 0, fmt.Errorf("invalid hex address '%s': %w", addrStr, err)
		}
		return val, nil
	}

	// 尝试解析为十进制（默认格式）
	val, err := strconv.ParseUint(addrStr, 10, 8)
	if err != nil {
		log.Warnf("FDM address '%s' parsed as hex (0x%X), consider using 0x prefix for clarity", addrStr, val)
		return val, nil
	}

	return val, nil
}
