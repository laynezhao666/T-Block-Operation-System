// Package marshaller 提供门禁协议数据的序列化和反序列化功能。
package marshaller

import (
	"bytes"
	"fmt"
	"unsafe"

	"dac/entity/model/driver/cacs"
	"dac/entity/utils/dtcp"
	"dac/logic/collect/driver/cacs/consts"
)

// CACSMarshal CACS协议的序列化器
type CACSMarshal struct {
}

// NewCACSMarshal 创建CACS序列化器实例
func NewCACSMarshal() *CACSMarshal {
	return &CACSMarshal{}
}

// cacsMarshalFunc 定义 CACS 序列化函数类型。
type cacsMarshalFunc func(req interface{}) ([]byte, error)

// cacsMarshalAs 将强类型序列化函数适配为通用的 cacsMarshalFunc。
func cacsMarshalAs[T any](fn func(T) []byte) cacsMarshalFunc {
	return func(req interface{}) ([]byte, error) {
		r, ok := req.(T)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		return fn(r), nil
	}
}

// cacsMarshalRegistry 存储 CACS 命令到序列化函数的映射。
var cacsMarshalRegistry = map[uint32]cacsMarshalFunc{
	consts.KCommandRequestDoorStatus: cacsMarshalAs[cacs.DoorStateReq](
		RequestDoorStatusMarshal),
	consts.KCommandRequestRemoteControl: cacsMarshalAs[cacs.DoorControlReq](
		RequestDoorControlMarshal),
	consts.KCommandRequestDownloadControllerParams: cacsMarshalAs[cacs.DownloadControllerParamsReq](
		ReqDlCtrlParamsMarshal),
	consts.KCommandRequestGetControllerParams: cacsMarshalAs[cacs.GetControllerParamsReq](
		RequestGetControllerParamsMarshal),
	consts.KCommandRequestDeleteCards: cacsMarshalAs[cacs.DeleteCardsReq](
		RequestDeleteCardsMarshal),
	consts.KCommandRequestDownloadCards: cacsMarshalAs[cacs.DownloadCardsReq](
		RequestDownloadCardsMarshal),
	consts.KCommandRequestGetCards: cacsMarshalAs[cacs.GetCardsReq](
		RequestGetCardsMarshal),
	consts.KCommandRequestDownloadDoorParams: cacsMarshalAs[cacs.DownloadDoorParamsReq](
		RequestDownloadDoorParamsMarshal),
	consts.KCommandRequestGetDoorParams: cacsMarshalAs[cacs.GetDoorParamsReq](
		RequestGetDoorParamsMarshal),
	consts.KCommandRequestSetTime: cacsMarshalAs[cacs.SetTimeReq](
		RequestSetTimeMarshal),
	consts.KCommandRequestAddTimeGroups: cacsMarshalAs[cacs.AddTimeGroupsReq](
		RequestAddTimeGroupsMarshal),
	consts.KCommandRequestGetTimeGroups: cacsMarshalAs[cacs.GetTimeGroupsReq](
		RequestGetTimeGroupsMarshal),
	consts.KCommandRequestDeleteTimeGroups: cacsMarshalAs[cacs.DeleteTimeGroupsReq](
		RequestDeleteTimeGroupsMarshal),
	consts.KCommandRequestGetCardsInfo: cacsMarshalAs[cacs.GetCardsInfoReq](
		RequestGetCardsInfoMarshal),
}

// Marshal 根据命令码序列化 CACS 请求数据。
func (c *CACSMarshal) Marshal(cmd uint32, req interface{}) ([]byte, error) {
	fn, ok := cacsMarshalRegistry[cmd]
	if !ok {
		return nil, fmt.Errorf("unsupported command cmd: %x", cmd)
	}
	return fn(req)
}

// RequestGetCardsInfoMarshal 序列化获取卡信息请求
func RequestGetCardsInfoMarshal(req cacs.GetCardsInfoReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint32Little(req.Index))
	return buf.Bytes()
}

// RequestDeleteTimeGroupsMarshal 序列化删除时间组请求
func RequestDeleteTimeGroupsMarshal(req cacs.DeleteTimeGroupsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.Type))
	buf.Write(dtcp.WriteUint8(req.Id))
	buf.Write(dtcp.WriteUint8(req.WhatDay))
	return buf.Bytes()
}

