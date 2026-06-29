# 告警管理服务 Alarm-Manage

Alarm-Manage 是 TBOS 告警管理服务,负责接收 alarm-compute 发送的告警消息,并完成告警的消费、存储、生命周期管理和通知推送。

## 模块介绍

alarm-manage 是 TBOS 告警系统的管理中枢,承担以下职责:

1. **告警消费**: 从 Kafka 消费 alarm-compute 产生的告警消息
2. **告警处理**: 对告警进行去重、字段填充、ID生成和存储
3. **恢复管理**: 自动恢复已消除的告警,将活动告警转移到历史表
4. **通知推送**: 将告警推送到 CGI 同步 Kafka 供下游消费
5. **分布式ID生成**: 基于雪花算法生成全局唯一的告警ID
6. **缓存同步**: 定期同步设备信息缓存,减少外部依赖

---

## 系统整体架构

### 架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          TBOS 告警管理系统架构                            │
└─────────────────────────────────────────────────────────────────────────┘

                          ┌──────────────────┐
                          │  alarm-compute   │
                          │   (告警计算)      │
                          └────────┬─────────┘
                                   │ 产生告警/恢复消息
                                   ▼
                          ┌──────────────────┐
                          │      Kafka       │
                          │  (消息队列)       │
                          │  Topic: alarm    │
                          └────────┬─────────┘
                                   │
                                   ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                          alarm-manage 服务                                │
│                                                                           │
│  ┌─────────────┐         ┌─────────────┐         ┌─────────────┐       │
│  │   Consumer  │────────▶│   Manager   │────────▶│ Notification│       │
│  │  (Kafka消费) │         │  (告警管理)  │         │   (通知)     │       │
│  └─────────────┘         └──────┬──────┘         └──────┬──────┘       │
│                                  │                       │               │
│         ┌────────────────────────┼───────────────────────┘               │
│         │                        │                                       │
│         ▼                        ▼                                       │
│  ┌─────────────┐         ┌─────────────┐                                │
│  │ Alert Mgr   │         │ Restore Mgr │                                │
│  │ (告警处理)   │         │  (恢复处理)  │                                │
│  └──────┬──────┘         └──────┬──────┘                                │
│         │                       │                                        │
│         │  ┌──────────────────┐ │                                        │
│         └─▶│ Batch Channel    │◀┘                                        │
│            │  (批处理通道)     │                                          │
│            └────────┬─────────┘                                          │
│                     │                                                     │
│                     ▼                                                     │
│            ┌──────────────────┐                                          │
│            │   Worker Pool    │                                          │
│            │   (线程池)        │                                          │
│            └────────┬─────────┘                                          │
│                     │                                                     │
│         ┌───────────┼───────────┐                                        │
│         │           │           │                                        │
│         ▼           ▼           ▼                                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐                                │
│  │ 去重检查  │ │ 填充字段  │ │ 生成ID   │                                │
│  └──────────┘ │(设备信息) │ │(雪花算法)│                                │
│               └──────────┘ └──────────┘                                 │
│                     │                                                     │
│  ┌──────────────────┴──────────────────┐                                │
│  │                                      │                                │
│  ▼                                      ▼                                │
│ ┌──────────────┐              ┌─────────────────┐                       │
│ │ Snowflake    │              │   Local Cache   │                       │
│ │(分布式ID生成) │              │   (设备缓存)     │                       │
│ │              │              │                 │                       │
│ │ • Redis锁    │              │ • 定期同步       │                       │
│ │ • 心跳保活    │              │ • 版本控制       │                       │
│ └──────────────┘              └─────────────────┘                       │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
         │                                          │
         ▼                                          ▼
┌─────────────────┐                        ┌─────────────────┐
│  MySQL 数据库    │                        │  外部依赖服务     │
│                 │                        │                 │
│ • t_alarm_active│                        │ • Redis (锁)    │
│ • t_alarm_history│                       │ • CMDB (设备)   │
│ • t_alarm_worker│                        │ • Kafka (通知)  │
└─────────────────┘                        └─────────────────┘

