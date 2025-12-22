// Package scheduler 存放业务逻辑对应的代码
package scheduler

import (
	"common/util/dislock"
	"context"
	"etrpc-go/log"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
	"scheduler/entity/model"
	"scheduler/repo/db"
	"strings"
	"sync"
	"time"
	"trpc.group/trpc-go/trpc-database/goredis/redlock"
	"trpc.group/trpc-go/trpc-go"
	tlog "trpc.group/trpc-go/trpc-go/log"
)

// ISchedulerLogic 任务调度接口,T为数据类型,R为接口请求参数类型
type ISchedulerLogic[T any, R any] interface {
	// RunTask 执行一次任务调度
	RunTask(wg *sync.WaitGroup)
	// PartitionData 划分数据的方案
	PartitionData(data []*model.TaskItem[T], workerMap map[string]*model.WorkerInfo, lastAssignResult map[string]string) error
	// ConvertToReq 将数据转化为请求参数
	ConvertToReq(addData []T, delData []string, fullPublish bool, verMark string) R
	// CallPublish 数据发布函数，将数据发送到worker节点，通常为一次网络调用
	CallPublish(ctx context.Context, worker *model.WorkerInfo, data R) (err error)
}

// DefaultSchedulerLogic 默认的调度器实现逻辑，包含部分默认的处理方案
type DefaultSchedulerLogic[T any, R any] struct {
	ISchedulerLogic[T, R]
	Dao     db.ISchedulerDao[T] // 和中间件交互的接口实现
	UnitCfg *model.TaskConfig   // 调度任务配置信息
}

var noAvailableWorkerErr = fmt.Errorf("no available worker exist")
var lessProcessCapErr = fmt.Errorf("total worker process cap less than total actual compute cost")

var tracer = otel.Tracer("")

// RunTask 加分布式锁发起一次任务调度
func (s *DefaultSchedulerLogic[T, R]) RunTask(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	// 增加TraceID信息，方便根据TraceID定位某次执行的所有日志
	ctx, span := tracer.Start(trpc.BackgroundContext(), "scheduler_task")
	defer span.End()
	tlog.WithContextFields(ctx, "traceID", span.SpanContext().TraceID().String())
	// 加分布式锁执行任务
	if err := dislock.DisLock(ctx, s.UnitCfg.RedisName, s.UnitCfg.LockKey, func() {
		var publishErr = s.schedulerTask(ctx)
		// 没执行下发动作，但是报错了, 通常是数据库/Redis异常,直接告警
		if publishErr != nil {
			log.AlarmContextf(ctx, "[%s]: scheduler fail, err: %v", s.UnitCfg.Name, publishErr)
			return
		} else {
			// 下发成功后持有锁等待5秒，等待下游处理完成，避免下游上报的数据版本和缓存的版本不一致
			time.Sleep(time.Second * 5)
		}
	}, redlock.WithKeyExpiration(time.Second*time.Duration(s.UnitCfg.LockKeyExpireSec))); err != nil {
		//log.InfoContextf(ctx, "[%s]: run publish task error: %v", s.UnitCfg.Name, err)
	}
}

