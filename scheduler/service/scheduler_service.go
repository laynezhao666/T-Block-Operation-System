package service

import (
	tconfig "etrpc-go/config"
	"etrpc-go/log"
	"scheduler/entity/dbmodel"
	"scheduler/entity/model"
	"scheduler/logic/scheduler"
	"sync"
	"time"
	"trpc.group/trpc-go/trpc-go"
	"trpcprotocol/agent"
	"trpcprotocol/alarm-compute"
	"trpcprotocol/data-compute"
)

var schedulerService = newSchedulerService()

func init() {
	tconfig.Register("scheduler-config", schedulerService.GetConfig(), tconfig.WithHotUpdate(true),
		tconfig.WithUpdateFunc(func(oldVal, newVal any) {
			// 启动时配置加载优先于trpc插件的初始化,直接调用可能因为缺少插件导致启动失败
			if schedulerService.IsReady() {
				schedulerService.RefreshTask()
			}
		}))
}

// GetSchedulerService 获取调度Service
func GetSchedulerService() ISchedulerService {
	return schedulerService
}

// ISchedulerService 调度服务接口
type ISchedulerService interface {
	// GetConfig 获取调度配置
	GetConfig() *model.AllTaskConfig
	// RefreshTask 刷新所有任务
	RefreshTask()
	// WaitTaskDone 等待所有任务执行完成
	WaitTaskDone()
	// CancelTask 取消所有任务

	CancelTask()
	// IsReady 调度任务是否就绪
	IsReady() bool
}

type schedulerServiceImpl struct {
	cfg     *model.AllTaskConfig      // 当前配置信息
	taskMap map[string]*schedulerTask // 当前任务map
	ready   bool                      // 是否就绪
	wg      *sync.WaitGroup           // 等待组,用于等待所有配置任务执行完成

}

func newSchedulerService() ISchedulerService {
	return &schedulerServiceImpl{
		cfg:     &model.AllTaskConfig{},
		wg:      &sync.WaitGroup{},
		taskMap: make(map[string]*schedulerTask),
	}
}

func (s *schedulerServiceImpl) GetConfig() *model.AllTaskConfig {
	return s.cfg
}

func (s *schedulerServiceImpl) RefreshTask() {
	// 根据配置解析出需要执行的任务
	validSchedulerCfg := make(map[string]*model.TaskConfig)
	for _, schedulerUnit := range s.cfg.Scheduler {
		if err := schedulerUnit.BuildDefaultAndValid(); err != nil {
			log.AlarmContextf(trpc.BackgroundContext(), "ignore bad scheduler unit cfg, err: %s", err.Error())
		} else {
			validSchedulerCfg[schedulerUnit.CalcUniqueKey()] = schedulerUnit
		}
	}
	// 移除掉需要删除的任务
	for unitKey, task := range s.taskMap {
		if _, ok := validSchedulerCfg[unitKey]; !ok {
			task.Stop()
			delete(s.taskMap, unitKey)
		}
	}
	// 需要新增的任务
	for unitKey, unitCfg := range validSchedulerCfg {
		if _, ok := s.taskMap[unitKey]; !ok && !unitCfg.Disable {
			task := newSchedulerTask(unitCfg, s.wg)
			s.taskMap[unitKey] = task
			task.Start()
		}
	}
	s.ready = true
	log.Infof("load scheduler task success, total valid cnt: %d", len(s.taskMap))
}

// WaitTaskDone 等待所有任务执行完成
func (s *schedulerServiceImpl) WaitTaskDone() {
	s.wg.Wait()
}

// IsReady 是否准备完毕
func (s *schedulerServiceImpl) IsReady() bool {
	return s.ready
}

// CancelTask 取消任务
func (s *schedulerServiceImpl) CancelTask() {
	for _, task := range s.taskMap {
		task.Stop()
	}
}

// schedulerTask 调度任务结构
type schedulerTask struct {
	running   bool              // 是否在运行
	unitCfg   *model.TaskConfig // 调度任务的配置
	interVal  time.Duration     // 调度间隔
	ticker    *time.Ticker      // 定时器
	waitGroup *sync.WaitGroup   // 等待组
	taskObj   any
	cancel    chan struct{}
}

func newSchedulerTask(unitCfg *model.TaskConfig, wg *sync.WaitGroup) *schedulerTask {
	interVal := time.Second * time.Duration(unitCfg.IntervalSec)
	return &schedulerTask{
		running:   false,
		unitCfg:   unitCfg,
		interVal:  interVal,
		ticker:    time.NewTicker(interVal),
		cancel:    make(chan struct{}),
		waitGroup: wg,
	}
}

// Start 启动调度任务
func (s *schedulerTask) Start() {
	log.Infof("scheduler task [%s] beging running, cfg:%v", s.unitCfg.Name, s.unitCfg)
	// 启动时先调度一次
	switch s.unitCfg.Type {
	case model.TaskTypeAlarm:
		alarmTask := scheduler.NewAlarmStrategyLogic(s.unitCfg)
		s.taskObj = alarmTask
		go alarmTask.RunTask(s.waitGroup)
	case model.TaskTypeCollector:
		collectorTask := scheduler.NewCollectorDeviceLogic(s.unitCfg)
		s.taskObj = collectorTask
		go collectorTask.RunTask(s.waitGroup)
	case model.TaskTypePoint:
		pointTask := scheduler.NewDevicePointLogic(s.unitCfg)
		s.taskObj = pointTask
		go pointTask.RunTask(s.waitGroup)
	default:
		// 正常来说不存在这种情况
		return
	}
	// 按频率定时调度
	s.running = true
	s.waitGroup.Add(1)
	go func(st *schedulerTask) {
		defer st.waitGroup.Done()
		for st.running {
			select {
			case <-st.cancel:
				st.running = false
				s.ticker.Stop()
			case <-st.ticker.C:
				switch st.unitCfg.Type {
				case model.TaskTypeAlarm:
					go st.taskObj.(scheduler.ISchedulerLogic[*dbmodel.AlarmStrategy,
						*alarm_compute.ReqStrategyRecv]).RunTask(st.waitGroup)
				case model.TaskTypeCollector:
					go st.taskObj.(scheduler.ISchedulerLogic[*dbmodel.CollectorDevice,
						*agent.RecvCollectTaskReq]).RunTask(st.waitGroup)
				case model.TaskTypePoint:
					go st.taskObj.(scheduler.ISchedulerLogic[*dbmodel.DevicePoint,
						*data_compute.ReqReceiveTask]).RunTask(st.waitGroup)
				}
				// 成功后重置一次下发时间
				st.ticker.Reset(st.interVal)
			}
		}
		log.Infof("scheduler task [%s] stoped", st.unitCfg.Name)
	}(s)
}

func (s *schedulerTask) Stop() {
	s.running = false
	s.cancel <- struct{}{}
	if s.ticker != nil {
		s.ticker.Stop()
	}
}
