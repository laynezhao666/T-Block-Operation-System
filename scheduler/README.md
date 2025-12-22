# Scheduler 调度中心服务

Scheduler 是 TBOS 系统的中心调度服务，负责将配置数据（告警策略、采集设备、标准测点）按负载均衡算法分配到下游 Worker 节点，实现任务的分布式调度和管理。

## 模块介绍

Scheduler 服务的主要职责：

1. **Worker 管理**：接收下游 Worker（alarm-compute、agent、data-compute）的心跳注册，维护 Worker 状态
2. **任务分配**：根据数据版本变化和 Worker 变化，触发任务重新分配
3. **增量下发**：支持增量和全量两种下发模式，最小化网络传输和下游处理开销
4. **负载均衡**：基于计算复杂度的任务分配算法，实现 Worker 间负载均衡

### 数据流向

```
┌─────────────────────────────────────────────────────────────────────┐
│                           MySQL 数据库                               │
│  (t_mozu_info、告警策略表、采集设备表、标准测点表)                     │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                          定时读取配置数据
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                           Scheduler                                  │
│                           调度中心                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │
│  │ alarm调度任务 │  │collector调度 │  │  point调度   │               │
│  │   (告警策略)  │  │  (采集设备)  │  │  (标准测点)  │               │
│  └──────────────┘  └──────────────┘  └──────────────┘               │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              │                  │                  │
              ▼                  ▼                  ▼
     ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
     │alarm-compute│    │    agent    │    │data-compute │
     │  告警计算    │    │  边缘采集器  │    │  测点计算    │
     └─────────────┘    └─────────────┘    └─────────────┘
              │                  │                  │
              └──────────────────┼──────────────────┘
                                 │
                          心跳注册到Redis
                                 │
                                 ▼
                    ┌─────────────────────────┐
                    │        Redis            │
                    │ (Worker注册、分配结果缓存) │
                    └─────────────────────────┘
```

## 核心能力

### 1. 调度任务泛型框架

Scheduler 使用 Go 泛型实现了通用的调度框架，支持不同类型任务的统一调度：

```go
// logic/scheduler/scheduler_logic.go
// ISchedulerLogic 任务调度接口,T为数据类型,R为接口请求参数类型
type ISchedulerLogic[T any, R any] interface {
    // RunTask 执行一次任务调度
    RunTask(wg *sync.WaitGroup)
    // PartitionData 划分数据的方案
    PartitionData(data []*model.TaskItem[T], workerMap map[string]*model.WorkerInfo, lastAssignResult map[string]string) error
    // ConvertToReq 将数据转化为请求参数
    ConvertToReq(addData []T, delData []string, fullPublish bool, verMark string) R
    // CallPublish 数据发布函数，将数据发送到worker节点
    CallPublish(ctx context.Context, worker *model.WorkerInfo, data R) (err error)
}
```

**三种调度任务实现**：
- `alarmStrategyLogic`：告警策略调度，下发到 alarm-compute
- `collectorDeviceLogic`：采集设备调度，下发到 agent
- `devicePointLogic`：标准测点调度，下发到 data-compute

### 2. 调度触发机制

调度任务通过定时器定期执行，默认每30秒触发一次：

```go
// service/scheduler_service.go
func (s *schedulerTask) Start() {
    // 启动时先调度一次
    switch s.unitCfg.Type {
    case model.TaskTypeAlarm:
        alarmTask := scheduler.NewAlarmStrategyLogic(s.unitCfg)
        go alarmTask.RunTask(s.waitGroup)
    case model.TaskTypeCollector:
        collectorTask := scheduler.NewCollectorDeviceLogic(s.unitCfg)
        go collectorTask.RunTask(s.waitGroup)
    case model.TaskTypePoint:
        pointTask := scheduler.NewDevicePointLogic(s.unitCfg)
        go pointTask.RunTask(s.waitGroup)
    }
    
    // 按频率定时调度
    go func(st *schedulerTask) {
        for st.running {
            select {
            case <-st.ticker.C:
                // 定时触发调度
            }
        }
    }(s)
}
```

