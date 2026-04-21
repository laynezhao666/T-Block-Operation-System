package parse

import (
	"reflect"
)

// FieldExtractor 从结构体字段中提取字段名
type FieldExtractor func(*reflect.StructField) (string, bool)

// Config 用于解析过程的相关配置
type Config struct {
	DisableConvert bool           // 是否禁止自动转换，例如：int <-> float32
	FieldExtractor FieldExtractor // 字段提取器
}

type parser struct {
	config Config
}

// NewParser 创建自定义解析器
func NewParser(c Config) *parser {
	if c.FieldExtractor == nil {
		return nil
	}
	p := new(parser)
	p.config = c
	return p
}
