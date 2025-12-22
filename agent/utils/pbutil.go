package utils

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// ConvertMapToStruct 将 map[string]interface{} 转换为 *structpb.Struct
func ConvertMapToStruct(m map[string]interface{}) (*structpb.Struct, error) {
	fields := make(map[string]*structpb.Value)

	for key, value := range m {
		var structValue *structpb.Value

		switch v := value.(type) {
		case string:
			structValue = structpb.NewStringValue(v)
		case float64:
			structValue = structpb.NewNumberValue(v)
		case bool:
			structValue = structpb.NewBoolValue(v)
		case []interface{}:
			// 处理数组
			arrayValue := make([]*structpb.Value, len(v))
			for i, item := range v {
				// 递归处理每个数组元素
				itemValue, err := convertValueToStructValue(item)
				if err != nil {
					return nil, err
				}
				arrayValue[i] = itemValue
			}
			structValue = structpb.NewListValue(&structpb.ListValue{Values: arrayValue})
		case map[string]interface{}:
			// 处理嵌套的 map
			nestedStruct, err := ConvertMapToStruct(v)
			if err != nil {
				return nil, err
			}
			structValue = structpb.NewStructValue(nestedStruct)
		default:
			return nil, fmt.Errorf("unsupported type: %T", v)
		}

		fields[key] = structValue
	}

	return &structpb.Struct{Fields: fields}, nil
}

func convertValueToStructValue(value interface{}) (*structpb.Value, error) {
	switch v := value.(type) {
	case string:
		return structpb.NewStringValue(v), nil
	case float64:
		return structpb.NewNumberValue(v), nil
	case bool:
		return structpb.NewBoolValue(v), nil
	case []interface{}:
		// 处理数组
		arrayValue := make([]*structpb.Value, len(v))
		for i, item := range v {
			itemValue, err := convertValueToStructValue(item)
			if err != nil {
				return nil, err
			}
			arrayValue[i] = itemValue
		}
		return structpb.NewListValue(&structpb.ListValue{Values: arrayValue}), nil
	case map[string]interface{}:
		// 处理嵌套的 map
		nestedStruct, err := ConvertMapToStruct(v)
		if err != nil {
			return nil, err
		}
		return structpb.NewStructValue(nestedStruct), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}
