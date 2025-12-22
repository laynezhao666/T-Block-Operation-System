'use strict';
require('./process-env-argv');
require('./check-versions')();
process.env.NODE_ENV = 'production';
global.TNF_curOperName = 'build';
require('./init');

const ora = require('ora');
// const rm = require('rimraf')
const chalk = require('chalk');
const webpack = require('webpack');

const webpackConfig = require('./webpack.prod.conf');
// const webpackHelper = require('./webpack.helper')

/* param check end */
console.log('> Building for production...');
const spinner = ora('building for production...');
spinner.start();

// 移出项目static文件夹（如果支持页面级别的打包，不需要移出）
// let rm_path = path.join(config.build.assetsRoot, global.TNF_curModuleName, config.build.assetsSubDirectory)
// rm(rm_path, err => {
//   if (err) throw err

webpack(webpackConfig, (err, stats) => {
  spinner.stop();
  if (err) throw err;
  process.stdout.write(`${stats.toString({
    colors: true,
    modules: false,
    children: false, // If you are using ts-loader
    chunks: false,
    chunkModules: false,
  })}\n\n`);

  if (stats.hasErrors()) {
    console.log(chalk.red('  Build failed with errors.\n'));
    process.exit(1);
  }

  console.log(chalk.cyan('  Build complete.\n'));
  console.log(chalk.yellow('  Tip: built files are meant to be served over an HTTP server.\n'
      + '  Opening index.html over file:// won\'t work.\n'));
});
// })
