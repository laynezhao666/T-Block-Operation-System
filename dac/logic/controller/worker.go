// Package controller 提供从CMDB同步门禁控制器数据的功能。
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/entity/model/tbox"
	"dac/logic/cgi/controller"
	"trpc.group/trpc-go/trpc-go/log"
	cmdbPb "trpcprotocol/cmdb"
)

// collectorDeviceDefaultSyncInterval 默认同步间隔（分钟）
// collectorDeviceTypeController 采集设备类型：门禁控制器
// defaultControllerAccount 默认控制器账号
// defaultControllerPassword 默认控制器密码
const (
	collectorDeviceDefaultSyncInterval = 60
	collectorDeviceTypeController      = 5
	defaultControllerAccount           = "admin"
	defaultControllerPassword          = "admin"
)

// w 全局Worker单例
var (
	w = &Worker{}
)

// Worker 负责从CMDB同步门控器数据的工作模块
type Worker struct {
	syncInterval time.Duration
}

// GetWorker 获取Worker实例
func GetWorker() *Worker {
	return w
}

// Init 初始化从cmdb同步门控器数据模块
func Init(ctx context.Context) {
	syncInterval := config.C.SyncInterval
	if syncInterval <= 0 {
		syncInterval = collectorDeviceDefaultSyncInterval
	}
	w.syncInterval = time.Duration(syncInterval) * time.Minute

	GetWorker().start(ctx)
}

// start 启动cmdb同步门控器数据任务
func (w *Worker) start(ctx context.Context) {
	go w.collectorDeviceSyncLoop(ctx)
}

// collectorDeviceSyncLoop 从cmdb同步门控器数据循环任务
func (w *Worker) collectorDeviceSyncLoop(ctx context.Context) {
	if config.C.IsSyncFromCMDB() {
		// 启动后立即执行一次同步
		w.syncCollectorDevice(ctx)
		config.Log.Infof("开始启动从cmdb同步门控器数据模块")
	} else {
		config.Log.Infof("当前配置下启动初始不从cmdb同步门控器数据模块")
	}

	for {
		select {
		case <-time.After(w.syncInterval):
			// 每次同步前检查配置，支持配置热更新停止同步
			if !config.C.IsSyncFromCMDB() {
				log.Infof("sync_from_cmdb=false，停止从cmdb同步门控器数据")
				return
			}
			w.syncCollectorDevice(ctx)
		case <-ctx.Done():
			config.Log.Infof("停止从cmdb同步门控器数据")
			return
		}
	}
}

// syncCollectorDevice 从cmdb同步门控器数据并创建记录
func (w *Worker) syncCollectorDevice(ctx context.Context) {
	// 1. 获取需要从cmdb同步门禁的模组id列表
	mozuIDs := config.C.CMDBSyncMozus
	if len(mozuIDs) == 0 {
		return
	}

	// 2. 查询采集设备类型为“门禁”的设备列表
	rspData, err := w.getCMDBControllers(ctx, mozuIDs)
	if err != nil {
		config.Log.Warnf("从cmdb获取采集设备列表失败，模组id列表：%v，error：%v", mozuIDs, err)
		return
	}

	// 3. 查询结果转化
	records, err := w.parseDoorController(ctx, rspData)
	if err != nil {
		config.Log.Warnf("门禁控制器数据转化失败，error：%v", err)
	}

	// 4. 批量创建门禁控制器设备
	// 判断是否需要更新门控器
	for mozuId, record := range records {
		err = controller.BatchCreate(ctx, mozuId, record, true)
		if err != nil {
			config.Log.Warnf("批量创建门禁控制器记录失败，error：%v", err)
			return
		}
	}

}