// RequestGetTimeGroupsMarshal 序列化获取时间组请求
func RequestGetTimeGroupsMarshal(req cacs.GetTimeGroupsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.Id))
	buf.Write(dtcp.WriteUint8(req.WhatDay))
	return buf.Bytes()
}

// TimeGroupMarshal 序列化单个时间组数据
func TimeGroupMarshal(req cacs.TimeGroup) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.StartHour))
	buf.Write(dtcp.WriteUint8(req.StartMinute))
	buf.Write(dtcp.WriteUint8(req.EndHour))
	buf.Write(dtcp.WriteUint8(req.EndMinute))
	return buf.Bytes()
}

// RequestAddTimeGroupsMarshal 序列化添加时间组请求
func RequestAddTimeGroupsMarshal(req cacs.AddTimeGroupsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.Id))
	buf.Write(dtcp.WriteUint8(req.WhatDay))
	for i := 0; i < len(req.TimeGroups); i++ {
		buf.Write(TimeGroupMarshal(req.TimeGroups[i]))
	}
	return buf.Bytes()
}

// RequestSetTimeMarshal 序列化设置时间请求
func RequestSetTimeMarshal(req cacs.SetTimeReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint16Little(req.Year))
	buf.Write(dtcp.WriteUint8(req.Month))
	buf.Write(dtcp.WriteUint8(req.Day))
	buf.Write(dtcp.WriteUint8(req.Hour))
	buf.Write(dtcp.WriteUint8(req.Minute))
	buf.Write(dtcp.WriteUint8(req.Second))
	return buf.Bytes()
}

// RequestGetDoorParamsMarshal 序列化获取门参数请求
func RequestGetDoorParamsMarshal(req cacs.GetDoorParamsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint32Little(req.Id))
	return buf.Bytes()
}

// RequestDownloadDoorParamsMarshal 序列化下载门参数请求
func RequestDownloadDoorParamsMarshal(req cacs.DownloadDoorParamsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint32Little(req.Id))
	buf.Write(dtcp.WriteUint8(req.AuxiliaryDI))
	buf.Write(dtcp.WriteUint8(req.ReservedDI1))
	buf.Write(dtcp.WriteUint8(req.ReservedDI2))
	buf.Write(dtcp.WriteUint8(req.AlarmDO))
	buf.Write(dtcp.WriteUint8(req.ReservedDO1))
	buf.Write(dtcp.WriteUint8(req.ReservedDO2))
	buf.Write(dtcp.WriteUint8(req.DoorSensorStatus))
	buf.Write(dtcp.WriteUint8(req.ElectricLockStatus))
	buf.Write(dtcp.WriteUint8(req.ButtonStatus))
	buf.Write(dtcp.WriteUint8(req.DoorMode))
	buf.Write(dtcp.WriteUint8(req.MultiCardsOpenDoorNum))
	buf.Write(dtcp.WriteUint8(req.AntiPassback))
	buf.Write(dtcp.WriteUint16Little(req.EntryAreaNum))
	buf.Write(dtcp.WriteUint16Little(req.ExitAreaNum))
	buf.Write(dtcp.WriteUint32Little(req.OpenDoorKeepTime))
	buf.Write(dtcp.WriteUint32Little(req.OpenDoorTimeoutTime))
	buf.Write(dtcp.WriteUint32Little(req.AlarmTime))
	buf.Write(dtcp.WriteUint32Little(req.CoercivePassword))
	buf.Write(dtcp.WriteUint32Little(req.Password))
	buf.Write(dtcp.WriteUint32Little(req.InterLockId1))
	buf.Write(dtcp.WriteUint32Little(req.InterLockId2))
	buf.Write(dtcp.WriteUint32Little(req.InterLockId3))
	buf.Write(dtcp.WriteUint32Little(req.MultiCardsOpenDoorInterval))
	buf.Write(dtcp.WriteUint32Little(req.CardPasswordInterval))
	buf.Write(dtcp.WriteUint32Little(req.LockType))
	return buf.Bytes()
}

// RequestGetCardsMarshal 序列化获取卡请求
func RequestGetCardsMarshal(req cacs.GetCardsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.Type))
	buf.Write(dtcp.WriteUint32Little(req.Id))
	return buf.Bytes()
}

