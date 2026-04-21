// Package marshaller 提供门禁控制器协议的序列化和反序列化功能。
package marshaller

import (
	"bytes"
	"fmt"
	"unsafe"

	"dac/entity/model/driver/xbrother"
	"dac/entity/utils/dtcp"
	"dac/logic/collect/driver/xbrother/consts"
)

// XBrotherMarshaller XBrother协议的序列化/反序列化器
type XBrotherMarshaller struct {
}

// NewXBrotherMarshaller 创建XBrother序列化器实例
func NewXBrotherMarshaller() *XBrotherMarshaller {
	return &XBrotherMarshaller{}
}

// marshalFunc 定义序列化函数类型，接收任意请求并返回字节切片。
type marshalFunc func(req interface{}) ([]byte, error)

// xbMarshalRegistry 存储命令到序列化函数的映射。
var xbMarshalRegistry = map[uint32]marshalFunc{
	uint32(consts.CommandSetControllerParams): marshalAs[xbrother.SetControllerParamsReq](setControllerParamsReqMarshal),
	uint32(consts.CommandOpenDoor):            marshalAs[xbrother.OpenDoorReq](openDoorReqMarshal),
	uint32(consts.CommandDoorOpenPermenently): marshalAs[xbrother.OpenDoorPermenentlyReq](openDoorPermenentlyReqMarshal),
	uint32(consts.CommandCloseDoor):           marshalAs[xbrother.CloseDoorReq](closeDoorReqMarshal),
	uint32(consts.CommandLockDoor):            marshalAs[xbrother.LockDoorReq](lockDoorReqMarshal),
	uint32(consts.CommandSetTime):             marshalAs[xbrother.SetTimeReq](setTimeReqMarshal),
	uint32(consts.CommandSetDoorParams):       marshalAs[xbrother.SetDoorParamsReq](setDoorParamsReqMarshal),
	uint32(consts.CommandClearTimeGroups):     marshalAs[xbrother.ClearDoorTimeGroupsReq](clearDoorTimeGroupsReqMarshal),
	uint32(consts.CommandAddTimeGroup):        marshalAs[xbrother.AddTimeGroupReq](addTimeGroupsReqMarshal),
	uint32(consts.CommandClearCards):          marshalAs[xbrother.ClearCardsReq](clearCardsReqMarshal),
	uint32(consts.CommandClean):               marshalAs[xbrother.CleanReq](resetReqMarshal),
	uint32(consts.CommandSetAlarm):            marshalAs[xbrother.AlarmSettingReq](alarmSettingReqMarshal),
	uint32(consts.CommandSetFireAlarm):        marshalAs[xbrother.AlarmSettingReq](alarmSettingReqMarshal),
	uint32(consts.CommandDeleteCard):          marshalAs[xbrother.DeleteCardReq](deleteCardReqMarshal),
	uint32(consts.CommandAddCard):             marshalAs[xbrother.AddCardReq](addCardReqMarshal),
}

// marshalAs 将强类型序列化函数适配为通用的 marshalFunc。
func marshalAs[T any](fn func(T) []byte) marshalFunc {
	return func(req interface{}) ([]byte, error) {
		r, ok := req.(T)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		return fn(r), nil
	}
}

// Marshal 根据命令码序列化请求数据。
func (c *XBrotherMarshaller) Marshal(cmd uint32, req interface{}) ([]byte, error) {
	fn, ok := xbMarshalRegistry[cmd]
	if !ok {
		return nil, fmt.Errorf("unsupported command cmd: %x", cmd)
	}
	return fn(req)
}

// addCardReqMarshal 序列化添加卡请求
func addCardReqMarshal(req xbrother.AddCardReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint16Little(req.CardIndex))
	buf.Write(dtcp.WriteUint32Little(req.CardId))
	buf.Write(dtcp.WriteUint16Big(req.Password))
	// 每个门的权限实际应该是小端序，四门是单字节控制，这里统一marshal不受影响；
	// 而单双门是双字节控制，字节序放transferAccessTimeGroup()处理
	buf.Write(dtcp.WriteUint32Big(req.AccessTimeGroup))
	buf.Write(dtcp.WriteUint32Big(req.Reserved))
	buf.Write(dtcp.WriteUint8(req.Status))
	return buf.Bytes()
}

// deleteCardReqMarshal 序列化删除卡请求
func deleteCardReqMarshal(req xbrother.DeleteCardReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint16Little(req.CardIndex))
	buf.Write(dtcp.WriteUint32Big(req.CardId))
	return buf.Bytes()
}

// lockDoorReqMarshal 序列化锁门请求
func lockDoorReqMarshal(req xbrother.LockDoorReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.LockStatus))
	return buf.Bytes()
}

// alarmSettingReqMarshal 序列化告警设置请求
func alarmSettingReqMarshal(req xbrother.AlarmSettingReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.DisableAlarm))
	buf.Write(dtcp.WriteUint8(req.KeepEnableAlarm))
	return buf.Bytes()
}

