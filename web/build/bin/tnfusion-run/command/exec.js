/* 匹配执行 */
// const _ = require('lodash');
const shell = require('shelljs');
const chalk = require('chalk');

// 检查tnwebtool工具包
function checkTnwebTool(cwd) {
  if (!shell.test('-d', `${cwd}/../____tnwebtool`)) {
    console.log(chalk.red(`
    
    您未安装工具包
  
    `));
    shell.exit();
  }
}

module.exports = {
  build: (fusionObj) => {
    // eslint-disable-next-line no-unused-vars
    let otherArgv = '';
    let maxOldSpaceSize = false;
    if (fusionObj.otherArgv.length) {
      // eslint-disable-next-line no-unused-vars
      fusionObj.otherArgv.map((item, i) => {
        if (item.includes('--max-old-space-size')) {
          maxOldSpaceSize = true;
        }
      });
      otherArgv = `${fusionObj.otherArgv.join(' ')} ${!maxOldSpaceSize ? ' --max-old-space-size=8192' : ''}`;
    }
    console.log('fusionObj.otherArgv', fusionObj.otherArgv);
    process.env.cliArgv = fusionObj.otherArgv;
    // process.exit(1)
    // const command = `node  build.js ${fusionObj.inputConf.projname} ${fusionObj.inputConf.pages.join(' ')}  `;
    const command = `node --max-old-space-size=8192 build.js ${fusionObj.inputConf.projname} ${fusionObj.inputConf.pages.join(' ')}`;

    console.log(`============执行命令 ${command} ===========`);
    shell.cd(`${__dirname}/../../../build`);
    shell.exec(command);
  },
  dev: (fusionObj) => {
    const command = `node --max-old-space-size=8192 fork.js ${fusionObj.inputConf.projname} ${fusionObj.inputConf.pages.join(' ')}`;
    console.log(`============执行命令 ${command} ===========`);
    shell.cd(`${__dirname}/../../../build`);
    shell.exec(command);
  },
  tnwebup: () => {
    const command = ' npm upgrade @tencent/TNWeb-ui &&   npm run build tnebula  ';
    console.log(`============执行命令 ${command} ===========`);
    shell.cd(`${__dirname}/../../../../../../`);
    shell.exec(command);
    shell.exec('rm -rf src/static/devstyle/index*');
    shell.exec('rm -rf src/static/thirdparty/tnwebui/*');
    shell.exec('cp dist/main/static/js-css/index/index*.js  src/static/devstyle/index.js');
    shell.exec('cp dist/main/static/js-css/index/index*.css  src/static/devstyle/index.css');

    // eslint-disable-next-line no-useless-escape
    shell.exec('cp -r node_modules/\@tencent/TNWeb-ui/lib/theme-chalk  src/static/thirdparty/tnwebui/');
    // eslint-disable-next-line no-useless-escape
    shell.exec('cp -r node_modules/\@tencent/TNWeb-ui/lib/index.js  src/static/thirdparty/tnwebui/');
    shell.exec('git status');
  },
  tnwebuiup: () => {
    const command = ' npm upgrade @tencent/TNWeb-ui ';
    console.log(`============执行命令 ${command} ===========`);
    shell.cd(`${__dirname}/../../../../../../`);
    shell.exec(command);
    shell.exec('rm -rf src/static/thirdparty/tnwebui/*');
    // eslint-disable-next-line no-useless-escape
    shell.exec('cp -r node_modules/\@tencent/TNWeb-ui/lib/theme-chalk  src/static/thirdparty/tnwebui/');
    // eslint-disable-next-line no-useless-escape
    shell.exec('cp -r node_modules/\@tencent/TNWeb-ui/lib/index.js  src/static/thirdparty/tnwebui/');
    shell.exec('git status');
  },
  tnebulaup: () => {
    const command = ' npm run build tnebula  ';
    console.log(`============执行命令 ${command} ===========`);
    shell.cd(`${__dirname}/../../../../../../`);
    shell.exec(command);
    shell.exec('rm -rf src/static/devstyle/index*');
    shell.exec('cp dist/main/static/js-css/index/index*.js  src/static/devstyle/index.js');
    shell.exec('cp dist/main/static/js-css/index/index*.css  src/static/devstyle/index.css');

    shell.exec('git status');
  },
  tnwebiconup: () => {
    const command = ' npm upgrade  @tencent/tnweb-icon ';
    console.log(`============执行命令 ${command} ===========`);
    shell.cd(`${__dirname}/../../../../../../`);
    shell.exec(command);
    shell.exec('rm -rf src/static/thirdparty/tnweb-icon/*');
    // eslint-disable-next-line no-useless-escape
    shell.exec('cp -r node_modules/\@tencent/tnweb-icon/lib/theme-chalk  src/static/thirdparty/tnweb-icon/');
    // eslint-disable-next-line no-useless-escape
    shell.exec('cp -r node_modules/\@tencent/tnweb-icon/lib/index.js  src/static/thirdparty/tnweb-icon/');
    shell.exec('git status');
  },

  // 合作伙伴开发工具包命令行工具
  'init-tnwebtool': () => {
    const cwd = process.cwd();
    const workspace = `${cwd}`;
    shell.cd(workspace);
    shell.exec('rm -rf ../____tnwebtool');
    shell.exit();
  },
  'init-docs': () => {
    const cwd = process.cwd();
    checkTnwebTool(cwd);
    const workspace = `${cwd}/../____tnwebtool/tnebula_shell`;
    shell.cd(workspace);
    shell.exec('pwd');
    shell.exec('npm run init-docs');
    shell.exit();
  },
  'init-static': () => {
    const cwd = process.cwd();
    checkTnwebTool(cwd);
    const workspace = `${cwd}/../____tnwebtool/tnebula_shell`;
    shell.cd(workspace);
    shell.exec('pwd');
    shell.exec(`npm run init-static  -- ${cwd}`);
    shell.exit();
  },
  'init-tnwebuisite': () => {
    const cwd = process.cwd();
    checkTnwebTool(cwd);
    const workspace = `${cwd}/../____tnwebtool/tnebula_shell`;
    shell.cd(workspace);
    shell.exec('pwd');
    shell.exec('npm run init-tnwebuisite');
    shell.exit();
  },
  'init-proj': () => {
    console.log('建设中...');
  },
  'init-page': () => {
    console.log('建设中...');
  },

  'update-tnebula': () => {
    const command = ' npm run build tnebula  ';
    console.log(`============执行命令 ${command} ===========`);
    const cwd = process.cwd();
    const workspace = `${cwd}`;
    shell.cd(workspace);

    shell.exec(command);
    shell.exec('rm -rf src/static/devstyle/index*');
    shell.exec('cp dist/main/static/js-css/index/index*.js  src/static/devstyle/index.js');
    shell.exec('cp dist/main/static/js-css/index/index*.css  src/static/devstyle/index.css');
    shell.exec('git status');
    shell.exit();
  },
  'update-tnwebui': () => {
    const command = ' npm upgrade @tencent/TNWeb-ui@tdesign';
    console.log(`============执行命令 ${command} ===========`);
    const cwd = process.cwd();
    const workspace = `${cwd}`;
    shell.cd(workspace);

    shell.exec(command);
    shell.exec('rm -rf src/static/thirdparty/tnwebui/*');
    // eslint-disable-next-line no-useless-escape
    shell.exec('cp -r node_modules/\@tencent/TNWeb-ui/lib/theme-chalk  src/static/thirdparty/tnwebui/');
    // eslint-disable-next-line no-useless-escape
    shell.exec('cp -r node_modules/\@tencent/TNWeb-ui/lib/index.js  src/static/thirdparty/tnwebui/');
    shell.exec('git status');
    shell.exit();
  },

};
