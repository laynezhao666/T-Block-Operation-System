'use strict';
const utils = require('./utils');
const webpack = require('webpack');
const { getConfItem, curModuleDir, prodEnvConf } = require('./config');
const merge = require('webpack-merge');
const baseWebpackConfig = require('./webpack.base.conf');
const VueLoaderPlugin = require('vue-loader/lib/plugin');

const ExtractTextPlugin = require('extract-text-webpack-plugin');
const OptimizeCSSPlugin = require('optimize-css-assets-webpack-plugin');
const UglifyJsPlugin = require('uglifyjs-webpack-plugin');

// 引入多页面支持
const webpackHelper = require('./webpack.helper');

const webpackConfig = merge(baseWebpackConfig, {
  module: {
    rules: utils.styleLoaders({
      sourceMap: getConfItem('productionSourceMap'),
      extract: true,
      usePostCSS: true,
    }, curModuleDir),
  },
  devtool: getConfItem('productionSourceMap') ? getConfItem('devtool') : false,
  output: {
    library:getConfItem('libraryPre')?`${getConfItem('libraryPre')}${getConfItem('outPutJsFilename')}`:'',
    libraryTarget: getConfItem('libraryTarget'),
    path: getConfItem('assetsRoot'),
    filename: utils.assetsPath(`${getConfItem('bundleDir')}/${getConfItem('outPutJsFilename')}`),
    chunkFilename: utils.assetsPath(`${getConfItem('bundleDir')}/${getConfItem('outPutChunkFilename')}`),
  },
  plugins: [

    // 添加webpack错误中断处理,抛出错误码
    function()
    {
        this.plugin("done", function(stats)
        {
            if (stats.compilation.errors && stats.compilation.errors.length && process.argv.indexOf('--watch') == -1)
            {
                console.log('======================== webpack构建出错 ========================');
                console.log(stats.compilation.errors);
                process.exit(1); // or throw new Error('webpack build failed.');
            }
            // ...
        });
    },

    ...webpackHelper.getProdCleanWebpackPluginList(),

    new webpack.DefinePlugin(getConfItem('define')),
    new VueLoaderPlugin(),
    // http://vuejs.github.io/vue-loader/en/workflow/production.html
    new webpack.DefinePlugin({
      'process.env': prodEnvConf,
    }),
    new UglifyJsPlugin({
      uglifyOptions: {
        compress: {
          warnings: false,
        },
      },
      sourceMap: getConfItem('productionSourceMap'),
      parallel: true,
    }),

    // extract css into its own file
    new ExtractTextPlugin({
      filename: utils.assetsPath(`${getConfItem('bundleDir')}/${getConfItem('outPutCssFilename')}`),
      // Setting the following option to `false` will not extract CSS from codesplit chunks.
      // Their CSS will instead be inserted dynamically with style-loader
      // when the codesplit chunk has been loaded by webpack.
      // It's currently set to `true` because we are seeing that sourcemaps are
      // included in the codesplit bundle as well when it's `false`,
      // increasing file size: https://github.com/vuejs-templates/webpack/issues/1110
      allChunks: true,
      ignoreOrder: true,
    }),
    // Compress extracted CSS. We are using this plugin so that possible
    // duplicated CSS from different components can be deduped.
    new OptimizeCSSPlugin({
      cssProcessorOptions: getConfItem('productionSourceMap')
        ? { safe: true, map: { inline: false } }
        : { safe: true },
      cssProcessor: require('cssnano'),
    }),
    // generate dist index.html with correct asset hash for caching.
    // you can customize output by editing /index.html
    // see https://github.com/ampedandwired/html-webpack-plugin

    /** ***************** 老的单页面入口 *******************/
    // new HtmlWebpackPlugin({
    //   filename: getConfItem('index'),
    //   template: 'index.html',
    //   inject: true,
    //   minify: {
    //     removeComments: true,
    //     collapseWhitespace: true,
    //     removeAttributeQuotes: true
    //     // more options:
    //     // https://github.com/kangax/html-minifier#options-quick-reference
    //   },
    //   // necessary to consistently work with multiple chunks via CommonsChunkPlugin
    //   chunksSortMode: 'dependency'
    // }),

    // keep module.id stable when vendor modules does not change
    new webpack.HashedModuleIdsPlugin(),
    // enable scope hoisting
    new webpack.optimize.ModuleConcatenationPlugin(),
    // split vendor js into its own file
    //
    // 公共库提取
    // new webpack.optimize.CommonsChunkPlugin({
    //   name: 'vendor',
    //   minChunks (module) {
    //     // any required modules inside node_modules are extracted to vendor
    //     return (
    //       module.resource &&
    //       /\.js$/.test(module.resource) &&
    //       module.resource.indexOf(path.join(__dirname, '../node_modules')) === 0
    //     )
    //   },
    // }),
    // extract webpack runtime and module manifest to its own file in order to
    // prevent vendor hash from being updated whenever index.vue bundle is updated
    // new webpack.optimize.CommonsChunkPlugin({
    //   name: 'manifest',
    //   minChunks: Infinity,
    // }),
    // This instance extracts shared chunks from code splitted chunks and bundles them
    // in a separate chunk, similar to the vendor chunk
    // see: https://webpack.js.org/plugins/commons-chunk-plugin/#extra-async-commons-chunk
    //
    // 公共库提取
    // new webpack.optimize.CommonsChunkPlugin({
    //   name: 'app',
    //   async: 'vendor-async',
    //   children: true,
    //   minChunks: 3,
    // }),

    // 支持多页面项目单独发布
    ...webpackHelper.getProdCommonsChunkPluginList(),

    webpackHelper.getCopyWebpackPlugin(getConfItem('assetsSubDirectory')),

    ...webpackHelper.getProdHtmlWebpackPluginList(),
    new webpackHelper.MicroFronentPlugin(),

    // 前后端分离项目直接返回
    ...webpackHelper.getCustomizePluginList(),
  ],
});

if (getConfItem('productionGzip')) {
  const CompressionWebpackPlugin = require('compression-webpack-plugin');

  webpackConfig.plugins.push(new CompressionWebpackPlugin({
    asset: '[path].gz[query]',
    algorithm: 'gzip',
    test: new RegExp(`\\.(${getConfItem('productionGzipExtensions').join('|')})$`),
    threshold: 10240,
    minRatio: 0.8,
  }));
}

if (getConfItem('bundleAnalyzerReport')) {
  const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer');
  webpackConfig.plugins.push(new BundleAnalyzerPlugin());
}

// 开放自定义配置
const rewriteWebpackConfigFn = getConfItem('rewriteWebpackConfigFn');
if (rewriteWebpackConfigFn && typeof rewriteWebpackConfigFn === 'function') {
  rewriteWebpackConfigFn(webpackConfig);
}

module.exports = webpackConfig;
