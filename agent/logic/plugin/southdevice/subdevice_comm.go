package southdevice

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/rtdb"
	"agent/logic/collector/rtdb/model"
)

// setPointQuaErrByCommPoint 根据XX.Comm测点的值， 设置该子设备的测点质量为异常，或置为特殊值
// todo: 后续跟进设备实例的扩展属性配置做启用
func setPointQuaErrByCommPoint(deviceID definition.DeviceGidType, points model.DataPoints) {
	commErrSubDevices := getCommErrSubDevices(points)
	dealInputDataPoints(commErrSubDevices, points)
}

// getCommErrSubDevices 获取对应的Comm测点且值为1（1表示通讯中断）
func getCommErrSubDevices(points model.DataPoints) map[definition.DeviceGidType]struct{} {
	commAbnormalDevices := make(map[definition.DeviceGidType]struct{})

	// 获取这批测点对应的子设备
	subDeviceReq := make(map[definition.DeviceGidType]struct{})
	for i := range points {
		deviceGid, pointIdType, err := definition.SplitDataPointID(points[i].ID)
		if err != nil {
			continue
		}

		if pointIdType == definition.CommID {
			getCommDeviceCache().AddDevices(deviceGid)

			// 当 XX.Comm 的值 = 0 时跳过，否则赋值到 commAbnormalDevices
			if v, err := points[i].Rtd.Val.Pv.AsBool(); err != nil || !v {
				continue
			}
			commAbnormalDevices[deviceGid] = struct{}{}
		} else {
			// 如果子设备不存在直接采集的 Comm 测点，则跳过
			// 否则会与【根据子设备测点 -99998 赋值 Comm 测点为中断】的功能冲突
			if !getCommDeviceCache().HasCommePoint(deviceGid) {
				continue
			}

			// 封装向 RTDB 请求的子设备
			if _, ok := subDeviceReq[deviceGid]; !ok {
				subDeviceReq[deviceGid] = struct{}{}
			}
		}
	}

	// 获取子设备Comm测点实时值
	rtdPoints := make(model.DataPoints, 0, len(subDeviceReq))
	for k := range subDeviceReq {
		rtdPoints = append(rtdPoints, model.DataPoint{
			ID:        definition.GenerateDataPointID(k, definition.CommID),
			DeviceGiD: k,
		})
	}
	rtdb.GetDataPoints(rtdPoints)
	for i := range rtdPoints {
		p := &rtdPoints[i]
		if v, err := p.Rtd.Val.Pv.AsBool(); err != nil || !v { // 当测点 XX.Comm 值 = 0 时，跳过
			continue
		}
		commAbnormalDevices[p.DeviceGiD] = struct{}{}
	}

	return commAbnormalDevices
}

// dealInputDataPoints 将传入 points 中与 Comm 测点同设备 ID 的其它测点置为中断
func dealInputDataPoints(commErrSubDevices map[definition.DeviceGidType]struct{}, points model.DataPoints) {
	for i := range points {
		deviceGid, pointIdType, err := definition.SplitDataPointID(points[i].ID)
		if err != nil {
			continue
		}
		if pointIdType == definition.CommID {
			continue
		}

		if _, ok := commErrSubDevices[deviceGid]; ok {
			points[i].Rtd.Val.Qua = consts.QualityCommDisconnected

			// 以下兼容旧版数据服务
			//// 测点值置为 -99998，qua 置为正常
			//// 存储侧不会存储 qua 异常的测点
			//points[i].Rtd.Val.Qua = consts.QualityOk
			//points[i].Rtd.Val.Pv.SetValue(definition.OfflineValue)
		}
	}
}
