// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dac/entity/model/driver"
	"dac/entity/model/driver/xbrother"
	"dac/entity/model/rt"
	"dac/entity/utils"
	"dac/logic/collect/driver/xbrother/consts"

	"dac/entity/utils/ttime"
)

// SetDoorState 设置门状态（开门/关门/常开/常闭）
func (c *Controller) SetDoorState(
	doorStates driver.SetDoorStateRequest,
) error {
	for doorNo, doorStateType := range doorStates {
		if doorNo < 1 || int(doorNo) > c.doorNum {
			c.logger.Errorf(
				"unknown doorNo, doorNo: %d, doorStateType: %d",
				doorNo, doorStateType)
			return fmt.Errorf(
				"unknown doorNo, doorNo: %d, doorStateType: %d",
				doorNo, doorStateType)
		}
		switch doorStateType {
		case driver.StateOpen:
			if err := c.unlockDoor(uint8(doorNo)); err != nil {
				return fmt.Errorf(
					"unlock door error, doorNo: %d, err: %s",
					doorNo, err.Error())
			}
			time.Sleep(consts.DurationSleepTime)
			openReq := xbrother.OpenDoorReq{
				DoorNo: uint8(doorNo),
			}
			if _, err := c.controllerOpenDoor(
				openReq, uint8(doorNo)); err != nil {
				return fmt.Errorf(
					"open door error, doorNo: %d, err: %s",
					doorNo, err.Error())
			}
		case driver.StateClose:
			if err := c.unlockDoor(uint8(doorNo)); err != nil {
				return fmt.Errorf(
					"unlock door error, doorNo: %d, err: %s",
					doorNo, err.Error())
			}
			time.Sleep(consts.DurationSleepTime)
			closeReq := xbrother.CloseDoorReq{
				DoorNo: uint8(doorNo),
			}
			if _, err := c.controllerCloseDoor(
				closeReq, uint8(doorNo)); err != nil {
				return fmt.Errorf(
					"close door error, doorNo: %d, err: %s",
					doorNo, err.Error())
			}
		case driver.StateNormallyOpen:
			if err := c.unlockDoor(uint8(doorNo)); err != nil {
				return fmt.Errorf(
					"unlock door error, doorNo: %d, err: %s",
					doorNo, err.Error())
			}
			time.Sleep(consts.DurationSleepTime)
			openPermReq := xbrother.OpenDoorPermenentlyReq{
				DoorNo: uint8(doorNo),
			}
			if _, err := c.controllerOpenDoorPermenently(
				openPermReq, uint8(doorNo)); err != nil {
				return fmt.Errorf(
					"open door permanently error, doorNo: %d, err: %s",
					doorNo, err.Error())
			}
			time.Sleep(consts.DurationSleepTime)
			if err := c.lockDoor(uint8(doorNo)); err != nil {
				return fmt.Errorf(
					"lock door error, doorNo: %d, err: %s",
					doorNo, err.Error())
			}
		case driver.StateNormallyClose:
			if err := c.unlockDoor(uint8(doorNo)); err != nil {
				return fmt.Errorf(
					"unlock door error, doorNo: %d, err: %s",
					doorNo, err.Error())
			}
			time.Sleep(consts.DurationSleepTime)
			lockReq := xbrother.LockDoorReq{
				LockStatus: consts.LockDoorStateLock,
			}
			if _, err := c.controllerLockDoor(
				lockReq, uint8(doorNo)); err != nil {
				return fmt.Errorf(
					"lock door error, doorNo: %d, err: %s",
					doorNo, err.Error())
			}
			time.Sleep(consts.DurationSleepTime)
			if err := c.lockDoor(uint8(doorNo)); err != nil {
				return fmt.Errorf(
					"lock door error, doorNo: %d, err: %s",
					doorNo, err.Error())
			}
		default:
			c.logger.Warnf(
				"unknown door state type, doorNo: %d, doorStateType: %d",
				doorNo, doorStateType)
			return fmt.Errorf(
				"unknown door state type, doorNo: %d, doorStateType: %d",
				doorNo, doorStateType)
		}
	}
	return nil
}