### 3. 分布式锁保证单点执行

使用 Redis 分布式锁确保同一时刻只有一个 Scheduler 实例执行调度：

```go
// logic/scheduler/scheduler_logic.go
func (s *DefaultSchedulerLogic[T, R]) RunTask(wg *sync.WaitGroup) {
    // 加分布式锁执行任务
    if err := dislock.DisLock(ctx, s.UnitCfg.RedisName, s.UnitCfg.LockKey, func() {
        var publishErr = s.schedulerTask(ctx)
        if publishErr != nil {
            log.AlarmContextf(ctx, "[%s]: scheduler fail, err: %v", s.UnitCfg.Name, publishErr)
            return
        }
        // 下发成功后持有锁等待5秒，避免下游上报的数据版本和缓存的版本不一致
        time.Sleep(time.Second * 5)
    }, redlock.WithKeyExpiration(time.Second*time.Duration(s.UnitCfg.LockKeyExpireSec))); err != nil {
        // 未获取到锁，跳过执行
    }
}
```

### 4. 版本变化检测

调度器通过比对版本标识判断是否需要重新下发：

```go
// logic/scheduler/scheduler_logic.go
func (s *DefaultSchedulerLogic[T, R]) schedulerTask(ctx context.Context) error {
    // 1. 判断数据版本标识是否出现变化
    // 1.1 读取数据库中最新的版本标识（从t_mozu_info表获取）
    curVerStr, err := s.Dao.GetCurVersionStr(ctx)
    // 1.2 从redis中获取上次下发使用的版本标识
    lastVerStr, err := s.Dao.GetLastVerStr(ctx)
    verNoChanged := strings.EqualFold(curVerStr, lastVerStr)

    // 2. 比对worker是否发生变化
    curWorkers, err := s.Dao.GetRegisterWorkerList(ctx)
    lastWorkers, err := s.Dao.GetLastWorkerList(ctx)
    workerNoChanged := !s.compareAndMarkWorkers(curWorkers, lastWorkers)
    
    // 版本和worker都没有发生变化,则无需执行下发
    if verNoChanged && workerNoChanged {
        return nil
    }
    // ... 执行下发逻辑
}
```

**触发下发的条件**：
- 数据版本变化（`t_mozu_info` 表的 `publish_version` 或 `alarm_version` 字段变化）
- Worker 列表变化（新增、移除、最大处理能力变化）
- Worker 任务版本标识不一致

### 5. Worker 变化检测与标记

```go
// logic/scheduler/scheduler_logic.go
func (s *DefaultSchedulerLogic[T, R]) compareAndMarkWorkers(curWorkers []*model.WorkerInfo, lastWorkers []*model.WorkerInfo) bool {
    changed := len(curWorkers) != len(lastWorkers)
    for _, curWorker := range curWorkers {
        if lastWorker, ok := lastWorkerMap[curWorker.GetWorkerKey()]; ok {
            // 最大处理能力变化
            if curWorker.MaxProcessCap != lastWorker.MaxProcessCap {
                changed = true
            }
            // 上次没分配过也标记为新的Worker
            if lastWorker.AssignTaskCnt == 0 {
                curWorker.IsNewWorker = true
            }
            // 数据版本不一致，标记为新的Worker，需要全量下发
            if curWorker.TaskVerMark != lastWorker.TaskVerMark {
                changed = true
                curWorker.IsNewWorker = true
            }
        } else {
            // 新Worker，需要全量下发
            changed = true
            curWorker.IsNewWorker = true
        }
    }
    return changed
}
```

### 6. 基于计算复杂度的负载均衡算法

