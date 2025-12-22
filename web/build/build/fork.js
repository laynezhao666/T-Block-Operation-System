const path = require('path');
const { fork } = require('child_process');
const chokidar = require('chokidar');
// const { spread } = require('lodash');

// console.log(process.env.TNF_ENV);

const [,, moduleName] = process.argv;
require('./init');

let child;

serve();

const pages = `src/module/*${moduleName}*/pages/*/main.js`;

chokidar.watch(pages, {
  ignoreInitial: true,
  cwd: path.resolve(),
}).on('add', () => {
  child.kill('SIGINT');
  serve();
});

function serve() {
  child = fork(require.resolve('./dev-server'), process.argv.slice(2), {
    stdio: 'inherit',
  });
}
