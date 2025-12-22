# Common（公共模块）

## 一、模块介绍

Common是TBOS动环系统的公共模块，提供各业务服务共享的实体定义、常量定义和工具类，作为独立的Go Module被其他服务引用。

### 1.1 主要职责

| 职责 | 说明 |
|------|------|
| 实体定义 | 提供数据库表对应的GORM实体模型 |
| 常量定义 | 提供测点类型、数据质量、客户端名称等公共常量 |
| 工具类 | 提供表达式计算、分布式锁等通用工具 |

### 1.2 模块依赖关系

```
┌─────────────────────────────────────────────────────────────────────┐
│                        TBOS 各业务服务                               │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐           │
│  │  CMDB     │ │ Scheduler │ │ Data-Comp │ │ Collector │  ...      │
│  └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘           │
└────────┼─────────────┼─────────────┼─────────────┼──────────────────┘
         │             │             │             │
         └─────────────┴──────┬──────┴─────────────┘
                              │ import
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          Common 模块                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐      │
│  │  entity/consts  │  │  entity/model   │  │     util        │      │
│  │  公共常量定义     │  │  GORM实体模型   │  │   工具类        │      │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 二、核心能力

### 2.1 公共常量定义

#### 2.1.1 客户端名称常量

| 常量名 | 值 | 说明 |
|--------|-----|------|
| `TbosMysqlName` | `trpc.mysql.tbos` | MySQL客户端名称 |
| `TbosRedisName` | `trpc.redis.tbos` | Redis客户端名称 |
| `TbosInfluxName` | `trpc.influx.tbos` | InfluxDB客户端名称 |
| `TbosMajorKafkaName` | `trpc.kafka.tbos.major` | 主用Kafka客户端名称 |
| `TbosBackupKafkaName` | `trpc.kafka.tbos.backup` | 备用Kafka客户端名称 |

#### 2.1.2 测点类型常量

| 常量名 | 值 | 说明 |
|--------|-----|------|
| `PointTypeCollect` | 1 | 采集测点 |
| `PointTypeStd` | 2 | 标准测点 |
| `PointTypeVirtual` | 3 | 虚拟测点 |
| `PointTypeAlarm` | 4 | 告警测点 |

#### 2.1.3 测点分类常量

| 常量名 | 值 | 说明 |
|--------|-----|------|
| `PointCategoryUnknown` | 0 | 未知分类 |
| `PointCategoryAllCollect` | 1 | 全采集（无需标准化计算） |
| `PointCategoryCollectStd` | 2 | 采集+标准（需查询采集测点计算） |
| `PointCategoryAllStd` | 3 | 全标准（需查询标准测点计算） |

#### 2.1.4 数据质量常量

| 常量名 | 值 | 说明 |
|--------|-----|------|
| `QualityOk` | 0 | 正常 |
| `QualityPushKafkaErr` | -600 | 测点推送Kafka失败 |
| `QualityCalcLessPointErr` | -601 | 测点计算缺少测点 |
| `QualityQueryCacheApiErr` | -602 | 查询缓存API错误 |
| `QualityStdExprErr` | -900 | 标准化表达式错误 |
| `QualityStdEvalErr` | -901 | 标准化计算错误 |
| `QualityStdValTypeErr` | -904 | 标准点数据类型非法 |
| `QualityStdValNilErr` | -905 | 标准点值为空(nil) |
| `QualityStdNaNInfErr` | -906 | 标准点值为NaN或Inf |

### 2.2 GORM实体模型

#### 2.2.1 告警相关模型

| 实体 | 表名 | 说明 |
|------|------|------|
| `AlarmActive` | `t_alarm_active` | 活动告警，包含告警ID、级别、发生时间、设备信息等 |
| `AlarmHistory` | `t_alarm_history` | 历史告警，包含恢复时间、恢复分析结果等 |
| `AlarmStrategy` | `t_alarm_strategy` | 告警策略，包含告警/恢复表达式、级别、内容模板等 |

**告警状态常量**：
- `ActiveAlarmCode = 0`：未挂起活动告警
- `HangupAlarmCode = 1`：挂起告警

**主要方法**：
- `AlarmActive.ConvertToAlarmMsg()`：转换为Markdown格式的告警通知消息
- `AlarmHistory.ConvertToRestoreMsg()`：转换为Markdown格式的恢复通知消息
- `ActiveAlert2History()`：活动告警转历史告警
- `AlarmStrategy.GetExprMap()`：获取策略的测点映射关系

#### 2.2.2 采集设备相关模型

| 实体 | 表名 | 说明 |
|------|------|------|
| `CollectorDevice` | `t_collector_device` | 采集设备，包含设备GID、编号、通道信息、模板等 |
| `CollectorTemplate` | `t_collector_template` | 采集模板，包含协议类型、版本、设备型号等 |
| `CollectorTemplatePoint` | `t_collector_template_point` | 采集模板测点，包含测点定义、协议定义等 |

**采集器类型常量**：
- `CollectorTypeTbox = 1`：TBOX采集器
- `CollectorTypeTboxSubDevice = 2`：TBOX下子设备
- `CollectorTypeVendorBox = 3`：厂商采集器
- `CollectorTypeVendorSubDevice = 4`：厂商下子设备

#### 2.2.3 设备相关模型

| 实体 | 表名 | 说明 |
|------|------|------|
| `DeviceEntity` | `t_device_entity` | 设备实体，包含设备GID、编号、名称、所属模组等 |
| `DevicePoint` | `t_device_point` | 设备测点，包含测点表达式、映射关系、值类型等 |
| `MozuInfo` | `t_mozu_info` | 模组信息，包含模组ID、名称、发布版本等 |

**DevicePoint主要方法**：
- `CalcDependCollector()`：计算测点依赖的采集器
- `CollectorToStdPoint()`：采集到标准测点映射转换
- `StdToCollectorPoint()`：标准到采集测点映射转换（支持AST表达式解析和嵌套展开）
- `FixGid()`：修正room类设备GID

### 2.3 表达式计算引擎

基于`github.com/expr-lang/expr`实现的表达式计算工具，支持测点标准化计算和告警规则计算。

#### 2.3.1 内置函数

| 函数 | 说明 |
|------|------|
| `max/MAX` | 计算最大值 |
| `min/MIN` | 计算最小值 |
| `sum/SUM` | 计算求和 |
| `avg/AVG` | 计算平均值 |
| `abs/ABS` | 计算绝对值 |
| `eq/EQ` | 判断相等（支持bool与数值比较） |
| `neq/NEQ` | 判断不等 |

#### 2.3.2 表达式转换

支持以下语法转换：
- `if(condition, true_expr, false_expr)` → 三元表达式 `condition ? true_expr : false_expr`
- `and` → `&&`
- `or` → `||`
- `Avg/Min/Max` → 小写形式
- 负数变量处理：`B+-A` → `B+(-A)`

#### 2.3.3 主要接口

| 函数 | 说明 |
|------|------|
| `Eval(expression, parameters)` | 执行表达式计算，返回结果和质量码 |
| `EvalStr(expression, parameters)` | 执行表达式计算，返回字符串结果 |
| `EvalFloat(expression, parameters)` | 执行表达式计算，返回float64结果 |
| `RegisterCommonParameter(name, parameter)` | 注册自定义函数或常量 |

### 2.4 分布式锁工具

基于Redis的分布式锁封装，简化锁的使用。

**函数签名**：
```go
func DisLock(ctx context.Context, redisName, key string, handler func(), options ...redlock.Option) error
```

**特性**：
- 默认锁过期时间60秒
- 支持通过options自定义过期时间
- 自动释放锁（defer机制）
- 支持错误处理和日志记录

---

## 三、代码结构

```
common/
├── go.mod                                   # Go Module定义
├── go.sum                                   # 依赖校验文件
├── README.md                                # 说明文档
├── entity/                                  # 实体定义
│   ├── consts/                              # 常量定义
│   │   ├── const.go                         # 客户端名称、测点类型等常量
│   │   └── qua.go                           # 数据质量常量定义
│   └── model/                               # GORM模型定义
│       ├── alarm_active.go                  # 活动告警模型
│       ├── alarm_history.go                 # 历史告警模型
│       ├── alarm_strategy.go                # 告警策略模型
│       ├── collector_device.go              # 采集设备模型
│       ├── collector_template.go            # 采集模板模型
│       ├── collector_template_point.go      # 采集模板测点模型
│       ├── device_entity.go                 # 设备实体模型
│       ├── device_point.go                  # 设备测点模型
│       └── mozu_info.go                     # 模组信息模型
└── util/                                    # 工具类
    ├── dislock/
    │   └── dislock.go                       # 分布式锁封装
    └── expr/
        ├── eval.go                          # 表达式计算引擎
        └── transform.go                     # 表达式语法转换
