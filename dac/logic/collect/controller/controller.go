// Package controller 实现门禁控制器的数据采集和测点管理。
package controller

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/rt"
	"dac/entity/utils"
	"dac/logic/collect/controller/virtualpoints"
	"dac/logic/collect/rtdb"
	"dac/logic/collect/template"
	"dac/logic/dlm"
	"dac/logic/push"

	"dac/entity/utils/flog"
	"dac/entity/utils/tlog"
)

// defaultCommandIntervalMs 默认命令间隔（毫秒）
const (
	defaultCommandIntervalMs = 500
)

// Controller 门禁控制器采集实例，负责设备连接、测点采集和数据推送
type Controller struct {
	record rt.DoorController
	// 门编号
	doorNumbers []int
	// 门编号对应的 gid
	doorGIDs     map[int]db.GIDType
	doorGIDMutex sync.RWMutex
	controller   driver.Controller
	// 是否尝试打开过驱动
	isDriverOpenCalled bool
	// 虚拟测点
	virtualPoints virtualpoints.VirtualPoints
	template      *template.Template
	ctx           context.Context

	cancel      context.CancelFunc
	refreshChan chan struct{}

	logger       tlog.Logger
	filterLogger *flog.Filter

	commandInterval time.Duration
}

// NewController 创建门禁控制器采集实例，初始化门编号、模板和虚拟测点
func NewController(ctx context.Context, record rt.DoorController) *Controller {
	attrs := map[string]string{
		consts.AttrChannel:  record.Channel.ID,
		consts.AttrProtocol: record.Protocol.Name,
	}

	c := new(Controller)
	c.record = record
	c.logger = tlog.NewPrefixLogger(fmt.Sprintf("[%v@%v]", record.Channel.ID, record.ID), config.Log)
	c.filterLogger = flog.NewFilterLoggerWithContext(ctx, time.Minute*10, c.logger)

	// 如果没有设定命令间隔，则默认间隔为0，不等待
	if len(record.Channel.CommandInterval) > 0 {
		commandIntervalMs, err := strconv.ParseInt(record.Channel.CommandInterval, 10, 64)
		if err != nil {
			c.Warnf("parse command interval \"%v\" error: %v", record.Channel.CommandInterval, err)
			commandIntervalMs = defaultCommandIntervalMs
		} else if commandIntervalMs < 0 {
			c.Warnf("command interval: \"%v\" < 0", record.Channel.CommandInterval)
			commandIntervalMs = defaultCommandIntervalMs
		}
		c.commandInterval = time.Duration(commandIntervalMs) * time.Millisecond
	}

	doorLen := len(record.Doors)
	c.doorNumbers = make([]int, 0, doorLen)
	for i := range record.Doors {
		c.doorNumbers = append(c.doorNumbers, record.Doors[i].Number)
	}
	c.doorGIDs = make(map[int]db.GIDType, doorLen)

	t := template.GetManager().GetTemplate(record.Protocol.Name)
	if t == nil {
		c.Warnf("get template \"%v\" failed", record.Protocol.Name)
		return nil
	}
	c.template = t

	c.virtualPoints = virtualpoints.NewVirtualPoints(c.ID(), attrs)

	c.refreshChan = make(chan struct{}, 1)

	c.ctx, c.cancel = context.WithCancel(ctx)
	go c.refreshGIDLoop(c.ctx)

	return c
}

// GetCollectCode 获取控制器的采集编码
func (c *Controller) GetCollectCode() string {
	return c.record.GetCollectCode()
}

// ID 获取控制器ID
func (c *Controller) ID() db.IDType {
	return c.record.ID
}

// MozuID 获取控制器所属模组ID
func (c *Controller) MozuID() string {
	return c.record.MozuID
}

// Close 关闭控制器采集，释放资源
func (c *Controller) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.controller != nil {
		_ = c.controller.Close()
	}
}

// initPoints 初始化门状态和门告警的测点数据，质量标记为不确定
func (c *Controller) initPoints(doors []int) map[string]map[int]*rt.Point {
	states := make(map[int]*rt.Point)
	alarms := make(map[int]*rt.Point)
	for _, d := range doors {
		p := new(rt.Point)
		p.ID = utils.GenerateDoorStateID(c.ID(), d)
		p.SetValue("0").SetQua(consts.QualityUncertain)
		states[d] = p

		p = new(rt.Point)
		p.ID = utils.GenerateDoorOpenAlarmID(c.ID(), d)
		p.SetValue("0").SetQua(consts.QualityUncertain)
		alarms[d] = p
	}
	return map[string]map[int]*rt.Point{
		consts.StandardIDDoorState: states,
		consts.StandardIDOpenAlarm: alarms,
	}
}

// pushPoints 推送门测点数据
func (c *Controller) pushPoints(pointsData map[string]map[int]*rt.Point) {
	// 有些们测点因为推送成本问题不配置gid，忽略那些模组来兼容这种情况
	if config.C.IgnoreGID(c.MozuID()) {
		return
	}
	toPushPoints := make(rt.Points, 0, len(c.doorNumbers)<<2)
	for pointID, data := range pointsData {
		for d, p := range data {
			gid, ok := c.GetGID(d)
			if !ok {
				c.NotifyRefreshGID()
				continue
			}
			newPoint := *p
			newPoint.ID = utils.GeneratePointID(gid, pointID)
			toPushPoints = append(toPushPoints, newPoint)
		}
	}
	for _, d := range c.doorNumbers {
		gid, ok := c.GetGID(d)
		if !ok {
			c.NotifyRefreshGID()
			continue
		}
		commPoint := c.virtualPoints.GetCommStatusPoint()
		commPoint.ID = utils.GeneratePointID(gid, consts.StandardIDCommunicationState)

		faultPoint := c.virtualPoints.GetCommStatusPoint()
		faultPoint.ID = utils.GeneratePointID(gid, consts.StandardIDFaultStatus)

		toPushPoints = append(toPushPoints, commPoint, faultPoint)
	}
	push.GetWorker().SetPoints(c.ID(), toPushPoints, true)
}

