# 统一网关 CGI

CGI 是TBOS的HTTP/RESTful API网关，提供统一的对外接口，支持WebSocket实时告警推送。其主要能力包括：

- 告警API：告警列表、详情、操作等接口
- 数据API：测点数据查询接口
- CMDB API：设备配置查询接口
- WebSocket：实时告警推送

## 路由总览

CGI 作为网关将前端请求路由到下游各业务服务：

### 告警路由

| 路由前缀 | 下游服务 | 说明 |
|----------|----------|------|
| `/alarm/server/` | alarm-server | 告警查询、统计、操作 |
| `/alarm/manage/` | alarm-manage | 告警推送管理 |

### 数据路由

| 路由前缀 | 下游服务 | 说明 |
|----------|----------|------|
| `/Data/Query` | data-cache / data-query | 综合数据查询 |
| `/Data/TracePoint` | data-cache | 测点溯源 |
| `/Data/QueryLatest` | data-cache | 批量最新值查询 |

### CMDB 路由

| 路由前缀 | 下游服务 | 说明 |
|----------|----------|------|
| `/Cmdb/` | cmdb | 设备/测点/采集器查询 |

### 通用接口

| 方法 | 路径 | 说明 |
|------|------|------|
| ExportData | `/Common/ExportData` | 通用数据导出 |
| GetKeyDict | `/Common/GetKeyDict` | 获取配置字典（单字段） |
| GetKvDict | `/Common/GetKvDict` | 获取配置字典（KV 字段） |

## 接口约定

- **协议**: tRPC-Go / HTTP REST
- **序列化**: Protocol Buffers 3（内部）、JSON（对外）
- **通用响应格式**: `{code: 0, message: "success", data: {...}}`
- **认证**: 基于 Token 的登录认证

## 配置文件示例

```yaml
server:
  app: tbos
  server: cgi
  service:
    - name: trpc.tbos.cgi.CGIService
      ip: 0.0.0.0
      port: 8080
      protocol: http

custom:
  kafka:
    brokers: ["localhost:9092"]
    alarm_topic: "tbos_alarm"
    group_id: "cgi-alarm-group"
    
  websocket:
    # 心跳间隔(秒)
    heartbeat_interval: 30
    # 写超时(秒)
    write_timeout: 10
    # 最大消息大小(字节)
    max_message_size: 65536
    
  services:
    alarm_server: "alarm-server:8086"
    data_cache: "data-cache:8084"
    data_query: "data-query:8090"
    cmdb: "cmdb:8087"
```

## API接口文档

### 告警接口

#### 获取告警列表
```
GET /api/v1/alarms?page=1&size=20&level=1,2&status=active
```

响应：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 100,
        "items": [
            {
                "alarm_id": "123456789",
                "device_id": "device_001",
                "device_name": "UPS-001",
                "point_id": "voltage",
                "alarm_level": 2,
                "alarm_content": "电压过低",
                "alarm_time": 1699999999,
                "status": "active"
            }
        ]
    }
}
```

#### 确认告警
```
POST /api/v1/alarms/{alarm_id}/ack
```

请求：
```json
{
    "operator": "admin",
    "comment": "已处理"
}
```

### 数据接口

#### 获取实时数据
```
GET /api/v1/data/realtime?point_ids=temp_001,humidity_001
```

#### 获取历史数据
```
POST /api/v1/data/history
```

请求：
```json
{
    "point_ids": ["temp_001"],
    "start_time": 1699900000,
    "end_time": 1699999999,
    "interval": 60
}
```

### WebSocket

#### 连接
```
ws://host:8080/ws/alarm
```

#### 消息格式
```json
{
    "type": "alarm",
    "data": {
        "alarm_id": "123456789",
        "device_name": "UPS-001",
        "alarm_content": "电压过低",
        "alarm_level": 2,
        "alarm_time": 1699999999
    }
}
```

## 常见问题

### 1. WebSocket连接断开

**问题**：客户端频繁断开连接

**解决方案**：
- 检查网络连接稳定性
- 调整心跳间隔配置
- 前端实现自动重连机制
- 检查代理/负载均衡超时配置

### 2. 接口响应超时

**问题**：API响应时间过长

**解决方案**：
- 检查后端服务是否正常
- 优化数据库查询
- 启用接口缓存
- 增加服务超时时间

### 3. 告警推送延迟

**问题**：WebSocket告警推送不及时

**解决方案**：
- 检查Kafka消费是否正常
- 优化消息处理逻辑
- 增加消费者并发

### 4. 跨域问题

**问题**：前端请求被CORS拦截

**解决方案**：
- 配置CORS中间件
- 允许指定的域名访问
- WebSocket同样需要配置跨域

## 性能优化

1. **接口缓存**
   - 对低频变化的数据启用缓存
   - 设置合理的缓存过期时间

2. **连接池**
   - 复用后端服务连接
   - 限制最大连接数

3. **限流**
   - 对高频接口进行限流
   - 防止恶意请求

## 安全建议

1. **身份认证**
   - 接入统一认证系统
   - API Token验证

2. **参数校验**
   - 严格校验请求参数
   - 防止SQL注入和XSS

3. **日志审计**
   - 记录关键操作日志
   - 异常行为告警
