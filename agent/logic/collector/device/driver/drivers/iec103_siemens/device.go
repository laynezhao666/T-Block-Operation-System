package iec103_siemens

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/consts"
	"agent/entity/definition"
	model3 "agent/entity/model"
	"agent/logic/collector/device/model"
)

// ConnectionManager 连接管理 - 负责TCP/UDP的建立和关闭
type ConnectionManager struct {
	tcp         net.Conn
	udp         *net.UDPConn
	mu          sync.Mutex
	isConnected atomic.Bool
	option      Option
	ip          string
	tcpPort     int
	udpPort     int
	lastRecv    atomic.Int64
}

// IEC103Protocol. 协议层 - 负责报文编解码
type IEC103Protocol struct {
	builder *msgBuilder
	parser  *msgParser
}

// DataCache 数据缓存 - 负责数据存储
type DataCache struct {
	mu   sync.RWMutex
	data map[uint32]*DataPoint
}

// ControlResponse 控制响应结构
type ControlResponse struct {
	Success  bool   // 是否成功
	COT      byte   // 传送原因
	InfoNum  byte   // 信息序号
	GroupNum byte   // 组号
	EntryNum byte   // 条目号
	Data     []byte // 响应数据
}

// PendingRequest 等待响应的请求信息
type PendingRequest struct {
	ExpectCOT []byte                // 期望的COT列表
	GroupNum  byte                  // 组号
	EntryNum  byte                  // 条目号
	RespChan  chan *ControlResponse // 响应通道
}

// Device 设备 - 组合上面的组件
type Device struct {
	gid      definition.DeviceGidType
	name     string
	conn     *ConnectionManager
	protocol *IEC103Protocol
	cache    *DataCache

	mu     sync.Mutex
	isOpen atomic.Bool
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 控制响应管理 - 支持多协程并发等待
	respMu       sync.Mutex
	pendingResps []*PendingRequest // 等待响应的请求列表
}

// NewDevice 创建设备实例
func NewDevice(gid definition.DeviceGidType, name string) *Device {
	return &Device{
		gid:  gid,
		name: name,
		conn: &ConnectionManager{},
		protocol: &IEC103Protocol{
			builder: new(msgBuilder),
			parser:  new(msgParser),
		},
		cache: &DataCache{
			data: make(map[uint32]*DataPoint),
		},
		pendingResps: make([]*PendingRequest, 0),
	}
}

// Open 打开设备通道
func (d *Device) Open(chanInfo model.ChannelInfo, packets model.ListCollectPackets) consts.Quality {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.IsOpen() {
		log.Warnf("device %s: already opened", d.name)
		return consts.QualityOk
	}
	d.conn.option.Load(chanInfo, packets)
	// 解析设备地址和IP配置
	if err := d.parseConfig(chanInfo); err != nil {
		log.Warnf("%s iec103 parse config failed: %v", d.name, err)
		return consts.QualityConfigError
	}
	log.Infof("%s option:%+v", d.name, d.conn.option)

	// 建立TCP连接
	if err := d.connectTCP(); err != nil {
		log.Warnf("%s: connect tcp failed: %v, target:%s:%d", d.name, err, d.conn.ip, d.conn.tcpPort)
		return consts.QualityDriverOpenFailed
	}

	// 建立UDP连接
	if err := d.connectUDP(); err != nil {
		log.Warnf("%s: connect udp failed: %v, target:%s:%d", d.name, err, d.conn.ip, d.conn.udpPort)
		d.closeTcp()
		return consts.QualityDriverOpenFailed
	}
	// 创建上下文用于协程管理
	ctx, cancel := context.WithCancel(context.Background())
	d.ctx = ctx
	d.cancel = cancel
	cur := time.Now().UnixMilli()
	// 初始化为当前时间
	d.conn.lastRecv.Store(cur)
	d.isOpen.Store(true)
	d.wg.Add(1)
	go func(ctx context.Context) {
		defer d.wg.Done()
		d.tcpReceiveRoutine(ctx)
	}(ctx)
	d.wg.Add(1)
	go func(ctx context.Context) {
		defer d.wg.Done()
		d.udpReceiveRoutine(ctx)
	}(ctx)
	d.periodicCallRoutine(ctx)
	return consts.QualityOk
}

// Close 关闭设备通道
func (d *Device) Close() consts.Quality {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.isOpen.Store(false)
	// 停止所有协程
	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}
	d.closeTcp()
	d.closeUdp()
	d.wg.Wait()
	return consts.QualityOk
}

// parseAddr 将0x201格式的字符串地址转换为uint32
func parseAddr(addrStr string) (uint32, error) {
	if len(addrStr) < 3 {
		return 0, fmt.Errorf("invalid address len: %s", addrStr)
	}
	if addrStr[:2] == "0x" || addrStr[:2] == "0X" {
		// 解析十六进制字符串
		addr, err := strconv.ParseUint(addrStr[2:], 16, 32)
		if err != nil {
			return 0, fmt.Errorf("failed to parse address %s: %v", addrStr, err)
		}
		return uint32(addr), nil
	} else {
		// 解析十进制字符串
		addr, err := strconv.ParseUint(addrStr, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("failed to parse address %s: %v", addrStr, err)
		}
		return uint32(addr), nil
	}
}

