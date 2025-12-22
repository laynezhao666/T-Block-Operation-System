const shell = require('shelljs');
const path = require('path');

module.exports = {
  _init: (fusionObj) => {
    const { typename, projname, templatePath } = fusionObj.inputConf;
    const { option } = fusionObj;
    process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';
    const choices = option.filter(v => v.alias === typename);
    const tempPath = path.join(templatePath, choices[0].tplpath);

    shell.exec(`cp -rf ${tempPath} ./${projname}`);
    console.log('初始化成功，所在路径为：', `${process.cwd()}/${projname}`);
    shell.exit(1);
  },
};