数据流向:
1. alarm-compute 产生告警 → Kafka
2. Consumer 消费 Kafka → Manager 分发到告警/恢复通道
3. Batch Channel 批量汇聚 → Worker Pool 并发处理
4. 处理完成 → 写入数据库 & 推送通知
```

### 告警生命周期

```
┌────────┐      ┌────────┐      ┌────────┐      ┌────────┐
│ 产生   │─────▶│ 活动   │─────▶│ 恢复   │─────▶│ 历史   │
│ Firing │      │ Active │      │Resolved│      │History │
└────────┘      └────┬───┘      └────────┘      └────────┘
                     │
                     ▼
                ┌────────┐
                │ 挂起   │
                │Suspend │
                └────┬───┘
                     │
                     ▼
                ┌────────┐
                │ 解挂   │
                │ Resume │
                └────────┘
```

---

## 核心能力

### 1. 告警消费与处理

- **Kafka消费**: 从指定 Topic 消费告警消息 (Protobuf 格式)
- **消息分流**: 根据 `EndAt` 字段自动识别告警或恢复消息
- **延迟检测**: 检测消息延迟,超过 3 分钟的消息将被丢弃
- **批量处理**: 使用 Batch Channel 将单条消息汇聚成批,提高处理效率
- **并发处理**: 使用 Worker Pool (基于 ants) 并发处理告警

**核心逻辑位置**: 
- `/logic/consumer/consumer.go:22` - `Handle()` 消费入口
- `/logic/manager/manager.go:68` - `AddAlertToCh()` 告警分发
- `/logic/manager/manager.go:97` - `AddRestoreToCh()` 恢复分发

### 2. 告警去重与指纹机制

**指纹生成规则**: `FingerPrint = "{Rid};{Gid}"`
- `Rid`: 规则ID (标识哪条告警规则)
- `Gid`: 设备全局ID (标识哪个设备)

**去重策略**:
1. 本批次内去重 (内存去重)
2. 数据库指纹查询去重 (SQL: `WHERE fingerprint IN (...)`)
3. 数据库唯一索引兜底 (GORM: `OnConflict.DoNothing`)

**核心逻辑位置**: `/logic/manager/manager.go:62` - `GeneFingerPrint()`

### 3. 字段填充与模板替换

**自动填充字段**:
- 设备信息: 设备名称、设备类型、机房、功能间等 (从缓存获取)
- 创建时间: 当前时间
- 事件状态: EventStatus = 1 (活动)

**模板变量替换**:
- 告警内容支持模板变量,格式: `{{变量名}}`
- 从 `AnalyzeResult` 中解析测点值替换变量
- 示例: `"温度{{A}}超标"` → `"温度35.2超标"`

**核心逻辑位置**: `/logic/manager/alert_manager.go` - `fillActivesAlert()`

### 4. 分布式ID生成 (雪花算法)

**ID结构** (共64位):
```
┌─────────────┬─────────────┬─────────────┬─────────────┐
│   41 bits   │   8 bits    │   8 bits    │   6 bits    │
│  Timestamp  │   SetID     │  SubnodeID  │  Sequence   │
│ (毫秒时间戳) │  (片区ID)    │ (Pod节点)    │  (序列号)    │
└─────────────┴─────────────┴─────────────┴─────────────┘
```

**分布式协调**:
1. 启动时清理过期 Worker (24小时未心跳)
2. 查询已占用的 SubnodeID 列表
3. 尝试使用 Redis 分布式锁抢占未占用的节点
4. 抢占成功后定期更新心跳 (每小时)
5. 服务退出时自动清理占用记录

**支持规模**:
- 256 个片区 (SetID)
- 每个片区 256 个 Pod (SubnodeID)
- 每毫秒 64 个ID (Sequence)
- 总计: 理论 QPS > 1600万/秒

**核心逻辑位置**:
- `/logic/snowflake/snowflake.go:40` - `Generate()` ID生成
- `/logic/snowflake/redis_snowflake.go:18` - `InitSnowflake()` 初始化

### 5. 告警恢复处理

**恢复流程**:
1. 根据指纹查询对应的活动告警
2. 时间校验: 恢复时间必须晚于触发时间 (防止消息乱序)
3. 数据转换: 活动告警转为历史告警
4. 事务操作:
   - 插入历史表 `t_alarm_history`
   - 删除活动表 `t_alarm_active` 对应记录
5. 推送恢复通知

**去重策略**: 相同指纹的恢复消息,取最早的恢复时间

**核心逻辑位置**: 
- `/logic/manager/restore_manager.go` - 恢复处理主逻辑
- `/repo/db/alarm_dao.go` - `RestoreAlerts()` 事务操作

### 6. 设备缓存同步

**同步策略**:
- **增量同步**: 每 2 小时检查版本号,有变化则同步
- **全量同步**: 每 12 次增量后执行 1 次全量加载
- **缓存过期**: 本地缓存 3 天过期

**同步内容**:
- 设备全局ID (DeviceGid)
- 设备名称 (DeviceName)
- 设备类型 (DeviceTypeZh)
- 模组ID (MozuId)
- 机房名称 (IdcArea)
- 功能间 (FuncRoom)

**核心逻辑位置**: 
- `/logic/lcache/c_agent.go:19` - `RegularSyncDevice()` 定期同步
- `/repo/cache/lcache.go` - 本地缓存实现

### 7. 通知推送

**推送渠道**:
1. **CGI同步Kafka** (主要): 推送到指定 Kafka Topic 供 CGI 消费
2. **企业微信机器人** (可选): 发送 Markdown 消息到企业微信群 (代码中已注释)

**推送时机**:
- 告警入库成功后异步推送
- 告警恢复成功后异步推送
- `ResendAlarm` 接口调用时批量推送

**核心逻辑位置**:
- `/logic/notification/notification.go:18` - `ReportAlert()` 推送告警
- `/logic/notification/notification.go:33` - `ReportRestore()` 推送恢复
- `/repo/rpc/ckafka.go` - Kafka 发送实现

---

## 代码结构

```
alarm-manage/
├── main.go                          # 程序入口,初始化和启动
│
├── conf/                            # 配置管理
│   └── conf.go                      # 配置结构体定义
│
├── entity/                          # 数据实体定义
│   ├── errcode/
│   │   └── errcode.go               # 错误码定义
│   ├── message/
│   │   ├── alert_result.go          # 告警计算结果消息
│   │   └── robot.go                 # 企业微信机器人消息
│   └── model/
│       └── alarm_worker.go          # 雪花算法 Worker 信息
│
├── logic/                           # 核心业务逻辑
│   ├── consumer/                    # Kafka 消费者
│   │   ├── consumer.go              # 消费入口和消息处理
│   │   └── consumer_test.go         # 单元测试
│   │
│   ├── manager/                     # 告警管理器
│   │   ├── manager.go               # 管理器主逻辑,通道分发
│   │   ├── alert_manager.go         # 告警处理流程
│   │   └── restore_manager.go       # 恢复处理流程
│   │
│   ├── snowflake/                   # 分布式ID生成
│   │   ├── snowflake.go             # 雪花算法核心实现
│   │   └── redis_snowflake.go       # Redis分布式协调
│   │
│   ├── lcache/                      # 本地缓存管理
│   │   └── c_agent.go               # 设备缓存同步Agent
│   │
│   └── notification/                # 通知推送
│       ├── notification.go          # 通知接口实现
│       └── robot/                   # 企业微信机器人
│           ├── notice.go            # 机器人消息发送
│           ├── crypt.go             # 消息加解密
│           └── callback.go          # 回调处理
│
├── service/                         # RPC服务接口
│   ├── alarm.go                     # 告警服务接口
│   │                                # - PushAlarm: 补发告警
│   │                                # - ResendAlarm: 重推全量
│   └── bot.go                       # 机器人回调接口
│
├── repo/                            # 数据访问层
│   ├── db/                          # 数据库操作
│   │   ├── alarm_dao.go             # 告警表CRUD
│   │   └── alarm_worker.go          # Worker表CRUD
│   │
│   ├── cache/                       # 本地缓存
│   │   └── lcache.go                # 缓存读写
│   │
│   └── rpc/                         # 外部RPC调用
│       ├── ckafka.go                # Kafka生产者
│       ├── credis.go                # Redis操作
│       ├── cmdb.go                  # CMDB服务调用
│       └── robot.go                 # 企业微信机器人API
│
├── utils/                           # 工具类
│   ├── batch/
│   │   └── batch_channel.go         # 批处理通道
│   ├── common/
│   │   └── json.go                  # JSON序列化工具
│   └── modcall/
│       └── modcall.go               # 监控指标上报
│
├── trpc_go.yaml                     # TRPC框架配置文件
└── README.md                        # 本文档
```

---

## 核心流程详解

### 1. 程序启动流程

**入口**: `main.go:21`

```
main()
├─ 1. 创建 TRPC Server
├─ 2. 初始化雪花算法 (抢占节点,失败则 panic)
├─ 3. 启动两个核心 goroutine:
│     ├─ Manager.Run() - 告警和恢复处理
│     └─ CacheAgent.RegularSyncDevice() - 设备缓存同步
├─ 4. 注册 RPC 服务:
│     ├─ ManageService (告警服务)
│     └─ KafkaConsumerService (Kafka消费)
└─ 5. 启动服务器并等待退出
```

### 2. 告警处理流程

**入口**: `consumer.go:22` → `manager.go:68` → `alert_manager.go`

```
步骤 1: 消费 Kafka 消息
├─ Consumer.Handle() 接收 Kafka 消息
├─ 反序列化 Protobuf 消息 (AlarmMsgPb)
├─ 判断 EndAt 字段:
│   ├─ EndAt = 0: 告警消息 → Manager.AddAlertToCh()
│   └─ EndAt > 0: 恢复消息 → Manager.AddRestoreToCh()
└─ 记录消费指标

