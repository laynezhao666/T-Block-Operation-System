// Package cgi 提供门禁系统的HTTP CGI服务注册和路由配置。
package cgi

import (
	"net/http"

	"dac/entity/server/cgi/access"
	"dac/entity/server/cgi/alarm"
	"dac/entity/server/cgi/controller"
	"dac/entity/server/cgi/debug"
	"dac/entity/server/cgi/door"
	"dac/entity/server/cgi/event"
	"dac/entity/server/cgi/group"
	"dac/entity/server/cgi/point"
	"dac/entity/server/cgi/request"
	"dac/entity/server/cgi/room"
	"dac/entity/server/cgi/test"

	"github.com/gin-gonic/gin"
)

// getHandler 创建并配置所有HTTP路由，返回gin引擎作为http.Handler
func getHandler() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	g := gin.Default()

	tdac := g.Group("/api/dcos/tdac-cgi")

	// 机房信息接口
	{
		tdac.GET("/rooms", room.GetAll)
	}

	// 实时测点数据接口
	{
		tdac.POST("/rtd", point.Rtd)
		tdac.POST("/rtd/list", point.RtdList)
	}

	// 门组管理接口（列表）
	{
		dg := tdac.Group("/groups")
		dg.GET("", group.Get)
	}

	// 异步请求管理接口
	{
		r := tdac.Group("/requests")
		r.POST("", request.GetByControllers)
		r.PUT("/update", request.Update)
		r.POST("/outdate", request.Outdate)
		r.DELETE("", request.Delete)
		r.GET("/all", request.GetAll)
		r.POST("/info", request.GetInfo)
		r.POST("/all", request.GetAllRequestInfo)
		r.GET("/methods", request.GetMethods)
		r.POST("/export", request.Export)
		r.GET("/export/all", request.ExportAll)
		r.POST("/re-execute", request.ReExecute)
		r.POST("/batch-re-execute", request.BatchReExecute)
	}

	// 门组管理接口（增删改）
	{
		dg := tdac.Group("/group")
		dg.POST("", group.Create)
		dg.PUT("", group.Update)
		dg.DELETE("", group.Delete)

		dg.POST("/doors", group.GetDoors)
	}

	// 控制器批量操作接口
	{
		c := tdac.Group("/controllers")
		c.GET("", controller.Get)
		c.DELETE("", controller.BatchDelete)

		c.POST("/import", controller.Import)
		c.GET("/export", controller.Export)

		c.POST("/sync-time", controller.AllSyncTime)
		c.POST("/reset", controller.AllReset)
	}

	// 控制器单个操作接口
	{
		c := tdac.Group("/controller")
		c.POST("", controller.Create)
		c.PUT("", controller.Update)
		c.DELETE("", controller.Delete)
		c.POST("/clean", controller.Clean)
		c.POST("/reset", controller.Reset)
		c.DELETE("/time-groups", controller.ClearTimeGroup)
		c.POST("/sync-time", controller.SyncTime)
		c.POST("/card", controller.GetCardFromController) // 从门控器查询卡是否存在

	}

	// 门操作接口（单个）
	{
		d := tdac.Group("/door")

		d.POST("", door.Get)
		d.PUT("", door.Update)

		d.POST("/state", door.SetState)
		d.PUT("/code", door.UpdateCode)
	}

	// 门操作接口（批量）
	{
		d := tdac.Group("/doors")
		d.PUT("", door.BatchUpdate)

		d.GET("/export/code", door.ExportCode)
		d.POST("/import/code", door.ImportCode)

		d.POST("/events", event.GetByDoors)
	}

	// 测试接口
	{
		tdac.POST("/test", test.Ping)
	}

	// 调试接口
	{
		tdac.POST("/debug", debug.Debug)
	}

	// 事件记录接口
	{
		e := tdac.Group("/events")
		e.POST("", event.Get)
		e.POST("/export", event.Export)
	}

	// 告警记录接口
	{
		a := tdac.Group("/alarms")
		a.POST("", alarm.Get)
		a.POST("/export", alarm.Export)
	}

	setAccessHandler(tdac)

	return g
}

// setAccessHandler 配置门禁权限管理相关的路由（时间组、人员、卡片、权限组）
func setAccessHandler(rg *gin.RouterGroup) {
	// 时间组管理接口
	{
		rg.GET("/time-groups", access.GetTimeGroups)

		rg.PUT("/time-group/:group_no", access.UpdateTimeGroup)
		rg.POST("/time-groups/sync", access.SyncTimeGroupToControllers)
	}

	// 人员管理接口
	{
		rg.GET("/staffs", access.GetStaffs)
		rg.GET("/staffs/company", access.GetAllStaffCompany)
		rg.POST("/staffs", access.AddStaff)
		rg.POST("/staffs/import", access.ImportStaffs)
		rg.POST("/staffs/export", access.ExportStaffs)

		rg.PUT("/staff/:id", access.UpdateStaff)
		rg.DELETE("/staff/:id", access.DeleteStaff)
	}

	// 卡片管理接口（批量）
	{
		rg.POST("/cards", access.GetCards)
		rg.DELETE("", access.DeleteCards)
		rg.POST("/cards/import", access.ImportCards)
		rg.POST("/cards/export", access.ExportCards)
	}

	// 卡片管理接口（单个）
	{
		cc := rg.Group("/card")
		cc.POST("", access.AddCard)
		cc.PUT("/flag", access.UpdateCardsFlag)
		cc.PUT("/type", access.UpdateCardsType)
		cc.PUT("/valid_time", access.UpdateCardValidTime)
		cc.DELETE("", access.DeleteCard)
		cc.PUT("/staff", access.UpdateCardStaff)
		cc.PUT("/unbind", access.UnbindCardStaff)
		cc.PUT("/access", access.UpdateCardAccess)
	}

	// 权限组管理接口
	{
		rg.GET("access-groups", access.GetAccessGroups)
		rg.POST("access-groups", access.AddAccessGroup)

		rg.GET("access-groups/card", access.GetAllAccessGroups)

		ag := rg.Group("/access-group/:id")
		ag.PUT("", access.UpdateAccessGroup)
		ag.DELETE("", access.DeleteAccessGroup)
	}

}
