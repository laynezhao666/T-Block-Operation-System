# Collector 配置与数据总线服务

Collector 是 TBOS 的配置收集与数据转发服务，负责为 Agent 提供采集配置获取能力，并将 Agent 上报的测点数据转发到 Kafka 或上游 Collector。

## 模块介绍

Collector 服务包含以下核心模块：

### 1. 配置总线 (ConfigBus)
负责为下游 Agent 提供采集配置获取服务，支持多种配置类型：

| 配置类型 | 说明 |
|----------|------|
| `FETCH_COLLECTOR_DEVICES` | 采集设备配置 |
| `FETCH_COLLECTOR_TEMPLATES` | 采集模板配置 |
| `FETCH_STD_POINTS` | 标准测点配置 |
| `FETCH_CONFIG_MODIFY_TIME` | 配置修改时间 |
| `FETCH_STD_DEVICES` | 标准设备配置 |

### 2. 数据总线 (DataBus)
负责接收 Agent 上报的测点数据并转发：

- **TBOS 测点数据**：标准化测点数据，上报到 TBOS 系统
- **采集测点数据**：原始采集数据，可转发到原有动环系统
- **外部平台数据**：接收外部平台的测点数据上报

### 3. 控制总线 (ControlBus)
负责处理控制类请求：

- 心跳转发：将 Agent 心跳转发到上游监控服务

## 核心能力

### 1. 多源配置获取

系统支持注册多个配置获取器（Fetcher），按注册顺序依次尝试获取配置：

```go
// IConfigFetcher 采集器配置获取器接口
type IConfigFetcher interface {
    // Name 获取器的名称
    Name() string
    // FetchCollectDevices 获取采集设备配置
    FetchCollectDevices(deviceNumbers []string) ([]byte, error)
    // FetchCollectTemplates 获取采集模板配置
    FetchCollectTemplates(templateNames []string) ([]byte, error)
    // FetchStdPoints 获取采集设备相关标准点配置
    FetchStdPoints(deviceNumbers []string) ([]byte, error)
    // FetchConfigModifyTime 获取采集设备配置修改时间
    FetchConfigModifyTime(deviceNumbers []string) ([]byte, error)
    // FetchStdDevices 获取采集设备对应标准设备的配置
    FetchStdDevices(collectDeviceNumbers []string) ([]byte, error)
}
```

**当前支持的配置获取器**：

| 获取器名称 | 优先级 | 说明 |
|------------|--------|------|
| `cmdb config fetcher` | 1（优先） | 从 CMDB 服务获取配置，成功后自动缓存到本地 |
| `local file config fetcher` | 2（兜底） | 从本地文件获取配置 |

### 2. 配置自动缓存

从 CMDB 获取配置成功后，会自动将配置保存到本地文件，作为后续的兜底缓存：

```
./conf/collect/
├── collect_devices/     # 采集设备配置缓存
├── collect_templates/   # 采集模板配置缓存
├── std_points/          # 标准测点配置缓存
├── std_devices/         # 标准设备配置缓存
└── modify_time/         # 配置修改时间缓存
```

### 3. 超时控制策略

配置获取采用分级超时策略：
- 主获取器（CMDB）：占用总超时时间的 50%
- 备用获取器（本地文件）：平分剩余超时时间

### 4. 数据转发模式

支持两种数据转发模式，通过 `features.send_type` 配置：

| 模式 | 配置值 | 说明 |
|------|--------|------|
| Kafka 直推 | `kafka` | 直接推送到 Kafka，支持主备切换 |
| Collector 转发 | `collector` | 转发到上游 Collector（NUC 场景） |

#### Kafka 直推模式

```
Agent → Collector → Kafka (main) → 失败 → Kafka (backup)
```

支持的 Kafka Topic：
- `mainStdPoint` / `backupStdPoint`：TBOS 标准测点
- `mainCollectPoint` / `backupCollectPoint`：采集测点
- `grayCollectPoint`：灰度环境测点

#### Collector 转发模式

```
Agent (边端) → Collector (边端) → Collector (云端) → Kafka
```

适用于 NUC 等边缘节点，边端 Collector 将数据转发到云端 Collector。

### 5. 指标上报

内置多维度指标上报能力：

| 指标名称 | 说明 |
|----------|------|
| `handle_cnt` | 请求处理数量 |
| `handle_fail_cnt` | 请求处理失败数量 |
| `max_latency` | 最大处理时延 |
| `avg_latency` | 平均处理时延 |
| `send_data_cnt` | 数据发送数量 |
| `send_data_fail_cnt` | 数据发送失败数量 |
| `fetch_config_cnt` | 配置获取数量 |
| `fetch_config_fail_cnt` | 配置获取失败数量 |

## 代码结构

