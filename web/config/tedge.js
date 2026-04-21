const DEV_HOST = 'localhost:8080';
const LOCAL_HOST = 'localhost';
const path = require('path');
module.exports = {
  common: {
    define: {
      IS_TEDGE: true,
    },
    extensions: ['.js', '.vue', '.json', '.ts', '.tsx'],
    rewriteWebpackConfigFn(config) {
      config.module.rules.forEach((item) => {
        if (item.use === 'ts-loader') {
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
            /echarts\.esm\.min\.js/,
            /echarts\.v5\.min\.js/,
            /showcase\/warn\/core\/ht\.js$/,
            /core\/ht\.js$/,
            /feature\/ht\/sdk/,
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
    libraryTarget: 'umd',
    host: LOCAL_HOST,
    proxyTable: {
      '/api/dcos': {
        target: `http://${DEV_HOST}`,
        changeOrigin: true,
      },
      '/cgi/idc-tbos-cgi': {
        target: `http://${DEV_HOST}`,
        changeOrigin: true,
      },
      '/cgi/tedge-bff': {
        target: `http://${DEV_HOST}`,
        changeOrigin: true,
      },
      '/cgi/singlegraph': {
        target: `http://${DEV_HOST}`,
        changeOrigin: true,
      },
      '/cgi': {
        target: `http://${DEV_HOST}`,
        changeOrigin: true,
      },
      // 临时注释 WebSocket 代理，避免连接失败导致进程崩溃
      // '/ws': {
      //   target: `ws://${DEV_HOST}`,
      //   changeOrigin: true,
      //   ws: true,
      // },
    },
  },
  build: {
    libraryTarget: 'umd',
    assetsRoot: path.resolve(__dirname, '../dist/main'),
  },
};
