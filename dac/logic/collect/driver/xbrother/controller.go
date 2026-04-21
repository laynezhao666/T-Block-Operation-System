// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/driver"
	"dac/entity/model/driver/xbrother"
	"dac/entity/model/marshaller"
	"dac/entity/redis"
	"dac/entity/utils"
	consts2 "dac/logic/collect/driver/xbrother/consts"
	"dac/logic/dlm"
	"dac/repo/dac"
	"encoding/json"

	"dac/entity/utils/rrpc"
	"dac/entity/utils/tlog"
	"dac/entity/utils/ttime"
)

// Controller XBrother协议门禁控制器驱动实例
type Controller struct {
	timeout  time.Duration              // 请求超时时间
	baseInfo driver.ControllerBasicInfo // 控制器基本信息
	chanInfo driver.ChannelInfo         // 通道信息
	version  string                     // 协议版本
	Server   *DoorServer                // 门禁服务端
	marshal  marshaller.TcpMarshal      // TCP协议序列化器

	ctx    context.Context    // 上下文
	cancel context.CancelFunc // 取消函数

	doorNum    int // 门数量
	nextCardNo int // 下一个卡编号

	redisClient *redis.Client // Redis客户端

	hasSetTime  bool        // 是否已同步时间
	isConnected atomic.Bool // 是否已连接
	once        sync.Once   // 确保只执行一次

	logger tlog.Logger // 日志记录器
}

// Open 打开XBrother控制器连接，初始化通道、Redis和TCP服务
func (c *Controller) Open(chanInfo driver.ChannelInfo) consts.Quality {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.chanInfo = chanInfo
	c.redisClient = redis.GetClient()
	c.marshal = marshaller.NewXBrotherMarshaller()
	c.timeout = chanInfo.TimeoutMS

	c.logger = tlog.NewPrefixLogger(fmt.Sprintf("[controller-%v@%v]",
		c.baseInfo.ID, chanInfo.ChannelID), config.Log)

	var err error
	c.doorNum, err = c.GetDoorNumber()
	if err != nil {
		c.logger.Errorf("get door number failed, err: %v", err)
		return consts.QualityUncertain
	}

	c.Server = NewDoorServer(c.timeout, c.chanInfo.ChannelID, c.chanInfo.ChannelID, c.doorNum, c.baseInfo)

	c.once.Do(func() {
		go c.hasLockAndOpen(context.Background())
	})

	if !dlm.GetWorker().HasLock() {
		c.isConnected.Store(false)
		// 没拿到锁，返回OK, 实际未连接门禁控制器
		return consts.QualityOK
	}

	go c.saveEvents(c.ctx)
	go c.saveAlarms(c.ctx)
	go c.saveControllerStatus(c.ctx)

	if err = c.Server.Connect(c.ctx); err != nil {
		c.logger.Errorf("connect error: %v", err)
		return consts.QualityUncertain
	}

	c.isConnected.Store(true)
	go c.ConnCheckAndReConnect(c.ctx)

	return consts.QualityOK
}

// Close 关闭XBrother控制器连接
func (c *Controller) Close() consts.Quality {
	c.cancel()
	c.logger.Infof("close... ")
	return consts.QualityOK
}

// Ping 测试控制器连通性（通过同步时间实现）
func (c *Controller) Ping() error {
	return c.SetTime()
}

// IsReady 返回控制器是否就绪
func (c *Controller) IsReady() bool {
	return true
}

// setControllerParams 设置控制器参数
func (c *Controller) setControllerParams(
	req xbrother.SetControllerParamsReq, doorNo uint8,
) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo,
		consts2.GetRRPCSetControllerParams(c.chanInfo.ChannelID),
		consts2.CommandSetControllerParams)
}

// sendRequest 发送请求到门控器并等待RRPC响应
func (c *Controller) sendRequest(
	req interface{}, doorNo uint8, rrpcKey string, cmd uint8,
) (xbrother.CommonResp, error) {
	data, err := c.marshal.Marshal(uint32(cmd), req)
	if err != nil {
		return xbrother.CommonResp{}, fmt.Errorf("req marshal failed, cmd: %d, err: %w", cmd, err)
	}

	if err = c.Server.Request(cmd, doorNo, data); err != nil {
		return xbrother.CommonResp{}, fmt.Errorf("req data send error: %w", err)
	}

	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		return xbrother.CommonResp{}, fmt.Errorf("rrpc get resp timeout, cmd: %d, doorNo: %d, req: %+v, timeout: %v",
			cmd, doorNo, req, c.timeout)
	}

	bytes, ok := respRaw.([]byte)
	if !ok {
		return xbrother.CommonResp{}, fmt.Errorf("respRaw converse to []byte error: %w", err)
	}

	resp, err := c.marshal.Unmarshal(uint32(cmd), bytes)
	if err != nil {
		return xbrother.CommonResp{}, fmt.Errorf("unmarshal CommonResp failed error: %w", err)
	}

	commonResp, ok := resp.(xbrother.CommonResp)
	if !ok {
		return xbrother.CommonResp{}, fmt.Errorf("resp type error, it should be CommonResp")
	}
	if commonResp.Rtn != consts2.ACK {
		return xbrother.CommonResp{}, fmt.Errorf("rtn %v is not ok", commonResp.Rtn)
	}
	return commonResp, nil
}

