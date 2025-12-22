# Data-Compute 标准测点计算服务

Data-Compute 是 TBOS 系统中负责标准测点计算的核心服务，接收 Scheduler 调度器下发的计算任务，根据表达式和引用测点数据计算出标准测点值，并将结果上报到 Kafka。

## 模块介绍

Data-Compute 服务的主要职责：

1. **接收计算任务**：接收 Scheduler 调度器下发的标准测点计算任务
2. **定时计算执行**：支持周期计算（60秒全量）和变化计算（5秒增量）两种模式
3. **表达式求值**：根据任务中的表达式和引用测点值进行数学表达式计算
4. **心跳注册**：定期向 Scheduler 上报心跳，维护 Worker 状态

### 数据流向

```
┌─────────────────┐           ┌─────────────────┐
│   Scheduler     │  下发任务  │  Data-Compute   │
│    调度器       │ ────────► │   计算服务       │
└─────────────────┘           └────────┬────────┘
                                       │
                              查询引用测点值
                                       │
                                       ▼
                              ┌─────────────────┐
                              │   Data-Cache    │
                              │   数据缓存服务   │
                              └─────────────────┘
                                       │
                                 返回测点数据
                                       │
                                       ▼
                              ┌─────────────────┐
                              │  表达式计算引擎  │
                              │   expr.EvalFloat │
                              └────────┬────────┘
                                       │
                                 计算结果上报
                                       │
                                       ▼
                              ┌─────────────────┐
                              │     Kafka       │
                              │   (主备双活)     │
                              └─────────────────┘
```

## 核心能力

### 1. 任务接收与管理

通过 `ReceiveTask` 接口接收 Scheduler 下发的计算任务：

```go
// service/compute_service.go
func (s *computeServiceImpl) ReceiveTask(ctx context.Context, req *data_compute.ReqReceiveTask) (*emptypb.Empty, error) {
    register.TaskVerMark = req.TaskVerMark  // 更新任务版本标识
    go s.computeApi.ReceiveTask(req)         // 异步处理任务
    return &emptypb.Empty{}, nil
}
```

任务下发支持两种模式：
- **全量下发**（`PublishType=1`）：清空现有任务，加载全部新任务
- **增量下发**（`PublishType=0`）：先删除指定任务，再添加新任务

任务Key格式：`{DeviceGid}.{PointNameEn}.{Version}`

### 2. 定时计算调度

通过 `StartCalcPoint` 启动计算循环，每5秒执行一次：

```go
// logic/compute/compute.go
func (d *computeApiImpl) StartCalcPoint(ctx context.Context, wg *sync.WaitGroup) {
    nowSec := time.Now().Unix()
    ticker := time.NewTicker(time.Second * 5)
    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-ticker.C:
            // 60s整个周期计算一次（全量）
            if (msg.Unix()-nowSec)%60 == 0 {
                go d.calcFullPoints(wg)
            } else {
                // 其他时间计算变化测点（增量）
                go d.calcChangedPoints(msg.Add(-6*time.Second), wg)
            }
        }
    }
}
```

**计算模式**：
- **周期计算（60秒）**：计算所有标准测点，结果以`PointIntervalPeriod`类型上报
- **变化计算（5秒）**：仅计算引用测点发生变化的标准测点，结果以`PointIntervalChanged`类型上报

### 3. 引用测点映射

将任务中的表达式映射关系（`ExpressionMap`）解析为内部数据结构：

```go
// ExpressionMap格式: "A=设备ID.测点名;B=设备ID.测点名;..."
// 解析后生成: refPointStrategyMap[引用测点] -> []*localStrategy

type localStrategy struct {
    stdPointName string                         // 标准测点名称: DeviceGid.PointNameEn
    mozuId       int32                          // 模组ID
    refVarMap    map[string]map[string]struct{} // 引用测点 -> 变量名列表
    expression   string                         // 计算表达式
}
```

**特殊情况处理**：
- **常量点**：`ExpressionMap`为空时，该测点为常量点，存储在`refPointStrategyMap["constants"]`

### 4. 表达式计算引擎

使用 `common/util/expr.EvalFloat` 执行表达式计算：

```go
// logic/compute/compute.go
func (d *computeApiImpl) evalPoints(...) {
    for _, stdPoint := range uniqueStdPoints {
        variableValMap := make(map[string]any)
        
        // 填充变量值
        for pointName, variables := range stdPoint.refVarMap {
            if val, ok := refPointsVal[pointName]; ok && val.IsValid(begin.Unix()) {
                for variable := range variables {
                    variableValMap[variable] = val.Value
                }
            } else {
                lessPoint = append(lessPoint, pointName)
            }
        }
        
        // 执行表达式计算
        val, qua, err := expr.EvalFloat(stdPoint.expression, variableValMap)
        newPoint.Quality = int32(qua)
        newPoint.Value = val
    }
}
```

