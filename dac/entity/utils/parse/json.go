// Package parse 提供基于struct tag的数据解析工具。
package parse

import (
	"reflect"
	"strings"
)

// defaultTagName 默认使用的struct tag名称
const (
	defaultTagName = "json"
)

// jsonParser 全局JSON解析器实例
var (
	jsonParser = NewParser(Config{
		DisableConvert: false,
		FieldExtractor: NewTagFieldExtractor(defaultTagName),
	})
)

// JSON 根据json tag解析src到dst
func JSON(dst, src interface{}) error {
	return jsonParser.Parse(dst, src)
}

// NewTagFieldExtractor 创建基于struct tag的字段名提取器
func NewTagFieldExtractor(tagName string) FieldExtractor {
	if len(tagName) == 0 {
		return nil
	}
	return func(field *reflect.StructField) (string, bool) {
		v, ok := field.Tag.Lookup(tagName)
		if !ok {
			return "", false
		}

		pos := strings.Index(v, ",")
		if pos < 0 {
			return v, true
		}

		return v[:pos], true
	}
}
