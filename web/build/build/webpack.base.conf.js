'use strict';
// const path = require('path');
const utils = require('./utils');
const { getConfItem } = require('./config');
const chalk = require('chalk');
const vueLoaderConfig = require('./vue-loader.conf');
require('babel-polyfill');// 兼容IE

// 引入多页面支持
const webpackHelper = require('./webpack.helper');

/* const createLintingRule = () => ({
  test: /\.(js|vue)$/,
  loader: 'eslint-loader',
  enforce: 'pre',
  include: [resolve('src'), resolve('test')],
  options: {
    formatter: require('eslint-friendly-formatter'),
    emitWarning: !getConfItem('showEslintErrorsInOverlay')
  }
}) */

if (Object.keys(webpackHelper.getEntries()).length === 0) {
  console.error(chalk.red('没有有效的入口页面'));
  process.exit();
}

module.exports = {
  context: utils.resolve(''), // 查找入口文件的根目录
  // entry: {
  //   app: './src/main.js'
  // },
  entry: webpackHelper.getEntries(), // 多入口
  output: {
    path: getConfItem('assetsRoot'),
    filename: '[name].js',
    publicPath: getConfItem('assetsPublicPath'),
  },
  resolve: {
    symlinks: false, // 避免链接检查
    extensions: getConfItem('extensions'),
    alias: getConfItem('alias'),
    modules: [utils.resolve('src'), 'node_modules'],
  },
  externals: getConfItem('externals'),
  module: {
    noParse: [/videojs-contrib-hls/],
    rules: [
      { parser: { System: false } },
      {
        test: /\.vue$/,
        loader: 'vue-loader',
        options: vueLoaderConfig,
      },
      {
        test: /\.tsx?$/,
        use: 'ts-loader',
        exclude: /node_modules/,
      },
      {
        test: /\.js$/,
        loader: 'babel-loader',
        include: [
          utils.resolve('src'),
          utils.resolve('node_modules/webpack-dev-server/client'),
          utils.resolve('node_modules/@tencent/tnweb-common-utils'),
          utils.resolve('node_modules/@tencent/tnweb-common-feature'),
        ],
      },
      {
        test: /\.jade$/,
        loader: 'jade',
      },
      {
        test: /\.(png|jpe?g|gif|svg|webp)(\?.*)?$/,
        loader: 'url-loader',
        options: {
          limit: 10000,
          name: utils.assetsPath('img/[name].[hash:7].[ext]', 'img'),
        },
      },
      {
        test: /\.(mp4|webm|ogg|mp3|wav|flac|aac)(\?.*)?$/,
        loader: 'url-loader',
        options: {
          limit: 10000,
          name: utils.assetsPath('media/[name].[hash:7].[ext]','media'),
        },
      },
      {
        test: /\.(woff2?|eot|ttf|otf)(\?.*)?$/,
        loader: 'url-loader',
        options: {
          limit: 10000,
          name: utils.assetsPath('fonts/[name].[hash:7].[ext]','fonts'),
        },
      },
    ],
  },
  node: {
    // prevent webpack from injecting useless setImmediate polyfill because Vue
    // source contains it (although only uses it if it's native).
    setImmediate: false,
    // prevent webpack from injecting mocks to Node native modules
    // that does not make sense for the client
    dgram: 'empty',
    fs: 'empty',
    net: 'empty',
    tls: 'empty',
    child_process: 'empty',
  },
};
