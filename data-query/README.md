# Data-Query 数据查询服务

Data-Query 是TBOS的测点数据查询服务，提供测点历史数据的区间查询、最新值查询以及测点变化查询能力。通过可插拔的读取插件机制，支持从不同数据源（缓存服务、InfluxDB）获取数据。

## 模块介绍

Data-Query 服务主要负责：
- 提供测点数据的区间查询（按时间段查询历史数据）
- 提供测点最新值查询
- 提供测点变化时间查询
- 支持数据填充和时间对齐
- 通过插件化的读取架构，支持从不同数据源获取数据

### 数据流向

```
┌──────────────────┐      ┌──────────────────┐
│    调用方         │      │                  │
│  (CGI/其他服务)   │─────▶│   Data-Query     │
└──────────────────┘      │                  │
                          │  DataService:    │
                          │  - DataQuery     │
                          │  - DataChange    │
                          │  - DataPointChange│
                          └────────┬─────────┘
                                   │
                   ┌───────────────┴───────────────┐
                   │ 读取插件 (按Order优先级选择)    │
                   ▼                               ▼
        ┌──────────────────┐            ┌──────────────────┐
        │   cache-api      │            │     influx       │
        │  (缓存服务读取)    │            │  (InfluxDB读取)   │
        │  Order: 1        │            │  Order: 2        │
        │  范围: 10分钟内   │            │  范围: 100天内    │
        └────────┬─────────┘            └────────┬─────────┘
                 │                               │
                 ▼                               ▼
        ┌──────────────────┐            ┌──────────────────┐
        │   Data-Cache     │            │    InfluxDB      │
        │   (缓存服务)      │            │   (时序数据库)    │
        └──────────────────┘            └──────────────────┘
```

## 核心能力

### 1. 可插拔读取插件架构

采用插件化设计，支持灵活配置多个数据源：

- **插件注册**：通过 `read.Register()` 注册读取插件
- **优先级排序**：按 `order` 配置值排序，数值越小优先级越高
- **自动选择**：根据查询时间范围自动选择合适的数据源

内置读取插件：

| 插件类型 | 说明 | 适用场景 |
|---------|------|---------|
| `cache-api` | 通过RPC调用Data-Cache服务获取数据 | 近期数据（默认10分钟内） |
| `influx` | 直接查询InfluxDB时序数据库 | 历史数据（默认100天内） |

插件接口定义：
```go
type IReadPlugin interface {
    Setup(cfg PlgConfig) (IReadPlugin, error)          // 初始化插件
    GetType() string                                    // 获取插件类型标识
    CanRead(begin, end int64) bool                      // 判断时间范围是否可查
    ReadRange(ctx, pointName[], begin, end) (map, error) // 区间查询
    ReadLatest(ctx, pointName[], max) (map, error)      // 最新值查询
    ReadChanged(ctx, pointName[], begin, end) (map, error) // 变化查询
}
```

### 2. 三种查询模式

#### 2.1 区间查询 (DataQuery)

根据时间范围查询测点数据，支持数据填充和时间对齐：

- **最新值查询**：`begin=0` 时只返回最新一个值
- **单时间点查询**：`begin=end` 时查询指定时间点数据
- **时间段查询**：`begin<end` 时查询时间范围内数据，按 `interval` 间隔填充

数据填充逻辑：
```
查询时间点 T，如果T没有数据：
1. 查找时间点<=T且最接近T的数据点
2. 判断该数据点是否在有效期内（ExpireTimeSinceQuery配置，默认180秒）
3. 如果有效则用该值填充，否则该时间点无数据
```

查询限制：
- 当 `interval < IntervalLimit`（默认60秒）时，查询时间跨度不能超过31分钟

#### 2.2 变化时间查询 (DataChange)

查询测点最近一次变化的时间戳：


#### 2.3 变化测点查询 (DataPointChange)

查询在指定时间范围内发生过变化的测点列表：

## 代码结构

```
data-query/
├── main.go                           # 主入口，初始化读取插件并注册服务
├── trpc_go.yaml                      # 服务配置文件
├── entity/                           # 实体定义
│   ├── config/                       # 配置定义
│   │   └── conf.go                   # ServerConfStruct 服务配置结构体
│   ├── errcode/                      # 错误码定义
│   │   └── errcode.go                # 自定义错误码
│   ├── constants.go                  # 常量定义（插件类型、默认值等）
│   └── point.go                      # Point、CachePoint 测点数据结构
├── logic/                            # 业务逻辑层
│   └── query/                        # 查询逻辑
│       ├── query.go                  # DataQueryHandler 区间查询核心逻辑
│       ├── change.go                 # DataChangeHandler 变化时间查询
│       └── pointChange.go            # DataPointChangeHandler 变化测点查询
├── repo/                             # 数据访问层
│   ├── config.go                     # 仓库层配置
│   └── read/                         # 读取插件
│       ├── plugin.go                 # 插件管理：注册、初始化、批量读取
│       ├── influxdb_read.go          # InfluxDB读取插件实现
│       └── store_api_read.go         # 缓存服务(Data-Cache)读取插件实现
├── service/                          # 服务层
│   └── data.go                       # DataServiceImpl 服务接口实现
└── utils/                            # 工具类
    └── slice.go                      # GetBatchStringList 分批工具函数
```

