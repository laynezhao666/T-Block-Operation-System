// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"strings"

	"dac/entity/consts"
)

// errorCodeIndexInvalidType 索引号无效的错误码描述
const errorCodeIndexInvalidType = "索引号无效"

// outOfRangeWords 参数超出范围的关键词列表
// underflowWords 索引号下溢的关键词列表
// invalidWords 索引号无效的关键词列表
var (
	outOfRangeWords = []string{
		"参数超出范围", "Parameter value out of range",
		"缺少必要的参数"}
	underflowWords = []string{
		"不存在的刷卡记录索引号", "不存在的告警记录索引号"}
	invalidWords = []string{
		"查询门禁刷卡记录错误", errorCodeIndexInvalidType,
		"缺少必要的参数"}
)

// IsIndexInvalidType 判断错误是否为索引号无效类型
func IsIndexInvalidType(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	for _, w := range invalidWords {
		if strings.Contains(errStr, w) {
			return true
		}
	}

	return false
}

// IsRecordIndexOutOfRange 判断错误是否为记录索引超出范围
func IsRecordIndexOutOfRange(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	for _, w := range outOfRangeWords {
		if strings.Contains(errStr, w) {
			return true
		}
	}

	return false
}

// IsRecordIndexUnderflow 判断错误是否为记录索引下溢
func IsRecordIndexUnderflow(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	for _, w := range underflowWords {
		if strings.Contains(errStr, w) {
			return true
		}
	}

	return false
}

// IsFetchByTimestamp 判断是否按时间戳方式拉取数据
func IsFetchByTimestamp(
	extend map[string]interface{},
) bool {
	if extend == nil {
		return false
	}

	v, ok := extend[consts.KeySyncedByTimestamp]
	if !ok {
		return false
	}

	switch vv := v.(type) {
	case bool:
		return vv
	case int:
		return vv > 0
	case float32:
		return vv > 0.0
	case float64:
		return vv > 0.0
	default:
		return false
	}
}
