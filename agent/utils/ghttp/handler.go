package ghttp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 返回结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// SendResponse 返回错误信息
func SendResponse(c *gin.Context, err error, data interface{}) {
	e := GetError(err)
	resp := Response{
		Code:    e.Code,
		Message: e.Message,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// SendResponseWithData 返回数据
func SendResponseWithData(c *gin.Context, data interface{}) {
	resp := Response{
		Code:    ErrorOK.Code,
		Message: ErrorOK.Message,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// SendResponseWithError 返回错误
func SendResponseWithError(c *gin.Context, err error) {
	e := GetError(err)
	resp := Response{
		Code:    e.Code,
		Message: e.Message,
		Data:    nil,
	}
	c.JSON(http.StatusOK, resp)
}

// SendResponseWithCodeMessage 返回错误信息
func SendResponseWithCodeMessage(c *gin.Context, code int, message string) {
	resp := Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
	c.JSON(http.StatusOK, resp)
}

// SendResponseWithErrorMessage 返回错误信息
func SendResponseWithErrorMessage(c *gin.Context, err string) {
	resp := Response{
		Code:    ErrorUnknown,
		Message: err,
		Data:    nil,
	}
	c.JSON(http.StatusOK, resp)
}

// SendResponseWithErrorFormatMessage 返回错误信息
func SendResponseWithErrorFormatMessage(c *gin.Context, format string, a ...interface{}) {
	resp := Response{
		Code:    ErrorUnknown,
		Message: fmt.Sprintf(format, a...),
		Data:    nil,
	}
	c.JSON(http.StatusOK, resp)
}
