// Package driver 提供门禁驱动层的通用数据模型和工具函数。
package driver

import (
	"encoding/json"
)

// Marshal 将驱动数据结构序列化为JSON字节流
func Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// Unmarshal 将JSON字节流反序列化到驱动数据结构
func Unmarshal(dataPointer interface{}, b []byte) error {
	return json.Unmarshal(b, dataPointer)
}
