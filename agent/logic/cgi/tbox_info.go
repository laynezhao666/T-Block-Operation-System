package cgi

import (
	"context"
	"fmt"

	econfig "etrpc-go/config"

	pb "trpcprotocol/agent"

	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/config"
	cts "agent/entity/consts"
	"agent/logic/setup"
	cm "agent/repo/cm"
	stdos "os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ReportSn field constants
const (
	ReportSnEnable  = 1
	ReportSnDisable = 2
)

// ReportMonitor field constants
const (
	ReportMonitorEnable  = 1
	ReportMonitorDisable = 2
)

const collectorServiceName = "idc-tbos-collector"

const (
	moduleTimeShPath    = "/opt/tbbox/shells/module_time.sh"
	multiIPMarker       = "for server in ${servers}; do"
	defaultProxyPort    = "30091"
	proxyTargetIPPrefix = "ip://"
)

// SetSnBindingHandle 设置SN绑定
func SetSnBindingHandle(ctx context.Context, req *pb.SetSnBindingReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// GetBasicInfoHandle 获取基本信息
func GetBasicInfoHandle(ctx context.Context, req *emptypb.Empty) (*pb.GetBasicInfoRsp, error) {
	rsp := &pb.GetBasicInfoRsp{
		DeviceNumber: make([]string, 0),
		Hotstandby:   make([]*pb.HotstandbyInfo, 0),
	}
	rsp.Source = config.GetRB().Project.Source
	ips := getCollectorIps()
	if len(ips) > 0 {
		rsp.ReportAddr = strings.Join(ips, ",")
	}
	for _, dev := range config.GetRB().Task.Local.Devs {
		rsp.DeviceNumber = append(rsp.DeviceNumber, dev)
	}
	for devNum, val := range config.GetRB().Task.Local.HotStandby {
		rsp.Hotstandby = append(rsp.Hotstandby, &pb.HotstandbyInfo{
			DeviceNumber: devNum,
			IsMaster:     val.IsMaster,
			Ip:           val.Ip})
	}
	if config.GetRB().Tbox.SnReportEnabled {
		rsp.ReportSn = ReportSnEnable
	} else {
		rsp.ReportSn = ReportSnDisable
	}
	if config.GetRB().MonitorProxy.Enabled {
		rsp.ReportMonitor = ReportMonitorEnable
	} else {
		rsp.ReportMonitor = ReportMonitorDisable
	}
	rsp.ProxyTarget = config.GetRB().MonitorProxy.ProxyTarget
	rsp.AppMark = config.GetRB().MonitorProxy.AppMark
	rsp.MetricGroup = config.GetRB().MonitorProxy.MetricGroup
	rsp.Distributor = buildDistributorRsp()
	rsp.Ms = time.Now().UnixMilli()
	return rsp, nil
}

// SetBasicInfoHandle 设置基本信息
func SetBasicInfoHandle(ctx context.Context, req *pb.SetBasicInfoReq) (*emptypb.Empty, error) {
	if config.GetRB().IsGatewayMode() {
		return &emptypb.Empty{}, fmt.Errorf("gateway mode not supported")
	}
	hasChange := false
	ips := getCollectorIps()
	ipsStr := strings.Join(ips, ",")
	if len(req.ReportAddr) > 0 && req.ReportAddr != ipsStr {
		if ok := updateCollectorTargets(req.ReportAddr); !ok {
			return &emptypb.Empty{}, fmt.Errorf("report address replace fail")
		}
		// 更新collector ip
		config.GetRB().Update([]string{"client", "service"}, trpc.GlobalConfig().Client.Service)

		// 判断 module_time.sh 是否存在，存在才更新 time.json
		applyNtpServerConfig(req.ReportAddr)

		hasChange = true
	}
	if len(req.Source) > 0 && req.Source != config.GetRB().Project.Source {
		// 仅支持 deviceModel、local、tlink 三种模式之间切换
		allowedSources := map[string]bool{
			cm.LocalFileConfigModName: true,
			cm.TLinkModName:           true,
		}
		currentSource := config.GetRB().Project.Source
		if !allowedSources[currentSource] {
			return &emptypb.Empty{}, fmt.Errorf("当前模式(%s)不支持切换Source", currentSource)
		}
		if !allowedSources[req.Source] {
			return &emptypb.Empty{}, fmt.Errorf("Source值不合法，仅支持: %s, %s",
				cm.LocalFileConfigModName, cm.TLinkModName)
		}
		config.GetRB().Update([]string{"project", "source"}, req.Source)
		config.GetRB().Project.Source = req.Source
		hasChange = true
	}
	// 更新 Distributor 转发配置
	if len(req.Distributor) > 0 {
		if changed := applyDistributorConfig(req.Distributor); changed {
			hasChange = true
		}
	}
	// 更新ReportSn字段
	if req.ReportSn == ReportSnEnable {
		config.GetRB().Update([]string{"tbox", "sn_report_enabled"}, true)
		config.GetRB().Tbox.SnReportEnabled = true
		log.Infof("SetBasicInfoHandle, enable sn report")
	} else if req.ReportSn == ReportSnDisable {
		config.GetRB().Update([]string{"tbox", "sn_report_enabled"}, false)
		config.GetRB().Tbox.SnReportEnabled = false
		log.Infof("SetBasicInfoHandle, disable sn report")
	}

	if len(req.DeviceNumber) > 0 && !reflect.DeepEqual(req.DeviceNumber, config.GetRB().Task.Local.Devs) {
		config.GetRB().Update([]string{"task", "local", "devs"}, req.DeviceNumber)
		config.GetRB().Task.Local.Devs = req.DeviceNumber
		config.GetRB().Update([]string{"global", "container_name"}, req.DeviceNumber[0])
		hasChange = true
	}
	newHotstandby := toHostandby(req.Hotstandby)
	if !hostandbyEqual(newHotstandby, config.GetRB().Task.Local.HotStandby) {
		config.GetRB().Update([]string{"task", "local", "hot_standby"}, newHotstandby)
		config.GetRB().Task.Local.HotStandby = newHotstandby
		hasChange = true
	}
	// 更新ReportSn,这个字段不需要重启生效，所以不设置hasChange
	if hasChange {
		log.Infof("SetBasicInfoHandle, hasChange: %+v, restart", req)
		agentRestart()
	}
	return &emptypb.Empty{}, nil
}

// applyNtpServerConfig 根据新的上报地址更新 NTP 服务器配置（time.json）。
// 若 module_time.sh 不存在则跳过；支持多IP时使用完整地址，否则只取第一个IP。
func applyNtpServerConfig(reportAddr string) {
	if _, err := stdos.Stat(moduleTimeShPath); err != nil {
		return
	}
	shContent, err := stdos.ReadFile(moduleTimeShPath)
	if err != nil {
		log.Errorf("applyNtpServerConfig, 读取module_time.sh失败: %v", err)
		return
	}
	var timeServerIP string
	if strings.Contains(string(shContent), multiIPMarker) {
		// 支持多IP，直接使用完整地址
		timeServerIP = reportAddr
	} else {
		// 不支持多IP，只取第一个IP
		if ips := extractIPsFromReportAddr(reportAddr); len(ips) > 0 {
			timeServerIP = ips[0]
		}
	}
	if timeServerIP == "" {
		return
	}
	if err := updateTimeJsonServer(timeServerIP); err != nil {
		log.Errorf("applyNtpServerConfig, 更新time.json失败: %v", err)
	}
}

const defaultMasterPort = 61000

func toHostandby(a []*pb.HotstandbyInfo) map[string]config.HotStandbyDev {
	dev2info := make(map[string]config.HotStandbyDev)
	port := econfig.GetInt64OrDefault("etrpc.service_port", defaultMasterPort)
	for _, val := range a {
		dev2info[val.DeviceNumber] = config.HotStandbyDev{
			IsMaster: val.IsMaster,
			Ip:       val.Ip,
			Port:     port,
		}
	}
	return dev2info
}

func hostandbyEqual(a, dev2info map[string]config.HotStandbyDev) bool {
	if len(a) != len(dev2info) {
		return false
	}
	for devNum, info := range dev2info {
		val, ok := a[devNum]
		if !ok || val.IsMaster != info.IsMaster || val.Ip != info.Ip || val.Port != info.Port {
			return false
		}
	}
	return true
}

func getCollectorIps() []string {
	ips := make([]string, 0)
	ipSet := make(map[string]bool)
	for _, v := range trpc.GlobalConfig().Client.Service {
		if !strings.HasPrefix(v.ServiceName, collectorServiceName) {
			continue
		}
		if len(v.Target) == 0 {
			continue
		}
		// 解析ip://10.5.39.220:30082,10.5.39.221:30082,10.5.39.222:30082格式，提取为10.5.39.220,10.5.39.221,10.5.39.222
		if !strings.HasPrefix(v.Target, "ip://") {
			continue
		}
		// 去掉ip://前缀
		ipsWithPorts := strings.TrimPrefix(v.Target, "ip://")
		// 按逗号分割多个地址
		addresses := strings.Split(ipsWithPorts, ",")
		for _, addr := range addresses {
			// 按冒号分割IP和端口
			parts := strings.Split(addr, ":")
			if len(parts) >= 2 {
				// 提取IP地址部分
				ipSet[parts[0]] = true
			}
		}
	}
	for ip, _ := range ipSet {
		ips = append(ips, ip)
	}
	sort.Strings(ips)
	return ips
}

// buildProxyTarget 将逗号分隔的 IP 列表转换为 proxy_target 格式
// 例如 "10.40.30.3,10.40.30.5" -> "ip://10.40.30.3:30091,10.40.30.5:30091"
func buildProxyTarget(collectorIP string) string {
	ips := strings.Split(collectorIP, ",")
	for i, ip := range ips {
		ip = strings.TrimSpace(ip)
		if !strings.Contains(ip, ":") {
			ip = ip + ":" + defaultProxyPort
		}
		ips[i] = ip
	}
	return proxyTargetIPPrefix + strings.Join(ips, ",")
}

// updateCollectorTargets 原地更新 trpc.GlobalConfig().Client.Service 中所有 collector 条目的 Target，
// 保留非 collector 条目不变。ips 为逗号分隔的 IP 列表。
// 若没有找到任何可更新的 collector 条目则返回 false。
func updateCollectorTargets(ips string) bool {
	ipList := strings.Split(ips, ",")
	if len(ipList) == 0 {
		return false
	}

	updated := false
	for _, v := range trpc.GlobalConfig().Client.Service {
		if !strings.HasPrefix(v.ServiceName, collectorServiceName) {
			continue
		}
		if !strings.HasPrefix(v.Target, "ip://") {
			continue
		}

		// 从原 Target 中提取端口
		var port string
		ipsWithPorts := strings.TrimPrefix(v.Target, "ip://")
		addresses := strings.Split(ipsWithPorts, ",")
		if len(addresses) > 0 {
			parts := strings.Split(addresses[0], ":")
			if len(parts) >= 2 {
				port = parts[1]
			}
		}
		if port == "" {
			continue
		}

		// 用新 IP 列表 + 原端口构建新 Target，原地修改
		newTargets := make([]string, 0, len(ipList))
		for _, ip := range ipList {
			newTargets = append(newTargets, fmt.Sprintf("%s:%s", ip, port))
		}
		v.Target = "ip://" + strings.Join(newTargets, ",")
		updated = true
	}

	return updated
}

// extractIPsFromReportAddr 从ReportAddr中提取IP地址列表
// ReportAddr格式：ip,ip,ip
func extractIPsFromReportAddr(reportAddr string) []string {
	ips := make([]string, 0)
	if reportAddr == "" {
		return ips
	}

	// 按逗号分割多个IP地址
	ipList := strings.Split(reportAddr, ",")
	for _, ip := range ipList {
		// 去除可能的空格
		cleanIP := strings.TrimSpace(ip)
		if cleanIP != "" {
			ips = append(ips, cleanIP)
		}
	}
	return ips
}

// updateTimeJsonServer 更新time.json文件中的server字段
func updateTimeJsonServer(ip string) error {
	const timeJsonPath = "/opt/tbbox/3rd/conf/time.json"

	// 检查文件是否存在
	if _, err := stdos.Stat(timeJsonPath); stdos.IsNotExist(err) {
		log.Warnf("time.json文件不存在: %s", timeJsonPath)
		return nil // 文件不存在时静默返回，不报错
	}

	// 读取文件内容
	content, err := stdos.ReadFile(timeJsonPath)
	if err != nil {
		return fmt.Errorf("读取time.json失败: %v", err)
	}

	// 使用正则表达式替换server字段，同时保留原有格式
	re := regexp.MustCompile(`("server"\s*:\s*")[^"]*(")`)
	newContent := re.ReplaceAllString(string(content), "${1}"+ip+"${2}")

	// 如果内容没有变化，直接返回
	if newContent == string(content) {
		log.Debugf("time.json内容未变化，无需更新")
		return nil
	}

	// 写回文件
	if err := stdos.WriteFile(timeJsonPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("写入time.json失败: %v", err)
	}

	log.Infof("成功更新time.json的server字段为: %s", ip)
	return nil
}

// agentRestart 重启agent，与 service.BoxManager.AgentRestart 逻辑一致
func agentRestart() {
	go func() {
		time.Sleep(time.Second * 3)
		log.Warn("Agent Restart")
		setup.UnInit()
		stdos.Exit(0)
	}()
}

// buildDistributorRsp 构建 distributor 响应，仅返回配置中已启用（Enable 非空）的项
func buildDistributorRsp() map[string]*pb.Distributor {
	result := make(map[string]*pb.Distributor)
	dist := config.GetRB().Distributor

	// tlink：仅有 enable
	if len(dist.Tlink.Enable) > 0 {
		result[cts.DistKeyTlink] = &pb.Distributor{Enable: dist.Tlink.Enable}
	}

	// bypass：enable + target / client_id
	if len(dist.Bypass.Enable) > 0 {
		cfg := map[string]string{}
		if dist.Bypass.Target != "" {
			cfg[cts.DistCfgTarget] = dist.Bypass.Target
		}
		if dist.Bypass.ClientId != "" {
			cfg[cts.DistCfgClientID] = dist.Bypass.ClientId
		}
		result[cts.DistKeyBypass] = &pb.Distributor{Enable: dist.Bypass.Enable, Cfg: cfg}
	}

	// deviceModel：enable + mqtt 相关配置
	if len(dist.MqttConfig.Enable) > 0 {
		cfg := map[string]string{
			cts.DistCfgQos:            strconv.Itoa(dist.MqttConfig.Qos),
			cts.DistCfgRetain:         strconv.FormatBool(dist.MqttConfig.Retain),
			cts.DistCfgTimeoutConnect: strconv.Itoa(dist.MqttConfig.TimeoutC),
			cts.DistCfgTimeoutRW:      strconv.Itoa(dist.MqttConfig.TimeoutR),
		}
		if dist.MqttConfig.Broker != "" {
			cfg[cts.DistCfgBroker] = dist.MqttConfig.Broker
		}
		if dist.MqttConfig.ClientID != "" {
			cfg[cts.DistCfgClientID] = dist.MqttConfig.ClientID
		}
		result[cts.DistKeyDeviceModel] = &pb.Distributor{Enable: dist.MqttConfig.Enable, Cfg: cfg}
	}

	return result
}

// isDistributorEmpty 判断请求中的 Distributor 是否全空（Enable 和 Cfg 均为空），全空表示删除
func isDistributorEmpty(dist *pb.Distributor) bool {
	return len(dist.Enable) == 0 && len(dist.Cfg) == 0
}

// applyDistributorConfig 应用 distributor 转发配置变更，返回是否有变更
func applyDistributorConfig(distributors map[string]*pb.Distributor) bool {
	changed := false
	for key, dist := range distributors {
		if dist == nil {
			continue
		}
		switch key {
		case cts.DistKeyTlink:
			changed = applyTlinkDistributor(dist) || changed
		case cts.DistKeyBypass:
			changed = applyBypassDistributor(dist) || changed
		case cts.DistKeyDeviceModel:
			changed = applyDeviceModelDistributor(dist) || changed
		default:
			log.Warnf("applyDistributorConfig: 未知的 distributor key: %s", key)
		}
	}
	return changed
}

// applyTlinkDistributor 应用 tlink 转发配置；全空则删除
func applyTlinkDistributor(dist *pb.Distributor) bool {
	if isDistributorEmpty(dist) {
		if len(config.GetRB().Distributor.Tlink.Enable) == 0 {
			return false
		}
		config.GetRB().Update([]string{"distributor", "tlink", "enable"}, []string{})
		config.GetRB().Distributor.Tlink.Enable = nil
		log.Infof("applyTlinkDistributor: 删除 tlink 配置")
		return true
	}
	if dist.Enable != nil && !reflect.DeepEqual(dist.Enable, config.GetRB().Distributor.Tlink.Enable) {
		config.GetRB().Update([]string{"distributor", "tlink", "enable"}, dist.Enable)
		config.GetRB().Distributor.Tlink.Enable = dist.Enable
		log.Infof("applyTlinkDistributor: 更新 enable: %v", dist.Enable)
		return true
	}
	return false
}

// applyBypassDistributor 应用 bypass 旁路转发配置；全空则删除，设置时 target 和 client_id 有默认值
func applyBypassDistributor(dist *pb.Distributor) bool {
	bypass := &config.GetRB().Distributor.Bypass
	if isDistributorEmpty(dist) {
		if len(bypass.Enable) == 0 && bypass.Target == "" && bypass.ClientId == "" {
			return false
		}
		config.GetRB().Update([]string{"distributor", "bypass"}, config.BypassInfo{})
		*bypass = config.BypassInfo{}
		log.Infof("applyBypassDistributor: 删除 bypass 配置")
		return true
	}
	changed := false
	if dist.Enable != nil && !reflect.DeepEqual(dist.Enable, bypass.Enable) {
		config.GetRB().Update([]string{"distributor", "bypass", "enable"}, dist.Enable)
		bypass.Enable = dist.Enable
		log.Infof("applyBypassDistributor: 更新 enable: %v", dist.Enable)
		changed = true
	}
	// target 默认值
	target, hasTarget := dist.Cfg[cts.DistCfgTarget]
	if hasTarget {
		if target == "" {
			target = cts.DistDefaultBypassTarget
		}
		if target != bypass.Target {
			config.GetRB().Update([]string{"distributor", "bypass", cts.DistCfgTarget}, target)
			bypass.Target = target
			log.Infof("applyBypassDistributor: 更新 target: %s", target)
			changed = true
		}
	}
	// client_id 默认值
	clientId, hasClientId := dist.Cfg[cts.DistCfgClientID]
	if hasClientId {
		if clientId == "" {
			clientId = cts.DistDefaultClientID
		}
		if clientId != bypass.ClientId {
			config.GetRB().Update([]string{"distributor", "bypass", cts.DistCfgClientID}, clientId)
			bypass.ClientId = clientId
			log.Infof("applyBypassDistributor: 更新 client_id: %s", clientId)
			changed = true
		}
	}
	return changed
}

// applyDeviceModelDistributor 应用 deviceModel MQTT 转发配置；全空则删除，client_id 有默认值
func applyDeviceModelDistributor(dist *pb.Distributor) bool {
	mqtt := &config.GetRB().Distributor.MqttConfig
	if isDistributorEmpty(dist) {
		if len(mqtt.Enable) == 0 && mqtt.Broker == "" && mqtt.ClientID == "" {
			return false
		}
		config.GetRB().Update([]string{"distributor", "deviceModel"}, config.MqttConfig{})
		*mqtt = config.MqttConfig{}
		log.Infof("applyDeviceModelDistributor: 删除 deviceModel 配置")
		return true
	}
	changed := false
	if dist.Enable != nil && !reflect.DeepEqual(dist.Enable, mqtt.Enable) {
		config.GetRB().Update([]string{"distributor", "deviceModel", "enable"}, dist.Enable)
		mqtt.Enable = dist.Enable
		log.Infof("applyDeviceModelDistributor: 更新 enable: %v", dist.Enable)
		changed = true
	}
	cfg := dist.Cfg
	if cfg == nil {
		return changed
	}
	// 逐字段更新
	type strField struct {
		key  string
		ptr  *string
		path string
		def  string // 默认值，空串表示无默认值
	}
	strFields := []strField{
		{cts.DistCfgBroker, &mqtt.Broker, cts.DistCfgBroker, ""},
		{cts.DistCfgClientID, &mqtt.ClientID, cts.DistCfgClientID, cts.DistDefaultClientID},
	}
	for _, f := range strFields {
		if v, ok := cfg[f.key]; ok {
			if v == "" && f.def != "" {
				v = f.def
			}
			if v != *f.ptr {
				config.GetRB().Update([]string{"distributor", "deviceModel", f.path}, v)
				*f.ptr = v
				log.Infof("applyDeviceModelDistributor: 更新 %s: %s", f.key, v)
				changed = true
			}
		}
	}
	type intField struct {
		key  string
		ptr  *int
		path string
	}
	intFields := []intField{
		{cts.DistCfgQos, &mqtt.Qos, cts.DistCfgQos},
		{cts.DistCfgTimeoutConnect, &mqtt.TimeoutC, cts.DistCfgTimeoutConnect},
		{cts.DistCfgTimeoutRW, &mqtt.TimeoutR, cts.DistCfgTimeoutRW},
	}
	for _, f := range intFields {
		if v, ok := cfg[f.key]; ok {
			if iv, err := strconv.Atoi(v); err == nil && iv != *f.ptr {
				config.GetRB().Update([]string{"distributor", "deviceModel", f.path}, iv)
				*f.ptr = iv
				log.Infof("applyDeviceModelDistributor: 更新 %s: %d", f.key, iv)
				changed = true
			}
		}
	}
	if v, ok := cfg[cts.DistCfgRetain]; ok {
		if bv, err := strconv.ParseBool(v); err == nil && bv != mqtt.Retain {
			config.GetRB().Update([]string{"distributor", "deviceModel", "retain"}, bv)
			mqtt.Retain = bv
			log.Infof("applyDeviceModelDistributor: 更新 retain: %v", bv)
			changed = true
		}
	}
	return changed
}

// ChangeModeHandle 切换运营态
// tone出厂时读取本地文件(deviceModel)，联网后需要切换运营态(tlink)
func ChangeModeHandle(ctx context.Context, req *pb.ChangeModeReq) (*emptypb.Empty, error) {
	source := req.GetSource()
	// 空则默认表示 tlink 切换为运营态
	if source == "" {
		source = cm.TLinkModName
	}

	// 仅支持  tlink 模式切换
	if source != cm.TLinkModName {
		return nil, fmt.Errorf("source 仅支持%s", cm.TLinkModName)
	}

	// 判断当前模式是否已经为目标模式，如果已经是则无需切换
	currentSource := config.GetRB().Project.Source
	if currentSource == source {
		log.Infof("ChangeModeHandle: 当前模式已为 %s，无需切换", source)
		return &emptypb.Empty{}, nil
	}

	// 切换前清理 project 目录（含子目录），切换重启后会重新生成
	if err := cleanProjectDir(); err != nil {
		log.Errorf("ChangeModeHandle: 清理 project 目录失败: %v", err)
		return nil, fmt.Errorf("清理 project 目录失败: %v", err)
	}

	// 构建 SetBasicInfoReq
	setReq := buildChangeModeRequest(source)

	log.Infof("ChangeModeHandle: 切换运营态为 %s, req: %+v", source, setReq)

	// 调用 SetBasicInfoHandle 进行配置切换
	return SetBasicInfoHandle(ctx, setReq)
}

// cleanProjectDir 清理 project 目录（含所有子目录），切换重启后会重新生成
func cleanProjectDir() error {
	projectPath := config.GetRB().GetProjectPath()
	if _, err := stdos.Stat(projectPath); stdos.IsNotExist(err) {
		log.Infof("cleanProjectDir: 目录 %s 不存在，无需清理", projectPath)
		return nil
	}

	log.Infof("cleanProjectDir: 开始删除目录 %s", projectPath)
	if err := stdos.RemoveAll(projectPath); err != nil {
		return fmt.Errorf("remove project dir %s fail: %v", projectPath, err)
	}
	log.Infof("cleanProjectDir: 删除目录 %s 成功", projectPath)
	return nil
}

// buildChangeModeRequest 根据目标模式构建 SetBasicInfoReq
func buildChangeModeRequest(source string) *pb.SetBasicInfoReq {
	req := &pb.SetBasicInfoReq{
		Source:      source,
		Distributor: make(map[string]*pb.Distributor),
	}

	// 所有数据类型
	allDataTypes := []string{"std_change", "std_interval", "collect_change", "collect_interval"}

	if source == cm.TLinkModName {
		// 切换为 tlink 运营态：开启智研上报，开启 tlink/bypass 转发，删除 deviceModel 转发
		req.ReportMonitor = ReportMonitorEnable

		// 若当前 ProxyTarget 为空，则根据 collector IP 列表自动构建
		if config.GetRB().MonitorProxy.ProxyTarget == "" {
			collectorIPs := getCollectorIps()
			if len(collectorIPs) > 0 {
				req.ProxyTarget = buildProxyTarget(strings.Join(collectorIPs, ","))
				log.Infof("buildChangeModeRequest: ProxyTarget is empty, built from collector IPs: %s",
					req.ProxyTarget)
			}
		}

		// 开启 tlink 转发
		req.Distributor[cts.DistKeyTlink] = &pb.Distributor{
			Enable: allDataTypes,
		}

		// 开启 bypass 转发
		req.Distributor[cts.DistKeyBypass] = &pb.Distributor{
			Enable: allDataTypes,
			Cfg: map[string]string{
				cts.DistCfgTarget: cts.DistDefaultBypassTarget,
			},
		}

		// 删除 deviceModel 转发（传空的 Distributor 表示删除）
		req.Distributor[cts.DistKeyDeviceModel] = &pb.Distributor{}
	} else {
		// 切换为 deviceModel 初始态：关闭智研上报，关闭 tlink/bypass 转发，打开 deviceModel 转发
		req.ReportMonitor = ReportMonitorDisable

		// 删除 tlink 转发
		req.Distributor[cts.DistKeyTlink] = &pb.Distributor{}

		// 删除 bypass 转发
		req.Distributor[cts.DistKeyBypass] = &pb.Distributor{}

		// 打开 deviceModel 转发（使用默认配置）
		req.Distributor[cts.DistKeyDeviceModel] = &pb.Distributor{
			Enable: allDataTypes,
			Cfg: map[string]string{
				cts.DistCfgBroker:         cts.DistDefaultDeviceModelBroker,
				cts.DistCfgClientID:       cts.DistDefaultDeviceModelClientID,
				cts.DistCfgQos:            cts.DistDefaultDeviceModelQos,
				cts.DistCfgRetain:         cts.DistDefaultDeviceModelRetain,
				cts.DistCfgTimeoutConnect: cts.DistDefaultDeviceModelTimeoutConnect,
				cts.DistCfgTimeoutRW:      cts.DistDefaultDeviceModelTimeoutRW,
			},
		}
	}

	return req
}
