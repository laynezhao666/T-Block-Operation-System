// Package typeutil provides various type tools
package typeutil

import (
	"reflect"
)

// Array2Interface 各种类型的 slice 转化为 interface slice
func Array2Interface(arr interface{}) (ret []interface{}) {
	if arr == nil {
		return nil
	}

	if reflect.TypeOf(arr).Kind() != reflect.Slice && reflect.TypeOf(arr).Kind() != reflect.Array {
		return nil
	}

	s := reflect.ValueOf(arr)
	ret = make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}
