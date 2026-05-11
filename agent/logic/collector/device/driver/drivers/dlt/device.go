// Package dlt DL/T 645-2007 多功能电能表通信协议驱动
package dlt

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

// DL/T 645-2007 协议常量
const (
	// 帧起始符
	FrameStart byte = 0x68
	// 帧结束符
	FrameEnd byte = 0x16
	// 数据域加0x33
	DataMask byte = 0x33

	// 控制码
	CtrlCodeRead       byte = 0x11 // 读数据
	CtrlCodeReadMore   byte = 0x12 // 读后续数据
	CtrlCodeReadReply  byte = 0x91 // 读数据正常应答
	CtrlCodeWrite      byte = 0x14 // 写数据
	CtrlCodeWriteReply byte = 0x94 // 写数据正常应答
	CtrlCodeError      byte = 0xD1 // 读数据异常应答

	// 默认超时时间
	DefaultTimeout = 2 * time.Second
	// 默认重试次数
	DefaultRetries = 3
	// 默认允许连续失败次数
	DefaultAllowFailTimes = 3
	// 唤醒帧前导符数量
	WakeUpPreambleCount = 4
	// 唤醒帧前导符
	WakeUpPreamble byte = 0xFE
)

// 共享串口连接
var serialPortMap sync.Map

// sharedSerialPort 共享串口结构体
type sharedSerialPort struct {
	port     serial.Port
	refCount int
	mu       sync.Mutex
}

// DLTDevice DL/T 645-2007 多功能电能表设备
type DLTDevice struct {
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
	mu sync.Mutex
	// 电表地址(6字节BCD码)
	meterAddress []byte
}

// Option 配置选项
type Option struct {
	// 读超时时间(毫秒)
	ReadTimeOut int `json:"read_timeout"`
	// 重试次数
	ReadRetries int `json:"read_retries"`
	// 是否共享传输
	ShareTransport bool `json:"share_transport"`
	// 是否发送唤醒帧
	SendWakeUp bool `json:"send_wakeup"`
	// 响应等待时间(毫秒)，发送请求后等待设备响应的时间
	ResponseDelay int `json:"response_delay"`
	// 是否启用数据域加0x33掩码（DL/T 645-2007标准要求）
	EnableDataMask bool `json:"enable_data_mask"`
}

// Load 加载配置
func (o *Option) Load(chanInfo model.ChannelInfo, packets model.ListCollectPackets) {
	// 设置默认值
	o.ReadTimeOut = int(DefaultTimeout.Milliseconds())
	o.ReadRetries = DefaultRetries
	o.ShareTransport = true
	o.SendWakeUp = true
	o.ResponseDelay = 50    // 默认50毫秒
	o.EnableDataMask = true // 默认启用数据域加0x33（符合DL/T 645-2007标准）

	// 从通道超时设置
	if chanInfo.TimeoutMs > 0 {
		o.ReadTimeOut = chanInfo.TimeoutMs
	}

	// 解析扩展参数
	if v, ok := chanInfo.ExtendKV["share_transport"]; ok && v == "0" {
		o.ShareTransport = false
	}
	if v, ok := chanInfo.ExtendKV["send_wakeup"]; ok && v == "1" {
		o.SendWakeUp = true
	}
	if v, ok := chanInfo.ExtendKV["retry"]; ok {
		if retryVal, err := strconv.Atoi(v); err == nil && retryVal >= 0 {
			o.ReadRetries = retryVal
		}
	}
	if v, ok := chanInfo.ExtendKV["response_delay"]; ok {
		if delayVal, err := strconv.Atoi(v); err == nil && delayVal >= 0 {
			o.ResponseDelay = delayVal
		}
	}
	// enable_data_mask: "0" 禁用数据域加0x33，"1" 或不设置则启用
	if v, ok := chanInfo.ExtendKV["enable_data_mask"]; ok && v == "0" {
		o.EnableDataMask = false
	}
}

// NewDLTDevice 创建DLT设备
func NewDLTDevice() *DLTDevice {
	return &DLTDevice{
		port:             nil,
		requestFailTimes: 0,
		meterAddress:     make([]byte, 6),
	}
}

