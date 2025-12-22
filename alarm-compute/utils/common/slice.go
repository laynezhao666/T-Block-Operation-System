package common

import (
	"fmt"

	"github.com/samber/lo"
)

// ChunkStringList ChunkStringList
func ChunkStringList(strList []string, size int) ([][]string, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size: cannot be less than 1")
	}
	return lo.Chunk(strList, size), nil
}

// UniqueStringSlice UniqueStringSlice
func UniqueStringSlice(es []string) []string {
	m := make(map[string]struct{})
	for _, e := range es {
		m[e] = struct{}{}
	}

	newEs := []string{}
	for e := range m {
		newEs = append(newEs, e)
	}
	return newEs
}

// RemoveEleFromSlice 删除切片中的元素，不保证元素顺序
func RemoveEleFromSlice(s []string, element string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == element {
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
			i-- // 重新检查这个位置，因为它现在包含了最后一个元素
		}
	}
	return s
}
