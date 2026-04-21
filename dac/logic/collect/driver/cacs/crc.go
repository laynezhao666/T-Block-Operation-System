package cacs

// tableCRC CRC查找表类型定义
type tableCRC []uint8

// tableCRCHi CRC高字节查找表
// tableCRCLo CRC低字节查找表
var (
	tableCRCHi tableCRC
	tableCRCLo tableCRC
)

// CRC16 计算给定字节切片的 Modbus CRC16 校验值。
func CRC16(buff []byte) uint16 {
	return CRC16Update(0xffff, buff, len(buff))
}

// CRC16Update 基于初始CRC值增量计算 Modbus CRC16 校验。
// 使用预计算的查找表加速计算过程。
func CRC16Update(crc uint16, buff []byte, length int) uint16 {
	var crcHi uint8 = uint8(crc & 0xff)
	var crcLo uint8 = uint8(crc >> 8)
	var i uint
	index := length
	for index > 0 {
		i = uint(crcHi ^ buff[length-index])
		index--
		crcHi = crcLo ^ tableCRCHi[i]
		crcLo = tableCRCLo[i]
	}
	var result uint16 = uint16(crcLo)
	result <<= 8
	result |= uint16(crcHi)
	return result

	//return 0
}

// init 通过 Modbus CRC16 多项式 0xA001 动态生成查找表。
func init() {
	// 通过 Modbus CRC16 多项式 0xA001 动态生成查找表，替代硬编码数据
	const polynomial uint16 = 0xA001
	tableCRCHi = make([]uint8, 256)
	tableCRCLo = make([]uint8, 256)
	for i := 0; i < 256; i++ {
		crc := uint16(i)
		for j := 0; j < 8; j++ {
			if crc&1 == 1 {
				crc = (crc >> 1) ^ polynomial
			} else {
				crc >>= 1
			}
		}
		tableCRCHi[i] = uint8(crc & 0xFF)
		tableCRCLo[i] = uint8(crc >> 8)
	}
}
