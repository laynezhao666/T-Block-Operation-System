// Package arrayutil provides various slice tools
package arrayutil

import (
	"reflect"
	"sort"
)

type BaseType = interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | ~string
}

// Remove remove target from slice.
func Remove[T any](slice []T, target T) []T {
	res := make([]T, 0, len(slice))
	for i := 0; i < len(slice); i++ {
		if !reflect.DeepEqual(slice[i], target) {
			res = append(res, slice[i])
		}
	}
	return res
}

// RemoveIf remove target from slice if condition is true.
func RemoveIf[T any](slice []T, condFunc func(value T) bool) []T {
	res := make([]T, 0, len(slice))
	for i := 0; i < len(slice); i++ {
		if !condFunc(slice[i]) {
			res = append(res, slice[i])
		}
	}
	return res
}

// Find return first index of target in slice, return -1 if not found.
func Find[T any](slice []T, target T) int {
	for idx, item := range slice {
		if reflect.DeepEqual(item, target) {
			return idx
		}
	}
	return -1
}

// Exist return true if target in slice, return false if not.
func Exist[T any](slice []T, target T) bool {
	return Find(slice, target) >= 0
}

// Filter do filter slice if condition is true.
func Filter[T any](slice []T, filter func(value T) bool) []T {
	res := make([]T, 0, len(slice))
	for _, item := range slice {
		if filter(item) {
			res = append(res, item)
		}
	}
	return res
}

// GroupBy group slice by keyFunc.
func GroupBy[T any, R comparable](slice []T, keyFunc func(val T) R) map[R][]T {
	res := map[R][]T{}
	for _, val := range slice {
		key := keyFunc(val)
		res[key] = append(res[key], val)
	}
	return res
}

// ToMap convert slice to map, key is element, value is index.
func ToMap[T comparable](slice []T) map[T]any {
	res := make(map[T]any)
	for idx, val := range slice {
		res[val] = idx
	}
	return res
}

// Reverse reverse slice.
func Reverse[T any](slice []T) []T {
	res := make([]T, 0, len(slice))
	for idx := len(slice) - 1; idx >= 0; idx-- {
		res = append(res, slice[idx])
	}
	return res
}

// Partition slice into pages.
func Partition[T any](slice []T, size int) [][]T {
	lens := len(slice)
	page := lens / size
	if lens%size != 0 {
		page += 1
	}
	res := make([][]T, page)
	for i := 0; i < page; i++ {
		if (i+1)*size < lens {
			res[i] = slice[i*size : (i+1)*size]
		} else {
			res[i] = slice[i*size : lens]
		}
	}
	return res
}

// Sort make slice sorted with given function.
func Sort[T any](slice []T, less func(T, T) bool) []T {
	sort.Slice(slice, func(i, j int) bool {
		return less(slice[i], slice[j])
	})
	return slice
}

// SortAsc sort slice in ascending order.
func SortAsc[T BaseType](slice []T) []T {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
	return slice
}

// SortDesc sort slice in descending order.
func SortDesc[T BaseType](slice []T) []T {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] > slice[j]
	})
	return slice
}

// Distinct return slice with distinct element.
func Distinct[T comparable](slice []T) []T {
	res := make([]T, 0, len(slice))
	existMap := make(map[T]bool)
	for _, val := range slice {
		if _, ok := existMap[val]; !ok {
			res = append(res, val)
			existMap[val] = true
		}
	}
	return res
}

// Map return slice with mapped element.
func Map[T any, R any](slice []T, mapFunc func(T) R) []R {
	res := make([]R, len(slice))
	for idx, val := range slice {
		res[idx] = mapFunc(val)
	}
	return res
}
