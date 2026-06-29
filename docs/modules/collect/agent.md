# 边缘采集器 Agent

Agent 是部署在边缘设备上的数据采集服务，负责对接各类动环设备，采集实时数据并上报至数据中心。支持 `agent` 和 `agent-gw`（网关）两种运行模式。

## 模块介绍

Agent 是一个完整的边缘数据采集解决方案，主要包含以下核心模块：

### 1. 采集模块 (collector)
负责从各类设备采集数据，支持多种工业协议，并且可以按需灵活扩展：

| 协议 | 驱动名称 | 说明 |
|------|----------|------|
| Modbus | `modbus` | 支持 RTU/TCP 通信，支持串口（RS485/RS232）和网口方式 |
| SNMP | `snmp` | 网络设备管理协议，支持 GET 操作 |
| SysDIO | `sysdio` | 系统数字量输入输出，支持多种控制器（hyiot/sysclass/tbox/ual） |
| Simulator | `simulator` | 模拟器驱动，用于测试，支持静态值、随机值、单调递增等模式 |

### 2. 数据分发模块 (distribution)
负责将采集数据上报到上游系统：

- **Kafka 分发**：支持数据压缩、SASL 认证、多 Kafka 实例转发
- **HTTP 分发**：支持 HTTPS/TLS，可配置北向白名单
- **周期上报 (interval)**：按配置周期（默认60秒）全量上报测点数据
- **变化上报 (change)**：每秒检测一次，仅上报值发生变化的测点

### 3. 配置管理模块 (cm)
负责管理采集配置：

- 设备配置管理
- 模板配置管理
- 标准测点配置管理
- 配置版本管理
- 支持从 tlink、本地文件、备份等多种来源读取配置

### 4. 标准测点计算模块 (std)
负责将采集测点转换为标准测点：

- 基于表达式的映射计算
- 支持多采集点到单标准点的转换
- 并发计算，支持配置 worker 数量
- 质量标签传递与处理

### 5. 实时数据库 (rtdb)
负责存储测点实时数据：

- 测点值缓存与查询
- 测点告警状态管理
- 值变化检测与回调通知
- 虚拟测点数据管理

### 6. 插件系统 (plugin)
提供可扩展的插件机制：

- 支持插件注册与调度
- 定时执行（可配置间隔）
- 事件订阅与通知
- 内置插件：expressioncmd（表达式命令）、southdevice（南向设备）

### 7. 网络管理模块 (network)
负责边缘设备网络配置：

- LAN/WLAN 接口配置
- DNS 服务器配置
- 交换机模式支持
- 网络状态查询与设置

## 核心能力

### 1. 多协议采集

#### Modbus 采集
支持 Modbus RTU（串口）和 Modbus TCP（网口）两种通信方式：

```yaml
# 串口配置示例 (trpc_go.yaml)
collector:
  modbus:
    serials_map:
      COM1:
        baud: "9600"
        databit: "8"
        dev: /usr/dev/serial/com1
        id: COM1
        mode: "485"
        parity: "N"
        stopbit: "1"
```

#### SNMP 采集
支持 SNMP v1/v2c 协议：

```yaml
collector:
  snmp:
    timeout: 2000           # 超时时间(ms)
    retry: 1                # 重试次数
    max_coroutine: 100      # 最大协程数
    per_coroutine_points: 3000  # 每协程处理测点数
    request_total_timeout: 40000  # 请求总超时(ms)
```

### 2. 虚拟测点

系统自动生成设备级虚拟测点，用于监控设备采集状态：

| 虚拟测点ID | 说明 |
|------------|------|
| `commste` | 通讯状态 |
| `point_throughput` | 测点吞吐量（采集成功测点数/时间） |
| `range_resp_time` | 响应时间极差 |
| `max_resp_time` | 最大响应时间 |
| `min_resp_time` | 最小响应时间 |
| `avg_resp_time` | 平均响应时间 |
| `success_req_in_period` | 周期内成功请求次数 |
| `total_req_in_period` | 周期内总请求次数 |
| `interruption` | 通讯中断标志 |
| `Comm` | 通讯状态（兼容物模型） |
| `Comm_N` | 第N通道通讯状态 |

