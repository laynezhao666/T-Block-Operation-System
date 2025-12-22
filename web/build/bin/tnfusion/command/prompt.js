/* 输入主命令提示 */
module.exports = function prompt(fusionObj) {
  return [{
    type: 'list',
    message: '请输选择一个命令：',
    name: 'option',
    choices: fusionObj.cmd,
    filter: (val) => {
      const fo = fusionObj;
      fo.inputConf.cmd = '_init';
      return val;
    },
  },
  ];
};
