package utils

import (
	"sync"
)

// Task 任务接口
type Task interface {
}

// TaskPool 任务池
type TaskPool struct {
	handle    func(task Task, worker TaskWorker)
	maxCount  int
	taskQueue chan Task
	wg        sync.WaitGroup
}

// TaskWorker 任务执行器
type TaskWorker struct {
	Arg any
}

// NewTaskPool 创建任务池
func NewTaskPool() *TaskPool {
	return &TaskPool{}
}

// Start 启动任务池
func (tp *TaskPool) Start(workers []TaskWorker, handle func(task Task, worker TaskWorker)) {
	if tp.taskQueue != nil {
		tp.Stop()
	}

	tp.handle = handle
	tp.taskQueue = make(chan Task)
	for i := 0; i < len(workers); i++ {
		go tp.process(workers[i])
	}
}

// AddTask 添加任务
func (tp *TaskPool) AddTask(task Task) {
	tp.wg.Add(1)
	tp.taskQueue <- task
}

// WaitFinish 等待任务完成
func (tp *TaskPool) WaitFinish() {
	tp.wg.Wait()
}

// Stop 停止任务池
func (tp *TaskPool) Stop() {
	tp.WaitFinish()
	close(tp.taskQueue)
	tp.taskQueue = nil
}

func (tp *TaskPool) process(worker TaskWorker) {
	for task := range tp.taskQueue {
		tp.handle(task, worker)
		tp.wg.Done()
	}
}
