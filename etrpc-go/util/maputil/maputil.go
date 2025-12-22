// Package maputil provides various map tools
package maputil

import (
	"reflect"
)

// Merge 合并一个或多个 map[string]string
// 注意：后面的 map 将覆盖前面的 map
func Merge(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

// Keys 返回 map 的所有key，这里 key 必须是 string
func Keys(input interface{}) []string {
	if input == nil {
		return []string{}
	}

	if reflect.TypeOf(input).Kind() != reflect.Map {
		// TODO: 考虑打印日志或返回error
		return nil
	}

	m := reflect.ValueOf(input)

	ret := make([]string, m.Len())

	var i int
	for _, key := range m.MapKeys() {
		if key.Kind() != reflect.String {
			// TODO: 考虑打印日志或返回error
			return nil
		}

		ret[i] = key.String()
		i++
	}

	return ret
}
