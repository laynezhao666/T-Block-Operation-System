// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"dac/entity/consts"
	"dac/entity/model/driver"
	"dac/entity/model/driver/cacs"
	"dac/entity/model/marshaller"
	consts2 "dac/logic/collect/driver/cacs/consts"

	"dac/entity/utils/rrpc"
)

// Controller CACS门控器控制器，管理与门控器的TCP连接和通信。
// 通过被动监听模式接收设备连接，使用MAC地址作为设备唯一标识。
type Controller struct {
	timeout      time.Duration              // 请求超时时间
	baseInfo     driver.ControllerBasicInfo // 控制器基本信息
	chanInfo     driver.ChannelInfo         // 通道信息
	version      string                     // 协议版本
	Server       *DoorServer                // TCP通信服务端
	serverMutex  sync.RWMutex               // 保护Server字段的并发访问
	conn         net.Conn                   // TCP连接
	tcpMarshal   marshaller.TcpMarshal      // TCP协议序列化器
	isConnected  bool                       // 是否已连接
	connectChan  chan struct{}              // 连接通知通道
	ctx          context.Context            // 上下文
	cancel       context.CancelFunc         // 取消函数
	doorCache    []getDoorInfo              // 门列表缓存
	doorCacheSet bool                       // 门列表缓存是否已设置
}

// Open 打开门控器连接，初始化通信通道。
// 从 chanInfo.Extend 中获取 MAC 地址作为设备唯一标识，
// 注册到全局 worker 的 macAddrMap 中等待设备主动连接。
func (c *Controller) Open(chanInfo driver.ChannelInfo) consts.Quality {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	var isWorkerStart bool = false
	c.connectChan = make(chan struct{}, 1)
	func() {
		worker.mutex.Lock()
		defer worker.mutex.Unlock()
		if worker.IsFirst {
			ok := worker.Start()
			if !ok {
				c.Errorf("worker start error")
				return
			}
			worker.IsFirst = false
		}
		isWorkerStart = true
	}()

	if !isWorkerStart {
		return consts.QualityUncertain
	}

	c.chanInfo = chanInfo
	c.version = strings.ToLower(chanInfo.ProtocolVersion)
	c.timeout = chanInfo.TimeoutMS
	c.tcpMarshal = marshaller.NewCACSMarshal()

	// 从 extend 字段获取 MAC 地址
	macAddr := ""
	if chanInfo.Extend != nil {
		if mac, ok := chanInfo.Extend["mac_address"].(string); ok && mac != "" {
			// 统一格式：大写，去掉分隔符
			macAddr = strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", ""))
		}
	}

	if macAddr == "" {
		c.Errorf("CACS 门控器必须配置 mac_address")
		return consts.QualityUncertain
	}

	// 使用 MAC 地址作为 key
	c.chanInfo.ChannelID = macAddr
	func() {
		worker.macAddrMapMutex.Lock()
		defer worker.macAddrMapMutex.Unlock()
		worker.macAddrMap[macAddr] = c
	}()
	c.Infof("CACS 门控器注册 MAC 地址白名单: %s", macAddr)

	go c.waitForConnect()
	return consts.QualityOK
}

// Close 关闭门控器连接，清理资源。
// 从 macAddrMap 和 serverMap 中移除当前门控器的注册信息。
func (c *Controller) Close() consts.Quality {
	c.cancel()

	// 从 macAddrMap 中移除
	func() {
		worker.macAddrMapMutex.Lock()
		defer worker.macAddrMapMutex.Unlock()
		delete(worker.macAddrMap, c.chanInfo.ChannelID)
	}()

	// 关闭 serverMap 中的连接
	func() {
		worker.serverMapMutex.Lock()
		defer worker.serverMapMutex.Unlock()
		if server, ok := worker.serverMap[c.chanInfo.ChannelID]; ok {
			server.Close()
			delete(worker.serverMap, c.chanInfo.ChannelID)
		}
	}()

	c.Infof("CACS 门控器已关闭, MAC: %s", c.chanInfo.ChannelID)
	return consts.QualityOK
}

// checkConnection 检查连接状态，返回 Server 和错误
func (c *Controller) checkConnection() (*DoorServer, error) {
	c.serverMutex.RLock()
	defer c.serverMutex.RUnlock()

	if !c.isConnected {
		return nil, fmt.Errorf("controller: %s, not connected", c.chanInfo.ChannelID)
	}
	if c.Server == nil {
		return nil, fmt.Errorf("controller: %s, server not initialized", c.chanInfo.ChannelID)
	}
	return c.Server, nil
}

// setServer 设置 Server（线程安全）
func (c *Controller) setServer(server *DoorServer) {
	c.serverMutex.Lock()
	defer c.serverMutex.Unlock()
	c.Server = server
	c.isConnected = server != nil
}

// Ping 检测门控器连接是否正常，通过发送门状态查询请求验证。
func (c *Controller) Ping() error {
	server, err := c.checkConnection()
	if err != nil {
		return err
	}
	_ = server // 后续使用
	_, ok, _, _, _ := c.getDoorState(cacs.DoorStateReq{Id: 1})
	if !ok {
		return fmt.Errorf("ping failed, send request failed")
	}
	return nil
}

