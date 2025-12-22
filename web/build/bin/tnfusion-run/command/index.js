const prompt = require('./prompt');
const program = require('./program');
const exec = require('./exec');

const version = '1.0.2';
const inputConf = {
  cmd: '',
  option: '',
  projname: '',
  pages: [],
};
const cmd = [
  'dev',
  'build',
  'tnwebup',
  'tnwebuiup',
  'tnebulaup',
  'tnwebiconup',
  'init-tnwebtool',
  'init-docs',
  'init-static',
  'init-tnwebuisite',
  'init-proj',
  'init-page',
  'update-tnebula',
  'update-tnwebui',
];

const option = [
];

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