// RequestDownloadCardsMarshal 序列化下载卡请求
func RequestDownloadCardsMarshal(req cacs.DownloadCardsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.Num))
	for i := 0; i < int(req.Num); i++ {
		card := req.Cards[i]
		buf.Write(dtcp.WriteUint32Little(card.Id))
		buf.Write(dtcp.WriteUint32Little(card.UserId))
		buf.Write(dtcp.WriteUint32Little(card.Password))
		buf.Write(dtcp.WriteUint16Little(card.StartYear))
		buf.Write(dtcp.WriteUint8(card.StartMonth))
		buf.Write(dtcp.WriteUint8(card.StartDay))
		buf.Write(dtcp.WriteUint8(card.StartHour))
		buf.Write(dtcp.WriteUint8(card.StartMinute))
		buf.Write(dtcp.WriteUint8(card.StartSecond))
		buf.Write(dtcp.WriteUint8(card.Reserved))
		buf.Write(dtcp.WriteUint16Little(card.EndYear))
		buf.Write(dtcp.WriteUint8(card.EndMonth))
		buf.Write(dtcp.WriteUint8(card.EndDay))
		buf.Write(dtcp.WriteUint8(card.EndHour))
		buf.Write(dtcp.WriteUint8(card.EndMinute))
		buf.Write(dtcp.WriteUint8(card.EndSecond))
		buf.Write(dtcp.WriteUint8(card.CardType))
		buf.Write([]byte{card.AuthDoor1.PermitPeriod, card.AuthDoor1.AuthType[0], card.AuthDoor1.AuthType[1]})
		buf.Write([]byte{card.AuthDoor2.PermitPeriod, card.AuthDoor2.AuthType[0], card.AuthDoor2.AuthType[1]})
		buf.Write([]byte{card.AuthDoor3.PermitPeriod, card.AuthDoor3.AuthType[0], card.AuthDoor3.AuthType[1]})
		buf.Write([]byte{card.AuthDoor4.PermitPeriod, card.AuthDoor4.AuthType[0], card.AuthDoor4.AuthType[1]})
		buf.Write(dtcp.WriteUint16Little(card.AreaId))
	}
	return buf.Bytes()
}

// RequestDeleteCardsMarshal 序列化删除卡请求
func RequestDeleteCardsMarshal(req cacs.DeleteCardsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.Type))
	buf.Write(dtcp.WriteUint32Little(req.Id))
	return buf.Bytes()
}

// RequestGetControllerParamsMarshal 序列化获取控制器参数请求
func RequestGetControllerParamsMarshal(req cacs.GetControllerParamsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	return buf.Bytes()
}

// ResponseControllerRegisterMarshal 序列化控制器注册响应
func ResponseControllerRegisterMarshal(resp cacs.ControllerRegisterResp) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint32Little(resp.Id))
	return buf.Bytes()
}

// RequestDoorStatusMarshal 序列化门状态查询请求
func RequestDoorStatusMarshal(req cacs.DoorStateReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint32Little(req.Id))
	return buf.Bytes()
}

// RequestDoorControlMarshal 序列化门控制请求
func RequestDoorControlMarshal(req cacs.DoorControlReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.ControlMode))
	buf.Write(dtcp.WriteUint32Little(req.Id))
	return buf.Bytes()
}

// ReqDlCtrlParamsMarshal 序列化下载控制器参数请求。
func ReqDlCtrlParamsMarshal(req cacs.DownloadControllerParamsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.Mode))
	buf.Write([]byte(req.Name))
	return buf.Bytes()
}

// cacsUnmarshalFunc 定义 CACS 反序列化函数类型。
type cacsUnmarshalFunc func(data []byte) (interface{}, error)

// unmarshalWrap 将强类型反序列化函数适配为通用的 cacsUnmarshalFunc。
func unmarshalWrap[T any](fn func([]byte) (T, error)) cacsUnmarshalFunc {
	return func(data []byte) (interface{}, error) {
		return fn(data)
	}
}

