package chd806d4

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/driver"
	"dac/entity/model/rt"
	"dac/entity/redis"
	"dac/entity/utils"
	consts2 "dac/logic/collect/driver/chd806d4/consts"
	"dac/logic/dlm"

	"dac/entity/utils/tlog"
)

// Controller CHD806D4 协议控制器
type Controller struct {
	timeout  time.Duration
	baseInfo driver.ControllerBasicInfo
	chanInfo driver.ChannelInfo
	version  string
	Server   *DoorServer
	ctx      context.Context
	cancel   context.CancelFunc

	// 权限认证状态
	isAuthorized bool
	lastAuthTime time.Time
	authTimeout  time.Duration // 权限超时时间（4分钟）

	// Redis 客户端（用于门状态存储）
	redisClient *redis.Client

	// 日志记录器
	logger tlog.Logger

	// 门数量
	doorNum int

	// 连接状态（用于分布式锁和重连）
	isConnected atomic.Bool
}

// ============ 连接管理 ============

// Open 打开门控器连接（初始化配置、启动分布式锁检查和连接协程）
func (c *Controller) Open(chanInfo driver.ChannelInfo) consts.Quality {
	c.chanInfo = chanInfo
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.timeout = chanInfo.TimeoutMS
	c.authTimeout = 4 * time.Minute
	c.redisClient = redis.GetClient()
	c.logger = tlog.NewPrefixLogger(fmt.Sprintf("[CHD-Controller@%v]", chanInfo.ChannelID), config.Log)

	// 获取门数量
	c.doorNum = 2 // 默认2门
	if num, ok := chanInfo.Extend["door_num"].(int); ok && num > 0 {
		c.doorNum = num
	}

	// 启动分布式锁检查协程
	go c.hasLockAndOpen(c.ctx)

	// 检查分布式锁
	if !dlm.GetWorker().HasLock() {
		c.isConnected.Store(false)
		c.logger.Infof("未获取分布式锁，等待锁释放...")
		// 没拿到锁，返回OK，实际未连接门禁控制器
		return consts.QualityOK
	}

	// 执行实际的连接逻辑
	return c.doConnect()
}

// doConnect 执行实际的连接逻辑
func (c *Controller) doConnect() consts.Quality {
	// 防止重复连接：如果已经连接，直接返回
	if c.isConnected.Load() {
		return consts.QualityOK
	}

	// 1. 清理旧的 Server（如果存在）
	if c.Server != nil {
		c.Server.Disconnect()
		c.Server = nil
	}

	// 2. 创建服务器（直接使用ChannelID，格式：IP:Port）
	c.Server = NewDoorServer(c.chanInfo.ChannelID, c.timeout)

	// 3. 连接门控器
	if err := c.Server.Connect(c.ctx); err != nil {
		c.logger.Errorf("连接门控器失败: %v", err)
		c.isConnected.Store(false)
		return consts.QualityUncertain
	}

	// 4. 权限认证（使用默认密码：全0）
	defaultPassword := [5]byte{0x00, 0x00, 0x00, 0x00, 0x00}
	if err := c.AuthVerify(defaultPassword); err != nil {
		c.logger.Errorf("门控器权限认证失败: %v", err)
		c.Server.Disconnect()
		c.isConnected.Store(false)
		return consts.QualityUncertain
	}

	c.isConnected.Store(true)

	// 5. 启动事件处理协程
	go c.handleEvents(c.ctx)

	// 6. 启动断线重连协程（只启动一次，由 ConnCheckAndReConnect 内部处理重连）
	go c.ConnCheckAndReConnect(c.ctx)

	c.logger.Infof("门控器连接成功")
	return consts.QualityOK
}

// hasLockAndOpen 持续检查分布式锁状态，获取锁后自动连接
// 注意：这个函数只负责首次连接，断线重连由 ConnCheckAndReConnect 处理
func (c *Controller) hasLockAndOpen(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 只有在 Server 为 nil（从未连接过）或者 Server 存在但未连接时，才尝试连接
			// 如果 Server 已存在且 isConnected = false，说明是断线状态，由 ConnCheckAndReConnect 处理
			if dlm.GetWorker().HasLock() && !c.isConnected.Load() && c.Server == nil {
				c.logger.Infof("获取到分布式锁，开始连接门控器...")
				c.doConnect()
			}
		}
	}
}

