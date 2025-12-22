/* 命令 */

const commander = require('commander');
const inquirer = require('inquirer');

module.exports = function program(fusionObj) {
  /* version */
  // commander.version(fusionObj.version, '-v, --version');

  /* dev */
  commander
    .command('dev <projname> [pages...]')
    .description('【套件命令】调试项目或多页面,仅限项目内使用')
    .action((projname, pages) => {
      const promps = [];
      if (!new RegExp('^[a-z]+', 'g').test(projname)) {
        promps.push({
          type: 'input',
          message: '请输一个由英文字母组成的项目名称：',
          name: 'projname',
          filter(val) { // 使用filter将回答变为小写
            return val.replace().toLowerCase()
              .trim();
          },
          validate(val) {
            if (new RegExp('^[a-z]+', 'g').test(val)) {
              return true;
            }
            return '项目名称由英文字母组成';
          },
        }, {
          type: 'input',
          message: '请输一个或多个页面(空格相隔,可不填回车即可)：',
          name: 'pages',
          filter(val) {
            return val.toLowerCase().split(' ')
              .filter(Boolean);
          },
          validate(val) {
            if (new RegExp('^[a-z]+', 'g').test(val) || !val.length) {
              return true;
            }
            return '页面名称由英文字母组成';
          },
        });
        return inquirer.prompt(promps).then(answers => fusionObj.matchCommand({
          cmd: 'dev',
          projname: answers.projname,
          pages: answers.pages,
        }));
      }

      return fusionObj.matchCommand({
        cmd: 'dev',
        projname,
        pages,
      });
    });

  /* build */
  commander
    .command('build <projname> [pages...]')
    .description('【套件命令】打包项目或多页面,仅限项目内使用')
    .action((projname, pages) => {
      const promps = [];
      if (!new RegExp('^[a-z]+', 'g').test(projname)) {
        promps.push({
          type: 'input',
          message: '请输一个由英文字母组成的项目名称：',
          name: 'projname',
          filter(val) { // 使用filter将回答变为小写
            return val.replace().toLowerCase()
              .trim();
          },
          validate(val) {
            if (new RegExp('^[a-z]+', 'g').test(val)) {
              return true;
            }
            return '项目名称由英文字母组成';
          },
        }, {
          type: 'input',
          message: '请输一个或多个页面(空格相隔,可不填回车即可)：',
          name: 'pages',
          filter(val) {
            return val.toLowerCase().split(' ')
              .filter(Boolean);
          },
          validate(val) {
            if (new RegExp('^[a-z]+', 'g').test(val) || !val.length) {
              return true;
            }
            return '页面名称由英文字母组成';
          },
        });
        return inquirer.prompt(promps).then(answers => fusionObj.matchCommand({
          cmd: 'build',
          projname: answers.projname,
          pages: answers.pages,
        }));
      }
      return fusionObj.matchCommand({
        cmd: 'build',
        projname,
        pages,
      });
    });

  /* tnwebup */
  commander
    .command('tnwebup')
    .description('【定制命令】微前端项目自动更新tnwebui,devstyle')
    .action(() => fusionObj.matchCommand({
      cmd: 'tnwebup',
      projname: '',
      pages: [],
    }));

  /* tnwebuiup */
  commander
    .command('tnwebuiup')
    .description('【定制命令】更新static下的tnwebui版本')
    .action(() => fusionObj.matchCommand({
      cmd: 'tnwebuiup',
      projname: '',
      pages: [],
    }));

  /* tnebulaup */
  commander
    .command('tnebulaup')
    .description('【定制命令】更新build tnebula到devstyle')
    .action(() => fusionObj.matchCommand({
      cmd: 'tnebulaup',
      projname: '',
      pages: [],
    }));

  /* tnwebiconup */
  commander
    .command('tnwebiconup')
    .description('【定制命令】更新static下的tnweb-icon的版本')
    .action(() => fusionObj.matchCommand({
      cmd: 'tnwebiconup',
      projname: '',
      pages: [],
    }));

  /* init-tnwebtool */
  commander
    .command('init-tnwebtool')
    .description('【定制命令】外网开发-初始化开发者tnweb工具包')
    .action(() => fusionObj.matchCommand({
      cmd: 'init-tnwebtool',
      projname: '',
      pages: [],
    }));

  /* init-docs */
  commander
    .command('init-docs')
    .description('【定制命令】外网开发-初始化开发文档')
    .action(() => fusionObj.matchCommand({
      cmd: 'init-docs',
      projname: '',
      pages: [],
    }));

  /* init-static */
  commander
    .command('init-static')
    .description('【定制命令】外网开发-初始化static')
    .action(() => fusionObj.matchCommand({
      cmd: 'init-static',
      projname: '',
      pages: [],
    }));

  /* init-tnwebuisite */
  commander
    .command('init-tnwebuisite')
    .description('【定制命令】外网开发-初始化tnwebuisite')
    .action(() => fusionObj.matchCommand({
      cmd: 'init-tnwebuisite',
      projname: '',
      pages: [],
    }));

  /* init-proj */
  commander
    .command('init-proj')
    .description('【定制命令】外网开发-初始化proj')
    .action(() => fusionObj.matchCommand({
      cmd: 'init-proj',
      projname: '',
      pages: [],
    }));

  /* init-page */
  commander
    .command('init-page')
    .description('【定制命令】外网开发-初始化page')
    .action(() => fusionObj.matchCommand({
      cmd: 'init-page',
      projname: '',
      pages: [],
    }));

  /* update-tnebula */
  commander
    .command('update-tnebula')
    .description('【定制命令】新UI开发-打包tnebula更新devstyle资源')
    .action(() => fusionObj.matchCommand({
      cmd: 'update-tnebula',
      projname: '',
      pages: [],
    }));

  /* update-tnwebui */
  commander
    .command('update-tnwebui')
    .description('【定制命令】新UI开发-升级@tencent/TNWeb-ui的5.*.*更新static下tnweb')
    .action(() => fusionObj.matchCommand({
      cmd: 'update-tnwebui',
      projname: '',
      pages: [],
    }));

  commander.outputHelp(fusionObj.makeRed);

  commander.parse(fusionObj.processArgv);
};
