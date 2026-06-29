// Package iec104 包提供了对 IEC 60870-5-14 通信协议的支持，包括数据帧解析、连接管理等功能。
package iec104

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// IEC104Client IEC104客户端
type IEC104Client struct {
	address        string   // 主地址
	subAddress     string   // 备地址
	conn           net.Conn // tcp连接
	retryLoopStart bool     // 重试循环是否已启动

	rsn uint16 // 接收I帧序号
	ssn uint16 // 发送I帧序号
	ifn int    // 接收I帧数
	sff int    // S帧回复频率>=1, 默认1, 表示每收到sff次I帧回复一次S帧

	dataChan                 chan *APDU    // 数据处理通道
	sendChan                 chan []byte   // 数据发送通道
	startChan                chan struct{} // 启动帧通道
	testChan                 chan struct{} // 测试帧通道
	totalCallChan            chan struct{} // 总召通道
	electricityTotalCallChan chan struct{} // 电度总召通道

	isTotalCalling atomic.Bool // 是否正在总召中 （此状态下不进行电度召）

	// 超时时间配置
	totalCallInterval            time.Duration
	electricityTotalCallInterval time.Duration
	timeoutT0                    time.Duration
	timeoutT1                    time.Duration
	timeoutT2                    time.Duration
	timeoutT3                    time.Duration
	maxReadTimeout               time.Duration

	// 定时器
	testTimer      *time.Timer // 测试帧t1超时
	totalCallTimer *time.Timer // 总召t1超时
	t2             *time.Timer // t2
	t3             *time.Timer // t3

	clientCancel context.CancelFunc // 客户端cancel
	workerCancel context.CancelFunc // 工作协程cancel
	wg           *sync.WaitGroup    // WaitGroup
	Log          Logger             // 日志引擎
	deal         func(c *APDU)      // 接收数据钩子处理函数
}

type Option func(client *IEC104Client)

// WithLogger WithLogger
func WithLogger(logger Logger) Option {
	return func(client *IEC104Client) {
		client.Log = logger
	}
}

// WithMaxReadTimeout WithMaxReadTimeout
func WithMaxReadTimeout(t time.Duration) Option {
	return func(client *IEC104Client) {
		client.maxReadTimeout = t
	}
}

