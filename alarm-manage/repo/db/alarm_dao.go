package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	tgorm "etrpc-go/client/gorm"

	"etrpc-go/log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"trpc.group/trpc-go/trpc-go"

	"alarm-manage/entity/model"
	"alarm-manage/utils/modcall"
	cmodel "common/entity/model"
)

const (
	ALARM_MYSQL_NAME   = "trpc.mysql.tbos.alarm"
	ACTIVE_TABLE       = "t_alarm_active"
	HISTORY_TABLE      = "t_alarm_history"
	ALARM_WORKER_TABLE = "t_alarm_mworker"
)

// IAlarmDBApi 告警DB接口
type IAlarmDBApi interface {
	GetActiveFingerprints(fingerprint []string) (list []string, err error)
	GetActiveAlarmId(alarmIds []int64) (list []int64, err error)
	GetTotalActiveList(ctx context.Context, mozuId int32, alarmStatus int) ([]cmodel.AlarmActive, error)
	BatchInsertActiveAlerts(ctx context.Context, alarms []cmodel.AlarmActive) (successAlarms []cmodel.AlarmActive, err error)
	GetActiveListByFp(fingerprint []string) (list []cmodel.AlarmActive, err error)
	RestoreAlerts(historyList []cmodel.AlarmHistory) (successRestores []cmodel.AlarmHistory, err error)
	/*--------------------------------注册节点相关---------------------------------------*/
	GetAlarmWorkerIdList(ctx context.Context, occupyStatus int32) ([]int32, error)
	DelWorkerInfo(ctx context.Context, workerId int32, uuid string) error
	DelInvalidWorker(ctx context.Context) error
	InsertAlarmWorkerInfo(ctx context.Context, workerInfo model.AlarmWorker) error
	DBHeartBeat(ctx context.Context, workerId int32, uuid string, podIp string) error
}

// AlarmDBImpl 告警DB实现
type AlarmDBImpl struct {
	db *gorm.DB
}

var (
	alarmDBImpl *AlarmDBImpl
	once        sync.Once
)

// GetAlarmDBImpl 获取告警DB实现
func GetAlarmDBImpl() IAlarmDBApi {
	once.Do(func() {
		alarmDBImpl = &AlarmDBImpl{
			db: tgorm.GetDB(ALARM_MYSQL_NAME),
		}
	})
	return alarmDBImpl
}

// GetActiveFingerprints 获取当前活动告警的指纹
func (a *AlarmDBImpl) GetActiveFingerprints(fingerprint []string) (list []string, err error) {
	if len(fingerprint) == 0 {
		return
	}
	defer modcall.RecordDBReqCnt(err == nil)
	ret := a.db.Table(ACTIVE_TABLE).Where("fingerprint in (?)", fingerprint).Pluck("fingerprint", &list)
	if ret.Error != nil {
		err = ret.Error
		return
	}
	return
}

// GetActiveAlarmId 获取当前活动告警的ID
func (a *AlarmDBImpl) GetActiveAlarmId(alarmIds []int64) (list []int64, err error) {
	if len(alarmIds) == 0 {
		return
	}
	defer modcall.RecordDBReqCnt(err == nil)
	ret := a.db.Table(ACTIVE_TABLE).Where("alarm_id in (?)", alarmIds).Pluck("alarm_id", &list)
	if ret.Error != nil {
		err = ret.Error
		return
	}
	return
}

// GetTotalActiveList 获取全量活动告警
// @param alarmStatus 活动告警状态 0: 未挂起 1: 挂起
func (a *AlarmDBImpl) GetTotalActiveList(ctx context.Context, mozuId int32, alarmStatus int) (alarms []cmodel.AlarmActive, err error) {
	defer modcall.RecordDBReqCnt(err == nil)
	query := a.db.Table(ACTIVE_TABLE).Where("status = ?", alarmStatus)
	if mozuId > 0 {
		query = query.Where("mozu_id = ?", mozuId)
	}
	ret := query.Find(&alarms)
	if ret.Error != nil {
		err = ret.Error
		return
	}
	return
}

