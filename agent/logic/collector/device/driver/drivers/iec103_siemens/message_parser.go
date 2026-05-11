package iec103_siemens

import (
	"encoding/binary"
	"fmt"
	model2 "agent/entity/consts"
	"math"

	"trpc.group/trpc-go/trpc-go/log"
)

// ValueParser 值解析器
type ValueParser struct {
	Addr     string
	Extend   string
	DataType string
}

type msgParser struct {
}

// ParseCommunication 解析通信数据
func (m *msgParser) ParseCommunication(data []byte) ([]*DataPoint, error) {
	var points = make([]*DataPoint, 0)

	if len(data) < IEC103_HeadLen {
		return nil, fmt.Errorf("data too short")
	}

	// 解析ASDU
	asduStart := IEC103_AsduOffset
	if len(data) < IEC103_CotOffset {
		return nil, fmt.Errorf("COT data too short")
	}
	cot := data[IEC103_CotOffset]
	if cot == COT_QueryEnd {
		return points, nil
	}
	if len(data) <= asduStart+IEC103_AsduFixLen {
		return nil, fmt.Errorf("ASDU data too short")
	}

	// 通用分类服务数据解析
	offset := asduStart + IEC103_AsduFixLen // 跳过ASDU固定部分

	for offset < len(data) {
		if offset+IEC103_MinItemLen > len(data) {
			break
		}

		groupNum := data[offset]
		entryNum := data[offset+1]
		//kod := data[offset+2]
		dataType := data[offset+3]
		dataWidth := int(data[offset+4])
		dataCount := int(data[offset+5])

		offset += IEC103_MinItemLen

		dataEnd := offset + dataWidth*dataCount
		if dataEnd > len(data) {
			break
		}
		point := &DataPoint{
			Addr: CombineAddr(groupNum, entryNum),
			Qua:  model2.QualityOk,
		}
		if dataType == 0x12 { // 带时标的报文,双点信息
			if !(dataWidth == 6 && dataCount == 1) {
				log.Warnf("width may err,data type: %02X, width: %d, count: %d",
					dataType, dataWidth, dataCount)
			}
			if offset < len(data) {
				dip := data[offset]
				offset += dataWidth * dataCount
				var value int64 = 0
				if dip == 0x02 {
					value = 1 // 合状态
				}
				point.Value = value
			}
		} else if dataType == 0x13 { // 带相对时间的时标,双点信息
			if !(dataWidth == 10 && dataCount == 1) {
				log.Warnf("width may err,data type: %02X, width: %d, count: %d",
					dataType, dataWidth, dataCount)
			}
			if offset < len(data) {
				dip := data[offset]
				offset += dataWidth * dataCount
				var value int64 = 0
				if dip == 0x02 {
					value = 1 // 合状态
				}
				point.Value = value
			}
		} else if dataType == 0x0C { // 带品质的测量值
			if !(dataWidth == 2 && dataCount == 1) {
				log.Warnf("width may err,data type: %02X, width: %d, count: %d",
					dataType, dataWidth, dataCount)
			}
			// 解析遥测量：GID占2个字节
			if offset+2 <= len(data) {
				rawValue := binary.LittleEndian.Uint16(data[offset : offset+2])

				// 解析品质位（低3位）OV=0表示无溢出；OV=1 表示溢出。第2位为ER，ER=0表示测量值有效；ER=1表示测量值无效
				ov := (rawValue & 0x0001) != 0 // 最低位：OV（溢出标志）
				er := (rawValue & 0x0002) != 0 // 第2位：ER（有效性标志）
				//res := (rawValue & 0x0004) != 0 // 第3位：RES（备用）

				// 提取13位测量值（第4位至第16位）
				measurementValue := (rawValue >> 3) & 0x0FFF

				// 处理符号位（最高位为符号位）
				var actualValue int32
				if (rawValue & 0x8000) != 0 {
					// 负数，补码转原码
					actualValue = int32(measurementValue) - 4096 - 1
				} else {
					actualValue = int32(measurementValue)
				}
				// 设置品质标志
				if er {
					point.Qua = model2.QualityValueInvalidError // 测量值无效
				} else if ov {
					point.Qua = model2.QualityValueOverflow // 溢出
				}
				point.Value = float64(actualValue)
				offset += dataWidth * dataCount
			}
		} else if dataType == 0x03 { // 无符号整数
			if offset+dataWidth*dataCount <= len(data) {
				if dataWidth == 1 {
					point.Value = float64(data[offset])
				} else if dataWidth == 2 {
					point.Value = float64(binary.LittleEndian.Uint16(data[offset : offset+2]))
				} else if dataWidth == 4 {
					point.Value = float64(binary.LittleEndian.Uint32(data[offset : offset+4]))
				}
			}
			offset += dataWidth * dataCount
		} else if dataType == 0x09 { // DIP双点信息
			if offset+dataWidth*dataCount <= len(data) {
				dip := data[offset]
				var value int64 = 0
				if dip == 0x02 {
					value = 1 // 合状态
				}
				point.Value = value
			}
			offset += dataWidth * dataCount
		} else if dataType == 0x01 { // ASCII 8 位码
			if offset+dataWidth*dataCount <= len(data) {
				point.Value = string(data[offset : offset+dataWidth*dataCount])
			}
			offset += dataWidth * dataCount
		} else if dataType == 0x07 { // 短实数
			if offset+dataWidth*dataCount <= len(data) {
				minVal, maxVal, stepVal, err := parseSettingRange(data[offset : offset+dataWidth*dataCount])
				if err != nil {
					log.Warnf("parse range error: %v, val: % X", err, data[offset:offset+dataWidth*dataCount])
					point.Qua = model2.QualityValueAbnormal
				} else {
					point.Value = []float32{minVal, maxVal, stepVal}
				}
			}
			offset += dataWidth * dataCount
		} else {
			point.Value = make([][]byte, dataCount)
			for i := 0; i < dataCount; i++ {
				point.Value.([][]byte)[i] = data[offset+i*dataWidth : offset+i*dataWidth+dataWidth]
			}
			offset += dataWidth * dataCount
		}
		points = append(points, point)
	}
	return points, nil
}

