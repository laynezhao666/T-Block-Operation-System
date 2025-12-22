package snowflake

import (
	"context"
	"fmt"
	"time"

	"alarm-manage/conf"
	"alarm-manage/entity/model"
	"alarm-manage/repo/db"
	"alarm-manage/repo/rpc"

	"etrpc-go/log"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-go"
)

// snowflakeNode 雪花算法节点Id -> 片区(Set)ID
var snowflakeNodeId int32

// snowflakeSubnode 雪花算法子节点
var snowflakeSubnodeId int32

// ps: 目前最多支持 2^SubnodeBits 个节点，目前 SubnodeBits 为 8，即支持256个节点
// 考虑到pod滚动更新，滚动更新间隔时间如果超过 30s（即 sfSubnodeTTL），则最多允许 15 个节点

// InitSnowflake InitSnowflake
func InitSnowflake(ctx context.Context) {
	err := initSnow(ctx)
	if err != nil {
		panic(err)
	}
}

/*
	initSnow 初始化雪花算法

修改redis强依赖 -> mysql依赖
 1. 从数据库中删除过期workerId
 2. 获取当前被占用的节点集合
 3. 通过redis分布式锁获取一个未被占用的子节点
 4. 更新数据库
 5. 释放锁，定期同步心跳
*/
func initSnow(ctx context.Context) error {
	var subnodeList []int32
	var subnodeMax int32 = -1 ^ (-1 << SubnodeBits)
	for i := int32(0); i < subnodeMax; i++ {
		subnodeList = append(subnodeList, i)
	}
	uidV1, err := uuid.NewV1()
	if err != nil {
		return fmt.Errorf("get uuid failed %w", err)
	}
	newSubnode := int32(-1)
	uid := uidV1.String()
	// 删除过期workerId
	if err := db.GetAlarmDBImpl().DelInvalidWorker(trpc.BackgroundContext()); err != nil {
		log.AlarmContextf(ctx, "删除无效节点失败, podIp:%s, err:%v",
			trpc.GlobalConfig().Global.LocalIP, err)
	}
	// 获取当前被占用的节点集合
	occupyIdList, err := db.GetAlarmDBImpl().GetAlarmWorkerIdList(ctx, 1)
	if err != nil {
		return fmt.Errorf("initSnow GetAlarmWorkerIdList failed %w", err)
	}
	for _, subnode := range subnodeList {
		if lo.Contains(occupyIdList, subnode) {
			// subnode 被用了，则换下一个 subnode 继续尝试
			continue
		}
		key := getSnowflakeLockKey(subnode)
		lockSuccess := rpc.NewRedisApi().TryLock(key, 10*time.Minute)
		if !lockSuccess {
			// 该节点被其他Pod抢到了，换下一个 subnode 继续尝试
			continue
		}
		// 构建worker model
		alarmWorker := model.NewAlarmWorker(uid, conf.ServerConf.SnowflakeConfig.SetId, subnode)
		// 插入数据库
		if err := db.GetAlarmDBImpl().InsertAlarmWorkerInfo(ctx, *alarmWorker); err != nil {
			log.Errorf("workerId插入数据库失败, workerId:%d, uid:%s, podIp:%d, err:%v",
				subnode, uid, alarmWorker.PodIp, err)
			rpc.NewRedisApi().UnLock(key)
			continue
		}
		newSubnode = subnode
		rpc.NewRedisApi().UnLock(key)
		break
	}
	if newSubnode < 0 {
		// 子节点列表中的节点都没有抢到，有异常，报错
		return fmt.Errorf("no snowflake subnode can be locked, subnodeMax: %d", subnodeMax)
	}
	snowflakeNodeId = conf.ServerConf.SnowflakeConfig.SetId
	snowflakeSubnodeId = newSubnode
	log.Infof("workerId of pod ip: %s, nodeId:%d, subnodeId:%d",
		trpc.GlobalConfig().Global.LocalIP, snowflakeNodeId, snowflakeSubnodeId)
	go keepWorker(ctx, snowflakeSubnodeId, uid)
	return nil
}

func getSnowflakeLockKey(subnode int32) string {
	key := fmt.Sprintf("alarm_snowflake_%d", subnode)
	return key
}

// keepWorker 定期更新worker信息
// interval 单位小时
func keepWorker(ctx context.Context, workerId int32, uid string) {
	interval := conf.ServerConf.SnowflakeConfig.UpdateInterval
	if interval == 0 {
		log.Errorf("conf.ServerConf.SnowflakeConfig.UpdateInterval is 0, use default 1 hour")
		interval = 1
	}
	itv := time.Hour * time.Duration(interval)
	tick := time.NewTicker(itv)
	for {
		select {
		case <-ctx.Done():
			// 程序结束，退出。退出前删除worker信息
			db.GetAlarmDBImpl().DelWorkerInfo(ctx, workerId, uid)
			return
		case <-tick.C:
			err := db.GetAlarmDBImpl().DBHeartBeat(ctx, workerId, uid, trpc.GlobalConfig().Global.LocalIP)
			if err != nil {
				log.Errorf("workerId上报心跳失败, workerId:%d, uid:%s, podIp:%s, err:%v",
					workerId, uid, trpc.GlobalConfig().Global.LocalIP, err)
			}
		}
	}
}

// GenerateAlarmId 生成告警唯一Id
func GenerateAlarmId() (ID, error) {
	/*
		使用 片区Id，和片区内Pod的subnodeId 作为唯一标识workerId，保证alarmId的唯一性
	*/
	sf, err := GetNode(int64(snowflakeNodeId), int64(snowflakeSubnodeId))
	if err != nil {
		return 0, err
	}
	return sf.Generate(), nil
}