// Open 打开通道
func (d *DLTDevice) Open(chanInfo model.ChannelInfo, packets model.ListCollectPackets) consts.Quality {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 先清理旧资源
	d.closeInternal()

	d.chanInfo = chanInfo
	d.packets = packets
	d.requestFailTimes = 0
	d.option.Load(chanInfo, packets)
	d.shareTransport = d.option.ShareTransport

	// 解析电表地址
	if err := d.parseMeterAddress(chanInfo.Address); err != nil {
		log.Errorf("DLT645 parse meter address error: %v", err)
		return consts.QualityConfigError
	}

	var err error
	d.port, err = d.getSerialPortAndConnect(chanInfo)
	if err != nil {
		log.Errorf("DLT645 open serial port error: %v", err)
		return consts.QualityCommDisconnected
	}

	if d.port == nil {
		return consts.QualityCommDisconnected
	}

	return consts.QualityOk
}

// parseMeterAddress 解析电表地址
// 地址格式: 12位数字字符串，如 "610769000012"
// 空字符串使用广播地址(0xAA)
// 每2位数字转换为1个BCD字节，低位在前存储到6字节地址域
// 例如: "610769000012" -> BCD [0x12, 0x00, 0x00, 0x69, 0x07, 0x61]
func (d *DLTDevice) parseMeterAddress(addrStr string) error {
	if addrStr == "" {
		// 使用广播地址
		for i := 0; i < 6; i++ {
			d.meterAddress[i] = 0xAA
		}
		return nil
	}

	addrStr = strings.TrimSpace(addrStr)

	// 检查是否全为数字
	for _, c := range addrStr {
		if c < '0' || c > '9' {
			return fmt.Errorf("invalid meter address: %s, must be digits only", addrStr)
		}
	}

	// 补齐到12位（左侧补0）
	if len(addrStr) > 12 {
		return fmt.Errorf("meter address too long: %s, max 12 digits", addrStr)
	}
	for len(addrStr) < 12 {
		addrStr = "0" + addrStr
	}

	// 每2位数字转换为1个BCD字节，低位在前
	// "610769000012" -> [0x12, 0x00, 0x00, 0x69, 0x07, 0x61]
	// 字符串从右往左: "12" "00" "00" "69" "07" "61"
	for i := 0; i < 6; i++ {
		// 从右往左取2位
		idx := 10 - i*2
		high := addrStr[idx] - '0'
		low := addrStr[idx+1] - '0'
		d.meterAddress[i] = (high << 4) | low
	}

	return nil
}