// getTimeStamp 将协议时间字段转换为Unix时间戳
func getTimeStamp(year uint8, month uint8, day uint8, hour uint8, minute uint8, second uint8) (int64, error) {
	t, err := ttime.ParseLocal(fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		int(year)+2000, month, day, hour, minute, second))
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// saveControllerStatus 监听并保存控制器状态上报数据
func (c *Controller) saveControllerStatus(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Infof("stop save controller status")
			return
		case status, ok := <-c.Server.controllerStatusChan:
			if !ok {
				c.logger.Errorf("controller status channel close")
				break
			}
			c.logger.Debug("heart beat")
			if err := c.saveDoorStatusInRedis(ctx, status); err != nil {
				c.logger.Errorf("save door status in redis error, err: %v", err)
				rrpc.Manager().Set(consts2.GetRRPCSetDoorStatus(c.chanInfo.ChannelID), err)
				return
			}
			if err := c.updateCurrentAlarm(ctx); err != nil {
				c.logger.Errorf("update current alarm in redis error, err: %v", err)
				rrpc.Manager().Set(consts2.GetRRPCSetDoorStatus(c.chanInfo.ChannelID), err)
				return
			}
			rrpc.Manager().Set(consts2.GetRRPCSetDoorStatus(c.chanInfo.ChannelID), nil)
		}
	}
}

// updateCurrentAlarm 根据门状态更新当前告警（关门时清除超时告警）
func (c *Controller) updateCurrentAlarm(ctx context.Context) error {
	key := utils.GenerateRedisKeyDoorStatus(c.chanInfo.ChannelID)
	doorStateBytes, err := c.redisClient.Get(
		context.Background(), key,
	).Bytes()
	if err != nil {
		return err
	}
	var doorState []int
	if err := json.Unmarshal(doorStateBytes, &doorState); err != nil {
		return err
	}
	if len(doorState) != c.doorNum {
		return fmt.Errorf("unexpected doorNum, get %d, but expected %d", len(doorState), c.doorNum)
	}

	closeDoors := make(map[string]struct{})
	for i, v := range doorState {
		if v == consts2.DriverDoorStatusClose {
			closeDoors[fmt.Sprintf("%d", i+1)] = struct{}{}
		}
	}
	keys, err := c.redisClient.HKeys(ctx, utils.GenerateRedisKeyDoorOpenTimeout(c.chanInfo.ChannelID)).Result()
	if err != nil {
		return err
	}
	for i := range keys {
		if _, ok := closeDoors[keys[i]]; ok {
			timeoutKey := utils.GenerateRedisKeyDoorOpenTimeout(
				c.chanInfo.ChannelID,
			)
			if err := c.redisClient.HDel(ctx, timeoutKey, keys[i]).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

// saveDoorStatusInRedis 将门状态数据保存到Redis
func (c *Controller) saveDoorStatusInRedis(_ context.Context, req xbrother.ControllerStatusUploadReq) error {
	doorState := make([]int, c.doorNum)
	doorStatus := req.DoorStatus
	for i := 0; i < c.doorNum; i++ {
		if doorStatus&(uint8(1)<<i) == consts2.ControllerDoorStatusOpen {
			doorState[i] = consts2.DriverDoorStatusOpen
		} else {
			doorState[i] = consts2.DriverDoorStatusClose
		}
	}
	doorStateJson, err := json.Marshal(doorState)
	if err != nil {
		return err
	}

	return c.redisClient.Set(context.Background(), utils.GenerateRedisKeyDoorStatus(c.chanInfo.ChannelID),
		doorStateJson, consts2.RedisDefaultTimeout).Err()

}

// getLastCardIndex 获取最后一张卡的索引号
func (c *Controller) getLastCardIndex() (int, error) {
	card, err := dac.GetRW().GetLastDriverCard(c.ctx, c.baseInfo.ID, c.chanInfo.ChannelID)
	if err != nil {
		return 0, err
	}
	return card.CardIndex, nil
}

// RedisDoorStatusKeyExpire 刷新Redis中门状态Key的过期时间
func (c *Controller) RedisDoorStatusKeyExpire() {
	redisKey := utils.GenerateRedisKeyDoorStatus(c.chanInfo.ChannelID)
	exists, err := c.redisClient.Exists(context.Background(), redisKey).Result()
	if err != nil {
		c.logger.Errorf("redis key exists err: %v", err)
		return
	}
	if exists == 0 {
		c.logger.Errorf("redis key: %s, not found", redisKey)
		return
	}

	err = c.redisClient.Expire(context.Background(), redisKey, consts2.RedisDefaultTimeout).Err()
	if err != nil {
		c.logger.Errorf("redis client expire error: %v, key: %s", err, redisKey)
	}
}

// ConnCheckAndReConnect 监听断连事件并自动重连
func (c *Controller) ConnCheckAndReConnect(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Infof("stop reconnect")
			return
		case <-c.Server.disConnectChan:
			c.logger.Warnf("recv from disConnectChan, try to reconnect...")
			for {
				time.Sleep(consts2.ConnCheckTime)
				if err := c.Server.Connect(c.ctx); err != nil {
					c.logger.Errorf("reconnect error: %v", err)
					continue
				}
				c.logger.Infof("reconnect success")
				break
			}
		}
	}
}

// hasLockAndOpen 检查分布式锁状态，获取锁后重新打开连接
func (c *Controller) hasLockAndOpen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(consts2.DLMAndReOpenCheckTime):
			if dlm.GetWorker().HasLock() && !c.isConnected.Load() {
				c.cancel()
				c.Open(c.chanInfo)
			}
		}
	}
}
