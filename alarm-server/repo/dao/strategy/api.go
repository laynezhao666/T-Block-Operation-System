package strategy

import (
	"context"
	"fmt"

	tgorm "etrpc-go/client/gorm"

	"gorm.io/gorm"

	cmodel "common/entity/model"
)

const (
	AlarmDB         = "trpc.mysql.tbos.alarm"
	AlarmDBReadOnly = "trpc.mysql.tbos.alarm_readonly"
	Strategy_TABLE  = "t_alarm_strategy"
)

// IStrategyDao ...
type IStrategyDao interface {
	GetAlarmName(ctx context.Context, offset, num int) ([]string, int64, error)
	GetStrategyList(ctx context.Context, con *StrategyFilter) ([]cmodel.AlarmStrategy, int64, error)
}

// NewStrategyDao 创建告警表相关操作实现类对象
func NewStrategyDao() IStrategyDao {
	return &strategyDaoImpl{
		db:     tgorm.GetDB(AlarmDB),
		readDB: tgorm.GetDB(AlarmDBReadOnly),
	}
}

type strategyDaoImpl struct {
	db     *gorm.DB // 写库
	readDB *gorm.DB // 只读库
}

func (s *strategyDaoImpl) GetAlarmName(ctx context.Context, page, size int) ([]string, int64, error) {
	query := s.readDB.Table(Strategy_TABLE).Distinct("alarm_name")
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	alarmNameList := []string{}
	if page == 0 || size == 0 {
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

func (s *strategyDaoImpl) GetStrategyList(ctx context.Context, con *StrategyFilter) ([]cmodel.AlarmStrategy, int64, error) {
	if con.MozuId == 0 && !con.MozuAllowZero {
		return nil, 0, fmt.Errorf("mozuId is zero")
	}
	query := s.readDB.Table(Strategy_TABLE)
	if con.MozuId > 0 {
		query = query.Where("mozu_id = ?", con.MozuId)
	}
	if len(con.RidType) > 0 {
		query = query.Where("rid_type in (?)", con.RidType)
	}
	if len(con.Rid) > 0 {
		query = query.Where("rid in (?)", con.Rid)
	}
	if len(con.Gid) > 0 {
		query = query.Where("device_gid in (?)", con.Gid)
	}
	if len(con.DeviceNumber) > 0 {
		query = query.Where("device_number in (?)", con.DeviceNumber)
	}
	if len(con.Level) > 0 {
		query = query.Where("alarm_level in (?)", con.Level)
	}
	if len(con.ApplyType) > 0 {
		query = query.Where("application_type_zh in (?)", con.ApplyType)
	}
	if len(con.AlarmName) > 0 {
		query = query.Where("alarm_name in (?)", con.AlarmName)
	}
	if len(con.DeviceType) > 0 {
		query = query.Where("device_type_zh in (?)", con.DeviceType)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if con.Page > 0 && con.Size > 0 {
		query = query.Offset(int(con.Page-1) * int(con.Size)).Limit(int(con.Size))
	}
	activeList := []cmodel.AlarmStrategy{}
	if err := query.Find(&activeList).Error; err != nil {
		return nil, 0, err
	}
	return activeList, count, nil
}
