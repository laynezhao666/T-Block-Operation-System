/* 命令行交互内容 */
module.exports = function prompt(fusionObj) {
  const { option } = fusionObj;
  const choices = option.map(v => [`${v.alias}`, `${v.desc}`].join('|'));

  return [{
    type: 'list',
    message: '请选择要初始化类型：',
    name: 'typename',
    choices,
    filter(val) {
      return val.split('|')[0];
    },
  }, {
    type: 'input',
    message: '请输一个项目名称：',
    name: 'projname',
    filter(val) { // 使用filter将回答变为小写
      return val.replace().toLowerCase()
        .trim();
    },
    // validate(val) {
    //   if (new RegExp('^[a-z]+', 'g').test(val)) {
    //     return true;
    //   }
    //   return '项目名称由小写英文字母组成';
    // },
  },
  //  {
  //   type: 'input',
  //   message: '请输一个或多个页面(空格相隔,可不填回车即可)：',
  //   name: 'pages',
  //   filter(val) {
  //     return val.toLowerCase().split(' ')
  //       .filter(Boolean);
  //   },
  //   validate(val) {
  //     if (new RegExp('^[a-z]+', 'g').test(val) || !val.length) {
  //       return true;
  //     }
  //     return '页面名称由英文字母组成';
  //   },
  //   when(answers) {
  //     return answers.option !== 'init';
  //   },
  // },
  ];
};