// cacsUnmarshalRegistry 存储 CACS 命令到反序列化函数的映射。
var cacsUnmarshalRegistry = map[uint32]cacsUnmarshalFunc{
	consts.KCommandResponseDoorStatus:               unmarshalWrap(ResponseDoorStatusUnMarshal),
	consts.KCommandResponseRemoteControl:            unmarshalWrap(ResponseRemoteControlUnMarshal),
	consts.KCommandResponseDownloadControllerParams: unmarshalWrap(RespDlCtrlParamsUnmarshal),
	consts.KCommandResponseGetControllerParams:      unmarshalWrap(RespGetCtrlParamsUnmarshal),
	consts.KCommandResponseDeleteCards:              unmarshalWrap(ResponseDeleteCardsUnMarshal),
	consts.KCommandResponseDownloadCards:            unmarshalWrap(ResponseDownloadCardsUnMarshal),
	consts.KCommandResponseGetCards:                 unmarshalWrap(ResponseGetCardsUnMarshal),
	consts.KCommandResponseDownloadDoorParams:       unmarshalWrap(ResponseDownloadDoorParamsUnMarshal),
	consts.KCommandResponseGetDoorParams:            unmarshalWrap(ResponseGetDoorParamsUnMarshal),
	consts.KCommandResponseSetTime:                  unmarshalWrap(ResponseSetTimeUnMarshal),
	consts.KCommandResponseAddTimeGroups:            unmarshalWrap(ResponseAddTimeGroupsUnMarshal),
	consts.KCommandResponseGetTimeGroups:            unmarshalWrap(ResponseGetTimeGroupsUnMarshal),
	consts.KCommandResponseDeleteTimeGroups:         unmarshalWrap(ResponseDeleteTimeGroupsUnMarshal),
	consts.KCommandResponseGetCardsInfo:             unmarshalWrap(ResponseGetCardsInfoUnMarshal),
}

// Unmarshal 根据命令码反序列化 CACS 响应数据。
func (c *CACSMarshal) Unmarshal(cmd uint32, data []byte) (interface{}, error) {
	fn, ok := cacsUnmarshalRegistry[cmd]
	if !ok {
		return nil, fmt.Errorf("unsupported command cmd: %x", cmd)
	}
	return fn(data)
}

// ResponseGetCardsInfoUnMarshal 反序列化获取卡信息响应。
func ResponseGetCardsInfoUnMarshal(data []byte) (cacs.GetCardsInfoResp, error) {
	var resp cacs.GetCardsInfoResp
	var err error
	minLen := int(unsafe.Sizeof(resp.NextIndex)) +
		int(unsafe.Sizeof(resp.Num))
	if len(data) < minLen {
		return resp, fmt.Errorf(
			"data length 不一致, len(data): %d", len(data))
	}
	buf := bytes.NewBuffer(data)
	resp.NextIndex = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.NextIndex))))
	resp.Num = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.Num))))
	expectLen := int(unsafe.Sizeof(resp.Num)) +
		int(unsafe.Sizeof(resp.NextIndex)) +
		int(resp.Num)*cacs.GetFieldSizeSum(cacs.CardInfo{})
	if expectLen != len(data) {
		return resp, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectLen)
	}
	for i := 0; i < int(resp.Num); i++ {
		cardInfo := cacs.CardInfo{}
		cardInfo, err = CardInfoUnMarshal(buf.Next(cacs.GetFieldSizeSum(cardInfo)))
		if err != nil {
			return resp, err
		}
		resp.Cards = append(resp.Cards, cardInfo)
	}
	return resp, nil
}

// ResponseDeleteTimeGroupsUnMarshal 反序列化删除时间组响应。
func ResponseDeleteTimeGroupsUnMarshal(data []byte) (cacs.DeleteTimeGroupsResp, error) {
	return cacs.DeleteTimeGroupsResp{}, nil
}

// ResponseGetTimeGroupsUnMarshal 反序列化获取时间组响应。
func ResponseGetTimeGroupsUnMarshal(data []byte) (cacs.GetTimeGroupsResp, error) {
	var resp cacs.GetTimeGroupsResp
	expectSize := cacs.GetFieldSizeSum(cacs.GetTimeGroupsResp{})
	if len(data) != expectSize {
		return resp, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectSize)
	}

	buf := bytes.NewBuffer(data)
	resp.Id = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.Id))))
	resp.WhatDay = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.WhatDay))))
	for i := 0; i < len(resp.TimeGroups); i++ {
		resp.TimeGroups[i].StartHour = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.TimeGroups[i].StartHour))))
		resp.TimeGroups[i].StartMinute = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.TimeGroups[i].StartMinute))))
		resp.TimeGroups[i].EndHour = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.TimeGroups[i].EndHour))))
		resp.TimeGroups[i].EndMinute = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.TimeGroups[i].EndMinute))))
	}
	return resp, nil
}

// ResponseAddTimeGroupsUnMarshal 反序列化添加时间组响应。
func ResponseAddTimeGroupsUnMarshal(data []byte) (cacs.AddTimeGroupsResp, error) {
	return cacs.AddTimeGroupsResp{}, nil
}

