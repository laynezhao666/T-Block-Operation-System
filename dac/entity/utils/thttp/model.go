// Package thttp 提供HTTP请求和响应的通用模型和常量。
package thttp

// responseType HTTP响应通用结构
type responseType struct {
	Code    int         `json:"code"`    // 响应码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// HTTP头部常量
const (
	HeaderContentType          = "Content-Type"
	HeaderValueApplicationJSON = "application/json"
)

// JSONHeader JSON格式的HTTP请求头
var (
	JSONHeader = map[string][]string{
		HeaderContentType: {HeaderValueApplicationJSON},
	}
)