// NewIEC104Client New IEC104客户端
func NewIEC104Client(opts ...Option) *IEC104Client {
	c := &IEC104Client{
		wg:                           new(sync.WaitGroup),
		Log:                          new(LogFmt),
		sff:                          SFrameFrequencyMin,
		totalCallInterval:            defaultTotalCallInterval,
		electricityTotalCallInterval: defaultElectricityTotalCallInterval,
		timeoutT0:                    defaultTimeoutT0,
		timeoutT1:                    defaultTimeoutT1,
		timeoutT2:                    defaultTimeoutT2,
		timeoutT3:                    defaultTimeoutT3,
		retryLoopStart:               false,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.maxReadTimeout < defaultMaxReadTimeout {
		c.maxReadTimeout = defaultMaxReadTimeout
	}
	if c.maxReadTimeout < c.timeoutT3*2 {
		c.maxReadTimeout = c.timeoutT3 * 2
	}

	return c
}

func newStopTimer() *time.Timer {
	tim := time.NewTimer(0)
	tim.Stop()
	return tim
}

// SetAddress 设置连接地址
func (c *IEC104Client) SetAddress(address, subAddress string) {
	c.address = address
	c.subAddress = subAddress
}

// SetDeal 设置数据处理函数
func (c *IEC104Client) SetDeal(deal func(c *APDU)) {
	c.deal = deal
}

// SetLogger 设置日志引擎
func (c *IEC104Client) SetLogger(logger Logger) {
	c.Log = logger
}

// SetSFrameFrequency 设置S帧回复频率
func (c *IEC104Client) SetSFrameFrequency(sff int) {
	if sff < SFrameFrequencyMin {
		sff = SFrameFrequencyMin
	}
	if sff > SFrameFrequencyMax {
		sff = SFrameFrequencyMax
	}
	c.sff = sff
}

// SetTimeoutT0 设置t0，单位秒
func (c *IEC104Client) SetTimeoutT0(t int) (err error) {
	if t < 0 {
		return fmt.Errorf("set value can not less than 0")
	}
	c.timeoutT0 = time.Duration(t) * time.Second
	return
}

// SetTimeoutT1 设置t1，单位秒
func (c *IEC104Client) SetTimeoutT1(t int) (err error) {
	if t < 0 {
		return fmt.Errorf("set value can not less than 0")
	}
	c.timeoutT1 = time.Duration(t) * time.Second
	return
}

// SetTimeoutT2 设置t2，单位秒
func (c *IEC104Client) SetTimeoutT2(t int) (err error) {
	if t < 0 {
		return fmt.Errorf("set value can not less than 0")
	}
	c.timeoutT2 = time.Duration(t) * time.Second
	return
}

// SetTimeoutT3 设置t3，单位秒
func (c *IEC104Client) SetTimeoutT3(t int) (err error) {
	if t < 0 {
		return fmt.Errorf("set value can not less than 0")
	}
	c.timeoutT3 = time.Duration(t) * time.Second
	return
}

// SetTimeoutTotalCall 设置总召间隔，单位分钟
func (c *IEC104Client) SetTimeoutTotalCall(t int) (err error) {
	if t < 0 {
		return fmt.Errorf("set value can not less than 0")
	}
	c.totalCallInterval = time.Duration(t) * time.Minute
	return
}

func (c *IEC104Client) readImpl(b []byte) (int, error) {
	if c.conn == nil {
		return 0, fmt.Errorf("conn is nil")
	}
	_ = c.conn.SetReadDeadline(time.Now().Add(c.maxReadTimeout))
	return c.conn.Read(b)
}

// readData 读取数据
func (c *IEC104Client) readData(ctx context.Context) error {
	if c.conn == nil {
		return fmt.Errorf("读取数据时连接为空")
	}

	buf := make([]byte, startAndNumLen)
	//读取启动符和长度
	n, err := c.readImpl(buf)
	if err != nil {
		c.Log.Errorf("读取启动符和长度错误: %v", err)
		return err
	}

	if n == 0 {
		return fmt.Errorf("读取启动符和长度为空")
	}
	length := int(buf[1])
	//读取正文
	contentBuf := make([]byte, length)
	n, err = c.readImpl(contentBuf)
	if err != nil {
		c.Log.Errorf("读取正文错误: %v", err)
		return err
	}
	//长度不够继续读取，直至达到期望长度
	for i := 2; n < length; i++ {
		nextLength := length - n
		nextBuf := make([]byte, nextLength)
		m, err := c.readImpl(nextBuf)
		if err != nil {
			c.Log.Error("循环读取正文:", err)
			return err
		}
		contentBuf = append(contentBuf[:n], nextBuf[:m]...)
		n = len(contentBuf)
		c.Log.Debugf("循环读取数据，当前为第 %d 次读取，期望长度: %d,本次长度: %d,当前总长度: %d", i, length, m, n)
	}

	singleData := make([]byte, 0, startAndNumLen+n)
	singleData = append(singleData, buf...)
	singleData = append(singleData, contentBuf...)
	c.Log.Debugf("收到原始数据: [%X], rsn: %d, ssn: %d, 长度: %d", singleData, c.rsn, c.ssn, startAndNumLen+n)

	return c.parseData(ctx, singleData)
}

// parseData 解析接收到的数据
func (c *IEC104Client) parseData(ctx context.Context, buf []byte) error {
	apdu, err := ParseAPDU(buf)
	if err != nil {
		c.Log.Warnf("解析APDU异常: %v", err)
		return nil // 解析出错忽略
	}
	switch apdu.CtrFrame.(type) {
	case IFrame:
		c.incrRsn()
		c.t3.Reset(c.timeoutT3)
		switch apdu.ASDU.TypeID {
		case CIcNa1:
			if apdu.ASDU.Cause == COTActCon {
				go func() {
					c.totalCallChan <- struct{}{}
					c.Log.Info("接收总召唤确认帧")
				}()
			} else if apdu.ASDU.Cause == COTActTerm {
				c.Log.Info("接收总召唤结束帧")
			}
		case CCiNa1:
			if apdu.ASDU.Cause == COTActCon {
				go func() {
					c.electricityTotalCallChan <- struct{}{}
					c.Log.Info("接收电度总召唤确认帧")
				}()
			} else if apdu.ASDU.Cause == COTActTerm {
				c.Log.Info("接收电度总召唤结束帧")
			}
		default:
			c.ifn++
			c.Log.Debugf("接收到第 %d 个I帧", c.ifn)
			c.dataChan <- apdu
		}
		if c.ifn%c.sff == 0 {
			c.sendSFrame() // 回复S帧
		}
		c.t2.Reset(c.timeoutT2) // 重置S帧t2回复
	case SFrame:
		c.Log.Debug("接收到S帧")
	case UFrame:
		c.Log.Debug("接收到U帧")
		u := apdu.CtrFrame.(UFrame)
		switch u.cmd {
		case startDtCon:
			c.startChan <- struct{}{}
		case testFrAct:
			c.Log.Debug("U帧为测试激活帧,发送测试确认帧")
			c.sendUFrame(testFrCon)
		case testFrCon:
			c.testChan <- struct{}{}
		}
	default:
		c.Log.Warn("接收到未知帧")
	}
	return nil
}

// 建立连接
func (c *IEC104Client) connect(ctx context.Context) (err error) {
	c.initProperty()
	c.resetSn()
	c.conn, err = c.dail()
	if err != nil {
		c.Log.Errorf("建立连接失败, err: %v", err)
		return err
	}

	workerCtx, cancel := context.WithCancel(ctx)
	c.Log.Infof("建立连接成功, addr: %s", c.conn.RemoteAddr().String())
	go c.run(workerCtx, cancel)
	return nil
}

// Connect 建立连接，断连重试
func (c *IEC104Client) Connect(ctx context.Context) (err error) {
	if c.IsConnected() || c.retryLoopStart {
		return nil
	}

	clientCtx, cancel := context.WithCancel(ctx)
	c.clientCancel = cancel

	return c.connect(clientCtx)
}

// Close 程序结束
func (c *IEC104Client) Close() (err error) {
	if c.clientCancel != nil {
		c.clientCancel()
		c.clientCancel = nil
	}
	if c.conn != nil {
		err = c.conn.Close()
		c.conn = nil
	}
	return err
}

// Dail 建立tcp连接，支持重试和主备切换
func (c *IEC104Client) dail() (conn net.Conn, err error) {
	if len(c.address) == 0 {
		return nil, fmt.Errorf("地址不能为空")
	}

	conn, err = net.DialTimeout("tcp", c.address, c.timeoutT0)
	if err != nil && c.subAddress != "" {
		conn, err = net.DialTimeout("tcp", c.subAddress, c.timeoutT0)
	}

	return conn, err
}

// run
func (c *IEC104Client) run(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		_ = c.workerClose()
	}()

	c.workerCancel = cancel
	// 启动读写协程
	c.wg.Add(2)
	go c.read(ctx)
	go c.write(ctx)

	c.Log.Info("发送启动激活帧")
	c.sendUFrame(startDtAct)

	// 接收启动确认帧
	timer := time.NewTimer(c.timeoutT1)
	select {
	case <-c.startChan:
		c.Log.Info("接收到启动确认帧, 发送总召唤")
	case <-timer.C:
		c.Log.Error("接收启动确认帧超时")
		return
	}

	// 启动数据处理协程
	c.wg.Add(1)
	go c.handler(ctx, c.deal)

	c.sendTotalCall()
	// 接收总召唤确认帧
	timer = time.NewTimer(c.timeoutT1)
	select {
	case <-c.totalCallChan:
		c.Log.Info("接收总召唤确认帧")
	case <-timer.C:
		c.Log.Error("接收总召唤确认帧超时")
		// 设备无总召，不返回
	}

	// 等待总召响应后再发电度召
	c.sendElectricityTotalCall()

	// 启动超时控制协程
	c.wg.Add(5)
	go c.timeout1Deal(ctx)
	go c.t2Send(ctx)
	go c.t3Send(ctx)

	// 定时总召
	go c.totalCall(ctx, c.wg)
	go c.electricityTotalCallLoop(ctx, c.wg)

	c.wg.Wait()
	return
}

