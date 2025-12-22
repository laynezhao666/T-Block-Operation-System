# Alarm-Compute 告警计算引擎

Alarm-Compute 是 TBOS 系统的告警计算服务，负责接收调度中心下发的告警策略，对测点数据进行实时计算和判断，产生告警并发送到 Kafka。支持实时告警、延时告警和虚拟测点告警三种类型。

## 总体架构

```
                        ┌─────────────────────┐
                        │  Scheduler          │
                        │  (策略调度中心)      │
                        └──────────┬──────────┘
                                   │
                                   │ 策略推送 (gRPC)
                                   │ 心跳上报 (HTTP)
                                   ▼
         ┌─────────────────────────────────────────────────┐
         │         Alarm-Compute 告警计算引擎              │
         │                                                  │
         │  ┌──────────────┐  ┌──────────────┐            │
         │  │ 心跳上报      │  │ 策略接收器    │            │
         │  │ (每5秒)       │  │ (接收推送)    │            │
         │  └──────────────┘  └──────────────┘            │
         │                                                  │
         │  ┌──────────────────────────────────────────┐  │
         │  │ 规则管理器 (RuleManager)                 │  │
         │  │                                          │  │
         │  │ ┌─────────────┐ ┌─────────────┐        │  │
         │  │ │ 实时规则    │ │ 延时规则    │        │  │
         │  │ │ (1秒周期)   │ │ (3秒周期)   │        │  │
         │  │ │ 查询变更点  │ │ 查询历史段  │        │  │
         │  │ └─────────────┘ └─────────────┘        │  │
         │  │                                          │  │
         │  │ ┌─────────────────────────────────┐    │  │
         │  │ │ 虚拟规则 (2秒周期)              │    │  │
         │  │ │ 计算虚拟测点 → Kafka            │    │  │
         │  │ └─────────────────────────────────┘    │  │
         │  └──────────────────────────────────────────┘  │
         │                                                  │
         │  ┌──────────────────────────────────────────┐  │
         │  │ 表达式求值引擎 (TNQL)                   │  │
         │  │ 支持: 比较、逻辑、算术、函数             │  │
         │  └──────────────────────────────────────────┘  │
         │                                                  │
         │  ┌──────────────┐  ┌──────────────┐            │
         │  │ 验证收集器    │  │ 虚拟点采集    │            │
         │  │ (批处理)      │  │ (Kafka发送)   │            │
         │  └──────────────┘  └──────────────┘            │
         │                                                  │
         │  ┌──────────────────────────────────────────┐  │
         │  │ 本地缓存 (活动告警状态)                 │  │
         │  │ 定期同步 from MySQL                      │  │
         │  └──────────────────────────────────────────┘  │
         │                                                  │
         └─────────────────────────────────────────────────┘
                     │      │      │      │
                     │      │      │      │
        ┌────────────┘      │      │      └────────────┐
        │                   │      │                   │
        ▼                   ▼      ▼                   ▼
    ┌────────────┐  ┌──────────┐ ┌──────────┐  ┌──────────────┐
    │  Kafka     │  │  MySQL   │ │Data-Query│  │Alarm-Manage  │
    │            │  │          │ │ (测点)   │  │ (告警推送)   │
    │ - 告警消息 │  │ - 活动   │ │          │  │              │
    │ - 验证消息 │  │ - 历史   │ └──────────┘  └──────────────┘
    │ - 虚拟点   │  └──────────┘
    └────────────┘
```

### 数据流转

1. **策略下发**：Scheduler → alarm-compute (gRPC) → RuleManager
2. **测点查询**：RuleManager → data-query (HTTP) → 测点值
3. **告警计算**：测点值 → TNQL引擎 → 告警判断 → 本地缓存
4. **告警发送**：告警消息 → Kafka (降级→ alarm-manage HTTP API)
5. **策略验证**：执行结果 → Kafka (批量1000条或1秒)
6. **虚拟测点**：虚拟点值 → Kafka (批量200条或100ms)
7. **心跳上报**：worker状态 → Scheduler (每5秒)

## 模块介绍

### 核心能力

