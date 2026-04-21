// Package access 提供门禁权限管理的HTTP接口处理器。
package access

import (
	"strconv"

	"dac/entity/model/db"
	"dac/entity/utils"
	"dac/logic/cgi/accessgroup"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// GetAccessGroups 分页查询权限组列表
func GetAccessGroups(c *gin.Context) {
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	result, err := accessgroup.Get(c, utils.GetMozuID(c), offset, limit)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	ghttp.SendResponseWithData(c, result)
}

// AddAccessGroup 新增权限组
func AddAccessGroup(c *gin.Context) {
	var accessGroupWrapper db.AccessGroupInfoWrapper
	var err error
	if err = c.ShouldBind(&accessGroupWrapper); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	var (
		data struct {
			ID db.IDType `json:"id"`
		}
	)
	if data.ID, err = accessgroup.Add(c, utils.GetMozuID(c), accessGroupWrapper); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, data)
}

// UpdateAccessGroup 更新指定ID的权限组
func UpdateAccessGroup(c *gin.Context) {
	accessGroupId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	var accessGroupWrapper db.AccessGroupInfoWrapper
	if err = c.ShouldBind(&accessGroupWrapper); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	err = accessgroup.Update(c, accessGroupId, utils.GetMozuID(c), accessGroupWrapper)
	ghttp.SendResponseWithError(c, err)
}

// DeleteAccessGroup 删除指定ID的权限组
func DeleteAccessGroup(c *gin.Context) {
	accessGroupId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	err = accessgroup.Delete(c, accessGroupId, utils.GetMozuID(c))
	ghttp.SendResponseWithError(c, err)
}

// GetAllAccessGroups 获取所有权限组（不分页）
func GetAllAccessGroups(c *gin.Context) {
	groups, err := accessgroup.GetAllCardGroups(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	ghttp.SendResponseWithData(c, groups)
}