步骤 2: 延迟检测
├─ 计算消息延迟 (当前时间 - OccurTime)
├─ 延迟 > 3分钟: 丢弃消息并记录日志
├─ 延迟 > 1分钟: 记录告警日志
└─ 延迟正常: 放入处理通道

步骤 3: 批量汇聚
├─ Batch Channel 汇聚消息
├─ 触发条件:
│   ├─ 达到批大小 (BatchChannelSize = 1000)
│   └─ 超过时间间隔 (BatchFetchIntervalMS = 50ms)
└─ 将批量消息发送到 Worker Pool

步骤 4: 并发处理 (Worker Pool)
├─ 去重检查:
│   ├─ 提取所有指纹
│   ├─ 查询数据库已存在的指纹
│   └─ 过滤出新告警
│
├─ 填充字段:
│   ├─ 从缓存获取设备信息 (DeviceName, IdcArea等)
│   ├─ 解析 AnalyzeResult 获取测点值
│   ├─ 替换告警内容模板变量 {{变量名}}
│   └─ 设置创建时间
│
├─ 生成告警ID:
│   ├─ 调用雪花算法 Generate()
│   └─ 生成全局唯一 64位 ID
│
└─ 写入数据库:
    ├─ GORM批量插入 (1000条/批)
    ├─ OnConflict.DoNothing 处理重复
    └─ 返回插入成功的记录

