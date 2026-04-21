// Package request 提供异步请求消息的查询、管理和导出功能。
package request

import (
	"context"
	"fmt"
	"time"

	"dac/entity/config"
	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/repo/dac"

	"dac/entity/utils/excel"
	"github.com/tealeg/xlsx/v3"
)

// ht 默认Excel行高
// layout 时间格式化模板
const (
	ht     = 14.0
	layout = "2006-01-02 15:04:05"
)

// titles 异步消息导出Excel表头
var (
	titles = []string{"序号", "消息类型", "消息内容", "创建时间", "门禁控制器名称", "状态", "实际执行时间"}
)

// GetByControllers 根据控制器ID查询异步请求
func GetByControllers(ctx context.Context, controllerID int) ([]db.Request, error) {
	return dac.GetRW().GetRequestsByControllers(ctx, controllerID)
}

// Delete 批量删除异步请求
func Delete(ctx context.Context, ids []int) error {
	return dac.GetRW().DeleteRequests(ctx, ids)
}

// Update 更新异步请求信息
func Update(ctx context.Context, ids []int, method, payload string, createTime int64, state string) error {
	return dac.GetRW().UpdateRequestsInfo(ctx, ids, method, payload, createTime, state)
}

// BatchReExecute 批量重新执行异步请求
func BatchReExecute(ctx context.Context, ids []int, createTime int64, state string) error {
	return dac.GetRW().BatchReExecuteRequestsInfo(ctx, ids, createTime, state)
}

// Outdate 将异步请求标记为过期
func Outdate(ctx context.Context, ids []int) error {
	return dac.GetRW().OutdateRequests(ctx, ids)
}

// GetAll 获取模组下所有异步请求
func GetAll(ctx context.Context, mozuID string) ([]db.Request, error) {
	return dac.GetRW().GetAllRequests(ctx, mozuID)
}

// convertToAsyncInfos 将数据库请求记录转换为前端展示的异步消息信息
// 封装了"收集控制器ID → 获取名称映射 → 构建AsynchronousInfo切片"的通用逻辑
func convertToAsyncInfos(ctx context.Context, reqs []db.Request) ([]cgi.AsynchronousInfo, error) {
	controllerIdSet := make(map[db.IDType]struct{})
	for i := range reqs {
		controllerIdSet[reqs[i].ControllerID] = struct{}{}
	}
	controllerIds := make([]db.IDType, 0, len(controllerIdSet))
	for id := range controllerIdSet {
		controllerIds = append(controllerIds, id)
	}

	controllerMap, err := dac.GetRW().GetControllerNames(ctx, controllerIds)
	if err != nil {
		return nil, err
	}

	result := make([]cgi.AsynchronousInfo, 0, len(reqs))
	for i := range reqs {
		req := &reqs[i]
		result = append(result, cgi.AsynchronousInfo{
			ID:             req.ID,
			Method:         req.Method,
			Payload:        string(req.Payload),
			ControllerName: controllerMap[req.ControllerID],
			CreateTime:     time.UnixMilli(req.CreateTime).Format(layout),
			AccessTime:     time.UnixMilli(req.AccessTime).Format(layout),
			State:          req.State,
		})
	}
	return result, nil
}

// GetInfo 获取所有异步消息，与前端页面适配
func GetInfo(ctx context.Context, mozuID string) ([]cgi.AsynchronousInfo, error) {
	reqs, err := GetAll(ctx, mozuID)
	if err != nil {
		return nil, err
	}
	return convertToAsyncInfos(ctx, reqs)
}