// timeout1Deal t1超时处理
func (c *IEC104Client) timeout1Deal(ctx context.Context) {
	c.Log.Info("t1超时处理协程启动")
	defer func() {
		c.workerCancel()
		c.wg.Done()
		c.Log.Info("t1超时处理协程停止")
	}()
	for {
		select {
		case <-ctx.Done():
			c.Log.Info("timeout1Deal ctx done")
			return
		case <-c.testTimer.C:
			c.Log.Error("t1 testTimer 超时")
			return
		case <-c.totalCallTimer.C:
			c.isTotalCalling.Store(false)
			c.Log.Error("t1 totalCallTimer 超时")
			return
		}
	}
}

// t2Send t2发送S帧
func (c *IEC104Client) t2Send(ctx context.Context) {
	c.Log.Infof("t2发送S帧协程启动")
	defer func() {
		c.workerCancel()
		c.wg.Done()
		c.Log.Infof("t2发送S帧协程停止")
	}()
	for {
		select {
		case <-ctx.Done():
			c.Log.Info("t2Send ctx done")
			return
		case <-c.t2.C:
			c.sendSFrame()
			c.t2.Stop()
		}
	}
}

// t3Send t3发送测试链路帧
func (c *IEC104Client) t3Send(ctx context.Context) {
	c.Log.Infof("t3发送测试链路帧协程启动")
	defer func() {
		c.workerCancel()
		c.wg.Done()
		c.Log.Infof("t3发送测试链路帧协程停止")
	}()
	for {
		select {
		case <-ctx.Done():
			c.Log.Info("t3Send ctx done")
			return
		case <-c.t3.C:
			c.Log.Debug("发送测试激活帧")
			c.sendUFrame(testFrAct)
			c.testTimer.Reset(c.timeoutT1)
			c.t3.Reset(c.timeoutT3)
		case <-c.testChan:
			c.Log.Debug("接收到测试确认帧")
			c.testTimer.Stop()
		}
	}
}

