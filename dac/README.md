# DAC（门禁控制服务）

## 一、模块介绍

DAC（DoorAccessControl）是TBOS系统的门禁控制服务，负责对门禁控制器的统一管理、多协议驱动适配、实时数据采集、门禁卡权限管理、事件告警等功能。适用于数据中心、楼宇、园区等场景的门禁管理需求。

### 1.1 主要职责

| 职责 | 说明 |
|------|------|
| 门禁控制器管理 | 控制器的增删改查、批量导入导出、时间同步、消防复位 |
| 多协议驱动适配 | 插件化驱动设计，内置 CACS、CHD806D4、XBrother、HTTP 四种协议 |
| 门禁卡与权限管理 | 卡片授权、人员管理、权限组配置、时间组通行策略 |
| 实时数据采集 | 自动采集门状态、通行事件、告警信息，支持增量同步 |
| 事件与告警 | 实时采集通行事件与告警信息，支持按时间/控制器查询与导出 |
| 异步请求管理 | 门禁操作请求的全生命周期管理，支持重试和批量重执行 |

### 1.2 数据流向

```
┌─────────────────────────────────────────────────────────────────────┐
│                        外部客户端 / 前端页面                          │
│                   HTTP 请求 (Gin)                                    │
│                   /api/dcos/tdac-cgi/*                               │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          DAC 服务                                    │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    CGI 业务逻辑层                              │   │
│  │  controller │ door │ access │ event │ alarm │ request │ point │   │
│  └──────────────────────────┬───────────────────────────────────┘   │
│                             │                                        │
│  ┌──────────────────────────▼───────────────────────────────────┐   │
│  │                    核心服务层                                  │   │
│  │  cache │ card │ dlm │ mapping │ push │ timegroup │ request    │   │
│  └──────────────────────────┬───────────────────────────────────┘   │
│                             │                                        │
│  ┌──────────────────────────▼───────────────────────────────────┐   │
│  │                    数据采集引擎                                │   │
│  │  dispatcher ──► worker ──► driver ──► 门禁控制器设备           │   │
│  │                              │                                │   │
│  │              ┌───────────────┼───────────────┐                │   │
│  │              │               │               │                │   │
│  │           CACS(TCP)    HTTP(v1/v2/v3)   XBrother(TCP)        │   │
│  │                         CHD806D4                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                             │                                        │
│  ┌──────────────────────────▼───────────────────────────────────┐   │
│  │                    数据存储层                                  │   │
│  │           MySQL (GORM)  │  Redis  │  Kafka (可选)             │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 二、核心能力

### 2.1 多协议驱动架构

采用插件化驱动设计，通过 Go 的 `init()` 函数自动注册驱动，支持自定义扩展：

| 驱动 | 协议 | 通信方式 | 目录 |
|------|------|---------|------|
| CACS | CACS | TCP 被动监听 | `logic/collect/driver/cacs/` |
| CHD806D4 | CHD806D4 | TCP | `logic/collect/driver/chd806d4/` |
| XBrother | XBrother | TCP 主动拨号 | `logic/collect/driver/xbrother/` |
| HTTP | HTTP | HTTP (v1/v2/v3/MDC) | `logic/collect/driver/http/` |
| Test | test | 测试用 | `logic/collect/driver/test/` |

### 2.2 数据采集引擎

采集引擎采用 dispatcher → worker → driver 三级架构：

- **Dispatcher**：任务调度器，将控制器分配到工作协程
- **Worker**：工作协程池，管理采集任务的并发执行
- **Driver**：协议驱动，负责与门禁控制器设备通信
- **Delta**：增量同步，仅同步变化的数据

### 2.3 门禁卡与权限管理

- 卡片管理：增删改查、批量导入导出、启用/禁用、有效期管理
- 人员管理：人员信息维护、卡片绑定/解绑
- 权限组：权限组配置、门与权限组关联
- 时间组：灵活的通行时段配置，支持按星期和时间段组合

### 2.4 异步请求管理

门禁操作（如开门、同步卡片等）通过异步请求管理，支持：
- 请求全生命周期管理（创建、执行、成功/失败、过期）
- 失败重试与批量重执行
- 请求记录查询与导出
- 自动清理过期请求

### 2.5 分布式锁

基于 Redis + Redsync 的分布式锁，支持多 Pod 部署时任务不重复执行。

### 2.6 数据推送

支持将门禁数据（事件、告警等）推送到 Kafka，供下游系统消费。

---

## 三、代码结构

```
dac/
├── main.go                 # 程序入口
├── conf/                   # 配置文件
│   └── trpc_go.yaml        # tRPC 框架 + 业务配置（支持环境变量注入）
├── entity/                 # 实体层（模型、常量、工具）
│   ├── config/             # 配置管理（热更新支持）
│   ├── consts/             # 全局常量（协议名、服务名、列名等）
│   ├── model/              # 数据模型
│   │   ├── db/             # 数据库模型（GORM）
│   │   ├── cgi/            # CGI 请求/响应模型
│   │   ├── driver/         # 驱动层接口与数据模型
│   │   └── rt/             # 实时数据模型
│   ├── server/cgi/         # HTTP 路由定义（Gin）
│   ├── redis/              # Redis 客户端与分布式锁
│   ├── location/           # 地理位置/时区管理
│   └── utils/              # 工具函数
├── logic/                  # 业务逻辑层
│   ├── setup/              # 服务初始化与退出清理
│   ├── collect/            # 数据采集模块
│   │   ├── controller/     # 控制器管理
│   │   ├── driver/         # 协议驱动实现
│   │   │   ├── cacs/       # CACS 协议（TCP 被动监听）
│   │   │   ├── chd806d4/   # CHD806D4 协议
│   │   │   ├── xbrother/   # XBrother 协议（TCP 主动拨号）
│   │   │   ├── http/       # HTTP 通用协议（v1/v2/v3/MDC）
│   │   │   └── test/       # 测试驱动
│   │   ├── dispatcher/     # 任务调度器
│   │   ├── worker/         # 工作协程池
│   │   ├── delta/          # 增量同步
│   │   └── template/       # 采集模板
│   ├── cgi/                # CGI 业务逻辑
│   ├── cache/              # 缓存管理
│   ├── card/               # 门禁卡管理
│   ├── controller/         # 控制器业务逻辑
│   ├── door/               # 门点管理
│   ├── push/               # 数据推送（Kafka）
│   ├── dlm/                # 分布式锁管理
│   ├── mapping/            # 编码映射（GID）
│   ├── request/            # 异步请求管理
│   └── timegroup/          # 时间组管理
├── repo/                   # 数据访问层
│   ├── dac/                # 门禁数据库操作（GORM）
│   ├── data/               # 外部数据源
│   ├── pb/                 # Protobuf 定义
│   └── redis/              # Redis 数据操作
├── etrpc-go/               # 基于 tRPC-Go 的扩展框架（本地模块）
├── Dockerfile              # Docker 多阶段构建文件
└── docker-compose.yml      # Docker Compose 一键编排
```

### 核心文件说明

| 文件 | 说明 |
|------|------|
| `main.go` | 服务入口，初始化数据库、Redis、注册路由、启动采集引擎 |
| `entity/config/config.go` | 业务配置结构定义，支持热更新 |
| `entity/server/cgi/router.go` | HTTP 路由定义，注册所有 API 接口 |
| `logic/setup/` | 服务初始化与退出清理逻辑 |
| `logic/collect/dispatcher/` | 采集任务调度器，分配控制器到工作协程 |
| `logic/collect/driver/` | 协议驱动实现，插件化注册 |
| `logic/cgi/` | CGI 业务逻辑，处理 HTTP 请求 |
| `repo/dac/` | 数据库操作层，基于 GORM |

---

## 四、API 接口

服务启动后，所有 API 的基础路径为 `/api/dcos/tdac-cgi`。

### 控制器管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/controllers` | 获取控制器列表 |
| `POST` | `/controller` | 创建控制器 |
| `PUT` | `/controller` | 更新控制器 |
| `DELETE` | `/controller` | 删除单个控制器 |
| `DELETE` | `/controllers` | 批量删除控制器 |
| `POST` | `/controllers/import` | 批量导入控制器（Excel） |
| `GET` | `/controllers/export` | 导出控制器（Excel） |
| `POST` | `/controllers/sync-time` | 批量同步时间 |
| `POST` | `/controllers/reset` | 批量消防复位 |
| `POST` | `/controller/clean` | 清除控制器数据 |
| `POST` | `/controller/reset` | 单个消防复位 |
| `POST` | `/controller/sync-time` | 单个同步时间 |
| `DELETE` | `/controller/time-groups` | 清除控制器时间组 |
| `POST` | `/controller/card` | 从控制器查询卡是否存在 |

