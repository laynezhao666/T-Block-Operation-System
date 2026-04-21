// Package request 提供异步请求管理的HTTP接口处理器。
package request

import (
	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/cgi"
	"dac/entity/utils"
	"dac/entity/utils/ghttp"
	"dac/entity/utils/ttime"
	"dac/logic/cgi/request"

	"fmt"

	"github.com/gin-gonic/gin"
)

// GetAll 获取当前模组的所有异步请求
func GetAll(c *gin.Context) {
	r, err := request.GetAll(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, r)
}

// GetInfo 分页查询异步请求，支持多条件过滤
func GetInfo(c *gin.Context) {
	var (
		req struct {
			Offset          int    `json:"offset"`
			Limit           int    `json:"limit"`
			Query           string `json:"query"`
			BeginTime       int64  `json:"begin_time"`
			EndTime         int64  `json:"end_time"`
			QueryCreateTime bool   `json:"query_create_time"`
			State           string `json:"state"`
			QueryState      bool   `json:"query_state"`
			Method          string `json:"method"`
			QueryMethod     bool   `json:"query_method"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	requests, err := request.GetRequests(c, utils.GetMozuID(c), req.Offset, req.Limit, req.Query, req.BeginTime,
		req.EndTime, req.QueryCreateTime, req.State, req.QueryState, req.Method, req.QueryMethod)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, requests)
}

// GetAllRequestInfo 分页查询异步请求详情（含控制器信息）
func GetAllRequestInfo(c *gin.Context) {
	var (
		req struct {
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
			Query  string `json:"query"`
			Method string `json:"method"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	n, requestInfo, err := request.GetAllRequestWithControllerInfo(
		c, utils.GetMozuID(c), req.Offset, req.Limit, req.Query, req.Method,
	)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	var resp struct {
		Total int               `json:"total"`
		List  []cgi.RequestInfo `json:"list"`
	}
	resp.Total = int(n)
	resp.List = requestInfo
	ghttp.SendResponseWithData(c, resp)
}

// Export 导出指定ID的异步请求为Excel文件
func Export(c *gin.Context) {
	var (
		req struct {
			RequestIDs []int `json:"request_ids"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	f, err := request.Export(c, utils.GetMozuID(c), req.RequestIDs)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", "异步消息.xlsx"))

	if err = f.Write(c.Writer); err != nil {
		config.Log.Warnf("write requests excel error: %v", err)
	}

}

// ExportAll 导出所有异步请求为Excel文件
func ExportAll(c *gin.Context) {

	f, err := request.ExportAll(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", "异步消息.xlsx"))

	if err = f.Write(c.Writer); err != nil {
		config.Log.Warnf("write requests excel error: %v", err)
	}

}

// GetByControllers 按控制器ID查询异步请求
func GetByControllers(c *gin.Context) {
	var (
		req struct {
			ControllerID int `json:"controller_id"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	r, err := request.GetByControllers(c, req.ControllerID)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, r)
}

// Delete 批量删除异步请求
func Delete(c *gin.Context) {
	var (
		req struct {
			IDs []int `json:"ids"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = request.Delete(c, req.IDs); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// GetMethods 获取所有可用的请求方法列表
func GetMethods(c *gin.Context) {
	r, err := request.GetMethods(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, r)
}

// ReExecute 重新执行指定的异步请求
func ReExecute(c *gin.Context) {
	var (
		req struct {
			IDs     []int  `json:"ids"`
			Method  string `json:"method"`
			Payload string `json:"payload"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	t := ttime.GetNowUTC().UnixMilli()
	if err = request.Update(c, req.IDs, req.Method, req.Payload, t, consts.StateToBeExecuted); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	config.Log.Infof("re-execute request: %v, method: %v, payload: %v, create time: %v",
		req.IDs, req.Method, req.Payload, t)

	ghttp.SendResponseWithData(c, "ok")
}

// BatchReExecute 批量重新执行异步请求
func BatchReExecute(c *gin.Context) {
	var (
		req struct {
			IDs []int `json:"ids"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	t := ttime.GetNowUTC().UnixMilli()
	if err = request.BatchReExecute(c, req.IDs, t, consts.StateToBeExecuted); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	config.Log.Infof("batch-re-execute request: %v, create time: %v", req.IDs, t)

	ghttp.SendResponseWithData(c, "ok")
}

// Update 更新异步请求的方法、负载和状态
func Update(c *gin.Context) {
	var (
		req struct {
			IDs        []int  `json:"ids"`
			Method     string `json:"method"`
			Payload    string `json:"payload"`
			CreateTime int64  `json:"create_time"`
			State      string `json:"state"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = request.Update(c, req.IDs, req.Method, req.Payload, req.CreateTime, req.State); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	config.Log.Infof("update request: %v, method: %v, payload: %v, create time: %v",
		req.IDs, req.Method, req.Payload, req.CreateTime)

	ghttp.SendResponseWithData(c, "ok")
}

// Outdate 将指定请求标记为过期
func Outdate(c *gin.Context) {
	var (
		req struct {
			IDs []int `json:"ids"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = request.Outdate(c, req.IDs); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	config.Log.Infof("outdate request: %v", req.IDs)

	ghttp.SendResponseWithData(c, "ok")
}
