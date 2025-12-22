/* 命令行交互内容 */
module.exports = function prompt(fusionObj) {
  return [{
    type: 'list',
    message: '请输选择一个命令：',
    name: 'cmd',
    choices: fusionObj.cmd,
  }, {
    type: 'list',
    message: '请选择要初始化的项目：',
    name: 'option',
    choices: fusionObj.option,
    // when(answers) {
    //   return answers.cmd === 'init';
    // },
    when(answers) {
      return !['init', 'tnwebup', 'tnwebuiup', 'tnebulaup', 'tnwebiconup'].includes(answers.cmd);
    },
  }, {
    type: 'input',
    message: '请输一个项目名称：',
    name: 'projname',
    filter(val) { // 使用filter将回答变为小写
      return val.replace().toLowerCase()
        .trim();
    },
    when(answers) {
      return !['init', 'tnwebup', 'tnwebuiup', 'tnebulaup', 'tnwebiconup'].includes(answers.cmd);
    },
    // validate(val) {
    //   if (new RegExp('^[a-z]+', 'g').test(val)) {
    //     return true;
    //   }
    //   return '项目名称由英文字母组成';
    // },
  }, {
    type: 'input',
    message: '请输一个或多个页面(空格相隔,可不填回车即可)：',
    name: 'pages',
    filter(val) {
      return val.toLowerCase().split(' ')
        .filter(Boolean);
    },
    // validate(val) {
    //   if (new RegExp('^[a-z]+', 'g').test(val) || !val.length) {
    //     return true;
    //   }
    //   return '页面名称由英文字母组成';
    // },
    when(answers) {
      return !['init', 'tnwebup', 'tnwebuiup', 'tnebulaup', 'tnwebiconup',
        'init-tnwebtool',
        'init-docs',
        'init-static',
        'init-tnwebuisite',
        'init-proj',
        'init-page',
        'update-tnebula',
        'update-tnwebui',
      ].includes(answers.cmd);
    },
  }];
};
