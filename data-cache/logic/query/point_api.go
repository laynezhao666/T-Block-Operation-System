package query

import (
	"context"
	"data-cache/entity/model"
	"data-cache/repo/cache"
	pb "trpcprotocol/data-cache"

	"github.com/samber/lo"
)

// IPointApi 测点相关接口
type IPointApi interface {
	// QueryData 查询测点值数据
	QueryData(ctx context.Context, req *pb.ReqQueryData) (*pb.RspQueryData, error)
	// QueryChanged 查询变化测点数据
	QueryChanged(ctx context.Context, req *pb.ReqQueryChanged) (*pb.RspQueryChanged, error)
	// QueryCachedPoint 查询缓存测点数据
	QueryCachedPoint(ctx context.Context, req *pb.ReqQueryCachedPoint) (*pb.RspQueryCachedPoint, error)
}

// pointLogicImpl 测点逻辑接口实现struct
type pointLogicImpl struct {
}

// NewDefaultPointLogic 新建一个测点逻辑接口实现类
func NewDefaultPointLogic() IPointApi {
	return &pointLogicImpl{}
}

// QueryCachedPoint 查询本地缓存的点位列表
func (d *pointLogicImpl) QueryCachedPoint(ctx context.Context, req *pb.ReqQueryCachedPoint) (*pb.RspQueryCachedPoint, error) {
	// 获取所有的测点名称
	keys := cache.DataCache.Keys()
	// 过滤出用户查询的一批测点
	pMap := lo.SliceToMap(req.PointList, func(pointName string) (string, any) {
		return pointName, struct{}{}
	})
	keys = lo.Filter(keys, func(key string, index int) bool {
		if _, ok := pMap[key]; !ok {
			return true
		}
		return false
	})
	// 处理分页相关逻辑
	begin := (req.Page - 1) * req.Size
	end := begin + req.Size
	if end > int32(len(keys)) {
		end = int32(len(keys))
	}
	return &pb.RspQueryCachedPoint{
		PointList: keys[begin:end],
		Total:     int32(len(keys)),
	}, nil
}

func (d *pointLogicImpl) QueryData(ctx context.Context, req *pb.ReqQueryData) (*pb.RspQueryData, error) {
	rsp := &pb.RspQueryData{
		PointMap: make(map[string]*pb.RspQueryData_PointValueList),
	}
	// begin=0,读取最新的值
	if req.Begin == 0 {
		res, err := cache.ReadLatest(req.PointList, req.End)
		if err != nil {
			return nil, err
		}
		for pointName, point := range res {
			if point == nil {
				continue
			}
			if sameQuality(point, req.QualityType) {
				rsp.PointMap[pointName] = &pb.RspQueryData_PointValueList{
					Values: []*pb.RspQueryData_PointValueList_PointValue{{
						Q: int64(point.Quality),
						T: int64(point.Time),
						V: point.Value,
					}},
				}
			}
		}
		return rsp, nil
	}
	// 区间查询的情况
	res, err := cache.ReadRange(req.PointList, req.Begin, req.End)
	if err != nil {
		return nil, err
	}
	for pointName, points := range res {
		if len(points) == 0 {
			continue
		}
		values := make([]*pb.RspQueryData_PointValueList_PointValue, 0, len(points))
		for _, point := range points {
			if sameQuality(point, req.QualityType) {
				values = append(values, &pb.RspQueryData_PointValueList_PointValue{
					Q: int64(point.Quality),
					T: int64(point.Time),
					V: point.Value,
				})
			}
		}
		rsp.PointMap[pointName] = &pb.RspQueryData_PointValueList{
			Values: values,
		}
	}
	return rsp, nil
}

func sameQuality(point *model.StdPoint, qualityType pb.QualityType) bool {
	switch qualityType {
	case pb.QualityType_QUALITY_TYPE_ALL:
		return true
	case pb.QualityType_QUALITY_TYPE_BAD:
		return point.Quality < 0
	case pb.QualityType_QUALITY_TYPE_NORMAL:
		return point.Quality >= 0
	}
	return true
}

func (d *pointLogicImpl) QueryChanged(ctx context.Context, req *pb.ReqQueryChanged) (*pb.RspQueryChanged, error) {
	changedPoints, err := cache.ReadChanged(req.PointList, req.Begin)
	if err != nil {
		return nil, err
	}
	return &pb.RspQueryChanged{
		ChangedMap: changedPoints,
	}, nil
}