// RemoteControl 远程控制门禁（开门/关门等），返回控制结果。
func (c *Controller) RemoteControl(
	req cacs.DoorControlReq,
) (cacs.DoorControlResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.DoorControlResp{}, false, 0, consts2.KRequestError, err
	}

	cmd := consts2.KCommandRequestRemoteControl
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.DoorControlResp{}, false, server.p.rtn,
			consts2.KMarshalError, fmt.Errorf("req marshal failed, err: %v", err)
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.DoorControlResp{}, false, server.p.rtn,
			consts2.KRequestError, fmt.Errorf("req data send failed, err: %v", err)
	}

	rrpcKey := consts2.GetRRPCRemoteControl(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.DoorControlResp{}, false, server.p.rtn,
			consts2.KRecvRespError, fmt.Errorf("rrpc get resp timeout")
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.DoorControlResp{}, false, server.p.rtn,
			consts2.KUnMarshalError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	resp, err := c.tcpMarshal.Unmarshal(consts2.KCommandResponseRemoteControl, bytes)
	if err != nil {
		c.Errorf("resp tcpUnmarshal to DoorControlResp failed, err: %v", err)
		return cacs.DoorControlResp{}, false, server.p.rtn,
			consts2.KUnMarshalError,
			fmt.Errorf("resp tcpUnmarshal to DoorControlResp failed, err: %v", err)
	}
	doorStatusResp, ok := resp.(cacs.DoorControlResp)
	if !ok {
		c.Errorf("resp type error, it should be DoorControlResp")
		return cacs.DoorControlResp{}, false, server.p.rtn,
			consts2.KUnMarshalError,
			fmt.Errorf("resp type error, it should be DoorControlResp")
	}
	return doorStatusResp, true, server.p.rtn, consts2.KNormal, nil
}

// DownloadControllerParams 下载控制器参数到门控器设备。
func (c *Controller) DownloadControllerParams(
	req cacs.DownloadControllerParamsReq,
) (cacs.DownloadControllerParamsResp, bool, uint32) {
	server, err := c.checkConnection()
	if err != nil {
		c.Errorf("checkConnection failed: %v", err)
		return cacs.DownloadControllerParamsResp{}, false, 0
	}

	cmd := consts2.KCommandRequestDownloadControllerParams
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.DownloadControllerParamsResp{}, false, server.p.rtn
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.DownloadControllerParamsResp{}, false, server.p.rtn
	}
	respRaw, ok := rrpc.Manager().Get(consts2.GetRRPCDownloadControllerParams(c.chanInfo.ChannelID), c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.DownloadControllerParamsResp{}, false, server.p.rtn
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.DownloadControllerParamsResp{}, false, server.p.rtn
	}
	resp, err := c.tcpMarshal.Unmarshal(consts2.KCommandResponseDownloadControllerParams, bytes)
	if err != nil {
		c.Errorf("resp tcpUnmarshal to DownloadControllerParamsResp failed, err: %v", err)
		return cacs.DownloadControllerParamsResp{}, false, server.p.rtn
	}
	downloadControllerParamsResp, ok := resp.(cacs.DownloadControllerParamsResp)
	if !ok {
		c.Errorf("resp type error, it should be DownloadControllerParamsResp")
		return cacs.DownloadControllerParamsResp{}, false, server.p.rtn
	}
	return downloadControllerParamsResp, true, server.p.rtn
}

// GetControllerParams 从门控器设备读取控制器参数。
func (c *Controller) GetControllerParams(
	req cacs.GetControllerParamsReq,
) (cacs.GetControllerParamsResp, bool, uint32) {
	server, err := c.checkConnection()
	if err != nil {
		c.Errorf("checkConnection failed: %v", err)
		return cacs.GetControllerParamsResp{}, false, 0
	}

	cmd := consts2.KCommandRequestGetControllerParams
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.GetControllerParamsResp{}, false, server.p.rtn
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.GetControllerParamsResp{}, false, server.p.rtn
	}
	rrpcKey := consts2.GetRRPCGetControllerParams(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, 10*time.Second)
	//respRaw, ok := rrpc.Manager().Get(consts2.KRRPCGetControllerParams, c.timeout)
	if !ok {
		c.Errorf("rrpc.Manager().Get KRRPCGetControllerParams false")
		return cacs.GetControllerParamsResp{}, false, server.p.rtn
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.GetControllerParamsResp{}, false, server.p.rtn
	}
	resp, err := c.tcpMarshal.Unmarshal(consts2.KCommandResponseGetControllerParams, bytes)
	if err != nil {
		c.Errorf("resp tcpUnmarshal to GetControllerParamsResp failed, err: %v", err)
		return cacs.GetControllerParamsResp{}, false, server.p.rtn
	}
	getControllerParamsResp, ok := resp.(cacs.GetControllerParamsResp)
	if !ok {
		c.Errorf("resp type error, it should be GetControllerParamsResp")
		return cacs.GetControllerParamsResp{}, false, server.p.rtn
	}
	return getControllerParamsResp, true, server.p.rtn
}

// resetForReconnect 重连时重置 Controller 状态
// 注意：不创建新的 waitForConnect 协程，由现有协程继续处理
func (c *Controller) resetForReconnect() {
	c.serverMutex.Lock()
	defer c.serverMutex.Unlock()
	c.isConnected = false
	c.Server = nil
}

// waitForConnect 等待门控器设备建立TCP连接。
// 循环监听 connectChan，支持多次重连场景。
func (c *Controller) waitForConnect() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case _, ok := <-c.connectChan:
			if !ok {
				// channel 已关闭，退出
				return
			}
			func() {
				worker.serverMapMutex.RLock()
				defer worker.serverMapMutex.RUnlock()
				if server, ok := worker.serverMap[c.chanInfo.ChannelID]; ok {
					c.setServer(server)
				}
			}()
			c.Infof("门控器连接已建立, MAC: %s", c.chanInfo.ChannelID)
			// 不 return，继续循环，支持多次重连
		}
	}
}

// IsReady 检查门控器是否已就绪（Server已初始化）。
func (c *Controller) IsReady() bool {
	return c.Server != nil
}
