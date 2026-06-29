# 告警服务 Alarm-Server

Alarm-Server 是 TBOS 系统的核心告警服务模块，负责告警数据的查询、策略验证记录的收集上报、以及本地缓存管理，为上层应用提供高性能的告警数据查询接口。

## 系统架构

### 整体架构图

```
                        ┌─────────────────────────────────┐
                        │      外部客户端/上层服务           │
                        │    (运维平台、告警前端等)          │
                        └──────────────┬──────────────────┘
                                       │ gRPC 请求
                                       ▼
        ┌──────────────────────────────────────────────────────────┐
        │                   Service API Layer                       │
        │              (service/api/api.go)                         │
        │  提供11个核心API接口: 告警查询、策略查询、状态更新等         │
        └───────┬────────────────────┬────────────────┬─────────────┘
                │                    │                │
                │                    │                │
        ┌───────▼──────┐     ┌──────▼──────┐  ┌─────▼──────────┐
        │  Alarm Logic │     │Strategy Logic│  │Cache Management│
        │  (告警逻辑)   │     │  (策略逻辑)   │  │  (缓存管理)     │
        └───────┬──────┘     └──────┬──────┘  └─────┬──────────┘
                │                    │                │
                │                    │                │
        ┌───────▼────────────────────▼────────────────▼──────────┐
        │              Data Access Layer (DAO)                    │
        │  ┌──────────┐  ┌──────────┐  ┌────────┐  ┌──────────┐ │
        │  │Alarm DAO │  │Strategy  │  │Redis   │  │RPC调用   │ │
        │  │          │  │DAO       │  │Store   │  │(CMDB等)  │ │
        │  └──────────┘  └──────────┘  └────────┘  └──────────┘ │
        └─────┬───────────────┬───────────────┬──────────┬───────┘
              │               │               │          │
              │               │               │          │
     ┌────────▼──────┐  ┌────▼────┐  ┌───────▼────┐  ┌─▼────────┐
     │  MySQL 主库    │  │MySQL只读│  │   Redis    │  │外部服务   │
     │ (t_alarm_*)   │  │  副本   │  │(验证记录)   │  │(CMDB等)  │
     └────────────────┘  └─────────┘  └───────┬────┘  └──────────┘
                                              │
                    ┌─────────────────────────┴──────────────────┐
                    │                                            │
            ┌───────▼───────┐                         ┌─────────▼────────┐
            │ Kafka Consumer│                         │  Kafka Producer  │
            │  (消费验证消息)│                         │  (上报生效率)     │
            └───────┬───────┘                         └─────────▲────────┘
                    │                                           │
            ┌───────▼────────────┐                  ┌──────────┴─────────┐
            │  Collector Module  │──────────────────▶│   Timer Tasks      │
            │  (验证记录收集器)    │  定时触发上报      │  (定时任务调度)     │
            └────────────────────┘                  │ • 生效率上报(10-40s)│
                                                    │ • 过期清理(每日)    │
                                                    └────────────────────┘

            ┌────────────────────────────────────────────────────────────┐
            │                     Background Services                    │
            │  • Cache Agent: 定时同步策略和设备缓存 (2小时)                │
            │  • Collector: 收集并定期存储验证消息到Redis                  │
            └────────────────────────────────────────────────────────────┘
```

### 数据流向

#### 1. 告警查询流程
```
客户端请求 → Service API → Logic Layer → DAO Layer → MySQL数据库
                                              ↓
                                        本地缓存增强
                                              ↓
                                        返回响应数据
```

#### 2. 策略验证记录处理流程
```
Alarm-Compute服务
        ↓
  生成验证结果
        ↓
Kafka Topic: ruleValid
        ↓
  Kafka Consumer (批量消费50条)
        ↓
  解析并去重 (按设备+策略)
        ↓
  Collector收集器 (内存累积)
        ↓
  定期存储到Redis (每N秒)
        ↓
  定时上报生效率 (10-40秒)
        ↓
Kafka Topic: alarm_admin → 运维平台
```

#### 3. 缓存同步流程
```
启动/定时触发 (每2小时)
        ↓
  查询CMDB获取模组信息
        ↓
  对比本地缓存版本号
        ↓
  版本变化?
   ├─ Yes → 批量拉取设备/策略
   │         ↓
   │    存入本地缓存 (3天有效期)
   │         ↓
   │    更新版本号
   └─ No → 跳过本次同步
```

---

## 模块介绍

Alarm-Server 是一个基于 TRPC 框架构建的微服务，采用分层架构设计，包含以下核心模块：

### 1. **Service API Layer** (服务接口层)
- 位置: `service/api/`
- 职责: 对外提供 gRPC 接口，负责请求参数校验和路由分发
- 特点: 统一的入口控制，支持 MozuId 验证

### 2. **Logic Layer** (业务逻辑层)
- 位置: `logic/api/`
- 职责: 实现具体的业务逻辑，包括告警查询、策略验证、数据聚合等
- 子模块:
  - `alarm/`: 告警相关业务逻辑
  - `strategy/`: 策略相关业务逻辑

### 3. **Data Access Layer** (数据访问层)
- 位置: `repo/`
- 职责: 封装所有数据访问操作，包括数据库、缓存、RPC调用
- 子模块:
  - `dao/`: MySQL 数据库访问
  - `store/`: Redis 缓存操作
  - `rpc/`: 外部服务调用 (CMDB、Alarm-Compute等)
  - `ckafka/`: Kafka 消息生产

