const fs = require('fs-extra');
const url = require('url');
const path = require('path');
const chalk = require('chalk');
const chokidar = require('chokidar');
const log = require('loglevel');
const { merge, isArray, debounce } = require('lodash/fp');

const defaultConf = require('./config');
const { CONFIG, RULE, MODEL_BASE, ENTITY_BASE, OTHER_BASE } = require('./const');
const { mockArray, mockObject, getDescriptor, match, autoResponse, getRules, warn } = require('./utils');

log.setDefaultLevel(log.levels.INFO);

function next(req, rsp, next) {
  next();
}

module.exports = function (cwd = path.resolve(), configFile = CONFIG, ruleFile = RULE) {
  const configPath = path.join(cwd, configFile);
  const rulePath = path.join(cwd, ruleFile);
  let config; let rule;
  if (fs.existsSync(configPath)) {
    config = merge(defaultConf)(require(configPath));
  } else {
    warn(`${CONFIG} 不存在，跳过 mock 中间件`);
    return next;
  }
  log.setLevel(log.levels[config.log.toUpperCase()]);
  if (fs.existsSync(rulePath)) {
    rule = require(rulePath);
    const modelPath = path.join(path.dirname(rulePath), MODEL_BASE);
    const entityPath = path.join(path.dirname(rulePath), ENTITY_BASE);
    const otherPath = path.join(path.dirname(rulePath), OTHER_BASE);
    rule = getRules(rulePath, config.mock, config.prefix);
    const onChange = debounce(1000)((filePath) => {
      log.info(chalk.green(`${path.relative(path.resolve(), filePath)}有改动，重新加载...`));
      delete require.cache[rulePath];
      delete require.cache[filePath];
      rule = getRules(rulePath, config.mock, config.prefix);
      if (rule) {
        log.info(chalk.green('重新加载完毕'));
      }
    });
    chokidar.watch([rulePath, modelPath, entityPath, otherPath]).on('change', onChange);
  } else {
    warn(`${RULE} 不存在，跳过 mock 中间件`);
    return next;
  }
  const { mock } = config;
  if (!mock.enable) {
    warn('mock 关闭，跳过 mock 中间件');
    return next;
  }
  return function (req, rsp, next) {
    // 如果 pathname 匹配 ignores 则直接传递
    // eslint-disable-next-line
    const { pathname, query } = url.parse(req.url)
    const { method } = req;

    if (!pathname.startsWith(config.prefix)) {
      log.debug(`${chalk.bgRed(pathname).padEnd(100)}${chalk.red('next(prefix no match)').padStart(30)}`);
      next();
      return;
    }
    const suffixPath = pathname.slice(config.prefix.length);
    if (!match(suffixPath, config.patterns)) {
      log.debug(`${chalk.bgRed(pathname).padEnd(100)}${chalk.red('next(pattern no match)').padStart(30)}`);
      next();
      return;
    }
    // 在 rulerc 中指定的接口
    const descriptor = getDescriptor(rule, suffixPath, query, mock.restful ? method : void 0);

    if (descriptor) {
      if (descriptor.body) {
        const { body, latency } = descriptor;
        if (body === 'xlsx') {
          // xlsx mock
          autoResponse(rsp, { type: body, latency }, mock);
        } else if (!isNaN(body)) {
          // 当规则为数字时 mock
          autoResponse(rsp, { code: +body, method, latency }, mock);
        } else {
          let data;
          try {
            data = JSON.parse(body);
          } catch (e) {
            data = body;
          }
          autoResponse(rsp, { data, method, latency }, mock);
        }
        log.info(`${chalk.bgBlue(pathname).padEnd(100)}${chalk.blue('from mock base').padStart(30)}`);
      } else if (descriptor.model) {
        const { range, model, latency } = descriptor;
        if (isArray(model)) {
          if (range[0] === 0 && range[1] === void 0) {
            autoResponse(rsp, { data: mockObject(model), method, latency }, mock);
          } else {
            autoResponse(rsp, { data: mockArray(model, range[0], range[1]), method, latency }, mock);
          }
        } else {
          autoResponse(rsp, { data: mockObject(model), method, latency }, mock);
        }
        log.info(`${chalk.bgBlue(pathname).padEnd(100)}${chalk.blue('from mock model').padStart(30)}`);
      }
    } else {
      log.debug(`${chalk.bgRed(pathname).padEnd(100)}${chalk.red('next(rule no match)').padStart(30)}`);
      next();
    }
  };
};