// ResponseGetDoorParamsUnMarshal 反序列化获取门参数响应。
func ResponseGetDoorParamsUnMarshal(data []byte) (cacs.GetDoorParamsResp, error) {
	var resp cacs.GetDoorParamsResp
	expectSize := cacs.GetFieldSizeSum(cacs.GetDoorParamsResp{})
	if len(data) != expectSize {
		return resp, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectSize)
	}

	buf := bytes.NewBuffer(data)
	// 解析门ID
	resp.Id = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.Id))))
	// 解析DI/DO端口配置
	resp.AuxiliaryDI = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.AuxiliaryDI))))
	resp.ReservedDI1 = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.ReservedDI1))))
	resp.ReservedDI2 = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.ReservedDI2))))
	resp.AlarmDO = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.AlarmDO))))
	resp.ReservedDO1 = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.ReservedDO1))))
	resp.ReservedDO2 = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.ReservedDO2))))
	// 解析门传感器和锁状态
	resp.DoorSensorStatus = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.DoorSensorStatus))))
	resp.ElectricLockStatus = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.ElectricLockStatus))))
	resp.ButtonStatus = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.ButtonStatus))))
	// 解析门模式和多卡开门配置
	resp.DoorMode = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.DoorMode))))
	resp.MultiCardsOpenDoorNum = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.MultiCardsOpenDoorNum))))
	resp.AntiPassback = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.AntiPassback))))
	// 解析区域编号
	resp.EntryAreaNum = dtcp.ReadUint16Little(buf.Next(int(unsafe.Sizeof(resp.EntryAreaNum))))
	resp.ExitAreaNum = dtcp.ReadUint16Little(buf.Next(int(unsafe.Sizeof(resp.ExitAreaNum))))
	// 解析时间参数（开门保持、超时、告警）
	resp.OpenDoorKeepTime = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.OpenDoorKeepTime))))
	resp.OpenDoorTimeoutTime = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.OpenDoorTimeoutTime))))
	resp.AlarmTime = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.AlarmTime))))
	// 解析密码和互锁配置
	resp.CoercivePassword = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.Password))))
	resp.Password = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.Password))))
	resp.InterLockId1 = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.InterLockId1))))
	resp.InterLockId2 = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.InterLockId2))))
	resp.InterLockId3 = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.InterLockId3))))
	resp.MultiCardsOpenDoorInterval = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.MultiCardsOpenDoorInterval))))
	resp.CardPasswordInterval = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.CardPasswordInterval))))
	resp.LockType = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.LockType))))
	return resp, nil
}

// ResponseDownloadDoorParamsUnMarshal 反序列化下载门参数响应。
func ResponseDownloadDoorParamsUnMarshal(data []byte) (cacs.DownloadDoorParamsResp, error) {
	return cacs.DownloadDoorParamsResp{}, nil
}

// AuthDoorUnMarshal 反序列化门授权信息。
func AuthDoorUnMarshal(data []byte) (cacs.DoorAuth, error) {
	var resp cacs.DoorAuth
	expectSize := cacs.GetFieldSizeSum(cacs.DoorAuth{})
	if len(data) != expectSize {
		return resp, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectSize)
	}

	buf := bytes.NewBuffer(data)
	resp.PermitPeriod = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.PermitPeriod))))
	copy(resp.AuthType[:], buf.Next(int(unsafe.Sizeof(resp.AuthType))))
	return resp, nil
}