### 4. **Cache Management** (缓存管理)
- 位置: `logic/cache/`
- 职责: 管理本地内存缓存和同步逻辑
- 缓存内容:
  - 策略缓存: 按 `MozuId -> Rid -> Gid` 三级索引
  - 设备缓存: 按 `Gid` 索引设备详细信息
  - 版本控制: 通过版本号实现增量更新

### 5. **Collector Module** (收集器模块)
- 位置: `logic/collector/`
- 职责: 收集策略验证记录，计算生效率，上报运维平台
- 功能:
  - 实时收集: 接收 Kafka 消息并累积到内存
  - 定期存储: 批量写入 Redis
  - 生效率上报: 定时计算并发送到 Kafka
  - 过期清理: 删除超过1小时未更新的记录

### 6. **Consumer Module** (消费者模块)
- 位置: `logic/consumer/`
- 职责: 消费 Kafka 验证消息，进行解析和去重
- 特点: 批量消费 (50条/批次)，按设备去重保留最新记录

---

## 核心能力

### 1. 告警数据查询

提供多维度的告警数据查询能力，支持活跃告警和历史告警查询。

#### 1.1 告警列表查询 (`GetAlarmList`)
- **功能**: 分页查询告警列表，支持多维度过滤
- **支持过滤条件**:
  - 时间范围: `OccurTimeStart`, `OccurTimeEnd`
  - 告警级别: `Level` (L0-L4)
  - 设备维度: `DeviceGid`, `DeviceNumber`, `DeviceTypeZh`
  - 告警属性: `AlarmName`, `Content`, `Fingerprint`
  - 状态: `Status` (0:活跃, 1:已关闭), `EventStatus` (0:未挂单, 1:已挂单)
- **数据来源**: 
  - 活跃告警: `t_alarm_active`
  - 历史告警: `t_alarm_history`
- **性能优化**: 
  - 使用只读数据库副本
  - 设备信息通过本地缓存补充
  - 支持自定义排序 (时间/级别)

#### 1.2 告警数量统计 (`GetAlarmCnt`)
- **功能**: 统计告警数量，支持多维度聚合
- **统计维度**:
  - 按告警级别: `level`
  - 按设备类型: `device_type_zh`
  - 按设备: `device_gid`
  - 按告警名称: `alarm_name`
  - 按指纹: `fingerprint`
- **返回数据**: 
  - 总数量: `TotalCount`
  - 分组统计: `Metrics` (键值对形式)

#### 1.3 告警趋势查询 (`GetAlarmCntTrend`)
- **功能**: 查询最近24小时告警数量趋势
- **时间粒度**: 按小时聚合
- **返回数据**: 24个时间点的告警数量数组

#### 1.4 告警诊断 (`AlarmDiagnose`)
- **功能**: 诊断指定策略在特定时间段的执行情况
- **实现方式**: 调用 Alarm-Compute 服务重新计算表达式
- **返回信息**: 
  - 表达式计算结果
  - 数据点值
  - 错误信息

### 2. 策略管理

#### 2.1 策略查询 (`GetStrategy`)
- **功能**: 查询告警策略配置
- **支持过滤**: MozuId, Rid, Gid, DeviceNumber, AlarmName, Level等
- **数据来源**: `t_alarm_strategy` 表
- **返回信息**: 策略详情 (表达式、阈值、告警内容模板等)

#### 2.2 策略实例查询 (`GetStrategyInstance`)
- **功能**: 查询策略的实例化列表
- **数据来源**: 
  - 策略基础信息: MySQL数据库
  - 设备信息: 本地缓存
- **应用场景**: 运维平台展示策略应用到哪些设备

#### 2.3 虚拟测点查询 (`GetVirtualPoint`)
- **功能**: 查询虚拟测点定义
- **数据来源**: 策略表达式中的虚拟测点映射

### 3. 策略验证与生效率监控

这是 Alarm-Server 的核心特色功能，实现对告警策略执行情况的全链路监控。

#### 3.1 验证记录收集
- **数据源**: Kafka Topic `trpc.kafka.tbos.ruleValid`
- **生产者**: Alarm-Compute 服务 (每次策略计算后发送验证结果)
- **消费方式**: 批量消费 (50条/批次，500ms刷新)
- **记录内容**:
  ```go
  type ValidStoreData struct {
      MozuId      int32   // 模组ID
      Rid         int64   // 策略ID
      Gid         string  // 设备GID
      AlarmLevel  string  // 告警级别
      EvalTime    int64   // 计算时间戳
      Success     bool    // 是否执行成功
      Fired       bool    // 是否触发告警
      ErrorCode   int32   // 错误码
      ErrorName   string  // 错误名称
      ErrorDetail string  // 错误详情
  }
  ```

#### 3.2 验证记录存储
- **存储方式**: Redis Hash结构
- **Key格式**: `v_[mozuId]:[rid]`
- **Field**: `[gid]`
- **Value**: JSON序列化的 `ValidStoreData`
- **TTL**: 60秒 (自动过期)
- **存储频率**: 可配置 (默认每N秒批量存储)

#### 3.3 生效率计算与上报
- **触发方式**: 定时任务 (10-40秒间隔)
- **计算逻辑**:
  1. 获取分布式锁 (基于Redis，防止多实例重复计算)
  2. 遍历本地策略缓存
  3. 批量从Redis查询验证记录
  4. 统计每个模组的:
     - 有效策略数 (`valid_num`): 成功执行的策略数
     - 失败策略数 (`failed_num`): 执行失败的策略数
     - 总策略数 (`strategy_num`): 应该执行的策略总数
     - 生效率 (`efficiency`): `valid_num / strategy_num`
