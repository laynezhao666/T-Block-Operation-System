# etrpc-go 微服务框架

etrpc-go 是基于 trpc-go 的微服务框架封装，提供了统一的配置加载、服务启动、数据库客户端封装、日志增强、指标上报等能力，简化微服务开发流程。

## 模块介绍

etrpc-go 是 TBOS 项目的基础框架层，主要职责包括：

- 封装 trpc-go 服务启动流程，增加配置校验和健康检查
- 提供统一的配置加载和热更新机制，支持环境变量替换和配置项注册
- 封装数据库客户端（Gorm、Redis、InfluxDB），提供连接池管理
- 提供日志增强功能，自动附加 Namespace、Env、Container、IP 等公共字段
- 提供指标上报封装，简化多维度指标上报
- 提供 HTTP 响应体统一包装过滤器

### 架构关系

```
┌─────────────────────────────────────────────────────────────────────┐
│                        业务服务层                                    │
│           (agent, scheduler, data-store, alarm-compute, ...)        │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          etrpc-go 框架层                             │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐        │
│  │  config   │  │  client   │  │    log    │  │  metric   │        │
│  │ 配置加载   │  │ 数据库客户端│  │  日志增强  │  │  指标上报  │        │
│  └───────────┘  └───────────┘  └───────────┘  └───────────┘        │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐        │
│  │  filter   │  │healthcheck│  │  database │  │   util    │        │
│  │ HTTP过滤器 │  │  健康检查  │  │InfluxDB层 │  │  工具类   │        │
│  └───────────┘  └───────────┘  └───────────┘  └───────────┘        │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          trpc-go 底层框架                            │
│                     (RPC、插件、编解码、服务发现...)                  │
└─────────────────────────────────────────────────────────────────────┘
```

## 核心能力

### 1. 服务启动与配置加载

#### 服务启动流程

通过 `etrpc.NewServer()` 创建服务器，封装了以下流程：

```go
// server.go
func NewServer(opt ...server.Option) *server.Server {
    // 1. 加载配置
    loader.LoadConfig()
    
    // 2. 校验配置：Service列表不允许为空
    if len(CfgTrpc.Server.Service) == 0 {
        panic("start etrpc server failed, no trpc service config found")
    }
    // 3. 校验配置：服务名称必须配置
    if len(CfgEtrpc.ServiceName) == 0 {
        panic("start etrpc server failed, config `etrpc.service_name` can not be empty")
    }
    
    // 4. 创建trpc服务器
    s := newServerWithCfg(CfgTrpc, opt...)
    return s
}
```

内部会执行：
- `trpc.RepairConfig()` 修复配置
- `trpc.SetupPlugins()` 初始化插件
- `trpc.SetupClients()` 初始化客户端
- `healthcheck.CheckClients()` 数据库连接健康检查
- 设置 `GOMAXPROCS` 适配容器环境

#### 配置加载机制

配置加载流程（`config/loader/loader.go`）：

```go
func LoadConfig() {
    // 1. 加载本地配置文件 (默认 ./trpc_go.yaml)
    fileCfg, err := file.Load(file.GetLocalConfigPath())
    
    // 2. 设置默认配置
    setDefaultConfig(fileCfg)
    
    // 3. 替换配置中的变量引用 ${xxx}
    fileCfg, err = util.ExpandEnv(fileCfg)
    
    // 4. 刷新配置缓存，触发注册对象的初始化
    cache.RefreshConfig(fileCfg)
}
```

默认配置值：
- `etrpc.service_port`: 8080
- `etrpc.trpc_service_port`: 8081
- `POD_IP`: 本机IP
- `POD_NAME`: "local-machine"
- `TRPC_NAMESPACE`: "Development"

### 2. 配置注册与热更新

#### 配置注册

支持在 `init()` 函数中注册配置对象，框架启动时自动加载：

```go
// 方式1：全量注册
config.RegisterConfig("cfgName", &YourConfig{}, hotUpdate)

// 方式2：带前缀注册（只加载指定前缀下的配置）
config.RegisterConfigWithPrefix("cfgName", "prefix.path", &YourConfig{}, hotUpdate)

// 方式3：完整配置（支持初始化和更新回调）
config.Register("cfgName", &YourConfig{},
    config.WithPrefix("prefix"),
    config.WithHotUpdate(true),
    config.WithInitFunc(func(val any) {
        // 首次加载时回调
    }),
    config.WithUpdateFunc(func(oldVal, newVal any) {
        // 热更新时回调
    }),
)
```