// Request 发送采集请求
func (d *Device) Request(ctx context.Context,
	packet *model.CollectProtocolPacket) (consts.Quality, model3.MessageStatistics) {
	stats := model3.MessageStatistics{SendCount: 0, SuccessCount: 0}

	if packet == nil {
		return consts.QualityOk, stats
	}
	stats.SendCount = uint64(len(packet.Points))
	// 设备没有打开或者连接已断开直接返回失败
	if !d.IsOpen() || !d.IsConnected() {
		return consts.QualityCommDisconnected, stats
	}
	// 正常每秒都会收到数据，如果超过一定时间没有收到数据则认为连接已断开
	now := time.Now().UnixMilli()
	if now > d.conn.lastRecv.Load()+int64(d.conn.option.ExpiredBuf) {
		return consts.QualityCommDisconnected, stats
	}
	d.cache.mu.RLock()
	defer d.cache.mu.RUnlock()
	for _, p := range packet.Points {
		valueParser, ok := p.Attr.ValParser.(*ValueParser)
		if !ok || valueParser == nil {
			p.RtVal.Qua = consts.QualityUncertain
			continue
		}
		// 将0x201格式的valueParser.Addr转换为uint32
		addr, err := parseAddr(valueParser.Addr)
		if err != nil {
			log.Warnf("failed to parse address %s: %v", valueParser.Addr, err)
			p.RtVal.Qua = consts.QualityAddrError
			continue
		}
		pointVal, ok := d.cache.data[addr]
		if !ok {
			p.RtVal.Qua = consts.QualityUncollected
			continue
		}
		if pointVal.ExpiredMs > 0 && now > pointVal.ExpiredMs { // 超过过期时间认为数据过期
			p.RtVal.Qua = consts.QualityValueTmsExpired
		} else {
			p.RtVal.Qua = pointVal.Qua
		}
		if p.RtVal.Qua == consts.QualityOk {
			if err := parseVariable(pointVal, valueParser, p); err != nil {
				continue
			}
		}
		if p.RtVal.Qua == consts.QualityOk {
			stats.SuccessCount++
		}
	}
	return consts.QualityOk, stats
}

// RequestPing 发送心跳检测,注意心跳不等待接收响应，只要收到主动上报就认为心跳正常
func (d *Device) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	if !d.IsOpen() {
		return consts.QualityCommDisconnected
	}

	heartbeatData := d.protocol.builder.BuildHeartbeat()

	if err := d.sendTCPRequest(heartbeatData); err != nil {
		return consts.QualityCmdSendError
	}
	return consts.QualityOk
}

// Control 发送控制指令
func (d *Device) Control(packet *model.ControlProtocolPacket, val string) consts.Quality {
	if !d.IsOpen() {
		return consts.QualityCommDisconnected
	}
	if !d.IsConnected() {
		return consts.QualityCommDisconnected
	}
	if packet == nil || packet.Point == nil {
		log.Warnf("%s control packet or point is nil", d.name)
		return consts.QualityConfigError
	}

	// 获取ValueParser来获取地址
	valueParser, ok := packet.Point.Attr.ValParser.(*ValueParser)
	if !ok || valueParser == nil {
		log.Warnf("%s control value parser is nil or invalid", d.name)
		return consts.QualityConfigError
	}

	// 从地址解析组号和条目号
	addr, err := parseAddr(valueParser.Addr)
	if err != nil {
		log.Warnf("%s control parse addr failed: %v", d.name, err)
		return consts.QualityAddrError
	}

	// 反向解析组号和条目号
	groupNum, entryNum := ParseAddrToGroupEntry(addr)

	// 解析信息序号，使用Command字段
	infoNum, err := parseInfoNum(packet.Command)
	if err != nil {
		log.Warnf("%s control parse info num failed: %v", d.name, err)
		return consts.QualityConfigError
	}

	// 解析控制值，参考modbus的方式转换枚举值
	controlValue, err := parseControlValue(val)
	if err != nil {
		log.Warnf("%s control parse value failed: %v", d.name, err)
		return consts.QualityConfigError
	}

	log.Warnf("%s control: groupNum=%d, entryNum=%d, infoNum=0x%02X, value=%d",
		d.name, groupNum, entryNum, infoNum, controlValue)

	// 获取响应超时时间
	respTimeout := time.Duration(d.conn.option.ControlRespTimeout) * time.Millisecond
	if respTimeout <= 0 {
		respTimeout = 10000 * time.Millisecond // 默认10秒超时
	}

	data := d.protocol.builder.BuildControlSelectWithInfoNum(groupNum, entryNum, controlValue, infoNum)
	resp, err := d.sendAndWaitResponse(data, groupNum, entryNum, respTimeout)
	if err != nil {
		log.Warnf("%s control send failed: %v", d.name, err)
		return consts.QualityCmdSendError
	}
	if resp == nil {
		log.Warnf("%s control response timeout, infoNum=0x%02X", d.name, infoNum)
		return consts.QualityCmdRespTimeout
	}
	if !resp.Success {
		log.Warnf("%s control response failed, COT=0x%02X, infoNum=0x%02X", d.name, resp.COT, infoNum)
		return consts.QualityCmdRespError
	}
	log.Debugf("%s control success with infoNum=0x%02X", d.name, infoNum)

	return consts.QualityOk
}

