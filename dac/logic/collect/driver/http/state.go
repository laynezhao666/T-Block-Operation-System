// Package http 实现HTTP门禁控制器协议的驱动层。
package http

import (
	"encoding/json"
	"fmt"

	"dac/entity/model/driver"
	"dac/entity/model/rt"
	"dac/entity/utils"
	"dac/entity/utils/dhttp"

	"dac/entity/utils/ttime"
)

// OpenMethodPlatform 平台开门方式标识
const (
	OpenMethodPlatform = 0
)

// DoorState 门状态数据
type DoorState struct {
	Door   int `json:"door"`   // 门编号
	State  int `json:"state"`  // 门状态
	Status int `json:"status"` // 门状态（兼容字段）
}

// DoorStateSetRequest 设置门状态请求
type DoorStateSetRequest struct {
	Door   int `json:"door"`        // 门编号
	State  int `json:"state"`       // 目标状态
	Method int `json:"open_method"` // 开门方式
}

// getRawDoorStateV1 获取V1版本的原始门状态数据
func (c *Controller) getRawDoorStateV1(doors []int) (map[string]int, error) {
	var states map[string]int
	url := c.urlProducer.GetDoorStateURL()

	var e error
	err := dhttp.GetJSONWithParseFunc(url, c.timeout, func(b []byte) error {
		if e = json.Unmarshal(b, &states); e == nil {
			return nil
		}
		c.logger.Warnf("unmarsh state %v error: %v, try parse as float", string(b), e)
		var tempStats map[string]float64
		if e = json.Unmarshal(b, &tempStats); e == nil {
			states = make(map[string]int, len(tempStats))
			for k, v := range tempStats {
				states[k] = int(v)
			}
			return nil
		}
		return e
	})
	return states, err
}

// getDoorStateV1 获取V1版本的门状态测点数据
func (c *Controller) getDoorStateV1(doors []int) (map[int]*rt.Point, error) {
	states, err := c.getRawDoorStateV1(doors)
	if err != nil {
		return nil, err
	}

	points := make(map[int]*rt.Point, len(doors))
	t := ttime.GetNowUTC()
	for _, doorNumber := range doors {
		k := fmt.Sprintf("door%v", doorNumber)

		v, ok := states[k]
		if !ok {
			continue
		}

		p := new(rt.Point)
		p.ID = utils.GenerateDoorStateID(c.baseInfo.ID, doorNumber)
		p.SetValueWithTime(v, t.UnixMilli())

		points[doorNumber] = p
	}

	return points, nil
}

// getRawDoorStateV2 获取V2版本的原始门状态数据
func (c *Controller) getRawDoorStateV2() ([]DoorState, error) {
	var states []DoorState
	url := c.urlProducer.GetDoorStateURL()

	err := dhttp.GetJSON(url, c.timeout, &states)
	return states, err
}

// getDoorStateV2 获取V2版本的门状态测点数据
func (c *Controller) getDoorStateV2() (map[int]*rt.Point, error) {
	states, err := c.getRawDoorStateV2()
	if err != nil {
		return nil, err
	}

	t := ttime.GetNowUTC()
	points := make(map[int]*rt.Point, len(states))
	for i := range states {
		d := &states[i]

		p := new(rt.Point)
		p.ID = utils.GenerateDoorStateID(c.baseInfo.ID, d.Door)
		p.SetValueWithTime(d.State, t.UnixMilli())

		points[d.Door] = p
	}

	return points, nil
}

// GetDoorState 获取门状态测点数据（自动选择V1或V2版本）
func (c *Controller) GetDoorState(doors []int) (map[int]*rt.Point, error) {
	if c.isVersion1 || c.isVersionMDC {
		return c.getDoorStateV1(doors)
	}

	return c.getDoorStateV2()
}

// GetRawDoorState 获取原始门状态数据（自动选择V1或V2版本）
func (c *Controller) GetRawDoorState(doors []int) (interface{}, error) {
	if c.isVersion1 || c.isVersionMDC {
		return c.getRawDoorStateV1(doors)
	}

	return c.getRawDoorStateV2()
}

// setDoorStateV1 设置V1版本的门状态
func (c *Controller) setDoorStateV1(doorStates driver.SetDoorStateRequest) error {
	req := make(map[string]int, len(doorStates)+1)
	for door, state := range doorStates {
		req[fmt.Sprintf("door%v", door)] = int(state)
	}
	req["open_method"] = OpenMethodPlatform
	return c.postJSON(c.urlProducer.SetDoorStateURL(), req, nil)
}

// setDoorStateV2 设置V2版本的门状态
func (c *Controller) setDoorStateV2(doorStates driver.SetDoorStateRequest) error {
	req := make([]DoorStateSetRequest, 0, len(doorStates))
	for d, s := range doorStates {
		req = append(req, DoorStateSetRequest{
			Door:   int(d),
			State:  int(s),
			Method: OpenMethodPlatform,
		})
	}

	return c.postJSON(c.urlProducer.SetDoorStateURL(), req, nil)
}

// SetDoorState 设置门状态（自动选择V1或V2版本）
func (c *Controller) SetDoorState(doorStates driver.SetDoorStateRequest) error {
	if c.isVersion1 || c.isVersionMDC {
		return c.setDoorStateV1(doorStates)
	}

	return c.setDoorStateV2(doorStates)
}

// GetDoorPositionState 获取门位置状态
func (c *Controller) GetDoorPositionState() (interface{}, error) {
	var temp interface{}
	err := dhttp.GetJSON(c.urlProducer.GetDoorPositionStateURL(), c.timeout, &temp)
	return temp, err
}
