// 配置文件处理

'use strict';

global.TNF_urlProjPrefix = ''; // 项目打包url前缀

const merge = require('webpack-merge');
const path = require('path');
const fs = require('fs');
const prodEnv = require('./prod.env');
const devEnv = require('./dev.env');
const _ = require('lodash');

const curModuleDir = path.join(global.TNF_projRootPath, './src/module', global.TNF_curModuleName);

const defaultConf = require('./default');

const envConf = fs.existsSync(path.resolve(global.TNF_projRootPath, './config/.env.js'))
  ? require(`${global.TNF_projRootPath}/config/.env`) : {};

const comConf = fs.existsSync(path.resolve(global.TNF_projRootPath, './config/.common.js'))
  ? require(`${global.TNF_projRootPath}/config/.common`) : {};

const moduleConf = fs.existsSync(path.resolve(global.TNF_projRootPath, `./config/${global.TNF_curModuleName}.js`)) ? require(`${global.TNF_projRootPath}/config/${global.TNF_curModuleName}`) : {};

const localModuleConf = fs.existsSync(path.resolve(global.TNF_projRootPath, `./config/local/${global.TNF_curModuleName}.js`))
  ? require(`${global.TNF_projRootPath}/config/local/${global.TNF_curModuleName}`) : {};

// 生成第一个版本的src.conf.js
// const srcConf = {
//   common: merge({}, comConf.common, moduleConf.common, localModuleConf.common),
//   dev: {
//     common: merge({}, comConf.dev),
//     [global.TNF_curModuleName]: merge({}, defaultConf.dev, moduleConf.dev, localModuleConf.dev),
//   },
//   build: {
//     common: merge({}, comConf.build),
//     [global.TNF_curModuleName]: merge({}, defaultConf.build, moduleConf.build, localModuleConf.build),
//   },
// };

// 生成新的配置
const curModuleConf = {
  dev: merge(
    {},
    defaultConf.dev,
    comConf.common,
    comConf.dev,
    moduleConf.common,
    moduleConf.dev,
    localModuleConf.common,
    localModuleConf.dev
  ),
  build: merge(
    {},
    defaultConf.build,
    comConf.common,
    comConf.build,
    moduleConf.common,
    moduleConf.build,
    localModuleConf.common,
    localModuleConf.build
  ),
};

// 获取配置项
function getConfItem(confkey) {
  if (!global.TNF_curOperName) {
    return false;
  }

  if (_.get(curModuleConf, `${global.TNF_curOperName}.${confkey}`) !== undefined) {
    return curModuleConf[global.TNF_curOperName][confkey];
  }
  return false;
};

/**
 * 打包路径前缀
 */
function getFormatPath(src = '') {
  return getUrlProjPrefix() ? `${getUrlProjPrefix()}/${src}` : src;
};

/**
 * 打包静态资源前缀处理
 */
function getUrlProjPrefix() {
  if (global.TNF_urlProjPrefix) {
    return global.TNF_urlProjPrefix;
  }

  let urlProjPrefix = '';

  // 默认使用项目名
  if (getConfItem('urlProjPrefix') === true) {
    urlProjPrefix = global.TNF_curModuleName;
  } else if (getConfItem('urlProjPrefix')) {
    urlProjPrefix = getConfItem('urlProjPrefix');
  }

  global.TNF_urlProjPrefix = urlProjPrefix;

  // 兼容系统老变量
  // eslint-disable-next-line no-underscore-dangle
  global._urlProjPrefix = urlProjPrefix;

  return urlProjPrefix;
};


// console.log('*************curModuleConf*********');
// console.log(curModuleConf);
// console.log('*************curModuleConf*********');

// exports.srcConf = srcConf;

exports.curModuleConf = curModuleConf;
exports.curModuleDir = curModuleDir;
exports.getConfItem = getConfItem;
exports.getFormatPath = getFormatPath;
exports.getUrlProjPrefix = getUrlProjPrefix;
exports.devEnvConf = merge(devEnv, envConf.common, envConf.dev);
exports.prodEnvConf = merge(prodEnv, envConf.common, envConf.build);