// GetRequests 分页查询异步请求，支持多条件过滤
func GetRequests(ctx context.Context, mozuID string,
	offset int, limit int, query string,
	beginTime int64, endTime int64, queryCreateTime bool,
	state string, queryState bool,
	method string, queryMethod bool,
) (cgi.Requests, error) {
	// 1. 获取数据库目标requests
	total, dbRequests, err := dac.GetRW().GetRequests(ctx, mozuID, offset, limit, query, beginTime,
		endTime, queryCreateTime, state, queryState, method, queryMethod)
	if err != nil {
		return cgi.Requests{}, fmt.Errorf("GetRequests error: %w", err)
	}

	// 2. 获取门禁控制器名称并构建cgi对象
	requestInfos, err := convertToAsyncInfos(ctx, dbRequests)
	if err != nil {
		return cgi.Requests{}, fmt.Errorf("get controller names error: %w", err)
	}

	return cgi.Requests{
		Total: total,
		List:  requestInfos,
	}, nil
}

// GetAllRequestWithControllerInfo 分页查询异步请求并关联控制器信息
func GetAllRequestWithControllerInfo(ctx context.Context,
	mozuID string, offset int, limit int,
	query string, method string,
) (int64, []cgi.RequestInfo, error) {
	total, requestDB, controllerInfoMap, err := dac.GetRW().
		GetAllRequestWithControllerInfo(ctx, mozuID, offset, limit, query, method)
	if err != nil {
		return 0, nil, err
	}

	requestInfo := make([]cgi.RequestInfo, 0)
	for i := range requestDB {
		r := &requestDB[i]
		requestInfo = append(requestInfo, cgi.RequestInfo{
			ID:             r.ID,
			ControllerID:   r.ControllerID,
			Method:         r.Method,
			Message:        r.Message,
			CreateTime:     r.CreateTime,
			AccessTime:     r.AccessTime,
			MozuID:         r.MozuID,
			ControllerName: controllerInfoMap[r.ControllerID].ControllerName,
			ControllerIP:   controllerInfoMap[r.ControllerID].ControllerIP,
		})
	}

	return total, requestInfo, err
}

// GetMethods 获取模组下所有异步请求的方法类型列表
func GetMethods(ctx context.Context, mozuID string) ([]string, error) {
	requests, err := dac.GetRW().GetAllRequests(ctx, mozuID)
	if err != nil {
		return nil, err
	}

	requestMethodMap := make(map[string]struct{})
	var methods []string
	for i := range requests {
		r := &requests[i]
		if _, ok := requestMethodMap[r.Method]; !ok {
			methods = append(methods, r.Method)
			requestMethodMap[r.Method] = struct{}{}
		}
	}
	requestMethods := make([]string, len(methods))
	for i := range methods {
		requestMethods[i] = methods[i]
	}
	return requestMethods, err
}

// writeExcel 将异步消息写入Excel文件
func writeExcel(requests []cgi.AsynchronousInfo) (*xlsx.File, error) {
	f := xlsx.NewFile()
	s, err := f.AddSheet("异步消息")
	if err != nil {
		return nil, err
	}

	if _, err = excel.AddStringRow(s, ht, titles...); err != nil {
		return nil, err
	}

	for i := range requests {
		r := &requests[i]
		if _, err = excel.AddRow(s, ht, r.ID, r.Method, r.Payload, r.CreateTime,
			r.ControllerName, r.State, r.AccessTime); err != nil {
			return nil, err
		}
	}

	return f, nil
}

// Export 按ID列表导出异步消息到Excel
func Export(ctx context.Context, mozuID string, requestIds []db.IDType) (*xlsx.File, error) {
	// 1. 获取指定id和模组id的request记录
	reqs, err := dac.GetRW().GetRequestsByIds(ctx, mozuID, requestIds)
	if err != nil {
		return nil, err
	}

	// 2. 获取控制器名称并封装异步消息结果
	requests, err := convertToAsyncInfos(ctx, reqs)
	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		config.Log.Warnf("根据id%s查询模组%v下的异步消息记录为空。", requestIds, mozuID)
		return nil, err
	}
	return writeExcel(requests)
}

// ExportAll 导出模组下所有异步消息到Excel
func ExportAll(ctx context.Context, mozuID string) (*xlsx.File, error) {
	requests, err := GetInfo(ctx, mozuID)
	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		config.Log.Warnf("查询模组%v下的异步消息记录为空。", mozuID)
		return nil, err
	}
	return writeExcel(requests)
}
