// Package influxdb provides influxdb v1 client
package influxdb

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	addrCfg     = make(map[string]*UserConfig)
	addrCfgLock sync.RWMutex
)

// UserConfig 用户配置
type UserConfig struct {
	Address            string
	Username           string
	Password           string
	Timeout            time.Duration
	InsecureSkipVerify bool   // 跳过 http 认证
	WriteEncoding      string // 数据编码
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *UserConfig {
	return &UserConfig{
		Address: "localhost:8086",
		Timeout: 10 * time.Second,
	}
}

// ParseAddress address 格式 user:password@trpc.influxdb.xxx.xxx?timeout=1
//
// @param address: 地址，client address 配置的内容
// @return *UserConfig: address 配置
// @return error: 错误信息
func ParseAddress(address string) (*UserConfig, error) {

	addrCfgLock.RLock()
	cfg, ok := addrCfg[address]
	addrCfgLock.RUnlock() // 不使用 defer 是为了尽快 unlock，减小锁的范围
	if ok {
		return cfg, nil
	}

	config := GetDefaultConfig()
	// 1.移除不必要字符
	uri := strings.TrimPrefix(address, "influxdb://")
	uri = strings.TrimPrefix(uri, "influxdb+polaris://")

	// 2.解析账号密码
	var userPassword = uri
	if idx := strings.LastIndex(uri, "@"); idx != -1 {
		userPassword = uri[:idx]
		uri = uri[idx+1:]
	} else {
		uri = ""
	}
	authInfos := strings.SplitN(userPassword, ":", 2)
	if len(authInfos) < 2 {
		return nil, fmt.Errorf("address:%s format invalid: user and password", address)
	}
	config.Username = authInfos[0]
	config.Password = authInfos[1]

	// 3.解析连接地址
	addrInfo := strings.SplitN(uri, "?", 2)
	if len(addrInfo) < 1 || (len(addrInfo) == 1 && addrInfo[0] == "") {
		return nil, fmt.Errorf("address:%s format invalid: addr", address)
	}
	config.Address = addrInfo[0]

	// 没有额外配置
	if len(addrInfo) == 1 {
		return config, nil
	}

	// 4.解析超时时间
	values, err := url.ParseQuery(addrInfo[1])
	if err != nil {
		return nil, fmt.Errorf("address:%s format invalid: extra config, err:%s", address, err)
	}
	if values.Get("timeout") != "" {
		timeout, err := strconv.Atoi(values.Get("timeout"))
		if err != nil {
			return nil, fmt.Errorf("address:%s format invalid: timeout:%s", address, values.Get("timeout"))
		} else {
			config.Timeout = time.Duration(timeout) * time.Millisecond
		}
	}
	// 5.解析是否跳过 http 认证
	if values.Get("insecure_skip_verify") == "true" {
		config.InsecureSkipVerify = true
	}

	// 6.解析数据编码
	if values.Get("write_encoding") != "" {
		config.WriteEncoding = values.Get("write_encoding")
	}

	return config, nil
}