步骤 5: 推送通知 (异步)
├─ 序列化为 JSON
├─ 发送到 CGI 同步 Kafka
└─ (可选) 发送企业微信消息
```

**关键配置**:
- **BatchChannelSize**: 1000 (批大小)
- **BatchFetchIntervalMS**: 50ms (批处理间隔)
- **PoolSize**: 15 (线程池大小)
- **通道缓冲**: 20000 (alertingCh 容量)

### 3. 恢复处理流程

**入口**: `consumer.go:44` → `manager.go:97` → `restore_manager.go`

```
步骤 1: 消息接收
└─ Manager.AddRestoreToCh() 将恢复消息放入恢复通道

步骤 2: 批量汇聚
├─ Batch Channel 汇聚消息
└─ 批大小: 1000, 间隔: 200ms

步骤 3: 去重和查询
├─ 指纹去重: 相同指纹取最早恢复时间
├─ 查询活动告警: 根据指纹列表批量查询
└─ 匹配恢复消息与活动告警

步骤 4: 时间校验
├─ 验证: RestoreTime > OccurTime
├─ 通过: 生成历史告警记录
└─ 失败: 丢弃该恢复消息

步骤 5: 数据库事务
├─ 开启事务
├─ 插入历史表: t_alarm_history
├─ 删除活动表: t_alarm_active (WHERE alarm_id IN (...))
└─ 提交事务

