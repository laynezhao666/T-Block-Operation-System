// Package group 提供门组管理的HTTP接口处理器。
package group

import (
	"dac/entity/utils"
	"dac/logic/cgi/group"
	"dac/repo/dac"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// Delete 删除指定ID的门组
func Delete(c *gin.Context) {
	var (
		req struct {
			ID int `json:"id"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = group.Delete(c, req.ID); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// Update 更新门组的名称和关联的门列表
func Update(c *gin.Context) {
	var (
		req struct {
			Name  string `json:"name"`
			ID    int    `json:"id"`
			Doors []int  `json:"doors"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = group.Update(c, req.ID, req.Name, req.Doors); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// Create 创建新的门组
func Create(c *gin.Context) {
	var (
		req struct {
			Name  string `json:"name"`
			Doors []int  `json:"doors"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = group.Create(c, utils.GetMozuID(c), req.Name, req.Doors); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// Get 获取所有门组列表
func Get(c *gin.Context) {
	groups, err := dac.GetRW().GetAllDoorGroups(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, groups)
}

// GetDoors 获取指定门组下的所有门
func GetDoors(c *gin.Context) {
	var (
		req struct {
			ID int `json:"id"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	doors, err := group.GetGroupDoors(c, req.ID)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, doors)
}
