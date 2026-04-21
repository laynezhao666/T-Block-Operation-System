package dac

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/db"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
)

// ControllerInfo 门禁控制器基本信息（IP和名称），用于请求列表展示。
type ControllerInfo struct {
	ControllerIP   string
	ControllerName string
}

// deleteRequests 根据控制器ID列表删除关联的请求记录。
func deleteRequests(tx *gorm.DB, controllerIDs []db.IDType) error {
	if len(controllerIDs) == 0 {
		return nil
	}

	return withControllerIDs(tx, controllerIDs).Delete(&db.Request{}).Error
}

// AddRequests 批量添加请求记录（自动设置创建时间和访问时间）。
func AddRequests(tx *gorm.DB, reqs []db.Request) error {
	if len(reqs) == 0 {
		return nil
	}

	t := time.Now().UTC()
	for i := range reqs {
		r := &reqs[i]
		r.CreateTime = t.UnixMilli()
		r.AccessTime = r.CreateTime
	}

	if config.C.Debug {
		for i := range reqs {
			config.Log.Infof("req: %+v, payload: %+v", reqs[i], string(reqs[i].Payload))
		}
	}

	return tx.Create(reqs).Error
}

// GetAllRequestWithControllerInfo 分页获取请求记录并附带控制器信息（支持按通道ID和名称模糊搜索）。
func (d *impl) GetAllRequestWithControllerInfo(ctx context.Context, mozuID string, offset int, limit int,
	query string, method string) (int64, []db.Request, map[db.IDType]ControllerInfo, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, nil, errors.New(fmt.Sprintf("GetAllRequestWithControllerInfo not support by offset:%v, limit:%v",
			offset, limit))
	}

	var (
		totalCount        int64
		requests          []db.Request
		controllerInfoMap = make(map[db.IDType]ControllerInfo)
		e                 error
		opts              = make([]tgorm.Option, 0, 10)
	)

	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	orOpts := []tgorm.Option{withJSONLike(db.ColumnChannel, []string{db.ColumnChannelID}, query), withNameLike(query)}

	e = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var controllers []db.DoorController
		opts = append(opts, tgorm.WithOr(tx, orOpts...))
		if err := tgorm.WithOptions(tx, opts...).Find(&controllers).Error; err != nil {
			return err
		}

		var controllerIDs []db.IDType
		for i := range controllers {
			c := &controllers[i]
			controllerInfoMap[c.ID] = ControllerInfo{
				ControllerIP:   c.Channel.ID,
				ControllerName: c.Name,
			}
			controllerIDs = append(controllerIDs, c.ID)
		}

		return getRequestsByControllerIDs(tx, offset, limit, method, controllerIDs, &requests, &totalCount)
	})

	return totalCount, requests, controllerInfoMap, e
}

// GetRequests 分页查询请求记录（支持时间范围、状态、方法和模糊搜索等多条件过滤）。
func (d *impl) GetRequests(ctx context.Context, mozuID string, offset int, limit int, query string,
	beginTime int64, endTime int64, queryCreateTime bool, state string, queryState bool, method string,
	queryMethod bool) (int64, []db.Request, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, fmt.Errorf("GetRequests not support by offset: %v, limit: %v", offset, limit)
	}

	var (
		totalCount int64
		requests   []db.Request
		opts       = make([]tgorm.Option, 0, 10)
		err        error
	)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	orOpts := make([]tgorm.Option, 0, 2)
	// 这里查询时间的范围为前闭后开，因此让end_time+1以适配查询边界点的情况
	if queryCreateTime {
		opts = append(opts, withCreateTimeBetweenOption(beginTime, endTime+1))
	}
	if queryState {
		opts = append(opts, withRequestState(state))
	}
	if queryMethod {
		opts = append(opts, withMethod(method))
	}

	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var controllers []db.DoorController
		if len(query) > 0 {
			if e := tgorm.WithOptions(tx, withNameLike(query)).Find(&controllers).Error; e != nil {
				return e
			}
			controllerIds := make([]db.IDType, 0, len(controllers))
			for i := range controllers {
				controllerIds = append(controllerIds, controllers[i].ID)
			}
			orOpts = append(orOpts, withControllerIDsOption(controllerIds))

			opts = append(opts, tgorm.WithOr(tx, orOpts...))
		}

		return queryAndCountRecords(
			tgorm.WithOptions(tx.Model(&db.Request{}), opts...),
			offset, limit, &requests, &totalCount,
		)
	})

	return totalCount, requests, err

}

// GetAllRequests 获取指定模组下所有请求记录。
func (d *impl) GetAllRequests(ctx context.Context, mozuID string) ([]db.Request, error) {
	var requests []db.Request
	opts := make([]tgorm.Option, 0, 1)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	err := tgorm.WithOptions(d.db.WithContext(ctx), opts...).Find(&requests).Error
	return requests, err
}

// CleanRequests 清理请求记录（通过回调函数决定保留哪些）。
func CleanRequests(tx *gorm.DB, mozuID string,
	afterGet func(*gorm.DB, []db.Request) ([]db.Request, error),
) ([]db.Request, error) {
	opts := make([]tgorm.Option, 0, 1)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)

	requests := make([]db.Request, 0)
	err := tgorm.WithOptions(tx, opts...).Find(&requests).Error

	if err != nil {
		return requests, err
	}

	if afterGet != nil {
		if requests, err = afterGet(tx, requests); err != nil {
			return requests, err
		}
	}

	return requests, nil
}

// CleanRequests 清理请求记录（impl 方法，在事务中执行）。
func (d *impl) CleanRequests(ctx context.Context,
	afterGet func(*gorm.DB, []db.Request) ([]db.Request, error),
) ([]db.Request, error) {
	var (
		requests []db.Request
		e        error
	)
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		requests, e = CleanRequests(tx, "", afterGet)
		return e
	})
	return requests, err
}