步骤 6: 推送通知 (异步)
└─ 发送恢复消息到 CGI 同步 Kafka
```

**关键配置**:
- **BatchChannelSize**: 1000
- **BatchFetchIntervalMS**: 200ms
- **PoolSize**: 5
- **通道缓冲**: 5000 (restoringCh 容量)

### 4. 雪花算法初始化流程

**入口**: `snowflake/redis_snowflake.go:18`

```
步骤 1: 清理过期 Worker
└─ DELETE FROM t_alarm_worker WHERE heart_beat < NOW() - 24小时

步骤 2: 查询已占用节点
├─ SELECT worker_id FROM t_alarm_worker WHERE occupy_status = 1
└─ 得到已占用的 SubnodeID 列表

步骤 3: 尝试抢占节点
├─ 遍历所有可用节点 (0-255)
├─ 跳过已占用节点
├─ 尝试 Redis 分布式锁:
│   ├─ key = "alarm_snowflake_{subnode}"
│   ├─ 过期时间 = 10分钟
│   └─ SETNX 原子操作
│
├─ 抢占成功:
│   ├─ 插入 Worker 记录到数据库
│   ├─ 启动心跳协程 keepWorker()
│   └─ 初始化雪花节点
│
└─ 抢占失败: 继续尝试下一个节点

步骤 4: 心跳保活
├─ 定时器: 每小时触发一次
├─ 更新心跳: UPDATE t_alarm_worker SET heart_beat = NOW()
└─ 程序退出时清理 Worker 记录

步骤 5: 失败处理
└─ 所有节点抢占失败: panic (告警ID无法生成)
```

### 5. 设备缓存同步流程

**入口**: `lcache/c_agent.go:19`

```
步骤 1: 定期触发
├─ 定时器: 每 2 小时触发
└─ 计数器: 每 12 次增量后触发 1 次全量

步骤 2: 检查是否需要同步
├─ 调用 CMDB 获取模组列表和版本号
├─ 比对本地缓存版本号
├─ 增量同步: 版本号不一致的模组
└─ 全量同步: 所有模组

步骤 3: 获取设备列表
├─ 调用 CMDB GetDeviceEntity 接口
├─ 分页查询: 每页 8000 条
└─ 循环获取所有设备

步骤 4: 更新本地缓存
├─ 设备信息写入缓存: localcache.Set()
├─ 缓存过期时间: 3 天
└─ 更新模组版本号

步骤 5: 缓存使用
├─ 告警处理时读取设备信息
├─ 缓存未命中: 跳过设备信息填充
└─ 缓存命中: 填充设备名称、机房等字段
```

---

## 配置说明

### 配置文件: trpc_go.yaml

```yaml
# ===== 服务配置 =====
etrpc:
  service_name: alarm-manage
  service_port: ${PORT_ALARM_MANAGE}