// ConnCheckAndReConnect 断线重连
func (c *Controller) ConnCheckAndReConnect(ctx context.Context) {
	for {
		// 检查 Server 是否已初始化，避免空指针访问
		if c.Server == nil {
			select {
			case <-ctx.Done():
				c.logger.Infof("停止重连协程")
				return
			case <-time.After(time.Second):
				continue
			}
		}

		select {
		case <-ctx.Done():
			c.logger.Infof("停止重连协程")
			return
		case <-c.Server.disConnectChan:
			c.logger.Warnf("检测到连接断开，准备重连...")
			c.isConnected.Store(false)

			// 循环重试连接
			for {
				time.Sleep(5 * time.Second) // 每5秒重试一次

				// 检查上下文是否已取消
				select {
				case <-ctx.Done():
					return
				default:
				}

				// 检查是否有分布式锁
				if !dlm.GetWorker().HasLock() {
					c.logger.Debugf("未获取分布式锁，等待...")
					continue
				}

				// 尝试重新连接
				if err := c.Server.Connect(c.ctx); err != nil {
					c.logger.Errorf("重连失败: %v", err)
					continue
				}

				// 重新认证
				defaultPassword := [5]byte{0x00, 0x00, 0x00, 0x00, 0x00}
				if err := c.AuthVerify(defaultPassword); err != nil {
					c.logger.Errorf("重连后认证失败: %v", err)
					c.Server.Disconnect()
					continue
				}

				c.isConnected.Store(true)
				c.logger.Infof("重连成功")
				break
			}
		}
	}
}

// Close 关闭门控器连接（取消权限、断开连接、取消上下文）
func (c *Controller) Close() consts.Quality {
	// 1. 取消权限
	_ = c.AuthCancel()

	// 2. 关闭服务器
	if c.Server != nil {
		c.Server.Disconnect()
	}

	// 3. 取消上下文
	if c.cancel != nil {
		c.cancel()
	}

	c.isAuthorized = false
	c.logger.Infof("门控器已关闭")
	return consts.QualityOK
}

// Ping 检查门控器连接和权限状态
func (c *Controller) Ping() error {
	if c.Server == nil || !c.Server.IsConnected() {
		return fmt.Errorf("未连接到门控器")
	}

	// 检查权限是否超时
	if err := c.checkAuth(); err != nil {
		return err
	}

	return nil
}

// IsReady 检查门控器是否已连接且已认证
func (c *Controller) IsReady() bool {
	return c.Server != nil && c.Server.IsConnected() && c.isAuthorized
}

// ============ 门状态 Redis 存储 ============

// saveDoorStatusToRedis 保存门状态到 Redis（参考 xbrother）
func (c *Controller) saveDoorStatusToRedis(doorStates []int) error {
	if c.redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	doorStateJson, err := json.Marshal(doorStates)
	if err != nil {
		return err
	}

	return c.redisClient.Set(
		context.Background(),
		utils.GenerateRedisKeyCHDDoorStatus(c.chanInfo.ChannelID),
		doorStateJson,
		consts2.RedisDefaultTimeout,
	).Err()
}

// updateDoorStatusFromState 从门状态更新 Redis（在获取门状态后调用）
func (c *Controller) updateDoorStatusFromState(state map[int]*rt.Point) error {
	if state == nil || len(state) == 0 {
		return nil
	}

	// 根据门的数量构建状态数组
	// CHD 协议通常支持 1-4 门
	doorStates := make([]int, 4)
	for i := 0; i < 4; i++ {
		doorNo := i + 1
		if point, ok := state[doorNo]; ok && point != nil {
			// rt.Point 的 Rtd.Pv 存储门状态值
			// "1" 表示开门，"0" 表示关门
			if point.Rtd.Pv == "1" {
				doorStates[i] = consts2.DriverDoorStatusOpen
			} else {
				doorStates[i] = consts2.DriverDoorStatusClose
			}
		}
	}

	return c.saveDoorStatusToRedis(doorStates)
}

