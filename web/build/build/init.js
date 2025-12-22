/* eslint-disable no-restricted-syntax */
/* eslint-disable no-undef */
const path = require('path');
const fs = require('fs');
const chalk = require('chalk');
const argvs = require('yargs').argv;

const srcPath = path.resolve(__dirname, '../../');
process.chdir(srcPath);

global.TNF_projRootPath = srcPath;
global.TNF_cliPath = path.resolve(__dirname);

global.TNF_curModuleName = '';
global.TNF_argvs_pages = [];

const moduleRootPath = 'src/module';

checkCurModuleName();
/**
 * 检查输入要打包的模块的名称
 */
function checkCurModuleName() {
  const [curinputproj, ...pages] = argvs._;
  if (!curinputproj) {
    exit('项目名称必填');
  }

  if (pages) {
    global.TNF_argvs_pages = pages;
  }
  // node 执行目录开始
  const pa = fs.readdirSync(moduleRootPath);

  for (const ele of pa) {
    // 匹配一个项目,确认完整的项目名称
    if (ele.includes(curinputproj)) {
      global.TNF_curModuleName = ele;

      // 兼容模板html文件内的变量
      // eslint-disable-next-line no-underscore-dangle
      global._curProjName = ele;
      break;
    }
  }

  if (!global.TNF_curModuleName) {
    exit('项目名称输入有误,模糊匹配失败');
  }

  return true;
}

function exit(msg) {
  console.log(chalk.red(`
    ${msg}

    Build/Dev failed with errors.
    Usage: npm run build <module dir name>  |  npm run dev <module dir name>
    where <module dir name> is Level 1 directory under 'fusion_proj/src/module/'
`));
  process.exit();
}
