// Package test 提供门禁控制器的测试驱动实现，用于模拟门禁设备行为。
package test

import (
	"fmt"
	"math/rand"
	"time"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/driver"
	"dac/entity/model/rt"
	"dac/entity/utils"

	"dac/entity/utils/ttime"
)

// 测试驱动常量
const (
	maxRecord = 1000 // 最大记录数

	doorNumber = 4 // 模拟门数量
)

// testController 测试用门禁控制器，模拟真实设备的所有接口
type testController struct {
	info driver.ControllerBasicInfo // 控制器基本信息
}

// IsReady 返回控制器是否就绪（测试驱动始终返回true）
func (h *testController) IsReady() bool {
	return true
}

// Open 打开控制器连接（测试驱动直接返回OK）
func (h *testController) Open(chanInfo driver.ChannelInfo) consts.Quality {
	return consts.QualityOK
}

// Close 关闭控制器连接
func (h *testController) Close() consts.Quality {
	return consts.QualityOK
}

// Ping 测试控制器连通性
func (h *testController) Ping() error {
	return nil
}

// GetDoorState 获取门状态，随机生成0/1状态值
func (h *testController) GetDoorState(numbers []int) (map[int]*rt.Point, error) {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)+50))

	t := ttime.GetNowUTC()

	points := make(map[int]*rt.Point, len(numbers))
	for _, num := range numbers {
		points[num] = &rt.Point{
			ID:  utils.GenerateDoorStateID(h.info.ID, num),
			Rtd: rt.NewRTValueWithPvTime(fmt.Sprintf("%v", rand.Int()&1), t),
		}
	}

	return points, nil
}

// GetDoorPoints 获取门的所有测点数据（状态+告警）
func (h *testController) GetDoorPoints(doors []int) (map[string]map[int]*rt.Point, error) {
	states, err := h.GetDoorState(doors)
	if err != nil {
		return nil, err
	}

	alarmData, err := h.GetCurrentAlarm()
	if err != nil {
		return nil, err
	}

	timestamp := ttime.GetNowUTC().UnixMilli()
	alarmPoints := make(map[int]*rt.Point, len(doors))
	for i := range alarmData {
		d := alarmData[i].Door
		alarms := alarmData[i].Alarms
		hasOpenAlarm := false
		for j := range alarms {
			if hasOpenAlarm {
				break
			}
			switch alarms[j].Type {
			case 0:
				p := new(rt.Point)
				p.ID = utils.GenerateDoorOpenAlarmID(h.info.ID, d)
				p.SetValueWithTime(1, timestamp)

				alarmPoints[d] = p
				hasOpenAlarm = true
			}
		}
	}
	for _, d := range doors {
		if _, ok := alarmPoints[d]; ok {
			continue
		}

		p := new(rt.Point)
		p.ID = utils.GenerateDoorOpenAlarmID(h.info.ID, d)
		p.SetValueWithTime(0, timestamp)

		alarmPoints[d] = p
	}

	data := map[string]map[int]*rt.Point{
		consts.StandardIDDoorState: states,
		consts.StandardIDOpenAlarm: alarmPoints,
	}

	return data, nil
}

// GetRawDoorState 获取原始门状态（测试驱动返回nil）
func (h *testController) GetRawDoorState([]int) (interface{}, error) {
	return nil, nil
}

// SetDoorState 设置门状态（测试驱动空实现）
func (h *testController) SetDoorState(doorStates driver.SetDoorStateRequest) error {
	return nil
}

// GetTimeGroup 获取时间组（测试驱动返回空时间组）
func (h *testController) GetTimeGroup(int) (driver.TimeGroup, error) {
	return driver.TimeGroup{}, nil
}

// SetTimeGroup 设置时间组（测试驱动空实现）
func (h *testController) SetTimeGroup(timeGroup driver.TimeGroup) error {
	return nil
}

// ClearTimeGroup 清除时间组
func (h *testController) ClearTimeGroup(timeGroup int) error {
	config.Log.Infof("delete time group %v in controller %+v", timeGroup, h.info)
	return nil
}

// AddCard 添加门禁卡
func (h *testController) AddCard(card driver.Card) error {
	config.Log.Infof("add card %+v in controller %+v", card, h.info)
	return nil
}

// UpdateCard 更新门禁卡（测试驱动空实现）
func (h *testController) UpdateCard(card driver.Card) error {
	return nil
}

// DeleteCard 删除门禁卡（测试驱动空实现）
func (h *testController) DeleteCard(cardNo string) error {
	return nil
}

// GetCard 获取门禁卡信息（测试驱动返回空卡）
func (h *testController) GetCard(cardNo string) (driver.Card, error) {
	return driver.Card{}, nil
}

// AddUser 添加用户（测试驱动空实现）
func (h *testController) AddUser(user driver.CardWithStaffInfo) error {
	return nil
}

// DeleteUser 删除用户（测试驱动空实现）
func (h *testController) DeleteUser(user driver.UserID) error {
	return nil
}

// SetDoorParameter 设置门参数（测试驱动空实现）
func (h *testController) SetDoorParameter(params []driver.DoorParameter) error {
	return nil
}

// GetDoorParameter 获取门参数，返回模拟的4门参数
func (h *testController) GetDoorParameter() ([]driver.DoorParameter, error) {
	n := doorNumber
	doors := make([]driver.DoorParameter, 0, n)
	for i := 0; i < n; i++ {
		doors = append(doors, driver.DoorParameter{
			Number:         driver.DoorNumberType(i + 1),
			Name:           fmt.Sprintf("门%v", i+1),
			Password:       "0000",
			KeepOpenTime:   1,
			OpenTimeout:    1,
			LockCount:      3,
			LockTime:       3,
			VerifyInterval: 5,
			OpenMode:       0,
			FireSignalMode: 0,
		})
	}
	return doors, nil
}

