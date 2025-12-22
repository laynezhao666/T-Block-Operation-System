const fs = require('fs');
const path = require('path');
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

const cmds = fs.readdirSync(path.join(__dirname, '../../'));

const cmd = [];
cmds.map((dirname) => {
  if (!['common', 'tnfusion', 'tnfusion-run'].includes(dirname) && dirname.includes('tnfusion-')) {
    return cmd.push(dirname.replace('tnfusion-', ''));
  }
  return false;
});

const option = [];

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