1. **策略接收与管理**：通过 gRPC 接收 scheduler 下发的告警策略，支持增量和全量更新
2. **三种告警类型**：实时告警（立即触发）、延时告警（持续满足条件后触发）、虚拟测点告警（基于计算结果触发）
3. **表达式计算引擎**：基于 TNQL 引擎，支持复杂的逻辑、比较和算术运算，以及多种内置函数
4. **测点数据查询**：从 data-query 服务获取实时和历史测点数据
5. **告警消息发送**：通过 Kafka 发送告警消息，支持降级到 HTTP API
6. **活动告警缓存**：本地缓存活动告警状态，避免重复告警
7. **心跳上报**：定期向 scheduler 上报 worker 状态和任务版本
8. **策略验证**：批量收集策略执行结果并上报
9. **失败重试**：对执行失败的规则进行重试

### 告警类型说明

| 类型 | 执行周期 | 触发条件 | 数据查询方式 |
|------|---------|---------|-------------|
| 实时告警 | 1秒 | 表达式条件满足立即触发 | 查询变更测点的最新值 |
| 延时告警 | 3秒 | 表达式条件持续满足指定时间后触发 | 查询测点的历史数据段 |
| 虚拟测点告警 | 2秒 | 基于虚拟测点计算结果触发 | 查询依赖的实测点值，计算虚拟点 |

## 代码结构

```
alarm-compute/
├── main.go                           # 主入口：启动6个后台协程和gRPC服务
├── trpc_go.yaml                      # 服务配置文件
│
├── conf/                             # 配置管理
│   └── conf.go                       # 配置结构体定义
│
├── entity/                           # 实体定义
│   ├── alarm_config.go               # 告警配置结构（策略下发的数据模型）
│   ├── epoint/                       # 测点相关实体
│   │   ├── point.go                  # 测点定义
│   │   ├── point_value.go            # 测点值结构
│   │   └── delay_point.go            # 延迟测点结构
│   ├── errcode/                      # 错误码定义
│   └── taskcode/                     # 任务类型码定义
│
├── logic/                            # 核心业务逻辑
│   ├── strategy/                     # 策略处理
│   │   └── strategy_handler.go       # 策略接收器：从scheduler接收策略，更新规则管理器
│   │
│   ├── rules/                        # 规则引擎（核心）
│   │   ├── rmanager/                 # 规则管理器
│   │   │   ├── rule_manager.go       # 全局规则管理：存储和管理所有规则
│   │   │   ├── manager_realtime.go   # 实时规则执行逻辑：查询变更测点，并发执行规则
│   │   │   ├── manager_delaytime.go  # 延时规则执行逻辑：查询历史数据，判断持续时长
│   │   │   └── manager_virtual.go    # 虚拟规则执行逻辑：计算虚拟测点值
│   │   └── rtask/                    # 规则任务
│   │       ├── rule_task.go          # RuleTask结构定义（规则的运行时对象）
│   │       ├── rule_task_virtual.go  # 虚拟规则任务实现
│   │       ├── rule_task_delaytime.go # 延时规则任务实现
│   │       ├── alarm_task.go         # AlarmTask结构定义（告警/恢复计算任务）
│   │       ├── alarm_task_virtual.go # 虚拟点告警任务实现
│   │       ├── alarm_task_delaytime.go # 延时告警任务实现
│   │       └── alert_produce.go      # 告警消息生成和发送
│   │
│   ├── heartbeat/                    # 心跳管理
│   │   └── heartbeat.go              # 定期向scheduler上报worker状态
│   │
│   ├── collector/                    # 数据采集器
│   │   ├── validate/                 # 策略验证采集器
│   │   │   ├── validate.go           # 批量收集策略执行结果并发送到Kafka
│   │   │   └── failed_rule.go        # 收集失败规则并重新执行
│   │   └── vtpoint/                  # 虚拟测点采集器
│   │       └── virtualPoint.go       # 收集虚拟测点数据并发送到Kafka
│   │
│   ├── pointeval/                    # 表达式计算引擎
│   │   ├── alarm_point.go            # 告警点表达式求值
│   │   ├── alarm_point_func.go       # 内置函数实现
│   │   ├── alarm_point_data.go       # 测点数据处理
│   │   └── analyze_result.go         # 计算结果分析
│   │
│   ├── point/                        # 测点管理
│   │   ├── point_valid.go            # 测点有效性验证
│   │   ├── point_value.go            # 测点值获取和处理
│   │   └── tools.go                  # 测点工具函数
│   │
│   ├── lcache/                       # 本地缓存
│   │   └── lcache.go                 # 活动告警状态缓存（避免重复告警）
│   │
│   └── diagnose/                     # 诊断工具
│       └── api.go                    # 表达式诊断API
│
├── service/                          # gRPC服务接口
│   └── alarm_service.go              # 实现两个接口：
│                                     # - RecvTask: 接收策略任务
│                                     # - ExpCompute: 表达式诊断计算
│
├── repo/                             # 数据访问层
│   ├── ckafka.go                     # Kafka客户端
│   │                                 # - SendAlertMsg: 发送告警消息
│   │                                 # - SendRuleValidMsg: 发送规则验证消息
│   │                                 # - SendPointMsg: 发送虚拟测点数据
│   │                                 # - SendAdminMsg: 发送运营平台消息
│   ├── point_data.go                 # 测点数据查询服务
│   │                                 # - GetPointDataTS: 获取测点最新值
│   │                                 # - GetPointDurationDataTS: 获取历史数据段
│   │                                 # - ParallelGetChangedPointList: 并发查询变更测点
│   ├── alarm_dao.go                  # 告警数据库访问（MySQL）
│   └── alarm_push.go                 # 告警推送API（降级方案）
│
└── utils/                            # 工具类
    ├── common/                       # 通用工具
    │   ├── slice.go                  # 切片操作
    │   ├── json.go                   # JSON处理
    │   └── trace.go                  # 链路追踪
    ├── modcall/                      # 指标记录
    │   └── metrics.go                # 记录性能指标
    └── tnql/                         # 表达式解析引擎
        ├── parsing.go                # 表达式解析器
        ├── func_symbols.go           # 支持的函数定义
        ├── operator_symbol.go        # 操作符定义
        └── evaluable_expression.go   # 可求值表达式
```

