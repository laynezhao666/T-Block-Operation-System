// Package driver 定义门禁驱动层的通用接口和数据模型。
package driver

import (
	"time"

	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/rt"
)

// ChannelInfo 门禁控制器通道访问参数
type ChannelInfo struct {
	ChannelID       string                 // 通道ID，格式为 "127.0.0.1" 或 "127.0.0.1:8888"
	Address         string                 // 地址，与具体协议相关
	Protocol        string                 // 协议名称
	ProtocolVersion string                 // 协议版本
	Extend          map[string]interface{} // 扩展参数
	TimeoutMS       time.Duration          // 请求超时时间
}

// ControllerBasicInfo 控制器基本信息
type ControllerBasicInfo struct {
	ID   db.IDType // 控制器ID
	Name string    // 控制器名称
}

// NewControllerBasicInfo 创建控制器基本信息实例
func NewControllerBasicInfo(id db.IDType, name string) ControllerBasicInfo {
	return ControllerBasicInfo{
		ID:   id,
		Name: name,
	}
}

// Controller 门禁控制器驱动接口，定义所有协议需实现的操作
type Controller interface {
	// Open 打开通道，等待发送指令
	Open(chanInfo ChannelInfo) consts.Quality
	// Close 关闭通道
	Close() consts.Quality

	Ping() error

	// GetDoorPoints  获取门相关测点
	GetDoorPoints(doors []int) (map[string]map[int]*rt.Point, error)
	// GetDoorState  获取门状态
	GetDoorState(doors []int) (map[int]*rt.Point, error)
	GetRawDoorState(doors []int) (interface{}, error)
	// SetDoorState 设置门状态
	SetDoorState(doorStates SetDoorStateRequest) error

	// GetTimeGroup 获取时间组
	GetTimeGroup(timeGroup int) (TimeGroup, error)
	// SetTimeGroup 修改时间组
	SetTimeGroup(timeGroup TimeGroup) error
	// ClearTimeGroup 删除时间组
	ClearTimeGroup(timeGroup int) error

	// GetTime 获取时间
	GetTime() (string, error)
	// SetTime 设置时间
	SetTime() error

	// GetCards 批量获取卡信息
	GetCards(offset int) (CardData, error)
	// GetAllCards 获取所有卡信息
	GetAllCards() ([]Card, error)
	// AddCard 添加新卡信息
	AddCard(card Card) error
	// UpdateCard 修改卡信息
	UpdateCard(card Card) error
	// DeleteCard 删除卡信息
	DeleteCard(cardNo string) error
	// GetCard 获取卡信息
	GetCard(cardNo string) (Card, error)

	// AddUser 添加新用户信息
	AddUser(user CardWithStaffInfo) error
	// DeleteUser 删除卡用户信息
	DeleteUser(user UserID) error

	// GetDoors 获取门信息
	GetDoors() (interface{}, error)

	// SetDoorParameter 设置门参数
	SetDoorParameter(params []DoorParameter) error
	// GetDoorParameter 获取门参数
	GetDoorParameter() ([]DoorParameter, error)

	// GetEvents 获取指定索引的刷卡记录，返回新的偏移、最后一条刷卡记录索引以及本次获取到的刷卡记录
	GetEvents(offset int) (EventData, error)
	GetEventsByTime(timeInterval TimeInterval) (EventData, error)
	GetEventsWhenVerify(offset interface{}) (EventData, error)
	// GetAlarms 获取指定索引的告警记录，返回新的偏移、最后一条告警记录索引以及本次获取到的告警记录
	GetAlarms(offset int) (AlarmData, error)
	GetAlarmsByTime(timeInterval TimeInterval) (AlarmData, error)
	GetAlarmsWhenVerify(offset interface{}) (AlarmData, error)

	GetDoorPositionState() (interface{}, error)

	// Clean 重置门禁控制器，恢复至出厂状态
	Clean() error

	// Reset 消防复位
	Reset() error

	// GetCurrentAlarm 获取当前告警
	GetCurrentAlarm() ([]CurrentAlarmData, error)

	// IsReady 判断Controller是否就绪，只有CACSController有用，其他返回true
	IsReady() bool
}
