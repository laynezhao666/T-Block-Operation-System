
// 公共接口
export const commonCgi = {
  // 按参数获取用户信息的接口，用于内外用户选择器
  getSimpleUserList: '/cgi/userinfo/account/all',

  // 文件上传下载
  uploadFile: '/cgi/filestorage/uploadFile',
  downloadFile: '/cgi/filestorage/downloadFile',

  // 图片上传下载
  uploadImage: '/cgi/filestorage/uploadImage',
  downloadImage: '/cgi/filestorage/getImage',
};

const host = '';
export const dataQuality = {
  getTreeInfoMapListByMozuId: `${host}/cgi/dataQua/edge/getTreeInfoMapListByMozuId`,
  getConfigInfoDetailBySubDeviceGid: `${host}/cgi/dataQua/edge/getConfigInfoDetailBySubDeviceGid`,
  getConfigInfoDetailByParentGid: `${host}/cgi/dataQua/edge/getConfigInfoDetailByParentGid`,
  queryCurrentInterfaceQuaByGid: `${host}/cgi/dataQua/edge/queryCurrentInterfaceQuaByGid`,
  queryInterfaceHistoryIndicatorByGidWithIndicatorName: `${host}/cgi/dataQua/edge/queryInterfaceHistoryIndicatorByGidWithIndicatorName`,
  queryCurrentCollectQuaByGid: `${host}/cgi/dataQua/edge/queryCurrentCollectQuaByGid`,
  queryCollectHistoryIndicatorByGidWithIndicatorName: `${host}/cgi/dataQua/edge/queryCollectHistoryIndicatorByGidWithIndicatorName`,
};

setTimeout(() => {
  dataQuery.getBizDeviceLevelTree = window.tnwebServices.customConfigService?.get('EnableDeviceNumberV2') || window.tnwebServices.customConfigService?.get('DeviceNumberV2Mapping')
    ? `${host}/cgi/dataQuery/edge/getBizDeviceLevelTree?type=v2`
    : `${host}/cgi/dataQuery/edge/getBizDeviceLevelTree`;
}, 1000);

