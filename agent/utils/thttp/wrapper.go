package thttp

import (
	"net/http"
)

// GetJSON get请求
func GetJSON(url string, timeout int, dataPointer interface{}) error {
	return RequestJSON(url, http.MethodGet, nil, timeout, dataPointer)
}

// PostJSON post请求
func PostJSON(url string, reqBody interface{}, timeout int, dataPointer interface{}) error {
	return RequestJSON(url, http.MethodPost, reqBody, timeout, dataPointer)
}

// PutJSON put请求
func PutJSON(url string, reqBody interface{}, timeout int, dataPointer interface{}) error {
	return RequestJSON(url, http.MethodPut, reqBody, timeout, dataPointer)
}
