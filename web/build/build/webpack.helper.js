/* eslint-disable no-restricted-syntax */
/* eslint-disable no-param-reassign */
/* eslint-disable no-underscore-dangle */
/**
 * 多页面打包辅助工具
 * 多页面=多业务=每个业务一个页面=多个单页面打包
 */

const path = require('path');
const fs = require('fs');
// const glob = require('glob')
const { isSplitDeploy } = require('./process-env-argv');
const webpack = require('webpack');
const merge = require('webpack-merge');
// const HtmlWebpackPlugin = require('html-webpack-plugin')
const HtmlWebpackPlugin = require('html-webpack-plugin-for-multihtml');
const CleanWebpackPlugin = require('clean-webpack-plugin');
const CopyWebpackPlugin = require('copy-webpack-plugin');
const PhpWebpackPlugin = require('./webpack-plugin-php');

let moduleList; // 缓存多页面模块列表
let entryList; // 缓存入口
const moduleRootPath = './src/module'; // 模块根目录(这个可以根据自己的需求命名)
const templatePath = './src/template'; // 模板路径

const curModuleName = global.TNF_curModuleName; // 当前要打包的模块名称
const pageList = global.TNF_argvs_pages;

// var is_use_entry = true // 是否使用入口文件
const initJs = 'main.js'; // 模块入口js
const initTs = 'main.ts'; // 模块入口ts
const initHtml = 'index.html'; // 模板文件

global.TNF_webIndex = ''; // 项目入口文件

const { getFormatPath, getConfItem } = require('./config');
const showWebpackConfigMsg = getConfItem('showWebpackConfigMsg');
let entryJS = getConfItem('entryJS') || [];
entryJS = Array.from(new Set(entryJS));

exports.curModuleDir = path.join(global.TNF_projRootPath, './src/module', curModuleName);

/**
 * 获取js入口数组
 */
exports.getEntries = function getEntries() {
  if (entryList) {
    return entryList;
  }

  const entries = {};

  this.getModuleList();

  moduleList.forEach((module) => {
    if (module.moduleID !== '' && module.moduleJS !== '') {
      entries[module.moduleID] = entryJS.concat([module.moduleJS]);
    }
  });

  if (showWebpackConfigMsg) {
    console.log('\n********* entries **********');
    console.log(entries);
  }

  entryList = entries;
  return entries;
};

const modulePath = path.resolve(`${moduleRootPath}/${curModuleName}`);

/**
 * 获取多页面模块列表
 * @returns {模块的信息集合}
 */
exports.getModuleList = function getModuleList() {
  if (moduleList) {
    return moduleList;
  }
  moduleList = [];

  // 获取目录下所有文件及文件夹
  const pageDirs = fs.readdirSync(modulePath);

  // 过滤一级文件夹检测
  const filterEntryDir = getConfItem('filterEntryDir');

  for (const curDirName of pageDirs) {
    const filepath = `${modulePath}/${curDirName}`;

    const info = fs.statSync(filepath);

    // 文件夹
    if (info.isDirectory() && !filterEntryDir.includes(curDirName)) {
      parseModuleDir(filepath, curDirName);
    }
  }

  if (showWebpackConfigMsg) {
    console.log('\n********* moduleList ********');
    console.log(moduleList);
  }

  return moduleList;
};

/**
 * 获取dev的Html模板集合
 * @returns {dev的Html模板集合}
 */
exports.getDevHtmlWebpackPluginList = function getDevHtmlWebpackPluginList() {
  // 缓存dev的Html模板集合
  const devHtmlWebpackPluginList = [];

  // 获取多页面模块集合
  const moduleList = this.getModuleList();

  // 遍历生成模块的HTML模板
  const confList = [];
  moduleList.forEach((mod) => {
    // 生成配置
    const conf = {
      moduleName: mod.pageName,
      // filename: curModuleName + '/' + mod.moduleID + '.html',
      filename: getFormatPath(`${mod.pageName}.html`),
      template: mod.moduleHTML,
      inject: true,
      customTemp: getCustomTempConf(mod.pageName),
      urlProjPrefix: getFormatPath(),
      multihtmlCache: true, // 多页实时编译慢的问题
      // chunks: [mod.moduleID],
      chunksSortMode: 'manual', // 改为手动调整
      chunks: ['manifest', 'vendor', mod.moduleID], // 开发模式使用公共vendor，减少内存开销
    };
    confList.push(conf);

    // 添加HtmlWebpackPlugin对象
    devHtmlWebpackPluginList.push(new HtmlWebpackPlugin(conf));
  });

  if (showWebpackConfigMsg) {
    console.log('\n******** devHtmlWebpackPluginList ********');
    console.log(confList);
  }
  return devHtmlWebpackPluginList;
};

