// Package access 提供门禁权限管理的HTTP接口处理器。
package access

import (
	"strconv"

	"dac/entity/model/cgi"
	"dac/entity/utils"
	"dac/logic/cgi/timegroup"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// GetTimeGroups 获取所有时间组配置
func GetTimeGroups(c *gin.Context) {
	timeGroups, err := timegroup.GetAll(c)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, timeGroups)
}

// UpdateTimeGroup 更新指定编号的时间组配置
func UpdateTimeGroup(c *gin.Context) {
	groupNo, err := strconv.Atoi(c.Param("group_no"))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	var timeGroup cgi.TimeGroup
	if err = c.ShouldBind(&timeGroup); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	err = timegroup.Update(c, groupNo, timeGroup)
	ghttp.SendResponseWithError(c, err)
}

// SyncTimeGroupToControllers 将时间组配置同步到所有控制器
func SyncTimeGroupToControllers(c *gin.Context) {
	err := timegroup.Sync(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}
