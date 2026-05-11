package fdm

import (
	"context"
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
	// 默认设备地址
	DefaultDeviceAddr = 16
	// 默认重试次数
	DefaultRetries = 3
	// 默认允许连续失败次数
	DefaultAllowFailTimes = 3
)

// FDM协议常量
const (
	// 起始符
	StartChar byte = 0x40 // '@'
	// 结束符
	EndChar byte = 0x0D // '\r'

	// 功能码1 - 读数据
	FuncCode1Read = "08"
	// 功能码1 - 写数据
	FuncCode1Write = "09"

	// 功能码2 - 读气体浓度
	FuncCodeReadGasConcentration uint16 = 4
	// 功能码2 - 读探头内部温度
	FuncCodeReadTemperature uint16 = 8
	// 功能码2 - 读第二报警点数值
	FuncCodeReadAlarm2 uint16 = 12
	// 功能码2 - 读第一报警点值
	FuncCodeReadAlarm1 uint16 = 16
	// 功能码2 - 满量程值
	FuncCodeReadFullScale uint16 = 32
	// 功能码2 - 小数点位数
	FuncCodeReadDecimalPoints uint16 = 80
)

// 共享串口连接
var serialPortMap sync.Map

// sharedSerialPort 共享串口结构体
type sharedSerialPort struct {
	port     serial.Port
	refCount int
	mu       sync.Mutex
}

// FDMDevice FDM气体探测器设备
type FDMDevice struct {
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

// NewFDMDevice 创建FDM设备
func NewFDMDevice() *FDMDevice {
	return &FDMDevice{
		port:             nil,
		requestFailTimes: 0,
	}
}

// Open 打开通道
func (d *FDMDevice) Open(chanInfo model.ChannelInfo, packets model.ListCollectPackets) consts.Quality {
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
		log.Errorf("FDM open serial port error: %v", err)
		return consts.QualityCommDisconnected
	}

	if d.port == nil {
		return consts.QualityCommDisconnected
	}

	return consts.QualityOk
}

