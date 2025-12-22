// Package osutil provides various os tools
package osutil

import (
	"os"
	"strconv"
)

// GetIntFromEnv 从环境获取int
func GetIntFromEnv(key string) int {
	vStr := os.Getenv(key)
	v, _ := strconv.Atoi(vStr)
	return v
}
