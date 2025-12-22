module.exports = {
  mock: {
    enable: true,
    // restful 模式的 code 会体现在 http status code 上
    restful: true,
    // 非 restful 模式的 code 字段
    codeField: 'code',
    // 非 restful 模式的 data 字段
    dataField: 'data',
    // auto response 返回的 json 格式器，传入原 data 和 code
    formatter: null,
    // model 文件名和路径映射到 path 上
    autolink: false,
    // 延迟时间，单位是 ms，大于 0 时所有的 mock 数据都会经过这个时延
    latency: 0,
  },
  log: 'info',
  // mock 中间件处理的请求
  patterns: ['!**/*.*', '!*webpack_hmr*'],
  prefix: '',
};
