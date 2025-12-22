'use strict';
// const path = require('path');

module.exports = {
  common: {
  },
  dev: {
    index: 'ui.html', // 浏览器默认打开页面
    rewriteWebpackConfigFn: (config) => {
      const webpackConfig = config;
      return webpackConfig;
    },
  },
  build: {
  },
};