/**
 * 获取prod的Html模板集合
 * @returns {prod的Html模板集合}
 */
exports.getProdHtmlWebpackPluginList = function getProdHtmlWebpackPluginList() {
  // umd 模式不生成页面 优化构建速度
  if (getConfItem('libraryTarget') === 'umd') {
    return [];
  }
  const prodHtmlWebpackPluginList = [];

  // 获取多页面模块集合
  const moduleList = this.getModuleList();
  const confList = [];
  // 遍历生成模块的HTML模板
  moduleList.forEach((mod) => {
    // 生成配置
    const conf = {
      // filename: curModuleName + '/' + mod.pageName + '.html',
      filename: getFormatPath(`${mod.pageName}.html`),
      template: mod.moduleHTML,
      inject: true,
      customTemp: getCustomTempConf(mod.pageName),
      urlProjPrefix: getFormatPath(),
      minify: {
        removeComments: true,
        collapseWhitespace: true,
        removeAttributeQuotes: true,
        // more options:
        // https://github.com/kangax/html-minifier#options-quick-reference
      },
      // necessary to consistently work with multiple chunks via CommonsChunkPlugin
      // 函数方式调整
      // chunksSortMode: function (a, b) {
      //   if (a.names[0].endsWith('manifest')) {
      //     return -1
      //   } else if (a.names[0].endsWith('vendor')) {
      //     if (b.names[0].endsWith('manifest')) {
      //       return 1
      //     } else {
      //       return -1
      //     }
      //   } else {
      //     return 1
      //   }
      // },
      chunksSortMode: 'manual', // 改为手动调整
      chunks: [`${mod.pageName}/manifest`, `${mod.pageName}/vendor`, mod.moduleID],
    };

    confList.push(conf);

    // 添加HtmlWebpackPlugin对象
    prodHtmlWebpackPluginList.push(new HtmlWebpackPlugin(conf));
  });

  if (showWebpackConfigMsg) {
    console.log('\n******* prodHtmlWebpackPluginList *******');
    console.log(confList);
  }

  return prodHtmlWebpackPluginList;
};
/**
 * @returns {自定义插件}
 */
exports.getCustomizePluginList = function getCustomizePluginList() {
  const customizePluginList = [];

  if (getConfItem('urlProjPrefix') && getConfItem('libraryTarget') !== 'umd') {
    customizePluginList.push(new PhpWebpackPlugin());
  }
  return customizePluginList;
};

/**
 * 提取共用库
 * @returns {prod的Html模板集合}
 */
