package utils

import (
	"fmt"

	"github.com/samber/lo"
)

// GetBatchStringList GetBatchStringList
func GetBatchStringList(strList []string, size int) ([][]string, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size: cannot be less than 1")
	}
	return lo.Chunk(strList, size), nil
}