```go
// logic/scheduler/scheduler_logic.go
func (s *DefaultSchedulerLogic[T, R]) DefaultPartitionData(tasks []*model.TaskItem[T], workerMap map[string]*model.WorkerInfo,
    resetAssignComputeCost bool, lastAssignMap map[string]string) error {
    
    // 1. 计算worker总的处理能力
    var totalLeftWorkerCap int64
    for _, worker := range workerMap {
        totalLeftWorkerCap += worker.MaxProcessCap - worker.AssignComputeCost
    }
    
    // 2. 计算所有任务的累计计算复杂度
    totalTaskComputeCost := lo.SumBy(tasks, func(item *model.TaskItem[T]) int64 {
        return item.ComputeCost
    })
    
    // 3. 总处理能力不足时报错
    if totalLeftWorkerCap < totalTaskComputeCost {
        return lessProcessCapErr
    }
    
    // 4. 优先保持原有分配（减少任务迁移）
    for _, task := range tasks {
        if oldWorkerStr, ok := lastAssignMap[task.TaskKey]; ok {
            if newWorker, ok := workerMap[oldWorkerStr]; ok {
                if newWorker.AssignComputeCost + task.ComputeCost <= newWorker.MaxProcessCap {
                    newWorker.AssignComputeCost += task.ComputeCost
                    task.AssignWorker = oldWorkerStr
                    continue
                }
            }
        }
        needResignTasks = append(needResignTasks, task)
    }
    
    // 5. 剩余任务优先分配到新启动的Worker
    for _, task := range needResignTasks {
        // 优先分配到新启动的Worker
        for _, worker := range newWorkers {
            leftProcessCap := worker.MaxProcessCap - worker.AssignComputeCost
            if leftProcessCap >= task.ComputeCost && leftProcessCap > maxWorkerCap {
                maxWorker = worker
            }
        }
        // 没有新Worker则分配到旧Worker
        if maxWorker == nil {
            for _, worker := range oldWorkers { ... }
        }
        maxWorker.AssignComputeCost += task.ComputeCost
        task.AssignWorker = maxWorker.GetWorkerKey()
    }
    return nil
}
```

**算法特点**：
- 优先保持原有分配，减少任务迁移
- 新任务优先分配到新启动的 Worker
- 每个任务分配给剩余处理能力最大的 Worker
- 支持按计算复杂度（`ComputeCost`）进行加权分配

### 7. 增量下发机制

```go
// logic/scheduler/scheduler_logic.go
func (s *DefaultSchedulerLogic[T, R]) buildReq(data []*model.TaskItem[T], workerMap map[string]*model.WorkerInfo,
    lastAssignResult map[string]string) (map[*model.WorkerInfo]R, map[string]string) {
    
    for workerStr, worker := range workerMap {
        newAssignMap := workerNewAssignMap[workerStr]  // 本次分配
        oldAssignMap := workerOldAssignMap[workerStr]  // 上次分配
        
        // 新Worker或从未分配过，全部加到新增列表
        if worker.IsNewWorker || len(oldAssignMap) == 0 {
            for taskKey, task := range newAssignMap {
                addTask = append(addTask, task.TaskData)
            }
        } else {
            // 增量模式：找出需要新增和删除的任务
            for taskKey, task := range newAssignMap {
                if _, ok := oldAssignMap[taskKey]; !ok {
                    addTask = append(addTask, task.TaskData)
                }
            }
            for taskKey := range oldAssignMap {
                if _, ok := newAssignMap[taskKey]; !ok {
                    delTaskKey = append(delTaskKey, taskKey)
                }
            }
        }
        
        // 数据变化了才需要下发
        if len(addTask) > 0 || len(delTaskKey) > 0 {
            req := s.ConvertToReq(addTask, delTaskKey, worker.IsNewWorker, verMarkStr)
            workerReq[worker] = req
        }
    }
}
```

**下发模式**：
- **全量下发**（`PublishType=1/UPDATEALL`）：清空Worker现有任务，加载全部新任务
- **增量下发**（`PublishType=0/INCREMENT`）：仅下发变更的任务（新增+删除）

### 8. Worker 心跳注册

Worker 通过 `Heartbeat` 接口注册到 Scheduler：

