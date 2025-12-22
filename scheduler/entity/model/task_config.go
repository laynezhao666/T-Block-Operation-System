// Package model 调度器相关配置信息
package model

import (
	"encoding/json"
	"fmt"
	"scheduler/entity/consts"
	"scheduler/util/convutil"
	"strings"
)

const (
	defaultLockKey               = "lock_scheduler"       // 默认分布式锁key
	defaultLastVerStrKey         = "last_ver_str"         // 默认数据上次更新的key
	DefaultRegisterWorkKey       = "register_worker"      // 默认注册的worker列表的key
	defaultLastRegisterWorkerKey = "last_register_worker" // 默认上次注册的worker列表的key
	defaultLastAssignResultKey   = "last_assign_result"   // 默认上次分配结果的key

	TaskTypeAlarm     = "alarm"     // 告警调度任务类型
	TaskTypeCollector = "collector" // 采集调度任务类型
	TaskTypePoint     = "point"     // 测点计算调度任务类型
)

// AllTaskConfig 调度配置
type AllTaskConfig struct {
	Scheduler []*TaskConfig `yaml:"scheduler"` // 所有调度任务单元
}

// TaskConfig 调度单元配置,某一个调度任务
type TaskConfig struct {
	Name                  string  `yaml:"-"`                        // 调度任务名称
	Type                  string  `yaml:"type"`                     // 调度任务类型
	Disable               bool    `yaml:"disable"`                  // 是否禁用
	MysqlName             string  `yaml:"mysql_name"`               // 使用的MySQL实例名称
	RedisName             string  `yaml:"redis_name"`               // 使用的Redis实例名称
	IntervalSec           int     `yaml:"interval_sec"`             // 调度间隔(秒), 默认30s
	LockKey               string  `yaml:"lock_key"`                 // 分布式锁key
	LockKeyExpireSec      int     `yaml:"lock_key_expire_sec"`      // 分布式锁key过期时长(秒),默认60s
	LastVerStrKey         string  `yaml:"last_ver_str_key"`         // 记录数据上次更新的key
	RegisterWorkerKey     string  `yaml:"register_worker_key"`      // 获取注册的worker列表的key
	LastRegisterWorkerKey string  `yaml:"last_register_worker_key"` // 记录上次注册的worker列表的key
	LastAssignResultKey   string  `yaml:"last_assign_result_key"`   // 记录上次分配结果的key
	LastAllocateShardCnt  int     `yaml:"last_assign_shard_cnt"`    // 记录上次分配结果缓存分片数
	SetGroup              string  `yaml:"set_group"`                // 跨园区调度时,使用这个字段区分不同的园区
	FilterMozu            []int32 `yaml:"filter_mozu"`              // 需要调度的模组,默认为空代表全部模组
	OldVer                bool    `yaml:"old_ver"`                  // 是否是旧版本
}

// BuildDefaultAndValid 设置默认值并校验
func (s *TaskConfig) BuildDefaultAndValid() error {
	if s.Type != TaskTypeAlarm && s.Type != TaskTypeCollector && s.Type != TaskTypePoint {
		return fmt.Errorf("bad scheduler config [%v], type [%s] error, only allow [%s,%s,%s]",
			s, s.Type, TaskTypeAlarm, TaskTypeCollector, TaskTypePoint)
	}
	mozuListStr := convutil.SliceToStr(s.FilterMozu, "-")
	s.Name = strings.Join([]string{s.Type, s.SetGroup, mozuListStr}, consts.CommonFieldSeq)
	if s.MysqlName == "" {
		s.MysqlName = consts.TBosMySQLName
	}
	if s.RedisName == "" {
		s.RedisName = consts.TBosRedisName
	}
	if s.IntervalSec <= 0 {
		s.IntervalSec = 30 // 默认30s
	}
	redisKeyMark := strings.Join([]string{s.Type, s.SetGroup}, consts.RedisJoinFieldSep)
	if s.LockKey == "" {
		s.LockKey = strings.Join([]string{defaultLockKey, redisKeyMark}, consts.RedisJoinFieldSep)
	}
	if s.LockKeyExpireSec <= 0 {
		s.LockKeyExpireSec = 60 // 默认60s
	}
	if s.LastVerStrKey == "" {
		s.LastVerStrKey = strings.Join([]string{defaultLastVerStrKey, redisKeyMark}, consts.RedisJoinFieldSep)
	}
	if s.RegisterWorkerKey == "" {
		s.RegisterWorkerKey = strings.Join([]string{DefaultRegisterWorkKey, redisKeyMark}, consts.RedisJoinFieldSep)
	}
	if s.LastRegisterWorkerKey == "" {
		s.LastRegisterWorkerKey = strings.Join([]string{defaultLastRegisterWorkerKey, redisKeyMark}, consts.RedisJoinFieldSep)
	}
	if s.LastAssignResultKey == "" {
		s.LastAssignResultKey = strings.Join([]string{defaultLastAssignResultKey, redisKeyMark}, consts.RedisJoinFieldSep)
	}
	return nil
}

// CalcUniqueKey 计算配置任务唯一标识
func (s *TaskConfig) CalcUniqueKey() string {
	keyBytes, _ := json.Marshal(s)
	return string(keyBytes)
}

// CalcAllHistoryKey 计算所有和任务历史数据相关的Key
func (s *TaskConfig) CalcAllHistoryKey() []string {
	res := make([]string, 0, s.LastAllocateShardCnt+3)
	res = append(res, s.LastVerStrKey, s.LastRegisterWorkerKey)
	if s.LastAllocateShardCnt > 1 {
		for idx := 0; idx < s.LastAllocateShardCnt; idx++ {
			res = append(res, fmt.Sprintf("%s%s%d", s.LastAssignResultKey, consts.RedisJoinFieldSep, idx))
		}
	} else {
		res = append(res, s.LastAssignResultKey)
	}
	return res
}