### 3. 数据上报机制

#### 周期上报
- 默认周期：60秒
- 内容：全量测点数据
- 支持配置上报秒数偏移，避免多 Pod 同时上报

#### 变化上报
- 周期：1秒
- 内容：值发生变化的测点
- 首次采集到的测点始终上报

### 4. 标准测点映射计算

支持基于表达式的标准测点转换：

```json
{
    "std_device": "设备标识",
    "std_point": "标准测点标识",
    "mapping": "采集测点ID",
    "expr": "x * 0.1 + 20"
}
```

### 5. 质量标签

测点采集质量状态定义：

| 质量码 | 说明 |
|--------|------|
| 0 | 正常 |
| -200 | 通信中断 |
| -201 | 报文发送失败 |
| -202 | 报文响应超时 |
| -203 | 报文响应错误 |
| -204 | 配置错误 |
| -300 | 通讯异常 |
| -301 | CRC校验错误 |
| -403 | 未采集 |
| -406 | 值越界 |
| -408 | 值格式转换错误 |

### 6. 配置热更新

支持运行时动态更新采集配置：

- 监听设备配置变化 → 重新加载采集任务
- 监听标准点配置变化 → 重新加载标准点计算
- 监听模板配置变化 → 重新加载设备模板

## 代码结构

```
agent/
├── main.go                              # 主入口，注册服务并启动
├── trpc_go.yaml                         # 服务配置文件
├── entity/                              # 实体定义
│   ├── config/                          # 配置相关
│   │   ├── config.go                    # 配置结构定义（采集、分发、插件等）
│   │   └── init.go                      # 配置初始化
│   ├── consts/                          # 常量定义
│   │   ├── version.go                   # 版本号
│   │   ├── qua.go                       # 质量标签定义
│   │   ├── device.go                    # 设备相关常量
│   │   ├── path.go                      # 路径常量
│   │   └── kafka.go                     # Kafka相关常量
│   ├── definition/                      # 数据结构定义
│   │   ├── definition.go                # 基础类型定义
│   │   ├── virtual_points.go            # 虚拟测点ID生成
│   │   └── datatype/                    # 数据类型定义
│   ├── model/                           # 业务模型
│   │   ├── device.go                    # 设备模型
│   │   ├── template.go                  # 模板模型
│   │   ├── std_device.go                # 标准设备模型
│   │   └── data/                        # 数据模型
│   ├── kafka/                           # Kafka消息定义
│   └── errcode/                         # 错误码定义
├── logic/                               # 业务逻辑
│   ├── setup/                           # 初始化逻辑
│   │   └── init.go                      # 初始化入口（网络→配置→cm→采集→标准点→分发）
│   ├── collector/                       # 采集模块
│   │   ├── init.go                      # 采集模块初始化
│   │   ├── device/                      # 设备管理
│   │   │   ├── device.go                # 设备实例
│   │   │   ├── channel.go               # 通道管理
│   │   │   ├── template.go              # 模板管理
│   │   │   ├── virtualpoints/           # 虚拟测点
│   │   │   │   └── virtual_points.go    # 虚拟测点计算
│   │   │   ├── driver/                  # 驱动层
│   │   │   │   ├── driver.go            # 驱动接口定义
│   │   │   │   ├── manager.go           # 驱动管理器
│   │   │   │   └── drivers/             # 驱动实现
│   │   │   │       ├── modbus/          # Modbus驱动
│   │   │   │       ├── snmp/            # SNMP驱动
│   │   │   │       ├── sysdio/          # SysDIO驱动
│   │   │   │       └── simulator/       # 模拟器驱动
│   │   │   └── model/                   # 设备模型
│   │   ├── dispatcher/                  # 任务调度
│   │   │   └── dispatcher.go            # 调度器（设备分配到WorkerChannel）
│   │   ├── worker/                      # 工作线程
│   │   │   ├── worker.go                # WorkerChannel实现
│   │   │   ├── pool.go                  # 采集池
│   │   │   └── template_manager.go      # 模板协议管理
│   │   ├── rtdb/                        # 实时数据库
│   │   │   ├── rtdb.go                  # RTDB接口
│   │   │   ├── db.go                    # RTDB实现
│   │   │   └── model/                   # 数据模型
│   │   └── processor/                   # 数据处理器
│   ├── distribution/                    # 数据分发
│   │   ├── init.go                      # 分发模块初始化
│   │   ├── interval/                    # 周期上报
│   │   │   ├── interval.go              # 周期处理器
│   │   │   └── manager.go               # 管理器
│   │   ├── change/                      # 变化上报
│   │   │   └── collect_manager.go       # 变化管理器
│   │   ├── distributor/                 # 分发器
│   │   │   ├── kafka/                   # Kafka分发
│   │   │   │   ├── kafka.go             # Kafka分发器
│   │   │   │   └── retry.go             # 重试逻辑
│   │   │   └── http/                    # HTTP分发
│   │   │       ├── http.go              # HTTP分发器
│   │   │       └── client.go            # HTTP客户端
│   │   └── base/                        # 基础分发逻辑
│   ├── std/                             # 标准测点计算
│   │   ├── init.go                      # 初始化
│   │   ├── calculator.go                # 计算器（表达式求值）
│   │   ├── mapping.go                   # 映射规则
│   │   └── report.go                    # 上报逻辑
│   ├── cm/                              # 配置管理
│   │   ├── init.go                      # 初始化
│   │   ├── worker.go                    # 配置工作器
│   │   ├── save.go                      # 配置保存
│   │   └── utils/                       # 工具类
│   ├── network/                         # 网络管理
│   │   ├── network.go                   # 网络配置
│   │   ├── switch.go                    # 交换机模式
│   │   └── utils/                       # 工具类
│   ├── plugin/                          # 插件系统
│   │   ├── plugins.go                   # 插件管理器
│   │   ├── notify.go                    # 事件通知
│   │   ├── init/                        # 插件初始化
│   │   ├── expressioncmd/               # 表达式命令插件
│   │   └── southdevice/                 # 南向设备插件
│   ├── cgi/                             # CGI接口实现
│   │   ├── rtd.go                       # 实时数据接口
│   │   ├── device.go                    # 设备接口
│   │   └── qua.go                       # 质量查询接口
│   ├── hmac/                            # 鉴权
│   │   └── auth.go                      # HMAC鉴权
│   ├── task/                            # 任务管理
│   │   └── heartbeat.go                 # 心跳上报（schedule模式）
│   ├── debug/                           # 调试
│   └── logfile/                         # 日志文件
├── service/                             # 服务接口
│   ├── cgi.go                           # CGI服务（设备列表、实时数据、调试等）
│   ├── cm.go                            # 配置管理服务
│   ├── box.go                           # 盒子管理服务（NTP、重启等）
│   ├── rtd.go                           # 实时数据服务（北向推送配置）
│   └── device.go                        # 设备服务
├── repo/                                # 数据访问层
│   ├── cm/                              # 配置存储
│   │   ├── api.go                       # 接口定义
│   │   ├── localfile/                   # 本地文件读取
│   │   ├── tlink/                       # tlink读取
│   │   ├── taskserver/                  # 任务服务读取
│   │   └── backup/                      # 备份读取
│   └── monitor/                         # 监控上报
├── utils/                               # 工具类
│   ├── byteorder/                       # 字节序处理
│   ├── bytes/                           # 字节操作
│   ├── encoding/                        # 编解码
│   ├── file/                            # 文件操作
│   ├── ghttp/                           # HTTP工具
│   ├── osal/                            # 操作系统抽象层
│   │   ├── variant.go                   # 变体类型
│   │   └── queue/                       # 队列实现
│   ├── parse/                           # 解析工具
│   └── network/                         # 网络工具
└── project/                             # 项目配置目录
    └── default/                         # 默认项目配置
        ├── devices@xxx.json             # 设备配置
        ├── std@xxx.json                 # 标准点配置
        ├── std_device.json              # 标准设备配置
        └── templates/                   # 模板目录
```