### 核心文件说明

#### main.go
服务主入口，负责：
1. 创建etrpc服务器
2. 初始化读取插件 `read.Init()`
3. 注册DataService服务

#### logic/query/query.go
区间查询核心逻辑：
- `DataQueryHandler()`：处理区间查询请求
- `processSinglePoint()`：单时间点数据填充
- `processRangePoint()`：时间段数据填充（双指针算法）
- `processLatestPoint()`：最新值处理
- `isValidPoint()`：检查数据点是否在有效期内

#### repo/read/plugin.go
读取插件管理：
- `IReadPlugin` 接口定义
- `Register()`：注册读取插件
- `Init()`：按优先级初始化插件
- `BatchReadRangePoints()`：区间批量读取
- `BatchReadLatestPoint()`：最新值批量读取
- `BatchReadChangedPoint()`：变化时间批量读取

#### repo/read/store_api_read.go
缓存服务读取插件（type: store-api）：
- 通过RPC调用Data-Cache服务
- 支持配置读取时间范围 `read_minutes`
- 变化查询支持分批并发

#### repo/read/influxdb_read.go
InfluxDB读取插件（type: influx）：
- 直接查询InfluxDB时序数据库
- 支持配置存储天数 `store_day`
- 查询时自动往前多取65秒数据用于填充

## 配置说明

### trpc_go.yaml 配置示例

```yaml
etrpc:
  service_name: data-query
  service_port: ${PORT_DATA_QUERY}

server:
  service:
    - name: ${etrpc.service_name}
      protocol: http
      port: ${etrpc.service_port}

client:
  service:
    # Data-Cache服务客户端
    - callee: tbos.data.cache.Point
      name: data-cache
      target: ip://${LOCAL_IP}:${PORT_DATA_CACHE}
      protocol: http
    # InfluxDB客户端（可选）
    - name: trpc.influxdb.idc.tbos
      target: influxdb://${INFLUXDB_USER}:${INFLUXDB_PASSWORD}@${INFLUXDB_ADDR}

# 读取插件配置
read:
  plugins:
    # 缓存服务读取插件（优先级高）
    - type: store-api
      name: tbos-cache
      order: 1
      extra:
        read_minutes: 60        # 可查询的时间范围（分钟）
    
    # InfluxDB读取插件（优先级低）
    - type: influx
      name: tbos-influx
      order: 2
      extra:
        influx_name: trpc.influxdb.idc.tbos
        influx_database: ${INFLUXDB_DBNAME}
        data_measurement: ${INFLUXDB_POINT_MEASUREMENT}
        store_day: 100          # 存储天数
        bath_size: 10000        # 批次大小

# 服务配置参数
ExpireTimeSinceQuery: 180        # 数据填充有效期（秒）
ExpireTimeMargin: 3              # 过期时间余量（秒）
NormalCostThreshold: 100         # 正常耗时判断阈值（毫秒）
ExtremelyCostThreshold: 200      # 长耗时判断阈值，超过会打印日志（毫秒）
QueryChangedBatchSize: 500       # 变化查询批次大小
QueryChangedConcurrencyLimit: 10 # 变化查询并发数
IntervalLimit: 60                # 间隔限制（秒）
```

### 配置项说明

| 配置项 | 说明 | 默认值 |
|-------|------|--------|
| ExpireTimeSinceQuery | 数据填充时，数据点的有效期（秒） | 180 |
| ExpireTimeMargin | 缓存查询时的过期余量（秒） | 3 |
| NormalCostThreshold | 正常耗时判断阈值（毫秒） | 100 |
| ExtremelyCostThreshold | 长耗时判断阈值，超过会打印日志（毫秒） | 200 |
| QueryChangedBatchSize | 变化查询单批测点数量 | 500 |
| QueryChangedConcurrencyLimit | 变化查询最大并发数 | 10 |
| IntervalLimit | 小间隔查询的间隔阈值（秒） | 60 |

### 读取插件配置项

#### store-api 插件
| 配置项 | 说明 | 默认值 |
|-------|------|--------|
| read_minutes | 可查询的时间范围（分钟） | 10 |

#### influx 插件
| 配置项 | 说明 | 默认值 |
|-------|------|--------|
| influx_name | InfluxDB客户端名称 | 必填 |
| influx_database | 数据库名称 | tbos |
| data_measurement | 测量名称 | points |
| store_day | 存储天数 | 100 |
| bath_size | 批量查询大小 | 10000 |

