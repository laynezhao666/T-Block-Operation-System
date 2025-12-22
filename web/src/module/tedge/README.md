# 边端动环开发指南

## 模组信息
- 之前有几种方式获取，如TNBL.getCurModule、window.__GetFrameDataByKey、调用接口/cgi/dataQuery/edge/getEdgeLocation
- 存在这么多种方式，说明都不好用，或者没有文档等
- 现新开发的页面等统一在vue中使用：this.$moduleInfo.xxx
- Vue.prototype.$moduleInfo已添加ts类型描述