- **上报目标**: Kafka Topic `trpc.kafka.tbos.alarm_admin`
- **批量大小**: 可配置 (默认1000条/批次)

#### 3.4 验证记录查询 (`GetValidate`)
- **功能**: 查询策略的验证记录列表
- **数据来源**: Redis缓存
- **过滤条件**: MozuId, Rid, Gid等
- **特殊处理**: 
  - 如果记录不存在或已过期，返回 "策略未上报" 状态
  - 错误详情中将测点符号转换为实际测点名称

#### 3.5 过期记录清理
- **触发方式**: 定时任务 (每天12:40执行)
- **清理规则**: 删除超过1小时未更新的验证记录
- **实现步骤**:
  1. 获取分布式锁
  2. 遍历所有策略键
  3. 批量查询Redis记录
  4. 识别过期记录 (当前时间 - 评估时间 > 1小时)
  5. 批量删除

### 4. 告警状态管理

#### 4.1 更新告警状态 (`UpdateAlarmStatus`)
- **功能**: 更新告警的挂单状态
- **操作类型**:
  - 挂单: `EventStatus = 1`
  - 取消挂单: `EventStatus = 0`
- **更新范围**: 活跃告警表 `t_alarm_active`

#### 4.2 删除历史告警 (`DelHistoryAlarm`)
- **功能**: 批量删除历史告警记录
- **安全保护**: 需要提供有效的 Token (从配置读取)
- **删除条件**: 支持按 MozuId, 设备, 时间范围等条件删除
- **数据表**: `t_alarm_history`

### 5. 本地缓存管理

#### 5.1 三级缓存架构
```
Level 1: 本地内存缓存 (LocalCache)
  ├─ 策略缓存: map[mozuId]map[rid]map[gid]StrategyCacheData
  ├─ 设备缓存: map[gid]DeviceEntity
  └─ 版本缓存: 用于增量更新判断

Level 2: Redis缓存
  └─ 验证记录: 60秒TTL

Level 3: MySQL数据库
  └─ 持久化数据
```

#### 5.2 缓存同步机制
- **同步触发**:
  - 启动时立即同步
  - 定时同步 (默认每2小时)
- **同步策略**:
  1. 从CMDB查询所有模组信息和版本号
  2. 对比本地缓存的版本号
  3. 如果版本变化，触发增量更新:
     - 策略同步: 批量查询数据库 (默认30000条/批次)
     - 设备同步: 批量调用CMDB接口 (默认8000条/批次)
  4. 更新本地缓存并记录新版本号
- **缓存过期**: 本地缓存 TTL 为 3天

#### 5.3 缓存数据结构
```go
type StrategyCacheData struct {
    ID                   int64      // 策略ID
    DeviceGid            string     // 设备GID
    Rid                  int64      // 规则ID
    RidVersion           string     // 规则版本
    RidType              int32      // 规则类型 (0:实时, 1:延时)
    MozuId               int32      // 模组ID
    AlarmName            string     // 告警名称
    AlarmExpression      string     // 告警表达式
    AlarmExpressionStr   string     // 表达式字符串
    RestoreExpression    string     // 恢复表达式
    ExpressionMap        *ExprMap   // 符号到测点映射
    AlarmLevel           string     // 告警级别
    ContentTemplate      string     // 内容模板
    Owner                string     // 责任人
    CreateAt, UpdateAt   time.Time  // 创建/更新时间
}
```

---

## 代码结构

```
alarm-server/
├── main.go                           # 服务入口: 初始化服务、注册goroutine和定时任务
├── trpc_go.yaml                      # TRPC框架配置: 服务端口、数据库连接、Kafka配置等
├── go.mod / go.sum                   # Go模块依赖管理
│
├── conf/                             # 配置管理
│   └── conf.go                       # 业务配置结构定义 (验证上报、缓存同步等)
│
├── entity/                           # 实体定义
│   ├── constant/
│   │   └── time.go                   # 时间常量定义
│   ├── errcode/
│   │   ├── errcode.go                # 错误码定义
│   │   └── taskcode/
│   │       └── taskcode.go           # 任务错误码
│   └── model/
│       └── persist.go                # 数据模型 (StrategyCacheData, ValidStoreData)
│
├── service/                          # 服务层
│   └── api/
│       └── api.go                    # gRPC接口实现 (11个核心API)
│
├── logic/                            # 业务逻辑层
│   ├── api/                          # API业务逻辑
│   │   ├── alarm/                    # 告警相关逻辑
│   │   │   ├── api.go                # 接口定义
│   │   │   ├── alarm_list.go         # 告警列表查询实现
│   │   │   ├── alarm_cnt.go          # 告警数量统计实现
│   │   │   ├── alarm_diagnose.go     # 告警诊断实现
│   │   │   └── foc.go                # 过滤和聚合逻辑
│   │   └── strategy/                 # 策略相关逻辑
│   │       ├── api.go                # 接口定义
│   │       ├── strategy.go           # 策略查询实现
│   │       ├── validate.go           # 验证记录查询实现
│   │       └── virtualpoint.go       # 虚拟测点查询实现
│   │
│   ├── cache/                        # 缓存管理
│   │   ├── cache.go                  # 本地缓存操作 (策略、设备)
│   │   ├── cache_agent.go            # 缓存同步代理 (定时同步任务)
│   │   └── strategy.go               # 策略缓存同步逻辑
│   │
│   ├── consumer/                     # Kafka消费者
│   │   └── consumer.go               # 批量消费验证消息, 解析并去重
│   │
│   └── collector/                    # 验证记录收集器
│       ├── collector.go              # 收集器核心逻辑 (实时/延时两个收集器)
│       ├── report.go                 # 生效率上报 (定时任务)
│       └── expire.go                 # 过期记录清理 (定时任务)
│
├── repo/                             # 数据访问层
│   ├── dao/                          # 数据访问对象
│   │   ├── alarm/                    # 告警DAO
│   │   │   ├── api.go                # 告警数据库操作 (查询、更新、删除)
│   │   │   └── con.go                # 查询条件构建器
│   │   └── strategy/                 # 策略DAO
│   │       ├── api.go                # 策略数据库操作
│   │       └── con.go                # 查询条件构建器
│   │
│   ├── rpc/                          # RPC调用封装
│   │   ├── cmdb.go                   # CMDB服务调用 (获取设备、模组信息)
│   │   └── alarm_compute.go          # Alarm-Compute服务调用 (表达式计算)
│   │
│   ├── store/                        # Redis存储
│   │   └── redis.go                  # Redis操作 (批量读写、分布式锁)
│   │
│   └── ckafka/                       # Kafka生产者
│       └── ckafka.go                 # 批量发送验证记录到运维平台
│
└── utils/                            # 工具类
    ├── common/
    │   └── order_map.go              # 有序Map实现
    └── modcall/
        └── modcall.go                # 监控打点工具
```