## API接口

### DataQuery - 区间查询

**请求参数**:

| 字段 | 类型 | 说明 |
|------|------|------|
| PointList | []string | 测点名称列表 |
| Begin | int64 | 开始时间戳（秒），0表示查最新值 |
| End | int64 | 结束时间戳（秒） |
| Interval | int64 | 数据间隔（秒），默认1 |

**返回参数**:

| 字段 | 类型 | 说明 |
|------|------|------|
| GetPointData | map[string]*InnerMap | 测点数据，key为测点名，value为时间戳->值的映射 |
| MissPointName | []string | 未查到数据的测点列表 |

### DataChange - 变化时间查询

**请求参数**:

| 字段 | 类型 | 说明 |
|------|------|------|
| PointList | []string | 测点名称列表 |
| Begin | int64 | 开始时间戳（秒） |

**返回参数**:
| 字段 | 类型 | 说明 |
|------|------|------|
| ChangedPointMap | map[string]int64 | 测点最近变化时间映射 |

### DataPointChange - 变化测点查询

**请求参数**:

| 字段 | 类型 | 说明 |
|------|------|------|
| PointList | []string | 测点名称列表 |
| Begin | int64 | 开始时间戳（秒） |
| End | int64 | 结束时间戳（秒） |

**返回参数**:
| 字段 | 类型 | 说明 |
|------|------|------|
| PointList | []string | 在时间范围内发生变化的测点列表 |

## 常见问题

### 1. 查询返回空数据

**问题表现**：查询结果 `GetPointData` 为空，测点都在 `MissPointName` 中

**可能原因**：
- 查询时间范围超出所有插件的可查范围
- 测点名称不存在
- 数据源（Data-Cache或InfluxDB）中没有数据

**解决方案**：
- 检查查询时间是否在配置的可查范围内
- 确认测点名称正确
- 检查上游数据源是否正常写入数据

### 2. 查询耗时过长

**问题表现**：日志中出现长耗时警告（超过 `ExtremelyCostThreshold`）

**可能原因**：
- 查询测点数量过多
- 查询时间跨度过大
- 数据填充计算量大
- 下游服务响应慢

**解决方案**：
- 减少单次查询的测点数量
- 缩小查询时间范围
- 增大 `interval` 减少填充点数
- 检查Data-Cache或InfluxDB服务状态

### 3. 数据填充不符合预期

**问题表现**：某些时间点没有数据，或填充的值不是最近的

**可能原因**：
- 数据点超出有效期（`ExpireTimeSinceQuery` 配置，默认180秒）
- 原始数据本身就缺失

**解决方案**：
- 调整 `ExpireTimeSinceQuery` 配置，增大有效期
- 检查上游数据采集是否正常

### 4. 小间隔查询被拒绝

**问题表现**：返回错误 "查询时间区间不能超过31分钟，当间隔小于60分钟时"

**可能原因**：
- `interval` 参数小于 `IntervalLimit`（默认60秒），且时间跨度超过31分钟

**解决方案**：
- 增大 `interval` 参数
- 缩小查询时间范围
- 调整 `IntervalLimit` 配置（需评估性能影响）

### 5. 读取插件初始化失败

**问题表现**：服务启动时panic，提示 "read plugin setup failed"

**可能原因**：
- 插件类型名称拼写错误
- 必填配置项缺失（如 `influx_name`）
- 下游服务连接失败

**解决方案**：
- 检查 `read.plugins` 配置的 `type` 是否正确（store-api/influx）
- 确认必填配置项已填写
- 检查网络连接和依赖服务状态

### 6. 缓存插件和InfluxDB数据不一致

**问题表现**：同一时间范围，从不同插件查到的数据不同

**可能原因**：
- Data-Cache中的数据有过期淘汰
- InfluxDB写入有延迟
- 两个数据源的数据周期不同

**解决方案**：
- 这是正常现象，缓存服务只保留近期数据
- 如需查历史数据，确保时间超出缓存范围，让请求路由到InfluxDB

### 7. 如何扩展新的读取插件

1. 实现 `IReadPlugin` 接口：
```go
type IReadPlugin interface {
    Setup(cfg PlgConfig) (IReadPlugin, error)
    GetType() string
    CanRead(begin, end int64) bool
    ReadRange(ctx context.Context, pointName []string, begin, end int64) (map[string][]*entity.Point, error)
    ReadLatest(ctx context.Context, pointName []string, max int64) (map[string]*entity.Point, error)
    ReadChanged(ctx context.Context, pointName []string, begin int64, end int64) (map[string]int64, error)
}
```

2. 在 `init()` 中注册插件：
```go
func init() {
    Register("your_plugin_type", &YourRead{})
}
```

3. 在配置文件中启用：
```yaml
read:
  plugins:
    - type: your_plugin_type
      name: your-plugin-name
      order: 3
      extra:
        # 自定义配置
```