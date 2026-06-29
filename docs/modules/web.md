# 前端 Web

Web是TBOS的前端项目，提供设备监控、告警管理、安防门禁、数据查询及系统配置等可视化操作界面。

## 覆盖范围

| 功能域 | 覆盖内容 | 说明 |
|--------|----------|------|
| **告警管理** | 活动告警、已挂起告警、历史告警、告警策略、策略生效验证 | 告警全生命周期管理：实时监控、挂起/恢复、历史追溯、策略 CRUD 与生效验证 |
| **安防门禁** | 门禁概览、门禁事件、授权发卡、时间组、异步消息 | 门控制器状态监控、出入记录与告警、卡片/持卡人/权限组管理、时间组配置、异步指令跟踪 |
| **数据查询** | 实时数据、历史数据、高级搜索 | 基于设备树的实时/历史测点数据查询、数据导出、全局高级搜索 |
| **系统配置** | 设备管理、模组配置管理 | 采集器/设备树管理、备份还原、3D 预览；模组增删与配置导入 |

## 技术栈

- **核心框架**: Vue 2.6.12 + Vue Router 3.x + Vuex 3.x
- **UI 组件库**: Element UI 2.8.2、TNWeb-ui（组件库）
- **图表可视化**: ECharts 5.x、@antv/g6 4.0.1、Three.js
- **网络请求**: Axios、socket.io-client（WebSocket）
- **构建工具**: Webpack 3.x + Babel 7.x
- **微前端**: single-spa-vue

## 功能模块

项目分为两大核心模块：

### 1. tedge（边端业务模块）

主要业务功能页面，位于 `src/module/tedge/pages/`：

| 页面 | 功能描述 |
|------|----------|
| actived-warning | 当前告警 - 展示未恢复且未挂起的告警列表，支持告警时间线图表 |
| hangup-warning | 挂起告警 - 展示已挂起的告警列表 |
| warning-history | 历史告警 - 展示已恢复的告警记录，支持分析统计 |
| warning-strategy | 告警策略 - 设备告警策略配置管理 |
| warning-strategy-detail | 告警策略详情 - 策略详细配置页面 |
| warning-detail | 告警详情 - 单条告警的详细信息展示 |
| warning-effective-verify | 策略生效验证 - 验证告警策略是否正确生效 |
| data-query-index | 数据查询首页 - 按位置/类型查询设备测点数据 |
| data-query-detail | 数据查询详情 - 测点历史数据详情 |
| advanced-search | 高级搜索 - 历史数据高级查询，支持多测点对比 |
| device-manage | 设备管理 - 采集设备管理 |
| menu-config | 菜单配置 - 系统菜单配置管理 |
| mozu-config-manage | 模组配置管理 - 模组相关配置 |

### 2. tnebula（星云平台框架模块）

框架和适配器相关，位于 `src/module/tnebula/`：

| 目录/文件 | 功能描述 |
|-----------|----------|
| pages/index/ | 主框架页面（frameNew.vue），包含导航栏、侧边栏、菜单等 |
| pages/adaptor/ | 边端适配器框架，用于 tedge 模块的页面容器 |
| pages/common/ | 公共页面（404、权限等） |
| config/api.js | API 接口定义 |
| config/authCode.js | 权限码配置 |

## 核心能力

### 1. 告警管理

基于 `src/feature/warning/` 实现完整的告警生命周期管理：

- **当前告警**: 实时展示活动告警，支持按等级（L0-L5）分类统计，提供告警时间线图表
- **挂起告警**: 管理已挂起的告警，支持恢复操作
- **历史告警**: 查询已恢复告警，支持数据导出和统计分析
- **告警策略**: 配置设备告警规则，支持策略基线对比和操作日志

### 2. 数据查询

基于 `src/feature/adaptor/data-query/` 和 `src/feature/adaptor/advanced-search/` 实现：

- **设备树导航**: 支持按位置、按类型两种模式浏览设备树
- **测点实时数据**: 查询设备测点的实时数值
- **历史数据查询**: 查询测点历史数据，支持时间范围和采样间隔配置
- **高级搜索**: 多测点对比分析，支持图表展示和数据导出

### 3. 设备管理

基于 `src/feature/systemconfig/device-manage/` 实现：

- **采集设备树**: 展示采集器和设备的层级结构
- **设备信息**: 查看设备详细信息和状态
- **设备配置**: 支持设备的新增、编辑操作

### 4. 实时数据服务

基于 `src/services/` 实现多种数据服务：

