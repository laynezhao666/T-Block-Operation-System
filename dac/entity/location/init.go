// Package location 提供边缘节点位置信息的获取和管理。
package location

import (
	"encoding/json"
	"fmt"
	"net/http"

	"dac/entity/config"

	"dac/entity/utils/thttp"
)

// Init 从GID映射服务获取当前边缘节点的位置信息。
// 支持普通模式和池化模式两种部署方式。
func Init() error {
	if config.C.Debug {
		return nil
	}

	retry := 3
	var err error
	var b []byte
	// 重试3次获取位置信息
	for retry > 0 {
		// 不需要区分模组
		b, err = thttp.Request(
			config.C.GIDMapping.URL.Location,
			http.MethodPost, nil, nil, 60000)
		if err == nil {
			break
		}
		retry--
	}
	if err != nil {
		return err
	}

	if !config.C.IsEnablePooling() {
		// 普通模式：单边缘节点
		var resp getEdgeLocationResp
		if err = json.Unmarshal(b, &resp); err != nil {
			return err
		}
		if resp.Code != 0 {
			return fmt.Errorf("code != 0, resp: %+v", resp)
		}
		l = &resp.Data
	} else {
		// 池化模式：多边缘节点
		var resp getPoolingEdgeLocationResp
		if err = json.Unmarshal(b, &resp); err != nil {
			return err
		}
		if resp.Code != 0 {
			return fmt.Errorf("code != 0, resp: %+v", resp)
		}
		l = resp.Data
	}

	config.Log.Infof("location: %s", string(b))

	return nil
}
