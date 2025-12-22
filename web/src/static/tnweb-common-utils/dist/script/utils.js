/**
 * 判断是否是手机端
 * @export
 * @returns
 */
export function isMobile () {
  const sUserAgent = navigator.userAgent.toLowerCase()

  const bIsIpad = sUserAgent.match(/ipad/i) === 'ipad'

  const bIsIphoneOs = sUserAgent.match(/iphone os/i) === 'iphone os'

  const bIsMidp = sUserAgent.match(/midp/i) === 'midp'

  const bIsUc7 = sUserAgent.match(/rv:1.2.3.4/i) === 'rv:1.2.3.4'

  const bIsUc = sUserAgent.match(/ucweb/i) === 'ucweb'

  const bIsAndroid = sUserAgent.match(/android/i) === 'android'

  const bIsCE = sUserAgent.match(/windows ce/i) === 'windows ce'

  const bIsWM = sUserAgent.match(/windows mobile/i) === 'windows mobile'

  const bIsWebview = sUserAgent.match(/webview/i) === 'webview'
  return (
    bIsIpad ||
    bIsIphoneOs ||
    bIsMidp ||
    bIsUc7 ||
    bIsUc ||
    bIsAndroid ||
    bIsCE ||
    bIsWM ||
    bIsWebview
  )
}

export function getQueryStrings (url) {
  const ps = {}
  const u = url || window.location.href
  if (u.indexOf('?') !== -1) {
    const s = u.split('?')[1]
    const strs = s.split('&')
    for (let i = 0; i < strs.length; i++) {
      const p = strs[i].split('=')
      ps[p[0]] = decodeURIComponent(p[1])
    }
  }
  return ps
}

/**
 * 获取url参数值
 * @export
 * @param {any} name
 * @returns
 */
export function getQueryString (name) {
  const reg = new RegExp(`(^|&)${name}=([^&]*)(&|$)`, 'i')
  let r
  try {
    r = window.location.href.split('?')[1].match(reg) // window.location.search.substr(1).match(reg);
  } catch (err) {
    r = null
  }
  if (r != null) {
    return decodeURIComponent(r[2])
  }
  return null
}

/**
 * 序列号对象
 * @param {object} obj
 * 注意：支持不完善，有bug，不建议使用，请import qs代替
 */
export function queryObject (obj) {
  return Object.keys(obj).map((key) => {
    return `${key}=${encodeURI(obj[key])}`
  }).join('&')
}

