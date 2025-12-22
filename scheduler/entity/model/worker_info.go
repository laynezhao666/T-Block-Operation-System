// Package model 定义调度Worker实体信息结构
package model

import (
	"encoding/json"
	"fmt"
	"scheduler/entity/consts"
	"strings"
)

// WorkerInfo Worker实例信息
type WorkerInfo struct {
	Ip             string `json:"ip"`             // IP
	Port           int32  `json:"port"`           // 端口
	StartTime      int64  `json:"startTime"`      // 启动版本号
	MaxProcessCap  int64  `json:"maxProcessCap"`  // 最大处理能力
	TaskVerMark    string `json:"taskVerMark"`    // 任务版本标识
	WorkerProtocol string `json:"workerProtocol"` // Worker调用协议
	ReportTime     int64  `json:"reportTime"`     // 上报时间

	IsNewWorker bool `json:"-"` // 是否是新的Worker

	AssignComputeCost int64    `json:"-"`             // 累计分配的计算复杂度
	AssignTaskCnt     int32    `json:"assignTaskCnt"` // 分配的任务数量,缓存这个字段标识是否有分配过,如果上次为0,也是新的Worker
	AddTaskKey        []string `json:"-"`             // 新增的任务列表
	DelTaskKey        []string `json:"-"`             // 删除的任务列表
}

// GetAddr 获取Worker地址
func (w *WorkerInfo) GetAddr() string {
	return fmt.Sprintf("%s:%d", w.Ip, w.Port)
}

// GetWorkerKey 获取Worker唯一标识
func (w *WorkerInfo) GetWorkerKey() string {
	return strings.Join([]string{w.Ip, fmt.Sprint(w.Port), fmt.Sprint(w.StartTime)}, consts.CommonFieldSeq)
}

// ToJsonString 获取Worker字符串
func (w *WorkerInfo) ToJsonString() string {
	data, _ := json.Marshal(w)
	return string(data)
}

func (w *WorkerInfo) ToLogString() string {
	return fmt.Sprintf("workerInfo:[%s], isNewWorker:[%v], assignComputeCost:[%d], "+
		"AssignTaskCnt:[%d], addTaskCnt:[%d], delTaskCnt:[%d] \naddTask:[%s]\ndelTask:[%s]",
		w.ToJsonString(), w.IsNewWorker, w.AssignComputeCost, w.AssignTaskCnt, len(w.AddTaskKey), len(w.DelTaskKey),
		strings.Join(w.AddTaskKey, ","), strings.Join(w.DelTaskKey, ","))
}
