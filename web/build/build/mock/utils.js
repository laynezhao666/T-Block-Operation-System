/* eslint-disable no-param-reassign */
const path = require('path');
const { some, reduce, every, each } = require('lodash');
const { merge } = require('lodash/fp');
const { isMatch: micromatch } = require('micromatch');
const Mock = require('mockjs');
const xlsx = require('node-xlsx');
const chalk = require('chalk');
const glob = require('fast-glob');
const log = require('loglevel');
const { MODEL_BASE } = require('./const');

Mock.Random.extend({
  phone: () => Mock.mock(/^1[385][1-9]\d{8}/),
  temp: () => Mock.mock('@float(20, 50, 0, 2)'), // 温度
  hum: () => Mock.mock('@float(50, 100, 0, 2)'), // 湿度
  rate: () => Mock.mock('@float(0, 0, 0, 4)'), // 比例
});

function sleep(timeout) {
  return new Promise((resolve) => {
    setTimeout(resolve, timeout);
  });
}

function mockArray(array, min = 5, max = 20) {
  const key = 'array';
  return Mock.mock({
    [`${key}|${min}-${max}`]: array,
  })[key];
}

function mockObject(object) {
  return Mock.mock(object);
}

function getRulesFromSyntax(api) {
  const rules = api.split('|');
  let main;
  let range = [];
  const params = {};
  let latency = 0;
  rules.forEach((rule, i) => {
    if (i === 0) {
      main = rule;
    } else if (rule.includes('=')) {
      const pair = rule.split('=');
      // eslint-disable-next-line prefer-destructuring
      params[pair[0]] = pair[1];
      // .3000，latency 语法
    } else if (rule.match(/^\.(?:\d+)$/)) {
      latency = +rule.slice(1);
    } else {
      const rangeMatch = rule.match(/^\d+(?:-\d+)?$/);
      if (rangeMatch) {
        range = rule.split('-').map(Number);
      } else {
        warn(`${api} 存在非法规则 ------ ${rule}`);
      }
    }
  });
  return {
    range,
    params,
    latency,
    main,
  };
}

function getDescriptor(config, pathname, query, method) {
  query = query && query.split('&');
  if (method) {
    if (config[method]) {
      config = config[method];
    } else {
      warn(`rule中没有配置${method}`);
      config = {};
    }
  }
  let rst;
  // 处理 base
  some(config.base, (pathnames, key) => {
    if (pathnames.indexOf(pathname) > -1) {
      const { latency, main: body } = getRulesFromSyntax(key);
      rst = {
        latency,
        body,
      };
      return true;
    }
  });
  // 处理 model
  if (!rst) {
    some(config.model, (model, api) => {
      const { params, range, latency, main: path } = getRulesFromSyntax(api);
      if (path === pathname) {
        // eslint-disable-next-line no-confusing-arrow
        const isMatch = every(params, param => query ? query.includes(param) : false);
        if (range.length === 1 && range[0] !== 0) {
          range.push(range[0]);
        }
        if (isMatch) {
          rst = {
            model,
            range,
            latency,
          };
        }
        return isMatch;
      }
    });
  }
  return rst;
}

function match(str, patterns) {
  if (str.startsWith('/')) {
    str = str.slice(1);
  }
  let mode; // true: black
  return reduce(
    patterns,
    (memo, pattern) => {
      const isWhite = !pattern.startsWith('!');
      if (memo && mode && isWhite) {
        // 当前是通过的，而且是由白名单通过的，后面的白名单全部跳过直到黑名单
        return memo;
      } if (!memo && mode === false && !isWhite) {
        // 当前是禁止的，而且是由黑名单禁止的，后面的黑名单全部跳过直到白名单
        return memo;
      }
      mode = isWhite;
      return micromatch(str, pattern);
    },
    false
  );
}

async function autoResponse(rsp, { code = 200, data, method, type = 'json', latency }, config) {
  rsp.append('X-Data-Origin', 'mock');
  if (latency) {
    await sleep(latency);
  } else if (config.latency) {
    await sleep(config.latency);
  }
  switch (type) {
    case 'json': {
      let formatter;
      if (config.formatter) {
        // eslint-disable-next-line prefer-destructuring
        formatter = config.formatter;
      } else if (config.restful) {
        formatter = data => data;
      } else {
        formatter = (data, code) => ({
          [config.dataField]: data,
          [config.codeField]: code,
        });
      }
      const body = JSON.stringify(formatter(data, code, method));
      rsp.setHeader('Content-Type', 'application/json;charset=UTF-8');
      if (config.restful) {
        rsp.status(code).end(body);
      } else {
        rsp.end(body);
      }
      break;
    }
    case 'xlsx': {
      const body = xlsx.build([{
        name: 'mock',
        data: [
          [1, 2, 3],
          [true, false, null],
          ['foo', 'bar', void 0],
        ],
      }]);
      rsp.setHeader('Content-Type', 'application/octet-stream');
      rsp.end(body);
    }
  }
}

// 这样做有个问题：如果后台接口用了下划线作为cgi路径的名字_，那么就没法正常解析了
// 所以改成^，在url属于非法字符，文件名属于合法字符
function nameToPath(name) {
  return path.join(...(name.split('^')));
}

/**
 * 解析指定目录下的文件, 如 ./model/cgi^go/
 * 根据指定的cgi根路径，如 /cgi/go/
 * 合成对应的cgi请求，/cgi/go/a/b/c 对应文件 ./model/cgi^go/a^b^c.anyname
 *
 * 如果目录下还有目录，如 xxx，yy^zz^oo
 * 则递归生成，并将文件名追加到路径中，如xxx，yy/zz/oo
 *
 * @param {string} folder mock 工作区
 * @param {string} blacklist model 下面不进行转换的路径
 */
function analysis(folder, blacklist = []) {
  const localDir = path.join(path.dirname(folder), MODEL_BASE);

  const files = glob.sync('**/**', {
    cwd: localDir,
  });
  const list = {};
  files.forEach((file) => {
    const filePath = nameToPath(path.join(path.dirname(file), path.basename(file, path.extname(file))));

    // 最终生成key的时候，需要是POSIX
    const cgiPath = path.join('/', filePath).split(path.sep)
      .join('/');
    if (!blacklist.some(glob => micromatch(cgiPath, glob))) {
      list[cgiPath] = require(path.join(localDir, file));
    }
  }, {});
  console.log('\n******** mock ********');
  console.log(Object.keys(list));
  return list;
}

function getRules(path, { autolink }, prefix) {
  let rule;
  try {
    rule = require(path);

    if (autolink) {
      rule = merge(rule)({ model: analysis(path, rule.blacklist) });
    }
  } catch (e) {
    warn('语法错误');
    console.log(e);
  }
  log.info('\n******** mock model list ********');
  each(Object.keys(rule.model || {}), (api) => {
    log.info(chalk.green(`${prefix}${api}`));
  });
  return rule;
}

function warn(msg) {
  log.warn(chalk.yellow(msg));
}

exports.match = match;
exports.mockArray = mockArray;
exports.mockObject = mockObject;
exports.getDescriptor = getDescriptor;
exports.autoResponse = autoResponse;
exports.getRules = getRules;
exports.warn = warn;
