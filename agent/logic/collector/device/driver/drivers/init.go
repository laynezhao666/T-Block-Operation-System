package drivers

import (
	// 初始化modbus驱动
	_ "agent/logic/collector/device/driver/drivers/modbus"
	// 初始化模拟驱动
	_ "agent/logic/collector/device/driver/drivers/simulator"
	"agent/logic/collector/device/driver/drivers/snmp"
	// 初始化sysdio驱动
	_ "agent/logic/collector/device/driver/drivers/sysdio"
)

// Init 初始化驱动
func Init() error {
	var err error
	if err = snmp.Init(); err != nil {
		return err
	}

	return nil
}
