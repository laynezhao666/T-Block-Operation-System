<!--该文件用于快速按fusion规范迁移旧php系统页面-->
<template>
  <!--没有直出的data-->
  <div>
  <!--div v-if="data"--><!--有直出的data-->

  <!-- php html放这里 -->
  </div>
</template>

<script>
/* eslint-disable */

// 用cgi拉取直出数据时需要用到
// import axios from 'axios'

// 旧系统上第三方库迁移的根路径
// import '../oldstatic/'

// 如果有读取url参数发送首屏cgi请求时，需要用到
// import { getQueryString } from 'common/script/utils.js'

// 输出方法/变量/mixin/components的库
// import {xxx} from './xxx.js'

// 引入js的字符串类型图片地址
// import img from 'common/image/xxx.png'

export default {
  // 如果有组件方式引入的库
  // components: {
  //   componentsName
  // },
  // 如果有以mixin方式引入的库
  // mixins: [
  //   mixinName
  // ]
  props: {
    // main.js传入
    url: {
      type: Object,
      required: true,
    },
    // main.js传入
    cgi: {
      type: Object,
      required: true,
    },
    // 其他可能用到的main.js传入内容
    // getCurrentModuleName: {
    //   type: function,
    //   default() {
    //     return ''
    //   }
    // }
  },
  data() {
    return {
      // 首屏cgi需要读取url参数放这里
      // params: getQueryStrings(),

      // 首屏cgi请求的数据放这里
      data: null
    }
  },
  mounted () {
      // 有首屏cgi的时候
//    this.getData().then(({ data: { code, data } }) => {
//      if (code) {
//        FO.mask('', 'close')
//        FO.showErr('出错了：' + code)
//      } else {
//        this.data = data
//        // 保证vue渲染延迟和jq ready
//        this.$nextTick(() => {
//          $(() => {
//            this.init()
//          })
//        })
//      }
//      console.log(data)
//    })

    // 没有首屏cgi的时候
      $(() => {
        this.init()
      })
   
  },
  methods: {
    getData() {
      // return axios.get(this.cgi.xxx, {
      //    params: this.params
      // })
    },
    init () {
      // let that = this // 用于cgi和url的读取
        
      // php的js代码放这里

      // 页面的js文件里的代码放这里
    },
    // 如果有其他业务文件可以放这里
    // anyname () {
    // }
  }
}
</script>

<style>

/* 引入第三方库的css */

/* @import '../oldstatic/xxx.css'; */

/* 引入页面的css */

/* @import '.css'; */

/* 引入内联的css */

</style>