// resetReqMarshal 序列化重置请求
func resetReqMarshal(req xbrother.CleanReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	return buf.Bytes()
}

// clearCardsReqMarshal 序列化清除所有卡请求
func clearCardsReqMarshal(req xbrother.ClearCardsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	return buf.Bytes()
}

// addTimeGroupsReqMarshal 序列化添加时间组请求
func addTimeGroupsReqMarshal(req xbrother.AddTimeGroupReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.TimeZone))
	buf.Write(dtcp.WriteUint8(req.StartHour))
	buf.Write(dtcp.WriteUint8(req.StartMinute))
	buf.Write(dtcp.WriteUint8(req.EndHour))
	buf.Write(dtcp.WriteUint8(req.EndMinute))
	buf.Write(dtcp.WriteUint8(req.WeekDay))
	buf.Write(dtcp.WriteUint8(req.OpenDoorType))
	buf.Write(dtcp.WriteUint8(req.DeadlineYear))
	buf.Write(dtcp.WriteUint8(req.DeadlineMonth))
	buf.Write(dtcp.WriteUint8(req.DeadlineDay))
	return buf.Bytes()
}

// clearDoorTimeGroupsReqMarshal 序列化清除门时间组请求
func clearDoorTimeGroupsReqMarshal(req xbrother.ClearDoorTimeGroupsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	return buf.Bytes()
}

// setDoorParamsReqMarshal 序列化设置门参数请求
func setDoorParamsReqMarshal(req xbrother.SetDoorParamsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(uint8(req.OpenDoorTime & 0xff))) // 低8位
	buf.Write(dtcp.WriteUint8(req.OpenDoorTimeout))
	buf.Write(dtcp.WriteUint8(req.BidirectionalDetect))
	buf.Write(dtcp.WriteUint8(req.LongTimeOpenAlarm))
	buf.Write(dtcp.WriteUint8(uint8(req.OpenDoorTime >> 8))) // 高8位
	buf.Write(dtcp.WriteUint8(req.AlarmType))
	buf.Write(dtcp.WriteUint16Little(req.AlarmTime))
	return buf.Bytes()
}

// setTimeReqMarshal 序列化设置时间请求
func setTimeReqMarshal(req xbrother.SetTimeReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(req.Second))
	buf.Write(dtcp.WriteUint8(req.Minute))
	buf.Write(dtcp.WriteUint8(req.Hour))
	buf.Write(dtcp.WriteUint8(req.Week))
	buf.Write(dtcp.WriteUint8(req.Day))
	buf.Write(dtcp.WriteUint8(req.Month))
	buf.Write(dtcp.WriteUint8(req.Year))
	return buf.Bytes()
}

// openDoorReqMarshal 序列化开门请求
func openDoorReqMarshal(r xbrother.OpenDoorReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(r.DoorNo))
	return buf.Bytes()
}

// openDoorPermenentlyReqMarshal 序列化常开门请求
func openDoorPermenentlyReqMarshal(r xbrother.OpenDoorPermenentlyReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(r.DoorNo))
	return buf.Bytes()
}

// closeDoorReqMarshal 序列化关门请求
func closeDoorReqMarshal(r xbrother.CloseDoorReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(r.DoorNo))
	return buf.Bytes()
}

// setControllerParamsReqMarshal 序列化设置控制器参数请求
func setControllerParamsReqMarshal(r xbrother.SetControllerParamsReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(r.InterLockType))
	buf.Write(dtcp.WriteUint16Little(r.FireAlarmTime))
	buf.Write(dtcp.WriteUint16Little(r.AlarmTime))
	buf.Write(dtcp.WriteUint16Big(r.CoercePassword))
	return buf.Bytes()
}

// setControllerAddrReqMarshal 序列化设置控制器地址请求
func setControllerAddrReqMarshal(req xbrother.SetControllerAddrReq) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	slice := make([]uint8, len(req.Addr))
	for i, v := range req.Addr {
		slice[i] = v
	}
	buf.Write(slice)
	return buf.Bytes()
}

// AlarmUploadRespMarshal 序列化告警上报响应
func AlarmUploadRespMarshal(resp xbrother.AlarmUploadResp) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(resp.Seq))
	return buf.Bytes()
}

// CommonRespMarshal 序列化通用响应
func CommonRespMarshal(resp xbrother.CommonResp) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(resp.Rtn))
	return buf.Bytes()
}

// EventUploadRespMarshal 序列化事件上报响应
func EventUploadRespMarshal(resp xbrother.EventUploadResp) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dtcp.WriteUint8(resp.Seq))
	return buf.Bytes()
}

