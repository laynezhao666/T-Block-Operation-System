// Package alarm dao
package alarm

import (
	"context"
	"fmt"
	"time"

	tgorm "etrpc-go/client/gorm"
	"etrpc-go/log"

	"github.com/avast/retry-go"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cmodel "common/entity/model"

	pb "trpcprotocol/alarm-server"
)

const (
	AlarmDB         = "trpc.mysql.tbos.alarm"
	AlarmDBReadOnly = "trpc.mysql.tbos.alarm_readonly"
	ACTIVE_TABLE    = "t_alarm_active"
	HISTORY_TABLE   = "t_alarm_history"
	CntKeyTemplate  = "count_t%d_t%d"
)

var (
	ActiveMetricNameList  = []string{"level", "device_type_zh", "device_gid", "alarm_name", "fingerprint"}
	HistoryMetricNameList = []string{"level", "device_type_zh", "device_gid", "alarm_name", "fingerprint"}
)

// IAlarmDao ...
type IAlarmDao interface {
	GetActiveAlarmCnt(ctx context.Context, con *ActiveCntFilter) (int32, error)
	GetHistoryAlarmCnt(ctx context.Context, con *HistoryCntFilter) (int32, error)
	GetAlarmCntTrend(ctx context.Context, mozuId int32) ([]*pb.RspAlarmCntTrend_AlarmCount, error)
	GetActiveAlarmList(ctx context.Context, con *ActiveAlarmFilter) ([]*cmodel.AlarmActive, int64, map[string]*pb.RspAlarmList_VList, error)
	GetHistoryAlarmList(ctx context.Context, con *HistoryAlarmFilter) ([]*cmodel.AlarmHistory, int64, map[string]*pb.RspAlarmList_VList, error)
	GetAlarmName(ctx context.Context, qType int, page, size int) ([]string, int64, error)
	UpdateAlarmStatus(ctx context.Context, opType int64, con *AlarmStatusCon) error
	CloseAlarms(ctx context.Context, con *CloseStatusCon) error
	DelHistoryAlarm(ctx context.Context, con *DelHistoryAlarmCon) error
}

// NewAlarmDao 创建告警表相关操作实现类对象
func NewAlarmDao() IAlarmDao {
	return &alarmDaoImpl{
		db:     tgorm.GetDB(AlarmDB),
		readDB: tgorm.GetDB(AlarmDBReadOnly),
	}
}

type alarmDaoImpl struct {
	db     *gorm.DB // 主数据库
	readDB *gorm.DB // 只读库
}

func (a *alarmDaoImpl) GetActiveAlarmCnt(ctx context.Context, con *ActiveCntFilter) (int32, error) {
	query := a.readDB.Table(ACTIVE_TABLE).
		Where("mozu_id = ?", con.MozuId)
	if len(con.Level) > 0 {
		query = query.Where("level in ?", con.Level)
	}
	if len(con.Status) > 0 {
		query = query.Where("status in ?", con.Status)
	}
	if len(con.EventStatus) > 0 {
		query = query.Where("event_status in ?", con.EventStatus)
	}
	if con.Begin < con.End {
		query = query.Where("occur_time >= ?", time.Unix(con.Begin, 0).Format(time.DateTime))
		query = query.Where("occur_time <= ?", time.Unix(con.End, 0).Format(time.DateTime))
	} else {
		query = query.Where("occur_time <= ?", time.Unix(con.End, 0).Format(time.DateTime))
	}
	var activeCnt int64
	ret := query.Count(&activeCnt)
	if ret.Error != nil {
		return 0, fmt.Errorf("GetActiveAlarmCnt err: %s", ret.Error.Error())
	}
	return int32(activeCnt), nil
}

