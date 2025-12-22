// Package rpc  定义数据查询相关接口调用逻辑
package rpc

import (
	"context"
	"data-compute/entity/model"
	"fmt"
	"strings"
	"sync"
	"time"
	pb "trpcprotocol/data-cache"

	"github.com/avast/retry-go"
	"github.com/samber/lo"
)

// ICacheQueryApi 数据缓存查询接口
type ICacheQueryApi interface {
	// ReadLatest 读取最新数据
	ReadLatest(ctx context.Context, pointNames []string, max int64) (map[string]*model.Point, error)
	// ReadChanged 读取变化测点
	ReadChanged(ctx context.Context, pointNames []string, max int64) (map[string]int64, error)
}

// NewCacheQueryApi 创建一个数据缓存查询接口实现类
func NewCacheQueryApi() ICacheQueryApi {
	return cacheApiImpl{
		cli: pb.NewPointClientProxy(),
	}
}

type cacheApiImpl struct {
	cli pb.PointClientProxy
}

// ReadRange 范围读取数据
func (s cacheApiImpl) ReadRange(ctx context.Context, pointNames []string, begin, end int64) (map[string][]*model.Point, error) {
	chunkPoints := lo.Chunk(pointNames, 5000)
	chunkResult := make([]*pb.RspQueryData, len(chunkPoints))
	errs := make([]string, 0, len(chunkPoints))
	wg := sync.WaitGroup{}
	// 分批并发请求
	for idx, chunkPoints := range chunkPoints {
		req := &pb.ReqQueryData{
			PointList:   chunkPoints,
			Begin:       begin,
			End:         end,
			QualityType: pb.QualityType_QUALITY_TYPE_ALL,
		}
		wg.Add(1)
		go func(i int, reqItem *pb.ReqQueryData) {
			defer wg.Done()
			// 重试读取
			err := retry.Do(func() (err error) {
				chunkResult[i], err = s.cli.QueryData(ctx, reqItem)
				return err
			}, retry.Attempts(3), retry.Delay(time.Millisecond*100), retry.RetryIf(func(err error) bool {
				return err != nil
			}))
			if err != nil {
				errs[i] = fmt.Sprintf("batch[%d] query cahce api for data points fail, err:%s", i, err.Error())
				return
			}
		}(idx, req)
	}
	wg.Wait()
	allErrs := lo.Filter(errs, func(item string, index int) bool {
		return len(item) > 0
	})
	if len(allErrs) > 0 {
		return nil, fmt.Errorf("query error: [%s]", strings.Join(allErrs, "\n"))
	}
	// 合并结果
	result := make(map[string][]*model.Point)
	for _, res := range chunkResult {
		for k, v := range res.PointMap {
			newV := lo.Map(v.Values, func(item *pb.RspQueryData_PointValueList_PointValue, index int) *model.Point {
				return &model.Point{
					Name:    k,
					Quality: int32(item.Q),
					Value:   item.V,
					Time:    item.T,
				}
			})
			result[k] = newV
		}
	}
	return result, nil
}

// ReadLatest 读取最新数据
func (s cacheApiImpl) ReadLatest(ctx context.Context, pointNames []string, max int64) (map[string]*model.Point, error) {
	pointData, err := s.ReadRange(ctx, pointNames, 0, max)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*model.Point)
	for k, v := range pointData {
		if len(v) > 0 {
			res[k] = v[0]
		}
	}
	return res, err
}

// ReadChanged 读取变化测点数据
func (s cacheApiImpl) ReadChanged(ctx context.Context, pointNames []string, max int64) (map[string]int64, error) {
	chunkPoints := lo.Chunk(pointNames, 5000)
	chunkResult := make([]*pb.RspQueryChanged, len(chunkPoints))
	errs := make([]string, 0)
	wg := sync.WaitGroup{}
	// 分批并发请求
	for idx, chunkPoints := range chunkPoints {
		req := &pb.ReqQueryChanged{
			PointList: chunkPoints,
			Begin:     max,
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// 重试读取
			err := retry.Do(func() (err error) {
				chunkResult[i], err = s.cli.QueryChanged(ctx, req)
				return err
			}, retry.Attempts(3), retry.Delay(time.Millisecond*100), retry.RetryIf(func(err error) bool {
				return err != nil
			}))
			if err != nil {
				errs = append(errs, fmt.Sprintf("batch[%d] query cahce api for changed points fail, err:%s", i, err.Error()))
				return
			}
		}(idx)
	}
	wg.Wait()
	if len(errs) > 0 {
		return nil, fmt.Errorf("query error: [%s]", strings.Join(errs, "\n"))
	}
	// 合并结果
	result := make(map[string]int64)
	for _, res := range chunkResult {
		for k, v := range res.ChangedMap {
			result[k] = v
		}
	}
	return result, nil
}