## 核心流程

### 1. 服务启动流程

```
main()
  │
  ├─> 创建 TRPC 服务器
  ├─> 创建全局 Context（用于控制所有协程生命周期）
  │
  ├─> 启动 6 个后台协程：
  │   │
  │   ├─> heartbeat.ReportHeartbeat()
  │   │   定期向 scheduler 上报心跳（默认5秒）
  │   │
  │   ├─> strategy.Run()
  │   │   接收来自 scheduler 的策略下发
  │   │   更新 RuleManager 中的规则
  │   │   同步活动告警到本地缓存
  │   │
  │   ├─> validate.StartValidate()
  │   │   批量收集策略执行结果
  │   │   发送到 Kafka（批量1000条或间隔1秒）
  │   │
  │   ├─> vtpoint.ReportVtPointData()
  │   │   收集虚拟测点数据
  │   │   发送到 Kafka（批量200条或间隔100ms）
  │   │
  │   ├─> validate.CollectFailed()
  │   │   收集执行失败的规则
  │   │   加入下次重试队列（间隔5秒）
  │   │
  │   └─> run()  # 规则计算引擎（核心）
  │       ├─> StartRealTimeRuleTask()  # 实时规则（1秒周期）
  │       ├─> StartDelayTimeRuleTask() # 延时规则（3秒周期）
  │       └─> StartVirtualRuleTask()   # 虚拟规则（2秒周期）
  │
  ├─> 注册 gRPC 服务（RecvTask、ExpCompute）
  ├─> 运行服务器
  └─> 等待所有协程完成
```

### 2. 策略下发流程

```
Scheduler
  │
  │ gRPC调用 RecvTask
  │ 携带策略数据（AddTask、DelTaskKey）
  ▼
StrategyHandler.AddStrategyReq()
  │ 将策略请求放入通道
  ▼
StrategyHandler.HandleStrategyReq()
  │
  ├─> 删除旧规则
  │   └─> RuleManager.DelRuleTaskByKey()
  │       解析 key（rid;gid;version）
  │       从对应的 RuleMap 中删除
  │       从 PointRuleMap 中删除测点映射
  │
  ├─> 解析新规则
  │   └─> parseStrategyPb2Alarm()
  │       JSON 反序列化 ExpressionMap
  │       构造 AlarmConfig 对象
  │
  ├─> 添加新规则
  │   └─> RuleManager.AddRuleTasks()
  │       └─> TransferConfig2Rules()
  │           └─> InitRuleTask()
  │               创建 RuleTask 对象
  │               设置 Alert 和 Restore 任务
  │               解析表达式（Evaluate DryRun）
  │               存储到 RuleMap 和 PointRuleMap
  │
  └─> 同步活动告警
      └─> readActiveAlarmFromDB()
          从 MySQL 查询活动告警（rid;gid）
          写入本地缓存（lcache）
          TTL: 86400 秒（1天）
```

### 3. 实时规则执行流程

