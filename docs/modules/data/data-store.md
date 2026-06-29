# 数据存储服务 Data-Store

Data-Store 是TBOS的测点数据持久化存储服务，负责消费Kafka中的测点数据并通过可插拔的存储插件写入到不同的存储后端（如InfluxDB、Kafka等）。

## 模块介绍

Data-Store 服务主要负责：
- 消费Kafka中的测点数据消息（支持主备双Kafka消费）
- 解析测点数据，包括普通测点和虚拟测点
- 通过插件化的存储架构，将数据写入不同的存储后端
- 提供完善的监控指标上报

### 数据流向

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  Kafka(主用)     │────▶│                  │────▶│   InfluxDB       │
│ tbos.major       │     │                  │     │   (持久化存储)    │
└──────────────────┘     │   Data-Store     │     └──────────────────┘
                         │                  │
┌──────────────────┐     │  - 消息解析       │     ┌──────────────────┐
│  Kafka(备用)     │────▶│  - 测点提取       │────▶│   Kafka          │
│ tbos.backup      │     │  - 插件分发       │     │   (数据转发)      │
└──────────────────┘     └──────────────────┘     └──────────────────┘
```

## 核心能力

### 1. 主备Kafka双活消费

服务支持同时消费主用和备用两个Kafka集群的数据，通过不同的消费者实例区分数据来源：

```go
// 主用Kafka消费
kafka.RegisterBatchHandlerService(majorKafkaSvr, service.MajorKafkaHandle)

// 备用Kafka消费  
kafka.RegisterBatchHandlerService(backUpKafkaSvr, service.BackupKafkaHandle)
```

### 2. 可插拔存储架构

采用插件化设计，支持灵活配置多个存储后端：

- **插件注册**：通过 `store.Register()` 注册存储插件
- **插件配置**：通过配置文件启用和配置插件
- **并发写入**：支持同步/异步两种写入模式

内置存储插件: 

| 插件类型 | 说明 |
|---------|------|
| `influx` | InfluxDB存储，用于持久化时序数据 |
| `kafka` | Kafka转发，用于数据分发给下游服务 |

### 3. 测点数据解析

支持解析多种类型的测点数据：
- **采集测点** (PointTypeCollect)：从设备采集的原始测点
- **标准测点** (PointTypeStd)：标准化后的测点
- **告警测点** (PointTypeAlarm)：告警相关测点
- **虚拟测点** (PointTypeVirtual)：计算生成的虚拟测点

### 4. 监控指标上报

提供完善的监控指标：

| 指标名称 | 说明 |
|---------|------|
| `kafka_consume_delay` | Kafka消费延迟（平均/最大） |
| `kafka_consume_msg_cnt` | Kafka消费消息数量 |
| `kafka_consume_point_cnt` | Kafka消费测点数量 |
| `influx_write_success_cnt` | InfluxDB写入成功数 |
| `influx_write_fail_cnt` | InfluxDB写入失败数 |
| `influx_write_cost` | InfluxDB写入耗时（平均/最大） |
| `kafka_write_success_cnt` | Kafka写入成功数 |
| `kafka_write_fail_cnt` | Kafka写入失败数 |
| `kafka_write_cost` | Kafka写入耗时（平均/最大） |

指标维度：
- `mozu_id`：模组ID
- `source`：Kafka数据源标识
- `point_type`：测点类型
- `interval`：测点周期

## 代码结构

```
data-store/
├── main.go                           # 主入口，初始化存储插件和Kafka消费者
├── trpc_go.yaml                      # 服务配置文件
├── entity/                           # 实体定义
│   ├── kafkamodel/                   # Kafka消息模型
│   │   └── kafka_msg.go              # 定义PointMsgKey、PointMsgValue、Point结构
│   └── model/                        # 内部数据模型
│       └── point.go                  # 定义Point、OriginPointMsg结构
├── logic/                            # 业务逻辑层
│   └── consumer/                     # 消费逻辑
│       └── point_consumer.go         # Kafka消息消费和测点解析核心逻辑
├── repo/                             # 数据访问层
│   ├── report/                       # 指标上报
│   │   └── report_metrics.go         # 定义各类监控指标
│   └── store/                        # 存储插件
│       ├── plugin.go                 # 插件管理：注册、初始化、批量写入
│       ├── influx_store.go           # InfluxDB存储插件实现
│       └── kafka_store.go            # Kafka转发插件实现
└── service/                          # 服务层
    └── consumer_service.go           # Kafka消费服务入口
```

### 核心文件说明

#### main.go
服务主入口，负责：
1. 创建etrpc服务器
2. 初始化存储插件 `store.Init()`
3. 注册主备Kafka消费者服务
4. 服务停止时关闭存储插件

#### logic/consumer/point_consumer.go
核心消费逻辑，实现 `IKafkaConsumer` 接口：
- `BatchHandle()`：批量消费Kafka消息
- `resolvePoints()`：解析测点数据，包括类型识别、模组ID解析、质量码解析
- `reportMetric()`：上报消费指标

#### repo/store/plugin.go
存储插件管理，定义：
- `IStorePlugin` 接口：Setup/Write/Close
- `Register()`：注册存储插件
- `Init()`：初始化所有启用的插件
- `BatchWritePoint()`：并发写入所有插件
- `Close()`：关闭所有存储通道

#### repo/store/influx_store.go
InfluxDB存储插件：
- 支持配置存储天数、数据库名、测量名、批次大小
- 分批并发写入
- 写入失败自动重试3次

#### repo/store/kafka_store.go
Kafka转发插件：
- 将原始Kafka消息转发到下游Kafka
- 写入失败自动重试3次

## 配置说明

### trpc_go.yaml 配置示例

```yaml
etrpc:
  service_name: data-store
  service_port: ${PORT_DATA_STORE}

