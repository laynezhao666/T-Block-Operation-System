/**
 * passport公用登录态相关处理函数
 */
import parse from './cookie'
import { create } from './http2'
/**
 * local:本地开发
 * dev:开发自测开发
 * test:公用测试环境
 * pre:公用测试环境
 * publish：正式环境
 */

// const {
//   DEFAULT_REDIRECT_ORIGIN,
//   DEFAULT_REDIRECT_ORIGIN_PRE,
//   DEFAULT_REDIRECT_ORIGIN_TEST,

//   LOGIN_ORIGIN,
//   LOGIN_ORIGIN_PRE,
//   LOGIN_ORIGIN_TEST,

//   QQ_APPID_PUBLISH,
//   QQ_APPID_PRE,
//   QQ_APPID_TEST,

//   WX_APPID_PUBLISH,
//   WX_APPID_PRE,
//   WX_APPID_TEST,
// } = LOGIN

const _ENV = ((port, origin) => {
  if (port) {
    return 'local'
  } else if (/dev/.test(origin)) {
    // 匹配任意含dev的域名
    return 'dev'
  } else if (/test/.test(origin)) {
    // 匹配任意含test的域名
    return 'test'
  } else if (/\Wpre\w/.test(origin) || /\wpre\./.test(origin)) {
    // 匹配*://prexxx.xxx或者xxxpre.xxx
    return 'pre'
  } else {
    return 'publish'
  }
})(window.location.port, window.location.origin)

const LOGIN_CONFIG = LOGIN[_ENV]

// const PASSPORT_ORIGIN = {
//   // 只有登录调试自身才会用到这个，其他时候都用测试环境的
//   dev: LOGIN_ORIGIN_TEST,
//   test: LOGIN_ORIGIN_TEST,
//   pre: LOGIN_ORIGIN_PRE,
//   publish: LOGIN_ORIGIN,
// }[_ENV]
const PASSPORT_ORIGIN = LOGIN_CONFIG.LOGIN_ORIGIN
// const PASSPORT_URL = `${PASSPORT_ORIGIN}/`

// const DEFAULT_REDIRECT = {
//   dev: DEFAULT_REDIRECT_ORIGIN_TEST,
//   test: DEFAULT_REDIRECT_ORIGIN_TEST,
//   pre: DEFAULT_REDIRECT_ORIGIN_PRE,
//   publish: DEFAULT_REDIRECT_ORIGIN,
// }[_ENV]

const DEFAULT_REDIRECT = LOGIN_CONFIG.DEFAULT_REDIRECT_ORIGIN

// const PASSPORT_URL = {
//   dev: `${PASSPORT_ORIGIN}/`,
//   test: `${PASSPORT_ORIGIN}/`,
//   pre: `${PASSPORT_ORIGIN}/`,
//   publish: `${PASSPORT_ORIGIN}/`,
// }[_ENV]

export const REDIRECT_KEY = 'r_uri'

export const ENV = {
  [_ENV]: true
}

export const ENV_NAME = _ENV

// export const WX_APPID = {
//   dev: WX_APPID_TEST,
//   test: WX_APPID_TEST,
//   pre: WX_APPID_PRE,
//   publish: WX_APPID_PUBLISH,
// }[_ENV]

export const WX_APPID = LOGIN_CONFIG.WX_APPID
export const MOBILE_WX_APPID = LOGIN_CONFIG.MOBILE_WX_APPID
export const QQ_APPID = LOGIN_CONFIG.QQ_APPID

// export const QQ_APPID = {
//   dev: QQ_APPID_TEST,
//   test: QQ_APPID_TEST,
//   pre: QQ_APPID_PRE,
//   publish: QQ_APPID_PUBLISH,
// }[_ENV]

// 默认跳转url，写成tnboss系统
export function getDefaultRedirectUrl () {
  return DEFAULT_REDIRECT
}

// 协议+域名+port
export function getBaseOrigin () {
  if (ENV.dev && window.location.host.indexOf('login') > -1) {
    // 方便本地调试
    return window.location.origin
  } else {
    return PASSPORT_ORIGIN
  }
}

// passport主页面，或其根地址
export function getBaseUrl () {
  // return PASSPORT_URL
  return PASSPORT_ORIGIN + '/'
}

// 获取登录信息(无法保证一定有效，可能是过期的)
export function getLoginInfo () {
  const cookie = parse(document.cookie)
  return {
    // 四个cookie都已经设置了相同有效期
    // 有昵称则可说明cookie未过期
    status: !!cookie.tnebula_token,
    name: decodeURIComponent(cookie.tnebula_username || '')
  }
}

// 跳转到passport登录页方法
export function redirectToLogin ({ newTarget } = {}) {
  const url = `${PASSPORT_ORIGIN}/login.html?${REDIRECT_KEY}=${encodeURIComponent(window.location.href)}`
  if (newTarget) {
    // 防止被拦截
    const win = window.open('', '_blank')
    win.location.href = url
  } else {
    window.location.replace(url)
  }
}

// 退出
export function logout () {
  window.location.replace(`${PASSPORT_ORIGIN}/login.html?logout=1&${REDIRECT_KEY}=${encodeURIComponent(window.location.href)}`)
}

export const plugin = {
  install (Vue, {
    data,
    codeHandler,
    opts,
    httpRestOpts
  } = {}) {
    if (ENV.dev) {
      console.warn('vue plugin已升级为passport_plugin.js，请及时替换')
    }
    const that = this
    let init
    this.$storage = {
      ...this.$storage,
      ...data
    }
    Vue.prototype.$axios = create({
      ...opts,
      config: {
        withCredentials: true
      },
      codeHandler: (code, ...args) => {
        if (code === -99) {
          redirectToLogin()
        } else if (codeHandler) {
          return codeHandler.call(this, code, ...args)
        }
      },
      ...httpRestOpts
    })
    Vue.mixin({
      beforeCreate () {
        Vue.util.defineReactive(this, '$storage', that.$storage)
        if (!init) {
          if (!this.$storage.account || !this.$storage.account.status) {
            return redirectToLogin()
          }

          this.$storage.loginStatus = 'ok'
          init = true
        }
      },
      methods: {
        logout
      }
    })
  },
  $storage: {
    loginStatus: '',
    account: getLoginInfo()
  }
}
