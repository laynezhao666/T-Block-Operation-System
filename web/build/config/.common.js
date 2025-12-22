'use strict';
const path = require('path');

const LOCAL_HOST = 'local.dcim.dev.gmcc.net';
const DEV_HOST = 'dcim.dev.gmcc.net';

module.exports = {
  common: {
    // 1.true，代表使用modulname为前缀
    // 2.false 或 ''，代表从/项目访问，完全前后端分离项目
    // 3.'xxxxx'，自定义项目前缀
    // urlProjPrefix: true,
    // entryJS: ['common/script/console_text.js'],
    // entryJS: [],


    // AutoMapLocalHost: false, // 是否自动映射host域名到本地hosts文件，默认开启
    define: {
      LOGIN: JSON.stringify({
        local: {
          LOGIN_ORIGIN: 'http://logindcim.dev.gmcc.net',
        },
        dev: {
          LOGIN_ORIGIN: 'http://logindcim.dev.gmcc.net',
        },
        test: {
          LOGIN_ORIGIN: 'http://logindcim.test.gmcc.net',
        },
        publish: {
          LOGIN_ORIGIN: 'http://lndcim.gmcc.net',
        },
      }),
      // CGI地址
      HOST: JSON.stringify({
        local: '',
        dev: DEV_HOST,
        test: 'dcim.test.gmcc.net',
        publish: 'dcim.gmcc.net',
      }),
      // 主站地址
      HOME: JSON.stringify({
        local: DEV_HOST,
        dev: DEV_HOST,
        test: 'dcim.test.gmcc.net',
        publish: 'dcim.gmcc.net',
      }),
      // 网管站点
      NWHOST: JSON.stringify({
        local: 'netdcim.dev.gmcc.net',
        dev: 'netdcim.dev.gmcc.net',
        test: 'netdcim.test.gmcc.net',
        publish: 'nwdcim.gmcc.net',
      }),
    },
  },
  dev: {
  },
  build: {
      // isCleanLastBundlefile: true, // 是否清理上次打包的js，css文件
      // // showWebpackConfigMsg: false,
      // bundleDir: 'js-css',
      // outPutJsFilename: '[name].[chunkhash].js',
      // outPutCssFilename: '[name].[contenthash].css',
      // outPutChunkFilename: '[name].[chunkhash].js',
      // index: 'index.html',
      // assetsSubDirectory: 'static',
      // assetsPublicPath: '/',
      // assetsRoot: path.resolve(__dirname, '../dist'),
      // productionSourceMap: false,
      // devtool: '#source-map', // cheap-module-source-map
      // productionGzip: false,
      // externals: {
      //   vue: 'Vue',
      //   'vue-i18n': 'VueI18n',
      //   '@tencent/TNWeb-ui': 'ELEMENT',
      //   echarts: 'echarts',
      //   'element-ui': 'element-ui',
      //   $: 'jQuery',
      //   'single-spa': 'single-spa',
      // },
      // libraryTarget: 'var',
      // // 打包分析
      // bundleAnalyzerReport: process.env.npm_config_report,
  },
};
