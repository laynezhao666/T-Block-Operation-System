package conf

import (
	"etrpc-go/config"
)

// AlarmConf 告警配置
type AlarmConf struct {
	AlarmNameInterval int `yaml:"AlarmNameInterval"`
	PushWSInterval    int `yaml:"PushWSInterval"`
	WSTaskChannelSize int `yaml:"WSTaskChannelSize"`
}

var (
	// AlarmConfImpl cgi告警接口相关配置
	AlarmConfImpl AlarmConf
)

func init() {
	// 注册配置
	config.RegisterConfigWithPrefix("alarm-ws", "AlarmConf", &AlarmConfImpl, true)
}