### 核心文件说明

| 文件路径 | 代码行数 | 核心职责 |
|---------|---------|---------|
| `main.go` | 40 | 服务启动入口，注册后台goroutine和定时任务 |
| `service/api/api.go` | 126 | 11个gRPC接口的入口实现，参数校验和路由 |
| `logic/cache/cache.go` | 161 | 本地缓存的读写操作，策略和设备缓存管理 |
| `logic/cache/cache_agent.go` | - | 定时同步缓存，版本比对和增量更新 |
| `logic/consumer/consumer.go` | 80 | Kafka批量消费验证消息，解析protobuf并去重 |
| `logic/collector/collector.go` | 114 | 验证记录收集器，内存累积待上报数据 |
| `logic/collector/report.go` | 87 | 生效率计算和上报逻辑 |
| `logic/collector/expire.go` | 93 | 过期验证记录清理逻辑 |
| `repo/dao/alarm/api.go` | 483 | 告警表的CRUD操作，支持复杂条件查询和聚合 |
| `repo/dao/strategy/api.go` | - | 策略表的查询操作 |
| `repo/store/redis.go` | 249 | Redis批量操作、分布式锁实现 |
| `repo/rpc/cmdb.go` | 60 | CMDB服务调用封装 |
| `repo/ckafka/ckafka.go` | - | Kafka消息批量发送 |

---

## 工作流程

### 1. 服务启动流程

```
main.go 启动
    ↓
创建 TRPC Server
    ↓
注册 Background Services:
    ├─ cache.RegularSyncCache()          // 定时缓存同步 (goroutine)
    ├─ collector.RegularStoreValidMsg()  // 定期存储验证消息 (goroutine)
    └─ kafka.RegisterBatchHandlerService() // Kafka批量消费 (自动启动)
    ↓
注册 Timer Services:
    ├─ trpc.timer.tbos.alarmValid        // 生效率上报 (10-40秒)
    └─ trpc.timer.tbos.delRuleRecord     // 过期记录清理 (每天12:40)
    ↓
注册 gRPC Services:
    └─ pb.RegisterAlarmServerService()   // 注册11个API接口
    ↓
启动服务监听 (HTTP + gRPC)
    ↓
等待所有 goroutine 结束 (优雅关闭)
```

### 2. 告警查询流程 (以 GetAlarmList 为例)

```
1. 客户端发起 gRPC 请求
    ↓
2. service/api/api.go::GetAlarmList
    - 校验 MozuId 非空
    ↓
3. logic/api/alarm/alarm_list.go
    - 构建查询过滤条件 (ActiveAlarmFilter)
    - 决定查询 t_alarm_active 或 t_alarm_history
    ↓
4. repo/dao/alarm/api.go
    - 执行 SQL 查询 (支持分页、排序、多维度过滤)
    - 使用 GORM 构建复杂查询
    ↓
5. MySQL 数据库
    - 返回告警记录列表
    ↓
6. 数据增强 (logic/api/alarm/alarm_list.go)
    - 从本地缓存获取设备详细信息
    - 补充设备名称、类型、位置等
    ↓
7. 统计信息计算 (如果需要)
    - 按指定维度 (level, device_type_zh等) 聚合
    - 计算每个分组的数量
    ↓
8. 返回响应
    - RspAlarmList { TotalCount, AlarmList, Metrics }
```

### 3. 策略验证记录处理完整流程