// CardInfoUnMarshal 反序列化单张卡信息。
func CardInfoUnMarshal(data []byte) (cacs.CardInfo, error) {
	var cardInfo cacs.CardInfo
	var err error
	expectSize := cacs.GetFieldSizeSum(cacs.CardInfo{})
	if len(data) != expectSize {
		return cardInfo, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectSize)
	}

	buf := bytes.NewBuffer(data)
	// 解析卡基本信息（卡号、用户ID、密码）
	cardInfo.Id = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(cardInfo.Id))))
	cardInfo.UserId = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(cardInfo.UserId))))
	cardInfo.Password = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(cardInfo.Password))))
	// 解析卡有效期起始时间
	cardInfo.StartYear = dtcp.ReadUint16Little(buf.Next(int(unsafe.Sizeof(cardInfo.StartYear))))
	cardInfo.StartMonth = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.StartMonth))))
	cardInfo.StartDay = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.StartDay))))
	cardInfo.StartHour = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.StartHour))))
	cardInfo.StartMinute = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.StartMinute))))
	cardInfo.StartSecond = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.StartSecond))))
	cardInfo.Reserved = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.Reserved))))
	// 解析卡有效期结束时间
	cardInfo.EndYear = dtcp.ReadUint16Little(buf.Next(int(unsafe.Sizeof(cardInfo.EndYear))))
	cardInfo.EndMonth = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.EndMonth))))
	cardInfo.EndDay = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.EndDay))))
	cardInfo.EndHour = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.EndHour))))
	cardInfo.EndMinute = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.EndMinute))))
	cardInfo.EndSecond = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.EndSecond))))
	cardInfo.CardType = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(cardInfo.CardType))))
	// 解析4个门的授权信息
	authDoor1Data := buf.Next(cacs.GetFieldSizeSum(cardInfo.AuthDoor1))
	if cardInfo.AuthDoor1, err = AuthDoorUnMarshal(authDoor1Data); err != nil {
		return cardInfo, err
	}
	authDoor2Data := buf.Next(cacs.GetFieldSizeSum(cardInfo.AuthDoor2))
	if cardInfo.AuthDoor2, err = AuthDoorUnMarshal(authDoor2Data); err != nil {
		return cardInfo, err
	}
	authDoor3Data := buf.Next(cacs.GetFieldSizeSum(cardInfo.AuthDoor3))
	if cardInfo.AuthDoor3, err = AuthDoorUnMarshal(authDoor3Data); err != nil {
		return cardInfo, err
	}
	authDoor4Data := buf.Next(cacs.GetFieldSizeSum(cardInfo.AuthDoor4))
	if cardInfo.AuthDoor4, err = AuthDoorUnMarshal(authDoor4Data); err != nil {
		return cardInfo, err
	}
	cardInfo.AreaId = dtcp.ReadUint16Little(buf.Next(int(unsafe.Sizeof(cardInfo.AreaId))))
	return cardInfo, nil
}

// ResponseGetCardsUnMarshal 反序列化获取卡响应。
func ResponseGetCardsUnMarshal(data []byte) (cacs.GetCardsResp, error) {
	var resp cacs.GetCardsResp
	var err error
	expectSize := cacs.GetFieldSizeSum(cacs.GetCardsResp{})
	if len(data) != expectSize {
		return resp, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectSize)
	}

	buf := bytes.NewBuffer(data)
	resp.CardNum = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.CardNum))))

	// 按协议固定解析4张卡的信息（无效卡为全F）
	resp.Card1, err = CardInfoUnMarshal(buf.Next(cacs.GetFieldSizeSum(resp.Card1)))
	if err != nil {
		return resp, err
	}
	resp.Card2, err = CardInfoUnMarshal(buf.Next(cacs.GetFieldSizeSum(resp.Card2)))
	if err != nil {
		return resp, err
	}
	resp.Card3, err = CardInfoUnMarshal(buf.Next(cacs.GetFieldSizeSum(resp.Card3)))
	if err != nil {
		return resp, err
	}
	resp.Card4, err = CardInfoUnMarshal(buf.Next(cacs.GetFieldSizeSum(resp.Card4)))
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// ResponseDownloadCardsUnMarshal 反序列化下载卡响应。
func ResponseDownloadCardsUnMarshal(data []byte) (cacs.DownloadCardsResp, error) {
	var resp cacs.DownloadCardsResp
	expectSize := cacs.GetFieldSizeSum(cacs.DownloadCardsResp{})
	if len(data) != expectSize {
		return resp, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectSize)
	}

	buf := bytes.NewBuffer(data)
	resp.SuccessNum = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.SuccessNum))))
	resp.FailCardId = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.FailCardId))))
	return resp, nil
}

// ResponseDeleteCardsUnMarshal 反序列化删除卡响应。
func ResponseDeleteCardsUnMarshal(data []byte) (cacs.DeleteCardsResp, error) {
	return cacs.DeleteCardsResp{}, nil
}

// RespGetCtrlParamsUnmarshal 反序列化获取控制器参数响应。
func RespGetCtrlParamsUnmarshal(data []byte) (cacs.GetControllerParamsResp, error) {
	var resp cacs.GetControllerParamsResp
	expectLen := cacs.GetFieldSizeSum(cacs.GetControllerParamsResp{})
	if len(data) != expectLen {
		return resp, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectLen)
	}

	buf := bytes.NewBuffer(data)
	resp.Mode = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.Mode))))
	copy(resp.Name[:], buf.Next(consts.KControllerNameLen))
	return resp, nil
}

