package utils

import (
	"encoding/json"
	"fmt"

	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/entity/model/driver"
)

// ProcessDBDoor 处理数据库门对象，清空敏感信息并调整门名称
func ProcessDBDoor(d *db.Door) {
	if d == nil {
		return
	}
	// 清空密码
	d.Parameters.Password = ""

	// 调整门名称
	d.SetName(d.GetName())
}

// GetCGIDoor 将数据库门对象转换为CGI门对象，附带状态ID
func GetCGIDoor(controllerID db.IDType, dbDoor db.Door) cgi.Door {
	cgiDoor := cgi.Door{
		Door:    dbDoor,
		StateID: GenerateDoorStateID(controllerID, dbDoor.Number),
	}

	ProcessDBDoor(&cgiDoor.Door)

	return cgiDoor
}

// IntMapToSlice 将int类型的map键集合转换为切片
func IntMapToSlice(m map[int]struct{}) []int {
	n := make([]int, 0, len(m))
	for x := range m {
		n = append(n, x)
	}
	return n
}

// IntSliceToMap 将int切片转换为map集合（用于去重和快速查找）
func IntSliceToMap(n []int) map[int]struct{} {
	m := make(map[int]struct{}, len(n))
	for _, x := range n {
		m[x] = struct{}{}
	}
	return m
}

// GetDoorsMap 将门列表转换为以门ID为key的map
func GetDoorsMap(doors []db.Door) map[db.IDType]db.Door {
	doorMap := make(map[db.IDType]db.Door, len(doors))
	for i := range doors {
		doorMap[doors[i].ID] = doors[i]
	}
	return doorMap
}

// timezoneLess 比较两个时区字符串的大小（格式 "HH:MM"）
func timezoneLess(lhs, rhs string) bool {
	if lhs[0:2] < rhs[0:2] {
		return true
	}
	return lhs[3:5] < rhs[3:5]
}

// ConvertTimeGroupDriverToDB 将驱动时间组转换为数据库时间组模型，包含时区格式校验
func ConvertTimeGroupDriverToDB(tg driver.TimeGroup) (db.TimeGroup, error) {
	var (
		timeGroup db.TimeGroup
		err       error
		b         []byte
	)
	timeGroup.GroupNo = tg.GroupNo

	for i := range tg.TimeZone {
		t := &tg.TimeZone[i]
		if len(t.Begin) != 5 || len(t.End) != 5 {
			return timeGroup, fmt.Errorf("时间段 %+v 设置错误", *t)
		}
		if !(timezoneLess(t.Begin, t.End)) {
			return timeGroup, fmt.Errorf("时间段 %+v 设置错误", tg.TimeZone[i])
		}
	}

	if b, err = json.Marshal(tg.Week); err != nil {
		return timeGroup, err
	}
	timeGroup.Week = string(b)

	if b, err = json.Marshal(tg.TimeZone); err != nil {
		return timeGroup, err
	}
	timeGroup.TimeZone = string(b)

	return timeGroup, nil
}

// ConvertTimeGroupDBToDriver 将数据库时间组转换为驱动时间组模型
func ConvertTimeGroupDBToDriver(tg db.TimeGroup) (driver.TimeGroup, error) {
	var (
		timeGroup driver.TimeGroup
		err       error
	)
	timeGroup.GroupNo = tg.GroupNo

	if err = json.Unmarshal([]byte(tg.Week), &timeGroup.Week); err != nil {
		return timeGroup, err
	}

	if err = json.Unmarshal([]byte(tg.TimeZone), &timeGroup.TimeZone); err != nil {
		return timeGroup, err
	}

	return timeGroup, nil
}

// ConvertDBDoorParameter 将驱动门参数转换为数据库门参数模型（不含控制器和通道信息）
func ConvertDBDoorParameter(p *driver.DoorParameter) db.DoorParameter {
	return db.DoorParameter{
		Name:           p.Name,
		Password:       p.Password,
		KeepOpenTime:   p.KeepOpenTime,
		OpenTimeout:    p.OpenTimeout,
		LockCount:      p.LockCount,
		LockTime:       p.LockTime,
		VerifyInterval: p.VerifyInterval,
		OpenMode:       int(p.OpenMode),
		FireSignalMode: p.FireSignalMode,
	}
}