exports.getProdCommonsChunkPluginList = function getProdCommonsChunkPluginList() {
  // 微前端模式不处理
  if (getConfItem('libraryTarget') === 'umd' || getConfItem('isMicrofrontMainNav')) {
    return [];
  }
  // 缓存dev的Html模板集合
  const CommonsChunkPluginList = [];

  // 获取多页面模块集合
  const moduleList = this.getModuleList();
  const confList = [];
  // 遍历生成模块的HTML模板
  moduleList.forEach((mod) => {
    // 生成配置
    let conf = {
      name: `${mod.pageName}/vendor`,
      minChunks(module) {
        // any required modules inside node_modules are extracted to vendor
        return (
          module.resource
          && /\.js$/.test(module.resource)
          && module.resource.indexOf(path.join(global.TNF_projRootPath, './node_modules')) === 0
        );
      },
      chunks: [mod.moduleID],
    };

    confList.push(conf);
    CommonsChunkPluginList.push(new webpack.optimize.CommonsChunkPlugin(conf));
    // extract webpack runtime and module manifest to its own file in order to
    // prevent vendor hash from being updated whenever index.vue bundle is updated

    conf = {
      name: `${mod.pageName}/manifest`,
      minChunks: Infinity,
      chunks: [`${mod.pageName}/vendor`, mod.moduleID],
    };
    confList.push(conf);
    CommonsChunkPluginList.push(new webpack.optimize.CommonsChunkPlugin(conf));

    // new webpack.optimize.CommonsChunkPlugin({
    //   name: 'app',
    //   async: 'vendor-async',
    //   children: true,
    //   minChunks: 3,
    // }),
  });

  if (showWebpackConfigMsg) {
    console.log('\n******* CommonsChunkPluginList *******');
    console.log(confList);
  }

  return CommonsChunkPluginList;
};

/**
 * 获取prod的Html模板集合
 * @returns {prod的Html模板集合}
 */
exports.getProdCleanWebpackPluginList = function getProdCleanWebpackPluginList() {
  const CleanWebpackPluginList = [];
  const clearbasepath = `${getConfItem('assetsRoot')}/${getFormatPath(getConfItem('assetsSubDirectory'))}/${getConfItem('bundleDir')}`;

  // 获取多页面模块集合
  const moduleList = this.getModuleList();
  const confList = [];
  // 遍历生成模块的HTML模板
  moduleList.forEach((mod) => {
    // 生成配置
    confList.push(`${clearbasepath}/${mod.pageName}/*.*`);
  });

  if (getConfItem('isCleanLastBundlefile')) {
    CleanWebpackPluginList.push(new CleanWebpackPlugin(confList, {
      root: global.TNF_projRootPath,
      allowExternal: true,
    }));
  } else {
    CleanWebpackPluginList.push(new CleanWebpackPlugin());
  }

  if (showWebpackConfigMsg) {
    console.log('\n******* CleanWebpackPluginList *******');
    console.log(confList);
  }

  return CleanWebpackPluginList;
};

/**
 * copt static asset
 * @returns new CopyWebpackPlugin static
 */
exports.getCopyWebpackPlugin = function getCopyWebpackPlugin(assetsSubDirectory) {
  let staticPath = path.resolve('./src/static');
  const projStaticPath = path.resolve(`./src/module/${curModuleName}/static`);

  const copyList = [];

  // 兼容微前端项目
  if (getConfItem('libraryTarget') === 'umd') {
    // 项目static
    if (fs.existsSync(projStaticPath)) {
      copyList.push({
        from: projStaticPath,
        to: getFormatPath(assetsSubDirectory), // 项目static复制到项目文件夹下
        ignore: ['.*'],
      });
    }

    // 公共static
    if (fs.existsSync(staticPath)) {
      copyList.push({
        from: staticPath,
        to: assetsSubDirectory, // 公共static复制打包根文件夹下
        ignore: ['.*'],
      });
    }
  } else {
    // 有项目的static，则不使用公共static
    if (fs.existsSync(projStaticPath)) {
      staticPath = projStaticPath;
    }

    if (fs.existsSync(staticPath)) {
      copyList.push({
        from: staticPath,
        to: this.getFormatPath(assetsSubDirectory),
        ignore: ['.*'],
      });
    }
  }

  const dirCopyMap = getConfItem('dirCopyMap');
  if (dirCopyMap) {
    dirCopyMap.map((item) => {
      copyList.push({
        from: path.resolve(`./src/module/${curModuleName}/`, item.from),
        to: path.join(getConfItem('assetsRoot'), item.to),
        ignore: item.ignore || ['.*'],
      });
    });
  }

  if (copyList.length > 0) {
    return new CopyWebpackPlugin(copyList);
  }
  return new CopyWebpackPlugin();
};

const projTemp = `${moduleRootPath}/${curModuleName}/${initHtml}`;
const allProjTemp = `${templatePath}/${initHtml}`;
const projPath = path.resolve(`${moduleRootPath}/${curModuleName}`);