// RequestControllerRegisterUnMarshal 反序列化控制器注册请求。
func RequestControllerRegisterUnMarshal(data []byte) (cacs.ControllerRegisterReq, error) {
	var resp cacs.ControllerRegisterReq
	expectSize := cacs.GetFieldSizeSum(cacs.ControllerRegisterReq{})
	if len(data) != expectSize {
		return resp, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectSize)
	}

	buf := bytes.NewBuffer(data)
	resp.Id = dtcp.ReadUint32Little(
		buf.Next(int(unsafe.Sizeof(resp.Id))))
	resp.Version = dtcp.ReadUint32Little(
		buf.Next(int(unsafe.Sizeof(resp.Version))))
	resp.Seq = dtcp.ReadUint32Little(
		buf.Next(int(unsafe.Sizeof(resp.Seq))))
	resp.EventAlarmSeq = dtcp.ReadUint32Little(
		buf.Next(int(unsafe.Sizeof(resp.EventAlarmSeq))))
	copy(resp.MAC[:], buf.Next(consts.KMACAddrLen))
	copy(resp.Name[:], buf.Next(consts.KControllerNameLen))
	resp.Mode = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(resp.Mode))))
	return resp, nil
}

// ResponseDoorStatusUnMarshal 反序列化门状态响应。
func ResponseDoorStatusUnMarshal(data []byte) (cacs.DoorStateResp, error) {
	var resp cacs.DoorStateResp
	expectSize := cacs.GetFieldSizeSum(cacs.DoorStateResp{})
	if len(data) != expectSize {
		return resp, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectSize)
	}

	buf := bytes.NewBuffer(data)
	resp.Id = dtcp.ReadUint32Little(
		buf.Next(int(unsafe.Sizeof(resp.Id))))
	resp.AuxiliaryDI = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.AuxiliaryDI))))
	resp.ReservedDI1 = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.ReservedDI1))))
	resp.ReservedDI2 = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.ReservedDI2))))
	resp.AlarmDO = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.AlarmDO))))
	resp.ReservedDO1 = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.ReservedDO1))))
	resp.ReservedDO2 = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.ReservedDO2))))
	resp.DoorSensorStatus = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.DoorSensorStatus))))
	resp.ButtonStatus = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.ButtonStatus))))
	resp.ElectricLockStatus = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.ElectricLockStatus))))
	resp.OpenDoorStatus = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.OpenDoorStatus))))
	resp.DoorMode = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(resp.DoorMode))))

	return resp, nil
}

// ResponseRemoteControlUnMarshal 反序列化远程控制响应。
func ResponseRemoteControlUnMarshal(data []byte) (cacs.DoorControlResp, error) {
	return cacs.DoorControlResp{}, nil
}

// RespDlCtrlParamsUnmarshal 反序列化下载控制器参数响应。
func RespDlCtrlParamsUnmarshal(data []byte) (cacs.DownloadControllerParamsResp, error) {
	return cacs.DownloadControllerParamsResp{}, nil
}

// RequestUploadDoorStatusUnMarshal 反序列化上传门状态请求。
func RequestUploadDoorStatusUnMarshal(data []byte) (cacs.UploadDoorStatus, error) {
	var doorStatus cacs.UploadDoorStatus
	expectSize := cacs.GetFieldSizeSum(cacs.UploadDoorStatus{})
	if len(data) != expectSize {
		return doorStatus, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectSize)
	}

	buf := bytes.NewBuffer(data)
	doorStatus.Id = dtcp.ReadUint32Little(
		buf.Next(int(unsafe.Sizeof(doorStatus.Id))))
	doorStatus.AuxiliaryDI = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.AuxiliaryDI))))
	doorStatus.ReservedDI1 = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.ReservedDI1))))
	doorStatus.ReservedDI2 = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.ReservedDI2))))
	doorStatus.AlarmDO = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.AlarmDO))))
	doorStatus.ReservedDO1 = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.ReservedDO1))))
	doorStatus.ReservedDO2 = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.ReservedDO2))))
	doorStatus.DoorSensorStatus = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.DoorSensorStatus))))
	doorStatus.ButtonStatus = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.ButtonStatus))))
	doorStatus.ElectricLockStatus = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.ElectricLockStatus))))
	doorStatus.OpenDoorStatus = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.OpenDoorStatus))))
	doorStatus.DoorMode = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(doorStatus.DoorMode))))
	return doorStatus, nil
}

