/* 页面引导程序 */
import Vue from 'vue'
import axios from './http'
// import 'babel-polyfill' // 兼容IE

import '@tencent/TNWeb-ui/lib/theme-chalk/reset.css'
import '@tencent/TNWeb-ui/lib/theme-chalk/index.css'

import Tnwebui from '@tencent/TNWeb-ui'
Vue.use(Tnwebui)

// 全局引入ajax
Vue.prototype.$axios = axios

// 全局过滤
Vue.filter('filterEmpty', function (value) {
  return (value && value !== '') ? value : '--'
})

// export default Vue
