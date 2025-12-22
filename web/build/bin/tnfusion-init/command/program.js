/* 命令 */

const commander = require('commander');
// const inquirer = require('inquirer');

/* function actionCallback(fusionObj, projname) {
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
    });
    return inquirer.prompt(promps).then(answers => fusionObj.matchCommand({
      projname: answers.projname,
    }));
  }

  return fusionObj.matchCommand({
    projname,
  });
}
 */
module.exports = function program(fusionObj) {
  /* _init */

  const { option } = fusionObj;
  const typenameArr = [];
  const choices = option.map((v) => {
    typenameArr.push(v.alias);
    return [`${v.alias}`, `${v.desc}`].join('|');
  });
  console.log(`
####################### 初始化项目类型如下 ####################### 

${choices.join(`
`)}

tnfusion-init -t <type> -n <name>
  `);
  commander
    .option('-t --typename <type>', `请输入初始化项目类型${typenameArr.join(',')}`)
    .option('-n --projname <proj>', '请输入项目名称')
    .action(({ typename, projname }) => {
      if (!typenameArr.includes(typename)) {
        console.log('error:', `请输入有效的初始化项目类型${typenameArr.join(',')}`);
        return false;
      }

      if (typeof projname !== 'string' || !projname) {
        console.log('error:', '请输入项目名称');
        return false;
      }
      return fusionObj.matchCommand({
        typename,
        projname,
      });
    });

  // commander.outputHelp(fusionObj.makeRed);

  commander.parse(fusionObj.processArgv);
};