### 门点管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/door` | 获取门点信息 |
| `PUT` | `/door` | 更新门点 |
| `POST` | `/door/state` | 设置门状态（开门/关门/常开） |
| `PUT` | `/door/code` | 更新门点编码 |
| `PUT` | `/doors` | 批量更新门点 |
| `GET` | `/doors/export/code` | 导出门点编码 |
| `POST` | `/doors/import/code` | 导入门点编码 |

### 门组管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/groups` | 获取门组列表 |
| `POST` | `/group` | 创建门组 |
| `PUT` | `/group` | 更新门组 |
| `DELETE` | `/group` | 删除门组 |
| `POST` | `/group/doors` | 获取门组下的门列表 |

### 实时数据

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/rtd` | 查询单个实时数据 |
| `POST` | `/rtd/list` | 批量查询实时数据 |

### 事件与告警

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/events` | 查询通行事件 |
| `POST` | `/events/export` | 导出通行事件 |
| `POST` | `/doors/events` | 按门查询事件 |
| `POST` | `/alarms` | 查询告警记录 |
| `POST` | `/alarms/export` | 导出告警记录 |

### 异步请求管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/requests` | 按控制器查询请求 |
| `PUT` | `/requests/update` | 更新请求状态 |
| `POST` | `/requests/outdate` | 标记请求过期 |
| `DELETE` | `/requests` | 删除请求 |
| `GET` | `/requests/all` | 获取所有请求 |
| `POST` | `/requests/info` | 获取请求详情 |
| `POST` | `/requests/all` | 获取所有请求详情 |
| `GET` | `/requests/methods` | 获取支持的请求方法列表 |
| `POST` | `/requests/export` | 导出请求 |
| `GET` | `/requests/export/all` | 导出所有请求 |
| `POST` | `/requests/re-execute` | 重新执行请求 |
| `POST` | `/requests/batch-re-execute` | 批量重新执行请求 |