#### 配置获取API

提供多种类型的配置获取方法：

```go
// 获取任意类型
val, ok := config.Get("a.b.c")
val := config.GetOrDefault("a.b.c", defaultVal)

// 获取指定类型
intVal, ok := config.GetInt32("a.b.c")
intVal := config.GetInt32OrDefault("a.b.c", 0)

int64Val, ok := config.GetInt64("a.b.c")
float32Val, ok := config.GetFloat32("a.b.c")
float64Val, ok := config.GetFloat64("a.b.c")
strVal, ok := config.GetString("a.b.c")
boolVal, ok := config.GetBool("a.b.c")
timeVal, ok := config.GetTime("a.b.c")

// 加载到结构体
config.Load(&YourConfig{})
config.LoadWithPrefix(&YourConfig{}, "prefix.path")
```

#### 环境变量替换

支持在配置中使用 `${ENV_VAR}` 格式引用环境变量：

```yaml
server:
  port: ${PORT}              # 环境变量替换
  host: ${HOST:-localhost}   # 暂不支持默认值语法
  
etrpc:
  service_name: ${SERVICE_NAME}
  service_port: ${etrpc.service_port}  # 支持引用其他配置项
```

### 3. 数据库客户端封装

#### Gorm 客户端（MySQL/PostgreSQL/SQLite）

```go
import "etrpc-go/client/gorm"

// 获取数据库连接（自动缓存，支持链式调用）
db := gorm.GetDB("trpc.mysql.xxx.xxx")

// 创建新连接
db, err := gorm.NewClientProxy("trpc.mysql.xxx.xxx")
```

支持的数据库类型（根据 serviceName 中的类型字段判断）：
- `trpc.mysql.xxx` - MySQL
- `trpc.postgres.xxx` - PostgreSQL
- `trpc.sqlite.xxx` - SQLite
- `trpc.clickhouse.xxx` - ClickHouse

#### Redis 客户端

```go
import "etrpc-go/client/redis"

// 获取Redis连接
rdb := redis.GetRedis("trpc.redis.xxx.xxx")

// 创建新连接
rdb, err := redis.NewClientProxy("trpc.redis.xxx.xxx")
```

#### InfluxDB 客户端

```go
import "etrpc-go/client/influxdb"

// 获取InfluxDB连接
cli := influxdb.GetClient("trpc.influxdb.xxx.xxx")

// 创建新连接（自动执行Ping检测）
cli, err := influxdb.NewClientProxy("trpc.influxdb.xxx.xxx")

// 查询数据
rsp, err := cli.Query(ctx, influxdb.Query{
    Command:  "SELECT * FROM measurement WHERE time > now() - 1h",
    Database: "mydb",
})

// 写入数据
err := cli.Write(ctx, influxdb.BatchPoints{
    Database: "mydb",
    Points:   points,
})
```

InfluxDB 地址格式：
```
influxdb://user:password@host:port?timeout=5000&insecure_skip_verify=true
```

### 4. 健康检查

服务启动时自动检查数据库连接健康状态：

```go
// healthcheck/healthcheck.go
func CheckClients(config *trpc.Config) error {
    for _, client := range config.Client.Service {
        // 根据 serviceName 中的类型字段判断客户端类型
        // trpc.mysql.xxx -> MySQL
        // trpc.redis.xxx -> Redis
        // trpc.mongodb.xxx -> MongoDB
        // trpc.influxdb.xxx -> InfluxDB
    }
}
```

支持检查的客户端类型：
- MySQL/PostgreSQL/ClickHouse/SQLite（通过Gorm）
- Redis
- MongoDB
- InfluxDB

### 5. 日志增强

提供带上下文的日志方法，自动附加公共字段：

```go
import "etrpc-go/log"

// 普通日志
log.Debug("message")
log.Info("message")
log.Warn("message")
log.Error("message")
log.Fatal("message")

// 格式化日志
log.Debugf("format %s", args)
log.Infof("format %s", args)

// 带上下文的日志（自动附加TraceID等信息）
log.DebugContext(ctx, "message")
log.InfoContext(ctx, "message")
log.WarnContext(ctx, "message")
log.ErrorContext(ctx, "message")

// 带告警的日志（记录日志同时发送告警）
log.AlarmContext(ctx, "alarm message")
log.AlarmContextf(ctx, "alarm %s", args)

// Trace日志（需设置环境变量 TRPC_LOG_TRACE=1 启用）
log.Trace("trace message")
log.TraceContext(ctx, "trace message")
```

