'use strict';
const path = require('path');
const fs = require('fs');
const ExtractTextPlugin = require('extract-text-webpack-plugin');
const { getFormatPath, getConfItem } = require('./config');

exports.resolve = function (dir) {
  // return path.join(__dirname, '..', dir)
  return path.join(global.TNF_projRootPath, dir);
};

/**
* 输出文件夹，每个module一个static
*/
exports.assetsPath = function (_path,type) {
  const assetsSubDirectory = getConfItem('assetsSubDirectory');

  return path.posix.join(getFormatPath(assetsSubDirectory), _path);
};

exports.cssLoaders = function (options, cwd) {
  // eslint-disable-next-line no-param-reassign
  options = options || {};

  const cssLoader = {
    loader: 'css-loader',
    options: {
      sourceMap: options.sourceMap,
    },
  };

  const postcssLoader = {
    loader: 'postcss-loader',
    options: {
      sourceMap: options.sourceMap,

      // 解决移动端自定义配置，采取配置文件形式提取到根目录
      // plugins: () => [
      //   require('autoprefixer')({ browsers: 'last 5 version' }), // CSS浏览器兼容
      // ],
    },
  };

  const varFile = path.join(cwd, 'style/_variable.scss');

  // generate loader string to be used with extract text plugin
  function generateLoaders(loader, loaderOptions) {
    const loaders = options.usePostCSS ? [cssLoader, postcssLoader] : [cssLoader];

    if (loader) {
      loaders.push({
        loader: `${loader}-loader`,
        options: Object.assign({}, loaderOptions, {
          sourceMap: options.sourceMap,
        }),
      });
    }

    // Extract CSS when that option is specified
    // (which is the case during production build)
    if (options.extract) {
      return ExtractTextPlugin.extract({
        use: loaders,
        // 解决字体，图片相对引用路径问题
        publicPath: getConfItem('assetsPublicPath')==='/'?'../../../../':'../../../',
        fallback: 'vue-style-loader',
      });
    }
    return ['vue-style-loader'].concat(loaders);
  }

  // https://vue-loader.vuejs.org/en/configurations/extract-css.html
  return {
    css: generateLoaders(),
    postcss: generateLoaders(),
    less: generateLoaders('less'),
    sass: generateLoaders('sass', { indentedSyntax: true }),
    scss: generateLoaders('sass').concat(fs.existsSync(varFile) && {
      loader: 'sass-resources-loader',
      options: {
        resources: varFile,
      },
    })
      .filter(Boolean),
    // scss: generateLoaders('sass').concat(
    //   {
    //     loader: 'sass-resources-loader',
    //     options: {
    //       resources: path.resolve(__dirname, '../src/module/_variable.scss')
    //     }
    //   }
    // ),
    stylus: generateLoaders('stylus'),
    styl: generateLoaders('stylus'),
  };
};

// Generate loaders for standalone style files (outside of .vue)
exports.styleLoaders = function (options, cwd) {
  const output = [];
  const loaders = exports.cssLoaders(options, cwd);

  // eslint-disable-next-line no-restricted-syntax
  for (const extension in loaders) {
    const loader = loaders[extension];
    output.push({
      test: new RegExp(`\\.${extension}$`),
      use: loader,
    });
  }

  return output;
};

// exports.createNotifierCallback = () => {
//   const notifier = require('node-notifier')

//   return (severity, errors) => {
//     if (severity !== 'error') return

//     const error = errors[0]
//     const filename = error.file && error.file.split('!').pop()

//     notifier.notify({
//       title: packageConfig.name,
//       message: severity + ': ' + error.name,
//       subtitle: filename || '',
//     })
//   }
// }