```
[外部] Alarm-Compute 服务
    ↓
执行策略计算 (每个策略每次计算)
    ↓
生成验证结果 (ValidateTaskItem)
    ↓
发送到 Kafka Topic: trpc.kafka.tbos.ruleValid
    ↓
─────────────────────────────────────────
[alarm-server] Kafka Consumer (consumer.go)
    ↓
批量消费 (50条/批次, 500ms刷新)
    ↓
解析 Protobuf 消息
    ↓
按 RidType 分类:
    ├─ RidType = 0 → 实时策略 (RealTimeCollector)
    └─ RidType = 1 → 延时策略 (DelayCollector)
    ↓
去重逻辑 (按 Rid+Gid 分组，保留最新的 EvalTime)
    ↓
传递给 Collector 模块
    ↓
─────────────────────────────────────────
[Collector] collector.go
    ↓
存入内存 Map:
    - key: "[rid]:[gid]"
    - value: ValidStoreData
    ↓
定期触发 (RegularStoreValidMsg, 每N秒)
    ↓
批量写入 Redis (store/redis.go)
    - Key: "v_[mozuId]:[rid]"
    - Field: "[gid]"
    - Value: JSON(ValidStoreData)
    - TTL: 60秒
    ↓
清空内存 Map
    ↓
─────────────────────────────────────────
[Timer Task] report.go (每10-40秒执行)
    ↓
1. 尝试获取分布式锁 (Redis)
    - 防止多实例重复计算
    ↓
2. 遍历本地策略缓存
    - 按 MozuId 分组
    ↓
3. 批量查询 Redis 验证记录
    - 使用 MGET 批量获取
    ↓
4. 计算生效率指标:
    for each mozu:
        valid_num = 成功执行的策略数
        failed_num = 失败的策略数
        strategy_num = 应执行的总策略数
        efficiency = valid_num / strategy_num
    ↓
5. 构造上报消息 (ValidCheck)
    ↓
6. 批量发送到 Kafka (ckafka/ckafka.go)
    - Topic: trpc.kafka.tbos.alarm_admin
    - 批量大小: 1000条/批次
    ↓
7. 释放分布式锁
    ↓
─────────────────────────────────────────
[运维平台] 接收生效率数据
    ↓
展示告警策略健康度
```

### 4. 缓存同步流程详解

```
[启动时 / 定时触发 (每2小时)]
    ↓
cache_agent.go::RegularSyncCache()
    ↓
──────────── 设备缓存同步 ────────────
1. 调用 CMDB 获取所有模组信息
    - rpc/cmdb.go::GetMozuInfoList()
    - 返回: [{MozuId, Name, AlarmVersion, PublishVersion}]
    ↓
2. 对比每个模组的 AlarmVersion
    - 从本地缓存读取已存储的版本号
    - NeedUpdateDeviceCache(mozuId, version)
    ↓
3. 如果版本变化 → 触发设备缓存更新
    - 批量查询 CMDB (默认8000条/批次)
    - rpc/cmdb.go::GetDeviceEntity(mozuId, page, size)
    - 返回设备信息: GID, 名称, 类型, 位置等
    ↓
4. 存入本地缓存
    - cache.go::SetDeviceCache(mozuId, devices)
    - 使用 LocalCache (TTL: 3天)
    ↓
──────────── 策略缓存同步 ────────────
5. 对比每个模组的 PublishVersion
    - NeedUpdateStrategy(mozuId, version)
    ↓
6. 如果版本变化 → 触发策略缓存更新
    - 批量查询 MySQL (默认30000条/批次)
    - dao/strategy/api.go::GetStrategyList()
    - 返回策略详情: Rid, 表达式, 级别, 设备等
    ↓
7. 转换为缓存数据结构
    - StrategyCacheData (包含解析后的表达式映射)
    ↓
8. 存入本地缓存 (三级索引)
    - cache.go::SetStrategyCache()
    - map[mozuId][rid][gid] = StrategyCacheData
    - 重试3次 (防止并发冲突)
    ↓
9. 更新版本号缓存
    ↓
[等待下一次同步触发]
```

### 5. 过期记录清理流程

```
[Timer Task] expire.go (每天12:40执行)
    ↓
1. 尝试获取分布式锁
    - 防止多实例重复执行
    ↓
2. 获取所有策略键
    - 遍历本地策略缓存
    - 生成 Redis key 列表: ["v_[mozuId]:[rid]", ...]
    ↓
3. 批量查询 Redis 验证记录
    - store/redis.go::BatchGetRuleRecord()
    - 使用 HGETALL 批量获取
    ↓
4. 识别过期记录
    - 当前时间 - record.EvalTime > 1小时?
    - 是 → 标记为待删除
    ↓
5. 批量删除过期记录
    - store/redis.go::BatchDelRuleRecord()
    - 使用 HDEL 批量删除 (默认300条/批次)
    ↓
6. 释放分布式锁
    ↓
记录日志: 删除数量
```

---

## 配置说明

### 1. 主配置文件 (trpc_go.yaml)

```yaml
server:
  app: tbos                           # 应用名称
  server: alarm-server                # 服务名称
  service:
    - name: trpc.tbos.alarm-server.AlarmServer
      ip: 0.0.0.0
      port: ${PORT_ALARM_SERVER}      # 服务端口 (环境变量)
      protocol: http                  # 协议类型
      timeout: 10000                  # 超时时间 (ms)

# Kafka 消费者配置
client:
  service:
    - name: trpc.kafka.tbos.ruleValid
      target: kafka://127.0.0.1:9092
      topics:
        - ${KAFKA_ALARM_VALID_TOPIC}  # 验证消息Topic (环境变量)
      consumer_group: alarm-server    # 消费者组
      batch: 50                       # 批量消费条数
      batch_flush_interval: 500       # 批量刷新间隔 (ms)

    # Timer 定时任务配置
    - name: trpc.timer.tbos.alarmValid
      schedule: 10-40/10 * * * * *    # 每10-40秒执行
      timeout: 2000                   # 超时时间 (ms)

    - name: trpc.timer.tbos.delRuleRecord
      schedule: 0 40 12 * * *         # 每天12:40执行
      timeout: 2000

# MySQL 数据库配置
plugins:
  database:
    mysql:
      - name: trpc.mysql.tbos.alarm
        dsn: ${MYSQL_ALARM_DSN}       # 主库DSN (环境变量)
        max_idle: 10
        max_open: 20

      - name: trpc.mysql.tbos.alarm_readonly
        dsn: ${MYSQL_ALARM_READONLY_DSN}  # 只读副本DSN
        max_idle: 10
        max_open: 20

    # Redis 配置
    redis:
      - name: trpc.redis.tbos.valid
        addrs: ${REDIS_ADDRS}         # Redis地址 (环境变量)
        password: ${REDIS_PASSWORD}
        db: 0
        max_retries: 3
```

