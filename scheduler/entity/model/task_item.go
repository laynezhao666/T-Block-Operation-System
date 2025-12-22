// Package model 定义任务项结构
package model

// TaskItem 任务项
type TaskItem[T any] struct {
	TaskData    T      // 任务项具体数据
	TaskKey     string // 任务项唯一标识
	ComputeCost int64  // 任务项计算复杂度

	AssignWorker string // 任务项分配的Worker
}
