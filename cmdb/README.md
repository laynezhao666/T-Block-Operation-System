# CMDB（配置管理数据库）

## 一、模块介绍

CMDB（Configuration Management Database）是TBOS动环系统的配置管理服务，负责管理和提供动环设备的配置数据。该服务主要为Agent采集器、Collector配置收集器等下游服务提供设备配置、采集模版、测点配置等数据的查询和导出功能。

### 1.1 主要职责

| 职责 | 说明 |
|------|------|
| 采集器配置查询 | 提供采集设备（TBox/VendorBox）及其子设备的配置查询 |
| 标准测点查询 | 提供设备标准测点的配置查询，支持标准测点到采集测点的转换 |
| 采集模版管理 | 提供采集模版及模版测点的查询（包含协议配置） |
| 模组信息管理 | 提供模组（Mozu）信息的增删改查 |
| 配置版本监控 | 实时监控配置版本变化，支持增量更新 |
| 配置导出 | 批量导出采集器配置为ZIP压缩包 |

### 1.2 数据流向

```
┌─────────────────────────────────────────────────────────────────────┐
│                          边缘侧（机房）                               │
│  ┌─────────────────┐                                                │
│  │   Agent         │  GetCollectorDevice / GetCollectorPoint        │
│  │   (边缘采集器)   │  GetCollectorTemplate                          │
│  └────────┬────────┘                                                │
│           │                                                         │
│           │ 本地HTTP调用                                             │
│           ▼                                                         │
│  ┌─────────────────┐                                                │
│  │   Collector     │  (配置收集器，部署在边缘侧，代理CMDB请求)         │
│  └────────┬────────┘                                                │
└───────────┼─────────────────────────────────────────────────────────┘
            │
            │ 跨网络调用（边缘侧 → 中心侧）
            │ GetCollectorConfig / ExportCollectorConfig
            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          中心侧（云端）                               │
│  ┌─────────────────────────────────────────────┐                    │
│  │                   CMDB                       │                    │
│  │            (配置管理服务)                     │                    │
│  └────────────────────┬────────────────────────┘                    │
│                       │                                              │
│                       ▼                                              │
│  ┌─────────────────────────────────────────────┐                    │
│  │                  MySQL                       │                    │
│  │  t_collector_device    t_device_entity      │                    │
│  │  t_collector_template  t_device_point       │                    │
│  │  t_collector_template_point  t_mozu_info    │                    │
│  │  t_alarm_strategy                           │                    │
│  └─────────────────────────────────────────────┘                    │
└─────────────────────────────────────────────────────────────────────┘
```

> **说明**：Agent部署在边缘采集设备上，通常与中心侧的CMDB服务网络不互通，因此Agent通过部署在同一边缘侧的Collector服务来间接获取CMDB的配置数据。Collector会缓存配置并提供本地HTTP接口供Agent调用。

---

## 二、核心能力

### 2.1 配置版本监控

- **监控周期**：每5秒检测一次模组发布版本变化
- **实现机制**：
  - 使用 `mozuVerCache` 缓存模组版本信息
  - 使用 `collectorVerChange` 标记版本发生变化的采集器
  - 使用 `collectorVerCache` 缓存采集器配置版本
- **触发条件**：检测到模组的 `publish_version` 发生变化

### 2.2 采集设备查询

获取采集器（TBox/VendorBox）及其子设备的完整配置信息。

**采集器类型**：

| CollectorType | 说明 |
|--------------|------|
| 1 | TBox（腾讯自研采集器） |
| 2 | TBox下接的传感器 |
| 3 | VendorBox（第三方采集盒） |
| 4 | VendorBox子设备 |

### 2.3 标准测点查询

获取采集器下所有标准测点配置，支持标准测点到采集测点的转换：
- 查询归属于指定采集器的设备测点
- 查询常量测点（部分标准到标准的点涉及常量点）
- 执行标准测点到采集测点的转换

### 2.4 采集模版查询

获取采集模版及其测点定义，包含以下信息：

| 字段 | 说明 |
|------|------|
| Cls | 设备类型 |
| Drvlib | 协议类型（modbus/snmp等） |
| Protver | 协议版本 |
| Vendor | 厂商 |
| Extend | 协议扩展配置 |

**模版测点字段**：PointNameEn、PointNameZh、PointType、PointRw、SubDevice、DeltaDef、VerifyDef、ExpDef、ProtDef、ValDef、Simulator

### 2.5 配置导出

批量导出采集器的完整配置为ZIP压缩包，导出文件结构：

```
data_2024_01_01_12_00_00.zip
├── {device_number_1}/
│   ├── devices.json      # 采集设备配置
│   ├── std.json          # 标准测点配置
│   ├── std_device.json   # 标准设备配置
│   └── templates/
│       └── {template_name}.json  # 采集模版配置
├── {device_number_2}/
│   └── ...
```

### 2.6 数据差异对比与批量更新

使用泛型 `FindDiff` 算法进行数据增量更新：
- **对比逻辑**：查找新增（新有旧无）、删除（旧有新无）、不变的元素
- **深度比较**：支持字段级变化检测
- **批量操作**：删除分批（每批5000条）、插入分批（每批1000条）

