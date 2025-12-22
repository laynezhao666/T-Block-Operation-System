# Changelog

> **Frame Tags:**  type:content
> **Src Tags:**   type(scope):content
>
> - :boom:     [feat :新功能]
> - :bug:        [fix :修改bug]
> - :memo:       [docs :文档修改]
> - :bird:       [style :格式修改]
> - :wrench:     [test :增加测试代码]
> - :house:       [chore :构建过程或辅助工具的变动]
> - :rocket:      [perf :性能优化]
> - :wrench:      [revert :版本回退]
> - :bird:      [refactor: 代码重构]

> **Example:** 
> - feat:框架增加分包功能
> - fix(sm):修改分页bug
> - style(fo):修改巡检详情页样式 
>


_Note: Gaps between patch versions are faulty, broken or test releases._

See [CHANGELOG](/build/CHANGELOG.md).
See [README](/README.md).

<!-- insert-new-changelog-here -->

## v1.0.0 (2018-12-14)
创建新框架基本目录结构

## v1.1.1 (2018-12-17)
* `feat` - 添加watch功能
* `docs` - 添加readme

## v1.2.0 (2018-12-19)
* `feat` - 添加gulp配置

## v1.3.0 (2018-12-19)
* `feat` - 增加代码标准按照atandard标准格式化代码

## v1.4.0 (2019-01-02)
* `feat` - 分拆webpack配置

## v1.5.0 (2019-01-07)
* `feat` - gulp打包：加入文件发布url配置

## v2.0.1 (2019-01-12)
* `refactor` - 框架改造，支持vue脚手架搭建项，修改框架目录结构，目
* `feat` 
  * 使用express框架，搭建开发环境node服务
  * 新增多项目，多页面打包配置提取辅助工具
  * 修改配置，支持多项目共用资源
  * 新增开发模式下域名自定义配置

* `fix` 
  * 完善目录结构，新增公共组件，工具，静态资源目录
  * 修复sass打包错误问题
  * 删除根目录static文件夹
  * 引入sm源码，修复sm源码适配框架

## v2.1.1 (2019-01-13)
* `feat` 
  * 修改打包结构，支持项目单独打包
  * 打包支持项目级别静态资源共用
  * 新增fusion所有项目页面的主入口index占用
  * 支持命令行输入项目名称
  * 新增dev模式浏览器默认打开页面配置

* `fix` 
  * 优化打包配置，及webpack配置文件控制台打印

## v2.2.1 (2019-01-19)
* `feat` 
  * 提取项目配置，独立配置文件，及公共配置项提取
  * 增加dev模式下端口检测功能

* `fix` 
  * 修改服务启动默认界面错误问题


## v2.3.0 (2019-01-20)
* `feat` 
  * 增加静态文件夹static，打包直接拷贝功能
  * 增加项目通用index.html模板功能
  * 取消页面级别的index模板
  * 简化了入口文件，支持自定义入口文件名称
  * 提取开发模式的配置信息到src


## v2.4.2 (2019-01-21)
* `feat` 
  * 增加开发模式下热重载功能

* `fix`
  * 修复打包后图片相对路径问题
  * 修复打包后资源丢失问题
  * 增加fo系统老样式，解决开发模式界面变形问题

* `docs`
  * 更新readme，添加文件夹说明

## v2.5.0 (2019-01-25)
* `feat` 
  * 支持解析ts文件，增加了ts-demo


## v2.6.1 (2019-01-28)
* `feat` 
  * 添加了php打包逻辑
  * 优化了报错信息
  * 增加了eslint配置文件，支持eslint-config-standard规范
  * 升级vue-loader，style-loader，修改vue-loader的plugin配置

* `fix`
  * eslint问题修复


## v2.7.0 (2019-01-30)
* `feat` 
  * 修改eslint，js的报错检查
  * 增加了sytlelint检查


## v2.8.1 (2019-02-01)
* `feat` 
  * 添加git hooks功能，commit前检查代码规范性
  * 增加了sytlelint检查

* `perf` 
  * 添加eslint ignore，忽略dist,node_modules

## v2.8.2 (2019-02-13)
* `fix` 
  * 提交vscode配置项

  
## v2.8.4 (2019-02-14)
* `fix` 
  * 修改git add自动修复

* `perf` 
  * 移出vscode不必要配置项
  
## v2.9.2 (2019-02-15)
* `feat` 
  * 添加了公共组件
  * 添加了组件demo页面

* `fix` 
  * hook添加scss的format
  * 修改开发和生产环境包依赖
  * 添加了vscode setting配置

* `perf` 
  * 移出vscode不必要配置项
  
## v2.10.1 (2019-02-18)
* `feat` 
  * 升级babel^7.2.2
  * 升级cssnano^4.1.10
  * 公用console代码加入
  * 提起开发和生产的公用配置
  * 增加src.local.js本地配置文件读取

* `fix` 
  * 移出样式eslint --fix
  
