// Package event 提供门禁事件记录的HTTP接口处理器。
package event

import (
	"fmt"

	"dac/entity/config"
	"dac/entity/model/cgi"
	"dac/entity/utils"
	"dac/logic/cgi/event"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// GetByDoors 按门查询进出事件记录
func GetByDoors(c *gin.Context) {
	var (
		req struct {
			Controllers []struct {
				ControllerID int   `json:"id"`
				DoorNumbers  []int `json:"doors"`
			} `json:"controllers"`
			cgi.QueryCondition
		}
		err error
	)

	// 绑定请求参数
	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	// 构建控制器-门编号映射
	controllerDoors := make(map[int][]int, len(req.Controllers))
	for _, c := range req.Controllers {
		controllerDoors[c.ControllerID] = make([]int, 0, len(c.DoorNumbers))
		for _, d := range c.DoorNumbers {
			controllerDoors[c.ControllerID] = append(controllerDoors[c.ControllerID], d)
		}
	}

	// 查询事件并构建响应
	n, events, err := event.GetByDoors(c, controllerDoors, req.QueryCondition)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	var resp struct {
		Total int         `json:"total"`
		List  []cgi.Event `json:"list"`
	}
	resp.Total = int(n)
	resp.List = events
	ghttp.SendResponseWithData(c, resp)
}

// Export 导出进出事件记录为Excel文件
func Export(c *gin.Context) {
	var (
		req struct {
			DoorName    string `json:"door_name"`
			Controllers []struct {
				ControllerID int   `json:"id"`
				DoorNumbers  []int `json:"doors"`
			} `json:"controllers"`
			cgi.TimeCondition
			Query string `json:"query"`
		}
		err error
	)

	// 绑定请求参数
	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	// 构建控制器-门编号映射
	controllerDoors := make(map[int][]int, len(req.Controllers))
	for _, ct := range req.Controllers {
		controllerDoors[ct.ControllerID] = make([]int, 0, len(ct.DoorNumbers))
		for _, d := range ct.DoorNumbers {
			controllerDoors[ct.ControllerID] = append(controllerDoors[ct.ControllerID], d)
		}
	}

	// 导出Excel文件
	f, err := event.Export(c, utils.GetMozuID(c), req.DoorName, controllerDoors, req.TimeCondition, req.Query)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	// 设置文件下载响应头
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", "进出记录.xlsx"))

	if err = f.Write(c.Writer); err != nil {
		config.Log.Warnf("write events excel error: %v", err)
	}
}

// Get 查询进出事件记录，支持分页和条件过滤
func Get(c *gin.Context) {
	var (
		req struct {
			ControllerIDs []int  `json:"controller_ids"`
			Query         string `json:"query"`
			DoorName      string `json:"door_name"`
			cgi.QueryCondition
		}
		err error
	)

	// 绑定请求参数
	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	// 查询事件并构建响应
	n, events, err := event.Get(c, utils.GetMozuID(c), req.ControllerIDs, req.Query, req.DoorName, req.QueryCondition)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	var resp struct {
		Total int         `json:"total"`
		List  []cgi.Event `json:"list"`
	}
	resp.Total = int(n)
	resp.List = events
	ghttp.SendResponseWithData(c, resp)
}