// GetDoorNameMap 将门列表转换为以门号为key、门名称为value的map
func GetDoorNameMap(doors []db.Door) map[int]string {
	doorNameMap := make(map[int]string, len(doors))
	for i := range doors {
		d := &doors[i]
		doorNameMap[d.Number] = d.GetName()
	}
	return doorNameMap
}

// GetDoorsBaseInfo 从门列表中提取基础信息（ID和名称）
func GetDoorsBaseInfo(doors []db.Door) []db.DoorBaseInfo {
	r := make([]db.DoorBaseInfo, 0, len(doors))
	for i := range doors {
		d := &doors[i]
		r = append(r, db.DoorBaseInfo{
			ID:   d.ID,
			Name: d.GetName(),
		})
	}
	return r
}

// GetJSONString 将任意对象序列化为JSON字符串，序列化失败时返回空字符串
func GetJSONString(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// ConvertDBDriverDoorParams 将数据库门参数列表转换为驱动门参数列表
func ConvertDBDriverDoorParams(res []db.DriverDoorParameter) []driver.DoorParameter {
	doorParams := make([]driver.DoorParameter, len(res))
	for i := range res {
		dbDoorParam := &res[i]
		doorParams[i] = driver.DoorParameter{
			Number:         driver.DoorNumberType(dbDoorParam.Number),
			Name:           dbDoorParam.Name,
			Password:       dbDoorParam.Password,
			KeepOpenTime:   dbDoorParam.KeepOpenTime,
			OpenTimeout:    dbDoorParam.OpenTimeout,
			LockCount:      dbDoorParam.LockCount,
			LockTime:       dbDoorParam.LockTime,
			VerifyInterval: dbDoorParam.VerifyInterval,
			OpenMode:       driver.OpenModeType(dbDoorParam.OpenMode),
			FireSignalMode: dbDoorParam.FireSignalMode,
		}
	}
	return doorParams
}

// ConvertDriverDoorParamToDB 将驱动门参数转换为数据库模型
func ConvertDriverDoorParamToDB(controllerID db.IDType,
	channelID string, param driver.DoorParameter,
) db.DriverDoorParameter {
	return db.DriverDoorParameter{
		ControllerID:   controllerID,
		ChannelID:      channelID,
		Number:         db.DoorNumberType(param.Number),
		Name:           param.Name,
		Password:       param.Password,
		KeepOpenTime:   param.KeepOpenTime,
		OpenTimeout:    param.OpenTimeout,
		LockCount:      param.LockCount,
		LockTime:       param.LockTime,
		VerifyInterval: param.VerifyInterval,
		OpenMode:       db.OpenModeType(param.OpenMode),
		FireSignalMode: param.FireSignalMode,
	}
}

// DriverTimeZone2DBTimeZone 将驱动时区列表转换为数据库时区列表
func DriverTimeZone2DBTimeZone(driverTimeZone []driver.TimeZone) []db.TimeZone {
	res := make([]db.TimeZone, len(driverTimeZone))
	for i := 0; i < len(driverTimeZone); i++ {
		res[i].Begin = driverTimeZone[i].Begin
		res[i].End = driverTimeZone[i].End
	}
	return res
}

// DBTimeZone2DriverTimeZone 将数据库时区列表转换为驱动时区列表
func DBTimeZone2DriverTimeZone(dbTimeZone []db.TimeZone) []driver.TimeZone {
	res := make([]driver.TimeZone, len(dbTimeZone))
	for i := 0; i < len(dbTimeZone); i++ {
		res[i].Begin = dbTimeZone[i].Begin
		res[i].End = dbTimeZone[i].End
	}
	return res
}

// ConvertDBEventsToDriver 将数据库事件列表转换为驱动事件列表
func ConvertDBEventsToDriver(driverEvents []db.DriverEvent) []driver.Event {
	events := make([]driver.Event, len(driverEvents))
	for i := range driverEvents {
		event := &driverEvents[i]
		events[i] = driver.Event{
			Index:       event.Index,
			Timestamp:   event.Timestamp,
			CardNumber:  event.CardNumber,
			Username:    event.Username,
			DoorNumber:  driver.DoorNumberType(event.DoorNumber),
			Direction:   driver.DirectionType(event.Direction),
			Type:        driver.EventType(event.Type),
			Description: event.Description,
		}
	}
	return events
}
