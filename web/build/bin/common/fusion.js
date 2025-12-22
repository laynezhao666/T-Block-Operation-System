const inquirer = require('inquirer');
const _ = require('lodash');
const colors = require('colors');

class Fusion {
  constructor(command) {
    this.processArgv = process.argv;
    this.inputArgv = process.argv.slice(2);
    this.otherArgv = []
    this.npmArgvOriginal = process.env.npm_config_argv && JSON.parse(process.env.npm_config_argv).original || undefined
    if(this.npmArgvOriginal){
      this.otherArgv =this.npmArgvOriginal.filter(v=>v.indexOf('--') === 0)
    }
    this.command = command
    this.inputConf = command.inputConf;
    this.version = command.version;
    this.cmd = command.cmd;
    this.option = command.option;
    this.welcomeStr = command.welcomeStr;
  }

  // 检查输入的参数
  init() {
    if (this.inputArgv.length) {
      return this.setCommand();
    }

    console.log(this.welcomeStr);

    // 提示输入
    const promps = this.command.prompt(this);

    return inquirer.prompt(promps).then(answers => {
      this.matchCommand(answers)
    });
  }

  // 初始化命令
  setCommand() {
    this.command.program(this);
    return true;
  }

  // 匹配命令
  matchCommand(input) {
    this.inputConf = _.assign(this.inputConf, input);

    try {
      if(typeof this.command.exec[this.inputConf.cmd] !== 'function'){
       this.makeRed('命令匹配失败')
      }
      this.command.exec[this.inputConf.cmd](this)
    } catch (error) {
      console.log(error)
    }

    return true;
  }

  makeRed(txt) {
    return colors.red(txt);
  }
}

// const fusion = new Fusion(command);
// fusion.init();

function create(command) {
  return new Fusion(command).init()
}

module.exports = {create}
