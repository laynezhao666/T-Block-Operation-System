package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// RestfulSerializer 是为restful接口单独实现的序列化，
type RestfulSerializer struct {
}

// Name implements Serializer.
func (c RestfulSerializer) Name() string {
	return "application/json"
}

// ContentType implements Serializer.
func (c RestfulSerializer) ContentType() string {
	return "application/json"
}

// Marshal implements Serializer.
func (c RestfulSerializer) Marshal(v interface{}) ([]byte, error) {
	val := reflect.ValueOf(v)
	processed := c.processValue(val)
	return c.marshalJSON(processed)
}

// 递归处理所有类型
func (c RestfulSerializer) processValue(val reflect.Value) interface{} {
	// 解引用指针和接口
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		return c.processStruct(val)
	case reflect.Map:
		return c.processMap(val)
	case reflect.Slice, reflect.Array:
		return c.processSlice(val)
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint()
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.Bool:
		return val.Bool()
	default:
		return nil // 未知类型返回空
	}
}

// 处理结构体（忽略 omitempty）
func (c RestfulSerializer) processStruct(val reflect.Value) map[string]interface{} {
	res := make(map[string]interface{})
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" || !val.Field(i).CanInterface() {
			continue // 跳过非导出字段
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		// 解析字段名（忽略 omitempty）
		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]
		if fieldName == "" {
			fieldName = field.Name
		}

		// 递归处理字段值
		fieldVal := c.processValue(val.Field(i))
		res[fieldName] = fieldVal
	}
	return res
}

// 处理 map
func (c RestfulSerializer) processMap(val reflect.Value) map[string]interface{} {
	res := make(map[string]interface{})
	for _, key := range val.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		res[keyStr] = c.processValue(val.MapIndex(key))
	}
	return res
}

// 处理切片/数组
func (c RestfulSerializer) processSlice(val reflect.Value) []interface{} {
	res := make([]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		res[i] = c.processValue(val.Index(i))
	}
	return res
}

// 手动生成 JSON（不依赖标准库的 omitempty）
func (c RestfulSerializer) marshalJSON(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return c.marshalObject(v)
	case []interface{}:
		return c.marshalArray(v)
	case string:
		return json.Marshal(v)
	case float64, int, int64, bool, uint64, uint:
		return json.Marshal(v) // 基本类型直接序列化
	case nil:
		return []byte("null"), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

func (c RestfulSerializer) marshalObject(obj map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	for key, value := range obj {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		buf.WriteString(`"` + key + `":`)
		b, err := c.marshalJSON(value)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (c RestfulSerializer) marshalArray(arr []interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, item := range arr {
		if i > 0 {
			buf.WriteByte(',')
		}
		b, err := c.marshalJSON(item)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

// Unmarshal implements Serializer.
func (c RestfulSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