**测点有效性判断**（`Point.IsValid`）：
- 质量值（Quality）必须 >= 0
- 时间戳在当前时间的 [now, now-3分钟] 范围内

**质量码设置**：
- `QualityCalcLessPointErr`：引用测点缺失
- `QualityQueryCacheApiErr`：查询缓存接口失败
- `QualityPushKafkaErr`：推送Kafka失败

### 5. 变化检测机制

```go
// logic/compute/compute.go
func (d *computeApiImpl) calcChangedPoints(begin time.Time, wg *sync.WaitGroup) {
    // 1. 查询短时间内变化的测点
    changedPoints, err := d.cacheApi.ReadChanged(ctx, allRefPointNames, begin.Unix())
    
    // 2. 根据变化的测点找出需要重新计算的策略
    for pointName := range changedPoints {
        if relateStrategies, ok := d.refPointStrategyMap[pointName]; !ok {
            for _, strategy := range relateStrategies {
                needCalcStrategyMap[strategy.stdPointName] = strategy
            }
        }
    }
    
    // 3. 只计算受影响的标准测点
    _, calcChangedPoints := d.calcPoints(lo.Values(needCalcStrategyMap), consts.CalcTypeChanged)
}
```

**变化对比逻辑**：
```go
// 比对测点是否发生变化
lastData, loaded := d.lastValueMap.LoadOrStore(stdPoint.stdPointName, newPoint)
lastPoint := lastData.(*model.Point)
if !loaded || (lastPoint.Value != newPoint.Value || lastPoint.Quality != newPoint.Quality) {
    changedPoints = append(changedPoints, newPoint)
}
```

### 6. Worker心跳注册

定期向 Scheduler 上报心跳，维护 Worker 存活状态：

```go
// logic/register/register.go
func ReportHeartbeat(ctx context.Context, wg *sync.WaitGroup) {
    initWorkerInfo()
    interval := 5 * time.Second
    for {
        select {
        case <-ctx.Done():
            unregister(trpc.BackgroundContext())  // 优雅关闭时注销
            return
        case <-time.After(interval):
            register(ctx)  // 每5秒上报心跳
        }
    }
}

func initWorkerInfo() {
    worker.Ip = trpc.GlobalConfig().Global.LocalIP
    worker.Port = config.GetInt32OrDefault("heartbeat.port", 8080)
    worker.StartTime = time.Now().Unix()
    worker.WorkerType = scheduler.WorkerInfo_POINT        // 测点计算Worker
    worker.WorkerSet = trpc.GlobalConfig().Global.FullSetName
    worker.WorkerProtocol = scheduler.WorkerInfo_HTTP
}
```

**心跳上报内容**：
- `Ip/Port`：Worker地址
- `WorkerType`：POINT（测点计算类型）
- `WorkerStatus`：HEALTHY（健康）/ SHUTDOWN（关闭）
- `TaskVerMark`：当前任务版本标识
- `MaxProcessCap`：最大处理能力

### 7. Kafka结果上报

计算结果通过 Kafka 上报，支持主备双活：

```go
// repo/store/kafka_store.go
func (obj *kafkaStoreImpl) BatchWrite(allPoints []*model.Point, dataType int32) {
    // 按模组ID分组
    mozuPoints := lo.GroupBy(allPoints, func(item *model.Point) int32 {
        return item.MozuId
    })
    
    for mozuId, points := range mozuPoints {
        // 生成Kafka Key
        kafkaKey := kafkamodel.KafkaMsgKey{
            T:     now.Unix(),
            D:     dataType,         // 数据类型：周期/变化
            Type:  consts.PointTypeStd,  // 标准测点类型
            MID:   fmt.Sprint(mozuId),
            PubMs: now.UnixMilli(),
        }
        
        // 分批写入（每批2000条）
        chunkPoints := lo.Chunk(points, 2000)
        for _, chunk := range chunkPoints {
            obj.pushPoints(ctx, keyBytes, valueBytes)
        }
    }
}
```

**主备切换逻辑**：
```go
func (obj *kafkaStoreImpl) pushPoints(ctx context.Context, key, value []byte) (string, bool, error) {
    // 1. 主Kafka推送（重试3次）
    err := retry.Do(func() error {
        return obj.majorKafka.Produce(ctx, key, value)
    }, retry.Attempts(3))
    
    // 2. 主Kafka失败时切换备Kafka
    if err != nil && obj.backupKafka != nil {
        backupErr := retry.Do(func() error {
            return obj.backupKafka.Produce(ctx, key, value)
        }, retry.Attempts(3))
        if backupErr == nil {
            return consts.TbosBackupKafkaName, true, err  // 备Kafka成功
        }
    }
    return consts.TbosMajorKafkaName, err == nil, err
}
```

