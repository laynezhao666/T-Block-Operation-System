const prompt = require('./prompt');
const program = require('./program');
const exec = require('./exec');
const { downLoadGitRemoteFile } = require('../../common/utils');
const templatePath = downLoadGitRemoteFile('template');
// const templatePath = downLoadGitRemoteFile('template', void (0), 'dev-tnfusion');
const config = require(`${templatePath}/config.js`);
const version = '1.0.2';
const inputConf = {
  cmd: '_init',
  option: '',
  projname: '',
  pages: [],
  templatePath,
};
const cmd = [];

const option = config;

/* eslint-disable no-useless-escape */
const welcomeStr = `
______  _   _  _____                _             
|_   _|| \\ | ||  ___| _   _    __  (_)  ___   _ __  
  | |  |  \\| || |__  | | | | / __| | | / _ \\ | |_ \\ 
  | |  | |\\  ||  __| | |_| | \_\\ \\_ | || (_) || | | |
  |_|  |_| \\_||_|    \\_____| |___/ |_| \\___/ |_| |_|

`;

const command = {
  version,
  welcomeStr,
  inputConf,
  cmd,
  option,
  prompt,
  program,
  exec,
};

module.exports = command;
