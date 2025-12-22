#! /usr/bin/env node

// const fs = require('fs');
// const path = require('path');

const program = require('commander');
const inquirer = require('inquirer');
const _ = require('lodash');
const shell = require('shelljs');
const colors = require('colors');

class Fusion {
  constructor(processArgv) {
    this.processArgv = processArgv;
    this.inputArgv = processArgv.slice(2);
    this.inputConf = {
      cmd: '',
      option: '',
      projname: '',
      pages: [],
    };

    this.version = '0.0.1';
    this.cmd = [
      // 'init',
      'dev',
      'build',
      // 'tnwebup',
    ];
    this.option = [
      // 'single',
      // 'multi',
      // 'mobile',
    ];
    /* eslint-disable no-useless-escape */
    this.welcomeStr = `
    ______  _   _  _____                _             
    |_   _|| \\ | ||  ___| _   _    __  (_)  ___   _ __  
      | |  |  \\| || |__  | | | | / __| | | / _ \\ | |_ \\ 
      | |  | |\\  ||  __| | |_| | \_\\ \\_ | || (_) || | | |
      |_|  |_| \\_||_|    \\_____| |___/ |_| \\___/ |_| |_|
    
    `;
  }

  // 检查输入的参数
  init() {
    console.log(this.welcomeStr);

    const fusionObj = this;

    if (this.inputArgv.length) {
      return this.setCommand();
    }

    // 提示输入
    const promps = [{
      type: 'list',
      message: '请输选择一个命令：',
      name: 'cmd',
      choices: this.cmd,
    }, {
      type: 'list',
      message: '请选择要初始化的项目：',
      name: 'option',
      choices: this.option,
      when(answers) {
        return answers.cmd === 'init';
      },
    }, {
      type: 'input',
      message: '请输一个项目名称：',
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
      when(answers) {
        return answers.cmd !== 'init';
      },
    }];

    /*
    { cmd: 'init',
      option: 'single',
      projname: 'qwefqwef',
      pages: [ 'sss', 'dfsf', 'sdfasdf' ] }
    */
    return inquirer.prompt(promps).then(answers => fusionObj.matchCommand(answers));
  }

  // 初始化命令
  setCommand() {
    const fusionObj = this;

    program
      .version(this.version, '-v, --version');

    // program
    //   .command('tnwebup')
    //   .description('更新tnwebui,devstyle更新')
    //   .action(() => fusionObj.matchCommand({
    //     cmd: 'tnwebup',
    //     projname: '',
    //     pages: [],
    //   }));

    /*
    program
      .command('init  <projname>')
      .description('创建项目,[option] 选项请输入命令$ fusion i -h 查看')
      .option('-s, --single ', '单页面项目')
      .option('-m, --multi', '多页面项目')
      .option('-mo, --mobile', '移动端项目')
      .action((projname, cmd) => {
        let promps = []

        if (!new RegExp('^[a-z]+', 'g').test(projname)) {
          promps.push({
            type: 'input',
            message: '请输一个由英文字母组成的项目名称：',
            name: 'projname',
            filter: function (val) { // 使用filter将回答变为小写
              return val.replace().toLowerCase().trim()
            },
            validate: function (val) {
              if (new RegExp('^[a-z]+', 'g').test(val)) {
                return true
              } else {
                return '项目名称由英文字母组成'
              }
            },
          })
        }

        if (!cmd.single && !cmd.multi && !cmd.mobile) {
          promps.push({
            type: 'list',
            message: '请选择要初始化的项目：',
            name: 'option',
            choices: this.option,
          })
        }

        if (promps.length) {
          return inquirer.prompt(promps).then(function (answers) {
            return fusionObj.matchCommand({
              cmd: 'init',
              option: answers.option,
              projname,
            })
          })
        }
        return fusionObj.matchCommand({
          cmd: 'init',
          projname,
          option: cmd.single ? 'single' : 'multi',
        })
      })

       */
    program
      .command('build <projname> [pages...]')
      .description('打包一个项目或一个项目的多个页面')
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

    program
      .command('dev <projname> [pages...]')
      .description('启动一个项目的开发环境，或一个项目的多个页面开发环境')
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

    program.outputHelp(this.makeRed);

    program.parse(this.processArgv);

    return true;
  }