自动附加的公共字段：
- `Namespace`: 命名空间
- `Env`: 环境名称
- `Container`: 容器名称
- `IP`: 本机IP

### 6. 指标上报

简化多维度指标上报：

```go
import "etrpc-go/metric"

// 创建指标
m := metric.NewMetric("metric_name",
    metric.WithPolicy(metrics.PolicySUM, metrics.PolicyAVG),  // 聚合策略
    metric.WithDimensions(map[string]string{"dim1": "val1"}), // 默认维度
    metric.WithMetricGroup("group_name"),                     // 指标组
)

// 上报单个值
m.Report(100.0)

// 上报带维度的值
m.ReportWithDim(100.0, map[string]string{"extra_dim": "val"})

// 批量上报
m.ReportBatch([]float64{1.0, 2.0, 3.0})
```

支持的聚合策略：
- `metrics.PolicySUM` - 求和（默认）
- `metrics.PolicyAVG` - 平均值
- `metrics.PolicyMAX` - 最大值
- `metrics.PolicyMIN` - 最小值
- `metrics.PolicyMID` - 中位数

默认维度（除非使用 `WithoutDefaultDims()` 禁用）：
- `server`: 服务名
- `env`: 环境名
- `ip`: 本机IP
- `container`: 容器名
- `set_name`: Set名（启用Set路由时）

### 7. HTTP响应体包装

自动将 HTTP 响应包装为统一格式：

```go
// filter/rsp/filter.go
type responseEntity struct {
    Code    int32           `json:"code"`
    Message string          `json:"message"`
    Data    json.RawMessage `json:"data"`
    TraceId string          `json:"trace_id"`
}
```

配置控制：
```yaml
etrpc:
  disable_rsp_wrapper: false           # 全局禁用响应包装
  ignore_rsp_wrapper_path:             # 指定路径不包装
    - /api/health
    - /api/metrics
```

不进行包装的情况：
- 全局配置 `disable_rsp_wrapper: true`
- 请求路径在 `ignore_rsp_wrapper_path` 列表中
- 请求 Content-Type 为 `application/proto` 或 `application/pb`
- 请求 MetaData 中 `call_type` 为 `pb`

### 8. HTTP工具类

简化 HTTP 请求发送：

```go
import "etrpc-go/util/httputil"

// GET JSON请求
err := httputil.GetJson(ctx, "http://host/path", headers, &response)

// POST JSON请求
err := httputil.PostJson(ctx, "http://host/path", headers, &request, &response)

// 通用请求
rspHeader, err := httputil.Request(ctx, "POST", url, headers, &req, &resp, opts...)
```

支持的URL格式：
- `http://ip:port/path` - IP直连
- `http://domain/path` - 域名
- `http://service-name/path` - 北极星服务名

## 代码结构

