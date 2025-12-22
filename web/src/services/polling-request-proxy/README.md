# 轮询代理Agent
代理接口轮询，有websocket/http两种模式，websocket有效时优先使用该模式。

## websocket机制
1. 如果有websocket中间层支持，与中间层建立websocket
2. 将请求数据发送给中间层
3. 由中间层代理执行请求
4. agent订阅websocket的请求结果数据，每当接收到服务请求结果则回调调用者

## http机制
1. 前端定时调用http接口，获取了结果后回调
2. 上一次请求结束后才开始下一次轮询定时


## websocket、http升级、降级机制
1. websocket断开连接后，启动http轮询
2. websocket重新连接后，停用http轮询，启用websocket模式，并重新发送代理请求

## 基础用法
```javascript
// 建立代理
const proxy = pollingProxyAgentService.proxy({
  interval: 1000, // 轮询间隔，单位毫秒
  request: { // 请求配置，同axios.request({ ... })的参数
    url: '...请求地址',
    method: 'POST',
    data: payload, // 请求体
  },
}, (data) => {
  // 请求数据结果回调
});

// 当不需要代理是，请调用退出代理
pollingProxyAgentService.exit([proxy]);
```

## 用法：轮询切换器
通过创建轮询切换器，可在每次调用request时停止上一次轮询、启用新的轮询。
>> 注意如果切换的两次请求的JSON.stringify一模一样，则不会做任何处理

```javascript
const proxySwitcher = new PollingProxySwitcher({
  interval: 1000, // 轮询间隔，单位毫秒
});

// 轮询请求A
proxySwitcher.proxy({ // 请求配置，同axios.request({ ... })的参数
  url: '...请求地址A',
  method: 'POST',
  data: payload, // 请求体
}, (data) => {
  // 请求数据结果回调
});

// 轮询请求B，终止请求A，若A与B请求一模一样，不执行任何操作，也不会重置轮询等候时间
proxySwitcher.proxy({ // 请求配置，同axios.request({ ... })的参数
  url: '...请求地址B',
  method: 'POST',
  data: payload, // 请求体
}, (data) => {
  // 请求数据结果回调
});

// 不需要轮询时，取消轮询
proxySwitcher.cancel();
```