```go
// logic/register/api.go
func (r *registerApiImpl) Heartbeat(ctx context.Context, req *scheduler.WorkerInfo) error {
    workerInfo := &model.WorkerInfo{
        Ip:             req.Ip,
        Port:           req.Port,
        StartTime:      req.StartTime,
        MaxProcessCap:  int64(req.MaxProcessCap),
        TaskVerMark:    req.TaskVerMark,
        WorkerProtocol: strings.ToLower(req.WorkerProtocol.String()),
        ReportTime:     time.Now().Unix(),
    }
    
    redisKey := strings.Join([]string{model.DefaultRegisterWorkKey, 
        strings.ToLower(req.WorkerType.String()), req.WorkerSet}, consts.RedisJoinFieldSep)
    
    if req.WorkerStatus == scheduler.WorkerInfo_SHUTDOWN {
        // 取消注册
        return r.cache.HDel(ctx, redisKey, workerInfo.GetWorkerKey()).Err()
    } else {
        // 执行注册
        return r.cache.HSet(ctx, redisKey, workerInfo.GetWorkerKey(), workerInfo.ToJsonString()).Err()
    }
}
```

**Worker 过期处理**：
```go
// repo/cache/redis_util.go
// worker每5s上报一次心跳,如果3次未收到心跳(17秒),则认为worker挂掉
expireTs := time.Now().Add(-time.Second * 17).Unix()
if worker.ReportTime <= expireTs {
    invalidWorkers = append(invalidWorkers, workerKey)
}
```

### 9. 告警策略特殊分配逻辑

告警策略按 `ridType` 分组后再进行负载均衡：

```go
// logic/scheduler/alarm_strategy_logic.go
func (a *alarmStrategyLogic) PartitionData(data []*model.TaskItem[*dbmodel.AlarmStrategy], ...) error {
    // 先按ridType分组
    ridTypeDataMap := lo.GroupBy(data, func(item *model.TaskItem[*dbmodel.AlarmStrategy]) int64 {
        return item.TaskData.RidType
    })
    // 针对每一种ridType分别进行均分
    for _, strategies := range ridTypeDataMap {
        if err := a.DefaultPartitionData(strategies, workerMap, idx == 0, lastAllocateMap); err != nil {
            return err
        }
    }
    return nil
}
```

### 10. 采集设备计算复杂度计算

采集设备的计算复杂度 = 采集测点数 + 标准测点数：

```go
// repo/db/collector_device_dao.go
func (c *collectorDeviceDao) GetPublishData(ctx context.Context, verNoChanged bool) (...) {
    // 1. 查询所有采集设备
    // 2. 计算每个采集器下面的采集测点数
    collectorPointCntSql := c.Db.Table("t_collector_device c " +
        "left join t_collector_template_point p on c.template_name = p.template_name").
        Select("c.parent_device_number as device_number, count(1) as cnt").
        Group("c.parent_device_number")
    
    // 3. 计算每个采集器下面的标准测点数
    stdPointCntSql := c.Db.Table("t_device_point").
        Select("belong_collector as device_number, count(1) as cnt").
        Group("belong_collector")
    
    // 4. 计算量 = 采集测点数 + 标准测点数
    data := lo.Map(res, func(item *dbmodel.CollectorDevice, index int) *model.TaskItem[*dbmodel.CollectorDevice] {
        return &model.TaskItem[*dbmodel.CollectorDevice]{
            ComputeCost: collectorCntMap[item.DeviceNumber] + stdPointCntMap[item.DeviceNumber],
        }
    })
}
```

### 11. 配置热更新

调度配置支持热更新，配置变更时自动刷新任务：

```go
// service/scheduler_service.go
func init() {
    tconfig.Register("scheduler-config", schedulerService.GetConfig(), tconfig.WithHotUpdate(true),
        tconfig.WithUpdateFunc(func(oldVal, newVal any) {
            if schedulerService.IsReady() {
                schedulerService.RefreshTask()  // 配置更新后刷新任务
            }
        }))
}

func (s *schedulerServiceImpl) RefreshTask() {
    // 解析有效配置
    validSchedulerCfg := make(map[string]*model.TaskConfig)
    for _, schedulerUnit := range s.cfg.Scheduler {
        schedulerUnit.BuildDefaultAndValid()
        validSchedulerCfg[schedulerUnit.CalcUniqueKey()] = schedulerUnit
    }
    
    // 移除需要删除的任务
    for unitKey, task := range s.taskMap {
        if _, ok := validSchedulerCfg[unitKey]; !ok {
            task.Stop()
            delete(s.taskMap, unitKey)
        }
    }
    
    // 新增任务
    for unitKey, unitCfg := range validSchedulerCfg {
        if _, ok := s.taskMap[unitKey]; !ok && !unitCfg.Disable {
            task := newSchedulerTask(unitCfg, s.wg)
            s.taskMap[unitKey] = task
            task.Start()
        }
    }
}
```

