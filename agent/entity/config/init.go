package config

import (
	"net"

	"etrpc-go/config"
)

var (
	Conf    Config
	LocalIP string
)

func init() {
	// 业务配置不进行热更新，避免map并发访问
	config.RegisterConfig("agent.yaml", &Conf, false)
}

// Init 初始化配置
func Init() error {
	LocalIP = GetLocalIP()
	//var err error
	//
	//if err = GetRB().validate(); err != nil {
	//	return fmt.Errorf("validate config error: %w", err)
	//}

	return nil
}

// GetLocalIP 获取本地IP
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// IsKafkaLogEnable 是否开启kafka日志
func IsKafkaLogEnable() bool {
	return GetRB().IsKafkaLogEnable()
}
