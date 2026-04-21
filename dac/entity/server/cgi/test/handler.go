// Package test 提供门禁系统测试相关的HTTP接口处理器。
package test

import (
	"dac/entity/model/rt"
	"dac/logic/cgi/test"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// Ping 测试控制器网络连通性
func Ping(c *gin.Context) {
	var (
		req rt.PingArgs
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = test.Ping(req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}
