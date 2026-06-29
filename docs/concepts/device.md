# 设备

**设备(Device)** 是一个统一抽象的概念，将传统的多层设备概念抽象为单一的实体，包括四类：

- **空间位置设备**（园区、楼栋、区域、模组）
- **标准设备**（空调机组、UPS、传感器）
- **采集设备**（采集器、网管、通信模块）
- **虚拟设备**（聚合计算点、虚拟测点）

## 设备树

设备树是设备之间的归属关系形成的树状结构，树中的每一个节点都遵循腾讯统一的设备编号规范：

1. 园区(Park)
2. 楼栋(Build)
3. 区域(Area)
4. 方仓/模组(Mozu)
5. 应用类型(Application)

举例来说，一台组合式空调的编号可能为XXXX-B01-A01-M01-AHU01。

设备树是通过 **单张扁平表 `t_device_entity` + `ParentDeviceNumber` 自引用** 实现的，并非真正的嵌套结构：

- 每行代表一个设备节点，有全局唯一 `DeviceGid` 和域内唯一 `DeviceNumber`
- `ParentDeviceNumber` 指向父节点的 `DeviceNumber`，以此形成树状层级
- `DeviceNumberRoute` 存储从根到当前节点的完整路径编码，方便快速定位

例如：`园区A → 楼栋B → 模组C → 空调机组X` 就是四条记录通过 `ParentDeviceNumber` 串联起来的链路。

## Schema

设备数据存储在 MySQL 中的`t_device_entity`表中。

`t_device_entity` 自身只存身份信息（名称、类型、所属区域等），**配置属性存放在独立的卫星表中，通过 `DeviceGid` 外键关联**，而非挂在树节点上：

| 卫星表 | 关联字段 | 存储内容 |
|--------|----------|----------|
| `t_device_point` | `DeviceGid` | 测点名称、类型、值域、单位、表达式等 |
| `t_alarm_strategy` | `DeviceGid` | 告警规则表达式、告警级别、恢复条件、通知模版 |
| `t_collector_device` | `DeviceNumber` + `ParentDeviceNumber` | 采集器本身的设备属性与子设备拓扑 |

所有卫星表也带有 `MozuId`，确保与 `t_device_entity` 处于同一模组隔离域。这意味着：
- 同一设备节点可以有零到多个测点、零到多个告警策略
- 测点通过表达式可引用其他设备的测点（`ExpressionMap` 存储 `DeviceGid → PointKey` 映射），形成跨设备的计算依赖

## 相关概念

- [模组 (Mozu)](./mozu.md) — 设备树中的空间位置节点

## 相关模块

- [CMDB](../modules/config/cmdb.md) — 设备的统一建模、配置管理与查询
