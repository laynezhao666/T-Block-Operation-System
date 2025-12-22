/* eslint-disable linebreak-style */
'use strict';
const path = require('path');
const LOCAL_HOST = '';
const DEV_HOST = '';

module.exports = {
  common: {
    extensions: ['.js', '.vue', '.json', '.ts', '.tsx'],
  },
  dev: {
    host: LOCAL_HOST,
    urlProjPrefix: false,
    proxyTable: {
      '/cgi/idc-tbos-cmdb': {
        target: `http://${DEV_HOST}`,
        changeOrigin: true,
        pathRewrite: {
          '^/cgi/idc-tbos-cmdb': '' // 保留部分路径（按需调整）
        }
      },
      '/cgi': {
        target: `http://${DEV_HOST}`,
        changeOrigin: true,
      },

    },
  },
  build: {
    isMicrofrontMainNav: true,
    urlProjPrefix: false,
    assetsRoot: path.resolve(__dirname, '../dist/main'),
  },
};