### 门禁权限管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/time-groups` | 获取时间组列表 |
| `PUT` | `/time-group/:group_no` | 更新时间组 |
| `POST` | `/time-groups/sync` | 同步时间组到控制器 |
| `GET` | `/staffs` | 获取人员列表 |
| `GET` | `/staffs/company` | 获取所有人员公司 |
| `POST` | `/staffs` | 添加人员 |
| `POST` | `/staffs/import` | 导入人员 |
| `POST` | `/staffs/export` | 导出人员 |
| `PUT` | `/staff/:id` | 更新人员 |
| `DELETE` | `/staff/:id` | 删除人员 |
| `POST` | `/cards` | 查询卡片列表 |
| `POST` | `/card` | 添加卡片 |
| `DELETE` | `/cards` | 批量删除卡片 |
| `POST` | `/cards/import` | 导入卡片 |
| `POST` | `/cards/export` | 导出卡片 |
| `POST` | `/cards/sync` | 同步卡片到控制器 |
| `PUT` | `/card/flag` | 更新卡片启用/禁用状态 |
| `PUT` | `/card/type` | 更新卡片类型 |
| `PUT` | `/card/valid_time` | 更新卡片有效期 |
| `DELETE` | `/card` | 删除单张卡片 |
| `PUT` | `/card/staff` | 绑定卡片与人员 |
| `PUT` | `/card/unbind` | 解绑卡片与人员 |
| `PUT` | `/card/access` | 更新卡片权限组 |
| `GET` | `/access-groups` | 获取权限组列表 |
| `POST` | `/access-groups` | 创建权限组 |
| `GET` | `/access-groups/card` | 获取所有权限组（含卡片） |
| `PUT` | `/access-group/:id` | 更新权限组 |
| `DELETE` | `/access-group/:id` | 删除权限组 |

### 其他

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/rooms` | 获取机房列表 |
| `POST` | `/test` | 测试接口（Ping） |
| `POST` | `/debug` | 调试接口 |

### 创建控制器示例

```bash
curl -X POST http://127.0.0.1:8080/api/dcos/tdac-cgi/controller \
  -H "Content-Type: application/json" \
  -H "mozuid: your_mozu_id" \
  -d '{
    "name": "一号门控器",
    "profile": {
      "vendor": "厂商名称",
      "model": "设备型号",
      "sn": "序列号"
    },
    "position": {
      "room": "机房A",
      "block": "A区",
      "no": "01",
      "mark": "",
      "desc": "一楼入口"
    },
    "channel": {
      "chid": "192.168.1.100:9999",
      "cmd_interval": "3000",
      "timeout": "5000"
    },
    "protocol": {
      "name": "http",
      "version": "v1"
    },
    "extend": {
      "door_num": 4,
      "url_mode": "0"
    },
    "account": "admin",
    "password": "123456"
  }'
