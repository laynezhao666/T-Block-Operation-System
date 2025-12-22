/**
 * 常用ui提示
 */
import Vue from 'vue'

let unloggedTipOpened = false

export function unloggedTip () {
  if (unloggedTipOpened) return Promise.reject(new Error('opened'))
  unloggedTipOpened = true

  return Vue.prototype.$alert('您还没有登录，请点击确定打开新窗口登录', {
    showCancelButton: true,
    cancelButtonText: '取消',
    confirmButtonText: '确定',
  }).finally(() => {
    unloggedTipOpened = false
  })
}