// parseInfoNum 解析信息序号（从Command字段）
// 支持十六进制格式（如"0xF9"）和十进制格式
func parseInfoNum(command string) (byte, error) {
	if command == "" {
		return 0, fmt.Errorf("command is empty")
	}
	command = strings.TrimSpace(command)
	if len(command) >= 2 && (command[:2] == "0x" || command[:2] == "0X") {
		val, err := strconv.ParseUint(command[2:], 16, 8)
		if err != nil {
			return 0, fmt.Errorf("failed to parse hex command %s: %v", command, err)
		}
		return byte(val), nil
	}
	val, err := strconv.ParseUint(command, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("failed to parse command %s: %v", command, err)
	}
	return byte(val), nil
}

// parseControlValue 解析控制值
// 支持多种格式："0"/"1", "off"/"on", "分"/"合", 数字字符串
func parseControlValue(val string) (byte, error) {
	if val == "" {
		return 0, fmt.Errorf("control value is empty")
	}
	val = strings.TrimSpace(val)
	lowerVal := strings.ToLower(val)

	// 支持常见的布尔表示方式
	switch lowerVal {
	case "1", "off", "分":
		return ControlValue_Off, nil
	case "2", "on", "合":
		return ControlValue_On, nil
	}

	// 尝试解析为数字
	if len(val) >= 2 && (val[:2] == "0x" || val[:2] == "0X") {
		// 十六进制格式
		parsedVal, err := strconv.ParseUint(val[2:], 16, 8)
		if err != nil {
			return 0, fmt.Errorf("failed to parse hex control value %s: %v", val, err)
		}
		return byte(parsedVal), nil
	}

	// 十进制格式
	parsedVal, err := strconv.ParseUint(val, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("failed to parse control value %s: %v", val, err)
	}
	return byte(parsedVal), nil
}

// sendTotalCall 发送定期总召唤请求（不等待回复）
func (d *Device) sendTotalCall() consts.Quality {
	if !d.IsOpen() {
		return consts.QualityCommDisconnected
	}

	// 构建总召唤报文
	generalInterrogationData := d.protocol.builder.BuildTotalCall()

	// 直接发送总召唤请求，不等待回复
	if err := d.sendTCPRequest(generalInterrogationData); err != nil {
		log.Warnf("%s iec103 periodic total call send failed: %v", d.name, err)
		return consts.QualityCmdSendError
	}

	log.Debugf("%s iec103 periodic total call sent", d.name)
	return consts.QualityOk
}

// sendEnergyCall 发送定期电度召唤请求（不等待回复）
func (d *Device) sendEnergyCall() consts.Quality {
	if !d.IsOpen() {
		return consts.QualityCommDisconnected
	}

	// 构建电度召唤报文
	energyRequestData := d.protocol.builder.BuildEnergyCall()

	// 直接发送电度召唤请求，不等待回复
	if err := d.sendTCPRequest(energyRequestData); err != nil {
		log.Warnf("%s iec103 periodic energy call send failed: %v", d.name, err)
		return consts.QualityCmdSendError
	}

	log.Debugf("%s iec103 periodic energy call sent", d.name)
	return consts.QualityOk
}

// sendClockSync 发送时间同步请求（不等待回复）
func (d *Device) sendClockSync() consts.Quality {
	if !d.IsOpen() {
		return consts.QualityCommDisconnected
	}

	// 根据时区配置获取正确的时间
	var syncTime time.Time
	if d.conn.option.Timezone != 0 {
		// 使用数字时区偏移量计算时间
		syncTime = time.Now().UTC().Add(time.Duration(d.conn.option.Timezone) * time.Hour)
	} else {
		// 默认使用系统时区
		syncTime = time.Now()
	}

	// 构建时间同步报文
	clockSyncRequestData := d.protocol.builder.BuildClockSync(syncTime, d.conn.option.SummerTime > 0)

	// 直接发送时间同步请求，不等待回复
	if err := d.sendTCPRequest(clockSyncRequestData); err != nil {
		log.Warnf("%s iec103 periodic clock sync send failed: %v", d.name, err)
		return consts.QualityCmdSendError
	}

	log.Debugf("%s iec103 periodic clock sync sent with timezone offset: %+d hours, time: %s",
		d.name, d.conn.option.Timezone, syncTime.Format("2006-01-02 15:04:05"))
	return consts.QualityOk
}