// getSerialPortAndConnect 获取串口连接
func (d *FDMDevice) getSerialPortAndConnect(chanInfo model.ChannelInfo) (serial.Port, error) {
	// 解析串口参数
	baudRate, parity, dataBits, stopBits, err := model.ParseRTUParam(chanInfo.Params)
	if err != nil {
		// 使用默认参数：9600:N:8:1
		baudRate = 9600
		parity = "N"
		dataBits = 8
		stopBits = 1
		log.Warnf("FDM parse params error: %v, use default 9600:N:8:1", err)
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
func (d *FDMDevice) closeInternal() {
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
func (d *FDMDevice) Close() consts.Quality {
	if d == nil {
		return consts.QualityOk
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.closeInternal()
	return consts.QualityOk
}

// Request 发送采集指令
func (d *FDMDevice) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model2.MessageStatistics) {
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

	cmd, err := parseIntStr(packet.Command)
	if err != nil {
		return consts.QualityConfigError, statics
	}
	// 遍历所有测点进行采集
	for i := range packet.Points {
		statics.SendCount++
		// 构建并发送请求
		value, qua, err := d.readData(d.deviceAddr, uint16(cmd))
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
func (d *FDMDevice) readData(deviceAddr uint8, funcCode uint16) (float64, consts.Quality, error) {
	var lastErr error

	for i := 0; i <= d.option.ReadRetries; i++ {
		// 构建请求包
		reqPacket := d.buildRequestPacket(deviceAddr, funcCode)

		// 清空接收缓冲区
		d.port.ResetInputBuffer()

		// 发送请求
		_, err := d.port.Write(reqPacket)
		if err != nil {
			lastErr = fmt.Errorf("write error: %w, req:% X", err, reqPacket)
			continue
		}

		// 读取响应
		respData, err := d.readResponse()
		if err != nil {
			lastErr = fmt.Errorf("read error: %w, req:% X", err, reqPacket)
			continue
		}

		// 解析响应
		value, err := d.parseResponse(respData, deviceAddr)
		if err != nil {
			lastErr = fmt.Errorf("parse error: %w", err)
			continue
		}

		return value, consts.QualityOk, nil
	}

	return 0, consts.QualityCmdRespError, lastErr
}

// buildRequestPacket 构建请求数据包
// 格式: 起始符(1) + 地址(2) + 功能码1(2) + 功能码2(4) + 字节长度(2) + 校验码(2) + 结束符(1)
func (d *FDMDevice) buildRequestPacket(deviceAddr uint8, funcCode uint16) []byte {
	packet := make([]byte, 0, 14)

	// 起始符 '@'
	packet = append(packet, StartChar)

	// 设备地址 (2字符ASCII)
	packet = append(packet, byteToASCII(deviceAddr)...)

	// 功能码1 - 读数据 "08"
	packet = append(packet, []byte(FuncCode1Read)...)

	// 功能码2 (4字符ASCII)
	packet = append(packet, uint16ToASCII(funcCode)...)

	// 字节长度 "04" (数据区4字节)
	packet = append(packet, []byte("04")...)

	// 计算校验码 (从地址开始到字节长度结束)
	checksum := calculateChecksum(packet[1:])
	packet = append(packet, byteToASCII(checksum)...)

	// 结束符
	packet = append(packet, EndChar)

	return packet
}

// readResponse 读取响应数据
func (d *FDMDevice) readResponse() ([]byte, error) {
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

			// 检查是否收到结束符
			if len(result) > 0 && result[len(result)-1] == EndChar {
				break
			}
		}
	}

	if len(result) == 0 {
		return nil, errors.New("read timeout, no data received")
	}

	// 验证起始符
	if result[0] != StartChar {
		return nil, fmt.Errorf("invalid start char: 0x%02X", result[0])
	}

	return result, nil
}

// parseResponse 解析响应数据
// 响应格式: 起始符(1) + 地址(2) + 字节长度(2) + 数据区(8) + 校验码(2) + 结束符(1)
func (d *FDMDevice) parseResponse(data []byte, expectedAddr uint8) (float64, error) {
	// 最小长度检查: 1 + 2 + 2 + 8 + 2 + 1 = 16
	if len(data) < 16 {
		return 0, fmt.Errorf("response too short: %d bytes", len(data))
	}

	// 验证地址
	addrHigh, err := asciiToByte(data[1])
	if err != nil {
		return 0, fmt.Errorf("parse addr high error: %w", err)
	}
	addrLow, err := asciiToByte(data[2])
	if err != nil {
		return 0, fmt.Errorf("parse addr low error: %w", err)
	}
	addr := (addrHigh << 4) | addrLow
	if addr != expectedAddr {
		return 0, fmt.Errorf("address mismatch: expected %d, got %d", expectedAddr, addr)
	}

	// 解析字节长度
	lenHigh, _ := asciiToByte(data[3])
	lenLow, _ := asciiToByte(data[4])
	dataLen := int((lenHigh << 4) | lenLow)
	if dataLen != 4 {
		return 0, fmt.Errorf("unexpected data length: %d", dataLen)
	}

	// 提取数据区 (8个ASCII字符，表示4字节浮点数)
	if len(data) < 5+8+2+1 {
		return 0, fmt.Errorf("response data incomplete")
	}
	dataArea := data[5 : 5+8]

	// 验证校验码
	checksumStart := 5 + 8
	expectedChecksum := calculateChecksum(data[1:checksumStart])
	actualChecksumHigh, _ := asciiToByte(data[checksumStart])
	actualChecksumLow, _ := asciiToByte(data[checksumStart+1])
	actualChecksum := (actualChecksumHigh << 4) | actualChecksumLow
	if expectedChecksum != actualChecksum {
		return 0, fmt.Errorf("checksum mismatch: expected 0x%02X, got 0x%02X", expectedChecksum, actualChecksum)
	}

	// 解析浮点数数据
	return parseFloatData(dataArea)
}

// parseFloatData 解析浮点数数据
// 根据协议文档，数据区为8个ASCII字符，表示4字节浮点数
// 第一字节: D7=数符(0正1负), D6=阶符(0正1负), D5-D0=阶码
// 第二、三、四字节: 小数部分
func parseFloatData(data []byte) (float64, error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("invalid float data length: %d", len(data))
	}

	// 将8个ASCII字符转换为4个字节
	bytes := make([]byte, 4)
	for i := 0; i < 4; i++ {
		high, err := asciiToByte(data[i*2])
		if err != nil {
			return 0, fmt.Errorf("parse float byte %d high error: %w", i, err)
		}
		low, err := asciiToByte(data[i*2+1])
		if err != nil {
			return 0, fmt.Errorf("parse float byte %d low error: %w", i, err)
		}
		bytes[i] = (high << 4) | low
	}

	// 第一字节解析
	signBit := (bytes[0] >> 7) & 0x01    // 数符
	expSignBit := (bytes[0] >> 6) & 0x01 // 阶符
	expValue := int(bytes[0] & 0x3F)     // 阶码 (6位)

	// 小数部分: A2, A3, A4
	a2 := float64(bytes[1])
	a3 := float64(bytes[2])
	a4 := float64(bytes[3])

	// 计算数值: ((((A4/256)+A3)/256)+A2)/256 × 2^阶码
	mantissa := ((((a4 / 256.0) + a3) / 256.0) + a2) / 256.0

	// 计算2的阶码次方
	var exp float64 = 1.0
	if expSignBit == 0 {
		// 正阶
		for i := 0; i < expValue; i++ {
			exp *= 2.0
		}
	} else {
		// 负阶
		for i := 0; i < expValue; i++ {
			exp /= 2.0
		}
	}

	result := mantissa * exp

	// 应用数符
	if signBit == 1 {
		result = -result
	}

	return result, nil
}

// RequestPing 发送采集指令，最小化指令发送包
func (d *FDMDevice) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	qua, _ := d.Request(ctx, &packet)
	return qua
}

// Control 发送控制指令
func (d *FDMDevice) Control(packet *model.ControlProtocolPacket, val string) consts.Quality {
	// FDM协议目前只开放读气体浓度功能，写功能为锁定状态
	log.Warnf("FDM Control not supported, write function is locked")
	return consts.QualityConfigError
}

// setPointValueErrorQua 设置测点错误质量
func setPointValueErrorQua(point *model.PointInfo, qua consts.Quality) {
	point.RtVal.Qua = qua
	point.RtVal.Tms = utils.GetNowUTCTimeStamp()
}

// byteToASCII 将单字节转换为2字节ASCII
func byteToASCII(b byte) []byte {
	high := b >> 4
	low := b & 0x0F
	return []byte{
		nibbleToASCII(high),
		nibbleToASCII(low),
	}
}

// uint16ToASCII 将uint16转换为4字节ASCII
func uint16ToASCII(v uint16) []byte {
	return []byte{
		nibbleToASCII(byte((v >> 12) & 0x0F)),
		nibbleToASCII(byte((v >> 8) & 0x0F)),
		nibbleToASCII(byte((v >> 4) & 0x0F)),
		nibbleToASCII(byte(v & 0x0F)),
	}
}

// nibbleToASCII 将4位值转换为ASCII字符
func nibbleToASCII(n byte) byte {
	if n < 10 {
		return '0' + n
	}
	return 'A' + (n - 10)
}

// asciiToByte 将ASCII字符转换为4位值
func asciiToByte(c byte) (byte, error) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	default:
		return 0, fmt.Errorf("invalid ASCII char: 0x%02X", c)
	}
}

// calculateChecksum 计算校验码
// 从第二个字符开始到数据区最后一个字符的所有ASCII码按字节求异或
func calculateChecksum(data []byte) byte {
	var checksum byte = 0
	for _, b := range data {
		checksum ^= b
	}
	return checksum
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
