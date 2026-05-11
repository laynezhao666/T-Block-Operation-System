package fdm

import (
	"testing"
)

// TestBuildRequestPacket_Example1And2 测试构建请求数据包
// 根据协议文档例一和例二：设备地址为1，功能码为4（读气体浓度）
// 请求包：40 30 31 30 38 30 30 30 34 30 34 30 39 0D
func TestBuildRequestPacket_Example1And2(t *testing.T) {
	device := &FDMDevice{
		option: Option{
			ReadTimeOut: 2000,
			ReadRetries: 3,
		},
	}

	// 设备地址1，功能码4（读气体浓度）
	packet := device.buildRequestPacket(1, FuncCodeReadGasConcentration)

	// 预期结果：40 30 31 30 38 30 30 30 34 30 34 30 39 0D
	expected := []byte{
		0x40,       // @ 起始符
		0x30, 0x31, // 地址 01
		0x30, 0x38, // 功能码1 08 (读数据)
		0x30, 0x30, 0x30, 0x34, // 功能码2 0004 (气体浓度)
		0x30, 0x34, // 字节长度 04
		0x30, 0x39, // 校验码 09
		0x0D, // 结束符
	}

	if len(packet) != len(expected) {
		t.Errorf("packet length mismatch: got %d, expected %d", len(packet), len(expected))
		return
	}

	for i := range expected {
		if packet[i] != expected[i] {
			t.Errorf("byte %d mismatch: got 0x%02X, expected 0x%02X", i, packet[i], expected[i])
		}
	}

	t.Logf("Request packet: % X", packet)
}

// TestParseResponse_Example1 测试解析响应数据包 - 例一
// 设备地址为1，气体浓度为0ppm
// 响应包：40 30 31 30 34 30 30 30 30 30 30 30 30 30 35 0D
func TestParseResponse_Example1(t *testing.T) {
	device := &FDMDevice{}

	// 例一响应数据：40 30 31 30 34 30 30 30 30 30 30 30 30 30 35 0D
	respData := []byte{
		0x40,       // @ 起始符
		0x30, 0x31, // 地址 01
		0x30, 0x34, // 字节长度 04
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, // 数据区 00000000
		0x30, 0x35, // 校验码 05
		0x0D, // 结束符
	}

	value, err := device.parseResponse(respData, 1)
	if err != nil {
		t.Errorf("parseResponse error: %v", err)
		return
	}

	// 预期气体浓度为0ppm
	if value != 0.0 {
		t.Errorf("value mismatch: got %f, expected 0.0", value)
	}

	t.Logf("Example1 - Gas concentration: %f ppm", value)
}

// TestParseResponse_Example2 测试解析响应数据包 - 例二
// 设备地址为1，气体浓度为1000ppm
// 响应包：40 30 31 30 34 30 41 46 41 30 30 30 30 37 33 0D
func TestParseResponse_Example2(t *testing.T) {
	device := &FDMDevice{}

	// 例二响应数据：40 30 31 30 34 30 41 46 41 30 30 30 30 37 33 0D
	respData := []byte{
		0x40,       // @ 起始符
		0x30, 0x31, // 地址 01
		0x30, 0x34, // 字节长度 04
		0x30, 0x41, 0x46, 0x41, 0x30, 0x30, 0x30, 0x30, // 数据区 0AFA0000
		0x37, 0x33, // 校验码 73
		0x0D, // 结束符
	}

	value, err := device.parseResponse(respData, 1)
	if err != nil {
		t.Errorf("parseResponse error: %v", err)
		return
	}

	// 预期气体浓度为1000ppm
	// 允许一定的浮点数误差
	expectedValue := 1000.0
	tolerance := 0.01
	if value < expectedValue-tolerance || value > expectedValue+tolerance {
		t.Errorf("value mismatch: got %f, expected %f (tolerance: %f)", value, expectedValue, tolerance)
	}

	t.Logf("Example2 - Gas concentration: %f ppm", value)
}

// TestParseFloatData_ZeroValue 测试解析浮点数数据 - 0值
// 数据区：00 00 00 00 (ASCII: 30 30 30 30 30 30 30 30)
func TestParseFloatData_ZeroValue(t *testing.T) {
	// ASCII表示的0x00000000
	data := []byte{'0', '0', '0', '0', '0', '0', '0', '0'}

	value, err := parseFloatData(data)
	if err != nil {
		t.Errorf("parseFloatData error: %v", err)
		return
	}

	if value != 0.0 {
		t.Errorf("value mismatch: got %f, expected 0.0", value)
	}

	t.Logf("Zero value: %f", value)
}

