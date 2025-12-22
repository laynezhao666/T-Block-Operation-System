// Package task 任务相关
package task

import (
	"sync"
)

// TaskManager 结构体用于管理任务
type TaskManager struct {
	tasks []string
	sync.RWMutex
}

var (
	instance *TaskManager
	once     sync.Once
)

// GetInstance 获取 TaskManager 的单例实例
func GetInstance() *TaskManager {
	once.Do(func() {
		instance = &TaskManager{
			tasks: make([]string, 0),
		}
	})
	return instance
}

// UpdateTasks 全量更新任务
func (tm *TaskManager) UpdateTasks(newTasks []string) {
	tm.Lock()
	defer tm.Unlock()
	tm.tasks = newTasks

}

// GetTasks 获取当前任务
func (tm *TaskManager) GetTasks() []string {
	tm.RLock()
	defer tm.RUnlock()
	return tm.tasks
}