// schedulerTask 执行任务调度逻辑
func (s *DefaultSchedulerLogic[T, R]) schedulerTask(ctx context.Context) error {
	// 1 判断数据版本标识是否出现变化
	// 1.1 读取数据库中最新的版本标识
	curVerStr, err := s.Dao.GetCurVersionStr(ctx)
	if err != nil {
		return errors.Wrapf(err, "get latest verion str from mysql fail")
	}
	// 1.1 从redis中获取上次下发使用的版本标识
	lastVerStr, err := s.Dao.GetLastVerStr(ctx)
	if err != nil {
		return errors.Wrapf(err, "get last verion str from redis fail")
	}
	verNoChanged := strings.EqualFold(curVerStr, lastVerStr)

	// 2 比对worker是否发生变化
	// 2.1、获取最新的Worker实例列表
	curWorkers, err := s.Dao.GetRegisterWorkerList(ctx)
	if err != nil {
		return errors.Wrapf(err, "get register worker list from redis fail")
	}
	// 2.2、从redis中获取上次下发使用的Worker实例列表
	lastWorkers, err := s.Dao.GetLastWorkerList(ctx)
	if err != nil {
		return errors.Wrapf(err, "get last worker list from redis fail")
	}
	workerNoChanged := !s.compareAndMarkWorkers(curWorkers, lastWorkers)
	//2.3 版本和worker都没有发生变化,则无需执行下发
	if verNoChanged && workerNoChanged {
		log.InfoContextf(ctx, "[%s]: data version and worker not changed", s.UnitCfg.Name)
		return nil
	}

	// 临时逻辑，如果是旧版本的服务,全部为全量下发
	if s.UnitCfg.OldVer {
		for _, worker := range curWorkers {
			worker.IsNewWorker = true
		}
	}
	log.InfoContextf(ctx, "[%s]: data version status:[%v], worker info status:[%v]",
		s.UnitCfg.Name, !verNoChanged, !workerNoChanged)
	log.InfoContextf(ctx, "[%s]: last worker: [%s]", s.UnitCfg.Name, strings.Join(lo.Map(lastWorkers,
		func(item *model.WorkerInfo, index int) string { return item.ToJsonString() }), "\n"))
	log.InfoContextf(ctx, "[%s]: cur worker: [%s]", s.UnitCfg.Name, strings.Join(lo.Map(curWorkers,
		func(item *model.WorkerInfo, index int) string { return item.ToJsonString() }), "\n"))

	// 3、获取待下发的数据以及上次下发的分配结果
	// 3.1 从DB读取待下发的最新数据
	publishData, err := s.Dao.GetPublishData(ctx, verNoChanged)
	if err != nil {
		return errors.Wrapf(err, "get publish data from mysql fail")
	}
	// 3.2 从Redis获取上次下发分配方案
	lastAssignResult, err := s.Dao.GetLastAssignResult(ctx)
	if err != nil {
		return errors.Wrapf(err, "get last assign result from redis fail")
	}
	// 3.3 执行下发动作
	newAssignResult, assignWorkers, err := s.doPublish(ctx, publishData, curWorkers, lastAssignResult)
	if err != nil {
		return err
	}

	// 4、缓存本次下发的一些信息
	// 4.1 本次下发的数据版本
	if err := s.Dao.SetLastVerStr(ctx, curVerStr); err != nil {
		log.AlarmContextf(ctx, "cache last verion str to redis err: %v", err)
	}
	// 4.2、本次分配的Workers列表
	if err := s.Dao.SetLastWorkerList(ctx, assignWorkers); err != nil {
		log.AlarmContextf(ctx, "cache last worker list to redis err: %v", err)
	}
	// 4.2、本次分配的分配结果
	if err := s.Dao.SetLastAssignResult(ctx, newAssignResult); err != nil {
		log.AlarmContextf(ctx, "cache last assign result to redis err: %v", err)
	}
	return nil
}

func (s *DefaultSchedulerLogic[T, R]) doPublish(ctx context.Context, publishData []*model.TaskItem[T],
	workers []*model.WorkerInfo, lastAssignResult map[string]string) (map[string]string, []*model.WorkerInfo, error) {
	// 1 下发前数据准备,将worker转化为map,方便后续使用
	workerMap := lo.SliceToMap(workers, func(item *model.WorkerInfo) (string, *model.WorkerInfo) {
		return item.GetWorkerKey(), item
	})
	// 2 重试执行数据分配&数据下发
	newAssignResult := make(map[string]string)
	if err := retry.Do(func() error {
		if len(workerMap) == 0 {
			return noAvailableWorkerErr
		}
		log.InfoContextf(ctx, "[%s]: begin publish data...", s.UnitCfg.Name)
		// 执行数据分配逻辑
		if err := s.PartitionData(publishData, workerMap, lastAssignResult); err != nil {
			return err
		}
		// 将分配好的数据转化为请求参数
		workerReqMap, assignResult := s.buildReq(publishData, workerMap, lastAssignResult)
		for worker, req := range workerReqMap {
			// 最多重试三次
			err := retry.Do(func() error {
				return s.CallPublish(ctx, worker, req)
			}, retry.Attempts(3), retry.Delay(time.Millisecond*500), retry.RetryIf(func(err error) bool {
				return err != nil
			}))
			if err != nil {
				// 三次都失败了,移除这个worker,尝试重新分配重新下发
				delete(workerMap, worker.GetWorkerKey())
				log.InfoContextf(ctx, "[%s]: publish data to worker %s fail, err: %v, begin remove cur worker and retry",
					s.UnitCfg.Name, worker.GetWorkerKey(), err)
				return errors.Wrapf(err, "publish data to worker %s fail", worker.GetWorkerKey())
			}
			log.InfoContextf(ctx, "[%s]: success publish data to worker %s, publish info: %s",
				s.UnitCfg.Name, worker.GetWorkerKey(), worker.ToLogString())
		}
		newAssignResult = assignResult
		return nil
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil && !errors.Is(err, noAvailableWorkerErr) && !errors.Is(err, lessProcessCapErr)
	})); err != nil {
		return nil, nil, fmt.Errorf("publish data fail after 3 retry, err: %v,  more info please check log", err)
	}
	return newAssignResult, lo.Values(workerMap), nil
}

