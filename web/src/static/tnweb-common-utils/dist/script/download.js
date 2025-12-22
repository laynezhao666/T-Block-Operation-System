// 链接：https://www.zhihu.com/question/263323250/answer/267842980

import axiosDownload from "axios";

const downloadPost = (config) => {
    const url = config.url
    const data = JSON.parse(config.data)
    const form = document.createElement('form')
    form.action = url
    form.method = 'post'
    form.style.display = 'none'
    Object.keys(data).forEach(key => {
        const input = document.createElement('input')
        input.name = key
        input.value = data[key]
        form.appendChild(input)
    })
    const button = document.createElement('input')
    button.type = 'submit'
    form.appendChild(button)
    document.body.appendChild(form)
    form.submit()
    document.body.removeChild(form)
}

const downloadGet = (config) => {
    const params = []
    for (const item in config.params) {
        params.push(`${item}=${config.params[item]}`)
    }
    const url = params.length ? `${config.url}?${params.join('&')}` : `${config.url}`
    let iframe = document.createElement('iframe')
    iframe.style.display = 'none'
    iframe.src = url
    iframe.onload = function () {
        document.body.removeChild(iframe)
    }
    document.body.appendChild(iframe)
}

axiosDownload.interceptors.response.use(res => {
    // 处理流
    if (res.headers && res.headers['content-type'] === 'application/octet-stream') {
        const config = res.config
        if (config.method === 'post') {
            downloadPost(config)
        } else if (config.method === 'get') {
            downloadGet(config)
        }
        return
    }
}, error => {
    // Do something with response error
    return Promise.reject(error.response.data || error.message)
})

export default axiosDownload
