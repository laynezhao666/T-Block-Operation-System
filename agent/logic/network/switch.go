package network

import (
	"context"
	"fmt"
	"agent/logic/network/utils"
	osutils "agent/utils/os"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	pb "trpcprotocol/agent"

	"trpc.group/trpc-go/trpc-go/log"
)

var (
	allInterfaces = []string{"lan1", "lan2", "lan3", "lan4", "lan5", "lan6", "lan7", "lan8", "lan9", "lan10"}
	allCommands   = []string{"ip", "modprobe"}
)

// EnableSwitch 启用交换机模式
func EnableSwitch(ctx context.Context, req *pb.EnableSwitchReq) error {
	var err error

	// 修改配置文件
	if err := changeNetworkConfig(func(c *NetworkConfig) {
		c.Mode = networkModeSwitch

		c.Bridge.IP = req.GetBridge().GetIp()
		c.Bridge.Mask = req.GetBridge().GetMask()

		c.Bond.IP = req.GetBond().GetIp()
		c.Bond.Mask = req.GetBond().GetMask()
		c.Bond.Gateway = req.GetBond().GetGateway()
	}); err != nil {
		return err
	}

	// 创建启动交换机模式的脚本文件
	err = createShell()
	if err != nil {
		return fmt.Errorf("req %+v error: %v", req, err)
	}

	err = RunSwitchShell()
	if err != nil {
		return err
	}

	go func(ctx context.Context) {
		time.Sleep(time.Second * 3)
		log.WarnContext(ctx, "enable switch mode, restart...")
		osutils.ReStartApp()
	}(ctx)

	return nil
}

func createShell() error {
	networkConfig := getConfig()
	bridgeMaskLen, _ := net.IPMask(net.ParseIP(networkConfig.Bridge.Mask).To4()).Size()
	if bridgeMaskLen == 0 {
		return fmt.Errorf("bridge mask \"%v\" invalid", networkConfig.Bridge.Mask)
	}
	bondMaskLen, _ := net.IPMask(net.ParseIP(networkConfig.Bond.Mask).To4()).Size()
	if bondMaskLen == 0 {
		return fmt.Errorf("bond mask \"%v\" invalid", bridgeMaskLen)
	}
	shell := fmt.Sprintf(commandTemplate, networkConfig.Bridge.IP,
		bridgeMaskLen, networkConfig.Bond.IP, bondMaskLen, networkConfig.Bond.Gateway)
	// 创建时已设置可执行权限
	if err := os.WriteFile(enableSwitchShellFile, []byte(shell), os.ModePerm); err != nil {
		return err
	}
	return nil
}

// RunSwitchShell 执行交换机模式脚本
func RunSwitchShell() error {
	var err error
	if err = verifyCommand(); err != nil {
		return err
	}
	if err = verifyInterfaces(); err != nil {
		return err
	}
	// 执行修改脚本
	b, err := exec.Command("sh", "-c", enableSwitchShellFile).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run %v, output: %v, error: %v", enableSwitchShellFile, string(b), err)
	}
	log.Infof("run %v success, output: %v", enableSwitchShellFile, string(b))
	return nil
}

func verifyCommand() error {
	var err error
	for _, cmd := range allCommands {
		if _, err = exec.LookPath(cmd); err != nil {
			return fmt.Errorf("get command \"%v\" error: %v", cmd, err)
		}
	}
	return nil
}

func verifyInterfaces() error {
	var err error
	for _, name := range allInterfaces {
		if _, err = net.InterfaceByName(name); err != nil {
			return fmt.Errorf("get interface \"%v\" error: %v", name, err)
		}
	}
	return nil
}

// DisableSwitch 关闭交换机模式
func DisableSwitch(ctx context.Context) error {
	var err error
	// 删除网络配置脚本
	err = os.Remove(enableSwitchShellFile)
	if err != nil {
		return err
	}
	// 修改配置文件
	err = changeNetworkConfig(func(c *NetworkConfig) {
		c.Mode = networkModeDefault
	})
	if err != nil {
		return err
	}
	// 暂无较好的还原网络配置方法，故直接重启
	go func(ctx context.Context) {
		time.Sleep(time.Second * 3)
		log.WarnContext(ctx, "disable switch mode, call reboot...")
		_ = osutils.Reboot()
	}(ctx)

	return nil
}

func getSwitchNetworkStatus() (*pb.NetworkStatus, error) {
	networkConfig := getConfig()
	var runningStatus int
	networkStatus := &pb.NetworkStatus{
		Mode: networkModeSwitch,
		Lan:  &pb.Lan{},
		Wlan: &pb.Wlan{},
	}

	lan, err := utils.GetInterfaceStatus(bridge0Interface)
	if err != nil {
		return nil, err
	}

	log.Debugf("switch lan config: %+v", lan)
	networkStatus.Lan.Ip = lan[utils.IpKey]
	networkStatus.Lan.Mask = lan[utils.MaskKey]
	runningStatus, _ = strconv.Atoi(lan[utils.StatusKey])
	networkStatus.Lan.Status = int32(runningStatus)

	wlan, err := utils.GetInterfaceStatus(bond0Interface)
	if err != nil {
		return nil, err
	}
	log.Debugf("switch wlan config: %+v", lan)
	networkStatus.Wlan.Ip = wlan[utils.IpKey]
	networkStatus.Wlan.Mask = wlan[utils.MaskKey]
	networkStatus.Wlan.Gateway = networkConfig.Bond.Gateway
	runningStatus, _ = strconv.Atoi(wlan[utils.StatusKey])
	networkStatus.Wlan.Status = int32(runningStatus)

	return networkStatus, nil
}