func (s *DefaultSchedulerLogic[T, R]) buildReq(data []*model.TaskItem[T], workerMap map[string]*model.WorkerInfo,
	lastAssignResult map[string]string) (map[*model.WorkerInfo]R, map[string]string) {
	// worker组装的请求参数,Key: Worker, Value:请求体
	workerReq := make(map[*model.WorkerInfo]R, len(workerMap))
	// 新的任务分配结果，TaskKey: 任务标识, Value: Worker标识
	newAssignResult := make(map[string]string)
	// worker新分配的任务信息,TaskKey: Worker标识, Value: 任务Map
	workerNewAssignMap := make(map[string]map[string]*model.TaskItem[T])
	for _, task := range data {
		assignMap, ok := workerNewAssignMap[task.AssignWorker]
		if !ok {
			assignMap = make(map[string]*model.TaskItem[T])
			workerNewAssignMap[task.AssignWorker] = assignMap
		}
		assignMap[task.TaskKey] = task
		newAssignResult[task.TaskKey] = task.AssignWorker
	}
	// worker旧的分配列表
	workerOldAssignMap := make(map[string]map[string]struct{})
	for taskKey, lastWorkerStr := range lastAssignResult {
		assignMap, ok := workerOldAssignMap[lastWorkerStr]
		if !ok {
			assignMap = make(map[string]struct{})
			workerOldAssignMap[lastWorkerStr] = assignMap
		}
		assignMap[taskKey] = struct{}{}
	}
	// 每个worker找出新增和删除的数据
	verMarkStr := time.Now().Format(time.DateTime)
	for workerStr, worker := range workerMap {
		newAssignMap := workerNewAssignMap[workerStr]
		oldAssignMap := workerOldAssignMap[workerStr]
		// 查找出新增的数据
		addTask := make([]T, 0, len(newAssignMap))
		addTaskKey := make([]string, 0, len(newAssignMap))
		delTaskKey := make([]string, 0, len(newAssignMap))
		// 新worker或者没分配过,全部加到新增列表
		if worker.IsNewWorker || len(oldAssignMap) == 0 {
			for taskKey, task := range newAssignMap {
				addTask = append(addTask, task.TaskData)
				addTaskKey = append(addTaskKey, taskKey)
			}
		} else {
			// 找出需要新增的任务
			for taskKey, task := range newAssignMap {
				if _, ok := oldAssignMap[taskKey]; !ok {
					addTask = append(addTask, task.TaskData)
					addTaskKey = append(addTaskKey, taskKey)
				}
			}
			// 找出需要移除的任务
			for taskKey := range oldAssignMap {
				if _, ok := newAssignMap[taskKey]; !ok {
					delTaskKey = append(delTaskKey, taskKey)
				}
			}
		}
		// 分配数量用于标记下次下发时该Worker是否为新Worker
		worker.AssignTaskCnt = int32(len(newAssignMap))
		// 数据变化了才需要下发
		if len(addTask) > 0 || len(delTaskKey) > 0 {
			// 保存一些分配的信息,用于日志输出
			worker.AddTaskKey = addTaskKey
			worker.DelTaskKey = delTaskKey
			worker.TaskVerMark = verMarkStr
			// 生成请求参数
			req := s.ConvertToReq(addTask, delTaskKey, worker.IsNewWorker, verMarkStr)
			workerReq[worker] = req
		} else {
			log.Infof("[%s]: task not changed, ignore publish, worker info:[%s]", s.UnitCfg.Name, worker.ToLogString())
		}
	}
	return workerReq, newAssignResult
}