## 配置文件

### trpc_go.yaml 主要配置项

```yaml
# 服务配置
server:
  service:
    - name: agent
      ip: 0.0.0.0
      port: 61000
      protocol: http
    - name: agent-restful
      port: 61001
      protocol: restful

# 功能开关
feature:
  standard_calculation: 1    # 标准测点计算开关
  collect_report: 1          # 采集上报开关
  simulation: 1              # 模拟器开关
  devs_local: 1              # 本地设备配置模式
  backup_push: 0             # 备份推送开关

# 采集配置
collector:
  common:
    packet_failed_count: 2       # 报文失败计数阈值
    request_failed_count: 7      # 请求失败计数阈值（通讯中断判断）
    request_failed_time: 60      # 请求失败时间阈值(秒)
  snmp:
    timeout: 2000
    retry: 1
  modbus:
    serials_map:                 # 串口映射配置
      COM1:
        dev: /usr/dev/serial/com1
        baud: "9600"
        # ...

# 分发配置
distributor:
  common:
    interval_report_second: 0    # 周期上报基准秒数（0为随机）
  kafka:
    brokers: ["localhost:9092"]
    topic:
      points: "tbos_point_data"
    sasl:
      mechanism: ""
      username: ""
      password: ""
  http:
    enable:
      - collect_change
      - collect_interval
    north_whitelist: []          # 北向白名单

# 插件配置
plugin:
  interruption_judge_threshold: 50   # 通讯中断判断阈值(%)
  plugin_call_interval: 30           # 插件调用间隔(秒)

# 项目配置
project:
  mode: agent                    # 运行模式：agent 或 agent-gw
  module_group: default          # 模组ID
  source: tlink                  # 配置来源

# 任务配置
task:
  mode: local                    # 任务模式：local 或 schedule
  local:
    devs:                        # 本地模式设备列表
      - device_gid_1
      - device_gid_2

# TBox配置
tbox:
  heartbeat_interval_ms: 5000    # 心跳间隔(ms)
```