---

## 三、代码结构

```
cmdb/
├── main.go                           # 服务入口
├── trpc_go.yaml                      # 服务配置文件
├── entity/                           # 实体定义
│   └── cond/                         # 查询条件定义
│       ├── collector_device_cond.go  # 采集设备查询条件
│       ├── collector_template_cond.go # 采集模版查询条件
│       ├── device_entity_cond.go     # 标准设备查询条件
│       ├── device_point_cond.go      # 标准测点查询条件
│       └── mozu_info_cond.go         # 模组信息查询条件
├── logic/                            # 业务逻辑层
│   └── query/
│       └── api.go                    # 配置查询核心逻辑
├── service/                          # 服务接口层
│   ├── config_build_service.go       # 配置构建服务（模组管理、模型导入）
│   └── config_query_service.go       # 配置查询服务（设备/测点/模版查询）
├── repo/                             # 数据访问层
│   └── db/
│       ├── db_util.go                # 数据库工具类（泛型事务更新）
│       ├── collector_device_dao.go   # 采集设备DAO
│       ├── collector_template_dao.go # 采集模版DAO
│       ├── collector_template_point_dao.go # 采集模版测点DAO
│       ├── device_entity_dao.go      # 标准设备DAO
│       ├── device_point_dao.go       # 标准测点DAO
│       ├── mozu_info_dao.go          # 模组信息DAO
│       └── alarm_strategy_dao.go     # 告警策略DAO
└── util/                             # 工具类
    ├── collutil/
    │   └── collection_util.go        # 集合差异对比工具
    └── convutil/
        └── conv_util.go              # JSON转换工具
```

### 核心文件说明

| 文件 | 说明 |
|------|------|
| `main.go` | 服务入口，注册ConfigBuildService和ConfigQueryService两个服务 |
| `logic/query/api.go` | 核心查询逻辑，包含版本监控、设备查询、测点查询、模版查询、配置导出等功能 |
| `service/config_query_service.go` | 配置查询服务接口实现，负责参数校验和调用logic层 |
| `service/config_build_service.go` | 配置构建服务接口实现，提供模组管理和模型导入功能 |
| `repo/db/db_util.go` | 数据库工具类，提供泛型事务更新函数 |
| `util/collutil/collection_util.go` | 集合工具类，提供数据差异对比算法 |

---

## 四、API接口

### ConfigQueryService

| 接口 | 说明 |
|------|------|
| `GetCollectorDevice` | 获取采集设备及子设备配置 |
| `GetCollectorPoint` | 获取采集器下的标准测点 |
| `GetCollectorTemplate` | 获取采集模版配置 |
| `ExportCollectorConfig` | 导出采集器配置为ZIP文件 |
| `GetDeviceEntity` | 获取标准设备列表 |
| `GetDevicePoint` | 获取标准测点列表 |
| `ListCollectorDevice` | 获取采集设备列表（支持分页筛选） |
| `GetMozuInfo` | 获取模组信息 |
| `GetConfigModifyTime` | 获取配置版本信息 |

### ConfigBuildService

| 接口 | 说明 |
|------|------|
| `SaveMozu` | 保存/更新模组信息 |
| `ListMozu` | 查询模组列表 |
| `DeleteMozu` | 删除模组 |
| `ImportModel` | 导入模型配置 |

---

## 五、常见问题

### 5.1 采集器配置版本不更新

**问题描述**：修改了配置后，采集器获取到的配置版本没有变化。

**解决方案**：
1. 确认模组的 `publish_version` 字段已更新
2. 等待5秒后重新查询版本（版本监控周期为5秒）
3. 检查日志中是否有版本变化记录

### 5.2 导出配置失败

**问题描述**：调用 `ExportCollectorConfig` 接口返回错误。

**解决方案**：
1. 确认请求参数中的 `device_number`、`mozu_id`、`collector_type` 是否正确
2. 检查是否存在匹配的采集设备
3. 确认采集设备关联的模版是否存在

### 5.3 测点转换失败

**问题描述**：日志中出现 `convert to collector point fail` 错误。

**解决方案**：
1. 检查测点的 `point_key` 是否正确
2. 确认依赖的常量测点是否存在
3. 检查测点的表达式定义是否正确

### 5.4 查询数据量过大导致超时

**问题描述**：查询接口响应缓慢或超时。

**解决方案**：
1. 使用分页参数限制返回数量（默认最大10000条）
2. 添加更多筛选条件缩小查询范围
3. 检查数据库索引是否完整

### 5.5 数据库连接失败

**问题描述**：服务启动时报数据库连接错误。

**解决方案**：
1. 检查环境变量配置：`MYSQL_USER`、`MYSQL_PASSWORD`、`MYSQL_ADDR`、`MYSQL_DATABASE`
2. 确认数据库服务正常运行
3. 检查网络连通性

### 5.6 配置缓存数据不一致

**问题描述**：查询到的配置版本与实际不符。

**解决方案**：
1. 重启服务清除缓存
2. 检查 `watchCollectorVer` 协程是否正常运行
3. 确认模组版本是否正确标记变化