  // 匹配命令
  matchCommand(input) {
    this.inputConf = _.assign(this.inputConf, input);
    let cli; let execCommand; let tnwebup;
    // let projPath; let tempPath; let from; let to;

    switch (this.inputConf.cmd) {
      case 'tnwebup':

        shell.cd(`${__dirname}/../`);
        shell.exec('pwd');

        tnwebup = ' npm upgrade @tencent/TNWeb-ui &&   npm run build tnebula  ';
        console.log(`============执行命令 ${tnwebup} ===========`);
        if (shell.exec(tnwebup).code !== 0) {
          shell.echo(`Error: exec " ${tnwebup} " failed`);
          shell.exit(1);

          return false;
        }

        shell.exec('rm -rf src/static/devstyle/*');
        shell.exec('rm -rf src/static/thirdparty/tnwebui/*');
        shell.exec('cp dist/main/static/js-css/index/index*.js  src/static/devstyle/index.js');
        shell.exec('cp dist/main/static/js-css/index/index*.css  src/static/devstyle/index.css');
        // eslint-disable-next-line no-useless-escape
        shell.exec('cp -r   node_modules/\@tencent/TNWeb-ui/lib/theme-chalk  src/static/thirdparty/tnwebui/');
        // eslint-disable-next-line no-useless-escape
        shell.exec('cp -r   node_modules/\@tencent/TNWeb-ui/lib/index.js  src/static/thirdparty/tnwebui/');
        shell.exec('git status');

        break;
        /*
      case 'init':
        // 检查项目是否存在
        // projPath = path.resolve(__dirname, '..', projRootPath, this.inputConf.projname);
        if (fs.existsSync(projPath)) {
          console.log(this.makeRed(`该项目已经存在${projPath}`));
          return false;
        }

        // 检查模板项目是否存在
        tempPath = path.resolve(__dirname, '..', templateMap[`${this.inputConf.option}`]);
        if (!fs.existsSync(tempPath)) {
          console.log(this.makeRed(`选择模板不存在${tempPath}`));
          return false;
        }
        from = `${templateMap[`${this.inputConf.option}`]}/`;
        // to = `${projRootPath}/${this.inputConf.projname}`;

        // 支持用户在其他文件夹操作
        shell.cd(__dirname);
        shell.cd('..');

        // 拷贝模板项目
        if (shell.cp('-R', from, to).code !== 0) {
          shell.echo(`Error: copy  ${from}  -->  ${to} failed`);
          shell.exit(1);
          return false;
        }
        break;
 */
      case 'build':
        cli = 'node build.js';
        execCommand = `${cli} ${this.inputConf.projname} ${this.inputConf.pages.join(' ')}`;

        console.log(`============执行命令 ${execCommand} ===========`);
        shell.cd(__dirname);
        if (shell.exec(execCommand).code !== 0) {
          shell.echo(`Error: exec " ${execCommand} " failed`);
          shell.exit(1);

          return false;
        }
        break;

      case 'dev':
        cli = 'node --max-old-space-size=8192 fork.js';
        // cli = 'node dev-server.js'
        execCommand = `${cli} ${this.inputConf.projname} ${this.inputConf.pages.join(' ')}`;

        console.log(`============执行命令 ${execCommand} ===========`);
        shell.cd(__dirname);
        if (shell.exec(execCommand).code !== 0) {
          shell.echo(`Error: exec " ${execCommand} " failed`);
          shell.exit(1);

          return false;
        }
        break;

      default:
        console.log('命令参数错误');
        break;
    }

    return true;
  }

  makeRed(txt) {
    return colors.red(txt);
  }
}
const fusion = new Fusion(process.argv);
fusion.init();
