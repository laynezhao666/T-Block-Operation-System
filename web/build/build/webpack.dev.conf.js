'use strict';
const utils = require('./utils');
const webpack = require('webpack');
const { getConfItem, curModuleDir, devEnvConf } = require('./config');
const merge = require('webpack-merge');
const path = require('path');
const baseWebpackConfig = require('./webpack.base.conf');
const VueLoaderPlugin = require('vue-loader/lib/plugin');
const FriendlyErrorsWebpackPlugin = require('friendly-errors-webpack-plugin');
// 引入多页面支持
const webpackHelper = require('./webpack.helper');

const { HOST } = process.env;
const PORT = process.env.PORT && Number(process.env.PORT);

// add hot-reload related code to entry chunks
Object.keys(baseWebpackConfig.entry).forEach((name) => {
  baseWebpackConfig.entry[name] = [`${global.TNF_cliPath}/dev-client`].concat(baseWebpackConfig.entry[name]);
});

const devWebpackConfig = merge(baseWebpackConfig, {
  output: {
    library:`${getConfItem('libraryPre')}`,
    libraryTarget: getConfItem('libraryTarget'),
  },
  module: {
    rules: utils.styleLoaders({ sourceMap: getConfItem('cssSourceMap'), usePostCSS: true }, curModuleDir),
  },
  // cheap-module-eval-source-map is faster for development
  devtool: getConfItem('devtool'),

  // these devServer options should be customized in /config/index.js
  devServer: {
    clientLogLevel: 'warning',
    historyApiFallback: {
      rewrites: [
        { from: /.*/, to: path.posix.join(getConfItem('assetsPublicPath'), 'index.html') },
      ],
    },
    hot: true,
    contentBase: false, // since we use CopyWebpackPlugin.
    compress: true,
    https: true, // 开启https
    host: HOST || getConfItem('host'),
    port: PORT || getConfItem('port'),
    open: getConfItem('autoOpenBrowser'),
    overlay: getConfItem('errorOverlay')
      ? { warnings: false, errors: true }
      : false,
    publicPath: getConfItem('assetsPublicPath'),
    proxy: getConfItem('proxyTable'),
    quiet: true, // necessary for FriendlyErrorsPlugin
    // watchOptions: {
    //   ignored: '/node_modules/',
    //   poll: getConfItem('poll'),
    // },
    useLocalIp: true,
  },
  // watch: true,
  // watchOptions: {
  //   poll: getConfItem('poll'), // 轮询间隔时间
  //   aggregateTimeout: 500, // 防抖（在输入时间停止刷新计时）
  //   ignored: /node_modules/,
  // },
  plugins: [
    new webpack.DefinePlugin(getConfItem('define')),
    new VueLoaderPlugin(),
    new webpack.DefinePlugin({
      'process.env': devEnvConf,
    }),
    new webpack.HotModuleReplacementPlugin(),
    new webpack.NamedModulesPlugin(), // HMR shows correct file names in console on update.
    new webpack.NoEmitOnErrorsPlugin(),

    new FriendlyErrorsWebpackPlugin(),

    // 调试时提取公共库，减少内存开销
    new webpack.optimize.CommonsChunkPlugin({
      name: 'vendor',
      minChunks(module) {
        // any required modules inside node_modules are extracted to vendor
        return (
          module.resource
          && /\.js$/.test(module.resource)
          && module.resource.indexOf(path.join(global.TNF_projRootPath, './node_modules')) === 0
        );
      },
    }),
    new webpack.optimize.CommonsChunkPlugin({
      name: 'manifest',
      minChunks: Infinity,
    }),

    webpackHelper.getCopyWebpackPlugin(getConfItem('assetsSubDirectory')),

    ...webpackHelper.getDevHtmlWebpackPluginList(),
  ],
});

// 开放自定义配置
const rewriteWebpackConfigFn = getConfItem('rewriteWebpackConfigFn');
if (rewriteWebpackConfigFn && typeof rewriteWebpackConfigFn === 'function') {
  rewriteWebpackConfigFn(devWebpackConfig);
}

module.exports = devWebpackConfig;
