package xbm

import (
	"encoding/binary"
	"testing"
)

func TestCalculateCRC(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint16
	}{
		{"读取电压请求", []byte{0xAB, 0x01, 0x01, 0x00, 0x50}, 0xDDEA},
		{"清除报警", []byte{0xAB, 0x01, 0x00, 0x00, 0xB0}, 0x17F4},
		{"读取所有传感器电压", []byte{0xAB, 0x01, 0x00, 0x00, 0x50}, 0xEADA},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCRC(tt.data)
			if result != tt.expected {
				t.Errorf("CRC错误: 期望 0x%04X, 实际 0x%04X", tt.expected, result)
			}
		})
	}
}

func TestBuildRequestPacket(t *testing.T) {
	tests := []struct {
		name            string
		transformerAddr byte
		sensorAddr      uint16
		command         byte
		expected        []byte
	}{
		{"读取传感器1电压", 0x01, 0x0001, CmdAcquireVoltage, []byte{0xAB, 0x01, 0x01, 0x00, 0x50, 0xEA, 0xDD}},
		{"读取所有传感器电压", 0x01, 0x0000, CmdAcquireVoltage, []byte{0xAB, 0x01, 0x00, 0x00, 0x50, 0xDA, 0xEA}},
		{"清除传感器报警", 0x01, 0x0000, 0xB0, []byte{0xAB, 0x01, 0x00, 0x00, 0xB0, 0xF4, 0x17}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildRequestPacket(tt.transformerAddr, tt.sensorAddr, tt.command)
			for i := range tt.expected {
				if result[i] != tt.expected[i] {
					t.Errorf("字节 %d 不匹配: 期望 0x%02X, 实际 0x%02X", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name            string
		response        []byte
		transformerAddr byte
		sensorAddr      uint16
		command         byte
		expectedValue   float64
	}{
		{"电压12.35V", buildTestResponse(0x01, 0x0001, CmdAcquireVoltage, 1235), 0x01, 0x0001, CmdAcquireVoltage, 12.35},
		{"电阻5.03mOhm", buildTestResponse(0x01, 0x0001, CmdAcquireResistance, 503), 0x01, 0x0001, CmdAcquireResistance, 5.03},
		{"温度26.8C", buildTestResponse(0x01, 0x0001, CmdAcquireTemp1, 268), 0x01, 0x0001, CmdAcquireTemp1, 26.8},
		{"环流321.5A", buildTestResponse(0x01, 0x8001, CmdAcquireLoopCurrent, 3215), 0x01, 0x8001, CmdAcquireLoopCurrent, 321.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := ParseResponse(tt.response, tt.transformerAddr, tt.sensorAddr, tt.command)
			if err != nil {
				t.Errorf("解析错误: %v", err)
				return
			}
			if value < tt.expectedValue-0.01 || value > tt.expectedValue+0.01 {
				t.Errorf("值错误: 期望 %f, 实际 %f", tt.expectedValue, value)
			}
		})
	}
}

func buildTestResponse(transformerAddr byte, sensorAddr uint16, command byte, dataValue uint16) []byte {
	packet := []byte{FrameHeader, transformerAddr}
	sensorAddrBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(sensorAddrBytes, sensorAddr)
	packet = append(packet, sensorAddrBytes...)
	packet = append(packet, command)
	dataBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(dataBytes, dataValue)
	packet = append(packet, dataBytes...)
	crc := CalculateCRC(packet)
	crcBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(crcBytes, crc)
	packet = append(packet, crcBytes...)
	return packet
}
