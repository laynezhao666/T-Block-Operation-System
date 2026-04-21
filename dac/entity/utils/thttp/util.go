// Package thttp 提供HTTP请求工具函数，支持文件上传和JSON解析。
package thttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// getReader 将请求体转换为io.Reader，支持多种类型
func getReader(requestBody interface{}) (io.Reader, error) {
	if requestBody == nil {
		return bytes.NewReader([]byte{}), nil
	}
	switch body := requestBody.(type) {
	case *bytes.Reader:
		return body, nil
	case bytes.Reader:
		return &body, nil
	case *bytes.Buffer:
		return body, nil
	case bytes.Buffer:
		return &body, nil
	default:
		break
	}

	var b []byte
	var err error
	switch body := requestBody.(type) {
	case string:
		b = []byte(body)
	case []byte:
		b = body
	default:
		if b, err = json.Marshal(body); err != nil {
			return nil, err
		}
	}
	return bytes.NewReader(b), nil
}

// parseJSONResult 解析JSON响应并检查返回码
func parseJSONResult(dataPointer interface{}, responseBody []byte) error {
	var temp responseType
	if dataPointer != nil {
		temp.Data = dataPointer
	}
	if err := json.Unmarshal(responseBody, &temp); err != nil {
		return err
	}
	if temp.Code != 0 {
		return fmt.Errorf("code != 0, response: %+v", temp)
	}

	return nil
}