// getSerialPortAndConnect 获取串口连接
func (d *DLTDevice) getSerialPortAndConnect(chanInfo model.ChannelInfo) (serial.Port, error) {
	// 解析串口参数
	baudRate, parity, dataBits, stopBits, err := model.ParseRTUParam(chanInfo.Params)
	if err != nil {
		// DL/T 645-2007 默认参数：2400:E:8:1 或 9600:E:8:1
		baudRate = 2400
		parity = "E"
		dataBits = 8
		stopBits = 1
		log.Warnf("DLT645 parse params error: %v, use default 2400:E:8:1", err)
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
func (d *DLTDevice) closeInternal() {
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
func (d *DLTDevice) Close() consts.Quality {
	if d == nil {
		return consts.QualityOk
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.closeInternal()
	return consts.QualityOk
}

// Request 发送采集指令
func (d *DLTDevice) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model2.MessageStatistics) {
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

	// 解析数据标识（DI0-DI3）
	dataId, err := parseDataIdentifier(packet.Command)
	if err != nil {
		statics.ErrLog = fmt.Errorf("parse data identifier error: %v", err)
		return consts.QualityConfigError, statics
	}

	// 发送读数据请求
	statics.SendCount++
	respData, qua, err := d.readData(dataId)
	if err != nil {
		log.Errorf("DLT645 readData failed, dataId: %08X, error: %v", dataId, err)
		statics.ErrLog = err
		// 设置所有测点错误状态
		for i := range packet.Points {
			setPointValueErrorQua(packet.Points[i], qua)
		}
		return qua, statics
	}

	statics.SuccessCount++

	// 解析响应数据，填充到各个测点
	for i := range packet.Points {
		parser, err := GetPointValParser(packet.Points[i])
		if err != nil {
			log.Errorf("DLT645 get value parser error: %v", err)
			setPointValueErrorQua(packet.Points[i], consts.QualityConfigError)
			continue
		}

		// 根据解析器配置解析数据
		value, err := parser.ParseValue(respData, dataId)
		if err != nil {
			log.Errorf("DLT645 parse value error: %v, dataId: %08X", err, dataId)
			setPointValueErrorQua(packet.Points[i], consts.QualityCmdRespError)
			continue
		}

		packet.Points[i].RtVal.Pv.SetValue(value)
		packet.Points[i].RtVal.Qua = consts.QualityOk
		packet.Points[i].RtVal.Tms = currentTime
	}

	return consts.QualityOk, statics
}

// readData 读取数据
func (d *DLTDevice) readData(dataId uint32) ([]byte, consts.Quality, error) {
	var lastErr error

	for i := 0; i <= d.option.ReadRetries; i++ {
		// 发送唤醒帧（可选）
		if d.option.SendWakeUp {
			d.sendWakeUp()
			//time.Sleep(time.Duration(d.option.ResponseDelay) * time.Millisecond)
		}

		// 构建请求帧
		reqFrame := d.buildReadFrame(dataId)

		// 清空接收缓冲区
		d.port.ResetInputBuffer()

		// 发送请求
		_, err := d.port.Write(reqFrame)
		if err != nil {
			lastErr = fmt.Errorf("write error: %w, req:% X", err, reqFrame)
			continue
		}

		// 等待一小段时间让设备响应
		time.Sleep(time.Duration(d.option.ResponseDelay) * time.Millisecond)

		// 读取响应
		respFrame, err := d.readResponse()
		if err != nil {
			lastErr = fmt.Errorf("read error: %w, req:% X", err, reqFrame)
			continue
		}

		// 解析响应帧
		data, err := d.parseResponseFrame(respFrame)
		if err != nil {
			lastErr = fmt.Errorf("parse error: %w, resp:% X", err, respFrame)
			continue
		}

		return data, consts.QualityOk, nil
	}

	d.requestFailTimes++
	if d.requestFailTimes >= DefaultAllowFailTimes {
		return nil, consts.QualityCommDisconnected, lastErr
	}
	return nil, consts.QualityCmdRespError, lastErr
}

// sendWakeUp 发送唤醒帧
func (d *DLTDevice) sendWakeUp() {
	wakeUp := make([]byte, WakeUpPreambleCount)
	for i := 0; i < WakeUpPreambleCount; i++ {
		wakeUp[i] = WakeUpPreamble
	}
	d.port.Write(wakeUp)
}

// buildReadFrame 构建读数据请求帧
// 帧格式: 68H + 地址域(6字节) + 68H + 控制码(1字节) + 数据长度(1字节) + 数据域 + 校验(1字节) + 16H
func (d *DLTDevice) buildReadFrame(dataId uint32) []byte {
	// 数据域: DI0-DI3 (4字节)
	dataField := make([]byte, 4)
	var mask byte = 0
	if d.option.EnableDataMask {
		mask = DataMask
	}
	dataField[0] = byte(dataId&0xFF) + mask
	dataField[1] = byte((dataId>>8)&0xFF) + mask
	dataField[2] = byte((dataId>>16)&0xFF) + mask
	dataField[3] = byte((dataId>>24)&0xFF) + mask

	// 构建帧
	frame := make([]byte, 0, 16)

	// 帧头
	frame = append(frame, FrameStart)

	// 地址域（低位在前）
	frame = append(frame, d.meterAddress...)

	// 第二个帧头
	frame = append(frame, FrameStart)

	// 控制码
	frame = append(frame, CtrlCodeRead)

	// 数据长度
	frame = append(frame, byte(len(dataField)))

	// 数据域
	frame = append(frame, dataField...)

	// 计算校验码
	cs := calculateChecksum(frame)
	frame = append(frame, cs)

	// 结束符
	frame = append(frame, FrameEnd)

	return frame
}

// readResponse 读取响应数据
func (d *DLTDevice) readResponse() ([]byte, error) {
	buf := make([]byte, 256)
	result := make([]byte, 0, 256)

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
			if len(result) > 0 && result[len(result)-1] == FrameEnd {
				// 找到帧起始位置
				startIdx := -1
				for i := 0; i < len(result); i++ {
					if result[i] == FrameStart {
						startIdx = i
						break
					}
				}
				if startIdx >= 0 {
					result = result[startIdx:]
					// 验证帧长度是否足够
					if len(result) >= 12 {
						// 获取数据长度
						if len(result) > 9 {
							dataLen := int(result[9])
							expectedLen := 12 + dataLen
							if len(result) >= expectedLen {
								return result[:expectedLen], nil
							}
						}
					}
				}
			}
		}
	}

	if len(result) == 0 {
		return nil, errors.New("read timeout, no data received")
	}

	return result, nil
}

// parseResponseFrame 解析响应帧
func (d *DLTDevice) parseResponseFrame(frame []byte) ([]byte, error) {
	// 最小帧长度: 68 + A0-A5(6) + 68 + C(1) + L(1) + CS(1) + 16 = 12字节
	if len(frame) < 12 {
		return nil, fmt.Errorf("frame too short: %d bytes", len(frame))
	}

	// 验证帧头
	if frame[0] != FrameStart || frame[7] != FrameStart {
		return nil, fmt.Errorf("invalid frame header: %02X %02X", frame[0], frame[7])
	}

	// 验证帧尾
	if frame[len(frame)-1] != FrameEnd {
		return nil, fmt.Errorf("invalid frame end: %02X", frame[len(frame)-1])
	}

	// 获取控制码
	ctrlCode := frame[8]

	// 检查是否为异常应答
	if ctrlCode&0xC0 == 0xC0 {
		// 异常应答，D7=1且D6=1
		errCode := byte(0)
		if len(frame) > 10 {
			if d.option.EnableDataMask {
				errCode = frame[10] - DataMask
			} else {
				errCode = frame[10]
			}
		}
		return nil, fmt.Errorf("meter error response, ctrl: %02X, err: %02X", ctrlCode, errCode)
	}

	// 正常应答检查
	if ctrlCode != CtrlCodeReadReply {
		return nil, fmt.Errorf("unexpected control code: %02X", ctrlCode)
	}

	// 获取数据长度
	dataLen := int(frame[9])
	if len(frame) < 12+dataLen {
		return nil, fmt.Errorf("frame data incomplete: expected %d, got %d", 12+dataLen, len(frame))
	}

	// 验证校验码
	expectedCS := calculateChecksum(frame[:10+dataLen])
	actualCS := frame[10+dataLen]
	if expectedCS != actualCS {
		return nil, fmt.Errorf("checksum mismatch: expected %02X, got %02X", expectedCS, actualCS)
	}

	// 提取数据域（根据配置决定是否减去0x33）
	data := make([]byte, dataLen)
	for i := 0; i < dataLen; i++ {
		if d.option.EnableDataMask {
			data[i] = frame[10+i] - DataMask
		} else {
			data[i] = frame[10+i]
		}
	}

	return data, nil
}

// RequestPing 发送采集指令，最小化指令发送包
func (d *DLTDevice) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	qua, _ := d.Request(ctx, &packet)
	return qua
}

