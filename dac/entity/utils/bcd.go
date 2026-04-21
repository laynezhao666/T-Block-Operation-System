// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"fmt"
	"strconv"
)

// FourDigitIntToBCDUint16 将4位十进制整数转换为BCD编码的uint16值。
// 例如：2024 -> 0x2024。每个十进制位占4个二进制位。
func FourDigitIntToBCDUint16(n int) uint16 {
	decimalStr := strconv.Itoa(n)
	bcdStr := ""

	// 逐位将十进制数字转换为4位BCD编码
	for _, digit := range decimalStr {
		bcd := fmt.Sprintf("%04b", digit-'0')
		bcdStr += bcd
	}
	bcdInt, err := strconv.ParseUint(bcdStr, 2, 16)
	if err != nil {
		panic(err)
	}

	return uint16(bcdInt)
}