// GetRequestsByIds 根据ID列表获取请求记录。
func (d *impl) GetRequestsByIds(ctx context.Context, mozuID string, requestIds []db.IDType) ([]db.Request, error) {
	var requests []db.Request
	opts := make([]tgorm.Option, 0, 2)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	opts = addIDsOptionIfNotEmpty(opts, requestIds)
	err := tgorm.WithOptions(d.db.WithContext(ctx), opts...).Find(&requests).Error
	return requests, err
}

// GetRequestsByControllers 获取指定控制器的所有请求记录。
func (d *impl) GetRequestsByControllers(ctx context.Context, controllerID db.IDType) ([]db.Request, error) {
	var requests []db.Request
	err := tgorm.WithOptions(d.db.WithContext(ctx), withControllerIDOption(controllerID)).Find(&requests).Error
	return requests, err
}

// AddRequests 批量添加请求记录（impl 方法）。
func (d *impl) AddRequests(ctx context.Context, reqs []db.Request) error {
	return AddRequests(d.db.WithContext(ctx), reqs)
}

// DeleteRequests 批量删除请求记录。
func (d *impl) DeleteRequests(ctx context.Context, ids []db.IDType) error {
	if len(ids) == 0 {
		return nil
	}

	return deleteRecordsByID(d.db.WithContext(ctx), ids, &db.Request{})
}

// FetchRequests 从数据库中获取一定数量的待执行的请求
func (d *impl) FetchRequests(ctx context.Context, n int) ([]db.Request, error) {
	reqs := make([]db.Request, 0)
	query := d.db.WithContext(ctx).
		Where("state = ?", consts.StateToBeExecuted).Limit(n)
	err := withASC(query, "access_time").Find(&reqs).Error
	return reqs, err
}

// UpdateFailedRequests 更新失败请求执行的结果message
func (d *impl) UpdateFailedRequests(ctx context.Context, failedIds []db.IDType, messages map[db.IDType]string) error {
	if len(failedIds) == 0 {
		return nil
	}

	t := time.Now().UTC().UnixMilli()
	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range failedIds {
			id := failedIds[i]
			if err = withID(tx.Model(&db.Request{}), id).Updates(map[string]interface{}{
				"access_time": t,
				"message":     messages[id],
				"state":       consts.StateFailed,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateSuccessRequests 更新成功请求执行的状态
func (d *impl) UpdateSuccessRequests(ctx context.Context, successIds []db.IDType) error {
	if len(successIds) == 0 {
		return nil
	}

	t := time.Now().UTC().UnixMilli()
	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range successIds {
			id := successIds[i]
			if err = withID(tx.Model(&db.Request{}), id).Updates(map[string]interface{}{
				"access_time": t,
				"state":       consts.StateSuccess,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// OutdateRequests 使请求过期
func (d *impl) OutdateRequests(ctx context.Context, ids []db.IDType) error {
	if len(ids) == 0 {
		return nil
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tgorm.WithOptions(tx.Model(&db.Request{}), withIDsOption(ids)).Updates(map[string]interface{}{
			"state": consts.StateExpired,
		}).Error
	})
}

// DeleteRequestsByTime 根据时间阈值删除请求记录
func (d *impl) DeleteRequestsByTime(ctx context.Context, timeThreshold int64) (int64, error) {
	result := d.db.WithContext(ctx).Where("create_time <= ?", timeThreshold).Delete(&db.Request{})
	return result.RowsAffected, result.Error
}

// OutdatedRequestsByTime 根据时间阈值将指定状态的请求标记为过期
func (d *impl) OutdatedRequestsByTime(ctx context.Context, timeThreshold int64, currentState string) (int64, error) {
	result := d.db.WithContext(ctx).Model(&db.Request{}).
		Where("create_time <= ? AND state = ?", timeThreshold, currentState).
		Update("state", consts.StateExpired)
	return result.RowsAffected, result.Error
}

// getRequestsByControllerIDs 根据控制器ID列表分页查询请求记录。
func getRequestsByControllerIDs(tx *gorm.DB, offset, limit int,
	method string, controllerIDs []db.IDType,
	requests *[]db.Request, count *int64,
) error {
	opts := make([]tgorm.Option, 0, 2)
	opts = append(opts, withControllerIDsOption(controllerIDs))
	if method != "" {
		opts = append(opts, withMethod(method))
	}
	return tgorm.WithOptions(tx.Model(&db.Request{}), opts...).
		Offset(offset).Limit(limit).Find(requests).
		Offset(-1).Limit(-1).Count(count).Error
}

// UpdateRequestsInfo 更新请求的方法、载荷、创建时间和状态。
func (d *impl) UpdateRequestsInfo(ctx context.Context, ids []db.IDType,
	method, payload string, createTime int64, state string,
) error {
	if len(ids) == 0 {
		return nil
	}

	req := db.Request{
		Method:     method,
		Payload:    []byte(payload),
		CreateTime: createTime,
		State:      state,
	}
	if len(payload) == 0 {
		req.Payload = nil
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tgorm.WithOptions(tx.Model(&db.Request{}), withIDsOption(ids)).Updates(req).Error
	})
}

// BatchReExecuteRequestsInfo 批量重新执行请求（更新创建时间和状态）。
func (d *impl) BatchReExecuteRequestsInfo(ctx context.Context, ids []db.IDType, createTime int64, state string) error {
	if len(ids) == 0 {
		return nil
	}

	updateFields := map[string]interface{}{
		"create_time": createTime,
		"state":       state,
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tgorm.WithOptions(tx.Model(&db.Request{}), withIDsOption(ids)).Updates(updateFields).Error
	})
}
