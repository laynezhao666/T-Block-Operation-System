/* 页面引导程序 */
import Vue from 'vue'
import VueI18n from 'vue-i18n'
import { isFunction } from 'lodash'
import 'common/style/theme.scss'
import Tnwebui from '@tencent/TNWeb-ui'
import http from './http2'

// 全局过滤
Vue.filter('filterEmpty', function (value) {
  return (value || value === 0) ? value : '--'
})

export default function init (opts = {}) {
  const { locales, lang = 'en', overwriteUI = false, customHttp = false, httpKey = '$axios', plugins = [] } = opts
  if (locales) {
    Vue.use(VueI18n)
    Object.keys(locales).forEach(function (lang) {
      Vue.locale(lang, locales[lang])
    })
    Vue.config.lang = lang
    if (overwriteUI) {
      Vue.use(Tnwebui, { locale: locales[lang] })
    } else {
      Vue.use(Tnwebui)
    }
  } else {
    Vue.use(Tnwebui)
  }
  plugins.forEach(plugin => {
    if (isFunction(plugin) || plugin.install) {
      Vue.use(plugin)
    } else {
      Vue.use(plugin.plugin, plugin.args)
    }
  })
  if (!customHttp) {
    Vue.prototype[httpKey] = http
  }
}
