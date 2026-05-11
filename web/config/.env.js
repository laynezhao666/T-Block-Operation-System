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
    SPLIT_DEPLOY_FLAG:JSON.stringify(process.env.SPLIT_DEPLOY_FLAG),
    // NFC 读卡器密钥配置（由部署环境注入，此处为默认值）
    NFC_MF_OLD_KEY: JSON.stringify(process.env.NFC_MF_OLD_KEY || 'FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF'),
    NFC_MF_NEW_KEY: JSON.stringify(process.env.NFC_MF_NEW_KEY || 'FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF'),
    NFC_DF_KEY: JSON.stringify(process.env.NFC_DF_KEY || 'FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF'),
  },
  dev: {
    // TNF_ENV_XXX: '"XXXX"',
  },
  build: {
  },
};