### 8. 数据查询接口

通过 Data-Cache 服务查询引用测点数据：

```go
// repo/rpc/cache_api_read.go
type ICacheQueryApi interface {
    // 读取最新数据
    ReadLatest(ctx context.Context, pointNames []string, max int64) (map[string]*model.Point, error)
    // 读取变化测点
    ReadChanged(ctx context.Context, pointNames []string, max int64) (map[string]int64, error)
}
```

**分批并发查询**：
- 每批5000条测点
- 并发请求
- 失败重试3次，间隔100ms

### 9. 监控指标

```go
// repo/report/report_metrics.go
var (
    PointCalcExpectCnt  = metric.NewMetric("point_calc_expect_cnt")   // 期望计算数量
    PointCalcSuccessCnt = metric.NewMetric("point_calc_success_cnt")  // 成功计算数量
    PointCalcCost       = metric.NewMetric("point_calc_cost", ...)    // 计算耗时(MAX/AVG)
    
    PointPushSuccessCnt = metric.NewMetric("point_push_success_cnt")  // 推送成功数量
    PointPushFailCnt    = metric.NewMetric("point_push_fail_cnt")     // 推送失败数量
    PointPushCost       = metric.NewMetric("point_push_cost", ...)    // 推送耗时(MAX/AVG)
)
```

**维度标签**：
- `计算类型`：period（周期）/ changed（变化）
- `interval`：数据类型
- `mozu_id`：模组ID
- `source`：数据来源（major/backup）

## 代码结构

```
data-compute/
├── main.go                          # 服务入口
│                                    # - 创建etrpc服务器
│                                    # - 注册优雅关闭hook
│                                    # - 启动计算循环和心跳上报
├── go.mod                           # Go模块定义
├── trpc_go.yaml                     # 服务配置文件
├── entity/                          # 实体定义
│   ├── kafkamodel/                  # Kafka消息模型
│   │   └── point_kafka.go           # KafkaMsgKey、KafkaMsgValue、KafkaMsgPoint
│   └── model/                       # 数据模型
│       └── point.go                 # Point测点结构体、IsValid有效性判断
├── logic/                           # 业务逻辑层
│   ├── compute/                     # 计算核心逻辑
│   │   └── compute.go               # IComputeApi接口实现
│   │                                # - ReceiveTask: 接收任务
│   │                                # - StartCalcPoint: 启动计算循环
│   │                                # - calcFullPoints: 全量计算
│   │                                # - calcChangedPoints: 增量计算
│   │                                # - evalPoints: 表达式求值
│   └── register/                    # 注册逻辑
│       └── register.go              # Worker心跳上报
│                                    # - ReportHeartbeat: 定时心跳
│                                    # - initWorkerInfo: 初始化Worker信息
│                                    # - register/unregister: 注册/注销
├── repo/                            # 数据访问层
│   ├── report/                      # 指标上报
│   │   └── report_metrics.go        # 监控指标定义
│   ├── rpc/                         # RPC调用
│   │   └── cache_api_read.go        # Data-Cache查询接口
│   │                                # - ReadLatest: 查询最新值
│   │                                # - ReadChanged: 查询变化测点
│   │                                # - ReadRange: 范围查询
│   └── store/                       # 数据存储
│       └── kafka_store.go           # Kafka写入
│                                    # - BatchWrite: 批量写入
│                                    # - pushPoints: 主备切换推送
└── service/                         # 服务接口层
    └── compute_service.go           # ComputeService实现
                                     # - ReceiveTask: 接收任务
                                     # - ShowTask: 查询任务
                                     # - ShowData: 查询计算数据
```

## 配置说明

### trpc_go.yaml 配置

```yaml
etrpc:
  service_name: data-compute          # 服务名称
  service_port: ${PORT_DATA_COMPUTE}  # 服务端口

global:
  max_frame_size: 1048576000          # 最大帧大小
  namespace: Production
  local_ip: ${LOCAL_IP}

server:
  filter:
    - recovery
  service:
    - name: ${etrpc.service_name}
      protocol: http
      port: ${etrpc.service_port}

client:
  service:
    # Kafka配置
    - name: trpc.kafka.tbos.major     # 主Kafka
      target: kafka://${KAFKA_ADDR}?topic=${KAFKA_POINT_TOPIC}&...
    
    # Data-Cache服务
    - name: data-cache
      callee: tbos.data.cache.Point
      protocol: http
      target: ip://${LOCAL_IP}:${PORT_DATA_CACHE}
    
    # Scheduler服务
    - name: scheduler
      callee: tbos.scheduler.Register
      protocol: http
      target: ip://${LOCAL_IP}:${PORT_SCHEDULER}

# 心跳配置
heartbeat:
  port: ${PORT_DATA_COMPUTE}          # 心跳上报端口
  max_process_cap: 0                  # 最大处理能力（0表示不限制）
```