export const dataQuery = {
  getCollectDeviceTree: `${host}/cgi/dataQuery/edge/getCollectDeviceTree`,
  getBizDeviceTree: `${host}/cgi/dataQuery/edge/getBizDeviceTree`,
  getBizDeviceLevelTree: window.tnwebServices?.customConfigService?.get('EnableDeviceNumberV2') || window.tnwebServices?.customConfigService?.get('DeviceNumberV2Mapping')
    ? `${host}/cgi/dataQuery/edge/getBizDeviceLevelTree?type=v2`
    : `${host}/cgi/dataQuery/edge/getBizDeviceLevelTree`,
  queryPointDetailInfoWithCurrentValueByGid: `${host}/cgi/dataQuery/edge/queryPointDetailInfoWithCurrentValueByGid`,
  queryHistoryPointInfoByTimeRangeAndPageAndOrder: `${host}/cgi/dataQuery/edge/queryHistoryPointInfoByTimeRangeAndPageAndOrder`,
  exportExcelByCollectorGidWithCurrentData: `${host}/cgi/dataQuery/edge/exportExcelByCollectorGidWithCurrentData`,
  exportHistoryPointInfoByTimeRangeAndPageAndOrder: `${host}/cgi/dataQuery/edge/exportHistoryPointInfoByTimeRangeAndPageAndOrder`,
  getDistinctByFieldName: `${host}/cgi/dataQuery/edge/getDistinctByFieldName`,
  getCurrentBizGidAttrsWithValueByConditions: `${host}/cgi/dataQuery/edge/getCurrentBizGidAttrsWithValueByConditions`,
  exportCurrentBizGidAttrsWithValueByConditions: `${host}/cgi/dataQuery/edge/exportCurrentBizGidAttrsWithValueByConditions`,
  getMatchByFieldNameAndValue: `${host}/cgi/dataQuery/edge/getMatchByFieldNameAndValue`,
  getDistinctFieldByCascade: `${host}/cgi/dataQuery/edge/getDistinctFieldByCascade`,
  getHistoryBizGidAttrValuesByTemplate: `${host}/cgi/dataQuery/edge/getHistoryBizGidAttrValuesByTemplate`,
  getHistoryBizGidAttrValues: `${host}/cgi/dataQuery/edge//getHistoryBizGidAttrValues`,
  insertOrUpdateTemplate: `${host}/cgi/dataQuery/edge/insertOrUpdateTemplate`,
  selectTemplateByCondition: `${host}/cgi/dataQuery/edge/selectTemplateByCondition`,
  deleteTemplateById: `${host}/cgi/dataQuery/edge/deleteTemplateById`,
  exportHistoryBizGidAttrValuesByTemplate: `${host}/cgi/dataQuery/edge/exportHistoryBizGidAttrValuesByTemplate`,
  calByExpressAndTimeRange: `${host}/cgi/dataQuery/edge/calByExpressAndTimeRange`,
  validateDetailByExpress: `${host}/cgi/dataQuery/edge/validateDetailByExpress`,
  diagnosisByExpressAndUpdateTime: `${host}/cgi/dataQuery/edge/diagnosisByExpressAndUpdateTime`,
  exportByExpressAndTimeRange: `${host}/cgi/dataQuery/edge/exportByExpressAndTimeRange`,
  getEdgeDevices: `${host}/cgi/dataQuery/edge/getEdgeDevices`,
  getPointInfo: `${host}/api/dcos/tboxmonitor-cgi/trace/point`,
  getDeviceInfo: `${host}/api/dcos/tboxmonitor-cgi/collector/device/detail`,
  getCollectValues: `${host}/cgi/dataQuery/edge/queryCurrentByMibNameListWithValidDuration`,
};
export const pue = {
  // queryHistoryIndicatorByMibNameWithTimeRange:
  // `${host}/cgi/pointQuery/edge/queryHistoryIndicatorByMibNameWithTimeRange`,
  queryHistoryIndicatorByMibNameWithTimeRange: `${host}/cgi/dataQuery/edge/queryHistoryIndicatorByMibNameWithTimeRange`,
  getLevelPue: `${host}/cgi/dataQuery/edge/getLevelPue`,
  // queryHistoryIndicatorByMibNameWithTimeRange: `${host}/queryHistoryIndicatorByMibNameWithTimeRange`,

};

export const home = {
  getActiveOverview: `${host}/cgi/alarm/active/getActiveOverview`,
  // 获取边端模组
  getEdgeLocation: `${host}/cgi/dataQuery/edge/getEdgeLocation`,
  // 获取告警平台列表
  getActiveOverviewList: `${host}/cgi/alarm/active/getList`,
  // 获取链接状态
  getStatus: `${host}/cgi/alarmha/get`,
  // 查实时指标
  getByMibNameList: `${host}/cgi/dataQuery/edge/queryCurrentByMibNameListWithValidDuration`,
  // 查历史指标
  getTimeRangeList: `${host}/cgi/dataQuery/edge/queryHistoryIndicatorByMibNameWithTimeRange`,

};