```

### 核心文件说明

| 文件 | 说明 |
|------|------|
| `entity/consts/const.go` | 定义客户端名称、测点类型、测点分类等常量 |
| `entity/consts/qua.go` | 定义数据质量类型和错误码 |
| `entity/model/device_point.go` | 最核心的设备测点模型，包含表达式解析和AST操作 |
| `entity/model/alarm_strategy.go` | 告警策略模型，包含表达式映射解析 |
| `util/expr/eval.go` | 表达式计算引擎，支持缓存编译结果 |
| `util/expr/transform.go` | 表达式语法转换，兼容多种写法 |
| `util/dislock/dislock.go` | 分布式锁封装，简化锁操作 |

---

## 四、常见问题

### 4.1 表达式计算返回质量码异常

**问题描述**：调用`Eval()`函数返回非`QualityOk`的质量码。

**可能原因及解决方案**：

| 质量码 | 原因 | 解决方案 |
|--------|------|----------|
| `QualityStdExprErr(-900)` | 表达式语法错误 | 检查表达式语法是否正确 |
| `QualityStdEvalErr(-901)` | 计算执行错误 | 检查变量是否都已传入parameters |
| `QualityStdValTypeErr(-904)` | 返回值类型非法 | 确保表达式返回数值或布尔类型 |
| `QualityStdValNilErr(-905)` | 返回值为nil | 检查表达式是否有返回值 |
| `QualityStdNaNInfErr(-906)` | 返回NaN或Inf | 检查是否有除零等异常计算 |

### 4.2 分布式锁获取失败

**问题描述**：调用`DisLock()`返回错误。

**可能原因**：
1. Redis连接异常
2. 锁已被其他进程持有

**解决方案**：
1. 检查Redis客户端配置和网络连通性
2. 确认是否有其他进程长时间持有锁
3. 适当调整锁过期时间

### 4.3 实体唯一标识计算

**问题描述**：如何正确使用实体的`CalcUniqueKey()`方法？

**说明**：每个实体都实现了`CalcUniqueKey()`方法，返回该实体在业务上的唯一标识：

| 实体 | 唯一标识格式 |
|------|-------------|
| `AlarmStrategy` | `{DeviceGid}|{Rid}|{RidVersion}` |
| `CollectorDevice` | `{DeviceGid}|{TemplateName}` |
| `CollectorTemplate` | `{TemplateName}` |
| `CollectorTemplatePoint` | `{TemplateName}|{PointNameEn}|{SubDevice}` |
| `DeviceEntity` | `{DeviceGid}.{MozuId}` |
| `DevicePoint` | `{DeviceGid}|{PointNameEn}` |

### 4.4 测点表达式展开

**问题描述**：如何将标准测点的嵌套表达式展开为采集测点表达式？

**解决方案**：使用`DevicePoint.StdToCollectorPoint()`方法，该方法会：
1. 解析当前测点的表达式为AST
2. 递归查找引用的标准测点
3. 将嵌套的标准测点表达式替换为采集测点引用
4. 生成最终展开的表达式和映射关系

### 4.5 如何引用Common模块

**问题描述**：其他服务如何引用Common模块？

**解决方案**：在`go.mod`中添加依赖：
```go
require common v0.0.0

replace common => ../common
```

然后在代码中导入：
```go
import (
    "common/entity/consts"
    "common/entity/model"
    "common/util/expr"
    "common/util/dislock"
)
```

### 4.6 表达式语法兼容问题

**问题描述**：历史配置的表达式使用了`if()`语法或大写函数名。

**解决方案**：表达式引擎内置了语法转换功能，自动兼容以下写法：
- `if(a>0, 1, 0)` 会自动转换为 `a>0 ? 1 : 0`
- `a and b` 会自动转换为 `a && b`
- `Avg(a,b)` 会自动转换为 `avg(a,b)`

无需手动修改历史配置。