- **轮询代理服务** (`polling-request-proxy/`): 支持 WebSocket 和 HTTP 两种模式的轮询请求代理，自动切换降级
- **测点实时数据服务** (`tedge/check-point-realtime-data.service.ts`): 支持缓存模式的测点数据获取，优化重复请求性能
- **设备树服务** (`tedge/device-tree.service.ts`): 设备树数据获取，支持 V1/V2 两种版本
- **自定义配置服务** (`custom-config.service.ts`): 页面级、模块级的动态配置管理

### 5. API 转换层

项目实现了请求/响应转换机制，用于适配不同后端系统：

- 接口转换文件位于 `web/src/services/polling-request-proxy/transformMap.ts` 和 `web/src/module/tnebula/pages/adaptor/assets/tbos-transform/transformMap.json`两个文件相同
- 在`web/src/module/tnebula/pages/adaptor/assets/tbos-transform/funcs`目录下修改文件后，运行 `node transFuncToObj.js` 打包成上一步的转换文件
- 支持请求参数转换和响应数据转换
- 通过 `isTbos` 标识区分 TBOS 模式和其他模式

### 6. 公共组件库

位于 `src/feature/component/`：

| 组件 | 功能描述 |
|------|----------|
| ConfigPanel/ | 配置面板，包含高级搜索、表格、表单弹窗等 |
| chart/line-chart.vue | 折线图组件，基于 ECharts 封装 |
| Table/ | 通用表格组件 |
| searchComponent/ | 搜索组件，支持高级搜索和日历视图 |
| MergeAlarm/ | 告警合并组件 |
| imFrame/ | iframe 嵌入组件 |
| processLog/ | 流程日志组件 |
| user/user-selector.vue | 用户选择器组件 |
| tedge-components/ | tedge 专用组件集合 |

## 代码结构

```
web/
├── build/                          # 构建相关配置和脚本
│   ├── bin/                        # 命令行工具
│   │   ├── tnfusion/               # 融合构建主命令
│   │   ├── tnfusion-init/          # 项目初始化
│   │   └── tnfusion-run/           # 运行命令入口
│   ├── build/                      # Webpack 构建配置
│   │   ├── webpack.base.conf.js    # 基础配置
│   │   ├── webpack.dev.conf.js     # 开发环境配置
│   │   ├── webpack.prod.conf.js    # 生产环境配置
│   │   └── fusion.js               # 融合构建核心逻辑
│   └── config/                     # 构建环境配置
│
├── config/                         # 项目配置
│   ├── .common.js                  # 通用配置（别名、外部依赖等）
│   ├── tedge.js                    # tedge 模块配置
│   └── tnebula.js                  # tnebula 模块配置
│
├── deploy/                         # 部署相关
│   ├── deploy.sh                   # 部署脚本
│   └── tnebula/                    # 星云平台部署配置
│       ├── dockerfile
│       └── nginx.conf
│
├── src/                            # 源代码
│   ├── feature/                    # 功能模块（核心业务组件）
│   │   ├── adaptor/                # 适配器相关页面
│   │   │   ├── data-query/         # 数据查询页面
│   │   │   ├── advanced-search/    # 高级搜索页面
│   │   │   ├── menu-config/        # 菜单配置
│   │   │   └── mozu-config-manage/ # 模组配置管理
│   │   ├── warning/                # 告警相关组件
│   │   │   ├── actived-warning/    # 当前告警
│   │   │   ├── hangup-warning/     # 挂起告警
│   │   │   ├── warning-history/    # 历史告警
│   │   │   ├── warning-strategy/   # 告警策略
│   │   │   ├── warning-detail/     # 告警详情
│   │   │   └── strategy-effective-verify/ # 策略生效验证
│   │   ├── systemconfig/           # 系统配置
│   │   │   └── device-manage/      # 设备管理
│   │   ├── component/              # 公共组件库
│   │   ├── config/                 # CGI/HTTP 配置
│   │   ├── utils/                  # 工具函数
│   │   └── style/                  # 样式文件
│   │
│   ├── module/                     # 应用模块入口
│   │   ├── index/                  # 默认首页入口
│   │   ├── tedge/                  # 边端模块
│   │   │   ├── pages/              # 页面入口
│   │   │   ├── config/             # 模块配置
│   │   │   └── script/             # 脚本（spa 入口封装等）
│   │   └── tnebula/                # 星云平台模块
│   │       ├── pages/              # 框架页面
│   │       │   ├── index/          # 主框架
│   │       │   ├── adaptor/        # 边端适配器框架
│   │       │   └── common/         # 公共页面
│   │       └── config/             # API 和权限配置
│   │
│   ├── services/                   # 服务层
│   │   ├── custom-config.service.ts        # 自定义配置服务
│   │   ├── login-status.service.ts         # 登录状态服务
│   │   ├── v2-device-number-transformer.service.ts # 设备编号转换
│   │   ├── polling-request-proxy/          # 轮询请求代理
│   │   │   ├── polling-proxy.service.ts    # 代理主服务
│   │   │   ├── transformMap.ts             # API 转换映射
│   │   │   └── plugin/                     # 插件系统
│   │   └── tedge/                          # tedge 专用服务
│   │       ├── device-tree.service.ts      # 设备树服务
│   │       └── check-point-realtime-data.service.ts # 测点实时数据
│   │
│   ├── static/                     # 静态资源
│   │   ├── thirdparty/             # 第三方库（Vue、Element UI、ECharts 等）
│   │   ├── tnweb-common-utils/     # 公共工具库
│   │   ├── css/                    # 全局样式
│   │   ├── fonts/                  # 字体文件
│   │   ├── images/                 # 图片资源
│   │   └── audio/                  # 音频文件（告警提示音）
│   │
│   ├── template/                   # HTML 模板
│   ├── typings/                    # TypeScript 类型定义
│   └── utils/                      # 工具函数
│       ├── axios-methods.ts        # Axios 请求封装
│       ├── tree.ts                 # 树形数据处理
│       ├── xlsx-utils.ts           # Excel 处理
│       ├── download.ts             # 下载工具
│       └── pagination.ts           # 分页工具
│
├── package.json                    # 项目依赖配置
├── tsconfig.json                   # TypeScript 配置
├── babel.config.js                 # Babel 配置
├── Dockerfile                      # Docker 构建文件
└── nginx.conf                      # Nginx 配置
```