```

---

## 五、配置说明

### 5.1 环境变量

所有敏感配置通过环境变量注入，配置文件 `conf/trpc_go.yaml` 中使用 `${VAR_NAME}` 引用：

| 环境变量 | 必填 | 说明 |
|---------|------|------|
| `PORT_DAC` | 是 | 服务监听端口 |
| `LOCAL_IP` | 是 | 服务绑定 IP（本地开发用 `127.0.0.1`） |
| `MYSQL_USER` | 是 | MySQL 用户名 |
| `MYSQL_PASSWORD` | 是 | MySQL 密码 |
| `MYSQL_ADDR` | 是 | MySQL 地址（格式：`host:port`） |
| `MYSQL_DATABASE` | 是 | MySQL 数据库名 |
| `REDIS_PASSWORD` | 是 | Redis 密码（无密码则留空） |
| `REDIS_ADDR` | 是 | Redis 地址（格式：`host:port`） |
| `KAFKA_ADDR` | 否 | Kafka 地址（不启用推送可不设置） |
| `KAFKA_POINT_TOPIC` | 否 | Kafka Topic 名称 |
| `POD_IP` | 否 | K8s 环境下的 Pod IP |

### 5.2 业务配置

在 `conf/trpc_go.yaml` 的 `dac` 节点下配置业务参数：

```yaml
dac:
  # 调试模式，开启后会输出更详细的日志（如通信报文）
  debug: false

  # 是否启用连接池模式（建议生产环境开启）
  enable_pooling: false

  # 是否上报数据到 TBOS 平台（独立部署时设为 false）
  report_to_tbos: false

  # 是否从 CMDB 同步控制器数据（独立部署时设为 false）
  sync_from_cmdb: false
  sync_gid_from_cmdb: false

  # 忽略 GID 映射的模组列表（本地测试时可添加 "test"）
  ignore_gid_mozus: ["test"]

  # 请求记录清理策略
  expiration_time: 1    # 请求过期天数
  deletion_time: 7      # 请求删除天数
```

完整配置项定义请参考 `entity/config/config.go` 中的 `Config` 结构体。

### 5.3 Redis 连接模式

Redis 客户端支持三种部署模式，通过连接字符串的 `mode` 参数指定：

```yaml
# 单机模式（默认）
target: redis://:password@127.0.0.1:6379/0

# 哨兵模式
target: redis://:password@sentinel1:26379,sentinel2:26379/0?mode=sentinel&master=mymaster

# 集群模式
target: redis://:password@node1:6379,node2:6379?mode=cluster
```

---

## 六、数据库设计

系统使用 GORM 自动建表，启动时会自动创建以下数据表：

| 表名 | 说明 |
|------|------|
| `t_dac_controller` | 门禁控制器 |
| `t_dac_door` | 门点信息 |
| `t_dac_door_group` | 门组 |
| `t_dac_event` | 通行事件记录 |
| `t_dac_alarm` | 告警记录 |
| `t_dac_card` | 门禁卡信息 |
| `t_dac_staff` | 人员信息 |
| `t_dac_time_group` | 时间组 |
| `t_dac_access_group` | 权限组 |
| `t_dac_access_group_relation` | 权限组与门的关联 |
| `t_card_access_relation` | 卡与权限组的关联 |
| `t_dac_request` | 异步请求记录 |
| `t_dac_driver_card` | 驱动层卡同步记录 |
| `t_dac_driver_event` | 驱动层事件同步记录 |
| `t_dac_driver_alarm` | 驱动层告警同步记录 |
| `t_dac_driver_time_group` | 驱动层时间组同步记录 |
| `t_dac_driver_door_parameter` | 驱动层门参数同步记录 |
| `t_dac_event_index` | 事件同步索引 |
| `t_dac_alarm_index` | 告警同步索引 |
| `t_dac_event_timestamp_index` | 事件时间戳同步索引 |
| `t_dac_alarm_timestamp_index` | 告警时间戳同步索引 |

---

## 七、部署说明

### 7.1 编译

```bash
cd tbos/dac
go build -o dac
```

### 7.2 运行

```bash
./dac
```

### 7.3 Docker Compose 部署

提供 Docker Compose 编排，自动拉起 MySQL、Redis、Kafka 和 DAC 服务：

```bash
cd tbos/dac
docker-compose up -d
```

默认配置：

| 服务 | 端口 | 默认密码 |
|------|------|---------|
| DAC API | 8080 | — |
| MySQL | 3306 | dac123456 |
| Redis | 6379 | dac123456 |
| Kafka | 9092 | — |

### 7.4 本地开发运行

```bash
# 1. 确保 MySQL 和 Redis 已启动
docker run -d --name dac-mysql -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=123456 \
  -e MYSQL_DATABASE=dac \
  mysql:8.0

