package dhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"dac/entity/utils"

	"dac/entity/utils/thttp"
)

var (
	defaultClient = &http.Client{
		Timeout: time.Second * 30,
	}
	clientMap   = make(map[time.Duration]*http.Client)
	clientMutex sync.Mutex
)

func getClient(timeout time.Duration) *http.Client {
	if timeout == 0 {
		return defaultClient
	}
	clientMutex.Lock()
	defer clientMutex.Unlock()

	c, ok := clientMap[timeout]
	if !ok {
		c = &http.Client{
			Timeout: timeout,
		}
		clientMap[timeout] = c
	}
	return c
}

type responseType struct {
	Code    int         `json:"err_code"`
	Message string      `json:"err_msg"`
	Data    interface{} `json:"data"`
}

func requestJSON(url, method string, header map[string][]string, reqBody interface{}, timeout time.Duration,
	dataPointer interface{}) error {

	b, err := thttp.Request(url, method, header, reqBody, int(timeout.Milliseconds()))
	if err != nil {
		return err
	}

	return parseJSONResult(dataPointer, b)
}

func requestJSONWithParseFun(url, method string, header map[string][]string, reqBody interface{}, timeout time.Duration,
	parseFunc func([]byte) error) error {
	b, err := thttp.Request(url, method, header, reqBody, int(timeout.Milliseconds()))
	if err != nil {
		return err
	}

	return parseJSONResultWithParseFunc(b, parseFunc)
}

func requestJSONWithoutContentLength(url, method string, headers map[string][]string, timeout time.Duration) error {
	r, err := http.NewRequest(method, url, bytes.NewBufferString(""))
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, vs := range headers {
			for _, v := range vs {
				r.Header.Add(k, v)
			}
		}
	}

	r.TransferEncoding = []string{"chunked"}

	resp, err := getClient(timeout).Do(r)
	if resp != nil && resp.Body != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("status code: %v, body: %v", resp.StatusCode, string(respBody))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return parseJSONResult(nil, b)
}

func parseJSONResult(dataPointer interface{}, data []byte) error {
	var temp responseType
	temp.Data = dataPointer
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("unmarshal error: %w, response: %v", err, string(data))
	}
	if temp.Code != 0 {
		return fmt.Errorf("code != 0, response: %+v", utils.GetJSONString(temp))
	}

	return nil
}

func parseJSONResultWithParseFunc(data []byte, parseFunc func([]byte) error) error {
	var temp responseType
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("unmarshal error: %w, response: %v", err, string(data))
	}
	if temp.Code != 0 {
		return fmt.Errorf("code != 0, response: %+v", utils.GetJSONString(temp))
	}

	if parseFunc == nil {
		return nil
	}

	b, err := json.Marshal(temp.Data)
	if err != nil {
		return fmt.Errorf("marshal error: %w, response: %+v", err, utils.GetJSONString(temp))
	}
	return parseFunc(b)
}

// doRequestAndParse 执行 HTTP 请求并解析 JSON 响应，提取公共的请求发送和响应处理逻辑
func doRequestAndParse(req *http.Request, timeout time.Duration, dataPointer interface{}) error {
	resp, err := getClient(timeout).Do(req)
	if resp != nil && resp.Body != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return fmt.Errorf("http request error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status code: %v, body: %v", resp.StatusCode, string(respBody))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response error: %w", err)
	}

	return parseJSONResult(dataPointer, b)
}

// requestPureJSON 使用标准库直接发送纯 JSON 请求，避免 thttp 对 string body 进行 URL 编码处理
// 这是为了解决 Base64 数据中的 + 号被错误转换为空格的问题
func requestPureJSON(url string, method string, reqBody interface{},
	timeout time.Duration, dataPointer interface{},
) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return doRequestAndParse(req, timeout, dataPointer)
}

// requestFormJSON 使用标准库发送带 data= 前缀的表单请求，正确处理 URL 编码
// 门控器要求 application/x-www-form-urlencoded 格式，body 为 data={json}
// 通过 url.QueryEscape 对 JSON 进行编码，确保 base64 中的 + 号被编码为 %2B
func requestFormJSON(targetUrl string, method string, reqBody interface{},
	timeout time.Duration, dataPointer interface{},
) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	// 使用 url.QueryEscape 对 JSON 进行 URL 编码，+ 会被编码为 %2B
	encodedJSON := url.QueryEscape(string(jsonData))
	body := "data=" + encodedJSON

	req, err := http.NewRequest(method, targetUrl, bytes.NewReader([]byte(body)))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return doRequestAndParse(req, timeout, dataPointer)
}
