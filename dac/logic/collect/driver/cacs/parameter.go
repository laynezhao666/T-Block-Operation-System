// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"dac/entity/model/driver"
	"dac/entity/model/driver/cacs"
	"dac/logic/collect/driver/cacs/consts"
	"fmt"
	"strconv"

	"dac/entity/utils/rrpc"
)

// 门参数默认值常量
var (
	DefaultAreaNum   uint16 = 0xff // 默认区域编号
	InterLockInvalid uint32 = 0    // 互锁无效值
)

// recoverDoorParameter 尝试恢复门参数，如果恢复失败不做处理。
func (c *Controller) recoverDoorParameter(
	doorParameterMap map[uint32]cacs.DownloadDoorParamsReq,
) {
	for _, doorParameter := range doorParameterMap {
		c.downloadDoorParams(doorParameter)
	}
}

// SetDoorParameter 设置门参数到门控器。
// 设置前先备份当前参数，失败时自动恢复。
func (c *Controller) SetDoorParameter(
	params []driver.DoorParameter,
) error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	// 保存修改前的doorParameter
	doorParameterMap := make(map[uint32]cacs.DownloadDoorParamsReq)
	doorParameters, err := c.GetDoorParameter()
	if err == nil {
		// 备份当前门参数，用于失败时恢复
		for i := range doorParameters {
			doorParams := doorParameters[i]
			password, _ := strconv.Atoi(doorParams.Password)
			doorParameterMap[uint32(doorParams.Number)] =
				cacs.DownloadDoorParamsReq{
					Id:                    uint32(doorParams.Number),
					DoorMode:              uint8(doorParams.OpenMode),
					MultiCardsOpenDoorNum: 1,
					EntryAreaNum:          DefaultAreaNum,
					ExitAreaNum:           DefaultAreaNum,
					OpenDoorKeepTime:      uint32(doorParams.KeepOpenTime),
					OpenDoorTimeoutTime:   uint32(doorParams.OpenTimeout),
					Password:              uint32(password),
					InterLockId1:          InterLockInvalid,
					InterLockId2:          InterLockInvalid,
					InterLockId3:          InterLockInvalid,
					CardPasswordInterval:  uint32(doorParams.VerifyInterval),
				}
		}
	}
	// 遍历新参数，逐个下发到门控器
	for i := range params {
		doorParameter := params[i]
		// CACS协议中密码长度最大为8位
		if len(doorParameter.Password) > 8 {
			return fmt.Errorf("密码长度不能超过8位")
		}
		password, err := strconv.Atoi(doorParameter.Password)
		if err != nil {
			return err
		}

		req := cacs.DownloadDoorParamsReq{
			Id:                    uint32(doorParameter.Number),
			DoorMode:              uint8(doorParameter.OpenMode),
			MultiCardsOpenDoorNum: 1,
			EntryAreaNum:          DefaultAreaNum,
			ExitAreaNum:           DefaultAreaNum,
			OpenDoorKeepTime:      uint32(doorParameter.KeepOpenTime),
			OpenDoorTimeoutTime:   uint32(doorParameter.OpenTimeout),
			Password:              uint32(password),
			InterLockId1:          InterLockInvalid,
			InterLockId2:          InterLockInvalid,
			InterLockId3:          InterLockInvalid,
		}
		_, ok, packetRtn, _, err := c.downloadDoorParams(req)
		if !ok {
			c.recoverDoorParameter(doorParameterMap)
			return fmt.Errorf(
				"download door params failed, err: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			c.recoverDoorParameter(doorParameterMap)
			return fmt.Errorf(consts.RtnInfoMap[packetRtn])
		}
	}
	return nil
}