```
StartRealTimeRuleTask (每1秒触发一次)
  │
  ├─> 判断是否全量分析
  │   每24个周期执行一次全量分析
  │   其他周期执行增量分析
  │
  ▼ 增量分析
GetExecRealTimeRuleTask()
  │
  ├─> 获取所有实时规则的测点列表
  │   └─> RtPointRuleMap.GetTotalPointList()
  │
  ├─> 查询最近N秒变更的测点
  │   └─> ParallelGetChangedPointList()
  │       并发查询（协程池200，批量3000）
  │       调用 data-query 服务
  │       返回变更的测点列表
  │
  ├─> 根据变更测点找到关联的规则
  │   遍历变更测点
  │   从 RtPointRuleMap 查找对应的规则key
  │   加入 execRules
  │
  └─> 加入失败规则重试集合
      └─> GetFailRuleCollector().GetFailedRt()

EvalRealTimeRuleTaskWithTime (执行规则计算)
  │
  ├─> 将规则分批（批量200）
  │
  ├─> 使用协程池并发执行（池大小100）
  │   对每批规则：
  │   │
  │   ├─> 获取测点值
  │   │   └─> BatchGetRTPointValue()
  │   │       收集所有规则需要的测点
  │   │       去重后批量查询
  │   │       调用 data-query 服务
  │   │
  │   ├─> 对每个规则并发执行
  │   │   └─> RuleTask.StartRealtimeByData()
  │   │       │
  │   │       ├─> 检查活动告警缓存
  │   │       │   └─> lcache.CheckActiveAlarmCache()
  │   │       │
  │   │       ├─> 告警表达式求值
  │   │       │   └─> AlarmTask.Evaluate()
  │   │       │       使用 TNQL 引擎计算表达式
  │   │       │       传入测点值映射
  │   │       │       返回 true/false
  │   │       │
  │   │       ├─> 恢复表达式求值（如果有）
  │   │       │   └─> AlarmTask.Evaluate()
  │   │       │
  │   │       ├─> 判断告警状态
  │   │       │   │
  │   │       │   ├─ 缓存中无告警 + 告警条件满足
  │   │       │   │  └─> SendAlert()
  │   │       │   │      生成 FireAlertMsg
  │   │       │   │      缓存告警状态（lcache）
  │   │       │   │      发送到 Kafka（带重试）
  │   │       │   │      Kafka失败则调用API
  │   │       │   │
  │   │       │   └─ 缓存中有告警 + 恢复条件满足
  │   │       │      └─> SendRestoreAlert()
  │   │       │          生成 FireAlertMsg（设置 EndAt）
  │   │       │          删除缓存（lcache）
  │   │       │          发送到 Kafka
  │   │       │
  │   │       └─> 记录验证信息
  │   │           └─> ValidCollector.AddValidateRecord()
  │   │               记录执行状态和结果
  │   │               批量发送到 Kafka
  │   │
  │   └─> 收集失败的规则
  │       测点数据缺失导致的失败
  │       加入失败重试队列
  │
  └─> 等待所有批次完成
```

### 4. 延时规则执行流程

```
StartDelayTimeRuleTask (每3秒触发一次)
  │
  ├─> 判断是否全量分析
  │   每6个周期执行一次全量分析
  │
  ▼ 增量分析
GetExecDelayTimeRuleTask()
  │
  ├─> 获取所有延时规则的测点列表
  │   └─> DtPointRuleMap.GetTotalPointList()
  │
  ├─> 查询最近N秒变更的测点
  │   └─> ParallelGetChangedPointList()
  │
  └─> 根据变更测点找到关联的规则

EvalDelayTimeRuleTaskWithTime (执行延时规则计算)
  │
  ├─> 将规则分批
  │
  ├─> 使用协程池并发执行（池大小6）
  │   对每个规则：
  │   │
  │   ├─> 获取测点历史数据
  │   │   └─> GetPointDurationDataTS()
  │   │       根据规则的 PointDelayMap
  │   │       查询每个测点的时间段数据
  │   │       并发查询（协程池400）
  │   │       返回 HistoryValueMap
  │   │
  │   ├─> 对历史数据的每个时间点求值
  │   │   └─> RuleTask.StartDelayTimeByData()
  │   │       │
  │   │       ├─> 遍历历史时间点
  │   │       │   对每个时间点：
  │   │       │   └─> AlarmTask.Evaluate()
  │   │       │       计算该时间点的表达式结果
  │   │       │
  │   │       ├─> 判断是否持续满足延迟条件
  │   │       │   检查连续满足的时长
  │   │       │   是否超过配置的延迟时间
  │   │       │
  │   │       ├─> 检查活动告警缓存
  │   │       │   └─> lcache.CheckActiveAlarmCache()
  │   │       │
  │   │       ├─> 产生告警或恢复
  │   │       │   │
  │   │       │   ├─ 无告警 + 持续满足延迟
  │   │       │   │  └─> SendAlert()
  │   │       │   │
  │   │       │   └─ 有告警 + 恢复条件满足
  │   │       │      └─> SendRestoreAlert()
  │   │       │
  │   │       └─> 记录验证信息
  │   │
  │   └─> 收集失败的规则
  │
  └─> 等待所有规则完成
```

