/* 匹配执行 */
// const _ = require('lodash');
const shell = require('shelljs');

module.exports = {
  _init: (fusionObj) => {
    const { option } = fusionObj.inputConf;
    const { inputArgv } = fusionObj;

    let command = '';
    if (inputArgv.length > 1) {
      inputArgv.shift();
      command = `tnfusion-${fusionObj.inputConf.option} ${inputArgv.join(' ')}`;
    } else {
      command = `tnfusion-${fusionObj.inputConf.option} -h`;
    }
    console.log(`
    
############### tnfusion ${option}  ==> alias:tnfusion-${option} ###############

正在执行命令：${command}    
    `);
    shell.exec(command);
  },
};
