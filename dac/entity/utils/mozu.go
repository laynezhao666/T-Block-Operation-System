// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HeaderMozuID 请求头中的模组ID字段名
const (
	HeaderMozuID = "mozuid"
)

// GetMozuIDByRawHeader 从原始HTTP请求头中获取模组ID
func GetMozuIDByRawHeader(h http.Header) string {
	return h.Get(http.CanonicalHeaderKey(HeaderMozuID))
}

// GetMozuID 从Gin上下文中获取模组ID
func GetMozuID(c *gin.Context) string {
	return c.GetHeader(http.CanonicalHeaderKey(HeaderMozuID))
}
