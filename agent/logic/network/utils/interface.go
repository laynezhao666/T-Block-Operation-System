package utils

import (
	"agent/entity/consts"
	"agent/utils/osal"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	backdoorIp = "172.16.1.250" // tbox ETH1口后门ip
)

var (
	interfacesConfigFile = "/etc/network/interfaces"
	tboxVersionFile      = "/etc/tbox/version.conf"
	IpKey                = "ip"
	MaskKey              = "mask"
	StatusKey            = "status"
	GatewayKey           = "gateway"
	NameServerKey        = "dns-nameservers"
	networking           = "/etc/init.d/networking"
)

var (
	interfacesMutex sync.RWMutex
	WlanInterface   = "eth0"
	LanInterface    = "eth1"
	blockReg        *regexp.Regexp
	fieldNameMap    = map[string]string{
		IpKey:      "address",
		MaskKey:    "netmask",
		GatewayKey: GatewayKey,
	}
)

// FieldValue 字段值
type FieldValue struct {
	rawValue   string
	wordNumber int
	rawKey     string
	index      int
}

var (
	// Version 版本号在编译时通过ldflags设置
	Version = consts.Version
	// GitCommit Git提交哈希
	GitCommit = "undefined"
	// GitBranch Git分支
	GitBranch = "undefined"
	// BuildTime 构建时间
	BuildTime = "undefined"
	// GoVersion Go版本
	GoVersion = "undefined"
	// hardwareVersion 硬件版本
	hardwareVersion = ""
)

// Init 初始化
func Init() {
	blockReg, _ = regexp.Compile(`\s*(?s).+?(\n\s*?\n|$)`)
	// 识别TBOX硬件版本，适配对应的网络接口名
	context, err := os.ReadFile(tboxVersionFile)
	if err == nil {
		hardwareVersion = strings.TrimSpace(string(context))
		// TBOX2.0的网口lan1，lan2，代替eth0, eth1
		if strings.Contains(hardwareVersion, "TBOX-2") {
			WlanInterface = "lan1"
			LanInterface = "lan2"
		}
	}
}

// GetVersion 获取完整的版本信息
func GetVersion() (string, map[string]string) {
	return Version, map[string]string{
		"branch":    GitBranch,
		"commit":    GitCommit,
		"buildTime": BuildTime,
		"goVersion": GoVersion,
		"hardware":  hardwareVersion,
	}
}

// GetInterfaceStatus 获取网卡状态
func GetInterfaceStatus(interfaceName string) (map[string]string, error) {
	log.Debugf("get interface %v status", interfaceName)
	ifData, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("net.InterfaceByName(%v) error: %w", interfaceName, err)
	}
	addrs, err := ifData.Addrs()
	if err != nil {
		return nil, err
	}

	ip := ""
	mask := ""
	flag := false
	for i := range addrs {
		addr := addrs[i]
		switch x := addr.(type) {
		case *net.IPNet:
			if x.IP.String() == backdoorIp {
				continue
			}
			ip = x.IP.String()
			mask = net.IP(x.Mask).String()
			flag = true
			break
		}
		if flag {
			break
		}
	}

	data := make(map[string]string)
	data[IpKey] = ip
	data[MaskKey] = mask

	running := 0
	log.Debugf("ifdata flag: %v; net flagup: %v \n", ifData.Flags, net.FlagUp)
	log.Debugf("ifdata flag: %b; net flagup: %b \n", ifData.Flags, net.FlagUp)
	if (ifData.Flags & net.FlagUp) != 0 {
		running = 1
	}
	data[StatusKey] = fmt.Sprintf("%d", running)

	return data, nil
}

// GetInterfaceValues 获取网卡配置
func GetInterfaceValues(interfaceName string, keys osal.Set[string]) (map[string]string, error) {
	blocks, err := getBlocks()
	if err != nil {
		return nil, err
	}
	index := getInterfaceContentIndex(blocks, interfaceName)
	if index == -1 {
		return nil, fmt.Errorf("%v not found", interfaceName)
	}
	interfaceContent := blocks[index]
	data := make(map[string]string, len(keys))
	traverseInterfaceLine(interfaceContent, func(_ int, rawKey string, words []string) {
		if has := keys.Contains(rawKey); has {
			data[rawKey] = strings.Join(words[1:], " ")
		}
	})
	return data, nil
}