/**
 * 深度遍历目录，并整理多页面模块
 * @param filepath 当前递归的文件夹路径，包括文件夹名称，
 *        如：D:\workspace\TNWeb-tnfusion\src\module\demo\pages\demo1
 * @param curDirName 模块名称,当前递归的文件夹名称，如：demo1
 */
function parseModuleDir(filepath, curDirName) {
  // 缓存模块对象
  const module = { moduleID: '', pageName: '', moduleHTML: '', moduleJS: '' };
  module.moduleID = `${curDirName}/${curDirName}`;
  module.pageName = curDirName;

  // 获取目录下所有文件及文件夹
  const pa = fs.readdirSync(filepath);

  // 有效页面
  const projPageValidPath = [
    `${projPath}/${curDirName}/${initJs}`,
    `${projPath}/${curDirName}/${initTs}`,
    `${projPath}/pages/${curDirName}/${initJs}`,
    `${projPath}/pages/${curDirName}/${initTs}`,
  ];

  for (const ele of pa) {
    const elePath = `${filepath}/${ele}`;
    const pageTemp = `${filepath}/${initHtml}`;

    const info = fs.statSync(elePath);

    // 文件夹
    if (info.isDirectory()) {
      parseModuleDir(elePath, ele);
    } else {
      // 判断入口文件是否存在
      if (projPageValidPath.includes(elePath)) {
        // 检查与用模糊输入的页面是否匹配
        if (!checkInputPagesMatch(curDirName)) {
          continue;
        }

        module.moduleJS = elePath;

        // 模板优先级 pages>moudle>template
        if (fs.existsSync(pageTemp)) {
          module.moduleHTML = pageTemp;
        } else if (fs.existsSync(projTemp)) {
          module.moduleHTML = projTemp;
        } else {
          module.moduleHTML = allProjTemp;
        }
      }
    }
  }

  // 判断模块是否真实
  if ((module.moduleID !== '' && module.moduleHTML !== '' && module.moduleJS !== '')) {
    moduleList.push(module);
  }
}
/**
 * 检查当前遍历页面是否与输入
 * @param {*} curpage
 */
function checkInputPagesMatch(curpage) {
  // 如果用户为输入页面，打包全部页面

  if (pageList.length === 0) {
    global.TNF_webIndex = global.TNF_webIndex || `${curpage}.html`;
    return true;
  }

  // 检查输入
  for (let page of pageList) {
    // if (curpage.includes(page)) {

    // 改为支持*输入
    page = page.replace('*', '.*');
    if (curpage.search(page) !== -1) {
      global.TNF_webIndex = global.TNF_webIndex || `${curpage}.html`;
      return true;
    }
  }

  // 未匹配到页面
  return false;
}

/**
 *
 * 获取自定义页面模板配置
 */
function getCustomTempConf(curPageHtml) {
  const tempBasePath = path.resolve(templatePath);
  const customTempConf = fs.existsSync(`${tempBasePath}/config.js`)
    ? require(`${tempBasePath}/config.js`) : null;
  const env = process.env.NODE_ENV;

  if (!customTempConf) {
    return null;
  }

  let [commonTemp, projTemp, pageTemp] = [{}, {}, {}];

  if (customTempConf[env]) {
    // 返回公共temp
    if (customTempConf[env]._common_) {
      commonTemp = customTempConf[env]._common_;
    }

    // 返回项目的temp
    if (customTempConf[env][curModuleName] && customTempConf[env][curModuleName]._common_) {
      projTemp = customTempConf[env][curModuleName]._common_;
    }

    // 返回页面temp
    if (customTempConf[env][curModuleName]) {
      // eslint-disable-next-line no-restricted-syntax
      for (const page in customTempConf[env][curModuleName]) {
        if (curPageHtml.search(page.replace('*', '.*')) !== -1) {
          pageTemp = customTempConf[env][curModuleName][page];
          break;
        }
      }
    }
  }

  return merge(commonTemp, projTemp, pageTemp);
}

exports.getFormatPath = getFormatPath;

