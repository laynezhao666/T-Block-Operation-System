// Package ghttp 提供HTTP响应的统一封装工具。
package ghttp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response HTTP统一响应结构体
type Response struct {
	Code    int         `json:"code"`    // 业务状态码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// SendResponse 发送带错误和数据的响应
func SendResponse(c *gin.Context, err error, data interface{}) {
	e := GetError(err)
	resp := Response{
		Code:    e.Code,
		Message: e.Message,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// SendResponseWithData 发送成功响应（带数据）
func SendResponseWithData(c *gin.Context, data interface{}) {
	resp := Response{
		Code:    ErrorOK.Code,
		Message: ErrorOK.Message,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// SendResponseWithError 发送错误响应
func SendResponseWithError(c *gin.Context, err error) {
	e := GetError(err)
	resp := Response{
		Code:    e.Code,
		Message: e.Message,
		Data:    nil,
	}
	c.JSON(http.StatusOK, resp)
}

// SendResponseWithCodeMessage 发送自定义状态码和消息的响应
func SendResponseWithCodeMessage(
	c *gin.Context, code int, message string,
) {
	resp := Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
	c.JSON(http.StatusOK, resp)
}

// SendResponseWithErrorMessage 发送自定义错误消息的响应
func SendResponseWithErrorMessage(
	c *gin.Context, err string,
) {
	resp := Response{
		Code:    ErrorUnknown,
		Message: err,
		Data:    nil,
	}
	c.JSON(http.StatusOK, resp)
}

// SendResponseWithErrorFormatMessage 发送格式化错误消息的响应
func SendResponseWithErrorFormatMessage(
	c *gin.Context, format string, a ...interface{},
) {
	resp := Response{
		Code:    ErrorUnknown,
		Message: fmt.Sprintf(format, a...),
		Data:    nil,
	}
	c.JSON(http.StatusOK, resp)
}