// BatchInsertActiveAlerts BatchInsertActiveAlerts
// 批量插入活动告警
// 根据写回的Id提取写入成功的Alarm
func (a *AlarmDBImpl) BatchInsertActiveAlerts(ctx context.Context, alarms []cmodel.AlarmActive) (successAlarms []cmodel.AlarmActive, err error) {
	successAlarms = make([]cmodel.AlarmActive, 0)
	if len(alarms) == 0 {
		return
	}
	// 设置5s超时时间
	cancelCtx, cancel := context.WithTimeout(trpc.BackgroundContext(), 5*time.Second)
	defer func() {
		cancel()
		modcall.RecordDBReqCnt(err == nil)
		for _, alarmItem := range alarms {
			modcall.RecordDBWriteCnt(int32(alarmItem.MozuId), 1, err == nil)
		}
	}()
	ret := a.db.WithContext(cancelCtx).Table(ACTIVE_TABLE).
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).CreateInBatches(&alarms, 1000)
	if ret.Error != nil {
		err = ret.Error
		return
	}
	for _, item := range alarms {
		if item.ID > 0 {
			successAlarms = append(successAlarms, item)
		}
	}
	if len(successAlarms) == 0 {
		log.Warnf("none of alarms inserted, total: %d", len(alarms))
	}
	return
}

// GetActiveListByFp 根据fingerprint获取活动告警
func (a *AlarmDBImpl) GetActiveListByFp(fingerprint []string) (list []cmodel.AlarmActive, err error) {
	if len(fingerprint) == 0 {
		return
	}
	defer modcall.RecordDBReqCnt(err == nil)
	ret := a.db.Table(ACTIVE_TABLE).Where("fingerprint in (?)", fingerprint).Find(&list)
	if ret.Error != nil {
		err = ret.Error
		return
	}
	return
}

// RestoreAlerts 批量关闭告警
//
// 事务两步操作 1. 活动告警拷贝到历史表 2. 删除活动告警
func (a *AlarmDBImpl) RestoreAlerts(historyList []cmodel.AlarmHistory) (successRestores []cmodel.AlarmHistory, err error) {
	successRestores = make([]cmodel.AlarmHistory, 0)
	if len(historyList) == 0 {
		err = fmt.Errorf("RestoreAlerts historyList is empty")
		return
	}
	alarmIDList := []int64{}
	for _, item := range historyList {
		alarmIDList = append(alarmIDList, int64(item.AlarmID))
	}
	// 设置10s超时时间
	cancelCtx, cancel := context.WithTimeout(trpc.BackgroundContext(), 10*time.Second)
	defer func() {
		cancel()
		modcall.RecordDBReqCnt(err == nil)
	}()
	err = a.db.WithContext(cancelCtx).
		Transaction(func(tx *gorm.DB) error {
			// 活动告警写入到历史告警表
			ret := tx.Table(HISTORY_TABLE).Clauses(clause.OnConflict{
				DoNothing: true,
			}).CreateInBatches(&historyList, 1000)
			if ret.Error != nil {
				err = ret.Error
				return err
			}
			// 删除活动告警表中的活动告警
			ret = tx.Table(ACTIVE_TABLE).Where("alarm_id in (?)", alarmIDList).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Delete(&cmodel.AlarmActive{})
			if ret.Error != nil {
				err = ret.Error
				return err
			}
			return nil
		})
	if err != nil {
		err = fmt.Errorf("RestoreAlerts db failed %w", err)
		return
	}
	for _, item := range historyList {
		if item.ID > 0 {
			successRestores = append(successRestores, item)
		}
	}
	if len(successRestores) == 0 {
		log.Warnf("none of history inserted, total: %d", len(historyList))
	}
	return
}
