package setup

import (
	"fmt"
	"agent/entity/config"
	"agent/entity/consts"
	"agent/logic/cm"
	"agent/logic/collector"
	"agent/logic/distribution"
	"agent/logic/network"
	"agent/logic/std"
	"agent/utils"
	"os"
	// 注册 snmp 驱动
	_ "agent/logic/collector/device/driver/drivers/snmp"

	"trpc.group/trpc-go/trpc-go/log"
)

// Init 初始化
func Init() error {
	var err error

	log.Warnf("start tbos agent, ver:%v", consts.Version)
	_ = os.WriteFile("/tmp/agent_ver", []byte(consts.Version), 0666)

	//网络配置
	if err = network.Init(); err != nil {
		log.Warnf("init network failed: %v", err)
	} else {
		log.Infof("init network success.")
	}

	// 服务配置获取和验证
	if err = config.Init(); err != nil {
		return fmt.Errorf("init config error: %w", err)
	}
	log.Infof("init config success.")

	// 采集配置获取
	if err = cm.Init(); err != nil {
		log.Warnf("init cm error: %v", err)
	} else {
		log.Infof("init cm success.")
	}

	// 异常检测
	// processor.Init()
	// log.Infof("init processor success.")

	// 采集
	if err = collector.Init(); err != nil {
		return fmt.Errorf("init collector error: %w", err)
	}
	log.Infof("init collector success.")

	// 标准测点映射计算
	if config.GetRB().IsStdCalEnable() {
		if err = std.Init(); err != nil {
			return fmt.Errorf("init std error: %w", err)
		}
		log.Infof("init std success.")
	}

	// 北向上报
	if err = distribution.Init(); err != nil {
		return fmt.Errorf("init distribution error: %w", err)
	}
	log.Infof("init distribution interval report success.")

	log.Warnf("worker id: %v", utils.WorkerID())

	return nil
}

// UnInit 退出
func UnInit() {
	if config.GetRB().IsStdCalEnable() {
		std.UnInit()
		log.Infof("std uninit finished.")
	}

	distribution.UnInit()
	log.Infof("distribution uninit finished.")

	collector.UnInit()
	log.Infof("collector uninit finished.")

	//// 等待异步清理逻辑完成
	//time.Sleep(time.Second * 2)
}
