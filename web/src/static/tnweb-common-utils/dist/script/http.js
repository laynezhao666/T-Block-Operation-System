import qs from 'qs'
import axios from 'axios'
import { isPlainObject, isNil } from 'lodash'
import { Loading, Message } from '@tencent/TNWeb-ui'

axios.defaults.timeout = 30000
axios.defaults.headers.common['X-Requested-With'] = 'XMLHttpRequest'
axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded; charset=UTF-8'
axios.interceptors.response.use(rsp => {
  // rsp -> wrapper(rsp.data)
  if (isPlainObject(rsp.data)) {
    return rsp.data || {}
  } else if (rsp.data instanceof Blob) {
    return {
      blob: rsp.data,
      headers: rsp.headers,
    }
  }
}, e => Promise.reject(e))
axios.interceptors.response.use(rspData => {
  // wrapper(rsp.data) -> bizData
  if (rspData && !rspData.blob) {
    const { code, data, message } = rspData
    if (code === 0) {
      return data
    } else if (isNil(code) && isNil(data) && isNil(message)) {
      // 兼容老接口
      return rspData
    } else {
      let msg = message || '系统错误'
      toast(msg)
      return Promise.reject(rspData)
    }
  }
  return rspData
})

function loading (text) {
  return Loading.service({
    fullscreen: true,
    lock: true,
    text,
    // target,
    // background: 'rgba(0, 0, 0, 0.1)',
    // spinner: 'nbl-loading',
  })
}

function toast (message) {
  Message({
    type: 'error',
    message,
  })
}

function createFile (fileName, blob) {
  let urlBlob = window.URL.createObjectURL(new Blob([blob]))
  let link = document.createElement('a')
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

let postCounter = 0
let getCounter = 0

const http = {
  post (url, data = {}, loadOpts = true, restAxios, seria = true) {
    let loadingInstance
    if (loadOpts) {
      let text
      if (isPlainObject(loadOpts)) {
        text = loadOpts.text
      } else {
        text = '加载中'
      }
      loadingInstance = loading(text)
      postCounter++
    }
    return axios({
      method: 'post',
      url,
      data: seria ? qs.stringify(data) : data,
      ...restAxios,
    }).finally(() => {
      if (loadingInstance && --postCounter === 0) {
        loadingInstance.close()
      }
    })
  },

  get (url, params = {}, loadOpts = true, restAxios) {
    let loadingInstance
    if (loadOpts) {
      let text
      if (isPlainObject(loadOpts)) {
        text = loadOpts.text
      } else {
        text = '加载中'
      }
      loadingInstance = loading(text)
      getCounter++
    }
    return axios({
      method: 'get',
      url,
      params,
      ...restAxios,
    }).finally(() => {
      if (loadingInstance && --getCounter === 0) {
        loadingInstance.close()
      }
    })
  },

  download (url, params = {}, method = 'get', fileName, loadOpts = true, isJson) {
    let loadingInstance
    if (loadOpts) {
      let text
      if (isPlainObject(loadOpts)) {
        text = loadOpts.text
      } else {
        text = '下载中'
      }
      loadingInstance = loading(text)
    }
    let config = {
      method,
      url,
      responseType: 'blob',
    }

    if (method === 'post') {
      if (isJson) {
        config.headers = { 'content-type': 'application/json' }
        config.data = params
      } else {
        config.data = qs.stringify(params)
      }
    } else {
      config.params = params
    }
    return axios(config).then(bizData => {
      /**
       * 如果是直接下载，默认获取 content-disposition 指定的 filename，如果不存在则截取 url 的最后一段
       * 比如 "/data/exceltemplate/demand/visitors_import_all.xlsx" 截取为 visitors_import_all.xlsx
       */
      if (!fileName) {
        const header = bizData.headers['content-disposition']
        const matches = header.match(/filename\*?="?(?:.*'')?(.*?)"?$/)
        if (matches[1]) {
          fileName = decodeURIComponent(matches[1])
        } else {
          let xlsName = url.split('/').pop()
          if (xlsName.endsWith('.xlsx')) {
            fileName = xlsName
          } else {
            fileName = `${xlsName}.xlsx`
          }
        }
      }
      const { blob } = bizData
      createFile(fileName, blob)
    }).finally(() => loadingInstance?.close?.())
  },
  stringify (params = {}) {
    return qs.stringify(params)
  },
  parse (params = '') {
    return qs.parse(params)
  },
}

export default http
