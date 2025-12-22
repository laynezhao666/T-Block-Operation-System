// Package parse provides json parser
package parse

import (
	"reflect"
)

const (
	defaultTagName = "json"
)

var (
	jsonParser = NewParser(Config{
		DisableConvert: false,
		FieldExtractor: NewTagFieldExtractor(defaultTagName),
	})
)

// JSON 根据 tag 中的 json 字段解析值
func JSON(dst, src interface{}) error {
	return jsonParser.Parse(dst, src)
}

// NewTagFieldExtractor 创建 tag 中指定名称的字段提取器
func NewTagFieldExtractor(tagName string) FieldExtractor {
	if len(tagName) == 0 {
		return nil
	}
	return func(field *reflect.StructField) (string, bool) {
		return field.Tag.Lookup(tagName)
	}
}
