// Package dhttp 提供门禁控制器HTTP协议的请求封装。
// 支持标准form-urlencoded和纯JSON两种请求格式。
package dhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FormContentHeader 表单编码请求头
// JSONContentHeader JSON编码请求头
var (
	FormContentHeader = map[string][]string{"Content-Type": {"application/x-www-form-urlencoded"}}
	// JSONContentHeader 用于发送纯 JSON 请求，避免 base64 数据被 URL 解码破坏
	JSONContentHeader = map[string][]string{"Content-Type": {"application/json"}}
)

// GetBody 构造带 data= 前缀的请求体（标准协议）
func GetBody(req interface{}) (string, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return "data=" + string(b), nil
}

// GetBodyWithoutDataPrefix 构造不带 data= 前缀的请求体（非标准协议）
func GetBodyWithoutDataPrefix(req interface{}) (string, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetJSON 发送GET请求并将响应解析为JSON
func GetJSON(url string, timeout time.Duration, dataPointer interface{}) error {
	return requestJSON(url, http.MethodGet, nil, nil, timeout, dataPointer)
}

// GetJSONWithParseFunc 发送GET请求并使用自定义函数解析响应
func GetJSONWithParseFunc(url string, timeout time.Duration, parseFunc func([]byte) error) error {
	return requestJSONWithParseFun(url, http.MethodGet, nil, nil, timeout, parseFunc)
}

// PostJSON 发送带data=前缀的POST请求并解析JSON响应
func PostJSON(url string, timeout time.Duration, reqBody interface{}, dataPointer interface{}) error {
	data, err := GetBody(reqBody)
	if err != nil {
		return err
	}

	if err = requestJSON(url, http.MethodPost, FormContentHeader, data, timeout, dataPointer); err != nil {
		return fmt.Errorf("post %v, data: '%v', error: %w", url, data, err)
	}
	return nil
}

// PostJSONWithoutDataPrefix 发送不带 data= 前缀的 POST 请求（非标准协议）
func PostJSONWithoutDataPrefix(url string, timeout time.Duration, reqBody interface{}, dataPointer interface{}) error {
	data, err := GetBodyWithoutDataPrefix(reqBody)
	if err != nil {
		return err
	}

	if err = requestJSON(url, http.MethodPost, FormContentHeader, data, timeout, dataPointer); err != nil {
		return fmt.Errorf("post %v, data: '%v', error: %w", url, data, err)
	}
	return nil
}

// PostJSONWithoutData 发送不带请求体的POST请求
func PostJSONWithoutData(url string, timeout time.Duration) error {
	return requestJSON(url, http.MethodPost, nil, nil, timeout, nil)
}

// PostJSONWithoutContentLength 发送不带Content-Length的POST请求
func PostJSONWithoutContentLength(url string, timeout time.Duration) error {
	return requestJSONWithoutContentLength(url, http.MethodPost, nil, timeout)
}

// PostPureJSON 发送纯 JSON 请求（Content-Type: application/json，无 data= 前缀）
// 用于 HTTP v3 协议门控器包含人脸图片等 base64 数据的请求，避免 + 号被解码为空格
func PostPureJSON(url string, timeout time.Duration, reqBody interface{}, dataPointer interface{}) error {
	// 直接传入原始结构体，使用标准库发送，避免 thttp 对 string body 的 URL 编码处理
	if err := requestPureJSON(url, http.MethodPost, reqBody, timeout, dataPointer); err != nil {
		return fmt.Errorf("post pure json %v error: %w", url, err)
	}
	return nil
}

// PostFormJSON 发送带 data= 前缀的 JSON 请求，使用正确的 URL 编码
// 用于门控器要求 application/x-www-form-urlencoded 格式但包含 base64 数据的场景
// 与 PostJSON 的区别是：这里会对整个 JSON 进行 URL 编码，确保 + 号不会被错误解析
func PostFormJSON(targetUrl string, timeout time.Duration, reqBody interface{}, dataPointer interface{}) error {
	if err := requestFormJSON(targetUrl, http.MethodPost, reqBody, timeout, dataPointer); err != nil {
		return fmt.Errorf("post form json %v error: %w", targetUrl, err)
	}
	return nil
}