### 2. 业务配置文件 (serverconf.yaml)

```yaml
# 策略验证配置
RuleValidConfig:
  RegularStoreInterval: 10            # 验证记录存储间隔 (秒)
  BatchSize: 1000                     # Kafka批量发送大小

# 缓存同步配置
SyncCacheConfig:
  StrategyCacheInterval: 7200         # 策略缓存同步间隔 (秒, 默认2小时)
  StrategyCacheBatchSize: 30000       # 策略批量查询大小
  StrategyTotalIntervalCnt: 12        # 总同步周期数 (每周期检查一次)
  DeviceCacheInterval: 7200           # 设备缓存同步间隔 (秒)
  DeviceCacheBatchSize: 8000          # 设备批量查询大小
  DeviceTotalIntervalCnt: 12          # 总同步周期数

# Redis 缓存配置
ValidRedisCacheConfig:
  MGetBatchSize: 500                  # MGET批量大小
  MSetBatchSize: 500                  # MSET批量大小
  MDelBatchSize: 300                  # MDEL批量大小

# 数据库管理配置
db_admin:
  del_token: "your-secure-token"      # 删除历史告警的安全令牌
```

### 3. 环境变量说明

| 环境变量 | 说明 | 示例 |
|---------|------|------|
| `PORT_ALARM_SERVER` | 服务监听端口 | `8086` |
| `KAFKA_ALARM_VALID_TOPIC` | 验证消息Kafka Topic | `alarm_valid` |
| `MYSQL_ALARM_DSN` | MySQL主库连接字符串 | `user:pass@tcp(host:3306)/db?charset=utf8mb4&parseTime=True&loc=Local` |
| `MYSQL_ALARM_READONLY_DSN` | MySQL只读副本连接字符串 | `user:pass@tcp(ro-host:3306)/db?charset=utf8mb4&parseTime=True&loc=Local` |
| `REDIS_ADDRS` | Redis地址 | `127.0.0.1:6379` |
| `REDIS_PASSWORD` | Redis密码 | `your-password` |

### 4. 数据库表结构

#### t_alarm_active (活跃告警表)
```sql
-- 关键字段
mozu_id         INT         -- 模组ID
alarm_id        VARCHAR     -- 告警唯一ID
rid             BIGINT      -- 策略ID
device_gid      VARCHAR     -- 设备GID
level           VARCHAR     -- 告警级别 (L0-L4)
alarm_name      VARCHAR     -- 告警名称
content         TEXT        -- 告警内容
occur_time      BIGINT      -- 发生时间戳
status          INT         -- 状态 (0:活跃, 1:已关闭)
event_status    INT         -- 挂单状态 (0:未挂单, 1:已挂单)
fingerprint     VARCHAR     -- 告警指纹

-- 索引建议
INDEX idx_mozu_device (mozu_id, device_gid)
INDEX idx_occur_time (occur_time)
INDEX idx_level (level)
INDEX idx_alarm_name (alarm_name)
```

#### t_alarm_history (历史告警表)
```sql
-- 字段同 t_alarm_active
-- 用于存储已关闭的告警
```

#### t_alarm_strategy (策略表)
```sql
-- 关键字段
id                      BIGINT      -- 策略ID
mozu_id                 INT         -- 模组ID
device_gid              VARCHAR     -- 设备GID
rid                     BIGINT      -- 规则ID
rid_version             VARCHAR     -- 规则版本
rid_type                INT         -- 规则类型 (0:实时, 1:延时)
alarm_name              VARCHAR     -- 告警名称
alarm_expression        TEXT        -- 告警表达式
restore_expression      TEXT        -- 恢复表达式
expression_map          TEXT        -- 表达式映射 (JSON)
alarm_level             VARCHAR     -- 告警级别
content_template        TEXT        -- 内容模板
owner                   VARCHAR     -- 责任人

-- 索引建议
INDEX idx_mozu_rid (mozu_id, rid)
INDEX idx_device_gid (device_gid)
INDEX idx_alarm_name (alarm_name)
```

---

## 常见问题

### 1. 性能优化相关

#### Q1.1: 告警查询速度慢怎么办?
**原因分析**:
- 数据量过大 (历史告警表记录数过多)
- 缺少合适的索引
- 查询条件未充分利用索引
- 未使用只读数据库副本

**解决方案**:
1. **数据库优化**:
   ```sql
   -- 创建复合索引
   CREATE INDEX idx_mozu_occur ON t_alarm_active(mozu_id, occur_time);
   CREATE INDEX idx_device_level ON t_alarm_active(device_gid, level);
   ```

2. **分表策略**:
   - 定期归档历史数据 (超过3个月的告警移至归档表)
   - 使用 `DelHistoryAlarm` API 删除过期数据

3. **查询优化**:
   - 始终指定时间范围 (避免全表扫描)
   - 使用分页查询 (PageNo, PageSize)
   - 限制返回字段 (不需要的字段不查询)

4. **缓存优化**:
   - 设备信息已缓存在本地，无需重复查询
   - 对于高频查询，考虑在应用层增加缓存

#### Q1.2: 缓存同步占用资源过高?
**原因分析**:
- 批量查询数据库时数据量过大
- 同步频率过高
- 网络带宽不足