func (a *alarmDaoImpl) GetHistoryAlarmCnt(ctx context.Context, con *HistoryCntFilter) (int32, error) {
	query := a.readDB.Table(HISTORY_TABLE).
		Where("mozu_id = ?", con.MozuId)
	if len(con.Level) > 0 {
		query = query.Where("level in ?", con.Level)
	}
	if con.Begin < con.End {
		query = query.Where("occur_time >= ?", time.Unix(con.Begin, 0).Format(time.DateTime))
		query = query.Where("occur_time <= ?", time.Unix(con.End, 0).Format(time.DateTime))
	} else {
		query = query.Where("occur_time <= ?", con.Begin)
	}
	var count int64
	ret := query.Count(&count)
	if ret.Error != nil {
		return 0, fmt.Errorf("GetHistoryAlarmCnt err: %s", ret.Error.Error())
	}
	return int32(count), nil
}

func (a *alarmDaoImpl) GetAlarmCntTrend(ctx context.Context, mozuId int32) ([]*pb.RspAlarmCntTrend_AlarmCount, error) {
	now := time.Now()
	oneDayAgo := now.Add(-23 * time.Hour)
	oneDayAgoStr := oneDayAgo.Format("2006-01-02 15:00:00")
	// 查询过去一天内每小时的数据条数
	var activeResults []struct {
		Hour  string
		Count int64
	}
	var historyResults []struct {
		Hour  string
		Count int64
	}
	activeErr := a.readDB.Table(ACTIVE_TABLE).Select(
		"DATE_FORMAT(occur_time, '%Y-%m-%d %H:00:00') AS hour, COUNT(*) AS count").
		Where("mozu_id = ?", mozuId).Where("occur_time >= ?", oneDayAgoStr).
		Group("hour").Order("hour").Scan(&activeResults).Error
	historyErr := a.readDB.Table(HISTORY_TABLE).Select(
		"DATE_FORMAT(occur_time, '%Y-%m-%d %H:00:00') AS hour, COUNT(*) AS count").
		Where("mozu_id = ?", mozuId).Where("occur_time >= ?", oneDayAgoStr).
		Group("hour").Order("hour").Scan(&historyResults).Error
	if activeErr != nil || historyErr != nil {
		return nil, fmt.Errorf("GetAlarmCntTrend get Alarm Cnt err: %s-%s", activeErr.Error(), historyErr.Error())
	}
	data := []*pb.RspAlarmCntTrend_AlarmCount{}
	activeIndex, historyIndex := 0, 0
	for curHour := oneDayAgo; curHour.Unix() <= now.Unix(); curHour = curHour.Add(time.Hour) {
		item := &pb.RspAlarmCntTrend_AlarmCount{
			UTime: curHour.Format("2006-01-02 15:00:00"),
			Count: 0,
		}
		if activeIndex < len(activeResults) && activeResults[activeIndex].Hour == curHour.Format("2006-01-02 15:00:00") {
			item.Count += activeResults[activeIndex].Count
			activeIndex++
		}
		if historyIndex < len(historyResults) && historyResults[historyIndex].Hour == curHour.Format("2006-01-02 15:00:00") {
			item.Count += historyResults[historyIndex].Count
			historyIndex++
		}
		data = append(data, item)
	}
	return data, nil
}

func (a *alarmDaoImpl) getActiveAlarmListSql(con *ActiveAlarmFilter) *gorm.DB {
	query := a.readDB.Table(ACTIVE_TABLE).
		Where("mozu_id = ?", con.MozuId)
	if len(con.OccurBegin) > 0 {
		query = query.Where("occur_time >= ?", con.OccurBegin)
	}
	if len(con.OccurEnd) > 0 {
		query = query.Where("occur_time <= ?", con.OccurEnd)
	}
	if con.AlarmId > 0 {
		query = query.Where("alarm_id = ?", con.AlarmId)
	}
	if len(con.Level) > 0 {
		query = query.Where("level in ?", con.Level)
	}
	if len(con.Status) > 0 {
		query = query.Where("status in ?", con.Status)
	}
	if len(con.EventStatus) > 0 {
		query = query.Where("event_status in ?", con.EventStatus)
	}
	if len(con.DeviceGid) > 0 {
		query = query.Where("device_gid in ?", con.DeviceGid)
	}
	if len(con.DeviceNumber) > 0 {
		query = query.Where("device_number in ?", con.DeviceNumber)
	}
	if con.Rid > 0 {
		query = query.Where("rid = ?", con.Rid)
	}
	if len(con.AlarmName) > 0 {
		query = query.Where("alarm_name in ?", con.AlarmName)
	}
	if len(con.Content) > 0 {
		query = query.Where("content like ?", "%"+con.Content+"%")
	}
	return query
}

