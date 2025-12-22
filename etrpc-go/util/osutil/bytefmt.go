// Package osutil provides various os tools
package osutil

import (
	"fmt"
	"strings"
)

const (
	Byte     = 1.0
	KiloByte = 1024 * Byte
	MegaByte = 1024 * KiloByte
	GigaByte = 1024 * MegaByte
	TeraByte = 1024 * GigaByte
)

// Byte2Str 将字节大小转成可读的大小，比如：12.5k、10M、6.35G、13.1T等
func Byte2Str(bytes uint64) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= TeraByte:
		unit = "T"
		value = value / TeraByte
	case bytes >= GigaByte:
		unit = "G"
		value = value / GigaByte
	case bytes >= MegaByte:
		unit = "M"
		value = value / MegaByte
	case bytes >= KiloByte:
		unit = "K"
		value = value / KiloByte
	case bytes >= Byte:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}