**解决方案**:
1. **调整配置 (serverconf.yaml)**:
   ```yaml
   SyncCacheConfig:
     StrategyCacheInterval: 10800      # 增加到3小时
     StrategyCacheBatchSize: 10000     # 减小批量大小
     DeviceCacheBatchSize: 5000        # 减小批量大小
   ```

2. **优化同步逻辑**:
   - 利用版本号机制，只在数据变化时同步
   - 错峰同步 (避免高峰期)

3. **资源隔离**:
   - 使用独立的数据库只读副本
   - 增加数据库连接池大小

### 2. 数据一致性问题

#### Q2.1: 验证记录查询不到或显示"策略未上报"?
**原因分析**:
- Alarm-Compute 服务未正常运行
- Kafka 消息丢失或延迟
- Redis 记录已过期 (TTL 60秒)
- 策略未绑定到设备

**排查步骤**:
1. **检查 Kafka 消费者状态**:
   ```bash
   # 查看消费者 Lag (未消费消息数)
   kafka-consumer-groups --bootstrap-server <kafka> \
     --group alarm-server --describe
   ```

2. **检查 Redis 缓存**:
   ```bash
   # 查看验证记录
   redis-cli HGETALL "v_<mozuId>:<rid>"
   ```

3. **检查 Alarm-Compute 服务**:
   - 确认服务正常运行
   - 查看日志是否有策略计算错误

4. **检查策略配置**:
   - 确认策略已启用
   - 确认策略绑定到正确的设备

**解决方案**:
- 如果是 Redis 过期: 等待下次策略计算 (通常1分钟内)
- 如果是 Kafka 延迟: 检查消费者性能，考虑增加分区
- 如果是策略问题: 在运维平台检查策略配置

#### Q2.2: 本地缓存数据不一致?
**原因分析**:
- 缓存同步失败
- 版本号未正确更新
- 并发写入冲突

**排查步骤**:
1. **查看日志**:
   ```bash
   # 搜索缓存同步相关日志
   grep "cache sync" alarm-server.log
   grep "NeedUpdate" alarm-server.log
   ```

2. **手动触发同步**:
   - 重启服务 (启动时会立即同步)
   - 等待下一个同步周期 (默认2小时)

**解决方案**:
- 确保 CMDB 服务可访问
- 检查数据库连接是否正常
- 增加缓存 TTL (避免频繁过期)

### 3. 功能使用问题

#### Q3.1: 如何删除历史告警?
**使用方式**:
1. 获取删除令牌 (从配置文件 `db_admin.del_token`)
2. 调用 `DelHistoryAlarm` API:
   ```protobuf
   req := &pb.ReqDelHistoryAlarm{
       MozuId:    123,
       Token:     "your-secure-token",
       StartTime: 1609459200,  // 2021-01-01 00:00:00
       EndTime:   1640995200,  // 2022-01-01 00:00:00
       DeviceGid: "device-gid",  // 可选
       Level:     "L1",          // 可选
   }
   ```

**注意事项**:
- 删除操作不可逆，请谨慎操作
- 建议先备份数据
- 支持按时间范围、设备、级别等条件删除
- Token 验证失败会拒绝删除

#### Q3.2: 如何查看策略生效率?
**方式一: 调用 API 查询**
```protobuf
req := &pb.ReqValidateList{
    MozuId: 123,
    Rid:    456,     // 可选: 指定策略
    Gid:    "gid",   // 可选: 指定设备
}
rsp := GetValidate(ctx, req)
// 返回每个设备的验证记录
```

**方式二: 查看 Kafka 上报数据**
- Topic: `trpc.kafka.tbos.alarm_admin`
- 每10-40秒上报一次
- 包含: 有效数、失败数、总数、生效率

**方式三: 运维平台查看**
- 运维平台会消费上报数据并展示图表

#### Q3.3: 如何诊断告警策略?
**使用 AlarmDiagnose API**:
```protobuf
req := &pb.ReqAlarmDiagnose{
    MozuId:    123,
    Rid:       456,
    Gid:       "device-gid",
    StartTime: 1640000000,
    EndTime:   1640001000,
    Interval:  60,  // 采样间隔 (秒)
}
rsp := AlarmDiagnose(ctx, req)
```

**返回信息**:
- 表达式计算结果 (每个时间点)
- 数据点值
- 是否触发告警
- 错误信息 (如果有)

**应用场景**:
- 排查策略为什么没有触发告警
- 验证表达式逻辑是否正确
- 分析数据点值是否异常

### 4. 运维部署问题

#### Q4.1: 服务启动失败?
**常见原因及解决**:

1. **端口被占用**:
   ```bash
   # 检查端口占用
   lsof -i :8086
   # 修改配置文件中的 PORT_ALARM_SERVER
   ```

2. **数据库连接失败**:
   - 检查 DSN 配置是否正确
   - 确认数据库可访问 (网络、防火墙)
   - 验证数据库用户权限

3. **Kafka 连接失败**:
   - 检查 Kafka 地址配置
   - 确认 Topic 已创建
   - 验证消费者组权限

4. **Redis 连接失败**:
   - 检查 Redis 地址和密码
   - 确认 Redis 版本兼容 (建议 >= 6.0)

#### Q4.2: 如何监控服务健康状态?
**关键监控指标**:

1. **服务可用性**:
   - API 响应时间
   - 错误率
   - QPS (每秒查询数)

2. **Kafka 消费**:
   - Consumer Lag (消费延迟)
   - 消费速率
   - 消费错误数

