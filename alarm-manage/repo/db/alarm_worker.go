package db

import (
	"context"
	"fmt"
	"time"

	"alarm-manage/entity/model"
)

// DelWorkerInfo 程序终止，删除worker信息
func (a *AlarmDBImpl) DelWorkerInfo(ctx context.Context, workerId int32, uuid string) error {
	query := a.db.Table(ALARM_WORKER_TABLE).Where("worker_id = ? and uuid = ?", workerId, uuid)
	if err := query.Delete(&model.AlarmWorker{}).Error; err != nil {
		return err
	}
	return nil
}

// DelInvalidWorker 删除无效的worker ->
// 超过24小时未上报心跳 or occupy_status=0
func (a *AlarmDBImpl) DelInvalidWorker(ctx context.Context) error {
	query := a.db.Table(ALARM_WORKER_TABLE).Where("heartbeat < ? or occupy_status = ?", time.Now().Add(-24*time.Hour), 0)
	if err := query.Delete(&model.AlarmWorker{}).Error; err != nil {
		return err
	}
	return nil
}

// GetAlarmWorkerIdList 获取当前占用的workerId列表
func (a *AlarmDBImpl) GetAlarmWorkerIdList(ctx context.Context, occupyStatus int32) ([]int32, error) {
	idList := []int32{}
	query := a.db.Table(ALARM_WORKER_TABLE).Where("occupy_status = ?", occupyStatus)
	if err := query.Pluck("worker_id", &idList).Error; err != nil {
		return nil, err
	}
	return idList, nil
}

// InsertAlarmWorkerInfo 插入告警worker信息
func (a *AlarmDBImpl) InsertAlarmWorkerInfo(ctx context.Context, workerInfo model.AlarmWorker) error {
	if err := a.db.Table(ALARM_WORKER_TABLE).Create(&workerInfo).Error; err != nil {
		return err
	}
	return nil
}

// DBHeartBeat DB心跳上报
// 1. 检查workId对应的uuid是否一致
// 2. 更新心跳时间 podIp
func (a *AlarmDBImpl) DBHeartBeat(ctx context.Context, workerId int32, uuid string, podIp string) error {
	workerIdList := []int32{}
	query := a.db.Table(ALARM_WORKER_TABLE).Where("worker_id = ? and uuid = ?", workerId, uuid)
	if err := query.Pluck("worker_id", &workerIdList).Error; err != nil {
		return err
	}
	if len(workerIdList) != 1 {
		panic(fmt.Sprintf("pod占用的workerId数量不匹配, workerId:%d, uuid:%s, podIp:%s", workerId, uuid, podIp))
	}
	// 更新心跳时间
	if err := a.db.Table(ALARM_WORKER_TABLE).Where("worker_id = ?", workerId).
		Update("heartbeat", time.Now().Format(time.DateTime)).Update("pod_ip", podIp).Error; err != nil {
		return err
	}
	return nil
}