// getCMDBControllers 通过cmdb获取采集设备接口
func (w *Worker) getCMDBControllers(ctx context.Context, mozuIDs []string) (*cmdbPb.RspListCollectorDevice, error) {
	params, err := json.Marshal(mozuIDs)
	if err != nil {
		log.Errorf("unmarshal cmdb sync params err: %v", err)
		return nil, err
	}

	mozus := ParseStringToInt32(mozuIDs)
	req := &cmdbPb.ReqListCollectorDevice{
		MozuId:        mozus,
		CollectorType: []int32{collectorDeviceTypeController},
	}

	configClient := cmdbPb.NewConfigQueryClientProxy()

	rsp, err := configClient.ListCollectorDevice(ctx, req)
	if err != nil {
		log.Errorf("查询CMDB采集设备门控器配置失败，模组id：%s，err：%s", params, err)
		return nil, err
	}

	return rsp, nil
}

// parseDoorController cmdb查询门禁控制器数据转化为业务模型，按模组ID分组
func (w *Worker) parseDoorController(ctx context.Context, collectorDevice *cmdbPb.RspListCollectorDevice) (map[string][]rt.DoorController, error) {
	if len(collectorDevice.List) == 0 || collectorDevice.Total == 0 {
		return nil, fmt.Errorf("查询模组没有门控器数据")
	}

	records := make(map[string][]rt.DoorController)
	for _, device := range collectorDevice.List {
		mozuStr := strconv.Itoa(int(device.MozuId))
		record, err := convert(device)
		if err != nil {
			log.Warnf("解析采集设备门控器配置失败，设备编号：%s，模组id：%s，err：%s", device.DeviceCode, mozuStr, err)
			continue
		}
		records[mozuStr] = append(records[mozuStr], record)
	}

	return records, nil
}

// ParseStringToInt32 将字符串切片转换为int32切片，跳过无法解析的值
func ParseStringToInt32(mozuIDs []string) []int32 {
	result := make([]int32, 0, len(mozuIDs))
	for _, idStr := range mozuIDs {
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			log.Warnf("模组ID转换失败，跳过该ID：%s，error：%v", idStr, err)
			continue
		}
		result = append(result, int32(id))
	}

	return result
}

// convert 转换采集设备门控器配置为[]rt.DoorController
func convert(c *cmdbPb.CollectorDevice) (rt.DoorController, error) {
	var result rt.DoorController
	var channelLink rt.ChannelLink
	var protocol rt.Extend
	if len(c.ChannelLink) > 0 {
		if err := json.Unmarshal([]byte(c.ChannelLink), &channelLink); err != nil {
			config.Log.Warnf("Error unmarshalling channel_link JSON: %v", err)
			// 解析失败时使用默认值
			channelLink.Timeout = "3000" // 设置默认超时
		}
	} else {
		channelLink.Timeout = "3000"
	}
	if len(c.Extend) > 0 {
		if err := json.Unmarshal([]byte(c.Extend), &protocol); err != nil {
			config.Log.Warnf("Error unmarshalling extend JSON: %v", err)
		}
	}

	cc := db.DoorController{
		Name:     c.DeviceName,
		Profile:  db.Profile{},
		Position: tbox.Position{},
		GID:      db.GIDType(c.DeviceGid),

		Channel: tbox.ChannelRaw{
			ID:              c.ChannelId,
			RequestTimeout:  channelLink.Timeout,
			CommandInterval: "",
		},
		Protocol: db.Protocol{
			Name:    protocol.ProtocolName,
			Version: protocol.Protocol_version,
		},
		Extend:   make(map[string]interface{}),
		Account:  defaultControllerAccount,
		Password: defaultControllerPassword,
	}

	//// FetchInterval 即使没有从excel中读取到，也有初值(int)(cmdb暂无此字段，下面门数量同理)

	//if len(c.DoorNum) > 0 {
	//	doorNum, err := strconv.Atoi(c.DoorNum)
	//	if err != nil {
	//		return result, fmt.Errorf("parse door number error: %w", err)
	//	}
	//	cc.Extend[consts.KeyDoorNum] = doorNum
	//}

	result.DoorController = cc
	result.Doors = make([]db.Door, 0)

	return result, nil
}
