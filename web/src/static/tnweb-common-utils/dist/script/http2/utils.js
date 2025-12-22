import qs from 'qs'
import { merge, isNil, isPlainObject, isFunction } from 'lodash'
import { Loading, Message } from '@tencent/TNWeb-ui'
import { defaultOpts, defaultConfig } from './config'

export class HTTPError extends Error {
  name = 'HTTPError'
}

export function loading ({ text, spinner }) {
  return Loading.service({
    fullscreen: true,
    lock: true,
    text,
    spinner,
  })
}

export function toast (message) {
  Message({
    type: 'error',
    message,
  })
}

export function createFile (fileName, blob) {
  const urlBlob = window.URL.createObjectURL(new Blob([blob]))
  const link = document.createElement('a')
  link.style.display = 'none'
  link.href = urlBlob
  link.setAttribute('download', fileName)
  document.body.appendChild(link)
  const fn = e => {
    e.stopPropagation()
  }
  link.addEventListener('click', fn)
  link.click()
  link.removeEventListener('click', fn)
  document.body.removeChild(link)
}

export function getOpts (opts) {
  const { isJson, config } = opts
  const extraConfig = { headers: { post: { } } }
  if (isJson) {
    extraConfig.headers.post['Content-Type'] = 'application/json'
  } else {
    extraConfig.headers.post['Content-Type'] = 'application/x-www-form-urlencoded; charset=UTF-8'
  }
  return merge({}, defaultConfig, {
    ...opts,
    config: merge({}, config, extraConfig),
  })
}

function generator (cb, { key, descriptor }) {
  const fn = descriptor.value
  descriptor.value = function (url, data, loadOpts, opts) {
    opts = merge({}, defaultOpts[key], opts)
    data = merge({}, this.extraParams, data)
    let promise = fn.call(this, url, data, loadOpts, opts)
    const chain = this.filters[key]
    const env = {
      opts,
      data,
      context: this,
    }
    for (let i = 0; i < chain.length; i++) {
      promise = promise.then(chain[i].bind(env))
    }
    return promise.then(cb?.bind(env))
  }
}

/**
 * 装饰器
 * 注入 extraParams、默认的 opts 和 filter
 */
export function inject (arg) {
  if (isFunction(arg)) {
    return generator.bind(null, arg)
  } else {
    return generator(void 0, arg)
  }
}

export function jsonify ({ key, descriptor }) {
  const fn = descriptor.value
  descriptor.value = function (url, data, loadOpts, opts) {
    let isPost
    if (key === 'post') {
      isPost = true
    } else {
      isPost = opts.method === 'post'
    }
    if (isPost) {
      const isJson = isNil(opts.isJson) ? this.opts.isJson : opts.isJson
      if (isJson) {
        opts.restAxios = merge({
          headers: {
            post: {
              'Content-Type': 'application/json',
            },
          },
        }, opts.restAxios)
      } else {
        opts.restAxios = merge({
          headers: {
            post: {
              'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
            },
          },
        }, opts.restAxios)
      }
      opts.isJson = isJson
      data = isJson ? data : qs.stringify(data)
    }
    return fn.call(this, url, data, loadOpts, opts)
  }
}

export function load ({ key, descriptor }) {
  const fn = descriptor.value
  descriptor.value = function (url, data, loadOpts = true, opts) {
    let loadingInstance
    if (loadOpts) {
      let text
      if (isPlainObject(loadOpts)) {
        text = loadOpts.text
      } else {
        text = '加载中'
      }
      if (key !== 'download') {
        this.counter++
      }
      loadingInstance = loading({ text, spinner: this.spinner })
    }
    if (key === 'download') {
      return fn.call(this, url, data, loadOpts, opts).finally(() => loadingInstance?.close?.())
    } else {
      return fn.call(this, url, data, loadOpts, opts).finally(() => {
        if (loadingInstance && loadOpts && --this.counter === 0) {
          loadingInstance.close()
        }
      })
    }
  }
}
