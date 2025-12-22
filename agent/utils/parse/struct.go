package parse

import "reflect"

var (
	structParser = NewParser(Config{
		DisableConvert: false,
		FieldExtractor: func(field *reflect.StructField) (string, bool) {
			return field.Name, true
		},
	})
)

// Struct 直接根据字段名解析值
func Struct(dst, src interface{}) error {
	return structParser.Parse(dst, src)
}