# ===== Kafka 配置 =====
server:
  service:
    # Kafka 消费者
    - name: trpc.kafka.consumer.service
      address: ${KAFKA_ADDR}?topics=${KAFKA_ALARM_PRODUCE_TOPIC}&group=alarm-manage&initial=newest
      protocol: kafka

# ===== 客户端配置 =====
client:
  service:
    # MySQL 数据库
    - name: trpc.mysql.tbos.alarm
      target: dsn://${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_ADDR})/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=True&loc=Local
    
    # Redis
    - name: trpc.redis.tbos.alert
      target: redis://:${REDIS_PASSWORD}@${REDIS_ADDR}/0
    
    # Kafka 生产者 (CGI同步)
    - name: trpc.kafka.tbos.cgi_sync
      target: kafka://${KAFKA_ADDR}?topic=${KAFKA_ALARM_PUSH_TOPIC}&partitioner=roundrobin&maxMessageBytes=3145728&compression=lz4
    
    # CMDB 服务
    - callee: tbos.cmdb.ConfigQuery
      name: cmdb
      protocol: http
      target: ip://${LOCAL_IP}:${PORT_CMDB}

# ===== 雪花算法配置 =====
snowflake_config:
  set_id: 1                 # 片区ID,每个片区唯一
  update_interval: 1        # 心跳更新间隔(小时)

# ===== 告警处理配置 =====
AlertManageConfig:
  BatchChannelSize: 1000          # 批处理大小
  BatchFetchIntervalMS: 50        # 批处理时间间隔(毫秒)
  PoolSize: 15                    # 线程池大小

# ===== 恢复处理配置 =====
RestoreManageConfig:
  BatchChannelSize: 1000
  BatchFetchIntervalMS: 200
  PoolSize: 5

# ===== 设备缓存同步配置 =====
SyncDeviceCacheConfig:
  BatchSize: 8000                    # 分页大小
  RegularSyncDeviceInterval: 7200    # 同步间隔(秒), 默认2小时
  TotalLoadIntervalCnt: 12           # 全量加载间隔(次), 每12次增量后1次全量

# ===== 数据库表名配置 =====
MysqlTableName:
  ActiveAlarm: t_alarm_active      # 活动告警表
  HistoryAlarm: t_alarm_history    # 历史告警表
```

### 配置项说明

| 配置项 | 说明 | 默认值 | 调优建议 |
|--------|------|--------|----------|
| BatchChannelSize | 批处理大小 | 1000 | 根据消息速率调整,建议500-2000 |
| BatchFetchIntervalMS | 批处理间隔 | 50/200ms | 延迟敏感可降低,吞吐优先可提高 |
| PoolSize | 线程池大小 | 15/5 | 根据CPU核心数调整,建议核心数1-2倍 |
| RegularSyncDeviceInterval | 缓存同步间隔 | 7200s | 设备变更频繁可降低 |
| TotalLoadIntervalCnt | 全量加载频率 | 12次 | 增量可靠可提高,降低全量频率 |

---

## 常见问题

### 1. 告警消费延迟

**现象**: 日志出现 "告警消费速度过慢" 或 "告警消费速度较慢"

**可能原因**:
1. Kafka 消费速度跟不上生产速度
2. 数据库写入性能不足
3. 线程池大小不足
4. 网络延迟或抖动

**排查步骤**:
```bash
# 1. 检查 Kafka 消费 Lag
kafka-consumer-groups --bootstrap-server <kafka> --group alarm-manage --describe

# 2. 检查数据库性能
# 查看慢查询日志,检查是否有慢SQL

# 3. 查看服务 CPU 和内存使用率
top -p <pid>

