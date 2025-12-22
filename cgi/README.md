# CGI API网关

CGI 是TBOS的HTTP/RESTful API网关，提供统一的对外接口，支持WebSocket实时告警推送。

## 模块功能介绍

### 核心能力

1. **告警API**：告警列表、详情、操作等接口
2. **数据API**：测点数据查询接口
3. **CMDB API**：设备配置查询接口
4. **WebSocket**：实时告警推送

## 代码结构

```
cgi/
├── main.go                        # 主入口
├── entity/                        # 实体定义
│   ├── request/                   # 请求实体
│   └── response/                  # 响应实体
├── logic/                         # 业务逻辑
│   ├── alarm/                     # 告警相关
│   │   ├── api/                   # 告警API
│   │   │   ├── list.go            # 列表查询
│   │   │   ├── detail.go          # 详情查询
│   │   │   ├── operate.go         # 告警操作
│   │   │   └── stats.go           # 统计查询
│   │   ├── consumer/              # 消息消费
│   │   │   └── alarm_consumer.go  # 告警消费
│   │   └── wslogic/               # WebSocket逻辑
│   │       ├── hub.go             # 连接管理
│   │       ├── client.go          # 客户端
│   │       └── message.go         # 消息处理
│   ├── cmdb/                      # CMDB相关
│   │   ├── device.go              # 设备查询
│   │   └── template.go            # 模版查询
│   ├── data/                      # 数据查询
│   │   ├── realtime.go            # 实时数据
│   │   └── history.go             # 历史数据
│   ├── common/                    # 通用接口
│   │   └── health.go              # 健康检查
│   ├── cache/                     # 缓存
│   │   └── cache.go
│   └── util/                      # 工具类
│       └── response.go            # 响应工具
├── service/                       # 服务实现
│   ├── alarm_handler.go           # 告警处理
│   ├── data_handler.go            # 数据处理
│   └── cmdb_handler.go            # CMDB处理
└── repo/                          # 数据访问
```

## 核心功能介绍

### 1. API路由

```go
// 告警API
router.GET("/api/v1/alarms", listAlarms)
router.GET("/api/v1/alarms/:id", getAlarmDetail)
router.POST("/api/v1/alarms/:id/ack", ackAlarm)
router.GET("/api/v1/alarms/stats", getAlarmStats)

// 数据API
router.GET("/api/v1/data/realtime", getRealtimeData)
router.POST("/api/v1/data/history", getHistoryData)

// CMDB API
router.GET("/api/v1/devices", listDevices)
router.GET("/api/v1/devices/:id", getDeviceDetail)

// WebSocket
router.GET("/ws/alarm", wsAlarmHandler)
```

### 2. WebSocket推送

```go
// 连接管理
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

// 推送告警
func (h *Hub) BroadcastAlarm(alarm Alarm) {
    msg, _ := json.Marshal(alarm)
    h.broadcast <- msg
}

// 客户端处理
func (c *Client) ReadPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()
    
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        // 处理客户端消息(如订阅、心跳等)
        c.handleMessage(message)
    }
}
```

### 3. 告警消费与推送

```go
func (c *AlarmConsumer) Start() {
    for msg := range c.kafka.Messages() {
        var alarm Alarm
        json.Unmarshal(msg.Value, &alarm)
        
        // 推送到WebSocket
        wsHub.BroadcastAlarm(alarm)
    }
}
```

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
