import { ENV_NAME } from 'common/script/passport_login';
import localforage from 'localforage';
export default {
  getModules: '/getModules',
};

export const bp = {
  projects: '/cgi/zt/tw/ZTapi/v1.0/projects',
  blueprints: '/cgi/zt/tw/ZTapi/v1.0/blueprints',
  lantuBlueprints: '/cgi/idc-deliver-tlink-api/v1/polaris/curd/genidc/blueprints/list',
  // todo
  mozuInfo: '/cgi/zt/tw/ZTapi/v1.0/cmdb/mozus',
  campusInfo: '/cgi/zt/tw/ZTapi/v1.0/cmdb/campus?campus_id=608',
  buildingInfo: '/cmdb/get/CBFTree?campusId=608',

  getTopolCharge: '/cgi/idcdbtopol/getTopolCharge', // 带电状态
  getTopolAlarm: '/cgi/idcdbtopol/getTopolAlarm', // 告警状态
  getDeviceInfo: '/cgi/idcdbtopol/getDeviceInfo', // 获取设备信息

  getEventManageAlarmStatus: '/cgi/eventmanage/alarm/status', // 获取告警转事件接口

  getAlarmList: '/cgi/alarm/active/getList',
  // 收敛告警数量
  getAlarmNumByMozuId: '/cgi/alarmconv/getAlarmNumByMozuId',
  // 收敛告警相关信息
  getAlarmInfoByMozuId: '/cgi/alarmconv/getAlarmInfoByMozuId',

  // 图相关
  getGraphTree: '/cgi/alarmconv/getGraphTree',
  getGraphIdByDeviceNum: '/cgi/alarmconv/getGraphIdByDeviceNum',
  // 边端
  getGraphTreeEdge: '/cgi/singlegraph/getGraphTree',
  getAlarmNumByMozuIdEdge: '/cgi/singlegraph/getAlarmNumByMozuId',
  getAlarmInfoByMozuIdEdge: '/cgi/singlegraph/getAlarmInfoByMozuId',
};

export const alarm = {
  getALarmList: '/cgi/alarm/active/getList',
};

export const electric = {
  getGraphTree: '/cgi/singlegraph/getGraphTreeByConfig',
};

let pre = '/cgi/forwardEdge';
if (['publish', 'pre'].includes(ENV_NAME)) {
  pre = '/cgi/forwardEdge';
}

// 暖通视图
export const bigData = {
  // 根据gid和属性值获取设备测点的值 批量
  getExoressValByGidAndAttrListMap: '/cgi/forwardEdge/cgi/expCompute/edge/getExoressValByGidAndAttrListMap',
  getExoressValByGidAndAttrListMapWithoutCache: '/cgi/forwardEdge/cgi/expCompute/edge/getExoressValByGidAndAttrListMapWithoutCache',

  getExoressValByGidAndAttrListMapEdge: '/cgi/expCompute/edge/getExoressValByGidAndAttrListMap',
  getExoressValByGidAndAttrListMapWithoutCacheEdge: '/cgi/expCompute/edge/getExoressValByGidAndAttrListMapWithoutCache',

  // 获取单线图带电状态
  // getTopolchargedQuery: '/cgi/topolcharged/query',
  getTopolchargedQuery: `${pre}/cgi/topolcharged/query`,
  getTopolchargedQueryEdge: '/cgi/topolcharged/query',

  // 根据设备编号拉取idcdb的设备
  getCmdbDeviceInfo: '/cgi/go/cmdb/get',
  getCmdbDeviceInfoEdge: '/cgi/dataQuery/edge/getEdgeDevices',

  // 获取设备告警
  getDeviceAlarmList: `${pre}/cgi/alarm/active/get`,
  getDeviceAlarmListEdge: '/cgi/alarm/active/get',
  // 获取制冷率历史数据
  getColdrateHistory: '/cgi/dataQuery/edge/queryHistoryIndicatorWithExp',
  // 导出
  exportData: '/cgi/dataQuery/edge/queryHistoryIndicatorExportExcel',
};

export const debugApi = {
  render: '',
  point: '',
};

// 获取缓存数据
export async function getDataCache(fn, key) {
  const item = await localforage.getItem(key);

  // 无缓存或缓存过期
  if (!item || item?.expireTime < +new Date()) {
    return setDataCache(fn, key);
  }

  // 异步发起一个更新请求
  setDataCache(fn, key);
  return item.data;
}

async function setDataCache(fn, key) {
  return fn().then((data) => {
    const expireTime = +new Date() + (1000 * 60 * 60); // 1 小时过期
    const item = { data, expireTime };
    localforage.setDriver(localforage.INDEXEDDB);
    localforage.setItem(key, item);
    return data;
  });
};
