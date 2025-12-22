// Package network 网络配置
package network

import (
	"context"
	"fmt"
	"agent/entity/config"
	"agent/logic/network/utils"
	"agent/utils/file"
	"agent/utils/file/io"
	"agent/utils/osal"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	osutils "agent/utils/os"

	"trpc.group/trpc-go/trpc-go/log"

	pb "trpcprotocol/agent"
)

const (
	dnsModeOn  = "1"
	dnsModeOff = "0"
)

// Init 初始化网络配置
func Init() error {
	if config.GetRB().IsGatewayMode() {
		return nil
	}
	utils.Init()
	if exist, _ := file.TestExist(networkConfigFile); !exist {
		_, err := os.Create(networkConfigFile)
		if err != nil {
			return fmt.Errorf("create network config file error: %w", err)
		}
		err = io.JSON.Write(networkConfigFile, conf)
		if err != nil {
			return fmt.Errorf("write network config file error: %w", err)
		}
	}
	if NetworkMode() == networkModeSwitch {
		// 创建启动交换机模式的脚本
		err := createShell()
		if err != nil {
			log.Warnf("create enable-switch shell error: %v", err)
		}
		if exist, _ := file.TestExist(enableSwitchShellFile); !exist {
			return fmt.Errorf("file %s not exist", enableSwitchShellFile)
		}
		RunSwitchShell()
		log.Warn("enable switch mode, sleep 3s...")
		time.Sleep(time.Second * 3)
	}
	return nil
}

func isIP(str string) bool {
	return net.ParseIP(str) != nil
}

// SetNetworkStatus 设置网络状态
func SetNetworkStatus(ctx context.Context, req *pb.NetworkStatus) error {
	networkConfig := getConfig()
	switch networkConfig.Mode {
	case networkModeSwitch:
		log.WarnContext(ctx, "reset network mode")
		if err := ResetNetworkMode(); err != nil {
			return err
		}
		defer func() {
			// 暂无较好的还原网络配置方法，故直接重启
			go func(ctx context.Context) {
				log.WarnContext(ctx, "disable switch mode, call reboot...")
				time.Sleep(time.Second * 3)
				_ = osutils.Reboot()
			}(ctx)
		}()
	case networkModeDefault:
		fallthrough
	default:
		break
	}

	lan := map[string]string{}
	wlan := map[string]string{}
	lan[utils.IpKey] = req.GetLan().GetIp()
	lan[utils.MaskKey] = req.GetLan().GetMask()
	wlan[utils.IpKey] = req.GetWlan().GetIp()
	wlan[utils.MaskKey] = req.GetWlan().GetMask()
	wlan[utils.GatewayKey] = req.GetWlan().GetGateway()

	for k, v := range lan {
		if !isIP(v) {
			return fmt.Errorf("lan config %v invalid ip: %v", k, v)
		}
	}
	for k, v := range wlan {
		if !isIP(v) {
			return fmt.Errorf("lan config %v invalid ip: %v", k, v)
		}
	}

	// 处理dns-nameserver
	servers := make([]string, 0, 2)
	var gateway string
	reqWlan := req.GetWlan()
	err := changeNetworkConfig(func(networkConfig *NetworkConfig) {
		networkConfig.DnsMode = reqWlan.GetDnsMode()
	})
	if err != nil {
		return err
	}
	if reqWlan.GetDnsMode() == dnsModeOn {
		gateway = req.GetWlan().GetGateway()
		if len(gateway) == 0 {
			d, err := utils.GetInterfaceValues(utils.WlanInterface, osal.NewSet[string](utils.GatewayKey))
			if err != nil {
				return fmt.Errorf("get %v of %v error: %v", utils.GatewayKey, utils.WlanInterface, err)
			}
			gateway = d[utils.GatewayKey]
			if len(gateway) > 0 {
				servers = append(servers, gateway)
			}
		}
	} else {
		if reqWlan.GetFirstDnsServer() != "" {
			servers = append(servers, reqWlan.GetFirstDnsServer())
		}
		if reqWlan.GetSecondDnsServer() != "" {
			servers = append(servers, reqWlan.GetSecondDnsServer())
		}
	}
	if len(servers) > 0 {
		wlan[utils.NameServerKey] = strings.Join(servers, " ")
	}
	log.Debugf("set network config, waln: %+v, lan: %+v", wlan, lan)
	err = utils.ModifyInterfaceConfig(wlan, lan)
	if err != nil {
		return err
	}
	return utils.RestartNetwork()
}

// GetNetworkStatus 获取网络状态
func GetNetworkStatus(ctx context.Context) (*pb.NetworkStatus, error) {
	networkConfig := getConfig()
	log.Debugf("network config: %+v", networkConfig)
	if networkConfig == nil {
		return nil, fmt.Errorf("get network config error")
	}
	switch networkConfig.Mode {
	case networkModeSwitch:
		return getSwitchNetworkStatus()
	case networkModeDefault:
		fallthrough
	default:
		return getDefaultNetworkStatus()
	}
}

func getDefaultNetworkStatus() (*pb.NetworkStatus, error) {
	networkConfig := getConfig()
	var runningStatus int
	networkStatus := &pb.NetworkStatus{
		Mode: networkModeDefault,
		Lan:  &pb.Lan{},
		Wlan: &pb.Wlan{},
	}

	lan, err := utils.GetInterfaceStatus(utils.LanInterface)
	if err != nil {
		return nil, err
	}

	log.Debugf("lan config: %+v", lan)
	networkStatus.Lan.Ip = lan[utils.IpKey]
	networkStatus.Lan.Mask = lan[utils.MaskKey]
	runningStatus, _ = strconv.Atoi(lan[utils.StatusKey])
	networkStatus.Lan.Status = int32(runningStatus)

	wlan, err := utils.GetInterfaceStatus(utils.WlanInterface)
	if err != nil {
		return nil, err
	}

	log.Debugf("wlan config: %+v", wlan)
	keys := osal.NewSet[string](utils.GatewayKey, utils.NameServerKey)
	data, err := utils.GetInterfaceValues(utils.WlanInterface, keys)
	if err != nil {
		log.Errorf("get %v of %v error: %v", utils.GatewayKey, utils.WlanInterface, err)
	}
	log.Debugf("data config: %+v", data)
	networkStatus.Wlan.Ip = wlan[utils.IpKey]
	networkStatus.Wlan.Mask = wlan[utils.MaskKey]
	runningStatus, _ = strconv.Atoi(wlan[utils.StatusKey])
	networkStatus.Wlan.Status = int32(runningStatus)
	networkStatus.Wlan.Gateway = data[utils.GatewayKey]
	networkStatus.Wlan.DnsMode = networkConfig.DnsMode

	ipStr := data[utils.NameServerKey]
	servers := make([]string, 2)
	ips := strings.Split(ipStr, " ")
	for i, ip := range ips {
		if i > 1 {
			break
		}
		servers[i] = strings.Trim(ip, " ")
	}
	networkStatus.Wlan.FirstDnsServer = servers[0]
	networkStatus.Wlan.SecondDnsServer = servers[1]

	return networkStatus, nil
}