### 5. 虚拟规则执行流程

```
StartVirtualRuleTask (每2秒触发一次)
  │
  ├─> 判断是否全量分析
  │   变更测点为空累计3次则触发全量
  │
  ▼ 增量分析
GetExecVirtualRuleTask()
  │
  ├─> 获取所有虚拟规则的测点列表
  │   └─> VtPointRuleMap.GetTotalPointList()
  │
  ├─> 查询最近N秒变更的测点
  │   └─> ParallelGetChangedPointList()
  │
  └─> 根据变更测点找到关联的规则

EvalVirtualRuleTaskWithTime (执行虚拟规则计算)
  │
  ├─> 将规则分批
  │
  ├─> 使用协程池并发执行
  │   对每个规则：
  │   │
  │   ├─> 获取依赖的实测点值
  │   │   └─> BatchGetRTPointValue()
  │   │
  │   ├─> 计算虚拟测点值
  │   │   └─> RuleTask.StartVirtualByData()
  │   │       │
  │   │       ├─> 计算虚拟测点表达式
  │   │       │   └─> AlarmTask.Evaluate()
  │   │       │       返回虚拟测点的计算值
  │   │       │
  │   │       ├─> 四舍五入到指定精度
  │   │       │   └─> RoundPrecision (默认2位小数)
  │   │       │
  │   │       ├─> 生成虚拟测点消息
  │   │       │   └─> VirtualPointMsg
  │   │       │       包含 rid、gid、mozuId
  │   │       │       虚拟测点值、时间戳
  │   │       │
  │   │       ├─> 发送虚拟测点到 Kafka
  │   │       │   └─> PointCollector.AddVirtualPointData()
  │   │       │       批量发送（200条或100ms）
  │   │       │
  │   │       └─> 虚拟规则也用于告警判断
  │   │           如果虚拟规则有告警表达式
  │   │           则执行告警逻辑（类似实时告警）
  │   │
  │   └─> 记录验证信息
  │
  └─> 等待所有规则完成
```

### 6. 表达式计算引擎

alarm-compute 使用 TNQL（基于 govaluate 修改）作为表达式计算引擎。

#### 支持的运算符

| 类型 | 运算符 | 示例 |
|------|--------|------|
| 比较运算 | `>` `<` `>=` `<=` `==` `!=` | `temperature > 30` |
| 逻辑运算 | `&&` `||` `!` | `temp > 30 && humidity > 80` |
| 算术运算 | `+` `-` `*` `/` `%` | `(power_a + power_b) > 1000` |
| 括号 | `(` `)` | `(a + b) * c` |

#### 支持的函数

详见 `utils/tnql/func_symbols.go`，常用函数包括：

| 函数 | 说明 | 示例 |
|------|------|------|
| DelayEQ | 延迟相等检查 | `DelayEQ(A, 10, 30)` # A==10持续30秒 |
| Rise | 上升沿检测 | `Rise(A)` # A从0变为1 |
| Fall | 下降沿检测 | `Fall(A)` # A从1变为0 |
| InRange | 范围检查 | `InRange(A, 10, 20)` # 10 <= A <= 20 |
| Avg | 平均值 | `Avg(A, B, C)` |
| Max | 最大值 | `Max(A, B, C)` |
| Min | 最小值 | `Min(A, B, C)` |

#### 表达式求值过程

```
表达式字符串
  │ "temperature > 30 && humidity > 80"
  ▼
词法分析（Lexer）
  │ 生成 Token 流
  ▼
语法分析（Parser）
  │ 构建表达式树
  ▼
EvaluableExpression
  │ 创建可求值对象
  ▼
Evaluate(parameters)
  │ 传入测点值映射
  │ {
  │   "temperature": 35.5,
  │   "humidity": 85.0
  │ }
  ▼
阶段求值（Stage Planner）
  │ 按优先级计算
  │ 1. temperature > 30  → true
  │ 2. humidity > 80     → true
  │ 3. true && true      → true
  ▼
返回结果
  │ true (触发告警)
```