// periodicTask 定时任务定义
type periodicTask struct {
	name     string
	interval time.Duration
	lastRun  time.Time
	handler  func() consts.Quality
}

// periodicCallRoutine 定期召唤协程，使用更优雅的定时任务管理
func (d *Device) periodicCallRoutine(ctx context.Context) {
	// 初始化定时任务列表
	tasks := []*periodicTask{
		{
			name:     "total_call",
			interval: time.Duration(d.conn.option.TotalCallInterval) * time.Millisecond,
			handler:  d.sendTotalCall,
		},
		{
			name:     "energy_call",
			interval: time.Duration(d.conn.option.ElecCallInterval) * time.Millisecond,
			handler:  d.sendEnergyCall,
		},
		{
			name:     "clock_sync",
			interval: time.Duration(d.conn.option.ClockSyncIntvl) * time.Millisecond,
			handler:  d.sendClockSync,
		},
		{
			name:     "heartbeat",
			interval: time.Duration(d.conn.option.HeartbeatIntvl) * time.Millisecond,
			handler:  func() consts.Quality { return d.RequestPing(ctx, model.CollectProtocolPacket{}) },
		},
		{
			name:     "expired",
			interval: time.Hour,
			handler:  d.expired,
		},
	}

	// 过滤掉间隔为0的任务（表示不启用）
	activeTasks := make([]*periodicTask, 0, len(tasks))
	for _, task := range tasks {
		if task.interval > 0 {
			activeTasks = append(activeTasks, task)
		}
	}

	if len(activeTasks) == 0 {
		log.Infof("device %s: no periodic tasks configured", d.name)
		return
	}

	interval := 300 * time.Millisecond
	// 立即执行一次所有任务作为初始化
	d.executeTasks(ctx, activeTasks, true)
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		// 使用ticker进行定期检查
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !d.IsOpen() {
					return
				}
				d.executeTasks(ctx, activeTasks, false)
			}
		}
	}()
}

// expired 对cache里的数据进行过期处理
func (d *Device) expired() consts.Quality {
	now := time.Now().UnixMilli()
	expiredCount := 0

	d.cache.mu.Lock()
	defer d.cache.mu.Unlock()

	for addr, point := range d.cache.data {
		if point.ExpiredMs > 0 && now > point.ExpiredMs {
			delete(d.cache.data, addr)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		log.Debugf("device %s: expired %d data points from cache", d.name, expiredCount)
	}

	return consts.QualityOk
}

// executeTasks 执行定时任务
func (d *Device) executeTasks(ctx context.Context, tasks []*periodicTask, isInitial bool) {
	now := time.Now()
	hasSent := false

	for _, task := range tasks {
		// 如果是初始化阶段，或者任务间隔已到且没有发送过其他任务
		if isInitial || (now.Sub(task.lastRun) >= task.interval && !hasSent) {
			if quality := task.handler(); quality != consts.QualityOk {
				log.Warnf("device %s %s failed: %v", d.name, task.name, quality)
			}
			task.lastRun = now
			hasSent = true
			if isInitial {
				log.Infof("device exec %s %s", d.name, task.name)
			}

			// 如果不是心跳报文，则都紧跟着发送心跳报文，因为实际验证中发现如果不跟着发送心跳，设备看起来会把下一条指令忽略
			if task.name != "heartbeat" {
				if heartbeatQuality := d.RequestPing(ctx, model.CollectProtocolPacket{}); heartbeatQuality != consts.QualityOk {
					log.Warnf("device %s heartbeat after %s failed: %v", d.name, task.name, heartbeatQuality)
				}
				// 如果是初始化阶段，发送完报文后休眠一定时间防止设备处理不过来
				if isInitial {
					time.Sleep(300 * time.Millisecond)
				}
			}
		}
	}
}

// GetData 获取数据（用于测试）
func (d *Device) GetData(addr uint32) (interface{}, bool) {
	d.cache.mu.RLock()
	defer d.cache.mu.RUnlock()
	v, has := d.cache.data[addr]
	return v, has
}

// IsConnected 检查连接状态
func (d *Device) IsConnected() bool {
	return d.conn.isConnected.Load()
}

// IsOpen 检查设备是否打开
func (d *Device) IsOpen() bool {
	return d.isOpen.Load()
}

// ========== 内部方法 ==========

// parseConfig 解析设备配置
func (d *Device) parseConfig(chanInfo model.ChannelInfo) error {
	// 解析IP地址和端口
	if items := strings.Split(chanInfo.Name, ":"); len(items) > 1 {
		d.conn.ip = items[0]
		port, err := strconv.Atoi(items[1])
		if err != nil {
			log.Warnf("invalid tcp_port: %s", items[1])
			return err
		}
		d.conn.tcpPort = port
	} else {
		d.conn.ip = chanInfo.Name
		d.conn.tcpPort = 6000
	}
	d.conn.udpPort = 6001
	if len(chanInfo.ExtendKV) > 0 {
		if v, has := chanInfo.ExtendKV["tcp_port"]; has {
			tmp, err := strconv.Atoi(v)
			if err != nil {
				log.Warnf("invalid tcp_port: %s", v)
				return err
			}
			d.conn.tcpPort = tmp
		}
		if v, has := chanInfo.ExtendKV["udp_port"]; has {
			tmp, err := strconv.Atoi(v)
			if err != nil {
				log.Warnf("iec103 invalid tcp_port: %s", v)
				return err
			}
			d.conn.udpPort = tmp
		}
	}
	// 解析设备地址
	addr, err := strconv.ParseUint(chanInfo.Address, 10, 16)
	if err != nil {
		log.Warnf("iec103 invalid device address: %s", chanInfo.Address)
		return fmt.Errorf("invalid device address: %s", chanInfo.Address)
	}
	d.protocol.builder.dstDeviceAddr = uint16(addr)
	d.protocol.builder.srcDeviceAddr = 0x0A0A
	d.protocol.builder.dstStationAddr = 0x0000 // 默认厂站地址
	d.protocol.builder.srcStationAddr = 0x0000
	// 如果有多个采集器连接一个设备，则需要配置这个地区进行区分
	if v, has := chanInfo.ExtendKV["src_device_addr"]; has {
		tmp, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			log.Warnf("invalid src_device_addr: %s", v)
			return err
		}
		d.protocol.builder.srcDeviceAddr = uint16(tmp)
	} else {
		// 随机生成一个地址赋值给srcDeviceAddr，避免重连时因为相同的srcDeviceAddr导致连接失败
		// 使用当前时间纳秒作为随机种子，生成范围在0x1000-0xFFFE之间的地址
		// 避免使用0x0000（可能被保留）和0xFFFF（广播地址）
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		d.protocol.builder.srcDeviceAddr = uint16(rng.Intn(0xEFFE) + 0x1000)
		log.Infof("%s srcDeviceAddr: 0x%04X", d.name, d.protocol.builder.srcDeviceAddr)
	}
	return nil
}