func (a *alarmDaoImpl) GetActiveAlarmList(ctx context.Context, con *ActiveAlarmFilter) ([]*cmodel.AlarmActive, int64, map[string]*pb.RspAlarmList_VList, error) {
	metricsMap := map[string]*pb.RspAlarmList_VList{}
	if con.CountByMetric {
		for _, metric := range ActiveMetricNameList {
			resList := []*cmodel.AlarmStatisticsGroup{}
			activeMetricQuery := a.getActiveAlarmListSql(con)
			ret := activeMetricQuery.Select(fmt.Sprintf("%s as name, count(*) as cnt", metric)).
				Group("name").Order("cnt desc").Limit(10).Scan(&resList)
			if ret.Error != nil {
				log.Errorf("GetActiveAlarmList getActiveAlarmListSql err: %s, name:", ret.Error.Error(), metric)
				continue
			}
			metricsMap[metric] = &pb.RspAlarmList_VList{
				List: []*pb.RspAlarmList_VItem{},
			}
			for _, item := range resList {
				metricsMap[metric].List = append(metricsMap[metric].List, &pb.RspAlarmList_VItem{
					Name:  item.Name,
					Count: item.Cnt,
				})
			}
		}
	}
	var count int64
	query := a.getActiveAlarmListSql(con)
	ret := query.Count(&count)
	if ret.Error != nil {
		return nil, 0, nil, ret.Error
	}
	switch con.SortType {
	case 1:
		query = query.Order("level asc, occur_time desc")
	case 2:
		query = query.Order("occur_time desc, level asc")
	default:
		{
		}
	}
	if con.Page > 0 && con.Size > 0 {
		query = query.Offset(int(con.Page-1) * int(con.Size)).Limit(int(con.Size))
	}
	activeList := []*cmodel.AlarmActive{}
	ret = query.Find(&activeList)
	if ret.Error != nil {
		return nil, 0, nil, ret.Error
	}
	return activeList, count, metricsMap, nil
}

func (a *alarmDaoImpl) getHistoryAlarmListBaseSql(con *HistoryAlarmFilter) *gorm.DB {
	query := a.readDB.Table(HISTORY_TABLE).
		Where("mozu_id = ?", con.MozuId)
	if len(con.Level) > 0 {
		query = query.Where("level in ?", con.Level)
	}
	if con.AlarmId > 0 {
		query = query.Where("alarm_id = ?", con.AlarmId)
	}
	if len(con.DeviceGid) > 0 {
		query = query.Where("device_gid in ?", con.DeviceGid)
	}
	if len(con.DeviceNumber) > 0 {
		query = query.Where("device_number in ?", con.DeviceNumber)
	}
	if len(con.AlarmName) > 0 {
		query = query.Where("alarm_name in ?", con.AlarmName)
	}
	if con.Rid > 0 {
		query = query.Where("rid = ?", con.Rid)
	}
	if len(con.OccurBegin) > 0 {
		query = query.Where("occur_time >= ?", con.OccurBegin)
	}
	if len(con.OccurEnd) > 0 {
		query = query.Where("occur_time <= ?", con.OccurEnd)
	}
	if con.RestoreBegin != "" {
		query = query.Where("restore_time >= ?", con.RestoreBegin)
	}
	if con.RestoreEnd != "" {
		query = query.Where("restore_time <= ?", con.RestoreEnd)
	}
	if con.MaxDuration > 0 {
		query = query.Where("restore_time - occur_time <= ?", con.MaxDuration)
	}
	if con.MinDuration > 0 {
		query = query.Where("restore_time - occur_time >= ?", con.MinDuration)
	}
	if len(con.Content) > 0 && con.Content[0] != '%' {
		query = query.Where("content like ?", con.Content+"%")
	}
	return query
}