### 12. 优雅关闭

服务停止时取消所有调度任务并等待执行完成：

```go
// main.go
func main() {
    s := etrpc.NewServer()
    
    // 首次启动所有任务
    service.GetSchedulerService().RefreshTask()
    
    // 收到停止信号后，取消所有调度任务
    s.RegisterOnShutdown(func() {
        service.GetSchedulerService().CancelTask()
    })
    
    etrpc.RunServer(s)
    
    // 结束后，等待所有调度任务执行完毕
    service.GetSchedulerService().WaitTaskDone()
}
```

## 代码结构

```
scheduler/
├── main.go                              # 服务入口
│                                        # - 注册RegisterService和AdminService
│                                        # - 启动调度任务
│                                        # - 注册优雅关闭hook
├── go.mod                               # Go模块定义
├── trpc_go.yaml                         # 服务配置文件
├── entity/                              # 实体定义
│   ├── consts/                          # 常量定义
│   │   └── consts.go                    # TBosRedisName、TBosMySQLName等
│   ├── dbmodel/                         # 数据库模型
│   │   ├── alarm_strategy.go            # 告警策略表模型
│   │   ├── collector_device.go          # 采集设备表模型
│   │   ├── device_point.go              # 标准测点表模型
│   │   └── mozu_info.go                 # 模组信息表模型
│   └── model/                           # 业务模型
│       ├── task_config.go               # 调度任务配置（TaskConfig、AllTaskConfig）
│       ├── task_item.go                 # 任务项泛型结构（TaskItem[T]）
│       └── worker_info.go               # Worker信息结构
├── logic/                               # 业务逻辑层
│   ├── register/                        # Worker注册逻辑
│   │   └── api.go                       # IRegisterApi接口实现
│   │                                    # - Heartbeat: 心跳注册/注销
│   └── scheduler/                       # 调度核心逻辑
│       ├── scheduler_logic.go           # 通用调度逻辑
│       │                                # - ISchedulerLogic: 调度接口定义
│       │                                # - DefaultSchedulerLogic: 默认实现
│       │                                # - RunTask: 分布式锁调度
│       │                                # - schedulerTask: 调度流程
│       │                                # - doPublish: 执行下发
│       │                                # - buildReq: 构建请求（增量计算）
│       │                                # - DefaultPartitionData: 负载均衡算法
│       │                                # - compareAndMarkWorkers: Worker变化检测
│       ├── alarm_strategy_logic.go      # 告警策略调度实现
│       │                                # - 按ridType分组后均分
│       │                                # - 下发到alarm-compute
│       ├── collector_device_logic.go    # 采集设备调度实现
│       │                                # - 使用默认负载均衡
│       │                                # - 下发到agent
│       └── device_point_logic.go        # 标准测点调度实现
│                                        # - 使用默认负载均衡
│                                        # - 下发到data-compute
├── repo/                                # 数据访问层
│   ├── cache/                           # Redis缓存操作
│   │   └── redis_util.go                # Redis工具类
│   │                                    # - GetRegisterWorkerList: 获取注册Worker
│   │                                    # - GetCacheObj/SetCacheObj: 缓存操作
│   │                                    # - 分片存储支持
│   └── db/                              # 数据库操作
│       ├── scheduler_dao.go             # 通用调度DAO接口
│       │                                # - ISchedulerDao[T]: 泛型DAO接口
│       │                                # - DefaultSchedulerDao: 默认实现
│       │                                # - GetCurVersionStr: 获取当前版本
│       │                                # - Get/SetLastVerStr: 版本缓存
│       │                                # - Get/SetLastWorkerList: Worker列表缓存
│       │                                # - Get/SetLastAssignResult: 分配结果缓存
│       ├── alarm_strategy_dao.go        # 告警策略DAO
│       │                                # - GetPublishData: 获取告警策略数据
│       ├── collector_device_dao.go      # 采集设备DAO
│       │                                # - GetPublishData: 获取采集设备数据
│       │                                # - 计算采集/标准测点数作为复杂度
│       └── device_point_dao.go          # 标准测点DAO
│                                        # - GetPublishData: 获取标准测点数据
├── service/                             # 服务接口层
│   ├── register_service.go              # Worker注册服务
│   │                                    # - Heartbeat: 心跳接口
│   ├── scheduler_service.go             # 调度服务
│   │                                    # - ISchedulerService接口
│   │                                    # - RefreshTask: 刷新任务
│   │                                    # - CancelTask: 取消任务
│   │                                    # - WaitTaskDone: 等待完成
│   │                                    # - schedulerTask: 定时任务结构
│   └── admin_service.go                 # 管理服务
│                                        # - ResetScheduler: 重置调度
│                                        # - ShowAllScheduler: 展示所有调度任务
└── util/                                # 工具类
    ├── convutil/                        # 转换工具
    │   └── conv_util.go                 # SliceToStr等
    └── timezset/                        # 时间有序集合
        └── timezset.go                  # Redis ZSet时间操作（旧版Worker兼容）
```