// connectTCP 建立TCP连接
func (d *Device) connectTCP() error {
	d.conn.mu.Lock()
	defer d.conn.mu.Unlock()
	if d.conn.tcp != nil {
		d.conn.tcp.Close()
		d.conn.tcp = nil
	}

	// 使用带超时的Dial，避免长时间阻塞
	address := fmt.Sprintf("%s:%d", d.conn.ip, d.conn.tcpPort)
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		log.Warnf("device %s: connectTCP failed: %v, target: %s", d.name, err, address)
		return err
	}
	d.conn.tcp = conn
	return nil
}

func (d *Device) closeTcp() {
	d.conn.mu.Lock()
	defer d.conn.mu.Unlock()
	if d.conn.tcp != nil {
		d.conn.tcp.Close()
		d.conn.tcp = nil
	}
}

// connectUDP 建立UDP连接
func (d *Device) connectUDP() error {
	d.conn.mu.Lock()
	defer d.conn.mu.Unlock()
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", d.conn.ip, d.conn.udpPort))
	if err != nil {
		log.Warnf("%s connectUDP resolve address failed: %v", d.name, err)
		return err
	}
	if d.conn.udp != nil {
		d.conn.udp.Close()
		d.conn.udp = nil
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Warnf("%s connectUDP failed: %v", d.name, err)
		return err
	}
	d.conn.udp = conn
	log.Debugf("%s connectUDP success: %s:%d", d.name, d.conn.ip, d.conn.udpPort)
	return nil
}

func (d *Device) closeUdp() {
	d.conn.mu.Lock()
	defer d.conn.mu.Unlock()
	if d.conn.udp != nil {
		d.conn.udp.Close()
		d.conn.udp = nil
	}
}

const minConnIntvlMs = 500
const defaultSleep = minConnIntvlMs * time.Millisecond

func (d *Device) tcpRead(buffer []byte) (int, error) {
	var tmp net.Conn
	d.conn.mu.Lock()
	tmp = d.conn.tcp
	d.conn.mu.Unlock()
	if tmp == nil {
		return 0, fmt.Errorf("tcp connection is nil")
	}
	// 设置读超时，避免永久阻塞
	if err := tmp.SetReadDeadline(time.Now().Add(time.Duration(d.conn.option.ReadTimeOut) * time.Millisecond)); err != nil {
		log.Debugf("device %s: set read deadline failed: %v", d.name, err)
	}
	n, err := tmp.Read(buffer)
	// 清除读超时
	_ = tmp.SetReadDeadline(time.Time{})
	return n, err
}