## 常见问题

### 1. 如何安装依赖？

```bash
# 要求 Node.js >= 14.17.0
npm install
```

### 2. 如何本地开发运行？

```bash
# 运行 tedge 全部页面
npm run dev tedge

# 运行 tedge 单个页面（页面名称对应 src/module/tedge/pages/ 下的目录）
npm run dev tedge warning          # 告警相关页面
npm run dev tedge data             # 数据查询页面
npm run dev tedge actived-warning  # 当前告警页面

# 运行菜单框架 
npm run dev tnebula adaptor

# 菜单框架打包 改动后需执行生效
npm run adaptordev
```

### 3. 如何打包构建？

```bash
# 完整打包（先构建 tnebula，再构建 tedge）
npm run public-tedge

# 单独构建 tnebula
npm run bud-tnebula

# 单独构建 tedge
npm run bud-tedge
```

### 4. 如何添加新页面？

1. 在 `src/module/tedge/pages/` 下创建新目录
2. 创建 `main.js` 入口文件：
   ```javascript
   import simpleEntry from '@@/script/spa';
   import page from 'feature/xxx/index.vue';
   export default simpleEntry(page);
   ```
3. 在 `src/feature/` 下创建对应的页面组件
4. 在 `config/tedge.js` 中配置页面信息

### 5. 如何配置 API 接口？

- 通用接口配置：`src/feature/config/cgi.js`
- tedge 模块接口：`src/module/tedge/config/cgi.js`
- tnebula 模块接口：`src/module/tnebula/config/api.js`

### 6. 如何使用轮询代理服务？

```javascript
// 获取轮询代理服务实例
const pollingService = window.tnwebServices.pollingProxyAgentService;

// 启动轮询代理
const proxy = pollingService.proxy({
  request: {
    url: '/cgi/xxx',
    method: 'POST',
    data: { ... }
  },
  interval: 5000  // 轮询间隔（毫秒）
}, (data) => {
  // 数据回调
  console.log(data);
});

// 退出轮询
pollingService.exit([proxy]);
```

### 7. 如何使用自定义配置服务？

```javascript
// 获取配置值
const value = window.tnwebServices.customConfigService.get('ConfigKey');

// 在组件中使用 custom-config-value 组件
<custom-config-value name="EnableDeviceNumberV2">
  <template #default="{ value }">
    <!-- 根据配置值渲染内容 -->
  </template>
</custom-config-value>
```

### 8. Webpack 别名说明

在 `config/.common.js` 中配置了以下别名：

| 别名 | 路径 |
|------|------|
| `@template` | `src/template` |
| `@module` | `src/feature/${moduleName}` |
| `common` | `src/static/tnweb-common-utils/dist` |
| `feature` | `src/feature` |
| `component` | `src/feature/component` |
| `@@` | `src/module/${moduleName}` |

### 9. 外部依赖（CDN 加载）

以下依赖通过 CDN 加载，不打包进 bundle：

- Vue、Vue-i18n
- Element UI、TNWeb-ui
- ECharts
- jQuery
- Moment.js
- Axios
- Lodash

### 10. 如何调试 API 转换？

在浏览器控制台执行：

```javascript
localStorage.setItem('logTransform', 'true');
```

刷新页面后，API 转换日志会输出到控制台。关闭调试：

```javascript
localStorage.removeItem('logTransform');
```