# 4. 查看告警处理耗时
# 检查日志中的处理时间指标
```

**解决方案**:
1. **提高并发**: 增大 `PoolSize` (线程池)
2. **优化批处理**: 调整 `BatchChannelSize` 和 `BatchFetchIntervalMS`
3. **扩容实例**: 增加 Pod 副本数,提高并行消费能力
4. **数据库优化**: 
   - 检查索引是否完善 (fingerprint 需要唯一索引)
   - 增大数据库连接池
   - 考虑读写分离
5. **缓存预热**: 确保设备缓存已同步,避免缓存未命中

### 2. 雪花ID生成冲突

**现象**: 出现重复的告警ID或启动时 panic

**可能原因**:
1. 多个实例占用相同的 SubnodeID
2. Redis 锁机制失效
3. 系统时钟回拨
4. Worker 心跳未正常更新

**排查步骤**:
```sql
-- 1. 检查 Worker 占用情况
SELECT * FROM t_alarm_worker WHERE occupy_status = 1;

-- 2. 检查是否有重复占用
SELECT worker_id, COUNT(*) 
FROM t_alarm_worker 
WHERE occupy_status = 1 
GROUP BY worker_id 
HAVING COUNT(*) > 1;

-- 3. 检查心跳时间
SELECT *, TIMESTAMPDIFF(HOUR, heart_beat, NOW()) as hours_ago 
FROM t_alarm_worker 
WHERE occupy_status = 1;
```

**解决方案**:
1. **清理过期 Worker**: 
   ```sql
   DELETE FROM t_alarm_worker WHERE heart_beat < DATE_SUB(NOW(), INTERVAL 24 HOUR);
   ```
2. **检查 Redis 连通性**: 确保 Redis 正常工作
3. **时钟同步**: 确保所有节点时钟同步 (使用 NTP)
4. **检查配置**: 确保每个片区的 `set_id` 唯一

### 3. 数据库写入超时

**现象**: 日志出现 "BatchInsertActiveAlerts failed" 或数据库连接超时

**可能原因**:
1. 数据库连接池耗尽
2. 数据库负载过高
3. 批量插入数据量过大
4. 网络抖动

**排查步骤**:
```sql
-- 1. 检查数据库连接数
SHOW PROCESSLIST;

-- 2. 检查是否有锁等待
SHOW ENGINE INNODB STATUS;

