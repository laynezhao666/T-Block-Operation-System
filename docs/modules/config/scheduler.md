# 调度中心 Scheduler

调度中心是TBOS系统的核心调度组件，负责将配置数据(告警策略、采集设备、标准测点)按负载均衡算法分配到下游Worker节点，实现任务的分布式调度和管理。

## 数据流向

![Scheduler数据流向](./scheduler_dataflow.svg)

## Worker管理

Scheduler通过调度下游的Worker节点来管理TBOS系统的任务，这些Worker节点包括alarm-compute、agent、data-compute等。

在TBOS中，下游节点Worker分为三种：

- **Collector**：采集Worker
- **Point**：测点计算Worker
- **Alarm**：告警计算Worker

这些节点通过心跳注册来维护其实时状态。

| 方法 | CGI 路径 | 说明 |
|------|----------|------|
| Heartbeat | `/Register/Heartbeat` | Worker 实例心跳上报（含 IP、端口、处理能力、类型） |

## 任务调度

Scheduler是TBOS中配置流的关键枢纽，负责将来自[CMDB](./cmdb.md)的配置数据动态、均匀地分发到执行层（告警计算、采集、数据计算）的各个Worker节点。

Scheduler的设计实现了计算任务的无状态横向扩展和弹性伸缩，是TBOS系统能够应对大规模机房监控计算需求的关键。

其任务调度方案大致如下图所示：

![任务调度流程](./scheduler_process.png)

| 方法 | CGI 路径 | 说明 |
|------|----------|------|
| ResetScheduler | `/Admin/ResetScheduler` | 重置指定类型的调度任务 |
| ShowAllScheduler | `/Admin/ShowAllScheduler` | 展示所有调度任务详情（含分布式锁、Redis、分片等配置） |

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