func (a *alarmDaoImpl) GetHistoryAlarmList(ctx context.Context, con *HistoryAlarmFilter) ([]*cmodel.AlarmHistory, int64, map[string]*pb.RspAlarmList_VList, error) {
	metricsMap := map[string]*pb.RspAlarmList_VList{}
	if con.CountByMetric {
		for _, metric := range HistoryMetricNameList {
			resList := []*cmodel.AlarmStatisticsGroup{}
			metricsQuery := a.getHistoryAlarmListBaseSql(con)
			ret := metricsQuery.Select(fmt.Sprintf("%s as name, count(*) as cnt", metric)).
				Group("name").Order("cnt desc").Limit(10).Scan(&resList)
			if ret.Error != nil {
				log.Errorf("GetHistoryAlarmList getActiveAlarmListSql err: %s, name:", ret.Error.Error(), metric)
				continue
			}
			metricList := []*pb.RspAlarmList_VItem{}
			for _, item := range resList {
				metricList = append(metricList, &pb.RspAlarmList_VItem{
					Name:  item.Name,
					Count: item.Cnt,
				})
			}
			metricsMap[metric] = &pb.RspAlarmList_VList{
				List: metricList,
			}
		}
	}
	var count int64
	query := a.getHistoryAlarmListBaseSql(con)
	ret := query.Count(&count)
	if ret.Error != nil {
		return nil, 0, nil, ret.Error
	}
	switch con.SortType {
	case 1:
		query = query.Order("level asc, occur_time desc")
	case 2:
		query = query.Order("occur_time desc, level asc")
	default:
		{
		}
	}
	historyList := []*cmodel.AlarmHistory{}
	if con.Page > 0 && con.Size > 0 {
		query = query.Offset(int(con.Page-1) * int(con.Size)).Limit(int(con.Size))
	} else {
		query = query.Limit(100000)
	}
	ret = query.Find(&historyList)
	if ret.Error != nil {
		return nil, 0, nil, ret.Error
	}
	return historyList, count, metricsMap, nil
}