### API接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/ReceiveTask` | POST | 接收Scheduler下发的计算任务 |
| `/ShowTask` | POST | 查询当前计算任务列表 |
| `/ShowData` | POST | 查询当前计算结果数据 |

#### ReceiveTask 请求参数

```json
{
  "task_ver_mark": "版本标识",
  "publish_type": 1,           // 1=全量, 0=增量
  "add_task": [{
    "device_gid": "设备GID",
    "point_name_en": "测点英文名",
    "version": "版本号",
    "mozu_id": 123,
    "expression": "A + B * 2",
    "expression_map": "A=设备1.测点1;B=设备2.测点2"
  }],
  "del_task": [...]
}
```

#### ShowTask/ShowData 查询参数

```json
{
  "device_gid": "设备GID",      // 可选，按设备筛选
  "point_name_en": "测点名",    // 可选，按测点名筛选
  "point_key": ["设备.测点"]    // 可选，按完整Key筛选
}
```

## 常见问题

### 1. 测点计算结果质量码异常

**问题表现**：计算出的测点 Quality 不为0

**可能原因及质量码含义**：
- `QualityCalcLessPointErr`：引用测点缺失或数据过期（超过3分钟）
- `QualityQueryCacheApiErr`：查询 Data-Cache 服务失败
- `QualityPushKafkaErr`：推送 Kafka 失败

**解决方案**：
- 检查 Data-Cache 服务是否正常
- 检查引用测点是否正常上报数据
- 检查 Kafka 连接状态

### 2. 计算任务未执行

**问题表现**：下发任务后，测点没有计算结果

**可能原因**：
- 任务版本标识冲突
- 表达式映射解析失败
- Worker未正常注册

**解决方案**：
- 调用 `/ShowTask` 接口检查任务是否正确接收
- 检查日志中是否有任务解析错误
- 确认心跳上报正常

### 3. Kafka推送失败

**问题表现**：日志中出现 "write point to major kafka fail" 告警

**可能原因**：
- Kafka服务不可用
- 网络连接异常
- 消息体过大

**解决方案**：
- 检查 Kafka 服务状态
- 检查网络连通性
- 配置备用 Kafka（`trpc.kafka.tbos.backup`）实现主备切换

### 4. 心跳注册失败

**问题表现**：日志中出现 "register worker failed" 告警

**可能原因**：
- Scheduler 服务不可用
- 网络连接异常
- 配置的端口与实际不符

**解决方案**：
- 检查 Scheduler 服务状态
- 检查 `heartbeat.port` 配置是否正确
- 确认网络连通性

### 5. 变化计算频繁降级为全量

**问题表现**：日志频繁出现 "read batch changed point err"

**可能原因**：
- Data-Cache 服务响应慢
- 查询的测点数量过大

**解决方案**：
- 检查 Data-Cache 服务性能
- 优化测点数量，减少单次查询量
- 增加 Data-Cache 服务资源

### 6. 表达式计算错误

**问题表现**：日志出现 "point[xxx] eval fail"

**可能原因**：
- 表达式语法错误
- 变量名与映射不匹配
- 除数为零等数学错误

**解决方案**：
- 检查任务中的 `expression` 字段语法
- 确认 `expression_map` 中的变量名与表达式一致
- 添加表达式防护逻辑

### 7. 服务优雅关闭时数据丢失

**问题表现**：重启服务后部分计算结果丢失

**原因说明**：
- 服务通过 `context.WithCancel` 实现优雅关闭
- 关闭时会等待执行中的任务完成（通过 `WaitGroup`）
- 但内存中的 `lastValueMap` 不会持久化

**解决方案**：
- 依赖下游 Data-Store 服务的数据存储
- 服务重启后会重新从 Scheduler 获取任务
- 首次计算使用全量模式补全数据

### 8. 如何查看当前计算任务

**解决方案**：
调用 `/ShowTask` 接口：
```bash
curl -X POST http://localhost:8080/ShowTask \
  -H "Content-Type: application/json" \
  -d '{"point_key": ["设备GID.测点名"]}'
```

返回结果包含任务详情：表达式、引用测点映射、模组ID等。