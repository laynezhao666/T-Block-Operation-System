// Package http 实现HTTP门禁控制器协议的驱动层。
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"dac/entity/config"
	"dac/entity/model/driver"
	"dac/entity/utils"
	"dac/entity/utils/batch"
	"dac/entity/utils/thttp"
)

// 非法JSON修复常量
const (
	invalidStr = "}{"  // 非法JSON分隔符
	validStr   = "},{" // 合法JSON分隔符
)

// DoorParameter 门参数（标准协议格式，door_no字段）
// 注意：所有字段必填，否则调用失败
type DoorParameter struct {
	Number         int    `json:"door_no"`          // 门编号
	Name           string `json:"door_name"`        // 门名称
	Password       string `json:"password"`         // 门密码
	KeepOpenTime   int    `json:"keep_time"`        // 门开保持时间，单位：秒
	OpenTimeout    int    `json:"open_time"`        // 门开超时时间，单位：秒
	LockErrorCount int    `json:"lock_err_cnt"`     // 卡封锁错误次数(连续刷多少次非法卡后门封卡)
	SlotInterval   int    `json:"slot_interval"`    // 非法卡刷卡间隔，单位秒
	LockTime       int    `json:"lock_time"`        // 非法卡的封锁时间，单位：秒
	OpenMode       int    `json:"open_mode"`        // 开门模式：0-刷卡，1-密码，2-卡+密码，3-卡或密码
	FireSignalMode int    `json:"fire_signal_mode"` // 火警信号：0-短路有效，1-断路有效
}

// DoorParameterForDoorCompatible 门参数兼容格式（door字段）
type DoorParameterForDoorCompatible struct {
	Number         int    `json:"door"`             // 门编号（兼容字段名）
	Name           string `json:"door_name"`        // 门名称
	Password       string `json:"password"`         // 门密码
	KeepOpenTime   int    `json:"keep_time"`        // 门开保持时间
	OpenTimeout    int    `json:"open_time"`        // 门开超时时间
	LockErrorCount int    `json:"lock_err_cnt"`     // 卡封锁错误次数
	SlotInterval   int    `json:"slot_interval"`    // 非法卡刷卡间隔
	LockTime       int    `json:"lock_time"`        // 非法卡封锁时间
	OpenMode       int    `json:"open_mode"`        // 开门模式
	FireSignalMode int    `json:"fire_signal_mode"` // 火警信号模式
}

// convertDoorCompatible 将兼容格式的门参数转换为标准格式
func convertDoorCompatible(s *DoorParameterForDoorCompatible) DoorParameter {
	return DoorParameter{
		Number:         s.Number,
		Name:           s.Name,
		Password:       s.Password,
		KeepOpenTime:   s.KeepOpenTime,
		OpenTimeout:    s.OpenTimeout,
		LockErrorCount: s.LockErrorCount,
		SlotInterval:   s.SlotInterval,
		LockTime:       s.LockTime,
		OpenMode:       s.OpenMode,
		FireSignalMode: s.FireSignalMode,
	}
}

// SetDoorParameter 批量设置门参数到控制器
func (c *Controller) SetDoorParameter(params []driver.DoorParameter) error {
	if c.isVersionMDC {
		// MDC 版本协议无该接口
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	l := len(params)

	reqs := make([]DoorParameter, 0, l)
	for i := range params {
		p := &params[i]
		reqs = append(reqs, DoorParameter{
			Number:         int(p.Number),
			Name:           p.Name,
			Password:       p.Password,
			KeepOpenTime:   p.KeepOpenTime,
			OpenTimeout:    p.OpenTimeout,
			LockErrorCount: p.LockCount,
			SlotInterval:   p.VerifyInterval,
			LockTime:       p.LockTime,
			OpenMode:       int(p.OpenMode),
			FireSignalMode: p.FireSignalMode,
		})
	}

	args := make([]interface{}, 0, l)
	for i := range reqs {
		args = append(args, &reqs[i])
	}

	err := batch.Execute(ctx, args, func(ctx context.Context, arg interface{}) error {
		r, ok := arg.(*DoorParameter)
		if !ok {
			return nil
		}

		url := c.urlProducer.SetDoorParameterURL()
		if config.C.Debug {
			c.logger.Infof("post %v, set door param: %+v", url, utils.GetJSONString(r))
		}
		return c.postJSON(url, r, nil)
	})

	return err
}

// GetDoorParameter 从控制器获取所有门参数配置
func (c *Controller) GetDoorParameter() ([]driver.DoorParameter, error) {
	// 结构体中 Name 字段在为空时忽略，尚未处理该情况
	// 故使用标准库 json 进行反序列化操作
	var params []DoorParameter
	url := c.urlProducer.GetDoorParameterURL()

	if c.isVersionMDC {
		// MDC 协议版本无该接口，使用获取门信息接口替代
		url = c.urlProducer.GetDoorsURL()
	}
	b, err := thttp.Request(url, http.MethodGet, nil, nil, int(c.timeout.Milliseconds()))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Code    int         `json:"err_code"`
		Message string      `json:"err_msg"`
		Data    interface{} `json:"data"`
	}
	if err = json.Unmarshal(b, &resp); err != nil {
		c.logger.Warnf("unmarshal door controller: %v  error: %v, response: %v",
			c.chanInfo.ChannelID, err, string(b))

		/*
			厂商返回数据不一定合法，某个门禁控制器返回如下：
			{"err_code":0,"err_msg":"ok","data":[{"door_no":1,"password":"123456","door_name":"监控中心北门",
			"keep_time":5, "open_time":30, "lock_err_cnt":5,"slot_interval":60,"lock_time":300,"open_mode":3,
			"fire_signal_mode":0},{"door_no":2,"password":"0000","door_name":"监控中心南门","keep_time":3,
			"open_time":30,"lock_err_cnt":5,"slot_interval":60,"lock_time":300,"open_mode":0,
			"fire_signal_mode":0}]}
			此处尝试修复 json 字符串。
		*/
		if err = json.Unmarshal([]byte(strings.ReplaceAll(string(b), invalidStr, validStr)), &resp); err != nil {
			return nil, err
		}
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("code != 0, resp: %+v", utils.GetJSONString(resp))
	}

	if b, err = json.Marshal(resp.Data); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, &params); err != nil {
		return nil, err
	}

	// 某些厂商的返回数据中，门编号字段可能不满足协议
	doorNumberHasZero := false
	for i := range params {
		if params[i].Number == 0 {
			doorNumberHasZero = true
			c.logger.Warnf("门编号为 0，尝试修复，response: %v", string(b))
			break
		}
	}
	if doorNumberHasZero {
		var tempParams []DoorParameterForDoorCompatible
		if err = json.Unmarshal(b, &tempParams); err != nil {
			return nil, err
		}
		params = make([]DoorParameter, 0, len(tempParams))
		for i := range tempParams {
			params = append(params, convertDoorCompatible(&tempParams[i]))
		}
	}

	results := make([]driver.DoorParameter, 0, len(params))
	for i := range params {
		p := &params[i]

		results = append(results, driver.DoorParameter{
			Number:         driver.DoorNumberType(p.Number),
			Name:           p.Name,
			Password:       p.Password,
			KeepOpenTime:   p.KeepOpenTime,
			OpenTimeout:    p.OpenTimeout,
			LockCount:      p.LockErrorCount,
			LockTime:       p.LockTime,
			VerifyInterval: p.SlotInterval,
			OpenMode:       driver.OpenModeType(p.OpenMode),
			FireSignalMode: p.FireSignalMode,
		})
	}

	return results, err
}
