// Package debug 提供门禁系统调试接口。
package debug

import (
	"dac/logic/collect/driver/test"
	"dac/logic/dlm"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// Debug 开启调试模式的HTTP接口处理函数。
// 需要当前节点持有分布式锁才能执行。
func Debug(c *gin.Context) {
	if !dlm.GetWorker().HasLock() {
		ghttp.SendResponseWithErrorMessage(c, "has no lock")
		return
	}

	test.EnableDebug()

	ghttp.SendResponseWithData(c, "ok")
}
