这个快速开始手册用于指导开发者通过 Docker 一键部署 TBOS 全栈环境，覆盖镜像构建、中间件启动、服务编排与 Web 访问。

> **注意**：开始之前，请先阅读 [单机部署](standalone.md) 了解 TBOS 架构与各组件说明。Docker 部署仅适用于体验和学习目的，不建议直接用于生产环境。生产环境建议使用 [分布式部署](cluster.md)。

如果希望跳过手动安装中间件的步骤，可以使用 Docker 一键部署，适合初次体验和功能验证。

```bash
# 1. 构建所有镜像
./tbos_docker.sh build

# 2. 安装 TBOS（包含中间件和业务服务）
./tbos_docker.sh install

# 3. 启动所有服务
./tbos_docker.sh start

# 访问 Web 界面：http://服务器IP:8080
```

| 依赖 | 版本要求 |
|------|---------|
| Docker Engine | 20.10+ |
| Docker Compose | 2.0+ |

相较于本地部署，Docker 部署的优势在于自动拉取并启动 MySQL、Redis、Kafka、InfluxDB 等中间件，无需手动逐个安装配置。同时自动完成数据库初始化和服务编排，一条命令即可启动全部 12 个服务。

> **注意**：Docker 部署仅适用于体验和学习目的，不建议直接用于生产环境。生产环境建议使用 [分布式部署](cluster.md)。
