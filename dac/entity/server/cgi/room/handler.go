// Package room 提供机房信息的HTTP接口处理器。
package room

import (
	"dac/entity/utils"
	"dac/logic/cgi/room"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// GetAll 获取包含当前模组的所有楼栋及机房信息
func GetAll(c *gin.Context) {
	rooms, err := room.GetAllRoomsByBuildingContainingMozu(
		c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	ghttp.SendResponseWithData(c, rooms)
}
