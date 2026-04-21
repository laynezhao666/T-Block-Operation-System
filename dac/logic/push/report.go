// Package push 提供门禁测点数据的上报推送功能。
package push

import (
	"context"
	"dac/logic/mapping"
	"sync"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/repo/data"

	pb "dac/repo/pb/tcommon_point_data"
)

// reportControllerPoints 上报单个控制器的测点数据到TBOS
func (w *worker) reportControllerPoints(
	ctx context.Context, timestamp int64, kind pb.DataKind,
	controllerID db.IDType, points rt.Points,
) {
	c, ok := w.getController(controllerID)
	if !ok {
		w.notifyGetController(controllerID)
		return
	}

	mozu := c.MozuID
	code := c.GetCollectCode()
	deviceID, ok := mapping.GetWorker().GetGID(code)
	if !ok {
		mapping.GetWorker().Notify(code, mozu)
		config.Log.Warnf("can not find gid of controller %v", controllerID)
		return
	}

	pbPoints := make([]*pb.Point, 0, len(points))
	for i := range points {
		p := new(pb.Point)

		dp := &points[i]
		p.Id = dp.ID
		p.Value = dp.Rtd.Pv
		p.Timestamp = dp.Rtd.Timestamp
		p.Quality = int32(dp.Rtd.Qua)
		p.Kind = pb.PointKind_Collected

		pbPoints = append(pbPoints, p)
	}

	if config.C.ReportToTBOS {
		tbosCtx, cancel := context.WithCancel(ctx) // 派生上下文
		defer cancel()
		err := data.GetWriter().SetTBOSPointsWithExtends(
			tbosCtx, timestamp, kind,
			map[string]string{consts.PointMessageMozuKey: mozu},
			code, deviceID, pbPoints)
		if err != nil {
			config.Log.Warnf(
				"push tbos points %v of controller %v error: %v",
				points, deviceID, err)
		}
	}

}

// reportAllControllerPoints 并发上报所有控制器的测点数据
func (w *worker) reportAllControllerPoints(
	ctx context.Context, timestamp int64, kind pb.DataKind,
	controllerPoints map[db.IDType]rt.Points,
) {
	var wg sync.WaitGroup
	for controllerID, points := range controllerPoints {
		wg.Add(1)
		go func(controllerID db.IDType, points rt.Points) {
			defer wg.Done()

			w.reportControllerPoints(ctx, timestamp, kind, controllerID, points)
		}(controllerID, points)
	}

	wg.Wait()
}
