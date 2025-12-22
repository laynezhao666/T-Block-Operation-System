import Vue from 'vue'
import singleSpaVue from 'single-spa-vue'

// import * as url from '@/module/fopage/config/urls'
// import * as cgi from '@/module/tompage/config/cgi'

// Vue.use(TNWebUI)

// let vueLifecycles

export default (entryComponent, options = {}) => {
  const vueLifecycles = singleSpaVue({
    Vue,
    // Tnwebui,
    appOptions: { // eslint-disable-line no-new
      el: '#app',
      ...options,
      render: h => h(entryComponent),
    },
  })

  bootstrap.push(vueLifecycles.bootstrap)
  mount.push(vueLifecycles.mount)
  unmount.push(vueLifecycles.unmount)
}

export const bootstrap = [
  // vueLifecycles.bootstrap,
]

export const mount = [
  // vueLifecycles.mount,
]

export const unmount = [
  // vueLifecycles.unmount,
]