// GetEventsWhenVerify 校验模式下获取事件
func (h *testController) GetEventsWhenVerify(offset interface{}) (driver.EventData, error) {
	if o, ok := offset.(int); ok {
		return h.GetEvents(o)
	}

	return driver.EventData{}, fmt.Errorf("invalid offset: %v, type: %T", offset, offset)
}

// GetEventsByTime 按时间范围获取事件
func (h *testController) GetEventsByTime(timeInterval driver.TimeInterval) (driver.EventData, error) {
	time.Sleep(time.Second * 1)
	var data driver.EventData
	data.EndTimestamp = timeInterval.BeginTimestamp + 30
	return data, nil
}

// GetEvents 获取事件列表，20%概率随机生成新事件
func (h *testController) GetEvents(offset int) (driver.EventData, error) {
	n := 0
	if (offset < maxRecord && rand.Float32() < 0.2) || isDebug() {
		// 20% 概率返回新数据
		n = 1
	}
	events := make([]driver.Event, 0, n)
	for i := 1; i <= n; i++ {
		events = append(events, driver.Event{
			Index:       offset + i,
			Timestamp:   ttime.GetNowUTC().Unix(),
			CardNumber:  "test",
			Username:    "program",
			DoorNumber:  driver.DoorNumberType(rand.Intn(doorNumber) + 1),
			Direction:   driver.DirectionType(rand.Int() & 1),
			Type:        driver.EventType(rand.Intn(5)),
			Description: fmt.Sprintf("%v", rand.Uint64()),
		})
	}
	return driver.EventData{
		Offset: offset + n,
		Last:   offset + n,
		Events: events,
	}, nil
}

// GetAlarmsWhenVerify 校验模式下获取告警
func (h *testController) GetAlarmsWhenVerify(offset interface{}) (driver.AlarmData, error) {
	if o, ok := offset.(int); ok {
		return h.GetAlarms(o)
	}

	return driver.AlarmData{}, fmt.Errorf("invalid offset: %v, type: %T", offset, offset)
}

// GetAlarms 获取告警列表，20%概率随机生成新告警
func (h *testController) GetAlarms(offset int) (driver.AlarmData, error) {
	n := 0
	if (offset < maxRecord && rand.Float32() < 0.2) || isDebug() {
		// 20% 概率返回新数据
		n = 1
	}
	alarms := make([]driver.Alarm, 0, n)
	for i := 1; i <= n; i++ {
		alarms = append(alarms, driver.Alarm{
			Index:       offset + i,
			Timestamp:   ttime.GetNowUTC().Unix(),
			DoorNumber:  driver.DoorNumberType(rand.Intn(doorNumber + 1)),
			Type:        driver.AlarmType(rand.Intn(5)),
			State:       driver.AlarmStateType(rand.Int() & 1),
			Description: fmt.Sprintf("%v", rand.Uint64()),
		})
	}
	return driver.AlarmData{
		Offset: offset + n,
		Last:   offset + n,
		Alarms: alarms,
	}, nil
}

// GetAlarmsByTime 按时间范围获取告警
func (h *testController) GetAlarmsByTime(timeInterval driver.TimeInterval) (driver.AlarmData, error) {
	time.Sleep(time.Second * 1)
	var data driver.AlarmData
	data.EndTimestamp = timeInterval.BeginTimestamp + 30
	return data, nil
}

// Clean 清空控制器数据
func (h *testController) Clean() error {
	config.Log.Infof("clean controller: %+v", h.info)
	return nil
}

// Reset 重置控制器
func (h *testController) Reset() error {
	config.Log.Infof("reset controller: %+v", h.info)
	return nil
}

// SetTime 同步控制器时间
func (h *testController) SetTime() error {
	config.Log.Infof("set time, controller: %+v", h.info)
	return nil
}

// GetAllCards 获取所有卡（测试驱动返回nil）
func (h *testController) GetAllCards() ([]driver.Card, error) {
	return nil, nil
}

// GetCards 分页获取卡列表（测试驱动返回空）
func (h *testController) GetCards(offset int) (driver.CardData, error) {
	return driver.CardData{}, nil
}

// GetDoors 获取门列表（测试驱动返回nil）
func (h *testController) GetDoors() (interface{}, error) {
	return nil, nil
}

// GetDoorPositionState 获取门位置状态（测试驱动返回nil）
func (h *testController) GetDoorPositionState() (interface{}, error) {
	return nil, nil
}

// GetTime 获取控制器时间（测试驱动返回空）
func (h *testController) GetTime() (string, error) {
	return "", nil
}

// GetCurrentAlarm 获取当前告警，随机生成门超时告警
func (h *testController) GetCurrentAlarm() ([]driver.CurrentAlarmData, error) {
	n := doorNumber
	alarms := make([]driver.CurrentAlarmData, 0, n)
	for i := 0; i < n; i++ {
		if (rand.Int() & 1) == 0 {
			continue
		}

		alarms = append(alarms, driver.CurrentAlarmData{
			Door: i + 1,
			Alarms: []driver.CurrentAlarmEvent{
				{
					Type: 0,
					Desc: "模拟告警",
				},
			},
		})
	}

	return alarms, nil
}
