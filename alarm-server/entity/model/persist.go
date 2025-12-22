package model

import (
	"alarm-server/entity/errcode/taskcode"
	cmodel "common/entity/model"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"etrpc-go/log"

	"github.com/samber/lo"
)

const (
	ValidKeyTemPlate = "%d:%s"
)

// StrategyCacheData 策略信息缓存
type StrategyCacheData struct {
	ID                   int64           `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	DeviceGid            string          `gorm:"column:device_gid" json:"device_gid"`
	Rid                  int64           `gorm:"column:rid" json:"rid"`
	RidVersion           string          `gorm:"column:rid_version" json:"rid_version"`
	RidType              int32           `gorm:"column:rid_type" json:"rid_type"`
	MozuId               int32           `gorm:"column:mozu_id" json:"mozu_id"`
	AlarmName            string          `gorm:"column:alarm_name" json:"alarm_name"`
	AlarmExpression      string          `gorm:"column:alarm_expression" json:"alarm_expression"`
	AlarmExpressionStr   string          `gorm:"column:alarm_expression_str" json:"alarm_expression_str"`
	RestoreExpression    string          `gorm:"column:restore_expression" json:"restore_expression"`
	RestoreExpressionStr string          `gorm:"column:restore_expression_str" json:"restore_expression_str"`
	ExpressionMap        *cmodel.ExprMap `gorm:"column:expression_map" json:"expression_map"`
	AlarmLevel           string          `gorm:"column:alarm_level" json:"alarm_level"`
	ContentTemplate      string          `gorm:"column:content_template" json:"content_template"`
	Owner                string          `gorm:"column:owner" json:"owner"`
	CreateAt             time.Time       `gorm:"column:create_at" json:"create_at"`
	UpdateAt             time.Time       `gorm:"column:update_at" json:"update_at"`
}

// GetPointList 获取策略关联测点列表
func (a *StrategyCacheData) GetPointList() []string {
	var exp = a.ExpressionMap
	if len(exp.Fire.ExprMap) == 0 {
		return []string{}
	}
	if len(exp.Restore.ExprMap) == 0 {
		return []string{}
	}
	return append(lo.Values(exp.Fire.ExprMap), lo.Values(exp.Restore.ExprMap)...)
}

// GetExprMapStr 获取测点映射字符串
func (a *StrategyCacheData) GetExprMapStr() string {
	str, err := json.Marshal(a.ExpressionMap)
	if err != nil {
		return ""
	}
	return string(str)
}

// ValidStoreData 策略生效信息
type ValidStoreData struct {
	MozuId      int32  `json:"mozu_id" gorm:"column:mozu_id"`
	Rid         int64  `json:"rid" gorm:"column:rid"`
	AlarmLevel  string `json:"alarm_level,omitempty" gorm:"column:alarm_level"`
	Gid         string `json:"gid" gorm:"column:gid"`
	EvalTime    int64  `json:"eval_time" gorm:"column:eval_time"`
	PvTime      int64  `json:"pv_time" gorm:"column:pv_time"`
	Success     bool   `json:"success" gorm:"column:success"`
	Fired       bool   `json:"fired" gorm:"column:fired"`
	ErrorCode   int32  `json:"error_code" gorm:"column:error_code"`
	ErrorName   string `json:"error_name" gorm:"column:error_name"`
	ErrorDetail string `json:"error_msg" gorm:"column:error_msg"`
}

// GetKey 获取key
func (v *ValidStoreData) GetKey() string {
	return fmt.Sprintf(ValidKeyTemPlate, v.Rid, v.Gid)
}

// GetErrDetail 获取错误详情
// 如果为确实测点，则将错误详情返回缺失测点列表
func (v *ValidStoreData) GetErrDetail(s *StrategyCacheData) string {
	missPointList := []string{}
	if v.ErrorCode == taskcode.PointDataLackErr.ErrCode {
		varList := strings.Split(v.ErrorDetail, ",")
		for _, sym := range varList {
			if point, ok := s.ExpressionMap.Fire.ExprMap[sym]; ok {
				missPointList = append(missPointList, point)
			} else {
				log.Errorf("sym:%s not found in fire, rid:%d, gid:%s", sym, s.Rid, s.DeviceGid)
			}
		}
	} else if v.ErrorCode == taskcode.RestoreAnalyzeErr.ErrCode {
		varList := strings.Split(v.ErrorDetail, ",")
		for _, sym := range varList {
			if point, ok := s.ExpressionMap.Restore.ExprMap[sym]; ok {
				missPointList = append(missPointList, point)
			} else {
				log.Errorf("sym:%s not found in restore, rid:%d, gid:%s", sym, s.Rid, s.DeviceGid)
			}
		}
	}
	if len(missPointList) > 0 {
		return strings.Join(missPointList, ",")
	} else {
		return v.ErrorDetail
	}
}