// Control 发送控制指令
func (d *DLTDevice) Control(packet *model.ControlProtocolPacket, val string) consts.Quality {
	// 暂不支持写操作
	log.Warnf("DLT645 Control not implemented yet")
	return consts.QualityConfigError
}

// setPointValueErrorQua 设置测点错误质量
func setPointValueErrorQua(point *model.PointInfo, qua consts.Quality) {
	point.RtVal.Qua = qua
	point.RtVal.Tms = utils.GetNowUTCTimeStamp()
}

// calculateChecksum 计算校验码
// 从第一个帧起始符开始到校验码之前的所有字节的算术和，取模256
func calculateChecksum(data []byte) byte {
	var sum byte = 0
	for _, b := range data {
		sum += b
	}
	return sum
}

// parseDataIdentifier 解析数据标识
// 支持格式:
// - 8位十六进制字符串: "00010000" (正向有功总电能)
// - 带0x前缀: "0x00010000"
func parseDataIdentifier(cmd string) (uint32, error) {
	if cmd == "" {
		return 0, errors.New("empty command")
	}

	cmd = strings.TrimSpace(cmd)
	cmd = strings.TrimPrefix(strings.ToLower(cmd), "0x")

	if len(cmd) != 8 {
		return 0, fmt.Errorf("invalid data identifier length: %d, expected 8", len(cmd))
	}

	val, err := strconv.ParseUint(cmd, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid data identifier: %s", cmd)
	}

	return uint32(val), nil
}
