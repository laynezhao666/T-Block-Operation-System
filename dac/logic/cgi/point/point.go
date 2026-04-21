// Package point 提供门禁测点实时数据的查询接口。
package point

import (
	"context"
	"fmt"

	"dac/entity/model/rt"
	"dac/logic/collect/rtdb"
)

// CGIPoint 前端展示用的测点数据结构
type CGIPoint struct {
	ID        string `json:"id"`  // 测点ID
	Value     string `json:"pv"`  // 测点值
	Qua       string `json:"qua"` // 数据质量
	Timestamp string `json:"tms"` // 时间戳
}

// getPoints 批量获取测点实时数据
func getPoints(ctx context.Context, ids []string) ([]CGIPoint, error) {
	points := make(rt.Points, len(ids))
	for i, id := range ids {
		points[i].ID = id
	}
	err := rtdb.GetPoints(ctx, points)
	if err != nil {
		return nil, fmt.Errorf("get points error: %w", err)
	}

	results := make([]CGIPoint, len(points))
	for i := range points {
		results[i].ID = ids[i]

		p := &points[i].Rtd
		results[i].Value = p.Pv
		results[i].Qua = fmt.Sprintf("%v", p.Qua)
		results[i].Timestamp = fmt.Sprintf("%v", p.Timestamp)
	}

	return results, nil
}

// GetPoints 获取测点数据并以Map形式返回（ID为key）
func GetPoints(ctx context.Context, ids []string) (map[string]CGIPoint, error) {
	points, err := getPoints(ctx, ids)
	if err != nil {
		return nil, err
	}

	data := make(map[string]CGIPoint, len(points))
	for i := range points {
		p := &points[i]
		data[p.ID] = *p
	}
	return data, nil
}

// GetPointsList 获取测点数据并以列表形式返回
func GetPointsList(ctx context.Context, ids []string) ([]CGIPoint, error) {
	return getPoints(ctx, ids)
}
