# 配置管理服务 CMDB

CMDB是TBOS的配置管理数据库，负责：

- **模组信息管理**：增删改查、配置模型导入导出
- **标准设备/测点管理**：设备树、测点列表、设备实体查询
- **采集设备管理**：采集器状态树、采集设备详情、采集测点列表
- **配置查询**：设备/测点/模版批量查询、配置更新时间、配置导出

## 1. 数据流向

![CMDB数据流向](./cmdb_dataflow.svg)

## 2. 模组信息管理

CMDB服务提供了[模组](../../concepts/mozu.md)的管理和查询接口。

| 方法 | 实现逻辑 |
|------|----------|
| `SaveMozu` | 校验 `mozu_id > 0` → DAO 层通过 `FirstOrCreate` + `Assign` 实现 upsert（存在则更新，不存在则插入） |
| `ListMozu` | 支持按 `mozu_id`、`mozu_name`（模糊匹配）、`mozu_code`（模糊匹配）多条件组合查询，支持分页 |
| `DeleteMozu` | 按 `mozu_id` 列表批量删除 |
| `ImportModel` | 校验 `mozu_id` 和 `version` → 批量导入设备/测点/采集器/模版/策略共 5 个 Excel 配置模型 |

**模组查询**

| 方法 | 实现逻辑 |
|------|----------|
| `GetMozuInfo` | 按 `mozu_id` 列表查询模组详情，返回名称/编码/类型/楼栋/园区/版本等信息 |


## 3. 标准设备/测点管理

CMDB提供了**设备配置管理**和**测点配置管理**两大核心功能。

[标准设备](../../concepts/device.md)和[测点](../../concepts/point.md)通常在elvDB平台中预先配置，然后以配置文件(Excel)的形式导入到CMDB中，实现标准化管理。

### 3.1 设备配置管理

设备配置管理负责维护标准设备的属性信息，如设备编号、类型、所属模组、机房区域等。

| 方法 | 说明 |
|------|------|
| `GetDeviceEntity` | 获取设备实体列表（多条件筛选、分页） |
| `GetDeviceTree` | 获取设备树（按位置 / 按设备类型） |
| `GetSubTree` | 获取设备子树 |
| `GetSubTreeFieldDic` | 获取子树字段字典 |

### 3.2 测点配置管理

测点配置管理定义了设备的测点属性，包括测点名称、类型、读写属性、级别、表达式等，分为标准测点和常量测点。

| 方法 | 说明 |
|------|------|
| `GetDevicePoint` | 获取测点信息（多条件筛选、分页） |

## 4. 采集设备管理

CMDB提供了统一的配置管理服务来管理[采集设备](../../concepts/device.md)。

CMDB通过三层结构管理采集设备：**采集设备表**定义设备拓扑，**采集模板**定义协议与通道，**模板测点**定义每个测点的采集规则。数据同样通过 `ImportModel` 的 Excel 导入 + `BatchUpdate` diff 机制批量写入，查询侧提供多维度检索和配置导出。

### 4.1 采集设备

采集设备表为MySQL中的表 `t_collector_device` 每行代表一个采集设备（采集器或其子设备），通过 `CollectorType` 区分角色：

| 类型常量 | 值 | 角色 |
|---------|----|-----|
| `CollectorTypeTbox` | 1 | TBOX 采集器（主） |
| `CollectorTypeTboxSubDevice` | 2 | TBOX 下子设备 |
| `CollectorTypeVendorBox` | 3 | 厂商采集器（主） |
| `CollectorTypeVendorSubDevice` | 4 | 厂商下子设备 |
| `CollectorTypeDoor` | 5 | 门禁采集器（主） |
| `CollectorTypeDoorSubDevice` | 6 | 门禁下子设备 |
| `CollectorTypeTone` | 7 | TONE 采集器（主） |
| `CollectorTypeToneSubDevice` | 8 | TONE 下子设备 |

采集器与子设备通过 `ParentDeviceNumber` 形成父子拓扑，每个设备挂在特定 `MozuId` 下并关联一个采集 `TemplateName`。