// tcpReceiveRoutine TCP数据接收协程
func (d *Device) tcpReceiveRoutine(ctx context.Context) {
	// 读取报文头的缓冲区
	headerBuffer := make([]byte, IEC103_FixHeadLen)

	for d.IsOpen() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// 第一步：读取固定长度的报文头
		begin := time.Now()
		n, err := d.tcpRead(headerBuffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Infof("%s read timeout:%v", d.name, time.Now().Sub(begin))
			} else if d.isConnectionClosedError(err) {
				log.Warnf("%s read close: %v", d.name, err)
				d.conn.isConnected.Store(false)
			} else {
				log.Warnf("%s read error: %v", d.name, err)
			}
			if time.Now().Sub(begin) < defaultSleep {
				time.Sleep(defaultSleep)
			}
			continue
		}
		d.conn.isConnected.Store(true)
		if n != IEC103_FixHeadLen {
			log.Warnf("%s read tcp packet header error, length: %d,% X", d.name, n, headerBuffer[:n])
			// 读取的报文头长度不正确，跳过
			continue
		}
		// 解析报文长度
		length := binary.LittleEndian.Uint32(headerBuffer[2:6])
		if length <= 0 {
			log.Warnf("%s invalid packet length: %d", d.name, length)
			continue
		}
		// 避免length过大，做一下判断限制
		if length > d.conn.option.MaxPacketSize {
			log.Warnf("%s packet length too large: %d", d.name, length)
			continue
		}
		packetBuffer := make([]byte, IEC103_FixHeadLen+length)
		copy(packetBuffer, headerBuffer)
		n, err = d.tcpRead(packetBuffer[IEC103_FixHeadLen:])
		if err != nil {
			log.Errorf("%s read tcp packet error: %v", d.name, err)
			continue
		}

		if n != int(length) {
			log.Warnf("%s read tcp packet error, length: %d,% X", d.name, n, packetBuffer[:n])
			continue
		}
		d.conn.lastRecv.Store(time.Now().UnixMilli())
		d.Handle(packetBuffer)
	}
}

// udpReceiveRoutine UDP数据接收协程
func (d *Device) udpReceiveRoutine(ctx context.Context) {
	buffer := make([]byte, 4096)

	for d.IsOpen() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if d.conn.udp == nil {
			time.Sleep(defaultSleep)
			continue
		}

		// 修复：设置UDP读超时，避免永久阻塞
		_ = d.conn.udp.SetReadDeadline(time.Now().Add(time.Duration(d.conn.option.ReadTimeOut) * time.Millisecond))
		n, err := d.conn.udp.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			time.Sleep(defaultSleep)
			continue
		}

		if n > 0 {
			d.Handle(buffer[:n])
		}
	}
}

// findAndRemovePendingRequest 查找并移除匹配的等待请求
// 根据COT、组号和条目号进行匹配，找到后从列表中删除并返回
// 内部已加锁，调用者无需关心锁的管理
func (d *Device) findAndRemovePendingRequest(recvCOT, recvGroupNum, recvEntryNum byte) *PendingRequest {
	d.respMu.Lock()
	defer d.respMu.Unlock()

	for i, req := range d.pendingResps {
		// 检查组号和条目号是否匹配
		if req.GroupNum == recvGroupNum && req.EntryNum == recvEntryNum {
			// 检查COT是否在期望列表中
			for _, expectCOT := range req.ExpectCOT {
				if expectCOT == recvCOT {
					// 找到匹配后立即从列表中删除
					d.pendingResps = append(d.pendingResps[:i], d.pendingResps[i+1:]...)
					return req
				}
			}
		}
	}
	return nil
}

// removePendingRequest 从等待列表中移除指定的请求
// 内部已加锁，调用者无需关心锁的管理
func (d *Device) removePendingRequest(req *PendingRequest) {
	d.respMu.Lock()
	defer d.respMu.Unlock()

	for i, r := range d.pendingResps {
		if r == req {
			d.pendingResps = append(d.pendingResps[:i], d.pendingResps[i+1:]...)
			return
		}
	}
}

// addPendingRequest 添加等待请求到列表
// 内部已加锁，调用者无需关心锁的管理
func (d *Device) addPendingRequest(req *PendingRequest) {
	d.respMu.Lock()
	defer d.respMu.Unlock()

	d.pendingResps = append(d.pendingResps, req)
}

