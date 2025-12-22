const os = require('os');
const shell = require('shelljs');
const chalk = require('chalk');
const path = require('path');
const ora = require('ora');
const spinner = ora();

const defaultRemoteOriginUrl = ''
const tmpDir = os.tmpdir()

/**
 * 获取git仓库配置
 * @param {*} pullFileName 拉取文件或文件夹名字，例如 template/config.js  或 template
 * @param {*} gitOriginUrl 仓库地址，默认 
 * @param {*} branchName 分支，默认 master
 */
function downLoadGitRemoteFile(pullFileName, gitOriginUrl = defaultRemoteOriginUrl, branchName = 'master', loadingText = '初始化配置文件...\n') {
  spinner.start(loadingText);

  const fusionTmpDir = path.join(tmpDir, `tmp_fusion_config_${+new Date()}`);
  const cwd = process.cwd()
  shell.exec(`rm -rf ${tmpDir}/tmp_fusion_config_*`);
  shell.mkdir(fusionTmpDir);
  shell.cd(fusionTmpDir);
  shell.exec(`git init`);
  shell.exec('git config core.sparsecheckout true');
  shell.exec(`echo ${pullFileName} >> .git/info/sparse-checkout`);
  shell.exec(`git remote add origin ${gitOriginUrl}`);
  const rtn = shell.exec(`git pull origin ${branchName} `);

  shell.cd(cwd);
  spinner.stop();
  if (rtn.code !== 0) {
    console.log(chalk.red('Error: Clone  failed '));
    shell.exit(1);
  }

  // 返回下载的文件夹路径
  return path.join(fusionTmpDir, pullFileName)
}

module.exports = {
  downLoadGitRemoteFile
}