### 4.2 采集模版

采集模版表为MySQL中的表 `t_collector_template`，模板定义了采集协议参数，与设备通过 `TemplateName` 关联：

| 字段 | 说明 |
|------|------|
| `TemplateName` | 模板唯一名称 |
| `ProtocolType` / `ProtocolVersion` | 采集协议类型与版本 |
| `Manufacturer` / `DeviceModelEn` | 设备制造商与型号 |
| `ProtocolExtend` | 协议扩展参数（JSON） |

### 4.3 模板测点

模版测点表为MySQL中的表 `t_collector_template_point`，每个模板下可定义多条测点，描述采集器如何采集某个测点：

| 字段 | 说明 |
|------|------|
| `TemplateName` + `SubDevice` + `PointNameEn` | 联合唯一键 |
| `DeltaDef` / `VerifyDef` / `ExpDef` / `ProtDef` / `ValDef` | 变化定义、校验规则、表达式规则、协议定义、值定义（均为 JSON） |
| `Simulator` | 模拟数据定义 |

## 5. 配置查询

CMDB提供配置查询服务（即`cmdb.ConfigQuery`服务），是下游服务（Agent、Collector、Scheduler 等）获取配置数据的核心入口。

### 5.1 设备与测点查询

| 方法 | 说明 |
|------|------|
| `GetDeviceEntity` | 获取设备实体列表（多条件筛选、分页） |
| `GetDevicePoint` | 获取测点信息 |
| `GetMozuInfo` | 获取模组信息（含版本） |

### 5.2 采集器配置查询

| 方法 | 说明 |
|------|------|
| `GetCollectorDevice` | 批量获取采集器下子设备 |
| `GetCollectorPoint` | 采集测点列表（含协议/校验/模拟定义） |
| `GetCollectorTemplate` | 获取采集模版配置 |
| `ListCollectorDevice` | 采集器列表查询 |

### 5.3 版本感知与配置导出

| 方法 | 说明 |
|------|------|
| `GetConfigModifyTime` | 获取配置更新时间 |
| `ExportCollectorConfig` | 按模组导出所有配置 |

## 6. API接口

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

## 7. 常见问题

### 7.1 采集器配置版本不更新

**问题描述**：修改了配置后，采集器获取到的配置版本没有变化。

**解决方案**：
1. 确认模组的 `publish_version` 字段已更新
2. 等待5秒后重新查询版本（版本监控周期为5秒）
3. 检查日志中是否有版本变化记录

### 7.2 导出配置失败

**问题描述**：调用 `ExportCollectorConfig` 接口返回错误。

**解决方案**：
1. 确认请求参数中的 `device_number`、`mozu_id`、`collector_type` 是否正确
2. 检查是否存在匹配的采集设备
3. 确认采集设备关联的模版是否存在

### 7.3 测点转换失败

**问题描述**：日志中出现 `convert to collector point fail` 错误。

**解决方案**：
1. 检查测点的 `point_key` 是否正确
2. 确认依赖的常量测点是否存在
3. 检查测点的表达式定义是否正确

### 7.4 查询数据量过大导致超时

**问题描述**：查询接口响应缓慢或超时。

**解决方案**：
1. 使用分页参数限制返回数量（默认最大10000条）
2. 添加更多筛选条件缩小查询范围
3. 检查数据库索引是否完整

### 7.5 数据库连接失败

**问题描述**：服务启动时报数据库连接错误。

**解决方案**：
1. 检查环境变量配置：`MYSQL_USER`、`MYSQL_PASSWORD`、`MYSQL_ADDR`、`MYSQL_DATABASE`
2. 确认数据库服务正常运行
3. 检查网络连通性

### 7.6 配置缓存数据不一致

**问题描述**：查询到的配置版本与实际不符。

**解决方案**：
1. 重启服务清除缓存
2. 检查 `watchCollectorVer` 协程是否正常运行
3. 确认模组版本是否正确标记变化