server:
  service:
    # HTTP服务
    - name: ${etrpc.service_name}
      protocol: http
      port: ${etrpc.service_port}
    # 主用Kafka消费者
    - name: trpc.kafka.tbos.major
      address: ${KAFKA_ADDR}?topics=${KAFKA_POINT_TOPIC}&group=data-store&batch=20&batchFlush=200&initial=oldest
      protocol: kafka

client:
  service:
    # InfluxDB客户端
    - name: trpc.influxdb.idc.tbos
      target: influxdb://${INFLUXDB_USER}:${INFLUXDB_PASSWORD}@${INFLUXDB_ADDR}?timeout=10000&write_encoding=gzip

# 存储插件配置
store:
  plugins:
    # InfluxDB存储插件
    - name: tbos-influx
      type: influx
      async: false          # 同步写入
      extra:
        influx_name: trpc.influxdb.idc.tbos
        influx_database: ${INFLUXDB_DBNAME}
        data_measurement: ${INFLUXDB_POINT_MEASUREMENT}
        bath_size: 50000    # 批次大小
        store_day: 100      # 存储天数
    
    # Kafka转发插件
    - name: tbos-kafka-forward
      type: kafka
      async: true           # 异步写入
      extra:
        kafka_name: trpc.kafka.downstream
```

### 存储插件配置项

#### InfluxDB插件 (type: influx)
| 配置项 | 说明 | 默认值 |
|-------|------|--------|
| influx_name | InfluxDB客户端名称 | 必填 |
| influx_database | 数据库名称 | tbos |
| data_measurement | 测量名称 | points |
| bath_size | 批量写入大小 | 10000 |
| store_day | 存储天数 | 100 |

#### Kafka插件 (type: kafka)
| 配置项 | 说明 | 默认值 |
|-------|------|--------|
| kafka_name | Kafka客户端名称 | 必填 |

## Kafka消息格式

### 消息Key (PointMsgKey)
```json
{
  "mID": "模组ID",
  "dID": "设备ID",
  "wID": "WorkerID",
  "seq": 1,
  "t": 1699999999,
  "d": 60,
  "bKey": "业务Key",
  "pubMs": 1699999999000,
  "type": 1
}
```

### 消息Value (PointMsgValue)
```json
{
  "interval": 60,
  "box_id": "TBox ID",
  "points": [
    {"i": "测点名称", "v": "25.5", "q": "0", "t": "1699999999"}
  ],
  "virtual_points": [
    {"i": "虚拟测点名称", "v": "100", "q": "0", "t": "1699999999"}
  ]
}
```

## 常见问题

### 1. InfluxDB写入失败

**问题表现**：日志中出现 "write points to influxdb fail" 告警

**可能原因**：
- InfluxDB服务不可用
- 网络连接异常
- 数据库配置错误
- 磁盘空间不足

**解决方案**：
- 检查InfluxDB服务状态
- 确认 `influx_name` 配置的客户端名称正确
- 检查网络连接和防火墙配置
- 检查InfluxDB磁盘空间

### 2. Kafka消费延迟过高

**问题表现**：`kafka_consume_delay` 指标值持续较高

**可能原因**：
- 存储插件写入速度慢
- 单批消息量过大
- 消费者处理能力不足

**解决方案**：
- 调整Kafka消费批次配置 `batch` 和 `batchFlush`
- 增加 `bath_size` 提高批量写入效率
- 启用异步写入 `async: true`
- 水平扩展消费者实例

### 3. 数据未能写入存储

**问题表现**：InfluxDB中查询不到数据

**可能原因**：
- 存储插件未正确配置
- 插件类型名称拼写错误
- 插件初始化失败

**解决方案**：
- 检查 `store.plugins` 配置是否正确
- 查看启动日志确认 "store plugin setup success"
- 确认插件类型与已注册的类型一致

### 4. 测点数据被过滤

**问题表现**：部分测点数据未被存储

**可能原因**：
- 测点时间戳超过当前时间1分钟（被视为未来数据）
- 测点质量码解析失败
- 消息Key或Value格式错误

**解决方案**：
- 检查上游数据源的时间戳是否正确
- 确认测点数据格式符合规范
- 查看日志中的 "receive bad" 警告信息

### 5. 服务停止时数据丢失

**问题表现**：服务重启后部分异步写入的数据丢失

**可能原因**：
- 异步写入的数据未完成就停止服务

**解决方案**：
- 服务会在停止时等待异步写入完成 `wg.Wait()`
- 确保使用正常的停止信号（SIGTERM）而非强制kill
- 对于关键数据，使用同步写入模式 `async: false`

### 6. 如何扩展新的存储插件

1. 实现 `IStorePlugin` 接口：
```go
type IStorePlugin interface {
    Setup(cfg PlgConfig) (IStorePlugin, error)
    Write([]*model.OriginPointMsg)
    Close() error
}
```

2. 在 `init()` 中注册插件：
```go
func init() {
    Register("your_plugin_type", &YourStore{})
}
```

3. 在配置文件中启用：
```yaml
store:
  plugins:
    - name: your-plugin
      type: your_plugin_type
      extra:
        # 自定义配置
```