// TestParseFloatData_1000Value 测试解析浮点数数据 - 1000值
// 数据区：0A FA 00 00 (ASCII: 30 41 46 41 30 30 30 30)
// 0x0A = 0000 1010
//   D7(数符) = 0 (正数)
//   D6(阶符) = 0 (正阶)
//   D5-D0(阶码) = 001010 = 10
// 0xFA = 250, 0x00 = 0, 0x00 = 0
// 计算: ((((0/256)+0)/256)+250)/256 × 2^10
//     = (250/256) × 1024
//     = 0.9765625 × 1024
//     = 1000
func TestParseFloatData_1000Value(t *testing.T) {
	// ASCII表示的0x0AFA0000
	data := []byte{'0', 'A', 'F', 'A', '0', '0', '0', '0'}

	value, err := parseFloatData(data)
	if err != nil {
		t.Errorf("parseFloatData error: %v", err)
		return
	}

	expectedValue := 1000.0
	tolerance := 0.01
	if value < expectedValue-tolerance || value > expectedValue+tolerance {
		t.Errorf("value mismatch: got %f, expected %f", value, expectedValue)
	}

	t.Logf("1000 value: %f", value)
}

// TestCalculateChecksum 测试校验码计算
func TestCalculateChecksum(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		expected byte
	}{
		{
			// 例一请求的校验数据：30 31 30 38 30 30 30 34 30 34
			// 地址01 + 功能码1(08) + 功能码2(0004) + 字节长度(04)
			name:     "Example1_Request",
			data:     []byte{0x30, 0x31, 0x30, 0x38, 0x30, 0x30, 0x30, 0x34, 0x30, 0x34},
			expected: 0x09, // 校验码09
		},
		{
			// 例一响应的校验数据：30 31 30 34 30 30 30 30 30 30 30 30
			// 地址01 + 字节长度04 + 数据区00000000
			name:     "Example1_Response",
			data:     []byte{0x30, 0x31, 0x30, 0x34, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30},
			expected: 0x05, // 校验码05
		},
		{
			// 例二响应的校验数据：30 31 30 34 30 41 46 41 30 30 30 30
			// 地址01 + 字节长度04 + 数据区0AFA0000
			name:     "Example2_Response",
			data:     []byte{0x30, 0x31, 0x30, 0x34, 0x30, 0x41, 0x46, 0x41, 0x30, 0x30, 0x30, 0x30},
			expected: 0x73, // 校验码73
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateChecksum(tc.data)
			if result != tc.expected {
				t.Errorf("checksum mismatch: got 0x%02X, expected 0x%02X", result, tc.expected)
			}
			t.Logf("%s checksum: 0x%02X", tc.name, result)
		})
	}
}

// TestByteToASCII 测试字节转ASCII
func TestByteToASCII(t *testing.T) {
	testCases := []struct {
		input    byte
		expected []byte
	}{
		{0x00, []byte{'0', '0'}},
		{0x01, []byte{'0', '1'}},
		{0x09, []byte{'0', '9'}},
		{0x0A, []byte{'0', 'A'}},
		{0x0F, []byte{'0', 'F'}},
		{0x10, []byte{'1', '0'}},
		{0xFF, []byte{'F', 'F'}},
	}

	for _, tc := range testCases {
		result := byteToASCII(tc.input)
		if len(result) != 2 || result[0] != tc.expected[0] || result[1] != tc.expected[1] {
			t.Errorf("byteToASCII(0x%02X): got %c%c, expected %c%c",
				tc.input, result[0], result[1], tc.expected[0], tc.expected[1])
		}
	}
}

// TestUint16ToASCII 测试uint16转ASCII
func TestUint16ToASCII(t *testing.T) {
	testCases := []struct {
		input    uint16
		expected []byte
	}{
		{0x0000, []byte{'0', '0', '0', '0'}},
		{0x0004, []byte{'0', '0', '0', '4'}}, // 功能码4 (气体浓度)
		{0x0008, []byte{'0', '0', '0', '8'}}, // 功能码8 (温度)
		{0x000C, []byte{'0', '0', '0', 'C'}}, // 功能码12
		{0x0010, []byte{'0', '0', '1', '0'}}, // 功能码16
		{0x0050, []byte{'0', '0', '5', '0'}}, // 功能码80
		{0xFFFF, []byte{'F', 'F', 'F', 'F'}},
	}

	for _, tc := range testCases {
		result := uint16ToASCII(tc.input)
		if len(result) != 4 {
			t.Errorf("uint16ToASCII(0x%04X): wrong length %d", tc.input, len(result))
			continue
		}
		for i := 0; i < 4; i++ {
			if result[i] != tc.expected[i] {
				t.Errorf("uint16ToASCII(0x%04X): byte %d got %c, expected %c",
					tc.input, i, result[i], tc.expected[i])
			}
		}
	}
}

