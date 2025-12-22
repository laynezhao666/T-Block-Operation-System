// import qs from 'qs'
import axios from 'axios'
// import { isNil } from 'lodash'
import codeFilter from './filters/codeFilter'
import stripFilter from './filters/stripFilter'
import blobFilter from './filters/blobFilter'
import { createFile, inject, load, getOpts, jsonify, HTTPError } from './utils'
import { defaultOpts } from './config'

function download (rsp) {
  let fileName = this.opts.fileName
  const url = rsp.config.url
  /**
   * 如果是直接下载，默认获取 content-disposition 指定的 filename，如果不存在则截取 url的最后一段
   * 比如 "/data/exceltemplate/demand/visitors_import_all.xlsx" 截取为 visitors_import_all.xlsx
   */
  if (!fileName) {
    const header = rsp.headers['content-disposition']
    if (header) {
      const matches = header.match(/filename\*?="?(?:.*'')?(.*?)"?$/)
      fileName = decodeURIComponent(matches[1])
    } else {
      const xlsName = url.split('/').pop()
      if (xlsName.endsWith('.xlsx')) {
        fileName = xlsName
      } else {
        fileName = `${xlsName}.xlsx`
      }
    }
  }
  const { data } = rsp
  createFile(fileName, data)
}

class HTTP {
  /**
   *Creates an instance of HTTP.
   * @param {hasCode = true, needStrip = true, extraParams = {}, config = {}, spinner, isJson, showToast} opts extraParams 是每个请求会附带的参数，如 tnebula 中的模组名，config 是 axios 定义的配置项
   * @param {post = {}, pre = {}} [filters={}] post 中的 filter 会在自带的 filter 后调用，pre 中的 filter 会在自带的 filter 前调用
   */
  constructor (opts = {}, filters = {}, interceptors = []) {
    opts = getOpts(opts)
    const { post = {}, pre = {} } = filters
    let { hasCode, needStrip, extraParams, config, spinner, isJson, codeHandler, showToast = true } = opts
    Object.values(defaultOpts).forEach(opt => {
      opt.showToast = showToast
    })
    if (codeHandler) {
      hasCode = true
    }
    if (!needStrip && hasCode) {
      throw new HTTPError('当 hasCode 为 true 时， needStrip 必须为 true')
    }
    this.filters = {
      post: [
        ...pre.post || [],
        needStrip && stripFilter,
        hasCode && codeFilter,
        ...post.post || []
      ].filter(Boolean),
      get: [
        ...pre.post || [],
        needStrip && stripFilter,
        hasCode && codeFilter,
        ...post.get || []
      ].filter(Boolean),
      download: [
        ...pre.post || [],
        blobFilter,
        hasCode && codeFilter,
        ...post.download || []
      ].filter(Boolean),
      delete: [
        ...pre.post || [],
        needStrip && stripFilter,
        hasCode && codeFilter,
        ...post.delete || []
      ].filter(Boolean)
    }
    this.extraParams = extraParams
    this.opts.spinner = spinner
    this.opts.isJson = isJson
    this.opts.codeHandler = codeHandler
    this.ins = axios.create(config)
    interceptors.forEach(interceptor => {
      this.ins.interceptors.request.use(interceptor.bind(this))
    })
  }

  counter = 0
  opts = {}

  @inject
  @jsonify
  @load
  post (url, data, loadOpts, opts) {
    return this.ins({
      method: 'post',
      url,
      data,
      ...opts.restAxios
    })
  }

  @inject
  @load
  get (url, params, loadOpts, opts) {
    return this.ins({
      method: 'get',
      url,
      params,
      ...opts.restAxios
    })
  }

  @inject(download)
  @jsonify
  @load
  download (url, params, loadOpts, opts) {
    const { method } = opts
    const config = {
      method,
      url,
      responseType: 'blob',
      ...opts.restAxios
    }

    if (method === 'post') {
      config.data = params
    } else {
      config.params = params
    }
    return this.ins(config)
  }

  @inject
  @load
  delete (url, params, loadOpts, opts) {
    return this.ins({
      method: 'delete',
      url,
      params,
      ...opts.restAxios
    })
  }
}

export function create (...args) {
  return new HTTP(...args)
}

export default new HTTP()