// DoRequest 执行一次采集请求，获取门测点数据并推送
func (c *Controller) DoRequest(ctx context.Context) {
	// 未获取到锁时，需要打开设备
	// 因为后续开关门等操作可能会分发到未获取到锁的 pod 执行
	c.tryOpenDevice(ctx)

	if !dlm.GetWorker().HasLock() {
		time.Sleep(consts.RedisLockExpireTime)
		return
	}

	// 初始化门状态和门告警的测点数据
	initPoints := c.initPoints(c.doorNumbers)

	reqStartTime := time.Now()
	pointsData, err := c.controller.GetDoorPoints(c.doorNumbers)
	if err != nil {
		c.filterLogger.Warnf("GetDoorPoints", "get door points error: %v", err)
	}
	reqEndTime := time.Now()

	costTime := reqEndTime.Sub(reqStartTime).Milliseconds()
	c.virtualPoints.AddPeriodCostTime(costTime)

	reqSuccess := err == nil
	// 部分模组门禁性能原因，最大连续中断次数单独考虑
	interrupted := c.virtualPoints.UpdateAfterOneRequestFinished(reqSuccess, c.MozuID())
	c.virtualPoints.ReportComm(interrupted)

	if interrupted {
		c.filterLogger.Warnf(1, "door controller: %v 通讯中断中...", c.record.ID)
	}

	// 插入未获取到的测点，初始化为不确定的测点，string是测点名，int是门编号
	if pointsData == nil {
		pointsData = make(map[string]map[int]*rt.Point)
	}
	for pointID, oldData := range initPoints {
		currentData, ok := pointsData[pointID]
		if !ok {
			pointsData[pointID] = oldData
			continue
		}
		for d, p := range oldData {
			if _, ok = currentData[d]; !ok {
				currentData[d] = p
			}
		}
	}

	if !reqSuccess {
		for _, data := range pointsData {
			for _, p := range data {
				p.Rtd.Qua = consts.QualityUncertain
				p.Rtd.Timestamp = reqEndTime.UnixMilli()
			}
		}
	}

	c.pushPoints(pointsData)

	points := make(rt.Points, 0, len(pointsData)*len(c.doorNumbers))
	for _, data := range pointsData {
		for _, p := range data {
			points = append(points, *p)
		}
	}
	rtdb.SetPoints(points, true)

	c.virtualPoints.UpdateAfterOnePeriodFinished()
	c.virtualPoints.ResetValueAfterOnePeriod()
}

// tryOpenDevice 尝试打开设备连接，仅在首次调用时执行
func (c *Controller) tryOpenDevice(ctx context.Context) {
	if c.isDriverOpenCalled {
		return
	}

	r := c.doOpenDevice()
	c.Infof("start collecting, protocol: %+v, return code: %v", c.record.Protocol, r)
	if r == consts.QualityOK {
		go func() {
			for {
				// 仅当持有锁时才获取门参数及下发时间组
				if dlm.GetWorker().HasLock() {
					c.Infof("add request: get door parameters and set time group")
					go c.addGetDoorParameterRequest()
					go c.addSetTimeGroupRequest()
					return
				}
				select {
				case <-ctx.Done():
					c.Infof("stop try add request: get door parameters and set time group")
					return
				case <-time.After(time.Second):
				}
			}
		}()
	}

	c.isDriverOpenCalled = true
}

// doOpenDevice 执行设备连接操作，创建驱动实例并打开通道
func (c *Controller) doOpenDevice() consts.Quality {
	c.virtualPoints.Clear()

	if c.controller == nil {
		c.controller = c.template.GetDriver().CreateController(c.record.ID, c.record.Name)
		if c.controller == nil {
			c.Warnf("create controller %+v failed", c.record)
			return consts.QualityUncertain
		}
	}

	t, err := strconv.ParseInt(c.record.Channel.RequestTimeout, 0, 64)
	if err != nil || t <= 0 || t >= 60000 {
		t = 10000
		c.Warnf("timeout \"%v\" is invalid, use default value: %v", c.record.Channel.RequestTimeout, consts.DefaultTimeoutMS)
	}

	ch := driver.ChannelInfo{
		ChannelID:       c.record.Channel.ID,
		Address:         c.record.Channel.Address,
		Protocol:        c.record.Protocol.Name,
		ProtocolVersion: c.record.Protocol.Version,
		Extend:          c.record.Extend,
		TimeoutMS:       time.Millisecond * time.Duration(t),
	}

	return c.controller.Open(ch)
}

// WaitBetweenRequest 在两次请求之间等待指定的命令间隔
func (c *Controller) WaitBetweenRequest() {
	if c.commandInterval > 0 {
		time.Sleep(c.commandInterval)
	}
}
