// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"fmt"

	"dac/entity/consts"
)

// GetDoorNumber 从控制器扩展配置中获取门数量
func (c *Controller) GetDoorNumber() (int, error) {
	v, ok := c.chanInfo.Extend[consts.KeyDoorNum]
	if !ok {
		return 0, fmt.Errorf("has no field \"%v\"", consts.KeyDoorNum)
	}

	switch x := v.(type) {
	case int:
		return x, nil
	case float64:
		return int(x), nil
	case float32:
		return int(x), nil
	default:
		return 0, fmt.Errorf("unsupported door number type: %T, value: %v", x, v)
	}
}
