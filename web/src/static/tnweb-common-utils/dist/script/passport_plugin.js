/**
 * vue插件
 * 常用业务配置抽离出来
 * 方便添加各种常用公用的业务逻辑
 */
import { create } from './http2'
import { unloggedTip } from './tips'
import {
  redirectToLogin,
  getLoginInfo,
  logout,
} from './passport_login'

/**
 * 默认的业务配置
 * opts: 用于http2的配置
 * data: 将放入this.$storage
 * codeHandler(code) context=vue，优先处理code逻辑，返回true表示终端后续处理
 */
export const defaultArgs = {
  opts: {
    config: {
      withCredentials: true,
    },
  },
  data: {
    pageStatus: 'ok',
  },
  codeHandler(code) {
    if (code === 0) {
      unloggedTip().then(() => {
        redirectToLogin({ newTarget: true })
      })
      return false
    } else if (code === 1400800) {
      this.$storage.pageStatus = 'no-permission'
      // 不再返回到业务
      return true
    }
  },
}

/**
 * 合并配置参数
 */
export function argsInit({ data, codeHandler, opts } = {}) {
  return {
    opts: {
      ...defaultArgs.opts,
      ...opts,
    },
    data: {
      ...defaultArgs.data,
      ...data,
    },
    codeHandler() {
      const ret = codeHandler && codeHandler.call(this, ...arguments)

      if (ret) return ret

      return defaultArgs.codeHandler.call(this, ...arguments)
    },
  }
}

/**
 * vue插件
 * 对参数处理
 */
const plugin = {
  install(Vue, {
    data,
    opts,
    codeHandler,
  } = {}) {
    const that = this
    let init

    this.$storage = {
      ...this.$storage,
      ...data,
    }
    Vue.prototype.$axios = create({
      codeHandler: codeHandler.bind(this),
      ...opts,
    })

    Vue.mixin({
      beforeCreate() {
        Vue.util.defineReactive(this, '$storage', that.$storage)
        if (!init) {
          // if (!this.$storage.account || !this.$storage.account.status) {
          //   return redirectToLogin()
          // }

          this.$storage.loginStatus = 'ok'
          init = true
        }
      },
      methods: {
        logout,
      },
    })
  },
  $storage: {
    loginStatus: '',
    account: getLoginInfo(),
  },
}

export function passportPlugin(opts) {
  return {
    plugin,
    args: argsInit(opts),
  }
}

export default passportPlugin
