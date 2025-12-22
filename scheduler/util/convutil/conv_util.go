// Package convutil 一些不同类型之间转化的工具类
package convutil

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JsonStrToMap 将Json字符串直接转化为map，忽略错误信息
func JsonStrToMap(jsonStr string) map[string]any {
	res := make(map[string]any)
	if len(jsonStr) > 0 {
		_ = json.Unmarshal([]byte(jsonStr), &res)
	}
	return res
}

// JsonMarshalIgnore 将任意对象转化为json,错误返回空字符串
func JsonMarshalIgnore(source any) string {
	data, err := json.Marshal(source)
	if err != nil {
		return ""
	}
	return string(data)
}

// SliceToStr 将Slice转化为字符串
func SliceToStr[T any](data []T, seq string) string {
	strList := make([]string, 0, len(data))
	for _, item := range data {
		strList = append(strList, fmt.Sprint(item))
	}
	return strings.Join(strList, seq)
}
