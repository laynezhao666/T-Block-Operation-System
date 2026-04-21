// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"dac/entity/model/driver"
	"dac/entity/model/driver/cacs"
	"dac/entity/model/rt"
	"dac/entity/utils"
	"dac/logic/collect/driver/cacs/consts"
	"fmt"

	"dac/entity/utils/rrpc"
	"dac/entity/utils/ttime"
)

// mapStateToControlMode 将前端 DoorStateType 映射到 CACS 协议的 ControlMode。
// 前端定义: 0=关门(恢复自动), 1=开门一次, 2=常开, 3=常闭
// CACS协议: 0=开门一次, 1=常开, 2=常闭, 3=恢复自动
func mapStateToControlMode(state driver.DoorStateType) uint8 {
	switch state {
	case driver.StateClose: // 0 -> 恢复自动控制
		return RemoteControlOpenDoorAuto // 3
	case driver.StateOpen: // 1 -> 开门一次
		return RemoteControlOpenDoorOnce // 0
	case driver.StateNormallyOpen: // 2 -> 常开
		return RemoteControlOpenDoorAlways // 1
	case driver.StateNormallyClose: // 3 -> 常闭
		return RemoteControlCloseDoorAlways // 2
	default:
		return RemoteControlOpenDoorAuto // 默认恢复自动控制
	}
}

// GetDoorState 获取指定门的状态点位信息。
func (c *Controller) GetDoorState(
	doors []int,
) (map[int]*rt.Point, error) {
	if _, err := c.checkConnection(); err != nil {
		return nil, err
	}
	r, err := c.GetDoors()
	if err != nil {
		c.Errorf("Get doors failed: %s", err)
		return nil, err
	}
	getDoorInfos, ok := r.([]getDoorInfo)
	if !ok {
		c.Errorf("GetDoors response type error")
		return nil, fmt.Errorf("GetDoors response type error")
	}
	points := make(map[int]*rt.Point)
	doorsMap := make(map[uint32]struct{})
	for i := range doors {
		doorsMap[uint32(doors[i])] = struct{}{}
	}

	t := ttime.GetNowUTC()
	for i := range getDoorInfos {
		if _, ok := doorsMap[getDoorInfos[i].DoorNo]; !ok {
			continue
		}
		resp, ok, packetRtn, _, err := c.getDoorState(
			cacs.DoorStateReq{Id: getDoorInfos[i].DoorNo})
		if !ok {
			return points, fmt.Errorf(
				"get door state failed: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			return points, fmt.Errorf(consts.RtnInfoMap[packetRtn])
		}
		p := new(rt.Point)
		p.ID = utils.GenerateDoorStateID(
			c.baseInfo.ID, int(resp.Id))
		p.SetValueWithTime(resp.DoorSensorStatus, t.UnixMilli())
		points[int(resp.Id)] = p
	}
	return points, nil
}

// GetRawDoorState 获取门的原始状态数据。
func (c *Controller) GetRawDoorState(
	doors []int,
) (interface{}, error) {
	if _, err := c.checkConnection(); err != nil {
		return nil, err
	}
	r, err := c.GetDoors()
	if err != nil {
		c.Errorf("Get doors failed: %s", err)
		return nil, err
	}
	getDoorInfos, ok := r.([]getDoorInfo)
	if !ok {
		c.Errorf("GetDoors response type error")
		return nil, fmt.Errorf("GetDoors response type error")
	}
	var res []cacs.DoorStateResp
	doorsMap := make(map[uint32]struct{})
	for i := range doors {
		doorsMap[uint32(doors[i])] = struct{}{}
	}
	for i := range getDoorInfos {
		if _, ok := doorsMap[getDoorInfos[i].DoorNo]; !ok {
			continue
		}
		resp, ok, packetRtn, _, err := c.getDoorState(
			cacs.DoorStateReq{Id: getDoorInfos[i].DoorNo})
		if !ok {
			return res, fmt.Errorf(
				"get door state failed: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			return res, fmt.Errorf(consts.RtnInfoMap[packetRtn])
		}
		res = append(res, resp)
	}
	return res, nil
}

// SetDoorState 设置门的状态（开门/关门/常开/常闭）。
func (c *Controller) SetDoorState(
	doorStates driver.SetDoorStateRequest,
) error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	for k, v := range doorStates {
		_, ok, packetRtn, _, err := c.RemoteControl(cacs.DoorControlReq{
			ControlMode: mapStateToControlMode(v),
			Id:          uint32(k),
		})
		if !ok {
			return fmt.Errorf("set door state failed: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			return fmt.Errorf(consts.RtnInfoMap[packetRtn])
		}
	}
	return nil
}

// doorPositionState 门位置状态
type doorPositionState struct {
	Door  uint32 `json:"door"`
	State uint8  `json:"state"`
}

// GetDoorPositionState 获取门的位置状态（按钮状态）。
func (c *Controller) GetDoorPositionState() (interface{}, error) {
	if _, err := c.checkConnection(); err != nil {
		return nil, err
	}
	r, err := c.GetDoors()
	if err != nil {
		c.Errorf("Get doors failed: %s", err)
		return nil, err
	}
	getDoorInfos, ok := r.([]getDoorInfo)
	if !ok {
		c.Errorf("GetDoors response type error")
		return nil, fmt.Errorf("GetDoors response type error")
	}
	res := make([]doorPositionState, 0)
	for i := range getDoorInfos {
		resp, ok, packetRtn, _, err := c.getDoorParams(
			cacs.GetDoorParamsReq{Id: getDoorInfos[i].DoorNo})
		if !ok {
			return nil, fmt.Errorf(
				"get door position state failed: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			return nil, fmt.Errorf(consts.RtnInfoMap[packetRtn])
		}
		res = append(res, doorPositionState{
			Door:  getDoorInfos[i].DoorNo,
			State: resp.ButtonStatus,
		})
	}
	return res, nil
}

// getDoorState 查询门状态的底层通信方法。
func (c *Controller) getDoorState(
	req cacs.DoorStateReq,
) (cacs.DoorStateResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.DoorStateResp{}, false, 0,
			consts.KRequestError, err
	}

	cmd := consts.KCommandRequestDoorStatus
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.DoorStateResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf("req marshal failed, err: %v", err)
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.DoorStateResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf("req data send failed, err: %v", err)
	}

	rrpcKey := consts.GetRRPCDoorStatus(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.DoorStateResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("rrpc get resp timeout")
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.DoorStateResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	// 检查RTN，非0则直接返回
	if server.p.rtn != consts.KRtnNormal {
		return cacs.DoorStateResp{}, true, server.p.rtn,
			consts.KRecvRespError, nil
	}
	resp, err := c.tcpMarshal.Unmarshal(
		consts.KCommandResponseDoorStatus, bytes)
	if err != nil {
		c.Errorf("resp unmarshal failed, err: %v", err)
		return cacs.DoorStateResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp unmarshal failed, err: %v", err)
	}
	doorStatusResp, ok := resp.(cacs.DoorStateResp)
	if !ok {
		c.Errorf("resp type error, expect DoorStateResp")
		return cacs.DoorStateResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp type error, expect DoorStateResp")
	}
	return doorStatusResp, true, server.p.rtn,
		consts.KNormal, nil
}
