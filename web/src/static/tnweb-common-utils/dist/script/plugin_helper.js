import objectAssign from 'object-assign'

const mergeOptions = function ($vm, options) {
  const defaults = {}
  for (let i in $vm.$options.props) {
    if (i !== 'value') {
      defaults[i] = $vm.$options.props[i].default
    }
  }
  const _options = objectAssign({}, defaults, options)
  for (let i in _options) {
    $vm[i] = _options[i]
  }
}

const rtxList = function ($vm) {
  return new Promise((resolve, reject) => {
    if (window._arrusers) {
      return resolve(window._arrusers)
    }

    let script = document.createElement('script')
    script.onload = () => {
      resolve(window._arrusers)
    }

    script.onerror = () => {
      $vm.$alert('加载名单失败')
    }
    script.async = true

    let t = new Date()

    script.src = '//' + `${t.getFullYear()}${t.getMonth()}${t.getDate()}`
    document.head.appendChild(script)
  })
}

export {
  mergeOptions,
  rtxList,
}
