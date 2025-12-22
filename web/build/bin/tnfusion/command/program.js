/* 命令 运行命令-h,或者直接运行命令会匹配 */

const commander = require('commander');

module.exports = function program(fusionObj) {
  /* version */
  commander
    .version(fusionObj.version, '-v, --version')
    .usage('<command> [params]');

  /* _init cmd */
  fusionObj.cmd.map(command => commander.command(command)
    .allowUnknownOption()
    .description(`command alias:tnfusion-${command}`)
    .action(() => {
      fusionObj.matchCommand({
        cmd: '_init',
        option: command,
      });
    }));

  commander.outputHelp(fusionObj.makeRed);
  commander.parse(fusionObj.processArgv);
};