// GetDoorParameter 从门控器获取所有门的参数。
// 直接扫描所有门，不依赖 GetDoors() 避免重复调用。
func (c *Controller) GetDoorParameter() ([]driver.DoorParameter, error) {
	if _, err := c.checkConnection(); err != nil {
		return nil, err
	}

	res := make([]driver.DoorParameter, 0)
	doorNos := make([]uint32, 0)

	// 遍历所有支持的门编号，逐个查询参数
	for i := 0; i < consts.KSupportedDoorNum; i++ {
		doorNo := uint32(i + 1)
		resp, ok, packetRtn, _, err := c.getDoorParams(
			cacs.GetDoorParamsReq{Id: doorNo})
		if !ok {
			return nil, fmt.Errorf(
				"get door params failed: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			// 返回码不正常，说明该门不存在，跳过
			continue
		}
		res = append(res, driver.DoorParameter{
			Number:       driver.DoorNumberType(resp.Id),
			Name:         fmt.Sprintf("door%d", resp.Id),
			Password:     strconv.Itoa(int(resp.Password)),
			KeepOpenTime: int(resp.OpenDoorKeepTime),
			OpenTimeout:  int(resp.OpenDoorTimeoutTime),
			OpenMode:     driver.OpenModeType(resp.DoorMode),
		})
		doorNos = append(doorNos, doorNo)
	}

	// 同时更新门列表缓存
	if len(doorNos) > 0 {
		c.updateDoorCacheFromParams(doorNos)
	}

	return res, nil
}

// downloadDoorParams 下载门参数到门控器的底层通信方法。
// 返回值依次为：响应体、是否成功、返回码、错误码、错误信息。
func (c *Controller) downloadDoorParams(
	req cacs.DownloadDoorParamsReq,
) (cacs.DownloadDoorParamsResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.DownloadDoorParamsResp{}, false, 0,
			consts.KRequestError, err
	}

	// 序列化请求数据
	cmd := consts.KCommandRequestDownloadDoorParams
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.DownloadDoorParamsResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf("req marshal failed, err: %v", err)
	}
	// 发送请求到门控器
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.DownloadDoorParamsResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf("req data send failed, err: %v", err)
	}
	// 等待RRPC响应
	rrpcKey := consts.GetRRPCDownloadDoorParams(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.DownloadDoorParamsResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("rrpc get resp timeout")
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.DownloadDoorParamsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	resp, err := c.tcpMarshal.Unmarshal(
		consts.KCommandResponseDownloadDoorParams, bytes)
	if err != nil {
		c.Errorf("resp unmarshal failed, err: %v", err)
		return cacs.DownloadDoorParamsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp unmarshal failed, err: %v", err)
	}
	dlResp, ok := resp.(cacs.DownloadDoorParamsResp)
	if !ok {
		c.Errorf("resp type error, expect DownloadDoorParamsResp")
		return cacs.DownloadDoorParamsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp type error")
	}
	return dlResp, true, server.p.rtn, consts.KNormal, nil
}

// getDoorParams 从门控器获取门参数的底层通信方法。
// 返回值依次为：响应体、是否成功、返回码、错误码、错误信息。
func (c *Controller) getDoorParams(
	req cacs.GetDoorParamsReq,
) (cacs.GetDoorParamsResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.GetDoorParamsResp{}, false, 0,
			consts.KRequestError, err
	}

	// 序列化请求数据
	cmd := consts.KCommandRequestGetDoorParams
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.GetDoorParamsResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf("req marshal failed, err: %v", err)
	}
	// 发送请求到门控器
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.GetDoorParamsResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf("req data send failed, err: %v", err)
	}
	// 等待RRPC响应
	rrpcKey := consts.GetRRPCGetDoorParams(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.GetDoorParamsResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("rrpc get resp timeout")
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.GetDoorParamsResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	resp, err := c.tcpMarshal.Unmarshal(
		consts.KCommandResponseGetDoorParams, bytes)
	if err != nil {
		c.Errorf("resp unmarshal failed, err: %v", err)
		return cacs.GetDoorParamsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp unmarshal failed, err: %v", err)
	}
	getResp, ok := resp.(cacs.GetDoorParamsResp)
	if !ok {
		c.Errorf("resp type error, expect GetDoorParamsResp")
		return cacs.GetDoorParamsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp type error")
	}
	return getResp, true, server.p.rtn, consts.KNormal, nil
}
