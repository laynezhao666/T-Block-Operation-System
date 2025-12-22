package cgi

import (
	"context"
	"fmt"
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/entity/errcode"
	"agent/logic/cm"
	"agent/logic/collector/rtdb"
	"agent/logic/collector/rtdb/model"
	"agent/logic/distribution/distributor"
	"agent/utils"
	"agent/utils/osal"
	"sort"
	"strconv"
	"time"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go/log"

	pb "trpcprotocol/agent"
)

func getVirtualPoints(deviceGid string) model.DataPoints {
	lists := []string{
		// "almcount",
		// "almste",
		"commste",
		"point_throughput",
		"range_resp_time",
		"total_resp_time",
		"avg_resp_time",
		"max_resp_time",
		"min_resp_time",
		"success_req_in_period",
		"total_req_in_period",
		"success_req",
		"minute_success_req",
		"total_req",
		"minute_total_req",
		"interruption",
		"timeout_req",
		// "tms_delay_count",
		// "qua_err_count",
		// "qua_origin_err_count",
		"success_msg_req",
		"total_msg_req",
	}

	dataPoints := make(model.DataPoints, 0, len(lists))
	for _, v := range lists {
		point := deviceGid + consts.DefaultIDSep + v
		dataPoints = append(dataPoints, model.DataPoint{
			ID: definition.DataPointIDType(point),
		},
		)
	}
	return dataPoints
}

// GetRtdHandle 获取测点数据
func GetRtdHandle(ctx context.Context, req *pb.ReqRtd) (*pb.RspRtd, error) {
	dataPoints := make(model.DataPoints, 0, (1+len(req.Devices))*len(req.Points))

	if req.PointType == definition.AllPoints {
		dataPoints = rtdb.GetAll()
	} else {
		for _, point := range req.Points {
			dataPoints = append(dataPoints, model.DataPoint{
				ID: definition.DataPointIDType(point),
			},
			)
		}
		for _, device := range req.Devices {
			deviceDataPoints := distributor.PointsDataManager().GetDataPoints(definition.DeviceGidType(device))
			dataPoints = append(dataPoints, deviceDataPoints...)

			// add virtual points
			virtualPoints := getVirtualPoints(device)
			dataPoints = append(dataPoints, virtualPoints...)
		}

		rtdb.GetDataPoints(dataPoints)
	}

	points := make([]*pb.RtdPoint, 0, len(dataPoints))
	for _, point := range dataPoints {
		val := &point.Rtd.Val
		points = append(
			points, &pb.RtdPoint{
				Id:      string(point.ID),
				Pv:      val.Pv.String(),
				Tms:     fmt.Sprintf("%v", val.Tms),
				Des:     val.Desc,
				Qua:     fmt.Sprintf("%v", val.Qua),
				Alm:     "-1", //fmt.Sprintf("%v", point.Rtd.Alarm.Level),
				Virtual: point.Rtd.Virtual,
			},
		)
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Id < points[j].Id
	})
	return &pb.RspRtd{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
		Data:    points,
	}, nil
}

// GetRtdByIdHandle 获取测点数据
func GetRtdByIdHandle(ctx context.Context, req *pb.ReqRtd) (*pb.RspRtd, error) {
	dataPoints := make(model.DataPoints, 0, (1+len(req.Devices))*len(req.Points))

	if req.PointType == definition.AllPoints {
		dataPoints = rtdb.GetAll()
	} else {
		for _, id := range req.Points {
			gidPointId, err := convertToGidPoint(definition.DataPointIDType(id), req.PointType)
			if err != nil {
				log.Infof("point not exist:%v", id)
				continue
			}
			dataPoints = append(dataPoints, model.DataPoint{
				ID: gidPointId,
			})
		}
		for _, deviceId := range req.Devices {
			deviceGid, ok := convertToDeviceGid(deviceId, req.PointType)
			if !ok {
				log.Infof("device not exist:%v", deviceId)
				continue
			}
			deviceDataPoints := distributor.PointsDataManager().GetDataPoints(deviceGid)
			dataPoints = append(dataPoints, deviceDataPoints...)

			// add virtual points
			virtualPoints := getVirtualPoints(deviceId)
			dataPoints = append(dataPoints, virtualPoints...)
		}
		rtdb.GetDataPoints(dataPoints)
	}

	points := make([]*pb.RtdPoint, 0, len(dataPoints))
	for _, point := range dataPoints {
		val := &point.Rtd.Val
		idPointId, err := convertToIdPoint(point.ID, req.PointType)
		if err != nil {
			continue
		}
		points = append(
			points, &pb.RtdPoint{
				Id:      string(idPointId),
				Pv:      val.Pv.String(),
				Tms:     fmt.Sprintf("%v", val.Tms),
				Des:     val.Desc,
				Qua:     fmt.Sprintf("%v", val.Qua),
				Alm:     "-1", //fmt.Sprintf("%v", point.Rtd.Alarm.Level),
				Virtual: point.Rtd.Virtual,
			},
		)
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Id < points[j].Id
	})
	return &pb.RspRtd{
		Code:    errcode.DefaultCgiRspCode,
		Message: errcode.DefaultCgiRspMessage,
		Data:    points,
	}, nil
}