## v2.11.1 (2019-02-19)
* `feat` 
  * 升级element-ui^2.5.4

* `fix` 
  * 修复代码检测
  * 补充样式，table修改
  
## v2.11.3 (2019-02-22)
* `docs` 
  * 更新readme

* `fix` 
  * php plugin换个实现模式
  
## v2.12.1 (2019-02-25)
* `feat` 
  * 添加mock功能
  * mock功能配置放在个个项目中

* `fix` 
  * 添加mock配置文件存在性的校验
  
## v2.13.1 (2019-02-26)
* `feat` 
  * 添mock支持参数配置

* `fix` 
  * 热更新增加debounce节流
  * 为formatter添加method参数
  * 修复dev 服务器启动自动打开浏览器页面错误
  
## v3.1.0 (2019-02-27)
* `refactor` - 框架改造，页面单独打包发布

* `feat` 
  * 框架支持单页面打包发布
  
## v3.2.1 (2019-02-28)
* `feat` 
  * 新增ui打包工具fusion.exe

* `fix` 
  * mock热更新的缓存规则


## v3.3.0 (2019-03-01)
* `feat` 
  * 命令行增加项目名称模糊输入

## v3.4.2 (2019-03-03)
* `feat` 
  * 增加页面单独build
  * 增加命令行打包页面的模糊输入
  * 增加了webpcak配置信息打印的展示隐藏功能

* `fix` 
  * 终端配置信息输出的修改

* `perf` 
  * dev 支持启动某个页面

## v3.5.1 (2019-03-05)
* `feat` 
  * 框架增加fusion全局命令
  * 框架支持项目脚手架功能

* `fix` 
  * 修复脚手架生成项目后配置文件不存在问题

## v3.5.2 (2019-03-06)
* `fix` 
  * 修复ui样式问题

## v3.5.2 (2019-03-07)
* `fix` 
  * mock文件夹添加示例

## v3.6.0 (2019-03-08)
* `feat` 
  * 修改了mockrule规则，自动读取model下文件夹生成cgi

## v3.7.1 (2019-03-11)
* `feat` 
  * 升级vue^2.6.8

* `fix` 
  * 更新组件demo
  * 更新table组件部分


## v3.7.2 (2019-03-12)
* `fix` 
  * 更新组件，增加日期范围筛选header


## v3.8.1 (2019-03-13)
* `feat` 
  * mock增加excel下载
  
* `fix` 
  * 修改cgi路径匹配规则，使用^作为路径分割符


## v3.9.0 (2019-03-14)
* `feat` 
  * mock增加autolink功能
  * mock增加prefix支持
  * mock增加log等级配置


## v3.9.1 (2019-03-18)
* `fix` 
  * 添加mock enable选项

## v3.9.2 (2019-03-19)
* `docs` 
  * 添加框架的changelog

## v3.9.3 (2019-03-20)
* `fix` 
  * common.entryjs合并配置文件重复问题

## v3.10.0 (2019-03-25)
* `feat` 
  * mock添加latency
  * mock延时增加两种方法

## v3.10.1 (2019-04-02)
* `fix` 
  * 浏览器默认打开页面错误

## v3.11.1 (2019-04-12)
* `feat` 
  * 支持pages目录

## v3.11.2 (2019-04-15)
* `fix` 
  * linux环境打开浏览器失败错误关闭
  * 修改框架变量命名

## v3.11.3 (2019-04-16)
* `fix` 
  * 修复跨pages文件夹打包多页面bug

## v3.12.0 (2019-04-22)
* `feat` 
  * 框架打包发布时，输入页面的参数支持*,正则匹配

## v3.13.0 (2019-04-25)
* `feat` 
  * webpack-dev-server支持path rewrite（需配置），避免页面内正式链接的跳转失败问题

## v3.13.1 (2019-04-28)
* `fix` 
  * rewritePaths为空的情况

## v3.14.0 (2019-04-29)
* `feat` 
  * 框架增加异步组件打包自定义文件夹归类功能

## v3.15.0 (2019-05-09)
* `perf` 
  * 解决webpack多页面热更新编译缓慢问题

## v3.17.0 (2019-05-11)
* `perf` 
  * 修改dev模式公共库提取逻辑，减少内存开销，同时也加快了编译速度，
  * 经测试，优化后的页面平均减少800kb，
  * 模板缓存，编译速度从分钟级、秒级变为毫秒级别完成

* `feat` 
  * html模板文件支持引入模板块

* `docs` 
  * 修改框架changelog

## v3.18.0 (2019-05-14)
* `feat` 
  * 框架增加自定义模板配置功能

## v3.19.0 (2019-06-04)
* `feat` 
  * 框架增加检测pages文件夹

## v3.20.3 (2019-06-04)
* `feat` 
  * 框架增加公共static
  * 框架增加公用index.html功能
  * 框架增加template配置功能

## v3.21.0 (2019-06-19)
* `feat` 
  * 框架增加配置url前缀功能


