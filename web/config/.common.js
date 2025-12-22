'use strict';
const path = require('path');

const LOCAL_HOST = '';
const DEV_HOST = '';
const TEST_HOST = '';
const PRE_HOST = '';
const PUBLISH_HOST = '';

function resolve(dir) {
  return path.join(__dirname, '..', dir);
}

module.exports = {
  common: {
    alias: {
      '@template': resolve('src/template'),
      '@smpage_index': resolve('src/module/smpage/index'),
      '@module': resolve(`src/feature/${global.TNF_curModuleName}`),
      common: resolve('src/static/tnweb-common-utils/dist'),
      feature: resolve('src/feature'),
      component: resolve('src/feature/component'),
      '@tencent/TNWeb-ui/packages/chart/src/mixin.js': 'src/utils/tnweb-ui-chart-mixin-override.js',
      '@tencent/TNWeb-ui': resolve('src/static/thirdparty/tnwebui'),
      '@tencent/nebulaplayer': resolve('src/static/thirdparty/NebulaPlayer.js'),
    },
    define: {
      LOGIN: JSON.stringify({
        local: {
          LOGIN_ORIGIN: '',
        },
        dev: {
          LOGIN_ORIGIN: '',
        },
        test: {
          LOGIN_ORIGIN: '',
        },
        pre: {
          LOGIN_ORIGIN: '',
        },
        publish: {
          LOGIN_ORIGIN: '',
        },
      }),
      BPM: JSON.stringify({
        local: {
          PC_ORIGIN: ''
        },
        dev: {
          PC_ORIGIN: ''
        },
        test: {
          PC_ORIGIN: ''
        },
        pre: {
          PC_ORIGIN: ''
        },
        publish: {
          PC_ORIGIN: ''
        },
      }),
      FAIRY: JSON.stringify({
        local: {
          PC_ORIGIN: '',
        },
        dev: {
          PC_ORIGIN: '',
          // 没有自测环境
        },
        test: {
          PC_ORIGIN: '',
        },
        pre: {
          // 没有预发布环境
          PC_ORIGIN: '',
        },
        publish: {
          PC_ORIGIN: '',
        },
      }),
      // TODO: CGI地址
      HOST: JSON.stringify({
        local: '',
        dev: DEV_HOST,
        test: TEST_HOST,
        pre: PRE_HOST,
        publish: PUBLISH_HOST,
      }),
      // 主站地址
      HOME: JSON.stringify({
        local: LOCAL_HOST,
        dev: DEV_HOST,
        test: TEST_HOST,
        pre: PRE_HOST,
        publish: PUBLISH_HOST,
      }),
    },
    rewriteWebpackConfigFn(config) {
      config.module.rules.forEach((item) => {
        if (item.use === 'ts-loader') {
          // 改为只编译、不做类型检查，加快速度（即使编译单个简单文件类型检查也至少耗费6秒）
          // eslint-disable-next-line no-param-reassign
          item.use = [{
            loader: 'ts-loader',
            options: {
              transpileOnly: true,
              appendTsSuffixTo: [/\.vue$/],
            },
          }];
        }

        if (item.test && item.test.toString() === /\.js$/.toString()) {
          // eslint-disable-next-line no-param-reassign
          item.exclude = [
            // /echarts\.esm\.min\.js/,
            /showcase\/warn\/core\/ht\.js$/,
          ];
          item.include.push(path.join(
            global.TNF_projRootPath,
            'node_modules/vue-virtual-scroller/dist/vue-virtual-scroller.esm.js',
          ),);
          item.include.push(path.join(
            global.TNF_projRootPath,
            'node_modules/v2-virtual-tree',
          ),);
          item.include.push(path.join(
            global.TNF_projRootPath,
            'node_modules/yaml',
          ),);
        }
      });
    },
  },

  dev: {
    externals: {
      vue: 'Vue',
      'vue-i18n': 'VueI18n',
      '@tencent/TNWeb-ui': 'ELEMENT',
      // 'element-ui': 'element-ui',
      'element-ui': 'ELEMENT',
      echarts: 'echarts',
      '@tencent/tnweb-icon': 'TNWEBICON',
      $: 'jQuery',
      jquery: 'jQuery',
      'single-spa': 'single-spa',
    },
  },
  build: {
    externals: {
      vue: 'Vue',
      'vue-i18n': 'VueI18n',
      '@tencent/TNWeb-ui': 'ELEMENT',
      '@tencent/tnweb-icon': 'TNWEBICON',
      'element-ui': 'ELEMENT',
      echarts: 'echarts',
      $: 'jQuery',
      jquery: 'jQuery',
      'single-spa': 'single-spa',
      'moment': 'moment',
      'axios': 'axios',
      'qs': 'Qs',
      'lodash': '_'
    },
  },
};