func convertToGidPoint(oldId definition.DataPointIDType, pointType string) (definition.DataPointIDType, error) {
	deviceId, pointId, err := definition.SplitDataPointID(definition.DataPointIDType(oldId))
	if err != nil {
		return "", fmt.Errorf("invalid point id [%v], split failed", oldId)
	}

	if pointType == definition.StdDevice {
		deviceGid, ok := cm.Worker().GetStdDeviceGidById(string(deviceId))
		if !ok {
			return "", fmt.Errorf("std device gid not found")
		}
		return definition.GenerateDataPointID(deviceGid, pointId), nil
	} else {
		deviceGid, ok := cm.Worker().GetDeviceGidById(string(deviceId))
		if !ok {
			return "", fmt.Errorf("device gid not found")
		}
		return definition.GenerateDataPointID(deviceGid, pointId), nil
	}
}

func convertToIdPoint(oldId definition.DataPointIDType, pointType string) (definition.DataPointIDType, error) {
	deviceGid, pointId, err := definition.SplitDataPointID(definition.DataPointIDType(oldId))
	if err != nil {
		return "", fmt.Errorf("invalid point id [%v], split failed", oldId)
	}

	if pointType == definition.StdDevice {
		stdDevice, ok := cm.Worker().GetStdDeviceByGid(deviceGid)
		if !ok {
			return "", fmt.Errorf("device id not found")
		}
		return definition.GenerateDataPointID(stdDevice.ConciseCode, pointId), nil
	} else {
		device, ok := cm.Worker().GetDeviceByGid(deviceGid)
		if !ok {
			return "", fmt.Errorf("device id not found")
		}
		return definition.GenerateDataPointID(device.ID, pointId), nil
	}
}

func convertToDeviceGid(deviceId string, pointType string) (definition.DeviceGidType, bool) {
	if pointType == definition.StdDevice {
		return cm.Worker().GetStdDeviceGidById(deviceId)
	} else {
		return cm.Worker().GetDeviceGidById(deviceId)
	}
}

// SetRtdByIdHandle 设置测点数据
func SetRtdByIdHandle(ctx context.Context, req *pb.SetRtdByIdReq) (*emptypb.Empty, error) {
	points := req.GetPoints()
	dataPoints := make(model.DataPoints, 0, len(points))
	nowTms := time.Now().UnixMilli()
	for _, p := range req.Points {
		val, err := strconv.ParseFloat(p.Pv, 64)
		if err != nil {
			log.DebugContextf(ctx, "parse point [%+v] pv failed, continue", p)
			continue
		}
		deviceId, pointId, err := definition.SplitDataPointID(definition.DataPointIDType(p.Id))
		if err != nil {
			log.DebugContextf(ctx, "invalid point id [%v], split failed, continue", p.Id)
			continue
		}
		deviceGid, exist := cm.Worker().GetDeviceGidById(string(deviceId))
		if !exist {
			log.DebugContextf(ctx, "cannot find gid of device id [%v], continue", deviceId)
			continue
		}
		gidPointId := definition.GenerateDataPointID(deviceGid, pointId)
		q, err := strconv.ParseInt(p.Qua, 10, 64)
		var qua consts.Quality
		if err != nil {
			log.DebugContextf(ctx, "parse point [%+v] qua failed, use ok qua", p)
			qua = consts.QualityOk
		} else {
			qua = consts.Quality(q)
		}

		var tms int64
		tms, err = strconv.ParseInt(p.Tms, 10, 64)
		if err != nil {
			log.DebugContextf(ctx, "parse point [%+v] tms failed, use now tms", p)
			tms = nowTms
		}
		ts := utils.ConvertIfMilliToSeconds(tms)
		point := model.DataPoint{
			ID:        gidPointId,
			DeviceGiD: deviceGid,
			Rtd: model.RTData{
				Val: model.RTValue{
					Pv:   osal.NewVariantWithValue(val),
					Qua:  qua,
					Tms:  ts,
					Desc: p.Des,
				},
				Virtual: false,
			},
			IsValueChanged: false,
			PointType:      definition.CollectPointType,
		}
		dataPoints = append(dataPoints, point)
	}
	rtdb.SetDataPoints(dataPoints, true)
	return &emptypb.Empty{}, nil
}
