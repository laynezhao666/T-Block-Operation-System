这个快速开始手册用于指导开发者在单机上从零完成 TBOS 全栈部署，覆盖环境准备、配置、编译、启动与关键组件说明。

**注意**: 单机部署适合开发调试、功能验证和集成测试场景。所有服务部署在同一台主机上，不具备高可用能力，不建议直接用于生产环境。

## 1. 环境准备

### 1.1. 语言环境

TBOS 后端使用 Go 编写，前端使用 Node.js 构建，部署前请确保已安装以下语言的对应版本。

| 语言 | 版本要求 | 说明 | 验证命令 |
|------|---------|------|---------|
| Go | 1.22+ | 用于编译所有后端服务（cgi、scheduler、agent 等 12 个模块） | `go version` |
| Node.js | 14+ (推荐 16+) | 用于构建前端应用（web 模块） | `node -v` |

> 建议使用 [nvm](https://github.com/nvm-sh/nvm)（Node.js）和 [gvm](https://github.com/moovweb/gvm)（Go）管理多版本语言环境，避免与系统自带版本冲突。

### 1.2. 组件版本

| 组件 | MySQL | Redis | Kafka | InfluxDB(可选) |
|------|-------|-------|-------|----------|
| 版本要求 | 5.7+ | 6.0+ | 2.8+ | 1.8+ |

### 1.3. 硬件要求

| 配置项 | CPU | 内存 | 存储 |
|--------|-----|------|------|
| 最低要求 | 8 核心 | 32 GB | 200 GB |

> 单模组总资源不低于52核心，资源规划建议 CPU:内存 = 1:4。

### 1.4. 操作系统

- Ubuntu 16.04 / 18.04 LTS (64-bit) 以上
- CentOS Linux 7.6 (64-bit) 以上
- Tencent Linux 2.2 以上

## 2. 下载 TBOS

```bash
git clone https://github.com/Tencent/T-Block-Operation-System.git
cd T-Block-Operation-System
```

## 3. 配置 server.cfg

编辑项目根目录下的 `server.cfg` 文件，按实际环境配置各组件连接参数和服务端口：

```bash
# 以下所有配置项中的 <...> 占位符均需手动替换为实际值

# 数据库配置
MYSQL_HOST=<mysql_host>          # 请替换为实际的 MySQL 地址
MYSQL_PORT=3306
MYSQL_USER=<mysql_user>          # 请替换为实际的 MySQL 用户名
MYSQL_PASSWORD=<mysql_password>  # 请替换为实际的 MySQL 密码
MYSQL_DATABASE=tbos
MYSQL_ADDR=${MYSQL_HOST}:${MYSQL_PORT}

# Redis 配置
REDIS_HOST=<redis_host>          # 请替换为实际的 Redis 地址
REDIS_PORT=6379
REDIS_PASSWORD=<redis_password>  # 请替换为实际的 Redis 密码（无密码则留空）
REDIS_ADDR=${REDIS_HOST}:${REDIS_PORT}

# Kafka 配置
KAFKA_HOST=<kafka_host>          # 请替换为实际的 Kafka 地址
KAFKA_PORT=9092
KAFKA_ADDR=${KAFKA_HOST}:${KAFKA_PORT}
KAFKA_POINT_TOPIC=tbos_point_data
KAFKA_ALARM_TOPIC=tbos_alarm_msg

# 本机 IP
LOCAL_IP=127.0.0.1

# 服务端口
PORT_CGI=9111
PORT_SCHEDULER=9103
PORT_ALARM_COMPUTE=9108
PORT_ALARM_MANAGE=9109
PORT_DATA_CACHE=9105
PORT_DATA_STORE=9106
PORT_ALARM_SERVER=9110
PORT_CMDB=9102
PORT_COLLECTOR=9101
PORT_DATA_COMPUTE=9104
PORT_DATA_QUERY=9107
PORT_AGENT=9100
```

上方的 `server.cfg` 已为各服务预置了默认端口（见"服务端口"区域）。`tbos.sh start` 会依次启动以下 12 个服务，端口号可按实际环境自行调整：

| 服务 | 默认端口 | 职责 |
|------|---------|------|
| cgi | 9111 | API 网关，对接前端与 WebSocket 推送 |
| scheduler | 9103 | 任务调度中心，负载均衡分配采集/计算/告警任务 |
| alarm-compute | 9108 | 告警计算引擎，运行 TNQL 表达式 |
| alarm-manage | 9109 | 告警去重、入库与生命周期管理 |
| data-cache | 9105 | 消费 Kafka 测点数据，内存实时缓存 |
| data-store | 9106 | 测点数据持久化到 InfluxDB |
| alarm-server | 9110 | 告警查询、统计与趋势分析 |
| cmdb | 9102 | 设备、测点、模板、策略的统一配置管理 |
| collector | 9101 | 配置收集与测点数据转发到 Kafka |
| data-compute | 9104 | 虚拟测点表达式计算 |
| data-query | 9107 | 实时/历史测点数据查询 |
| agent | 9100 | 多协议设备采集与边缘网关 |

## 4. 本地启动

```bash
# 1. 编译所有服务
./tbos.sh build

# 2. 初始化数据库
mysql -u root -p < ddl.sql

# 3. 启动所有服务
./tbos.sh start

# 4. 停止所有服务
./tbos.sh stop
```

`tbos.sh build` 将所有服务编译到 `target/` 目录，`tbos.sh start` 以后台进程拉起重置全部服务。启动日志写入 `target/<服务名>/server.log`，可用于排查启动异常。


### 4.1. 关键组件说明

#### 4.1.1. MySQL

上一步中执行 `mysql -u root -p < ddl.sql` 完成数据库初始化后，TBOS 会创建以下核心表：

| 表名 | 说明 |
|------|------|
| `t_alarm_active` | 活动告警表，存储当前未恢复的告警 |
| `t_alarm_history` | 历史告警表，存储已恢复的告警 |
| `t_alarm_strategy` | 告警策略表，包含告警/恢复表达式、级别、内容模板 |
| `t_alarm_worker` | 告警 Worker 表，用于雪花算法分布式协调 |
| `t_collector_device` | 采集设备表，包含设备 GID、通道、模板等 |
| `t_collector_template` | 采集模板表，包含协议类型、版本、设备型号 |
| `t_collector_template_point` | 模板测点表，包含测点定义、协议定义 |
| `t_device_entity` | 设备实体表，包含设备 GID、名称、所属模组 |
| `t_device_point` | 设备测点表，包含测点表达式、映射关系 |
| `t_mozu_info` | 模组信息表，包含模组 ID、名称、发布版本 |


#### 4.1.2. Redis

Redis 在 TBOS 中主要用于分布式锁，支撑告警 ID 的全局唯一性保障。

`alarm-manage` 服务在启动时，会基于雪花算法为自身抢占一个唯一的工作节点（SubnodeID）。抢占过程通过 Redis SETNX 原子操作实现分布式互斥，key 格式为 `alarm_snowflake_{subnode}`，过期时间 10 分钟。抢占成功后，服务每小时向 MySQL 更新一次心跳；服务退出时自动释放节点。若所有节点均被占用，服务将启动失败。

#### 4.1.3. Kafka

TBOS 使用 Kafka 作为测点数据和告警消息的传输总线，涉及以下两条数据链路：

**测点数据链路**

| 角色 | 服务 | Topic |
|------|------|-------|
| 生产者 | collector | `tbos_point_data`（`KAFKA_POINT_TOPIC`） |
| 消费者 | data-store | `tbos_point_data` |
| 消费者 | data-cache | `tbos_point_data` |

`collector` 接收 Agent 上报的测点数据后，将其写入 `tbos_point_data` Topic。`data-store` 订阅该 Topic，将数据持久化到 InfluxDB；`data-cache` 同样订阅该 Topic，将近期数据保存在内存中供实时查询。

**告警消息链路**

| 角色 | 服务 | Topic |
|------|------|-------|
| 生产者 | alarm-compute | `tbos_alarm_msg`（`KAFKA_ALARM_TOPIC`） |
| 消费者 | alarm-manage | `tbos_alarm_msg` |
| 生产者（通知） | alarm-manage | CGI 同步 Topic |

`alarm-compute` 计算出告警或恢复事件后写入 `tbos_alarm_msg` Topic，`alarm-manage` 消费后完成去重、入库和生命周期管理，并将结果再次写入 Kafka 供 CGI 推送给前端。

#### 4.1.4. InfluxDB

InfluxDB 用于长周期测点历史数据的持久化存储。`data-store` 消费 Kafka 测点数据后，通过 InfluxDB 存储插件将数据写入时序数据库，默认保留 100 天。

TBOS 的测点数据查询链路为：`data-cache`（内存缓存，覆盖近期数据）→ `data-query`（缓存未命中时回退到 InfluxDB）。如果业务场景中不需要查询长周期历史数据（如仅需实时监控或短期回溯），可以不部署 InfluxDB，仅依赖 `data-cache` 的内存缓存即可满足需求。

> InfluxDB 详细配置（存储天数、批次大小、写入模式等）将移至 [部署参考](#414-influxdb) 章节单独说明。

## 5. 服务

### 5.1 服务配置文件

每个服务目录下的`trpc_go.yaml`即为服务配置文件，包含：

| 配置项 | 说明 |
| --- | -- |
| etrpc | 服务名称、端口配置 |
| global | 全局配置（命名空间、本机IP等） |
| server.service | 服务监听配置（协议、端口） |
| client.service | 客户端配置（数据库、Redis、Kafka、RPC调用）|
| plugins | 插件配置（日志、监控等） |

### 5.2 服务端端口分配

| 服务 | 默认端口 | 协议 | 说明 |
| -- | ---- | -- | -- |
| CGI | 9111 | HTTP | API网关服务 |
| Scheduler | 9103 | HTTP | 调度中心 |
| Alarm-Compute | 9108 | HTTP | 告警计算引擎 |
| Alarm-Manage | 9109 | HTTP | 告警管理服务 |
| Data-Cache | 9105 | HTTP | 数据缓存服务 |
| Data-Store | 9106 | HTTP | 数据存储服务 |
| Alarm-Server | 9110 | HTTP | 告警服务 |
| CMDB | 9102 | HTTP | 配置管理服务 |
| Collector | 9101 | HTTP | 数据收集器 |
| Data-Compute | 9104 | HTTP | 数据计算服务 |
| Data-Query | 9107 | HTTP | 数据查询服务 |
| Agent | 9100 | HTTP | 边缘采集器 |
| Web前端 | 8080 | HTTP | 前端界面(Nginx) |
