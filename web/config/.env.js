/*
* 环境变量会注入到process.env
* 注意值为json对象字符串
*/
const MODULE_DEPLOY= JSON.stringify({
  itoperation: true,
  costs: true,
  tassets: true,
  idcdb: true,
  hr: true,
  tompage: true,
  resourcechart: true,
  integratedmanagement: true,
  timpage: true,
  safetymanager: true,
  equipment: true,
  taskcenter: true,
  carbonemission: true,
  facilityoperation: true,
  tt: true,
})

module.exports = {
  common: {
    MODULE_DEPLOY,
    SPLIT_DEPLOY_FLAG:JSON.stringify(process.env.SPLIT_DEPLOY_FLAG)
  },
  dev: {
    // TNF_ENV_XXX: '"XXXX"',
  },
  build: {
  },
};
