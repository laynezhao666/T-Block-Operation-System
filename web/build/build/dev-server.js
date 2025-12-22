const path = require('path');
require('./process-env-argv');
require('./check-versions')();
process.env.NODE_ENV = 'development';
global.TNF_curOperName = 'dev';
require('./init');

const opn = require('opn');
const ora = require('ora');
const cors = require('cors');

const express = require('express');
const webpack = require('webpack');
const proxyMiddleware = require('http-proxy-middleware');
const mockMiddleware = require('./mock');
const { getFormatPath, getConfItem, curModuleConf } = require('./config');
const webpackConfig = require('./webpack.dev.conf');
const webpackHelper = require('./webpack.helper');
const portfinder = require('portfinder');
const _ = require('lodash');

// default port where dev server listens for incoming traffic
const port = process.env.PORT || getConfItem('port');
const host = process.env.Host || getConfItem('host');
// automatically open browser, if not set will be false
const autoOpenBrowser = !!getConfItem('autoOpenBrowser');
// Define HTTP proxies to your custom API backend
// https://github.com/chimurai/http-proxy-middleware

const { proxyTable, rewrites } = curModuleConf[global.TNF_curOperName];

function buildRewrites() {
  // 返回的格式就是rewrite的配置格式
  if (!_.isEmpty(rewrites)) {
    return rewrites;
  }
  // 是否开启路径重写
  if (getConfItem('isOpenPathRewrite')) {
    return webpackHelper.getModuleList().map(mod => ({
      from: new RegExp(`^/${global.TNF_curModuleName}/${mod.pageName}(?!.html)$`),
      to(context) {
        return `${context.parsedUrl.pathname}.html`;
      },
    }));
  }
  return [];
};

console.log('++++++++++++++  rewrites ++++++++++++\r\n', buildRewrites());

const app = express();
app.use(cors());
const compiler = webpack(webpackConfig);
const devMiddleware = require('webpack-dev-middleware')(compiler, {
  publicPath: webpackConfig.output.publicPath,
  quiet: true,
});

const hotMiddleware = require('webpack-hot-middleware')(compiler, {
  log: () => {},
});

// perf：去掉多页面入口的模板reload
// compiler.plugin('compilation', function (compilation) {
//   compilation.plugin('html-webpack-plugin-after-emit', function (data, cb) {
//     hotMiddleware.publish({ action: 'reload' })
//     cb()
//   })
// })

app.use(mockMiddleware(path.join(process.cwd(), 'src/module', global.TNF_curModuleName, 'mock')));
// proxy api requests
Object.keys(proxyTable).forEach((context) => {
  let options = proxyTable[context];
  let path;
  if (context.indexOf(',') > -1) {
    path = context.split(',');
  } else {
    path = context;
  }
  if (typeof options === 'string') {
    options = { target: options };
  }
  app.use(proxyMiddleware(options.filter || path, options));
});

// handle fallback for HTML5 history API
app.use(require('connect-history-api-fallback')({
  rewrites: buildRewrites(),
}));

// serve webpack bundle output
app.use(devMiddleware);

// enable hot-reload and state-preserving
// compilation error display
app.use(hotMiddleware);

// serve pure static assets
const staticPath = path.posix.join(getConfItem('assetsPublicPath'), getConfItem('assetsSubDirectory'));

// 已经将根目录static目录删除
app.use(staticPath, express.static('./static'));

// fo本地开发是所需的静态服务资源
// app.use(express.static('src/module/fopage'))
// app.use(staticPath, express.static('src/module/fopage/static'))

let resolveFn;
let rejectFn;
let server;
const readyPromise = new Promise((resolve, reject) => {
  resolveFn = resolve;
  rejectFn = reject;
});

console.log('> Starting dev server...');
const spinner = ora('Starting dev server...').start();

devMiddleware.waitUntilValid((stats) => {
  if (stats.hasErrors()) {
    console.log(stats.toString('errors-only'));
    process.exit(1);
  }
  portfinder.basePort = port;
  portfinder.getPort((err, port) => {
    if (err) {
      rejectFn(err);
    } else {
      process.env.PORT = port;
      process.env.HOST = host;

      if (getConfItem('AutoMapLocalHost')) {
        // dev环境检查 域名是否有指向
        const hostile = require('hostile');
        hostile.set('127.0.0.1', host, (err) => {
          if (err) {
            console.error(err);
          } else {
            console.log(`set/etc/hosts  :${host} successfully!`);
            startServer(port, host);
          }
        });
      } else {
        startServer(port, host);
      }
      resolveFn();
    }
  });
});

function startServer(port, host) {
  server = app.listen(port, '0.0.0.0', () => {
    // var uri = 'http://' + host + ':' + port
    const uri = `http://${host}:${port}/${getFormatPath('')}${global.TNF_webIndex || getConfItem('index') || 'index.html'}`;
    spinner.stop();
    console.log(`> Listening at ${uri}\n`);

    if (autoOpenBrowser && process.env.NODE_ENV !== 'testing') {
      opn(uri).catch(() => {
        console.log('打开浏览器失败');
      });
    }
    return true;
  });
}

module.exports = {
  ready: readyPromise,
  close: () => {
    server.close();
  },
};