// Read 读数据
func (c *IEC104Client) read(ctx context.Context) {
	c.Log.Infof("socket读协程启动")
	defer func() {
		c.workerCancel()
		c.wg.Done()
		c.Log.Infof("socket读协程停止")
	}()
	for {
		select {
		case <-ctx.Done():
			c.Log.Info("read ctx done")
			return
		default:
		}

		err := c.readData(ctx)
		if err != nil {
			c.Log.Errorf("readData error: %v", err)
			return
		}
	}
}

// Write 写数据
func (c *IEC104Client) write(ctx context.Context) {
	c.Log.Infof("socket写协程启动")
	defer func() {
		c.workerCancel()
		c.wg.Done()
		c.Log.Infof("socket写协程停止")
	}()
	for {
		select {
		case <-ctx.Done():
			c.Log.Info("write ctx done")
			return
		case data := <-c.sendChan:
			_, err := c.writeData(data)
			if err != nil {
				return
			}
			c.Log.Debugf("写入: [%X]", data)
		}
	}
}

// writeData
func (c *IEC104Client) writeData(data []byte) (n int, err error) {
	if c.conn == nil {
		return 0, fmt.Errorf("写数据时连接为空")
	}
	return c.conn.Write(data)
}

// handler 处理接收到的已解析数据
func (c *IEC104Client) handler(ctx context.Context, deal func(c *APDU)) {
	c.Log.Infof("数据处理协程启动")
	defer func() {
		c.workerCancel()
		c.wg.Done()
		c.Log.Infof("数据处理协程停止")
	}()
	for {
		select {
		case resp := <-c.dataChan:
			c.Log.Debugf("接收到数据类型: %d, 原因: %d, 长度: %d", resp.ASDU.TypeID, resp.ASDU.Cause, len(resp.Signals))
			if deal != nil {
				go deal(resp)
			}
		case <-ctx.Done():
			return
		}
	}
}

// totalCall 总召唤
func (c *IEC104Client) totalCall(ctx context.Context, wg *sync.WaitGroup) {
	c.Log.Infof("定时总召唤协程启动")
	defer func() {
		c.workerCancel()
		wg.Done()
		c.Log.Infof("定时总召唤协程停止")
	}()
	// 定时器，每15分钟发送一次总召唤
	ticker := time.NewTimer(c.totalCallInterval)
	for {
		select {
		case <-ticker.C:
			c.Log.Infof("每隔 %v 发送一次总召唤", c.totalCallInterval)
			c.isTotalCalling.Store(true)
			c.sendTotalCall()
			c.totalCallTimer.Reset(c.timeoutT1)
			ticker.Reset(c.totalCallInterval)
		case <-c.totalCallChan:
			c.Log.Infof("接收总召唤确认帧")
			c.totalCallTimer.Stop()
			c.isTotalCalling.Store(false)
		case <-ctx.Done():
			return
		}
	}
}

