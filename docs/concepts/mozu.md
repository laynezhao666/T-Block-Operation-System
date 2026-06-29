# 模组

**模组(Mozu)** 是设备树层级结构中的一个节点，代表数据中心内的一个方舱或模块化机房。

在设备编号体系中，模组位于区域之下、设备之上。例如，编号 `"XXXX-B01-A01-M01"` 中的 `"M01"` 即表示 1 号模组。

模组作为空间位置设备，用于组织和管理其下的物理设备。

## 角色

模组是TBOS系统中的核心组织单元，主要起到如下作用：

- **设备分层标识**：模组用于精确定位机房内的设备。
- **配置管理基础**：模组是设备配置、采集模版和测点数据组织和查询的最小单元。
- **资源部署单元**：模组是资源规划和扩容的基本单位。

## Schema

模组信息存储在 MySQL 的 `t_mozu_info` 表中：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `int(11)` | 主键ID（自增） |
| `mozu_id` | `int(11)` | 模组ID（唯一标识，UNIQUE 约束） |
| `mozu_name` | `varchar(32)` | 模组名称 |
| `mozu_code` | `varchar(32)` | 模组编码 |
| `mozu_type` | `int(11)` | 模组类型 |
| `belong_building` | `varchar(32)` | 所属楼栋 |
| `belong_campus` | `varchar(32)` | 所属园区 |
| `belong_campus_code` | `varchar(32)` | 所属园区编码 |
| `publish_version` | `varchar(32)` | 配置下发版本号 |
| `alarm_version` | `varchar(32)` | 告警下发版本号 |
| `create_at` | `datetime` | 记录创建时间 |
| `update_at` | `datetime` | 记录更新时间 |

## 相关模块

- [CMDB](../modules/config/cmdb.md) — 模组的增删改查、配置导入导出