// tryHandleResponsePacket 尝试处理等待响应的报文
// 如果是等待中的响应报文则处理并返回true，否则返回false
// 通过COT、组号和条目号来匹配响应报文
func (d *Device) tryHandleResponsePacket(data []byte) bool {
	if len(data) < IEC103_AsduOffset+IEC103_AsduFixLen+3 {
		return false
	}

	// 解析报文中的COT、组号和条目号
	recvCOT := data[IEC103_CotOffset]
	// 组号和条目号在ASDU固定部分之后：ASDU偏移 + 8字节固定部分后
	groupEntryOffset := IEC103_AsduOffset + 8
	recvGroupNum := data[groupEntryOffset]
	recvEntryNum := data[groupEntryOffset+1]

	// 查找匹配的等待请求，找到后直接从列表中删除
	matchedReq := d.findAndRemovePendingRequest(recvCOT, recvGroupNum, recvEntryNum)

	if matchedReq == nil {
		return false
	}

	// 解析响应报文
	resp := d.parseControlResponse(data)
	if resp == nil {
		log.Warnf("%s parse control response failed for group=%d, entry=%d", d.name, recvGroupNum, recvEntryNum)
		return true // 虽然解析失败，但确实是等待的响应报文
	}

	// 发送响应到对应通道（非阻塞）
	select {
	case matchedReq.RespChan <- resp:
		log.Debugf("%s control response sent to channel, success=%v, COT=0x%02X, group=%d, entry=%d",
			d.name, resp.Success, resp.COT, resp.GroupNum, resp.EntryNum)
	default:
		log.Warnf("%s control response channel full for group=%d, entry=%d, discarded",
			d.name, resp.GroupNum, resp.EntryNum)
	}

	return true
}

// parseControlResponse 解析控制响应报文
func (d *Device) parseControlResponse(data []byte) *ControlResponse {
	if len(data) < IEC103_AsduOffset+IEC103_AsduFixLen+3 {
		return nil
	}

	resp := &ControlResponse{
		Data: data,
	}

	// 获取传送原因 COT
	resp.COT = data[IEC103_CotOffset]

	// 获取信息序号
	// ASDU结构: TypeID(1) + VSQ(1) + COT(1) + CommonAddr(1) + FunType(1) + InfNum(1)
	infNumOffset := IEC103_AsduOffset + 5
	resp.InfoNum = data[infNumOffset]

	// 获取组号和条目号
	// 组号和条目号在ASDU固定部分之后：ASDU偏移 + 8字节固定部分后
	groupEntryOffset := IEC103_AsduOffset + 8
	resp.GroupNum = data[groupEntryOffset]
	resp.EntryNum = data[groupEntryOffset+1]

	// 根据COT判断是否成功
	// COT_Write_Ack (0x2C) 表示写命令确认成功
	// COT_Write (0x28) 可能是请求或肯定确认
	// COT_WriteFail (0x29) 表示写失败
	resp.Success = resp.COT == COT_Write_Ack || resp.COT == COT_Write

	return resp
}

// sendAndWaitResponse 发送请求并等待响应
// 通过COT、组号和条目号来识别响应报文
func (d *Device) sendAndWaitResponse(data []byte, groupNum, entryNum byte, timeout time.Duration) (*ControlResponse, error) {
	// 创建该请求专属的响应通道
	respChan := make(chan *ControlResponse, 1)

	// 创建等待请求，期望的COT为写确认或写失败
	req := &PendingRequest{
		ExpectCOT: []byte{COT_Write_Ack, COT_Write, COT_WriteFail},
		GroupNum:  groupNum,
		EntryNum:  entryNum,
		RespChan:  respChan,
	}

	// 注册到等待响应的列表中
	d.addPendingRequest(req)

	log.Debugf("%s sending control request with group=%d, entry=%d", d.name, groupNum, entryNum)

	// 确保退出时从列表中移除
	defer d.removePendingRequest(req)

	// 发送请求
	if err := d.sendTCPRequest(data); err != nil {
		return nil, err
	}
	// 发送一个心跳
	d.RequestPing(context.Background(), model.CollectProtocolPacket{})

	// 等待响应或超时
	select {
	case resp := <-respChan:
		return resp, nil
	case <-time.After(timeout):
		log.Warnf("%s control request timeout, group=%d, entry=%d", d.name, groupNum, entryNum)
		return nil, nil // 超时返回nil响应
	}
}

// Handle 处理数据
func (d *Device) Handle(data []byte) []*DataPoint {
	// 解析数据
	report := ParseActiveReport(data)
	if report == nil {
		if isHearbeat(data) {
			log.Debugf("receive heartbeat")
		} else {
			log.Warnf("%s parse active report failed, data: % X", d.name, data)
		}
		return nil
	}

	// 根据解析类型处理数据
	return d.processActiveReport(report)
}

// processActiveReport 处理主动上报数据
func (d *Device) processActiveReport(report *ActiveReport) []*DataPoint {
	if len(report.Data) == 0 {
		return nil
	}
	var result []*DataPoint

	switch report.Type {
	case Spontaneous_communication:
		result = d.handleMeasurement(report.Data, d.conn.option.TotalCallInterval)
	case Cyclic_measurement:
		result = d.handleMeasurement(report.Data, d.conn.option.ElecCallInterval)
	case Call_data:
		result = d.handleMeasurement(report.Data, d.conn.option.TotalCallInterval)
	case Call_energy:
		result = d.handleMeasurement(report.Data, d.conn.option.ElecCallInterval)
	case Write_Rsp:
		d.tryHandleResponsePacket(report.Data)
		//result = d.handleMeasurement(report.Data, d.conn.option.TotalCallInterval)
	case Query_End:
		// do nothing
	case Fault_data:
	// todo 处理故障数据
	default:
		log.Warnf("%s unknown report type: % X", d.name, report.Data)
		result = d.handleMeasurement(report.Data, d.conn.option.ElecCallInterval)
	}
	return result
}

