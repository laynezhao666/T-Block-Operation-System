package model

// AlarmActiveWS 活动告警websocket
type AlarmActiveWS struct {
	AlarmId      string   `protobuf:"varint,1,opt,name=alarm_id,json=alarmId,proto3" json:"alarm_id,omitempty"`
	Level        string   `protobuf:"bytes,2,opt,name=level,proto3" json:"level,omitempty"`
	AlarmName    string   `protobuf:"bytes,3,opt,name=alarm_name,json=alarmName,proto3" json:"alarm_name,omitempty"` // 告警类型/告警名称
	Rid          int32    `protobuf:"varint,4,opt,name=rid,proto3" json:"rid,omitempty"`
	DeviceGid    string   `protobuf:"bytes,5,opt,name=device_gid,json=deviceGid,proto3" json:"device_gid,omitempty"`            // 设备gid
	DeviceNumber string   `protobuf:"bytes,6,opt,name=device_number,json=deviceNumber,proto3" json:"device_number,omitempty"`   // 设备编号
	DeviceTypeZh string   `protobuf:"bytes,7,opt,name=device_type_zh,json=deviceTypeZh,proto3" json:"device_type_zh,omitempty"` // 设备类型（中文） 告警源: device_type_zh【device_number】
	Box          string   `protobuf:"bytes,8,opt,name=box,proto3" json:"box,omitempty"`                                         // 方仓  定位 ： box/room
	Room         string   `protobuf:"bytes,9,opt,name=room,proto3" json:"room,omitempty"`                                       // 房间
	MozuId       int32    `protobuf:"varint,10,opt,name=mozu_id,json=mozuId,proto3" json:"mozu_id,omitempty"`                   // 模组Id
	MozuName     string   `protobuf:"bytes,11,opt,name=mozu_name,json=mozuName,proto3" json:"mozu_name,omitempty"`              // 模组名称
	AlarmContent string   `protobuf:"bytes,12,opt,name=alarm_content,json=alarmContent,proto3" json:"alarm_content,omitempty"`  // 告警内容
	AlarmStatus  int32    `protobuf:"varint,13,opt,name=alarm_status,json=alarmStatus,proto3" json:"alarm_status,omitempty"`    // 0 正常 1挂起
	EventStatus  int32    `protobuf:"varint,14,opt,name=event_status,json=eventStatus,proto3" json:"event_status,omitempty"`    // 1未转单，2已转单，3已结单
	Points       []string `protobuf:"bytes,15,rep,name=points,proto3" json:"points,omitempty"`                                  // 告警计算所用的测点列表
	OccurTime    string   `protobuf:"bytes,16,opt,name=occur_time,json=occurTime,proto3" json:"occur_time,omitempty"`           // 触发时间
	RestoreTime  string   `protobuf:"bytes,17,opt,name=restore_time,json=restoreTime,proto3" json:"restore_time,omitempty"`     // 恢复时间
	RestoreType  string   `protobuf:"bytes,18,opt,name=restore_type,json=restoreType,proto3" json:"restore_type,omitempty"`     // 恢复方式
	HangupReason string   `protobuf:"bytes,19,opt,name=hangup_reason,json=hangupReason,proto3" json:"hangup_reason,omitempty"`  // 挂起原因
}