exports.MicroFronentPlugin = MicroFronentPlugin;

function getSystemImportmap(apps) {
  const map = { imports: {}, scopes: {} };
  // eslint-disable-next-line no-restricted-syntax
  for (const k in apps) {
    const config = apps[k];
    const { scripts } = config;
    if (config.type > 1 && scripts && scripts.entry) {
      const { depts } = scripts;
      map.imports[config.id] = scripts.entry;
      if (depts) {
        if (depts.imports) {
          // eslint-disable-next-line no-restricted-syntax
          for (const i in depts.imports) {
            if (!map.imports[i]) {
              map.imports[i] = depts.imports[i];
            }
          }
        }
        if (depts.scopes) {
          // eslint-disable-next-line no-restricted-syntax
          for (const i in depts.scopes) {
            if (!map.scopes[i]) {
              map.scopes[i] = {};
            }
            // eslint-disable-next-line no-restricted-syntax
            for (const m in depts.scopes[i]) {
              if (!map.scopes[i][m]) {
                map.scopes[i][m] = depts.scopes[i][m];
              }
            }
          }
        }
      }
    }
  }
  return map;
}

const regMrocFrontend = /(.*?)\/static\/js-css\/(.*?)\/([^.]*)(.*)/;

// eslint-disable-next-line no-unused-vars
function MicroFronentPlugin(options) {}

MicroFronentPlugin.prototype.apply = function (compiler) {
  if (getConfItem('libraryTarget') !== 'umd') {
    return [];
  }

  compiler.plugin('emit', (compilation, callback) => {
    // 遍历所有编译过的资源文件，
    // 对于每个文件名称，都添加一行内容。
    // console.log(compilation.assets)
    let apps = {};

    // eslint-disable-next-line no-restricted-syntax
    for (const filename in compilation.assets) {
      if (regMrocFrontend.test(filename)) {
        const rets = filename.match(regMrocFrontend);
        let id = `/${rets[1]}/${rets[2]}`;

        // 如果有配置自定义key 移除
        if (getConfItem('removeAppjsonPrefix') && id.indexOf(getConfItem('removeAppjsonPrefix')) === 1) {
          id = id.substring(getConfItem('removeAppjsonPrefix').length + 1);
        }

        let app = apps[id];
        if (!app) {
          // eslint-disable-next-line no-multi-assign
          apps[id] = app = {
            id,
            name: id,
            title: '',
            type: 2,
            path: id,
            html: '<div id=\'app\'></div>',
            scripts: {
              entry: '',
              depts: {
                imports: {
                },
              },
            },
            css: [],
          };
        }
        if (filename.endsWith('.js')) {
          if (rets[3].startsWith(rets[2])) {
            app.scripts.entry = `/${filename}`;
          } else {
            app.scripts.depts.imports[rets[3]] = `/${filename}`;
          }
        }
        if (filename.endsWith('.css')) {
          app.css.push(`/${filename}`);
        }
      }
    }

    const appsJsonConfigPath = path.join(getConfItem('assetsRoot'), '/apps.json');

    // merge
    if (fs.existsSync(appsJsonConfigPath)) {
      const oldApps = JSON.parse(fs.readFileSync(appsJsonConfigPath));
      apps = Object.assign(oldApps, apps);
    }

    // 如果独立部署不打包importmap.json
    if (!isSplitDeploy() && getConfItem('buildImportJson')) {
      const importmapJsonConfigPath = path.join(getConfItem('assetsRoot'), '/importmap.json');
      compilation.assets['importmap.json'] = {
        source() {
          let maps = getSystemImportmap(apps);
          if (fs.existsSync(importmapJsonConfigPath)) {
            const oldImportmap = JSON.parse(fs.readFileSync(importmapJsonConfigPath));
            maps = Object.assign(oldImportmap, maps);
          }
          return JSON.stringify(maps);
        },
        size() {
          return 1;
        },
      };
    }
    compilation.assets['apps.json'] = {
      source() {
        return JSON.stringify(apps);
      },
      size() {
        return 1;
      },
    };

    callback();
  });
};