-- 3. 检查表大小
SELECT 
  TABLE_NAME,
  ROUND((DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024, 2) AS `Size (MB)`
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'your_database'
  AND TABLE_NAME IN ('t_alarm_active', 't_alarm_history');
```

**解决方案**:
1. **优化索引**: 确保 `fingerprint` 有唯一索引
2. **增大连接池**: 修改数据库连接池配置
3. **分批插入**: 降低 `BatchChannelSize` 减少单次插入量
4. **定期归档**: 定期清理历史告警数据
5. **检查慢查询**: 优化慢SQL

### 4. 告警重复入库

**现象**: 相同告警出现多条记录

**可能原因**:
1. 指纹生成逻辑变更导致指纹不一致
2. 数据库唯一索引缺失
3. Kafka 消息重复消费
4. 事务未正确处理

**排查步骤**:
```sql
-- 查找重复的告警
SELECT fingerprint, COUNT(*) as cnt 
FROM t_alarm_active 
GROUP BY fingerprint 
HAVING cnt > 1;

-- 检查索引
SHOW INDEX FROM t_alarm_active WHERE Key_name = 'fingerprint';
```

**解决方案**:
1. **添加唯一索引**:
   ```sql
   ALTER TABLE t_alarm_active ADD UNIQUE INDEX idx_fingerprint (fingerprint);
   ```
2. **检查指纹生成**: 确保 `GeneFingerPrint()` 逻辑正确
3. **Kafka 幂等性**: 代码中使用 `OnConflict.DoNothing` 保证幂等

### 5. 设备信息缺失

**现象**: 告警中设备名称、机房等字段为空

**可能原因**:
1. 设备缓存未同步
2. CMDB 服务异常
3. 设备在 CMDB 中不存在
4. 缓存过期

**排查步骤**:
```bash
# 1. 检查缓存同步日志
# 查找 "DoUpdateDeviceCache" 相关日志

# 2. 检查 CMDB 服务连通性
curl http://${CMDB_HOST}:${CMDB_PORT}/health

# 3. 手动触发缓存同步
# (需要暴露手动同步接口或重启服务)
```

**解决方案**:
1. **手动同步**: 重启服务触发全量同步
2. **缩短同步间隔**: 降低 `RegularSyncDeviceInterval`
3. **检查 CMDB**: 确认设备在 CMDB 中存在
4. **容错处理**: 代码中对缓存未命中做了容错,不会阻断告警流程

### 6. Kafka 推送失败

**现象**: 日志出现 "SendCgiAlarm failed"

**可能原因**:
1. Kafka 服务异常
2. Topic 不存在
3. 消息体过大
4. 网络问题

**排查步骤**:
```bash
# 1. 检查 Kafka 服务
kafka-topics --bootstrap-server <kafka> --list

# 2. 检查 Topic 配置
kafka-topics --bootstrap-server <kafka> --describe --topic <topic>

# 3. 检查消息大小限制
# 配置中 maxMessageBytes=3145728 (3MB)
```

**解决方案**:
1. **创建 Topic**: 确保目标 Topic 存在
2. **增大消息限制**: 调整 `maxMessageBytes` 配置
3. **分批发送**: 代码中已实现分批发送逻辑
4. **重试机制**: 代码中使用 retry.Attempts(3) 重试

### 7. 服务启动失败 (panic)

**现象**: 服务启动时 panic,日志显示雪花算法初始化失败

**可能原因**:
1. 所有 SubnodeID 都被占用 (超过256个实例)
2. Redis 连接失败
3. 数据库连接失败

**排查步骤**:
```bash
# 1. 检查 Worker 占用数量
mysql> SELECT COUNT(*) FROM t_alarm_worker WHERE occupy_status = 1;

# 2. 检查 Redis 连通性
redis-cli -h <host> -p <port> PING

# 3. 检查数据库连通性
mysql -h <host> -u <user> -p -e "SELECT 1"
```

**解决方案**:
1. **清理过期 Worker**: 手动清理 24 小时前的 Worker
2. **检查依赖**: 确保 Redis 和 MySQL 正常运行
3. **扩容节点**: 如果确实超过 256 实例,需要增加片区 (set_id)

---

## 性能指标

### 系统容量

| 指标 | 数值 | 说明 |
|------|------|------|
| 单实例 QPS | ~10000 | 取决于批处理配置和硬件 |
| 支持实例数 | 256/片区 | 受雪花算法限制 |
| 通道缓冲 | 20000 (告警) / 5000 (恢复) | 防止消息堆积 |
| 批处理大小 | 1000 | 提高数据库写入效率 |
| 线程池大小 | 15 (告警) / 5 (恢复) | 并发处理能力 |

### 监控指标

| 指标名 | 说明 | 维度 |
|--------|------|------|
| consume_alert_cnt | 消费告警数 | mozuId |
| total_db_req_cnt | DB请求总数 | - |
| success_db_req_cnt | DB成功请求数 | - |
| total_db_write_cnt | DB写入总数 | mozuId |
| success_db_write_cnt | DB成功写入数 | mozuId |

---

## 开发建议

### 添加新的告警处理逻辑

1. 在 `logic/manager/alert_manager.go` 中的 `processEdgeFireAlert()` 添加逻辑
2. 如需修改字段填充,修改 `fillActivesAlert()` 函数
3. 如需修改去重逻辑,修改 `getUniqueActiveList()` 函数

### 添加新的通知渠道

1. 在 `logic/notification/` 下新建子目录
2. 实现通知接口
3. 在 `notification.go` 的 `ReportAlert()` 中调用

### 修改指纹生成规则

⚠️ **警告**: 修改指纹规则会导致已有告警无法正确恢复!

如确需修改:
1. 修改 `manager.go:62` 的 `GeneFingerPrint()` 函数
2. 修改数据库唯一索引
3. 需要数据迁移方案