3. **缓存命中率**:
   - 本地缓存命中次数
   - Redis 查询次数
   - 缓存过期次数

4. **数据库性能**:
   - 慢查询数量
   - 连接池使用率
   - 死锁/锁等待

5. **业务指标**:
   - 告警总数
   - 策略生效率
   - 验证记录上报数

**监控工具**:
- 使用 `utils/modcall/modcall.go` 中的打点函数
- 集成 Prometheus + Grafana
- 配置告警规则

#### Q4.3: 如何进行灰度发布?
**推荐步骤**:

1. **部署新版本实例 (不加入负载均衡)**
2. **观察日志和监控指标**:
   - 启动是否成功
   - 缓存同步是否正常
   - Kafka 消费是否正常
3. **小流量测试**:
   - 手动调用 API 接口测试
   - 验证返回数据正确性
4. **加入负载均衡 (小比例流量)**
5. **逐步增加流量比例**
6. **全量切换并下线旧版本**

**注意事项**:
- 新旧版本并存期间，定时任务可能重复执行 (已有分布式锁保护)
- 缓存数据在多实例间不共享 (各自维护本地缓存)
- Kafka 消费者使用同一个 Group (消息只会被一个实例消费)

#### Q4.4: 如何优雅重启服务?
**步骤**:

1. **准备阶段**:
   - 确认无重要任务正在执行
   - 备份当前日志和配置

2. **下线流量**:
   - 从负载均衡摘除实例
   - 等待现有请求处理完毕 (建议等待30秒)

3. **发送停止信号**:
   ```bash
   # 发送 SIGTERM 信号 (优雅关闭)
   kill -TERM <pid>
   ```

4. **等待 goroutine 结束**:
   - main.go 中使用 WaitGroup 等待后台任务
   - 超时时间: 30秒

5. **启动新实例**

6. **加入负载均衡**

**优雅关闭机制**:
```go
// main.go 中的实现
s.RegisterOnShutdown(func() {
    cancel()  // 取消 context
})
wg.Wait()  // 等待所有 goroutine
```

### 5. 数据问题排查

#### Q5.1: 告警数量统计不准确?
**可能原因**:
1. **查询条件问题**:
   - 时间范围未正确设置
   - 过滤条件冲突

2. **活跃/历史表分布**:
   - 部分告警已转移到历史表
   - 需要同时查询两个表

3. **数据延迟**:
   - Alarm-Compute 还未生成告警
   - 数据库主从同步延迟

**排查方法**:
```go
// 查询活跃告警
reqActive := &pb.ReqAlarmCnt{
    MozuId: 123,
    Status: 0,  // 活跃
    ...
}

// 查询历史告警
reqHistory := &pb.ReqAlarmCnt{
    MozuId: 123,
    Status: 1,  // 已关闭
    ...
}

// 总数 = 活跃 + 历史
```

#### Q5.2: 策略实例数量与预期不符?
**排查步骤**:
1. **检查策略配置**:
   - 查看策略的 `ApplyType` (应用类型)
   - 确认设备过滤条件

2. **检查设备缓存**:
   - 设备信息可能未同步
   - 手动触发缓存刷新

3. **检查数据库数据**:
   ```sql
   SELECT COUNT(*) FROM t_alarm_strategy 
   WHERE mozu_id = ? AND rid = ?;
   ```

---

## 附录

### A. API 接口清单

| API 方法 | 功能 | 主要参数 |
|---------|------|---------|
| `GetAlarmList` | 查询告警列表 | MozuId, Status, 时间范围, 过滤条件, 分页 |
| `GetAlarmCnt` | 统计告警数量 | MozuId, Status, 过滤条件, 聚合维度 |
| `GetAlarmCntTrend` | 24小时趋势 | MozuId, Status, 过滤条件 |
| `GetAlarmName` | 查询告警名称 | MozuId, 过滤条件 |
| `GetStrategy` | 查询策略列表 | MozuId, Rid, 设备, 分页 |
| `GetStrategyInstance` | 查询策略实例 | MozuId, Rid |
| `GetValidate` | 查询验证记录 | MozuId, Rid, Gid, 分页 |
| `UpdateAlarmStatus` | 更新告警状态 | MozuId, AlarmId, EventStatus |
| `GetVirtualPoint` | 查询虚拟测点 | MozuId, Rid |
| `DelHistoryAlarm` | 删除历史告警 | Token, MozuId, 时间范围, 过滤条件 |
| `AlarmDiagnose` | 告警诊断 | MozuId, Rid, Gid, 时间范围, 间隔 |

### B. 错误码说明

错误码定义在 `entity/errcode/` 目录下，主要包括:

- **通用错误** (errcode.go):
  - 参数错误
  - 数据库错误
  - RPC 调用错误
  - 缓存错误

- **任务错误** (taskcode/taskcode.go):
  - 策略计算错误
  - 数据点读取错误
  - 表达式解析错误

### C. 依赖服务

| 服务名称 | 用途 | 调用方式 |
|---------|------|---------|
| CMDB | 获取设备、模组信息 | HTTP RPC |
| Alarm-Compute | 策略计算、表达式评估 | HTTP RPC |
| Kafka | 消息队列 (验证记录、上报) | Consumer/Producer |
| MySQL | 告警、策略数据持久化 | GORM |
| Redis | 验证记录缓存、分布式锁 | go-redis |

### D. 相关文档

- [TRPC-Go 框架文档](https://trpc.group/trpc-go/trpc-go)
- [Alarm-Compute 服务文档](./alarm-compute.md)
- [CMDB](../config/cmdb.md)
