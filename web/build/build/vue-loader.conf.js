'use strict';
const utils = require('./utils');
const { curModuleConf, curModuleDir } = require('./config');
const isProduction = process.env.NODE_ENV === 'production';
const sourceMapEnabled = isProduction
  ? curModuleConf.build.productionSourceMap
  : curModuleConf.dev.cssSourceMap;

module.exports = {
  loaders: utils.cssLoaders({
    sourceMap: sourceMapEnabled,
    extract: isProduction,
  }, curModuleDir),
  cssSourceMap: sourceMapEnabled,
  cacheBusting: curModuleConf.dev.cacheBusting,
  transformToRequire: {
    video: ['src', 'poster'],
    source: 'src',
    img: 'src',
    image: 'xlink:href',
  },
};