func getBlocks() ([]string, error) {
	interfacesMutex.RLock()
	defer interfacesMutex.RUnlock()

	fileData, err := os.ReadFile(interfacesConfigFile)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileData)
	blocks := blockReg.FindAllString(fileContent, -1)
	for i, b := range blocks {
		blocks[i] = strings.Trim(b, "\n")
	}
	return blocks, nil
}

func getInterfaceContentIndex(blocks []string, interfaceName string) int {
	flag := interfaceName + " "
	for i, block := range blocks {
		if strings.Contains(block, flag) {
			return i
		}
	}
	return -1
}

func traverseInterfaceLine(interfaceContent string, f func(index int, rawKey string, words []string)) {
	lines := strings.Split(interfaceContent, "\n")
	j := 0
	index := 0
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		trimLine := strings.Trim(line, " ")
		words := strings.Split(trimLine, " ")
		if len(words) < 2 {
			continue
		}
		for j = range line {
			if line[j] != ' ' {
				break
			}
		}
		prefixSpaces := line[:j]
		rawKey := strings.TrimSpace(prefixSpaces + words[0])
		f(index, rawKey, words)
		index++
	}
}

func convertToInterfaceFieldMap(conf map[string]string) map[string]string {
	newMap := make(map[string]string, len(conf))
	for k, v := range conf {
		newKey, ok := fieldNameMap[k]
		if !ok {
			continue
		}
		newMap[newKey] = v
	}
	return newMap
}

// ModifyInterfaceConfig 修改网络配置
func ModifyInterfaceConfig(wlanMap, lanMap map[string]string) error {
	blocks, err := getBlocks()
	if err != nil {
		return err
	}
	newWlanMap := convertToInterfaceFieldMap(wlanMap)
	newLanMap := convertToInterfaceFieldMap(lanMap)

	wlanIndex := getInterfaceContentIndex(blocks, WlanInterface)
	if wlanIndex == -1 {
		return fmt.Errorf("%v not found", WlanInterface)
	}
	wlanInterfaceContent := blocks[wlanIndex]
	modeifiedWlanBlock, err := modifyInterface(wlanInterfaceContent, newWlanMap)
	if err != nil {
		return err
	}
	blocks[wlanIndex] = modeifiedWlanBlock

	lanIndex := getInterfaceContentIndex(blocks, LanInterface)
	if lanIndex == -1 {
		return fmt.Errorf("%v not found", LanInterface)
	}
	lanInterfaceContent := blocks[lanIndex]
	modeifiedLanBlock, err := modifyInterface(lanInterfaceContent, newLanMap)
	if err != nil {
		return err
	}
	blocks[lanIndex] = modeifiedLanBlock
	modifiedContent := strings.Join(blocks, "\n\n")
	modifiedContent += "\n\n"
	return updateInterfaceConfigure(modifiedContent)
}

func modifyInterface(interfaceContent string, updateData map[string]string) (string, error) {
	if len(updateData) == 0 {
		return interfaceContent, nil
	}

	blockData := make(map[string]*FieldValue)

	traverseInterfaceLine(interfaceContent, func(index int, rawKey string, words []string) {
		newKey := strings.TrimSpace(rawKey)
		blockData[newKey] = &FieldValue{
			rawValue:   strings.Join(words[1:], " "),
			wordNumber: len(words),
			rawKey:     rawKey,
			index:      index,
		}
	})

	appendField := make([]string, 0)
	for k, v := range updateData {
		fieldValue, ok := blockData[k]
		if !ok {
			appendField = append(appendField, k)
		} else {
			if fieldValue.wordNumber != 2 {
				return "", fmt.Errorf("invalid format, try to set %v = %v", k, v)
			}
			fieldValue.rawValue = v
		}
	}

	lines := make([]string, len(blockData))
	for _, fieldValue := range blockData {
		lines[fieldValue.index] = fmt.Sprintf("%v %v", fieldValue.rawKey, fieldValue.rawValue)
	}
	for _, k := range appendField {
		line := fmt.Sprintf("%v %v", k, updateData[k])
		lines = append(lines, line)
	}
	interfaceContent = strings.Join(lines, "\n")
	interfaceContent = strings.Trim(interfaceContent, "\n")
	return interfaceContent, nil
}

func updateInterfaceConfigure(fileContent string) error {
	interfacesMutex.Lock()
	defer interfacesMutex.Unlock()

	if err := os.WriteFile(interfacesConfigFile, []byte(fileContent), os.ModePerm); err != nil {
		return err
	}

	return nil
}

// RestartNetwork 重启网络
func RestartNetwork() error {
	cmd := exec.Command(networking, "restart")
	return cmd.Run()
}
