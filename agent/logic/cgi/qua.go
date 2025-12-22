package cgi

import (
	"context"
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/cm"
	"agent/logic/collector/rtdb"
	pb "trpcprotocol/agent"
)

// QuaHandle Qua处理
func QuaHandle(ctx context.Context, req *pb.ReqQua) (*pb.RspQua, error) {
	if len(req.Qua) == 0 {
		return nil, nil
	}

	quaMap := make(map[int32]*pb.DataPointIDs, 10)
	for _, qua := range req.Qua {
		list := make([]string, 0)
		quaMap[qua] = &pb.DataPointIDs{DataPointId: list}
	}

	// qua异常点位：全量拉取后筛选
	var pointsID definition.DataPointIDsType
	if req.PointType == definition.StdDevice {
		stdData := cm.Worker().GetStdData()
		if stdData == nil || len(stdData.StdPointsInfo) == 0 {
			return nil, nil
		}
		pointsID = make(definition.DataPointIDsType, 0, len(stdData.StdPointsInfo))
		for _, v := range stdData.StdPointsInfo {
			point := definition.DataPointIDType(v.StdDevice + consts.DefaultIDSep + v.StdPoint)
			pointsID = append(pointsID, point)
		}
	} else {
		collectData := cm.Worker().GetCollectData()
		if collectData == nil || len(collectData) == 0 {
			return nil, nil
		}
		pointsID = make(definition.DataPointIDsType, 0, len(collectData))
		for _, v := range collectData {
			pointsID = append(pointsID, v.ID)
		}
	}

	allPoints := rtdb.GetDataPointsByID(pointsID)

	for _, point := range allPoints {
		quaInt := int32(point.Rtd.Val.Qua)
		if v, ok := quaMap[quaInt]; ok {
			v.DataPointId = append(v.DataPointId, string(point.ID))
		}
	}

	return &pb.RspQua{
		Data: quaMap,
	}, nil
}