## 配置说明

### trpc_go.yaml 配置

```yaml
etrpc:
  service_name: scheduler
  service_port: ${PORT_SCHEDULER}

global:
  namespace: Production
  local_ip: ${LOCAL_IP}

server:
  service:
    - name: ${etrpc.service_name}
      protocol: http
      port: ${etrpc.service_port}

client:
  service:
    # MySQL连接
    - name: trpc.mysql.tbos
      target: dns://${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_ADDR})/${MYSQL_DATABASE}
    
    # Redis连接
    - name: trpc.redis.tbos
      target: redis://:${REDIS_PASSWORD}@${REDIS_ADDR}/0
    
    # 下游Worker服务
    - name: data-compute
      callee: tbos.data.Strategy
      protocol: http
      target: ip://${LOCAL_IP}:${PORT_DATA_COMPUTE}
    
    - name: alarm-compute
      callee: tbos.AlarmComputeService
      protocol: http
      target: ip://${LOCAL_IP}:${PORT_ALARM_COMPUTE}
    
    - name: agent
      callee: tbos.TaskConfig
      protocol: http
      target: ip://${LOCAL_IP}:${PORT_AGENT}

# 调度任务配置
scheduler:
  - type: point                # 标准测点调度
    set_group:                 # 园区分组（跨园区调度用）
    filter_mozu: []            # 过滤模组（空表示全部）
  
  - type: alarm                # 告警策略调度
    set_group:
    filter_mozu: []
```

### TaskConfig 配置项说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `type` | string | - | 调度任务类型：`alarm`/`collector`/`point` |
| `disable` | bool | false | 是否禁用 |
| `mysql_name` | string | `trpc.mysql.tbos` | MySQL实例名称 |
| `redis_name` | string | `trpc.redis.tbos` | Redis实例名称 |
| `interval_sec` | int | 30 | 调度间隔（秒） |
| `lock_key` | string | `lock_scheduler#type#setGroup` | 分布式锁Key |
| `lock_key_expire_sec` | int | 60 | 分布式锁过期时长（秒） |
| `set_group` | string | - | 园区分组标识 |
| `filter_mozu` | []int32 | [] | 需要调度的模组ID列表 |
| `old_ver` | bool | false | 是否兼容旧版本Worker |
| `last_assign_shard_cnt` | int | 0 | 分配结果分片数（大数据量时避免BigKey） |

### API接口

| 接口 | 服务 | 方法 | 说明 |
|------|------|------|------|
| `/Heartbeat` | RegisterService | POST | Worker心跳注册 |
| `/ResetScheduler` | AdminService | POST | 重置调度任务（清空缓存） |
| `/ShowAllScheduler` | AdminService | GET | 展示所有调度任务配置 |