docker run -d --name dac-redis -p 6379:6379 \
  redis:7-alpine redis-server --requirepass 123456

# 2. 设置环境变量
export PORT_DAC=8080
export LOCAL_IP=127.0.0.1
export MYSQL_USER=root
export MYSQL_PASSWORD=123456
export MYSQL_ADDR=127.0.0.1:3306
export MYSQL_DATABASE=dac
export REDIS_PASSWORD=123456
export REDIS_ADDR=127.0.0.1:6379

# 3. 下载依赖并运行
go mod tidy
go run main.go
```

> **注意**：项目依赖本地模块 `etrpc-go`（位于 `../etrpc-go`），请确保目录结构正确。

---

## 八、常见问题

### 8.1 时区处理

**问题描述**：门控器设备上报的时间均为北京时间（UTC+8），而容器默认运行在 UTC 时区，直接使用 `time.Parse` 会导致 8 小时偏差。

**解决方案**：

```go
// 正确：使用 ParseInLocation 指定时区
loc, _ := time.LoadLocation("Asia/Shanghai")
t, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, loc)

// 错误：会按 UTC 解析
t, _ := time.Parse("2006-01-02 15:04:05", timeStr)
```

### 8.2 数据库事务中的网络 I/O

**问题描述**：在事务内部执行耗时的网络 I/O 操作（如调用外部接口）可能导致 MySQL 死锁。

**解决方案**：将网络请求移出事务范围，仅在事务内保留纯内存计算和 DB 读写。

### 8.3 HTTP 请求编码问题

**问题描述**：向门控器下发包含 Base64 编码数据（如人脸照片）时，使用 `application/x-www-form-urlencoded` 会导致 Base64 中的 `+` 号被错误解码为空格。

**解决方案**：必须使用 `application/json` 格式。

### 8.4 K8s 环境下的设备标识

**问题描述**：由于 SNAT 机制会将设备源 IP 转换为集群内部 IP，使用 IP 地址作为设备唯一标识会导致识别错误。

**解决方案**：使用设备的 MAC 地址或序列号进行身份绑定。

### 8.5 创建控制器报 GID 映射错误

**问题描述**：调用创建控制器接口返回 `get gid from codes error: dial tcp: lookup tadaptor-gidmapping: no such host`。

**解决方案**：
1. 在配置中将 `ignore_gid_mozus` 设置为 `["test"]`
2. 在请求头中使用 `mozuid: test`，以跳过 GID 映射服务的依赖

### 8.6 服务启动后查询到旧数据

**问题描述**：数据库已清空但接口仍返回旧数据。

**解决方案**：
1. 确认请求访问的是本服务端口，而非其他进程（如前端 dev server）占用的端口
2. 使用 `lsof -i :8080` 检查端口占用情况
3. 确认 `conf/trpc_go.yaml` 中的数据库连接地址正确

### 8.7 数据库连接失败

**问题描述**：服务启动时报数据库连接错误。

**解决方案**：
1. 检查环境变量配置：`MYSQL_USER`、`MYSQL_PASSWORD`、`MYSQL_ADDR`、`MYSQL_DATABASE`
2. 确认数据库服务正常运行
3. 检查网络连通性

### 8.8 如何扩展新的门禁协议驱动

1. 在 `logic/collect/driver/` 下创建新的驱动目录
2. 实现驱动接口
3. 在 `init()` 函数中注册驱动
4. 在控制器配置中指定新协议名称即可使用
