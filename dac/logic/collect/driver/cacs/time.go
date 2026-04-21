// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"dac/entity/model/driver/cacs"
	"dac/logic/collect/driver/cacs/consts"
	"fmt"
	"time"

	"dac/entity/utils/rrpc"
)

// GetTime 获取门控器时间（同步本地时间到门控器后返回时间字符串）
func (c *Controller) GetTime() (string, error) {
	now, err := c.doSetTime()
	if err != nil {
		return "", err
	}
	return now.String(), nil
}

// SetTime 设置门控器时间（将本地时间同步到门控器）
func (c *Controller) SetTime() error {
	_, err := c.doSetTime()
	return err
}

// doSetTime 执行设置时间的公共逻辑，返回设置的时间和错误
func (c *Controller) doSetTime() (time.Time, error) {
	if _, err := c.checkConnection(); err != nil {
		return time.Time{}, err
	}
	now := time.Now()
	req := cacs.SetTimeReq{
		Year:   uint16(now.Year()),
		Month:  uint8(now.Month()),
		Day:    uint8(now.Day()),
		Hour:   uint8(now.Hour()),
		Minute: uint8(now.Minute()),
		Second: uint8(now.Second()),
	}
	_, ok, packetRtn, _, err := c.setTime(req)
	if !ok {
		return time.Time{}, fmt.Errorf("set time failed, err: %s", err.Error())
	}
	if packetRtn != consts.KRtnNormal {
		return time.Time{}, fmt.Errorf(consts.RtnInfoMap[packetRtn])
	}
	return now, nil
}

// setTime 发送设置时间的协议请求并解析响应
func (c *Controller) setTime(req cacs.SetTimeReq) (cacs.SetTimeResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.SetTimeResp{}, false, 0, consts.KRequestError, err
	}

	cmd := consts.KCommandRequestSetTime
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.SetTimeResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf(consts.RequestInfoMap[consts.KMarshalError])
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.SetTimeResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf(consts.RequestInfoMap[consts.KRequestError])
	}
	rrpcKey := consts.GetRRPCSetTime(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.SetTimeResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf(consts.RequestInfoMap[consts.KRecvRespError])
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.SetTimeResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf(consts.RequestInfoMap[consts.KUnMarshalError])
	}
	resp, err := c.tcpMarshal.Unmarshal(
		consts.KCommandResponseSetTime, bytes)
	if err != nil {
		c.Errorf("resp tcpUnmarshal to SetTimeResp failed, err: %v", err)
		return cacs.SetTimeResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf(consts.RequestInfoMap[consts.KUnMarshalError])
	}
	setTimeResp, ok := resp.(cacs.SetTimeResp)
	if !ok {
		c.Errorf("resp type error, it should be SetTimeResp")
		return cacs.SetTimeResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf(consts.RequestInfoMap[consts.KUnMarshalError])
	}
	return setTimeResp, true, server.p.rtn, consts.KNormal, nil
}
