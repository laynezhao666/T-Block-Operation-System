package cgi

import (
	"context"
	"fmt"
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/cm"
	"agent/logic/collector/dispatcher"
	"agent/logic/collector/rtdb"
	"agent/logic/collector/rtdb/model"
	"sort"

	pb "trpcprotocol/agent"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

const (
	commstePointSuffix = "commste"
)

// DevicesHandle 设备列表
func DevicesHandle(ctx context.Context) (*pb.RspDevices, error) {
	s := dispatcher.Dispatcher().GetStatus()
	status := make(map[string]*pb.RunningStatus)
	for k, v := range s {
		d := make([]string, len(v.Devices))
		for i, x := range v.Devices {
			d[i] = string(x)
		}
		val := &pb.RunningStatus{
			Devices:   d,
			IsRunning: v.IsRunning,
		}
		status[k] = val
	}
	return &pb.RspDevices{
		Status: status,
	}, nil
}

// DevicesCommsteHandle 设备列表
func DevicesCommsteHandle(ctx context.Context, req *emptypb.Empty) (*pb.DevicesCommsteRsp, error) {
	devices := cm.Worker().GetAllDevices()
	rsp := &pb.DevicesCommsteRsp{
		Devices: make([]*pb.DeviceWithCommste, 0, len(devices)),
	}
	dataPoints := make(model.DataPoints, 0, len(devices))

	for _, d := range devices {
		commstePointId := d.Gid + consts.DefaultIDSep + commstePointSuffix
		dataPoints = append(dataPoints, model.DataPoint{
			ID: definition.DataPointIDType(commstePointId),
		})
		device := &pb.DeviceWithCommste{
			Id:   string(d.ID),
			Gid:  string(d.Gid),
			Name: d.Name,
			Type: d.TypeEn,
			Tpl: &pb.Template{
				Tplnm:   d.TemplateData.TemplateName,
				Tplpath: d.TemplateData.TemplatePath,
			},
			Channel: &pb.Channel{
				Addr:         d.ChData.Address,
				Chid:         d.ChData.ChannelID,
				Chparams:     d.ChData.ChannelParams,
				Chtype:       d.ChData.Chtype,
				Timeout:      fmt.Sprintf("%v", d.ChData.TimeoutMs),
				CmdIntetrval: fmt.Sprintf("%v", d.ChData.CmdInterval),
				WaitTime:     fmt.Sprintf("%v", d.ChData.WaitTimeMs),
			},
			MozuId: int32(d.MozuID),
		}
		rsp.Devices = append(rsp.Devices, device)
	}
	rtdb.GetDataPoints(dataPoints)
	for i := range dataPoints {
		val := &dataPoints[i].Rtd.Val
		pv := val.Pv.String()
		rsp.Devices[i].Commste = &pb.RtdPoint{
			Id:      string(dataPoints[i].ID),
			Pv:      pv,
			Tms:     fmt.Sprintf("%v", val.Tms),
			Des:     val.Desc,
			Qua:     fmt.Sprintf("%v", val.Qua),
			Alm:     "-1", //fmt.Sprintf("%v", point.Rtd.Alarm.Level),
			Virtual: dataPoints[i].Rtd.Virtual,
		}
	}
	sort.Slice(rsp.Devices, func(i, j int) bool {
		return rsp.Devices[i].Id < rsp.Devices[j].Id
	})
	return rsp, nil
}