## 外部依赖

### 1. Scheduler（调度中心）

**作用**：策略调度和 worker 管理

**通信方式**：
- Scheduler → alarm-compute: gRPC（RecvTask 接口）
- alarm-compute → Scheduler: HTTP（Heartbeat 接口）

**主要接口**：
- `RecvTask`: 接收策略下发（增量/全量）
- `Heartbeat`: 上报 worker 状态（IP、端口、任务版本、处理能力）

### 2. Data-Query（测点查询服务）

**作用**：查询测点实时值和历史数据

**通信方式**：HTTP

**主要接口**：
- `DataQuery()`: 查询测点值
- `DataChange()`: 获取测点变更时间戳
- `DataPointChange()`: 批量获取变更测点列表

### 3. Kafka（消息队列）

**作用**：告警消息、验证消息、虚拟测点数据传输

**使用的 Topic**：

| Kafka Client | Topic 用途 | 消息类型 |
|-------------|-----------|---------|
| trpc.kafka.tbos.alert | 告警消息 | FireAlertMsg（告警产生/恢复） |
| trpc.kafka.tbos.rule | 规则验证消息 | ValidateTaskList（策略执行结果） |
| trpc.kafka.tbos.data | 虚拟测点数据 | VirtualPointMsg（虚拟测点值） |
| trpc.kafka.tbos.admin | 运营平台消息 | MQAdminPointMsg（虚拟点数据） |

**特性**：
- 告警消息发送带重试（3次）
- 告警消息失败降级到 HTTP API

### 4. MySQL（数据库）

**作用**：活动告警持久化

**主要表**：
- `t_alarm_active`: 活动告警表
- `t_alarm_history`: 告警历史表（可选）

**使用场景**：
- 策略变更时同步活动告警到本地缓存
- 定期同步活动告警（默认300秒）

### 5. Alarm-Manage（告警管理服务）

**作用**：告警推送（Kafka 失败时的降级方案）

**通信方式**：HTTP

**主要接口**：
- `PushAlarmByApi()`: 通过 API 推送告警

## 配置说明

### 核心配置项

```yaml
# 实时规则配置
RealTimeConfig:
  RealTimeTaskInterval: 1            # 执行周期（秒）
  TotalAnalyzeCycleCount: 24         # 全量分析周期（每24次增量后执行一次全量）
  ParallelExecWpSize: 5              # 周期执行协程池大小
  RealTimeBatchSize: 200             # 规则分批大小
  RealTimeTaskPoolSize: 100          # 批内并发协程池大小
  VaryPointQueryTimeSpan: 5          # 查询变更测点的时间跨度（秒）
  VaryPointBatchSize: 3000           # 变更测点批量查询大小
  VaryPointPoolSize: 200             # 变更测点并发查询协程池

# 延时规则配置
DelayTimeConfig:
  DelayTimeTaskInterval: 3           # 执行周期（秒）
  TotalAnalyzeCycleCount: 6          # 全量分析周期
  ParallelExecWpSize: 6              # 周期执行协程池大小
  IntervalRequestPoolSize: 100       # 查询历史数据协程池（按interval）
  IntervalBatchPointSize: 300        # 历史数据批量大小
  DurationRequestPoolSize: 400       # 查询历史数据协程池（按duration）
  DurationBatchPointSize: 50         # 历史数据批量大小
  JPRangeSec: 30                     # 跳变检测时间范围（秒）

# 虚拟规则配置
VirtualConfig:
  VirtualTaskInterval: 2             # 执行周期（秒）
  PointKafkaBatchSize: 200           # 虚拟点批量发送大小
  FlushInterval: 100                 # 虚拟点发送间隔（毫秒）
  RoundPrecision: 2                  # 虚拟点值四舍五入精度

# 活动告警缓存配置
ActiveAlarmCache:
  ActiveNormalSyncInterval: 300      # 定期同步间隔（秒）
  ActiveRequestBatchSize: 5000       # 从数据库查询批量大小
  CacheKeyTimeDuration: 86400        # 缓存TTL（秒）

# 心跳配置
HeartBeatConfig:
  HeartBeatInterval: 5               # 心跳上报间隔（秒）

# 验证记录配置
ValidateRecordConfig:
  BatchSize: 1000                    # 验证消息批量大小
  FlushInterval: 1000                # 验证消息发送间隔（毫秒）
  FailedDispatchInterval: 5000       # 失败规则重试间隔（毫秒）
```