```
etrpc-go/
├── etrpc.go                          # 框架配置定义（Etrpc结构体、全局配置变量）
├── server.go                         # 服务器启动逻辑（NewServer、RunServer）
├── go.mod                            # Go模块定义
├── alarm/                            # 告警模块
│   ├── alarm.go                      # Alarmer接口定义、GetAlarmClient工厂方法
│   └── default_alarm.go              # 默认告警实现（空实现）
├── client/                           # 数据库客户端封装
│   ├── client.go                     # 包定义
│   ├── gorm/                         # Gorm客户端
│   │   └── client.go                 # GetDB、NewClientProxy（支持MySQL/PG/SQLite）
│   ├── redis/                        # Redis客户端
│   │   └── client.go                 # GetRedis、NewClientProxy
│   └── influxdb/                     # InfluxDB客户端
│       └── client.go                 # GetClient、NewClientProxy
├── config/                           # 配置模块
│   ├── config.go                     # 配置注册和获取API
│   ├── cache/                        # 配置缓存
│   │   └── cache.go                  # RefreshConfig、Register、配置热更新逻辑
│   ├── loader/                       # 配置加载器
│   │   ├── loader.go                 # LoadConfig主函数、默认配置设置
│   │   └── file/                     # 文件加载器
│   │       └── loader.go             # Load本地配置文件
│   └── util/                         # 配置工具
│       └── config_util.go            # 环境变量替换、配置合并、GetByKey
├── database/                         # 数据库适配层
│   └── influxdb/                     # InfluxDB V1客户端实现
│       ├── client.go                 # Client接口（Ping/Query/Write/Close）
│       ├── config.go                 # 地址解析（ParseAddress）
│       ├── codec.go                  # 编解码器
│       ├── transport.go              # 传输层实现
│       └── service_name_extractor.go # 服务名提取
├── errs/                             # 错误定义
│   ├── errs.go                       # 错误类型
│   └── codec.go                      # 错误编解码
├── filter/                           # 过滤器
│   └── rsp/                          # 响应过滤器
│       └── filter.go                 # HTTP响应体统一包装
├── healthcheck/                      # 健康检查
│   └── healthcheck.go                # CheckClients数据库连接检查
├── log/                              # 日志模块
│   ├── log.go                        # 日志方法（Debug/Info/Warn/Error/Fatal/Alarm）
│   └── logger.go                     # Logger接口定义
├── metric/                           # 指标上报
│   ├── metric.go                     # IMetric接口、NewMetric、Report方法
│   └── metric_demo.go                # 使用示例
└── util/                             # 工具类
    ├── apputil/apputil.go            # 应用工具（获取执行路径等）
    ├── arrayutil/arrayutil.go        # 数组工具（Remove/Find/Filter/GroupBy/Map等）
    ├── copyutil/copyutil.go          # 深拷贝工具
    ├── cryptoutil/common.go          # 加密工具（MD5/SHA等）
    ├── gormutil/                     # Gorm工具
    │   ├── gormutil.go               # 分页查询
    │   └── model.go                  # 基础Model定义
    ├── httputil/httputil.go          # HTTP请求工具（Get/Post/Request）
    ├── idutil/idutils.go             # ID生成工具（UUID/雪花ID）
    ├── iputil/iputil.go              # IP工具（获取本机IP）
    ├── maputil/maputil.go            # Map工具
    ├── mysqlutil/mysqlutil.go        # MySQL工具
    ├── netutil/net.go                # 网络工具（端口检测）
    ├── osutil/                       # 系统工具
    │   ├── bytefmt.go                # 字节格式化
    │   ├── env.go                    # 环境变量
    │   └── file.go                   # 文件操作
    ├── reflectutil/reflect.go        # 反射工具
    ├── regexputil/regexputil.go      # 正则工具
    ├── restutil/restutil.go          # REST工具
    ├── retryutil/retryutil.go        # 重试工具
    ├── signals/                      # 信号处理
    │   ├── signal.go                 # 信号监听
    │   ├── signal_posix.go           # POSIX信号
    │   └── signal_windows.go         # Windows信号
    ├── stringutil/                   # 字符串工具
    │   ├── stringutil.go             # 字符串操作（Diff/Unique/SubString等）
    │   └── conv.go                   # 类型转换
    ├── timeutil/timeutil.go          # 时间工具
    ├── traceutil/traceutil.go        # 追踪工具
    ├── trpcutil/trpcutil.go          # trpc工具
    ├── typeutil/                     # 类型工具
    │   ├── nil.go                    # Nil判断
    │   └── typ.go                    # 类型判断
    └── userutil/userutil.go          # 用户工具
```

## 配置说明

### trpc_go.yaml 配置示例

```yaml
# etrpc 配置
etrpc:
  service_name: your-service          # 服务名称（必填）
  service_port: 8080                  # HTTP端口（默认8080）
  trpc_service_port: 8081             # TRPC端口（默认8081）
  disable_rsp_wrapper: false          # 禁用响应体包装
  ignore_rsp_wrapper_path:            # 忽略响应体包装的路径
    - /health
    - /metrics

# trpc 框架配置
global:
  namespace: Development
  env_name: dev

server:
  app: your-app
  server: your-service
  service:
    - name: your-service
      protocol: http
      port: ${etrpc.service_port}

# 客户端配置
client:
  service:
    # MySQL
    - name: trpc.mysql.xxx.xxx
      target: dsn://user:password@tcp(host:3306)/dbname
      timeout: 10000
    
    # Redis
    - name: trpc.redis.xxx.xxx
      target: redis://user:password@host:6379/0
      timeout: 5000
    
    # InfluxDB
    - name: trpc.influxdb.xxx.xxx
      target: influxdb://user:password@host:8086?timeout=5000

# 插件配置
plugins:
  log:
    default:
      - writer: console
        level: debug
```