func (c *IEC104Client) electricityTotalCallLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		c.Log.Info("定时电度总召唤协程停止")
	}()
	c.Log.Info("定时电度总召唤协程启动")
	for {
		select {
		case <-time.After(c.electricityTotalCallInterval):
			c.Log.Infof("每隔 %v 发送一次电度总召唤", c.electricityTotalCallInterval)

			// 如果在总召期间，则等待
			waitCount := 10
			for c.isTotalCalling.Load() {
				c.Log.Infof("等待总召唤结束")
				time.Sleep(time.Second)
				waitCount--
				if waitCount <= 0 {
					break
				}
			}

			c.sendElectricityTotalCall()

		case <-c.electricityTotalCallChan:
			c.Log.Infof("接收电度总召唤确认帧")
		case <-ctx.Done():
			return
		}
	}
}

// sendUFrame 发送U帧
func (c *IEC104Client) sendUFrame(cmd [4]byte) {
	data := convertBytes(convert4BytesToSlice(cmd))
	c.Log.Debugf("发送U帧: [%X]", data)
	c.sendChan <- data
}

// sendSFrame 发送S帧
func (c *IEC104Client) sendSFrame() {
	rsnBytes := parseLittleEndianUInt16(c.rsn << 1)
	sendBytes := make([]byte, 0, len(rsnBytes)+2)
	sendBytes = append(sendBytes, 0x01, 0x00)
	sendBytes = append(sendBytes, rsnBytes...)
	data := convertBytes(sendBytes)
	c.Log.Debugf("发送S帧: [%X]", data)
	c.sendChan <- data
}

// sendTotalCall 发送总召唤
func (c *IEC104Client) sendTotalCall() {
	ssnBytes := parseLittleEndianUInt16(c.ssn << 1)
	rsnBytes := parseLittleEndianUInt16(c.rsn << 1)
	totalCallData := make([]byte, 0, len(ssnBytes)+len(rsnBytes)+len(totalCallAct))
	totalCallData = append(totalCallData, ssnBytes...)
	totalCallData = append(totalCallData, rsnBytes...)
	totalCallData = append(totalCallData, totalCallAct...)
	data := convertBytes(totalCallData)
	c.Log.Infof("发送总召唤: [%X]", data)
	c.sendChan <- data
}

// sendTotalCall 发送电度总召唤
func (c *IEC104Client) sendElectricityTotalCall() {
	ssnBytes := parseLittleEndianUInt16(c.ssn << 1)
	rsnBytes := parseLittleEndianUInt16(c.rsn << 1)
	totalCallData := make([]byte, 0, len(ssnBytes)+len(rsnBytes)+len(electricitytotalCallAct))
	totalCallData = append(totalCallData, ssnBytes...)
	totalCallData = append(totalCallData, rsnBytes...)
	totalCallData = append(totalCallData, electricitytotalCallAct...)
	data := convertBytes(totalCallData)
	c.Log.Infof("发送电度总召唤: [%X]", data)
	c.sendChan <- data
}

// incrRsn 增加rsn
func (c *IEC104Client) incrRsn() {
	c.rsn++
}

// 重置收发序号
func (c *IEC104Client) resetSn() {
	c.rsn = 0
	c.ssn = 0
}

// workerClose 内部工作关闭的收尾处理工作
func (c *IEC104Client) workerClose() error {
	var err error
	if c.conn != nil {
		if err = c.conn.Close(); err != nil {
			c.Log.Warnf("close connection error: %v", err)
		}
		c.conn = nil
	}
	c.Log.Infof("close connection, worker close")
	return err
}

// IsConnected 判断连接是否断开
func (c *IEC104Client) IsConnected() bool {
	return c.conn != nil
}

func (c *IEC104Client) initProperty() {
	c.dataChan = make(chan *APDU, 1)
	c.sendChan = make(chan []byte, 1)
	c.startChan = make(chan struct{}, 1)
	c.testChan = make(chan struct{}, 1)
	c.totalCallChan = make(chan struct{}, 1)
	c.electricityTotalCallChan = make(chan struct{}, 1)
	c.testTimer = newStopTimer()
	c.totalCallTimer = newStopTimer()
	c.t2 = newStopTimer()
	c.t3 = newStopTimer()
}