// xbUnmarshalCmds 列出所有使用 commonRespUnMarshal 的命令码。
var xbUnmarshalCmds = map[uint32]struct{}{
	uint32(consts.CommandSetControllerParams): {},
	uint32(consts.CommandOpenDoor):            {},
	uint32(consts.CommandDoorOpenPermenently): {},
	uint32(consts.CommandCloseDoor):           {},
	uint32(consts.CommandLockDoor):            {},
	uint32(consts.CommandSetTime):             {},
	uint32(consts.CommandSetDoorParams):       {},
	uint32(consts.CommandClearTimeGroups):     {},
	uint32(consts.CommandAddTimeGroup):        {},
	uint32(consts.CommandClearCards):          {},
	uint32(consts.CommandClean):               {},
	uint32(consts.CommandSetAlarm):            {},
	uint32(consts.CommandSetFireAlarm):        {},
	uint32(consts.CommandDeleteCard):          {},
	uint32(consts.CommandAddCard):             {},
}

// Unmarshal 根据命令码反序列化响应数据。
func (c *XBrotherMarshaller) Unmarshal(cmd uint32, data []byte) (interface{}, error) {
	if _, ok := xbUnmarshalCmds[cmd]; ok {
		return commonRespUnMarshal(data)
	}
	return nil, fmt.Errorf("unsupported command cmd: %x", cmd)
}

// commonRespUnMarshal 反序列化通用响应
func commonRespUnMarshal(data []byte) (xbrother.CommonResp, error) {
	var commonResp xbrother.CommonResp
	if len(data) != xbrother.GetFieldSizeSum(xbrother.CommonResp{}) {
		return commonResp, fmt.Errorf("data length 不一致, len(data): %d, expect: %d",
			len(data), xbrother.GetFieldSizeSum(xbrother.CommonResp{}))
	}

	buf := bytes.NewBuffer(data)
	commonResp.Rtn = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(commonResp.Rtn))))
	return commonResp, nil
}

// AlarmUploadReqUnMarshal 反序列化告警上报请求
func AlarmUploadReqUnMarshal(data []byte) (xbrother.AlarmUploadReq, error) {
	var alarmUploadItem xbrother.AlarmUploadReq
	buf := bytes.NewBuffer(data)
	alarmUploadItem.Second = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(alarmUploadItem.Second))))
	alarmUploadItem.Minute = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(alarmUploadItem.Minute))))
	alarmUploadItem.Hour = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(alarmUploadItem.Hour))))
	alarmUploadItem.Day = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(alarmUploadItem.Day))))
	alarmUploadItem.Month = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(alarmUploadItem.Month))))
	alarmUploadItem.Year = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(alarmUploadItem.Year))))
	alarmUploadItem.Type = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(alarmUploadItem.Type))))
	alarmUploadItem.Door = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(alarmUploadItem.Door))))
	alarmUploadItem.HasNext = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(alarmUploadItem.HasNext))))
	alarmUploadItem.Seq = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(alarmUploadItem.Seq))))
	return alarmUploadItem, nil
}

// EventUploadReqUnMarshal 反序列化事件上报请求
func EventUploadReqUnMarshal(data []byte) (xbrother.EventUploadReq, error) {
	var eventUploadReq xbrother.EventUploadReq
	buf := bytes.NewBuffer(data)
	eventUploadReq.CardNo = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(eventUploadReq.CardNo))))
	eventUploadReq.Second = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(eventUploadReq.Second))))
	eventUploadReq.Minute = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(eventUploadReq.Minute))))
	eventUploadReq.Hour = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(eventUploadReq.Hour))))
	eventUploadReq.Day = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(eventUploadReq.Day))))
	eventUploadReq.Month = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(eventUploadReq.Month))))
	eventUploadReq.Year = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(eventUploadReq.Year))))
	eventUploadReq.Type = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(eventUploadReq.Type))))
	eventUploadReq.Door = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(eventUploadReq.Door))))
	eventUploadReq.HasNext = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(eventUploadReq.HasNext))))
	eventUploadReq.Seq = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(eventUploadReq.Seq))))
	return eventUploadReq, nil
}

// ControllerStatusUploadReqUnMarshal 反序列化控制器状态上报请求
func ControllerStatusUploadReqUnMarshal(data []byte) (xbrother.ControllerStatusUploadReq, error) {
	var req xbrother.ControllerStatusUploadReq
	buf := bytes.NewBuffer(data)
	req.Reserve1 = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.Reserve1))))
	req.Year = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.Year))))
	req.Month = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.Month))))
	req.Day = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.Day))))
	req.Hour = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.Hour))))
	req.Minute = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.Minute))))
	req.Second = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.Second))))
	req.DoorStatus = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.DoorStatus))))
	req.BatchSize = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.BatchSize))))
	req.Reserve2 = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.Reserve2))))
	req.FunctionType = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.FunctionType))))
	req.ControllerType = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.ControllerType))))
	req.LockStatus = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.LockStatus))))
	req.Reserve3 = dtcp.ReadUint32Big(buf.Next(int(unsafe.Sizeof(req.Reserve3))))
	req.AuxiliaryRelay = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.AuxiliaryRelay))))
	req.Version = dtcp.ReadUint8(buf.Next(int(unsafe.Sizeof(req.Version))))
	req.Reserve4 = dtcp.ReadUint16Big(buf.Next(int(unsafe.Sizeof(req.Reserve4))))
	return req, nil
}