// handleEvents 处理触发包事件（从设备读取事件并保存）
func (c *Controller) handleEvents(ctx context.Context) {
	for {
		// 检查 Server 是否已初始化，避免空指针访问
		if c.Server == nil {
			select {
			case <-ctx.Done():
				c.logger.Infof("停止事件处理协程")
				return
			case <-time.After(time.Second):
				continue
			}
		}

		select {
		case <-ctx.Done():
			c.logger.Infof("停止事件处理协程")
			return
		case _, ok := <-c.Server.GetEventChan():
			if !ok {
				c.logger.Warnf("事件通道已关闭")
				return
			}
			c.logger.Debugf("收到触发包通知，准备读取事件")

			// 读取并处理记录（FetchAndSaveRecords 会先从设备读取记录，再调用 ProcessAndSaveRecords 保存）
			if err := c.FetchAndSaveRecords(); err != nil {
				c.logger.Errorf("处理记录失败: %v", err)
			}

			// 更新门状态到 Redis
			stateMap, err := c.GetDoorState([]int{1, 2, 3, 4})
			if err != nil {
				c.logger.Warnf("获取门状态失败: %v", err)
			} else {
				if err := c.updateDoorStatusFromState(stateMap); err != nil {
					c.logger.Warnf("更新门状态到 Redis 失败: %v", err)
				}
			}
		}
	}
}

// ============ 权限认证相关方法 ============

// AuthVerify 权限密码校验
func (c *Controller) AuthVerify(password [5]byte) error {
	// 发送请求（传递 RRPC Key 生成函数）
	respInfo, err := c.Server.Request(
		consts2.CID2AccessAuth,
		consts2.GroupAuth,
		consts2.TypeAuthVerify,
		password[:],
		consts2.GetRRPCAuthVerify, // 传递函数，而不是字符串
	)
	if err != nil {
		return fmt.Errorf("权限认证失败: %w", err)
	}

	// 检查响应（空响应表示成功）
	if len(respInfo) == 0 {
		c.isAuthorized = true
		c.lastAuthTime = time.Now()
		return nil
	}
	return fmt.Errorf("权限认证失败: 门控器返回非空响应")
}

// AuthCancel 取消访问权限
func (c *Controller) AuthCancel() error {
	if !c.isAuthorized {
		return nil // 已经没有权限了
	}

	// 发送请求（传递 RRPC Key 生成函数）
	_, err := c.Server.Request(
		consts2.CID2AccessAuth,
		consts2.GroupAuth,
		consts2.TypeAuthCancel,
		nil,
		consts2.GetRRPCAuthCancel, // 传递函数，而不是字符串
	)
	if err != nil {
		return fmt.Errorf("取消权限失败: %w", err)
	}

	c.isAuthorized = false
	return nil
}

// AuthModify 修改访问密码
func (c *Controller) AuthModify(newPassword [5]byte) error {
	// 检查权限
	if err := c.checkAuth(); err != nil {
		return err
	}

	// 计算校验码（5字节密码的异或值）
	checksum := newPassword[0] ^ newPassword[1] ^ newPassword[2] ^ newPassword[3] ^ newPassword[4]

	// 构建数据（5字节密码 + 1字节校验码）
	data := make([]byte, 6)
	copy(data[0:5], newPassword[:])
	data[5] = checksum

	// 发送请求（传递 RRPC Key 生成函数）
	_, err := c.Server.Request(
		consts2.CID2AccessAuth,
		consts2.GroupAuth,
		consts2.TypeAuthModify,
		data,
		consts2.GetRRPCAuthModify, // 传递函数，而不是字符串
	)
	if err != nil {
		return fmt.Errorf("修改密码失败: %w", err)
	}

	return nil
}

// checkAuth 检查权限状态（内部方法）
func (c *Controller) checkAuth() error {
	if !c.isAuthorized {
		return fmt.Errorf("未通过权限认证")
	}

	// 检查是否超时（4分钟）
	if time.Since(c.lastAuthTime) > c.authTimeout {
		c.logger.Infof("权限已超时，尝试重新认证...")

		// 自动重新认证
		defaultPassword := [5]byte{0x00, 0x00, 0x00, 0x00, 0x00}
		if err := c.AuthVerify(defaultPassword); err != nil {
			c.isAuthorized = false
			return fmt.Errorf("重新认证失败: %w", err)
		}

		c.logger.Infof("重新认证成功")
	}

	return nil
}
