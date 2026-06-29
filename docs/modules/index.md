# 系统架构

## 整体架构

TBOS将数据中心的动环系统划分为四个层级：

- **硬件层**：分布在机房各角落的物理硬件，可以在上面部署服务
- **采集层**：采集agent部署在硬件上，向下统一数据，向上统一上报
- **存储层**：部署在边端集群，使用消息队列存储数据
- **应用层**：直接用于告警服务、大数据服务

![整体架构](../assets/arch.png)

## 模块总览

| 模块 | 目录 | 包含服务 | 说明 |
|------|------|----------|------|
| **采集** | [collect/](collect/index.md) | Collector + Agent | 边缘数据采集层，设备接入、数据上报、配置拉取 |
| **配置** | [config/](config/index.md) | CMDB + Scheduler | 设备建模、配置版本管理、任务调度与配置下发 |
| **数据** | [data/](data/index.md) | Data Store + Data Cache + Data Compute + Data Query | 数据存储、查询、计算 |
| **告警** | [alarm/](alarm/index.md) | Alarm Compute + Alarm Manage + Alarm Server | 告警策略计算、告警管理与多通道推送 |
| **平台** | [cgi.md](cgi.md) | CGI | API 网关、请求路由与 WebSocket 实时推送 |
| **Web 前端** | [web.md](web.md) | Web | 前端监控管理界面（可能不开源） |

## 接口约定

- **协议**: tRPC-Go / HTTP REST（`@alias` 注解映射）
- **序列化**: Protocol Buffers 3
- **通用响应格式**: CGI 层统一包装 `{code, message, data}`
- **认证**: 基于 Token 的登录认证（参见 Agent `ConfigManager.Login`）

## 基础设施（非业务模块）

以下为框架/协议层，不属于业务 API 文档范围：

| 模块 | 路径 | 说明 |
|------|------|------|
| `etrpc-go` | `ref/tbos/etrpc-go/` | 内部 tRPC-Go 框架扩展 |
| `common` | `ref/tbos/common/` | 公共库（实体 + 工具） |
| `trpcprotocol` | `ref/tbos/trpcprotocol/` | tRPC 协议定义（12 个子模块） |
