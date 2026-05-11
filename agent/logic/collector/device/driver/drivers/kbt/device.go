package kbt

import (
	"bytes"
	"context"
	"fmt"
	"agent/entity/consts"
	"agent/entity/definition"
	model3 "agent/entity/model"
	"agent/logic/collector/device/model"
	"agent/utils"
	"agent/utils/encoding"
	"agent/utils/osal"
	"strconv"
	"strings"
	"time"

	"trpc.group/trpc-go/trpc-go/log"
)

// Device 实现 IDevice
type Device struct {
	gid  definition.DeviceGidType
	name string

	deviceAddr byte
	port       *SerialPort
}

// Open 打开设备通道
func (d *Device) Open(chanInfo model.ChannelInfo, _ model.ListCollectPackets) consts.Quality {
	baudRate, parity, dataBits, stopBits, err := model.ParseRTUParam(chanInfo.Params)
	if err != nil {
		log.Errorf("KBT RTU params parse err=%v, fallback 9600:N:8:1", err)
		baudRate, parity, dataBits, stopBits = 9600, "N", 8, 1
	}

	// 解析设备地址
	addr, err := strconv.Atoi(strings.TrimSpace(chanInfo.Address))
	if err != nil || addr < 1 || addr > 255 {
		log.Errorf("KBT device address must be 1-255, got %q", chanInfo.Address)
		return consts.QualityConfigError
	}
	d.deviceAddr = byte(addr)

	sp, err := OpenSerial(SerialOptions{
		Port:       chanInfo.Name,
		Baud:       baudRate,
		DataBits:   dataBits,
		StopBits:   stopBits,
		Parity:     model.NormalizeParity(parity),
		ReadTO:     time.Duration(utils.FirstNonZero(chanInfo.TimeoutMs, 3000)) * time.Millisecond,
		WriteTO:    2 * time.Second,
		HardwareFC: false,
		SoftwareFC: false,
	})
	if err != nil {
		log.Errorf("KBT open serial err=%v", err)
		return consts.QualityCannotOpen
	}
	d.port = sp
	return consts.QualityOk
}

func (d *Device) Close() consts.Quality {
	if d.port != nil {
		_ = d.port.Close()
		d.port = nil
	}
	return consts.QualityOk
}

// Request 发送采集指令并解析响应
func (d *Device) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model3.MessageStatistics) {
	var stat model3.MessageStatistics
	if d.port == nil {
		return consts.QualityCommDisconnected, stat
	}

	// KBT1000装置主动上传数据，无需发送请求指令
	// 等待接收装置上传的数据帧
	rx, err := d.port.Read(ctx, 256)
	if err != nil {
		log.Errorf("KBT read err=%v", err)
		stat.ErrLog = err
		return consts.QualityCmdRespError, stat
	}
	stat.RecvPackets = encoding.HexSp(rx)

	// 解析KBT1000协议帧
	frame, _, ok := parseKBTFrame(rx)
	if !ok {
		log.Errorf("KBT frame parse failed")
		return consts.QualityCmdRespError, stat
	}

	// 检查设备地址是否匹配
	if frame.DeviceAddr != d.deviceAddr {
		log.Warnf("KBT device address mismatch, expected %d, got %d", d.deviceAddr, frame.DeviceAddr)
	}

	// 映射到测点
	if err := fillPointsFromFrame(frame, packet); err != nil {
		log.Warnf("KBT map points warn: %v", err)
	}

	return consts.QualityOk, stat
}

func (d *Device) RequestPing(ctx context.Context, pkt model.CollectProtocolPacket) consts.Quality {
	q, _ := d.Request(ctx, &pkt)
	return q
}

func (d *Device) Control(_ *model.ControlProtocolPacket, _ string) consts.Quality {
	return consts.QualityOk
}

// parseKBTFrame 解析KBT1000协议帧
func parseKBTFrame(data []byte) (*KBTFrame, []byte, bool) {
	// 查找同步字
	syncIndex := bytes.Index(data, SyncWord)
	if syncIndex == -1 {
		// 未找到同步字，保留最后5个字节（可能是不完整的同步字）
		log.Debugf("未找到同步字，数据长度：%d，数据：%X", len(data), data)
		if len(data) > 5 {
			return nil, data[len(data)-5:], false
		}
		return nil, data, false
	}

	// 丢弃同步字之前的数据
	if syncIndex > 0 {
		log.Debugf("丢弃 %d 字节无效数据，剩余数据：%X", syncIndex, data[syncIndex:])
		data = data[syncIndex:]
	}

	// 检查是否有完整帧
	if len(data) < FrameLength {
		log.Debugf("数据长度不足，需要%d字节，实际%d字节，数据：%X", FrameLength, len(data), data)
		return nil, data, false
	}

	// 如果数据长度超过帧长度，只取前24字节作为帧数据
	if len(data) > FrameLength {
		log.Debugf("数据长度超过帧长度，实际%d字节，只取前%d字节", len(data), FrameLength)
	}

	// 提取帧数据
	frameData := data[:FrameLength]
	remaining := data[FrameLength:]

	log.Debugf("解析KBT帧，帧数据：%X，剩余数据：%X", frameData, remaining)

	frame := &KBTFrame{
		SyncWord:  frameData[0:6],
		CtrlWord:  frameData[6:12],
		InfoWord1: frameData[12:18],
		InfoWord2: frameData[18:24],
	}

	// 提取设备地址 (B10)
	frame.DeviceAddr = frameData[9]
	log.Debugf("提取设备地址：0x%02X", frame.DeviceAddr)

	// 验证控制字
	if !validateCtrlWord(frame.CtrlWord) {
		log.Errorf("KBT control word validation failed, control word: %X", frame.CtrlWord)
		return nil, remaining, false
	}

	// 验证信息字
	if !validateInfoWord(frame.InfoWord1) {
		log.Errorf("KBT info word1 validation failed")
		return nil, remaining, false
	}

	if !validateInfoWord(frame.InfoWord2) {
		log.Errorf("KBT info word2 validation failed")
		return nil, remaining, false
	}

	// 解析线路状态
	parseLineStatus(frame)

	return frame, remaining, true
}