export const warning = {
  getActivedWarning: '/cgi/alarm/active/getList',
  getActivedOverview: '/cgi/alarm/active/getActiveOverview',
  getActiveDeviceType: '/cgi/alarm/active/getActiveDeviceType',
  getActiveDeviceList: '/cgi/alarm/active/getActiveDeviceList',
  getWarningDetail: '/cgi/alarm/active/detail',
  exportActivedWarning: '/cgi/activedWarning/exportCgi',
  getRunDataAnalog: '/cgi/activedWarning/rundataAnalog',
  getRunDataStatus: '/cgi/activedWarning/rundataStatus',
  getWarningHistory: '/cgi/alarm/history/getList',
  exportWarningHistory: '/cgi/alarm/history/exportList',
  getStrategyList: '/cgi/alarm/rule/getList',
  exportStrategyList: '/cgi/alarm/rule/exportList',
  getStrategyStatistics: '/cgi/alarm/rule/getStatistics',
  getAlarmType: '/cgi/alarm/active/getAlarmType',
  exportList: '/cgi/alarm/active/exportList',
  getPointDataType: '/cgi/alarm/active/getPointDataType',
  getPointData: '/cgi/alarm/active/getPointData',
  getCltPointsData: '/cgi/alarm/tool/getCltPointsData',
  GetAlarmRuleDropdown: '/cgi/alarm/tool/GetAlarmRuleDropdown',
  getDeviceCltPoints: '/cgi/alarm/tool/getDeviceCltPoints',

  getValidOverview: '/cgi/alarm/validate/realtime/getOverview',
  getValidList: '/cgi/alarm/validate/realtime/getValidList',
  exportValidList: '/cgi/alarm/validate/realtime/exportValidList',
  getInvalidList: '/cgi/alarm/validate/realtime/getInvalidList',
  exportInValidList: '/cgi/alarm/validate/realtime/exportInValidList',
  getValidateDetail: '/cgi/alarm/validate/realtime/getValidateDetail',
  getHistoryList: '/cgi/alarm/validate/history/getHistoryList',
  getHistoryDetail: '/cgi/alarm/validate/history/getHistoryDetail',
  exportInvalidHistory: '/cgi/alarm/validate/history/exportHistory',
  getValidDropdown: '/cgi/alarm/validate/realtime/getDropdown',

  editCustom: '/cgi/alarm/rule/editCustom',
  getProtocolTypeDropdown: '/cgi/alarm/getProtocolTypeDropdown',
  getDeviceListByProtocolType: '/cgi/alarm/getDeviceListByProtocolType',
  queryPointTypeList: '/cgi/alarm/queryPointTypeList',

  getOperationLog: '/cgi/alarm/rule/getOperationLog',
};

export const importCgi = {
  importDictionary: '/cgi/gidmapping/importDictionary',
  importDefaultCollectTemplate: '/cgi/gidmapping/importDefaultCollectTemplate',
  importTboxDriverFiles: '/cgi/gidmapping/importTboxDriverFiles',
  importCollectorDevices: '/cgi/gidmapping/importCollectorDevices',
  importStandard2CollectMapping: '/cgi/gidmapping/importStandard2CollectMapping',
  importBizToStdMapping: '/cgi/gidmapping/importBizToStdMapping',
  refreshQYDeviceGid: '/cgi/gidmapping/refreshQYDeviceGid',
  warningImport: '/cgi/alarm/rule/import',
};

export const ba = {
  getHoursAlarmFixRate: '/cgi/alarm/ba/getHoursAlarmFixRate',
  getHoursAlarmTrend: '/cgi/alarm/ba/getHoursAlarmTrend',
  getHoursAlarmTypeTopCount: '/cgi/alarm/ba/getHoursAlarmTypeTopCount',
};

const websocketProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';

let wshost = `${websocketProtocol}://${location.host}`;
let alarmHost = wshost;
let haHost = wshost;
export const wsScreen = {
  screenBizIndicator: `${wshost}/ws/screenBizIndicator`,
  quaScreenIndicator: `${wshost}/ws/quaScreenIndicator`,
  alarm: `${alarmHost}/ws/alarm`,
  ha: `${haHost}/ws/HA`,
};