// ReqUploadCtrlStatusUnmarshal 反序列化上传控制器状态请求。
func ReqUploadCtrlStatusUnmarshal(data []byte) (cacs.UploadControllerStatus, error) {
	var uploadControllerStatus cacs.UploadControllerStatus
	expectSize := cacs.GetFieldSizeSum(cacs.UploadControllerStatus{})
	if len(data) != expectSize {
		return uploadControllerStatus, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectSize)
	}

	buf := bytes.NewBuffer(data)
	uploadControllerStatus.Year = dtcp.ReadUint16Little(
		buf.Next(int(unsafe.Sizeof(uploadControllerStatus.Year))))
	uploadControllerStatus.Month = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(uploadControllerStatus.Month))))
	uploadControllerStatus.Day = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(uploadControllerStatus.Day))))
	uploadControllerStatus.Hour = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(uploadControllerStatus.Hour))))
	uploadControllerStatus.Minute = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(uploadControllerStatus.Minute))))
	uploadControllerStatus.Second = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(uploadControllerStatus.Second))))
	uploadControllerStatus.FireAlarmStatus = dtcp.ReadUint8(
		buf.Next(int(unsafe.Sizeof(uploadControllerStatus.FireAlarmStatus))))
	return uploadControllerStatus, nil
}

// RequestUploadEventAlarmUnMarshal 反序列化上传事件告警请求。
func RequestUploadEventAlarmUnMarshal(data []byte) (cacs.UploadEventAlarmReq, error) {
	var uploadEventAlarm cacs.UploadEventAlarmReq
	minLen := int(unsafe.Sizeof(uploadEventAlarm.Num)) +
		int(unsafe.Sizeof(uploadEventAlarm.Seq))
	if len(data) < minLen {
		return uploadEventAlarm, fmt.Errorf(
			"data length 不一致, len(data): %d", len(data))
	}

	buf := bytes.NewBuffer(data)
	// 解析事件告警数量和序列号
	uploadEventAlarm.Num = dtcp.ReadUint16Little(buf.Next(int(unsafe.Sizeof(uploadEventAlarm.Num))))
	uploadEventAlarm.Seq = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(uploadEventAlarm.Seq))))
	itemSize := cacs.GetFieldSizeSum(cacs.EventAlarmItem{})
	// 校验数据总长度与事件数量是否匹配
	expectLen := int(uploadEventAlarm.Num)*itemSize +
		int(unsafe.Sizeof(uploadEventAlarm.Num)) +
		int(unsafe.Sizeof(uploadEventAlarm.Seq))
	if expectLen != len(data) {
		return uploadEventAlarm, fmt.Errorf(
			"data length 不一致, len(data): %d, expect: %d",
			len(data), expectLen)
	}
	for i := 0; i < int(uploadEventAlarm.Num); i++ {
		var item cacs.EventAlarmItem
		item.Type = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(item.Type))))
		item.Year = dtcp.ReadUint16Little(buf.Next(int(unsafe.Sizeof(item.Year))))
		item.Month = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(item.Month))))
		item.Day = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(item.Day))))
		item.Hour = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(item.Hour))))
		item.Minute = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(item.Minute))))
		item.Second = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(item.Second))))
		item.DoorId = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(item.DoorId))))
		item.CardId = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(item.CardId))))
		item.Extras = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(item.Extras))))
		item.CardReaderId = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(item.CardReaderId))))
		uploadEventAlarm.Items = append(uploadEventAlarm.Items, item)
	}

	fmt.Printf("\n--------------------------------------------------------------\n")
	fmt.Printf("✓ 解析完成，共 %d 个事件告警项\n", len(uploadEventAlarm.Items))
	fmt.Printf("==============================================================\n\n")

	return uploadEventAlarm, nil
}

// ResponseUploadEventAlarm 序列化上传事件告警响应。
func ResponseUploadEventAlarm(resp cacs.UploadEventAlarmResp) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint16Little(resp.SuccessNum))
	buf.Write(dtcp.WriteUint32Little(resp.EventAlarmSeq))
	return buf.Bytes()
}

// ResponseSetTimeUnMarshal 反序列化设置时间响应。
func ResponseSetTimeUnMarshal(data []byte) (cacs.SetTimeResp, error) {
	return cacs.SetTimeResp{}, nil
}