### 并发控制说明

alarm-compute 使用多级协程池控制并发：

1. **周期级并发**：控制每秒/每几秒启动多少个计算周期
   - 实时规则：5个协程
   - 延时规则：6个协程

2. **规则级并发**：控制同时计算多少个规则
   - 实时规则：先分批（200个/批），每批内100个协程并发

3. **数据查询并发**：控制同时查询多少个测点
   - 变更测点查询：200个协程，批量3000个测点
   - 延时历史查询：400个协程（duration），100个协程（interval）

## 常见问题

### 1. 策略下发后未生效

**可能原因**：
- 策略下发失败（通道满）
- 表达式解析失败
- 测点映射配置错误

**排查方法**：
- 查看日志：`strategy req, timeStamp`（策略接收日志）
- 查看日志：`realtimeCount: %d, delaytimeCount: %d, virtualCount: %d`（规则数量统计）
- 检查错误日志：`parse config error`（配置解析失败）
- 使用 ExpCompute 接口诊断表达式

**解决方案**：
- 检查 StrategyCh 通道容量（默认10）
- 验证表达式语法正确性
- 确认 ExpressionMap 中的测点名称正确

### 2. 告警未触发

**可能原因**：
- 测点数据未变更（实时规则依赖变更测点）
- 测点数据缺失
- 表达式计算结果不满足条件
- 活动告警缓存中已存在（避免重复告警）

**排查方法**：
- 查看日志：`get changed real rule`（找到的变更规则数量）
- 查看日志：`rt failed due to the lack of points`（测点缺失）
- 查看日志：`BatchGetPointValueRealTime failed`（测点查询失败）
- 使用 ExpCompute 接口诊断表达式计算

**解决方案**：
- 确认测点有数据上报
- 检查 data-query 服务状态
- 检查表达式是否正确
- 检查活动告警缓存（lcache）是否正确同步

### 3. 延时告警不准确

**可能原因**：
- 计算周期配置不当（默认3秒）
- 测点历史数据不完整
- 延时时间配置与数据采集频率不匹配

**排查方法**：
- 查看配置：`DelayTimeTaskInterval`（计算周期）
- 查看日志：查询历史数据的时间范围
- 检查测点的采集频率

**解决方案**：
- 延时精度受计算周期影响，最小精度为计算周期
- 确保历史数据查询范围覆盖延时时间
- 调整 JPRangeSec 配置（跳变检测范围）

### 4. 告警消息发送失败

**可能原因**：
- Kafka 连接失败
- Kafka topic 不存在
- 消息序列化失败

**排查方法**：
- 查看日志：`发送告警失败`
- 查看日志：`发送告警信息，Kafka和接口均失败`（双重失败）
- 查看日志：`发送告警信息，Kafka失败，调用接口成功`（降级成功）
- 检查 Kafka 服务状态

**解决方案**：
- 检查 Kafka 配置和连接
- 确认 topic 存在且有写权限
- 检查 alarm-manage 服务状态（降级方案）
- 告警消息带3次重试，降级到 HTTP API

### 5. 虚拟测点计算错误

**可能原因**：
- 依赖的实测点数据缺失
- 虚拟点表达式错误
- 四舍五入精度配置不当

**排查方法**：
- 查看日志：`failed parse alarm virtual expression`
- 检查虚拟规则的 Alert.Exp.PMap（测点映射）
- 使用 ExpCompute 接口诊断虚拟点表达式

**解决方案**：
- 确认所有依赖的实测点有数据
- 验证虚拟点表达式语法
- 调整 RoundPrecision 配置

### 6. 内存占用过高

**可能原因**：
- 规则数量过多
- 测点数据缓存过多
- 协程泄漏
- 活动告警缓存未清理

**排查方法**：
- 查看规则数量：`realtimeCount: %d, delaytimeCount: %d, virtualCount: %d`
- 检查协程数量和泄漏
- 检查缓存大小

**解决方案**：
- 减少单实例规则数量，增加 alarm-compute 实例
- 调整活动告警缓存 TTL（CacheKeyTimeDuration）
- 调整批量大小配置，减少内存占用
- 定期重启服务（生产环境需配合监控）

### 7. 规则执行延迟

