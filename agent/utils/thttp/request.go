package thttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Request 发送http请求
func Request(url, method string, headers map[string][]string, reqBody interface{}, timeout int) ([]byte, error) {
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
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("status code: %v, body: %v", resp.StatusCode, string(respBody))
	}
	return ioutil.ReadAll(resp.Body)
}

// RequestJSON 发送http请求
func RequestJSON(url, method string, reqBody interface{}, timeout int, dataPointer interface{}) error {
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}
	b, err := Request(url, method, headers, reqBody, timeout)
	if err != nil {
		return err
	}

	return parseJSONResult(dataPointer, b)
}