func (d *Device) handleMeasurement(data []byte, expireInterval int) []*DataPoint {
	points, err := d.protocol.parser.ParseCommunication(data)
	if err != nil {
		return nil
	}
	cur := time.Now().UnixMilli()
	expired := cur + int64(expireInterval+d.conn.option.ExpiredBuf)

	d.cache.mu.Lock()
	defer d.cache.mu.Unlock()

	for _, point := range points {
		point.Ms = cur
		point.ExpiredMs = expired
		d.cache.data[point.Addr] = point
	}
	return points
}

// handleFaultData 处理故障数据
func (d *Device) handleFaultData(data []byte) {
	// 检查数据长度
	if len(data) < 34 {
		log.Warnf("fault data too short: %d bytes", len(data))
		return
	}

	// 解析ASDU类型和传送原因
	asduType := data[28]
	// cot := data[30]
	// d.dataCacheLock.Lock()
	// defer d.dataCacheLock.Unlock()
	// 根据ASDU类型处理不同的故障数据
	switch asduType {
	case 0x17: // 扰动表
		if len(data) >= 40 {
			faultNumber := binary.LittleEndian.Uint16(data[34:36])
			// 将故障编号存入缓存，使用特殊的地址标识
			faultAddr := uint32(0xFF000000) | uint32(faultNumber)
			//d.dataCache[faultAddr] = faultNumber
			log.Debugf("fault table data received, fault addr:%v fault number: %v",
				faultAddr, faultNumber)
		}
	case 0x1A: // 扰动数据传输准备就绪
		if len(data) >= 40 {
			faultNumber := binary.LittleEndian.Uint16(data[34:36])
			// 将故障传输状态存入缓存
			faultAddr := uint32(0xFF010000) | uint32(faultNumber)
			// d.dataCache.Store(faultAddr, "ready")
			log.Debugf("fault data transfer ready, fault addr:%v fault number: %v",
				faultAddr, faultNumber)
		}
	default:
		log.Warnf("unknown fault data ASDU type: 0x%02X", asduType)
	}
}

// sendTCPRequest 发送TCP请求（支持超时和失败统计）
func (d *Device) sendTCPRequest(data []byte) error {
	d.conn.mu.Lock()
	defer d.conn.mu.Unlock()

	if d.conn.tcp == nil {
		log.Warnf("device %s: TCP nil", d.name)
		return fmt.Errorf("TCP nil")
	}
	deadline := time.Now().Add(time.Duration(d.conn.option.ReadTimeOut) * time.Millisecond)
	if err := d.conn.tcp.SetWriteDeadline(deadline); err != nil {
		log.Warnf("device %s: set write deadline failed: %v", d.name, err)
	}

	_, err := d.conn.tcp.Write(data)

	// 清除写超时
	if clearErr := d.conn.tcp.SetWriteDeadline(time.Time{}); clearErr != nil {
		log.Debugf("device %s: clear write deadline failed: %v", d.name, clearErr)
	}

	if err != nil {
		// 判断是否为连接断开错误
		if d.isConnectionClosedError(err) {
			log.Warnf("%s TCP connection closed: %v", d.name, err)
			d.conn.isConnected.Store(false)
		} else {
			log.Warnf("%s TCP write failed: %v", d.name, err)
		}
		return err
	}
	d.conn.isConnected.Store(true)
	return nil
}

// isConnectionClosedError 判断是否为连接断开错误
func (d *Device) isConnectionClosedError(err error) bool {
	if err == nil {
		return false
	}

	// 检查常见的连接断开错误类型
	if netErr, ok := err.(net.Error); ok {
		// 网络超时错误
		if netErr.Timeout() {
			return false
		}
	}

	// 检查连接关闭相关的错误字符串
	errorStr := err.Error()
	return strings.Contains(errorStr, "broken pipe") ||
		strings.Contains(errorStr, "connection reset") ||
		strings.Contains(errorStr, "closed") ||
		strings.Contains(errorStr, "EOF") ||
		strings.Contains(errorStr, "reset by peer") ||
		strings.Contains(errorStr, "forcibly closed")
}

func parseVariable(data *DataPoint, valParser *ValueParser, point *model.PointInfo) error {
	if data == nil || valParser == nil || point == nil {
		return fmt.Errorf("invalid params")
	}
	point.RtVal.Pv.SetValue(data.Value)
	return nil
}