#### Heartbeat 请求参数

```protobuf
message WorkerInfo {
    string ip = 1;                    // Worker IP
    int32 port = 2;                   // Worker 端口
    int64 start_time = 3;             // 启动时间戳
    int32 max_process_cap = 4;        // 最大处理能力
    string task_ver_mark = 5;         // 当前任务版本标识
    WorkerType worker_type = 6;       // Worker类型：ALARM/COLLECTOR/POINT
    WorkerStatus worker_status = 7;   // 状态：HEALTHY/SHUTDOWN
    WorkerProtocol worker_protocol = 8; // 协议：HTTP/TRPC
    string worker_set = 9;            // 所属Set（园区）
}
```

## 常见问题

### 1. 调度任务未触发下发

**问题表现**：日志显示 "data version and worker not changed"

**可能原因**：
- 数据版本未变化（`t_mozu_info` 表的版本字段未更新）
- Worker列表未变化

**解决方案**：
- 检查数据库 `t_mozu_info` 表的 `publish_version`（采集/测点）或 `alarm_version`（告警）是否更新
- 调用 `/ResetScheduler` 接口重置调度状态，强制触发全量下发

### 2. Worker未收到任务下发

**问题表现**：Worker日志无任务接收记录

**可能原因**：
- Worker心跳未注册成功
- Worker被判定为过期（心跳超时17秒）
- Worker任务版本与上次一致，无变更任务

**解决方案**：
- 检查Worker是否正常上报心跳（每5秒一次）
- 检查Redis中 `register_worker#type#setGroup` Key是否有Worker记录
- 查看Scheduler日志中的Worker变化检测结果

### 3. 任务分配不均匀

**问题表现**：部分Worker任务过多，部分Worker空闲

**可能原因**：
- 算法优先保持原有分配，导致新Worker分配较少
- Worker `max_process_cap` 设置不合理
- 任务 `compute_cost` 设置不合理

**解决方案**：
- 调用 `/ResetScheduler` 重置分配缓存，触发重新分配
- 调整Worker的 `max_process_cap` 参数
- 检查任务计算复杂度是否合理

### 4. 分布式锁获取失败

**问题表现**：日志中频繁出现锁获取相关错误

**可能原因**：
- Redis连接异常
- 锁过期时间设置过长，上一次调度未正常释放
- 多实例部署时锁竞争激烈

**解决方案**：
- 检查Redis连接状态
- 调整 `lock_key_expire_sec` 配置
- 单机部署时确保只启动一个Scheduler实例

### 5. Worker下发失败重试

**问题表现**：日志出现 "publish data to worker xxx fail, begin remove cur worker and retry"

**原因说明**：
- 下发到某个Worker失败后，会移除该Worker并重新分配任务
- 最多重试3次，每次失败会移除一个Worker

**解决方案**：
- 检查Worker服务是否正常
- 检查网络连通性
- 查看Worker端日志了解拒绝原因

### 6. 大数据量下Redis BigKey

**问题表现**：Redis操作超时，分配结果缓存失败

**原因说明**：
- 分配结果（`last_assign_result`）可能包含大量任务映射

**解决方案**：
- 配置 `last_assign_shard_cnt` 进行分片存储
- 例如设置为10，分配结果会分散到10个Key中存储

### 7. 如何手动触发全量下发

**解决方案**：
调用 AdminService 的 `/ResetScheduler` 接口：
```bash
curl -X POST http://localhost:8080/ResetScheduler \
  -H "Content-Type: application/json" \
  -d '{"type": "alarm", "set_group": ""}'
```

这会清空该类型调度任务的所有缓存（版本、Worker列表、分配结果），下一次调度时会触发全量下发。

### 8. 如何查看当前调度配置

**解决方案**：
调用 AdminService 的 `/ShowAllScheduler` 接口：
```bash
curl http://localhost:8080/ShowAllScheduler
```

返回所有调度任务的配置信息。