func (a *alarmDaoImpl) GetAlarmName(ctx context.Context, qType int, page, size int) ([]string, int64, error) {
	query := a.readDB
	if qType == 1 {
		query = query.Table(ACTIVE_TABLE).Where("status = ?", 0).Distinct("alarm_name")
	} else if qType == 2 {
		query = query.Table(ACTIVE_TABLE).Where("status = ?", 1).Distinct("alarm_name")
	} else if qType == 3 {
		query = query.Table(HISTORY_TABLE).Distinct("alarm_name")
	} else {
		log.Errorf("GetAlarmName qType error: %d", qType)
		return nil, 0, fmt.Errorf("GetAlarmName qType error: %d", qType)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	alarmNameList := []string{}
	if page <= 0 || size <= 0 {
		ret := query.Find(&alarmNameList)
		if ret.Error != nil {
			return nil, 0, ret.Error
		}
	} else {
		ret := query.Offset((page - 1) * size).Limit(size).Find(&alarmNameList)
		if ret.Error != nil {
			return nil, 0, ret.Error
		}
	}
	return alarmNameList, count, nil
}

func (a *alarmDaoImpl) UpdateAlarmStatus(ctx context.Context, opType int64, con *AlarmStatusCon) error {
	query := a.db.Table(ACTIVE_TABLE).
		Where("mozu_id = ?", con.MozuId).
		Where("alarm_id in (?)", con.AlarmIds)
	updateTime := time.Now().Format(time.DateTime)
	if len(con.UpdateTime) > 0 {
		updateTime = con.UpdateTime
	}
	updates := map[string]interface{}{
		"update_time": updateTime,
	}
	if opType == 0 {
		updates["status"] = con.AlarmStatus
		updates["hangup_reason"] = con.HangupReason
		updates["hangup_user_id"] = con.UserId
	} else if opType == 1 {
		updates["event_status"] = con.EventStatus
	} else {
		return fmt.Errorf("UpdateAlarmStatus opType error: %d", opType)
	}
	ret := query.Updates(updates)
	if ret.Error != nil {
		return fmt.Errorf("UpdateAlarmStatus err: %s", ret.Error.Error())
	}
	return nil
}

// CloseAlarms 服务台手动 关闭告警
func (a *alarmDaoImpl) CloseAlarms(ctx context.Context, con *CloseStatusCon) error {
	query := a.db.Table(ACTIVE_TABLE).
		Where("mozu_id = ?", con.MozuId).
		Where("alarm_id in (?)", con.AlarmIds)
	activeList := []*cmodel.AlarmActive{}
	ret := query.Find(&activeList)
	if ret.Error != nil {
		return fmt.Errorf("CloseAlarms: get active alarm err: %s", ret.Error.Error())
	}
	historyList := lo.Map(activeList, func(active *cmodel.AlarmActive, index int) *cmodel.AlarmHistory {
		return &cmodel.AlarmHistory{
			AlarmID:        active.AlarmID,
			Level:          active.Level,
			OccurTime:      active.OccurTime,
			Rid:            active.Rid,
			MozuId:         active.MozuId,
			AlarmName:      active.AlarmName,
			Content:        active.Content,
			AnalyzeResult:  active.AnalyzeResult,
			FingerPrint:    active.FingerPrint,
			DeviceGid:      active.DeviceGid,
			DeviceNumber:   active.DeviceNumber,
			BoxName:        active.BoxName,
			RoomName:       active.RoomName,
			RestoreTime:    time.Now(),
			CreateAt:       time.Now(),
			ActiveCreateAt: active.CreateAt,
			DeviceTypeZh:   active.DeviceTypeZh,
			OpUser:         fmt.Sprintf("%d", con.UserId),
			OpReason:       con.CloseReason,
		}
	})
	var err error
	retry.Do(func() error {
		err = a.db.Transaction(func(tx *gorm.DB) error {
			// 活动告警写入到历史告警表
			ret := tx.Table(HISTORY_TABLE).Clauses(clause.OnConflict{
				DoNothing: true,
			}).CreateInBatches(&historyList, 1000)
			if ret.Error != nil {
				historyErr := ret.Error
				return historyErr
			}
			// 删除活动告警表中的活动告警
			ret = tx.Table(ACTIVE_TABLE).Where("alarm_id in (?)", con.AlarmIds).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Delete(&cmodel.AlarmActive{})
			if ret.Error != nil {
				activeErr := ret.Error
				return activeErr
			}
			return nil
		})
		return err
	}, retry.Attempts(10), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	if err != nil {
		return fmt.Errorf("CloseAlarms: update history and active err: %s", err.Error())
	}
	return nil
}

func (a *alarmDaoImpl) DelHistoryAlarm(ctx context.Context, con *DelHistoryAlarmCon) error {
	query := a.db.Table(HISTORY_TABLE).
		Where("mozu_id = ?", con.MozuId)
	if len(con.DeviceGid) > 0 {
		query = query.Where("device_gid in (?)", con.DeviceGid)
	}
	if len(con.Rid) > 0 {
		query = query.Where("rid in (?)", con.Rid)
	}
	if len(con.Level) > 0 {
		query = query.Where("level in (?)", con.Level)
	}
	if len(con.EndTime) > 0 {
		query = query.Where("occur_time <= ?", con.EndTime)
	} else {
		// 最近一个月内的历史告警不删除
		limitTime := time.Now().Add(-time.Hour * 24 * 30)
		query = query.Where("occur_time <= ?", limitTime)
	}
	ret := query.Delete(&cmodel.AlarmHistory{})
	if ret.Error != nil {
		return fmt.Errorf("DelHistoryAlarm: delete history err: %s", ret.Error.Error())
	}
	return nil
}