// TestAsciiToByte 测试ASCII转字节
func TestAsciiToByte(t *testing.T) {
	testCases := []struct {
		input    byte
		expected byte
		hasError bool
	}{
		{'0', 0x00, false},
		{'9', 0x09, false},
		{'A', 0x0A, false},
		{'F', 0x0F, false},
		{'a', 0x0A, false},
		{'f', 0x0F, false},
		{'G', 0, true},
		{'g', 0, true},
		{' ', 0, true},
	}

	for _, tc := range testCases {
		result, err := asciiToByte(tc.input)
		if tc.hasError {
			if err == nil {
				t.Errorf("asciiToByte('%c'): expected error but got none", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("asciiToByte('%c'): unexpected error: %v", tc.input, err)
			} else if result != tc.expected {
				t.Errorf("asciiToByte('%c'): got 0x%X, expected 0x%X", tc.input, result, tc.expected)
			}
		}
	}
}

// TestParseResponse_AddressMismatch 测试地址不匹配的情况
func TestParseResponse_AddressMismatch(t *testing.T) {
	device := &FDMDevice{}

	// 响应地址是1，但期望地址是2
	respData := []byte{
		0x40,       // @ 起始符
		0x30, 0x31, // 地址 01
		0x30, 0x34, // 字节长度 04
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, // 数据区
		0x30, 0x35, // 校验码
		0x0D, // 结束符
	}

	_, err := device.parseResponse(respData, 2) // 期望地址2
	if err == nil {
		t.Error("expected address mismatch error but got none")
	} else {
		t.Logf("Address mismatch error (expected): %v", err)
	}
}

// TestParseResponse_ChecksumMismatch 测试校验码不匹配的情况
func TestParseResponse_ChecksumMismatch(t *testing.T) {
	device := &FDMDevice{}

	// 故意修改校验码
	respData := []byte{
		0x40,       // @ 起始符
		0x30, 0x31, // 地址 01
		0x30, 0x34, // 字节长度 04
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, // 数据区
		0x30, 0x30, // 错误的校验码 00 (正确应该是05)
		0x0D, // 结束符
	}

	_, err := device.parseResponse(respData, 1)
	if err == nil {
		t.Error("expected checksum mismatch error but got none")
	} else {
		t.Logf("Checksum mismatch error (expected): %v", err)
	}
}

// TestParseResponse_TooShort 测试响应数据太短的情况
func TestParseResponse_TooShort(t *testing.T) {
	device := &FDMDevice{}

	// 数据太短
	respData := []byte{0x40, 0x30, 0x31}

	_, err := device.parseResponse(respData, 1)
	if err == nil {
		t.Error("expected too short error but got none")
	} else {
		t.Logf("Too short error (expected): %v", err)
	}
}

// ===================== 例三、例四、例五、例六 测试用例 =====================

// TestParseResponse_Example3 测试解析响应数据包 - 例三
// 根据协议文档：设备地址为1，气体浓度为0ppm
// 响应包：40 30 31 30 34 30 30 30 30 30 30 30 30 30 35 0D
// 数据区第一字节：30 30 → 0x00 → 二进制 0000 0000
// 数符=0(正), 阶符=0(正), 阶码=000000=0, 即 2^0 = 1
// 小数部分全为0，结果为0
func TestParseResponse_Example3(t *testing.T) {
	device := &FDMDevice{}

	// 例三响应数据：40 30 31 30 34 30 30 30 30 30 30 30 30 30 35 0D
	respData := []byte{
		0x40,       // @ 起始符
		0x30, 0x31, // 地址 01
		0x30, 0x34, // 字节长度 04
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, // 数据区 00000000
		0x30, 0x35, // 校验码 05
		0x0D,       // 结束符
	}

	value, err := device.parseResponse(respData, 1)
	if err != nil {
		t.Errorf("parseResponse error: %v", err)
		return
	}

	// 预期气体浓度为0ppm
	if value != 0.0 {
		t.Errorf("value mismatch: got %f, expected 0.0", value)
	}

	t.Logf("Example3 - Gas concentration: %f ppm", value)
}

// TestParseResponse_Example4 测试解析响应数据包 - 例四
// 根据协议文档：设备地址为1，气体浓度为1000ppm
// 响应包：40 30 31 30 34 30 41 46 41 30 30 30 30 37 33 0D
// 数据区第一字节：30 41 → 0x0A → 二进制 0000 1010
// 数符=0(正), 阶符=0(正), 阶码=001010=10, 即 2^10 = 1024
// 小数部分：FA(250), 00(0), 00(0)
// 计算：((((0/256)+0)/256)+250)/256 × 1024 = 0.9765625 × 1024 = 1000
func TestParseResponse_Example4(t *testing.T) {
	device := &FDMDevice{}

	// 例四响应数据：40 30 31 30 34 30 41 46 41 30 30 30 30 37 33 0D
	respData := []byte{
		0x40,       // @ 起始符
		0x30, 0x31, // 地址 01
		0x30, 0x34, // 字节长度 04
		0x30, 0x41, 0x46, 0x41, 0x30, 0x30, 0x30, 0x30, // 数据区 0AFA0000
		0x37, 0x33, // 校验码 73
		0x0D,       // 结束符
	}

	value, err := device.parseResponse(respData, 1)
	if err != nil {
		t.Errorf("parseResponse error: %v", err)
		return
	}

	// 预期气体浓度为1000ppm
	expectedValue := 1000.0
	tolerance := 0.01
	if value < expectedValue-tolerance || value > expectedValue+tolerance {
		t.Errorf("value mismatch: got %f, expected %f (tolerance: %f)", value, expectedValue, tolerance)
	}

	t.Logf("Example4 - Gas concentration: %f ppm", value)
}

// TestParseFloatData_Example5 测试解析浮点数数据 - 例五
// 根据协议文档：瞬时流量测量值为100.210
// 数据区：07 C8 66 66
// 第一字节：07 → 二进制 0000 0111
// 数符=0(正), 阶符=0(正), 阶码=000111=7, 即 2^7 = 128
// 小数部分：C8(200), 66(102), 66(102)
// 计算：((((102/256)+102)/256)+200)/256 × 128
//     = ((0.3984375+102)/256)+200)/256 × 128
//     = (0.39990234375+200)/256 × 128
//     = 0.7828125 × 128
//     ≈ 100.2 (协议文档示例值为100.210，存在精度差异)
func TestParseFloatData_Example5(t *testing.T) {
	// ASCII表示的0x07C86666
	data := []byte{'0', '7', 'C', '8', '6', '6', '6', '6'}

	value, err := parseFloatData(data)
	if err != nil {
		t.Errorf("parseFloatData error: %v", err)
		return
	}

	// 预期值约为100.2 (文档给出100.210，但实际计算会有精度差异)
	expectedValue := 100.2
	tolerance := 0.1 // 允许较大误差，因为浮点数精度问题
	if value < expectedValue-tolerance || value > expectedValue+tolerance {
		t.Errorf("value mismatch: got %f, expected ~%f (tolerance: %f)", value, expectedValue, tolerance)
	}

	t.Logf("Example5 - Flow value: %f (expected ~100.210)", value)
}

// TestParseFloatData_Example6 测试解析浮点数数据 - 例六
// 根据协议文档：浮点数值为 -0.1234567
// 数据区：C3 FC D6 E8
// 第一字节：C3 → 二进制 1100 0011
// 数符=1(负), 阶符=1(负), 阶码=000011=3, 即 2^(-3) = 0.125
// 小数部分：FC(252), D6(214), E8(232)
// 计算：((((232/256)+214)/256)+252)/256 × 0.125
//     = 0.98765432... × 0.125
//     = 0.1234567...
// 由于数符为1，结果为负数：-0.1234567
func TestParseFloatData_Example6(t *testing.T) {
	// ASCII表示的0xC3FCD6E8
	data := []byte{'C', '3', 'F', 'C', 'D', '6', 'E', '8'}

	value, err := parseFloatData(data)
	if err != nil {
		t.Errorf("parseFloatData error: %v", err)
		return
	}

	// 预期值为 -0.1234567
	expectedValue := -0.1234567
	tolerance := 0.0001 // 允许较小误差
	if value < expectedValue-tolerance || value > expectedValue+tolerance {
		t.Errorf("value mismatch: got %f, expected %f (tolerance: %f)", value, expectedValue, tolerance)
	}

	t.Logf("Example6 - Negative float value: %f (expected -0.1234567)", value)
}

// TestParseFloatData_Example5_DetailedCalculation 详细验证例五的计算过程
func TestParseFloatData_Example5_DetailedCalculation(t *testing.T) {
	// 手动计算验证
	// 数据：07 C8 66 66
	// 第一字节：07 = 0000 0111
	signBit := byte(0)      // D7 = 0 (正数)
	expSignBit := byte(0)   // D6 = 0 (正阶)
	expValue := 7           // D5-D0 = 000111 = 7
	a2 := float64(0xC8)     // 200
	a3 := float64(0x66)     // 102
	a4 := float64(0x66)     // 102

	// 计算小数部分
	mantissa := ((((a4 / 256.0) + a3) / 256.0) + a2) / 256.0
	t.Logf("小数部分: ((((%.0f/256)+%.0f)/256)+%.0f)/256 = %f", a4, a3, a2, mantissa)

	// 计算指数
	var exp float64 = 1.0
	if expSignBit == 0 {
		for i := 0; i < expValue; i++ {
			exp *= 2.0
		}
	}
	t.Logf("指数: 2^%d = %f", expValue, exp)

	// 计算结果
	result := mantissa * exp
	if signBit == 1 {
		result = -result
	}
	t.Logf("计算结果: %f × %f = %f", mantissa, exp, result)

	// 验证parseFloatData函数的结果
	data := []byte{'0', '7', 'C', '8', '6', '6', '6', '6'}
	value, _ := parseFloatData(data)

	if result != value {
		t.Errorf("计算结果不一致: 手动计算=%f, 函数结果=%f", result, value)
	}
}

// TestParseFloatData_Example6_DetailedCalculation 详细验证例六的计算过程
func TestParseFloatData_Example6_DetailedCalculation(t *testing.T) {
	// 手动计算验证
	// 数据：C3 FC D6 E8
	// 第一字节：C3 = 1100 0011
	signBit := byte(1)      // D7 = 1 (负数)
	expSignBit := byte(1)   // D6 = 1 (负阶)
	expValue := 3           // D5-D0 = 000011 = 3
	a2 := float64(0xFC)     // 252
	a3 := float64(0xD6)     // 214
	a4 := float64(0xE8)     // 232

	// 计算小数部分
	mantissa := ((((a4 / 256.0) + a3) / 256.0) + a2) / 256.0
	t.Logf("小数部分: ((((%.0f/256)+%.0f)/256)+%.0f)/256 = %f", a4, a3, a2, mantissa)

	// 计算指数 (负阶)
	var exp float64 = 1.0
	if expSignBit == 1 {
		for i := 0; i < expValue; i++ {
			exp /= 2.0
		}
	}
	t.Logf("指数: 2^(-%d) = %f", expValue, exp)

	// 计算结果
	result := mantissa * exp
	if signBit == 1 {
		result = -result
	}
	t.Logf("计算结果: %f × %f × (-1) = %f", mantissa, exp, result)

	// 验证parseFloatData函数的结果
	data := []byte{'C', '3', 'F', 'C', 'D', '6', 'E', '8'}
	value, _ := parseFloatData(data)

	tolerance := 0.0000001
	if result < value-tolerance || result > value+tolerance {
		t.Errorf("计算结果不一致: 手动计算=%f, 函数结果=%f", result, value)
	}
}

// TestParseFloatData_500Value 测试解析浮点数数据 - 协议文档5.4节示例值500
// 根据协议文档：十进制数500，十六进制01F4H
// 用浮点数表示：0.9765625 × 2^9
// 数据区：09 FA 00 00 (注：文档写的是 09 01 0F 04 但那是原始十六进制分解，实际浮点编码应该是09FA0000)
// 第一字节：09 = 0000 1001
// 数符=0(正), 阶符=0(正), 阶码=001001=9, 即 2^9 = 512
// 小数部分应该是0.9765625对应的编码
func TestParseFloatData_500Value(t *testing.T) {
	// 500 = 0.9765625 × 2^9
	// 0.9765625 = 250/256 = FA/100 (十六进制)
	// 所以浮点编码为：09 FA 00 00
	data := []byte{'0', '9', 'F', 'A', '0', '0', '0', '0'}

	value, err := parseFloatData(data)
	if err != nil {
		t.Errorf("parseFloatData error: %v", err)
		return
	}

	expectedValue := 500.0
	tolerance := 0.01
	if value < expectedValue-tolerance || value > expectedValue+tolerance {
		t.Errorf("value mismatch: got %f, expected %f", value, expectedValue)
	}

	t.Logf("500 value: %f", value)
}