**可能原因**：
- 规则数量过多，协程池不足
- 测点查询耗时过长
- 表达式计算复杂

**排查方法**：
- 查看性能指标：`RecordStrategyTimeCost`（规则计算耗时）
- 查看性能指标：`RecordDataQueryTime`（数据查询耗时）
- 查看日志：协程池满（`realtime invoke failed`）

**解决方案**：
- 增加协程池大小（ParallelExecWpSize、RealTimeTaskPoolSize）
- 优化表达式，减少复杂度
- 增加 alarm-compute 实例分担负载
- 调整计算周期（增加 RealTimeTaskInterval）

### 8. 策略验证消息丢失

**可能原因**：
- Kafka 发送失败
- 批量累积未达到阈值未发送

**排查方法**：
- 查看日志：`SendRuleValidMsg err`
- 检查 ValidateRecordConfig 配置

**解决方案**：
- 确认 Kafka 连接正常
- 验证消息批量大小和间隔配置合理
- 服务关闭时确保 flush 剩余消息

## 性能优化建议

### 1. 表达式优化
- 简化复杂表达式，减少嵌套层级
- 避免在表达式中使用过多函数调用
- 使用预编译表达式（TNQL 会缓存解析结果）

### 2. 并发优化
- 根据 CPU 核数调整协程池大小
- 批量大小与协程池大小需要平衡
- 避免过多协程导致上下文切换开销

### 3. 数据查询优化
- 合理设置变更测点查询时间跨度（VaryPointQueryTimeSpan）
- 批量查询测点数据，减少 RPC 调用次数
- 并发查询协程池不宜过大（避免打爆 data-query）

### 4. 缓存优化
- 活动告警缓存避免重复告警，TTL 设置合理
- 定期同步数据库，确保缓存准确性
- 缓存批量查询大小要适中

### 5. 消息发送优化
- 批量发送消息到 Kafka，减少网络开销
- FlushInterval 和 BatchSize 需要平衡（延迟 vs 吞吐量）
- 异步发送，避免阻塞规则计算

## 监控指标

alarm-compute 通过 `utils/modcall` 记录性能指标（需配合监控系统）：

| 指标名称 | 说明 |
|---------|------|
| RecordAnalyzeTaskCnt | 规则分析数量（区分全量/增量） |
| RecordStrategyTimeCost | 规则计算耗时（毫秒） |
| RecordDataQueryTime | 数据查询耗时（毫秒） |
| RecordDataChangeTime | 变更测点查询耗时（毫秒） |
| RecordProduceAlertCnt | 告警产生数量（按模组ID） |
| RecordRuleValidKafkaCnt | 验证消息发送次数 |

## 开发调试

### 表达式诊断

使用 `ExpCompute` 接口诊断表达式计算：

```bash
# gRPC 调用示例
grpcurl -plaintext -d '{
  "expression": "temperature > 30 && humidity > 80",
  "expression_map": "{\"temperature\":[\"device001.temp\"],\"humidity\":[\"device001.hum\"]}",
  "begin_time": 1700000000,
  "end_time": 1700000300,
  "interval": 60
}' localhost:8080 AlarmCompute/ExpCompute
```

返回每个时间点的计算结果和测点值，用于排查表达式问题。

### 日志级别

关键日志：
- `strategy req`：策略下发
- `realtimeCount`：规则数量统计
- `get changed real rule`：变更规则数量
- `start to eval realtime rule task`：开始计算
- `发送告警消息成功`：告警发送成功
- `rt failed due to the lack of points`：测点缺失

错误日志：
- `parse config error`：策略解析失败
- `BatchGetPointValueRealTime failed`：测点查询失败
- `发送告警失败`：Kafka 发送失败
- `realtime invoke failed`：协程池满

## 总结

alarm-compute 是一个高性能的告警计算引擎，采用多级协程池实现并发计算，支持三种告警类型，具备完善的降级和重试机制。核心特点：

1. **高并发**：多级协程池，支持大规模规则并发计算
2. **高可用**：Kafka 失败降级到 HTTP API，失败规则自动重试
3. **高性能**：批量查询测点数据，异步发送消息，表达式预编译
4. **灵活性**：支持复杂表达式和多种内置函数
5. **可观测**：详细的日志和性能指标

代码库采用清晰的分层架构：
- **service**：对外接口层
- **logic**：业务逻辑层（策略、规则、计算、采集）
- **repo**：数据访问层（Kafka、MySQL、HTTP）
- **utils**：工具层（表达式引擎、通用工具）