// fillPointsFromFrame 将帧数据映射到测点
func fillPointsFromFrame(frame *KBTFrame, packet *model.CollectProtocolPacket) error {
	if packet == nil || len(packet.Points) == 0 {
		return nil
	}
	now := utils.GetNowUTCTimeStamp()

	for _, pt := range packet.Points {
		if pt == nil || pt.Attr.ValParser == nil {
			continue
		}
		vp, ok := pt.Attr.ValParser.(*ValueParser)
		if !ok {
			continue
		}

		// 根据线路编号获取状态值
		if vp.LineNum < 1 || vp.LineNum > 64 {
			continue
		}

		status := frame.LineStatus[vp.LineNum-1]
		var val int
		if status {
			val = 1 // 接地故障
		} else {
			val = 0 // 正常
		}

		pt.RtVal.Pv = osal.NewVariantWithValue(val)
		pt.RtVal.Tms = now
		pt.RtVal.Qua = consts.QualityOk
	}
	return nil
}

// SimulateFrame 生成模拟测试帧
func SimulateFrame(deviceAddr byte, faultLines []int) []byte {
	frame := make([]byte, FrameLength)

	// 同步字 (B1-B6)
	copy(frame[0:6], SyncWord)

	// 控制字 (B7-B12)
	frame[6] = 0x71                       // B7
	frame[7] = 0xF4                       // B8
	frame[8] = 0x01                       // B9
	frame[9] = deviceAddr                 // B10 设备地址
	frame[10] = 0x01                      // B11
	frame[11] = calculateCRC(frame[6:11]) // B12 CRC

	// 信息字1 (B13-B18)
	frame[12] = 0xF0 // B13 功能码
	// B14-B17 线路状态 (线路1-32)
	for _, lineNum := range faultLines {
		if lineNum >= 1 && lineNum <= 32 {
			byteIdx := (lineNum - 1) / 8
			bitIdx := (lineNum - 1) % 8
			frame[13+byteIdx] |= (1 << bitIdx)
		}
	}
	frame[17] = calculateCRC(frame[12:17]) // B18 CRC

	// 信息字2 (B19-B24)
	frame[18] = 0xF1 // B19 功能码
	// B20-B23 线路状态 (线路33-64)
	for _, lineNum := range faultLines {
		if lineNum >= 33 && lineNum <= 64 {
			byteIdx := (lineNum - 33) / 8
			bitIdx := (lineNum - 33) % 8
			frame[19+byteIdx] |= (1 << bitIdx)
		}
	}
	frame[23] = calculateCRC(frame[18:23]) // B24 CRC

	return frame
}

// ParseResponse 解析KBT1000协议响应
func ParseResponse(data []byte) (*ParsedResponse, error) {
	frame, _, ok := parseKBTFrame(data)
	if !ok {
		return nil, fmt.Errorf("failed to parse KBT frame")
	}

	parsed := &ParsedResponse{
		Frame:  frame,
		Parsed: make(map[string]interface{}),
	}

	// 将线路状态映射到解析结果中
	for i := 0; i < 64; i++ {
		lineNum := i + 1
		key := fmt.Sprintf("line_%d", lineNum)
		parsed.Parsed[key] = frame.LineStatus[i]
	}

	return parsed, nil
}

// ProcessFrame 处理解析后的帧
func ProcessFrame(frame *KBTFrame) {
	fmt.Println("\\n========== 帧解析结果 ==========")
	fmt.Printf("时间: %s\\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("设备地址: %d\\n", frame.DeviceAddr)

	// 显示线路状态
	fmt.Println("\\n---------- 线路状态 ----------")
	faultCount := 0

	for i := 0; i < 64; i++ {
		lineNum := i + 1
		status := frame.LineStatus[i]

		if status {
			faultCount++
			fmt.Printf("  ⚠ 线路%d: 接地故障\\n", lineNum)
		}
	}

	if faultCount == 0 {
		fmt.Println("所有线路运行正常 ✓")
	} else {
		fmt.Printf("!!! 发现 %d 路接地故障 !!!\\n", faultCount)
	}

	fmt.Println("================================\\n")
}

// ParsedResponse 解析后的响应结构
type ParsedResponse struct {
	Frame  *KBTFrame
	Parsed map[string]interface{}
}

// parseRegSpec 解析寄存器规格
func parseRegSpec(input string) (int, int, error) {
	if input == "" {
		return 0, -1, fmt.Errorf("empty input")
	}

	// 这里实现寄存器规格解析逻辑
	// 简化实现，返回默认值
	return 0, -1, nil
}
