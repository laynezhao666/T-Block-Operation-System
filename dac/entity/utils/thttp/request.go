// Package thttp 提供HTTP请求工具函数，支持文件上传和JSON解析。
package thttp

import (
	"fmt"
	"io"
	"net/http"
)

// Request 发送HTTP请求并返回响应体
func Request(url, method string,
	headers map[string][]string,
	reqBody interface{}, timeout int,
) ([]byte, error) {
	reader, err := getReader(reqBody)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	if len(headers) > 0 {
		for k, vs := range headers {
			for _, v := range vs {
				request.Header.Add(k, v)
			}
		}
	}

	resp, err := getClient(timeout).Do(request)
	if resp != nil && resp.Body != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("status code: %v, body: %v", resp.StatusCode, string(respBody))
	}
	return io.ReadAll(resp.Body)
}

// RequestJSON 发送JSON格式的HTTP请求并解析响应
func RequestJSON(url, method string,
	reqBody interface{}, timeout int,
	dataPointer interface{},
) error {
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}
	b, err := Request(url, method, headers, reqBody, timeout)
	if err != nil {
		return err
	}

	return parseJSONResult(dataPointer, b)
}

// RequestJSONWithHeader 发送带自定义Header的JSON请求并解析响应
func RequestJSONWithHeader(url, method string, header http.Header,
	reqBody interface{}, timeout int, dataPointer interface{},
) error {
	var h http.Header
	if header == nil {
		h = JSONHeader
	} else {
		h = header.Clone()
		h.Set(HeaderContentType, HeaderValueApplicationJSON)
	}
	b, err := Request(url, method, h, reqBody, timeout)
	if err != nil {
		return err
	}

	return parseJSONResult(dataPointer, b)
}
