'use strict';

global.cliArgv = [];
// 获取 npm 命令 --输入的参数
if (process.env.cliArgv) {
  global.cliArgv = process.env.cliArgv.split(',');
}

// 独立拆分部署标识
exports.isSplitDeploy = function () {
  let flag = false;
  global.cliArgv.map((item) => {
    if (item.includes('--split_deploy')) {
      flag = true;
    }
  });
  return flag;
};
