function isPlainObject (obj) {
  return typeof obj === 'object' && obj !== null && Object.getPrototypeOf(obj) === Object.prototype
}

export default function (rspData) {
  // wrapper(rsp.data) -> bizData
  if (rspData && isPlainObject(rspData) && !rspData.isBlob) {
    const { code, data, message } = rspData
    if (code === 0) {
      return data
    } else {
      let rst
      if (this.context.opts.codeHandler) {
        rst = this.context.opts.codeHandler.call(this, code, message, data)
      }
      if (rst === void 0) {
        return Promise.reject(rspData)
      } else if (rst === true) {
        console.warn('request promise 不建议返回 true')
        return Promise.resolve()
      } else {
        return Promise.reject(rst)
      }
    }
  }
  return rspData
}