## 使用示例

### 基本使用

```go
package main

import (
    "etrpc-go"
    "etrpc-go/log"
)

func main() {
    // 创建服务器
    s := etrpc.NewServer()
    
    // 注册服务
    // pb.RegisterYourService(s, &YourServiceImpl{})
    
    // 启动服务
    etrpc.RunServer(s)
}
```

### 配置注册

```go
package myconfig

import "etrpc-go/config"

type MyConfig struct {
    Timeout  int    `yaml:"timeout"`
    Endpoint string `yaml:"endpoint"`
}

var Cfg = &MyConfig{}

func init() {
    config.RegisterConfigWithPrefix("my.config", "myservice", Cfg, true)
}
```

## 常见问题

### 1. 服务启动失败：no trpc service config found

**问题表现**：启动时 panic，提示 "start etrpc server failed, no trpc service config found"

**可能原因**：
- `trpc_go.yaml` 中没有配置 `server.service` 列表
- 配置文件路径错误

**解决方案**：
```yaml
server:
  service:
    - name: your-service
      protocol: http
      port: 8080
```

### 2. 服务启动失败：service_name can not be empty

**问题表现**：启动时 panic，提示 "config `etrpc.service_name` can not be empty"

**可能原因**：
- `trpc_go.yaml` 中没有配置 `etrpc.service_name`

**解决方案**：
```yaml
etrpc:
  service_name: your-service-name
```

### 3. 数据库连接失败

**问题表现**：启动时 panic，提示 "connect xxx client for xxx fail"

**可能原因**：
- 数据库地址配置错误
- 数据库服务不可用
- 网络不通

**解决方案**：
- 检查 `client.service` 中的 `target` 配置
- 确认数据库服务正常运行
- 检查网络连通性

### 4. 配置注册的对象不是指针

**问题表现**：启动时 panic，提示 "配置对象仅允许为Struct指针类型"

**可能原因**：
- 注册配置时传递的不是结构体指针

**解决方案**：
```go
// 错误
config.RegisterConfig("name", MyConfig{}, false)

// 正确
config.RegisterConfig("name", &MyConfig{}, false)
```

### 5. 重复注册配置对象

**问题表现**：启动时 panic，提示 "重复注册"

**可能原因**：
- 同一个 `cfgName` 被注册了多次

**解决方案**：
- 确保每个配置名称唯一
- 检查是否有多个 init() 函数注册了相同名称

### 6. 环境变量替换无效

**问题表现**：配置中的 `${VAR}` 没有被替换

**可能原因**：
- 环境变量未设置
- 变量格式错误（只支持 `${VAR}`，不支持 `$VAR`）

**解决方案**：
- 确认环境变量已设置：`export VAR=value`
- 使用正确的格式：`${VAR}`

### 7. 配置热更新不生效

**问题表现**：修改配置后服务没有更新

**可能原因**：
- 注册配置时 `hotUpdate` 参数为 `false`
- 没有触发配置刷新

**解决方案**：
```go
// 启用热更新
config.RegisterConfig("name", &cfg, true)
```

### 8. HTTP响应没有被包装

**问题表现**：返回的 JSON 没有 code/message/data 结构

**可能原因**：
- 响应 Content-Type 不是 `application/json`
- 请求路径在 `ignore_rsp_wrapper_path` 中
- 请求使用了 PB 调用方式

**解决方案**：
- 确保响应 Content-Type 为 `application/json`
- 检查配置中的排除路径

### 9. 指标上报维度缺失

**问题表现**：上报的指标缺少 server/env/ip 等维度

**可能原因**：
- 使用了 `WithoutDefaultDims()` 选项
- trpc 全局配置未初始化（在服务启动前上报）

**解决方案**：
- 移除 `WithoutDefaultDims()` 选项
- 确保在 `etrpc.NewServer()` 之后创建指标对象

### 10. 日志没有公共字段

**问题表现**：日志中缺少 Namespace/Env/Container/IP 字段

**可能原因**：
- 使用了 trpc-go 原生的 log 包而不是 etrpc-go/log
- 在服务启动前打印日志

**解决方案**：
```go
// 使用 etrpc-go 的 log 包
import "etrpc-go/log"

log.Info("message")
```