// DefaultPartitionData 默认的数据划分算法
func (s *DefaultSchedulerLogic[T, R]) DefaultPartitionData(tasks []*model.TaskItem[T], workerMap map[string]*model.WorkerInfo,
	resetAssignComputeCost bool, lastAssignMap map[string]string) error {
	if len(workerMap) == 0 {
		return noAvailableWorkerErr
	}
	// 初始化每个worker已分配的计算复杂度为0，计算worker总的处理能力
	var totalLeftWorkerCap int64
	var totalAssignCap int64
	for _, worker := range workerMap {
		if resetAssignComputeCost {
			worker.AssignComputeCost = 0
		}
		totalAssignCap += worker.AssignComputeCost
		totalLeftWorkerCap += worker.MaxProcessCap - worker.AssignComputeCost
	}
	// 计算所有任务的累计计算复杂度
	totalTaskComputeCost := lo.SumBy(tasks, func(item *model.TaskItem[T]) int64 {
		return item.ComputeCost
	})
	// 总worker最大处理能力 < 数据的计算复杂度总和
	if totalLeftWorkerCap < totalTaskComputeCost {
		return errors.Wrapf(lessProcessCapErr, "totalAssignCap:[%d], totalLeftWorkerCap:[%d], totalTaskComputeCost:[%d]",
			totalAssignCap, totalLeftWorkerCap, totalTaskComputeCost)
	}
	// 记录需要重新分配的数据
	needResignTasks := make([]*model.TaskItem[T], 0)
	// 依次判断每条数据上次分配的Worker是否存活，以及存活情况下是否能够保持分配
	for _, task := range tasks {
		// 获取这个任务上一次分配的worker
		if oldWorkerStr, ok := lastAssignMap[task.TaskKey]; ok {
			// Worker还存活
			if newWorker, ok := workerMap[oldWorkerStr]; ok {
				newComputeCost := newWorker.AssignComputeCost + task.ComputeCost
				// 新分配的计算复杂度必须小于Worker最大处理能力
				if newComputeCost <= newWorker.MaxProcessCap {
					newWorker.AssignComputeCost = newComputeCost
					task.AssignWorker = oldWorkerStr
					continue
				}
			}
		}
		needResignTasks = append(needResignTasks, task)
	}
	// 过滤出新启动的Worker和旧的Worker
	oldWorkers := make([]*model.WorkerInfo, 0) // 旧的已存在的Worker
	newWorkers := make([]*model.WorkerInfo, 0) // 新启动的Worker
	for _, worker := range workerMap {
		if worker.IsNewWorker {
			newWorkers = append(newWorkers, worker)
		} else {
			oldWorkers = append(oldWorkers, worker)
		}
	}
	// 分配剩余的任务，每个任务找到剩余处理能力最大的Worker进行分配
	for _, task := range needResignTasks {
		var maxWorker *model.WorkerInfo
		var maxWorkerCap int64 = 0
		// 优先分配到新启动的Worker上
		for _, worker := range newWorkers {
			leftProcessCap := worker.MaxProcessCap - worker.AssignComputeCost
			if leftProcessCap >= task.ComputeCost && leftProcessCap > maxWorkerCap {
				maxWorkerCap = leftProcessCap
				maxWorker = worker
			}
		}
		// 没有新启动的或者新启动的分配满了,再分配到旧的上
		if maxWorker == nil {
			for _, worker := range oldWorkers {
				leftProcessCap := worker.MaxProcessCap - worker.AssignComputeCost
				if leftProcessCap >= task.ComputeCost && leftProcessCap > maxWorkerCap {
					maxWorkerCap = worker.MaxProcessCap - worker.AssignComputeCost
					maxWorker = worker
				}
			}
		}
		if maxWorker == nil {
			// 无法找到满足条件的最小Worker,说明超过所有Worker的最大处理能力
			return errors.Wrapf(lessProcessCapErr, "record:[%s] compute cost:[%d], can not assign",
				task.TaskKey, task.ComputeCost)
		}
		// 分配到处理能力最大的Worker上
		maxWorker.AssignComputeCost += task.ComputeCost
		task.AssignWorker = maxWorker.GetWorkerKey()
	}
	return nil
}

// compareAndMarkWorkers 比较Worker是否变化,以及标记哪些worker应该识别为新的worker
func (s *DefaultSchedulerLogic[T, R]) compareAndMarkWorkers(curWorkers []*model.WorkerInfo, lastWorkers []*model.WorkerInfo) bool {
	// 先将lastWorkers转化为Map,方便查找
	lastWorkerMap := lo.SliceToMap(lastWorkers, func(item *model.WorkerInfo) (string, *model.WorkerInfo) {
		return item.GetWorkerKey(), item
	})
	// worker数量是否变化
	changed := len(curWorkers) != len(lastWorkers)
	for _, curWorker := range curWorkers {
		// 前后worker同时存在,比对部分关键信息是否变化
		if lastWorker, ok := lastWorkerMap[curWorker.GetWorkerKey()]; ok {
			// 最大处理能力变化
			if curWorker.MaxProcessCap != lastWorker.MaxProcessCap {
				changed = true
			}
			// 上次没分配过也标记为新的Worker,但changed不一定需要重新标记
			if lastWorker.AssignTaskCnt == 0 {
				curWorker.IsNewWorker = true
			}
			// 数据版本不一致，这种也标记为新的Worker，需要重新全量下发数据
			// 旧版本没有上报这个版本信息,所以无需判断
			if !s.UnitCfg.OldVer && curWorker.TaskVerMark != lastWorker.TaskVerMark {
				changed = true
				curWorker.IsNewWorker = true
			}
		} else {
			// 标记为新的Worker，需要重新全量下发数据
			changed = true
			curWorker.IsNewWorker = true
		}
	}
	return changed
}
