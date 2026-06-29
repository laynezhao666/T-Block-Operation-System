# 测点

**测点(Point)** 指的是设备上的标准化数据采集或计算点位，用于设备数据的采集、监控、告警和控制。包括：

- **采集测点**：从设备直接采集的原始测点，数据的源头。
- **标准测点**：对采集测点进行标准化处理/映射转换后得到的测点。
- **虚拟测点**：不是直接采集，而是通过对一个或多个测点数据进行统计、数学运算、表达式求值而计算生成的衍生测点。
- **告警测点**：与告警相关的虚拟测点。

每个测点包含唯一标识（ID）、名称、值类型（模拟量/布尔量/枚举量）、读写权限（只读/读写/只写）等属性。

测点类型在TBOS中的`common`模块定义如下：

| 常量名 | 类型值 | 测点类型 |
| --- | --- | --- |
| PointTypeCollect | 1 | 采集测点 |
| PointTypeStd | 2 | 标准测点 |
| PointTypeVirtual | 3 | 虚拟测点 |
| PointTypeAlarm | 4 | 告警测点 |

## 1. 采集测点 {#collect-point}

采集测点是直接从设备采集得到的测点，是原始数据的来源。

采集测点由**Agent边缘采集器**通过Modbus、SNMP、DIO等工业协议直接从动环设备读取，得到的是设备返回的原始数据值。

采集测点是后续**标准测点**、**虚拟测点**计算的输入数据来源。

## 2. 标准测点

标准测点是对采集测点进行标准化处理/映射转换之后得到的测点，遵循统一的命名规范和数据类型。

标准测点由**Agent**的标准测点逻辑模块`agent/logic/std`通过映射配置和表达式计算生成。

标准测点的存在屏蔽了不同厂商设备的差异，向上层（告警、存储、展示）提供了统一格式的业务数据。

## 3. 虚拟测点

虚拟测点是通过对一个或多个测点数据进行统计、数学运算、表达式求值而计算生成的衍生测点，主要分为两大类：

- **设备性能监控点**：应用于设备健康度监控、通信质量评估、采集性能分析等。
- **虚拟计算测点**：应用于业务指标聚合、跨测点数据融合、复杂衍生计算等。

虚拟测点来源于Agent或Alarm-Compute。

- 在Agent侧，虚拟测点通过**cron定时器(默认每6秒)**周期性计算设备性能指标 (比如，`CommID`通信状态、`PointThroughputID`吞吐量等)。
- 在Alarm-Compute侧，虚拟测点通过 TQNL 表达式计算如 `A*10+B` 这类含变量映射的表达式，将结果推送到Kafka。

## 4. 告警测点

告警测点是与告警相关的虚拟测点，属于一种特殊用途的虚拟测点，承载告警计算/状态相关的数据。

告警测点与告警计算引擎配合，在计算的值触发告警条件时，告警的消息将会被写入到Kafka。

## 5. Schema

测点在 TBOS 中由两张 MySQL 表承载，分别对应不同层次的数据模型。

### 5.1 设备测点表

设备测点信息存储在 MySQL 中的表`t_device_point`中，存储挂载在设备上的所有测点（含标准测点、虚拟测点等）。

表中的每行可以通过 `device_gid + point_name_en` 唯一定位：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `bigint(20)` | 主键 ID（自增） |
| `device_gid` | `varchar(255)` | 所属设备 GID |
| `device_number` | `varchar(255)` | 所属设备编号 |
| `belong_collector` | `varchar(255)` | 数据来源的采集器编号 |
| `point_name_en` | `varchar(255)` | 测点英文名 |
| `point_name_zh` | `varchar(255)` | 测点中文名 |
| `point_key` | `varchar(255)` | 全局唯一标识（`DeviceGid.PointNameEn`） |
| `point_category` | `tinyint(4)` | 测点分类：0 未定义 / 1 全采集 / 2 标准+采集 / 3 全标准 |
| `point_rw` | `varchar(16)` | 读写权限（只读/读写/只写） |
| `point_level` | `varchar(16)` | 测点级别 |
| `expression` | `text` | 测点表达式（用于虚拟/标准测点的计算） |
| `expression_map` | `text` | 表达式变量 → `DeviceGid.PointNameEn` 映射（`;` 分隔） |
| `expression_map_zh` | `text` | 表达式变量 → `DeviceNumber.PointNameZh` 映射 |
| `value_type` | `varchar(16)` | 值类型（模拟量/布尔量/枚举量） |
| `value_valid_range` | `varchar(255)` | 值有效范围 |
| `value_unit` | `varchar(32)` | 值单位 |
| `value_precision` | `varchar(16)` | 值精度（小数位数） |
| `value_enum` | `varchar(255)` | 枚举值映射 |
| `mozu_id` | `int(11)` | 所属模组 ID（数据隔离） |
| `create_at` | `datetime` | 创建时间 |
| `update_at` | `datetime` | 更新时间 |

唯一键：`(device_gid, point_name_en)`，索引覆盖 `device_number`、`point_key`、`belong_collector`。

### 5.2 采集模板测点表

从**采集测点向标准测点转换**的测点采集规则在采集模版测点表中给出，即MySQL中的表 `t_collector_template_point`。

它是采集测点的实际配置来源：

| 字段 | 类型 | 说明 |
|------|------|------|
| `template_name` | `varchar(127)` | 所属模板名称 |
| `sub_device` | `varchar(127)` | 子设备名称 |
| `point_name_en` / `point_name_zh` | `varchar(255)` | 测点英文/中文名 |
| `point_type` / `point_rw` | `varchar(16)` | 测点类型 / 读写分类 |
| `point_standard` | `tinyint(4)` | 是否标准测点 |
| `delta_def` | `varchar(1024)` | 变化定义规则（JSON） |
| `verify_def` | `varchar(1024)` | 校验规则（JSON） |
| `exp_def` | `text` | 表达式定义规则（JSON） |
| `prot_def` | `text` | 协议定义规则（JSON，含寄存器地址等） |
| `val_def` | `varchar(1024)` | 值定义规则（JSON） |
| `simulator` | `varchar(255)` | 模拟数据定义 |

唯一键：`(template_name, sub_device, point_name_en)`。

### 5.3 关键设计

- **全局唯一标识**：`point_key = DeviceGid.PointNameEn`，是整个系统中引用测点的统一标识符。
- **表达式链**：`expression` 字段存储计算表达式（如 `a*10+b`），`expression_map` 将表达式中的变量映射到具体测点的 `point_key`，形成跨设备、跨模版的计算依赖链路。
- **数据隔离**：两张表均带 `mozu_id`，与设备、模版处于同一隔离域。