```
collector/
├── main.go                              # 主入口，注册服务并启动
├── trpc_go.yaml                         # 服务配置文件
├── config.json                          # 应用配置文件
├── conf/                                # 配置缓存目录
│   └── collect/                         # 采集配置缓存
│       ├── collect_devices/             # 设备配置缓存
│       ├── collect_templates/           # 模板配置缓存
│       ├── std_devices/                 # 标准设备配置缓存
│       ├── std_points/                  # 标准测点配置缓存
│       └── modify_time/                 # 修改时间缓存
├── entity/                              # 实体定义
│   ├── collectors/                      # 采集器相关
│   │   └── file.go                      # 配置文件路径常量
│   ├── config/                          # 配置相关
│   │   └── config.go                    # 配置结构定义
│   ├── errcode/                         # 错误码定义
│   │   └── errcode.go                   # 错误码常量
│   └── model/                           # 数据模型
│       └── kafka.go                     # Kafka 消息模型
├── logic/                               # 业务逻辑
│   └── bus/                             # 总线逻辑
│       ├── collectors_config/           # 配置获取模块
│       │   ├── fetcher.go               # IConfigFetcher 接口定义
│       │   ├── handle.go                # 配置获取处理器（多源切换、超时控制）
│       │   ├── cmdb/                    # CMDB 获取器实现
│       │   │   ├── impl.go              # CMDB 配置获取实现
│       │   │   └── save.go              # 配置本地缓存保存
│       │   └── localfile/               # 本地文件获取器实现
│       │       └── impl.go              # 本地文件配置获取实现
│       ├── control/                     # 控制总线
│       │   └── handle.go                # 心跳处理
│       └── data/                        # 数据总线
│           ├── sender.go                # 数据发送（Kafka/Collector）
│           ├── collect_point/           # 采集测点处理
│           │   └── handle.go            # 采集测点转发处理（支持灰度）
│           ├── tbos_point/              # TBOS 测点处理
│           │   └── handle.go            # TBOS 测点转发处理
│           └── external/                # 外部平台数据处理
│               └── handle.go            # 外部平台数据接收处理
├── service/                             # 服务接口实现
│   ├── config_fetch.go                  # ConfigBusService 实现
│   ├── data_bus.go                      # DataBusService 实现
│   ├── control_bus.go                   # MonitorService（心跳）实现
│   ├── collect_point_forward.go         # CollectPointForwardService 实现
│   └── external.go                      # ExternalPlatformService 实现
├── repo/                                # 数据访问层
│   ├── sender.go                        # 发送器初始化（根据配置选择 Kafka/Collector）
│   ├── kafka/                           # Kafka 发送器
│   │   └── sender.go                    # Kafka 生产者管理
│   ├── collector/                       # Collector 转发器
│   │   └── sender.go                    # 上游 Collector 转发
│   └── report/                          # 指标上报
│       └── report.go                    # 多维度指标上报实现
└── utils/                               # 工具类
    ├── file.go                          # 文件操作
    ├── md5.go                           # MD5 计算
    ├── net.go                           # 网络工具
    ├── panic.go                         # Panic 恢复
    └── upstream.go                      # 上游 IP 获取
```

## 配置文件

### trpc_go.yaml

```yaml
etrpc:
  service_name: collector
  service_port: ${PORT_COLLECTOR}

global:
  max_frame_size: 1048576000    # 最大帧大小（约 1GB）
  namespace: Production
  local_ip: ${LOCAL_IP}

server:
  service:
    - name: ${etrpc.service_name}
      network: tcp
      ip: 0.0.0.0
      port: ${etrpc.service_port}
      protocol: http

client:
  network: tcp
  service:
    # Kafka 生产者配置
    - name: trpc.kafka.producer.mainStdPoint
      target: kafka://${KAFKA_ADDR}?topic=${KAFKA_POINT_TOPIC}&partitioner=roundrobin&maxMessageBytes=3145728&compression=lz4
      timeout: 30000
    - name: trpc.kafka.producer.backupStdPoint
      target: kafka://${KAFKA_ADDR}?topic=${KAFKA_POINT_TOPIC}&partitioner=roundrobin&maxMessageBytes=3145728&compression=lz4
      timeout: 30000
    # CMDB 客户端配置
    - callee: tbos.cmdb.ConfigQuery
      name: cmdb
      protocol: http
      target: ip://${LOCAL_IP}:${PORT_CMDB}

# 功能开关
features:
  send_type: kafka              # 发送模式：kafka 或 collector
  trace: false                  # 追踪开关
  gray: false                   # 灰度开关
```

### 配置项说明

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `features.send_type` | 数据发送模式 | `kafka`（直推）/ `collector`（转发） |
| `features.trace` | 追踪开关 | `true` / `false` |
| `features.gray` | 灰度开关，开启后数据同时推送到灰度 Kafka | `true` / `false` |

## API 接口

### ConfigBusService

| 接口 | 方法 | 说明 |
|------|------|------|
| `/ConfigBus/FetchConfig` | POST | 获取采集配置 |

