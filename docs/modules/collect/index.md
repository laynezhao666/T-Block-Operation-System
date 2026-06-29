# 采集模块

采集模块是系统的数据入口和边缘计算核心，负责从各类动环设备（如空调、电力、电池等）采集实时数据，并进行标准化处理后上报至云端数据处理中心。

## 模块概览

采集模块包含以下 3 个核心服务：

| 服务 | 文档 | 说明 |
|------|------|------|
| **Collector** | [collector.md](collector.md) | 配置与数据总线服务，负责为 Agent 提供采集配置获取能力，并将测点数据转发到 Kafka 或上游 Collector |
| **Agent** | [agent.md](agent.md) | 边缘采集器，部署在边缘设备上，支持 Modbus/SNMP 等多协议采集、标准测点计算与实时数据库 |

## 数据流向

![collector数据流向](./dataflow.svg)

## 服务间关系

- **Agent** 通过 Collector 的 `ConfigBusService` 拉取采集配置（设备、模板、标准测点）
- **Agent** 通过 Collector 的 `DataBusService` 上报测点数据（TBOS 标准测点 / 原始采集测点）
- **Agent** 通过 Collector 的 `ControlBus` 上报心跳
