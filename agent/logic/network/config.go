// Package network 网络配置
package network

import (
	"fmt"
	"agent/utils/file/io"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	networkConfigFile = "/etc/network/tbos-network-config.json"
)

var (
	conf *NetworkConfig = &NetworkConfig{}
)

// NetworkConfig 网络配置
type NetworkConfig struct {
	DnsMode string `json:"dns_mode"`
	Mode    string `json:"mode"`
	Bridge  struct {
		IP   string `json:"ip"`
		Mask string `json:"mask"`
	} `json:"bridge"`
	Bond struct {
		IP      string `json:"ip"`
		Mask    string `json:"mask"`
		Gateway string `json:"gateway"`
	} `json:"bond"`
}

// NetworkMode 网络模式
func NetworkMode() string {
	return getConfig().Mode
}

func getConfig() *NetworkConfig {
	err := readConfig()
	if err != nil {
		log.Errorf("read network config error: %v", err)
		return nil
	}
	return conf
}

func readConfig() error {
	return io.JSON.Read(networkConfigFile, conf)
}

func writeConfig() error {
	return io.JSON.Write(networkConfigFile, conf)
}

func changeNetworkConfig(changeFunciton func(networkConfig *NetworkConfig)) error {
	var err error
	err = readConfig()
	if err != nil {
		return fmt.Errorf("read network config error: %w", err)
	}
	changeFunciton(conf)
	err = writeConfig()
	if err != nil {
		return fmt.Errorf("write network config error: %w", err)
	}
	return nil
}

// ResetNetworkMode 重置网络模式
func ResetNetworkMode() error {
	return changeNetworkConfig(func(c *NetworkConfig) {
		c.Mode = networkModeDefault
	})
}