**请求参数**：
```json
{
    "fetch_type": 0,         // 配置类型枚举
    "params": ["device1", "device2"]  // 设备编号或模板名称列表
}
```

**FetchType 枚举**：
- `0`: FETCH_COLLECTOR_DEVICES
- `1`: FETCH_COLLECTOR_TEMPLATES
- `2`: FETCH_STD_POINTS
- `3`: FETCH_CONFIG_MODIFY_TIME
- `4`: FETCH_STD_DEVICES

### DataBusService

| 接口 | 方法 | 说明 |
|------|------|------|
| `/DataBus/Send` | POST | 发送 TBOS 标准测点数据 |

### CollectPointForwardService

| 接口 | 方法 | 说明 |
|------|------|------|
| `/CollectPointForward/Forward` | POST | 转发采集测点到原有动环 |

### ExternalPlatformService

| 接口 | 方法 | 说明 |
|------|------|------|
| `/ExternalPlatform/Data` | POST | 接收外部平台测点数据 |

### MonitorService（ControlBus）

| 接口 | 方法 | 说明 |
|------|------|------|
| `/Monitor/Heartbeat` | POST | 心跳转发 |

## 错误码

| 错误码 | 常量名 | 说明 |
|--------|--------|------|
| 270200 | `ErrSendFail` | 数据发送失败 |
| 270201 | `ErrFetchConfigFail` | 配置获取失败 |
| 270202 | `ErrRequestContentMissed` | 请求内容缺失 |
| 270203 | `ErrHeartbeatFail` | 心跳发送失败 |

## 常见问题

### 1. 配置获取失败

**问题描述**：Agent 调用 FetchConfig 接口返回 `ErrFetchConfigFail`

**排查步骤**：
1. 检查 CMDB 服务是否正常运行
2. 检查 `trpc_go.yaml` 中 CMDB 地址配置是否正确
3. 检查本地缓存目录 `./conf/collect/` 是否有对应的配置文件
4. 查看日志确认具体的失败原因

**解决方案**：
- 确保 CMDB 服务可访问
- 如果 CMDB 不可用，确保本地缓存目录有有效的配置文件
- 检查请求的 `params` 参数是否正确

### 2. Kafka 发送失败

**问题描述**：测点数据发送到 Kafka 失败

**排查步骤**：
1. 检查 Kafka 服务是否正常
2. 确认 Kafka broker 地址配置正确
3. 检查网络连接
4. 查看日志中的具体错误信息

**解决方案**：
- 系统会自动尝试备用 Kafka，检查主备 Kafka 配置
- 确认 Kafka topic 是否存在
- 检查消息大小是否超过 `maxMessageBytes` 限制（默认 3MB）

### 3. 请求超时

**问题描述**：配置获取请求超时

**原因分析**：
- CMDB 响应慢
- 网络延迟高
- 配置数据量过大

**解决方案**：
- 调整客户端超时时间
- 减少单次请求的设备/模板数量
- CMDB 超时后会自动切换到本地文件获取

### 4. 本地缓存文件格式错误

**问题描述**：从本地文件获取配置失败，报 `unmarshal` 错误

**排查步骤**：
1. 检查对应的 JSON 文件格式是否正确
2. 确认文件编码为 UTF-8
3. 验证 JSON 语法

**解决方案**：
- 修复 JSON 文件格式
- 删除错误的缓存文件，等待 CMDB 重新同步

### 5. Collector 转发模式配置

**问题描述**：边端 Collector 无法转发数据到云端

**排查步骤**：
1. 确认 `features.send_type` 配置为 `collector`
2. 检查云端 Collector 服务是否可访问
3. 检查服务名配置是否正确

**解决方案**：
- 确保云端 Collector 服务 `idc-tbos-collector-std` 和 `idc-tbos-collector-collect` 可访问
- 检查网络连通性

### 6. 灰度模式数据推送

**问题描述**：需要将数据同时推送到灰度环境

**配置方法**：
```yaml
features:
  gray: true
```

开启后，采集测点数据会同时推送到 `grayCollectPoint` Kafka。

## 部署说明

### 编译

```bash
go build -o collector main.go
```

### 运行

```bash
./collector -conf trpc_go.yaml
```

### 环境变量

| 变量名 | 说明 | 示例 |
|--------|------|------|
| `PORT_COLLECTOR` | Collector 服务端口 | `8088` |
| `LOCAL_IP` | 本机 IP | `127.0.0.1` |
| `KAFKA_ADDR` | Kafka 地址 | `localhost:9092` |
| `KAFKA_POINT_TOPIC` | Kafka Topic | `tbos_point_data` |
| `PORT_CMDB` | CMDB 服务端口 | `8087` |

### 目录权限

确保以下目录有读写权限：
```bash
mkdir -p ./conf/collect/{collect_devices,collect_templates,std_points,std_devices,modify_time}
```