export function escapeHtml (string) {
  const matchHtmlRegExp = /["'&<>]/
  const str = `${string}`
  const match = matchHtmlRegExp.exec(str)

  if (!match) {
    return str
  }

  let escape
  let html = ''
  let index = 0
  let lastIndex = 0

  for (index = match.index; index < str.length; index++) {
    switch (str.charCodeAt(index)) {
      case 34: // "
        escape = '&quot;'
        break
      case 38: // &
        escape = '&amp;'
        break
      case 39: // '
        escape = '&#39;'
        break
      case 60: // <
        escape = '&lt;'
        break
      case 62: // >
        escape = '&gt;'
        break
      default:
        continue
    }

    if (lastIndex !== index) {
      html += str.substring(lastIndex, index)
    }

    lastIndex = index + 1
    html += escape
  }

  return lastIndex !== index ? html + str.substring(lastIndex, index) : html
}

export function cutStr (str, len) {
  if (typeof str === 'number') {
    str = str.toString()
  }
  if (str.toString.length * 2 <= len) {
    return str
  }
  var strlen = 0
  var s = ''
  for (var i = 0; i < str.length; i++) {
    s = s + str.charAt(i)
    if (str.charCodeAt(i) > 128) {
      strlen = strlen + 2
      if (strlen >= len) {
        return s.substring(0, s.length - 1) + '..'
      }
    } else {
      strlen = strlen + 1
      if (strlen >= len) {
        return s.substring(0, s.length - 2) + '..'
      }
    }
  }
  return s
}

/**
 * Get the first item that pass the test
 * by second argument function
 *
 * @param {Array} list
 * @param {Function} f
 * @return {*}
 */
export function find (list, f) {
  return list.filter(f)[0]
}

/**
 * Deep copy the given object considering circular structure.
 * This function caches all nested objects and its copies.
 * If it detects circular structure, use cached copy to avoid infinite loop.
 *
 * @param {*} obj
 * @param {Array<Object>} cache
 * @return {*}
 */
export function deepCopy (obj, cache = []) {
  // just return if obj is immutable value
  if (obj === null || typeof obj !== 'object') {
    return obj
  }

  // if obj is hit, it is in circular structure
  const hit = find(cache, c => c.original === obj)
  if (hit) {
    return hit.copy
  }

  const copy = Array.isArray(obj) ? [] : {}
  // put the copy into cache at first
  // because we want to refer it in recursive deepCopy
  cache.push({
    original: obj,
    copy,
  })

  Object.keys(obj).forEach(key => {
    copy[key] = deepCopy(obj[key], cache)
  })

  return copy
}

/**
 * forEach for object
 */
export function forEachValue (obj, fn) {
  Object.keys(obj).forEach(key => fn(obj[key], key))
}

export function isObject (obj) {
  return obj !== null && typeof obj === 'object'
}

export function isPromise (val) {
  return val && typeof val.then === 'function'
}

export function assert (condition, msg) {
  if (!condition) throw new Error(`[vuex] ${msg}`)
}

export const localStorageJson = {
  get: function (key) {
    if (typeof (Storage) !== 'undefined') {
      try {
        let value = localStorage.getItem(key)
        return value ? JSON.parse(value) : []
      } catch (oException) {
        console.log(oException)
      }
    } else {
      console.log('不支持localStorage！')
    }
  },
  set: function (key, value) {
    if (typeof (Storage) !== 'undefined') {
      try {
        value = JSON.stringify(value)
        localStorage.setItem(key, value)
      } catch (oException) {
        if (oException.name === 'QuotaExceededError') {
          console.log('已经超出本地存储限定大小！')
          // // 可进行超出限定大小之后的操作，如下面可以先清除记录，再次保存
          // localStorage.clear();
          // localStorage.setItem(key,value);
        }
      }
    } else {
      console.log('不支持localStorage！')
    }
  },
  push: function (key, value) {
    if (typeof (Storage) !== 'undefined') {
      try {
        let oldValue = localStorage.getItem(key)
        oldValue = JSON.parse(oldValue)
        if (!oldValue) {
          oldValue = []
        }
        oldValue.push(value)
        let newValue = JSON.stringify(oldValue)
        localStorage.setItem(key, newValue)
      } catch (oException) {
        if (oException.name === 'QuotaExceededError') {
          console.log('已经超出本地存储限定大小！')
          // // 可进行超出限定大小之后的操作，如下面可以先清除记录，再次保存
          // localStorage.clear();
          // localStorage.setItem(key,value);
        }
      }
    } else {
      console.log('不支持localStorage！')
    }
  },

}

export function doubleNum (n) {
  return n > 9 ? n : ('0' + n)
}

/**
 * 将字符串以指定分隔符，转换为驼峰
 * @param {string} str 需要转换的字符串
 * @param {string} separator 分隔符，例如' ','-','_'
 */
export function toCameCase (str, separator) {
  if (typeof str !== 'string') {
    return str
  }

  var arr = str.split(separator)

  for (var i = 1, len = arr.length; i < len; i++) {
    arr[i] = arr[i].charAt(0).toUpperCase() + arr[i].substring(1)
  }

  return arr.join('')
}

/**
 * @param {object|array} o 对象或数组
 * @param {string} separator 分隔符
 */
export function camelCaseObjectKeys (o, separator) {
  if (!o) {
    console.log('objectToCameCaseKey:', '参数错误', o, separator)
    return o
  }

  if (o instanceof Array) {
    return o.map((item) => {
      let newItem = {}
      Object.keys(item).forEach((key) => {
        newItem[toCameCase(key, '_')] = item[key]
      })
      return newItem
    })
  } else if (Object.keys(o).length) {
    let newItem = {}
    Object.keys(o).forEach((key) => {
      newItem[toCameCase(key, '_')] = o[key]
    })

    return newItem
  }
}
/**
 * 转换时间格式，暂时更改。后期可优化或改为修改date对象原型方法
 * @param {string} v 时间字符串，时间戳等
 */
export function formateDate (v) {
  const date = new Date(v)
  let _year = date.getFullYear()
  let _month = date.getMonth() + 1 // 月从0开始计数
  let _d = date.getDate()
  let _hour = date.getHours()
  let _min = date.getMinutes()
  let _sec = date.getSeconds()

  _month = (_month > 9) ? ('' + _month) : ('0' + _month)
  _d = (_d > 9) ? ('' + _d) : ('0' + _d)
  _hour = (_hour > 9) ? ('' + _hour) : ('0' + _hour)
  _min = (_min > 9) ? ('' + _min) : ('0' + _min)
  _sec = (_sec > 9) ? ('' + _sec) : ('0' + _sec)

  return _year + '-' + _month + '-' + _d + ' ' + _hour + ':' + _min + ':' + _sec
}
