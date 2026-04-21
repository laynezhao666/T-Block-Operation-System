// Package http 实现HTTP门禁控制器协议的驱动层。
package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"dac/entity/utils/thttp"
)

// GetDoors 获取控制器上所有门的信息
func (c *Controller) GetDoors() (interface{}, error) {
	url := c.urlProducer.GetDoorsURL()
	b, err := thttp.Request(
		url, http.MethodGet, nil, nil,
		int(c.timeout.Milliseconds()))
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code    int         `json:"err_code"` // 错误码
		Message string      `json:"err_msg"`  // 错误信息
		Data    interface{} `json:"data"`     // 门信息数据
	}
	if err = json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("code != 0, resp: %+v", resp)
	}

	return resp.Data, nil
}
