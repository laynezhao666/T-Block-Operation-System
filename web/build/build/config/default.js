// 'use strict';
// Template version: 1.3.1
// see http://vuejs-templates.github.io/webpack for documentation.
const path = require('path');
function resolve(dir) {
  return path.join(global.TNF_projRootPath, dir);
};
module.exports = {
  dev: {
    // 1.true，代表使用modulname为前缀
    // 2.false 或 ''，代表从/项目访问，完全前后端分离项目
    // 3.'xxxxx'，自定义项目前缀
    urlProjPrefix: true,

    entryJS: [],
    AutoMapLocalHost: false, // 是否自动映射host域名到本地hosts文件，默认开启

    showWebpackConfigMsg: true,
    bundleDir: 'js-css',
    index: 'index.html', // 浏览器默认打开页面
    host: 'localhost',
    assetsSubDirectory: 'static',
    assetsPublicPath: '/',
    externals: {},
    proxyTable: {},
    define: {},
    assetsRoot: `${global.TNF_projRootPath}/dist`,
    devtool: '#cheap-module-eval-source-map', // 修改编译sourcemap模式，提高性能

    rewrites: [],
    isOpenPathRewrite: true,
    libraryTarget: 'var',
    libraryPre: '',
    port: 8080,
    autoOpenBrowser: true,
    errorOverlay: true,
    notifyOnErrors: true,
    poll: false,
    useEslint: true,
    showEslintErrorsInOverlay: false,
    cacheBusting: true,
    cssSourceMap: true,
    extensions: ['.js', '.vue', '.json', 'ts', 'tsx'],
    alias: {
      '@': resolve('src'),
      '@@': resolve(`src/module/${global.TNF_curModuleName}`),
      vue$: 'vue/dist/vue.common',

      // '@template': resolve('src/template'),
      // '@smpage_index': resolve('src/module/smpage/index'),
      // '@dcsm_index': resolve('src/module/dcsm/index'),
      // '@module': resolve(`src/feature/${global.TNF_curModuleName}`),
      // common: '@tencent/tnweb-common-utils/dist',
      // feature: resolve('src/feature'),
      // component: resolve('src/feature/component'),

      // '@common': resolve('src/common'),
      // '@component': resolve('src/common/component'),
      // '@image': resolve('src/common/image'),
      // '@script': resolve('src/common/script'),
      // '@style': resolve('src/common/style'),
    },
    filterEntryDir: [
      'static', 'config', 'mock', 'assets', 'components', 'component',
      'style', 'common', 'script', 'utils', 'router', 'store',
    ],
    rewriteWebpackConfigFn: null, // (config) => {}
    // 自定义拷贝目录
    // [
    //   {
    //     from: 'assets',   //from:会自动从module下模块路径下的文件开始查找，
    //     to: '.',          //to:是相对于assetRoot目录
    //     ignore: ['.*'],   //ignore:忽略的文件夹
    //   },
    // ],
    dirCopyMap: null,
  },

  build: {
    buildImportJson: true, // false 则不构建import.json
    removeAppjsonPrefix: '', // 自定义部署模块，如果配置，打包出来的微前端apps.json key会从移除removeAppjsonPrefix部分

    isMicrofrontMainNav: false, // 是否为微前端主导航
    urlProjPrefix: true,

    entryJS: [],

    isCleanLastBundlefile: true, // 是否清理上次打包的js，css文件
    showWebpackConfigMsg: false,

    bundleDir: 'js-css',

    // outPutJsFilename: '[name].js',
    // outPutCssFilename: '[name].css',
    outPutJsFilename: '[name].[chunkhash].js',
    outPutCssFilename: '[name].[contenthash].css',

    outPutChunkFilename: '[name].[chunkhash].js',
    index: 'index.html',
    assetsSubDirectory: 'static',
    assetsPublicPath: '/',
    assetsRoot: `${global.TNF_projRootPath}/dist`,
    productionSourceMap: false,
    devtool: '#source-map', // cheap-module-source-map
    productionGzip: false,
    externals: {},
    // 打包分析
    bundleAnalyzerReport: process.env.npm_config_report,
    libraryTarget: 'var',
    libraryPre: '',
    define: {},
    productionGzipExtensions: ['js', 'css'],
    extensions: ['.js', '.vue', '.json', 'ts', 'tsx'],
    alias: {
      '@': resolve('src'),
      '@@': resolve(`src/module/${global.TNF_curModuleName}`),
      vue$: 'vue/dist/vue.common',

      // '@template': resolve('src/template'),
      // '@smpage_index': resolve('src/module/smpage/index'),
      // '@dcsm_index': resolve('src/module/dcsm/index'),
      // '@module': resolve(`src/feature/${global.TNF_curModuleName}`),
      // common: '@tencent/tnweb-common-utils/dist',
      // feature: resolve('src/feature'),
      // component: resolve('src/feature/component'),

      // '@common': resolve('src/common'),
      // '@component': resolve('src/common/component'),
      // '@image': resolve('src/common/image'),
      // '@script': resolve('src/common/script'),
      // '@style': resolve('src/common/style'),
    },
    filterEntryDir: [
      'static', 'config', 'mock', 'assets', 'components', 'component',
      'style', 'common', 'script', 'utils', 'router', 'store',
    ],
    rewriteWebpackConfigFn: null, // (config) => {}
    // 自定义拷贝目录
    // [
    //   {
    //     from: 'assets',   //from:会自动从module下模块路径下的文件开始查找，
    //     to: '.',          //to:是相对于assetRoot目录
    //     ignore: ['.*'],   //ignore:忽略的文件夹
    //   },
    // ],
    dirCopyMap: null,
  },
};