## API 接口

### CGI 服务 (AgentCgiService)

| 接口 | 方法 | 说明 |
|------|------|------|
| `/cgi/devices` | GET | 获取设备列表 |
| `/cgi/rtd` | POST | 获取实时数据 |
| `/cgi/rtdById` | POST | 按ID获取实时数据 |
| `/cgi/setRtdById` | POST | 设置实时数据（调试用） |
| `/cgi/intervalPoints` | GET | 获取周期上报测点 |
| `/cgi/qua` | POST | 获取测点质量信息 |
| `/cgi/devicesCommste` | GET | 获取设备通讯状态 |
| `/cgi/exprValidate` | POST | 表达式校验 |
| `/cgi/startupProbe` | GET | 启动探测 |
| `/cgi/debug` | POST | 调试接口 |

### 实时数据服务 (RealTimeDataManagerService)

| 接口 | 方法 | 说明 |
|------|------|------|
| `/north/set` | POST | 设置消息推送参数（HTTP分发） |
| `/north/get` | GET | 获取消息推送参数 |
| `/north/online_strategy_push` | POST | 在线策略推送 |

### 盒子管理服务 (BoxManagerService)

| 接口 | 方法 | 说明 |
|------|------|------|
| `/box/setNtp` | POST | 设置NTP服务器 |
| `/box/setRealTime` | POST | 设置系统时间 |
| `/box/osRestart` | POST | 重启操作系统 |
| `/box/agentRestart` | POST | 重启Agent服务 |

## 常见问题

### 1. 设备采集不到数据

**问题描述**：设备添加后，测点数据一直显示未采集状态（质量码 -403）

