/*
* 环境变量会注入到process.env
* 注意值为json对象字符串
*/
module.exports = {
  common: {
  },
  dev: {
    TNF_ENV_DEMO: '"test"',
  },
  build: {
  },
};