// unlockDoor 解锁指定门
func (c *Controller) unlockDoor(doorNo uint8) error {
	lockReq := xbrother.LockDoorReq{
		LockStatus: consts.LockDoorStateUnlock,
	}
	if _, err := c.controllerLockDoor(lockReq, doorNo); err != nil {
		return fmt.Errorf(
			"unlock door error, doorNo: %d, err: %s",
			doorNo, err.Error())
	}
	return nil
}

// lockDoor 锁定指定门
func (c *Controller) lockDoor(doorNo uint8) error {
	lockReq := xbrother.LockDoorReq{
		LockStatus: consts.LockDoorStateLock,
	}
	if _, err := c.controllerLockDoor(lockReq, doorNo); err != nil {
		return fmt.Errorf(
			"lock door error, doorNo: %d, err: %s",
			doorNo, err.Error())
	}
	return nil
}

// RawDoorState 原始门状态数据
type RawDoorState struct {
	Door   int `json:"door"`   // 门编号
	Status int `json:"status"` // 门状态
}

// GetRawDoorState 获取原始门状态数据
func (c *Controller) GetRawDoorState(doors []int) (interface{}, error) {
	return c.getRawDoorState(doors)
}

// getRawDoorState 从Redis获取门状态数据，返回指定门的原始状态列表。
// 门编号从1开始，Redis中存储的是所有门的状态数组。
func (c *Controller) getRawDoorState(doors []int) ([]RawDoorState, error) {
	rawDoorStates := make([]RawDoorState, 0)
	redisKey := utils.GenerateRedisKeyDoorStatus(c.chanInfo.ChannelID)
	doorStateBytes, err := c.redisClient.Get(context.Background(), redisKey).Bytes()
	if err != nil {
		return nil, fmt.Errorf("get door state error, redis key: %v, err: %w", redisKey, err)
	}
	var doorState []int
	if err = json.Unmarshal(doorStateBytes, &doorState); err != nil {
		return nil, err
	}
	if len(doorState) != c.doorNum {
		return nil, fmt.Errorf("unexpected doorNum, get %d, but expected %d", len(doorState), c.doorNum)
	}

	for i := range doors {
		doorId := doors[i]
		if doorId < 1 || doorId > c.doorNum {
			return nil, fmt.Errorf("unknown doorId: %d", doorId)
		}
		rawDoorStates = append(rawDoorStates, RawDoorState{
			Door:   doorId,
			Status: doorState[doorId-1],
		})
	}
	return rawDoorStates, nil
}

// GetDoorPositionState 获取门位置状态（XBrother不支持）
func (c *Controller) GetDoorPositionState() (interface{}, error) {
	return nil, nil
}

// GetDoorState 获取门状态测点数据
func (c *Controller) GetDoorState(doors []int) (map[int]*rt.Point, error) {
	rawDoorStates, err := c.getRawDoorState(doors)
	if err != nil {
		return nil, err
	}

	points := make(map[int]*rt.Point, len(doors))
	t := ttime.GetNowUTC()
	for i := range rawDoorStates {
		d := &rawDoorStates[i]
		p := new(rt.Point)
		p.ID = utils.GenerateDoorStateID(c.baseInfo.ID, d.Door)
		p.SetValueWithTime(d.Status, t.UnixMilli())
		points[d.Door] = p
	}
	return points, nil
}

// controllerOpenDoor 发送开门命令
func (c *Controller) controllerOpenDoor(
	req xbrother.OpenDoorReq, doorNo uint8,
) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo, consts.GetRRPCOpenDoor(c.chanInfo.ChannelID), consts.CommandOpenDoor)
}

// controllerOpenDoorPermenently 发送常开门命令
func (c *Controller) controllerOpenDoorPermenently(
	req xbrother.OpenDoorPermenentlyReq, doorNo uint8,
) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo,
		consts.GetRRPCDoorOpenPermanently(c.chanInfo.ChannelID),
		consts.CommandDoorOpenPermenently)
}

// controllerCloseDoor 发送关门命令
func (c *Controller) controllerCloseDoor(
	req xbrother.CloseDoorReq, doorNo uint8,
) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo, consts.GetRRPCCloseDoor(c.chanInfo.ChannelID), consts.CommandCloseDoor)
}

// controllerLockDoor 发送锁门命令
func (c *Controller) controllerLockDoor(
	req xbrother.LockDoorReq, doorNo uint8,
) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo, consts.GetRRPCLockDoor(c.chanInfo.ChannelID), consts.CommandLockDoor)
}