**排查步骤**：
1. 检查设备配置是否正确（通道地址、设备地址等）
2. 检查串口/网络连接是否正常
3. 查看日志确认设备是否成功加载
4. 使用调试接口检查采集状态

**解决方案**：
- Modbus：确认 slave_id、寄存器地址、数据类型配置正确
- SNMP：确认 community、OID 配置正确
- 检查防火墙设置，确保端口可访问

### 2. 测点质量显示通讯中断（-200）

**问题描述**：测点质量持续显示通讯中断

**原因分析**：
- 连续请求失败次数超过阈值（默认7次）
- 失败持续时间超过配置时间（默认60秒）

**解决方案**：
- 检查设备是否在线
- 检查网络/串口连接
- 调整 `request_failed_count` 和 `request_failed_time` 配置
- 查看日志中的具体错误信息

### 3. 配置更新后不生效

**问题描述**：修改配置文件后，采集行为没有变化

**解决方案**：
- 调用配置管理接口触发配置重载
- 检查配置文件格式是否正确
- 查看日志确认配置是否成功加载
- 必要时重启 Agent 服务

### 4. Kafka 发送失败

**问题描述**：数据无法上报到 Kafka

**排查步骤**：
1. 检查 Kafka 服务是否正常运行
2. 确认 Kafka broker 地址配置正确
3. 检查网络连接和防火墙
4. 如配置了 SASL 认证，确认用户名密码正确

**解决方案**：
- Kafka 发送失败时会自动重试（HTTP 回退）
- 检查配置中的 `max_attempt` 和 `write_timeout` 设置

### 5. 标准测点计算结果异常

**问题描述**：标准测点值与预期不符

**排查步骤**：
1. 检查映射表达式是否正确
2. 确认关联的采集测点数据是否正常
3. 检查采集测点的质量标签

**解决方案**：
- 使用 `/cgi/exprValidate` 接口校验表达式
- 标准点会继承采集点的质量状态，确保采集点数据正常
- 检查表达式中的变量是否都有对应的采集点

### 6. 内存占用持续增长

**问题描述**：Agent 运行一段时间后内存占用过高

**可能原因**：
- RTDB 缓存数据量过大
- 日志缓冲区过大
- Kafka 消息积压

**解决方案**：
- 检查测点数量是否超出预期
- 调整日志级别，减少日志输出
- 确保 Kafka 连接正常，避免消息积压
- 定期重启服务释放内存

### 7. 串口通信异常

**问题描述**：Modbus RTU 设备采集失败

**排查步骤**：
1. 确认串口设备路径是否正确（如 `/usr/dev/serial/com1`）
2. 检查波特率、数据位、停止位、校验位配置
3. 确认 RS485 模式配置正确

**解决方案**：
```yaml
# 确认串口配置
collector:
  modbus:
    serials_map:
      COM1:
        baud: "9600"        # 波特率
        databit: "8"        # 数据位
        stopbit: "1"        # 停止位
        parity: "N"         # 校验位: N(无)/E(偶)/O(奇)
        mode: "485"         # RS485模式
        dev: /usr/dev/serial/com1  # 设备路径
```

### 8. Agent 启动失败

**问题描述**：Agent 启动时报错退出

**常见原因**：
- 配置文件格式错误
- 端口被占用
- 依赖服务不可用

**解决方案**：
- 检查 `trpc_go.yaml` 格式
- 确认 61000、61001 等端口未被占用
- 查看启动日志定位具体错误

## 部署说明

### 编译

```bash
# 编译
go build -o agent main.go
```

### 运行

```bash
# 运行
./agent -conf trpc_go.yaml
```

### Docker 部署

```bash
# 构建镜像
docker build -t tbos-agent .

# 运行容器（需要串口访问权限）
docker run -d --name agent \
  -v /dev:/dev \
  -v /path/to/config:/app/config \
  --privileged \
  -p 61000:61000 \
  -p 61001:61001 \
  tbos-agent
```

注意：需要 `--privileged` 权限访问串口设备。
