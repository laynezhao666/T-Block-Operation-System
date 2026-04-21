// Package thttp 提供HTTP请求的便捷封装，支持GET/POST/PUT等方法。
package thttp

import (
	"net/http"
)

// GetJSON 发送GET请求并将响应解析为JSON
func GetJSON(url string, timeout int, dataPointer interface{}) error {
	return RequestJSON(url, http.MethodGet, nil, timeout, dataPointer)
}

// GetJSONWithHeader 发送带自定义Header的GET请求并解析JSON响应
func GetJSONWithHeader(url string, header http.Header, timeout int, dataPointer interface{}) error {
	return RequestJSONWithHeader(url, http.MethodGet, header, nil, timeout, dataPointer)
}

// PostJSON 发送POST请求并将响应解析为JSON
func PostJSON(url string, reqBody interface{}, timeout int, dataPointer interface{}) error {
	return RequestJSON(url, http.MethodPost, reqBody, timeout, dataPointer)
}

// PostJSONWithHeader 发送带自定义Header的POST请求并解析JSON响应
func PostJSONWithHeader(url string, header http.Header,
	reqBody interface{}, timeout int, dataPointer interface{},
) error {
	return RequestJSONWithHeader(url, http.MethodPost, header, reqBody, timeout, dataPointer)
}

// PutJSON 发送PUT请求并将响应解析为JSON
func PutJSON(url string, reqBody interface{}, timeout int, dataPointer interface{}) error {
	return RequestJSON(url, http.MethodPut, reqBody, timeout, dataPointer)
}

// PutJSONWithHeader 发送带自定义Header的PUT请求并解析JSON响应
func PutJSONWithHeader(url string, header http.Header,
	reqBody interface{}, timeout int, dataPointer interface{},
) error {
	return RequestJSONWithHeader(url, http.MethodPut, header, reqBody, timeout, dataPointer)
}
