// Package http 实现HTTP门禁控制器协议的驱动层。
package http

import (
	"dac/entity/config"
	"dac/entity/model/driver"
	"dac/entity/utils/dhttp"
	"encoding/json"
	"trpc.group/trpc-go/trpc-go/log"
)

// AddUser 添加用户（含人脸数据时使用表单提交）
func (c *Controller) AddUser(user driver.CardWithStaffInfo) error {
	userJson, _ := json.Marshal(user)
	if config.C.Debug {
		log.Infof("AddUser加卡 HTTP请求 URL: %s", c.urlProducer.AddUserURL())
		log.Infof("AddUser加卡 HTTP请求体: %s", string(userJson))
	}

	if user.FaceImage != "" {
		// 使用 PostFormJSON：保持 data= 前缀格式，但正确 URL 编码确保 base64 的 + 号不被破坏
		return dhttp.PostFormJSON(c.urlProducer.AddUserURL(), c.timeout, user, nil)
	} else {
		log.Infof("新增用户没有人脸数据")
	}
	return c.postJSON(c.urlProducer.AddUserURL(), user, nil)
}

// DeleteUser 删除用户
func (c *Controller) DeleteUser(user driver.UserID) error {
	userJson, _ := json.Marshal(user)
	if config.C.Debug {
		log.Infof("Delete删除卡用户 HTTP请求 URL: %s", c.urlProducer.AddUserURL())
		log.Infof("DeleteUser删除卡用户 HTTP请求体: %s", string(userJson))
	}
	return c.postJSON(c.urlProducer.DeleteUserURL(), user, nil)
}