const devicePointComparePrefix = '/cgi/dashboardaux/';
export const devicePointCompare = {
  getQueryTemplateByCondition: `${devicePointComparePrefix}getQueryTemplateByCondition`,
  addQueryTemplate: `${devicePointComparePrefix}addQueryTemplate`,
  getAllQueryTemplate: `${devicePointComparePrefix}getAllQueryTemplate`,
  addTemplateTag: `${devicePointComparePrefix}addTemplateTag`,
  getAllTemplateTag: `${devicePointComparePrefix}getAllTemplateTag`,
  getTemplateTagById: `${devicePointComparePrefix}getTemplateTagById`,
  getQueryTemplateById: `${devicePointComparePrefix}getQueryTemplateById`,
  updateQueryTemplate: `${devicePointComparePrefix}updateQueryTemplate`,
  deleteQueryTemplate: `${devicePointComparePrefix}deleteQueryTemplate`,
  updateTemplateTag: `${devicePointComparePrefix}updateTemplateTag`,
  deleteTemplateTag: `${devicePointComparePrefix}deleteTemplateTag`,
  exportHistoryBizGidAttrValuesByTemplate: '/cgi/dataQuery/edge/exportHistoryBizGidAttrValuesByTemplate',
};
// 采集器相关接口
export const collectorApi = {
  queryCollectorAssignedList: '/api/dcos/tboxmonitor-cgi/collector/all', // 采集器及设备列表（已分配）
  queryCollectorUnassignedList: '/api/dcos/tboxmonitor-cgi/unassigned_collector/all', // 采集器及设备列表（未分配）
  queryCollectorDetail: '/api/dcos/tboxmonitor-cgi/collector/detail', // 采集器详情
  queryDeviceDetail: '/api/dcos/tboxmonitor-cgi/collector/device/detail', // 设备详情
  restartOS: '/api/dcos/tboxmonitor-cgi/control/restart/os', // 重启主机
  restartCollector: '/api/dcos/tboxmonitor-cgi/control/restart/collector', // 重启服务
  queryPointData: '/api/dcos/tboxmonitor-cgi/collector/points/rtd', // 测点实时数据
  queryPointInfo: '/api/dcos/tboxmonitor-cgi/collector/points', // 设备（采集器）所有测点
  editCollector: '/api/dcos/tboxmonitor-cgi/collector/assign', // 分配采集器
  getRoomConfig: '/api/dcos/tboxmonitor-cgi/rooms', // 获取房间和方仓列表
  resetStatus: '/api/dcos/tboxmonitor-cgi/collector/discovered', // 重置采集器新发现状态
  resetDeviceStatus: '/api/dcos/tboxmonitor-cgi/collector/device/discovered',
  getCollectorBackupList: '/api/dcos/tboxmonitor-cgi/collector/backups', // 获取采集器备份列表
  createCollectorBackup: '/api/dcos/tboxmonitor-cgi/collector/backup', // 创建采集器备份
  downloadBackup: '/api/dcos/tboxmonitor-cgi/collector/backup', // 下载采集器备份
  deleteBackup: '/api/dcos/tboxmonitor-cgi/collector/backup', // 删除采集器备份
  restoreBackup: '/api/dcos/tboxmonitor-cgi/collector/restore', // 还原备份到原采集器
  queryCollectorAssignedListNew: '/api/dcos/tboxmonitor-cgi/collector/all/v2', // 采集器及设备列表V2（已分配）
  getAllBackups: '/api/dcos/tboxmonitor-cgi/collector/backups', // 获取所有采集器的备份列表
  getAllBackups2: '/api/dcos/tboxmonitor-cgi/backups', // 获取所有采集器的备份列表
  restoreBackup2: '/api/dcos/tboxmonitor-cgi/restore', // 还原已分配采集器的备份数据到未分配采集器
};

export const tbosWarning = {
  GetStrategy: '/cgi/idc-tbos-cgi/alarm/server/GetStrategy',
  GetAlarmList: '/cgi/idc-tbos-cgi/alarm/server/GetAlarmList',
  pointQuery: '/cgi/idc-tbos-cgi/Data/Query',
  GetValidate: '/cgi/idc-tbos-cgi/alarm/server/GetValidate'
};

export const tbosCollectorApi = {
  GetCollectorStatusTree: '/cgi/idc-tbos-cgi/Cmdb/GetCollectorStatusTree',
  GetCollectorInfo: '/cgi/idc-tbos-cgi/Cmdb/GetCollectorInfo',
  GetCollectorPoint: '/cgi/idc-tbos-cgi/Cmdb/GetCollectorPoint',
  QueryLatest: '/cgi/idc-tbos-cgi/Data/QueryLatest',
}

export const tbosImportCgi = {
  listMozu: '/cgi/idc-tbos-cmdb/ConfigBuild/ListMozu',
  saveMozu: '/cgi/idc-tbos-cmdb/ConfigBuild/SaveMozu',
  importModel: '/cgi/idc-tbos-cmdb/ConfigBuild/ImportModel',
  deleteMozu: '/cgi/idc-tbos-cmdb/ConfigBuild/DeleteMozu',
  getMozu: "/cgi/idc-public-privilege/getMozu"
}