// parseSettingRange 定值量程解析
func parseSettingRange(data []byte) (min, max, step float32, err error) {
	if len(data) < 12 {
		err = fmt.Errorf("定值量程数据需要12字节，实际%d字节", len(data))
		return
	}

	// 解析最小值（4字节）
	min, err = parseIEEE754ShortReal(data[0:4])
	if err != nil {
		return
	}

	// 解析最大值（4字节）
	max, err = parseIEEE754ShortReal(data[4:8])
	if err != nil {
		return
	}

	// 解析步长（4字节）
	step, err = parseIEEE754ShortReal(data[8:12])
	if err != nil {
		return
	}

	return
}

// parseIEEE754ShortReal 简化的解析函数
func parseIEEE754ShortReal(data []byte) (float32, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("需要4字节，实际%d字节", len(data))
	}
	bits := binary.LittleEndian.Uint32(data)
	return math.Float32frombits(bits), nil
}

// ParseFaultWaveData 解析故障录波数据
func ParseFaultWaveData(data []byte) (*FaultWaveData, error) {
	if len(data) < 28 {
		return nil, fmt.Errorf("data too short")
	}

	result := &FaultWaveData{}

	// todo 解析故障录波数据的具体逻辑
	// 这里根据文档中的故障录波报文格式进行解析

	return result, nil
}

// FaultWaveData 故障录波数据
type FaultWaveData struct {
	FaultNumber uint16
	Channels    []FaultChannel
	Timestamps  []int64
}

// FaultChannel 故障通道
type FaultChannel struct {
	ChannelNum uint8
	Data       []int16
}

func isHearbeat(data []byte) bool {
	return len(data) == IEC103_HeadLen && data[0] == 0x90 && data[1] == 0xEB && data[2] == 0x14
}
