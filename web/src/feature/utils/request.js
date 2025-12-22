/* eslint-disable quote-props */

import config from '@@/config/business';
import { cloneDeep, groupBy, has, mapValues } from 'lodash';
import business from '../../module/tedge/config/business';

const forwardUrls = [
  '/cgi/alarm/rule',
  '/cgi/alarm/history',
  '/cgi/alarm/active',
  '/cgi/alarm/validate/',
  '/cgi/alarm/history/getStat',
  // '/cgi/alarm/active/getActiveDeviceType',
  // '/cgi/alarm/active/getActiveDeviceList',
  // '/cgi/alarm/active/detail',
  // '/cgi/alarm/active/getPointDataType',
  // '/cgi/alarm/active/getPointData',
  '/cgi/dataQuery/edge/getCollectDeviceTree'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/getBizDeviceTree'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/getBizDeviceLevelTree'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/getV2DeviceTree'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/queryPointDetailInfoWithCurrentValueByGid'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/queryHistoryPointInfoByTimeRangeAndPageAndOrder'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/exportExcelByCollectorGidWithCurrentData'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/exportHistoryPointInfoByTimeRangeAndPageAndOrder'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/getDistinctByFieldName'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/getCurrentBizGidAttrsWithValueByConditions'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/exportCurrentBizGidAttrsWithValueByConditions'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/getMatchByFieldNameAndValue'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/getDistinctFieldByCascade'.toLocaleLowerCase(),
  '/cgi/alarm/getDeviceListByProtocolType'.toLocaleLowerCase(),
  '/cgi/alarm/queryPointTypeList'.toLocaleLowerCase(),
  '/cgi/alarm/getProtocolTypeDropdown'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/queryPointDetailInfoWithCurrentValueByGid'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/getHistoryBizGidAttrValuesByTemplate'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/insertOrUpdateTemplate'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/selectTemplateByCondition'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/deleteTemplateById'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/calByExpressAndTimeRange'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/validateDetailByExpress'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/diagnosisByExpressAndUpdateTime'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/exportByExpressAndTimeRange'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/exportHistoryBizGidAttrValuesByTemplate'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/queryHistoryIndicatorExportExcel'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/queryHistoryIndicatorWithExp'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/exportTemplateByCondition'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/importTemplateByCondition'.toLocaleLowerCase(),
  '/cgi/dataQuery/edge/getDeviceNumberListAndAttrs'.toLocaleLowerCase(),
  '/cgi/alarm/getDeviceNumberDropdown'.toLocaleLowerCase(),
];

function isForward(mozuId, url) {
  if (mozuId && config.isForwardEdge) {
    const lowerUrl = url.toLowerCase();
    let matched = false;
    forwardUrls.forEach((i) => {
      if (lowerUrl.startsWith(i)) {
        matched = true;
      }
    });
    return matched;
  }
  return false;
}
function getUrl(url) {
  // 监控管理（当前、已挂起、历史页面接口需要去掉/cgi/forwardEdge
  // 便于不影响其它页面）
  const pageUrl = window.location.href;
  if (pageUrl.includes('timpage/actived-warning') || pageUrl.includes('timpage/hangup-warning')
    || pageUrl.includes('timpage/warning-history')) {
    return `${url}`;
  }
  return `/cgi/forwardEdge${url}`;
}
function getOpts(mozuId, opts) {
  const headers = {
    platform: 'cloud',
    mozuId,
  };
  if (opts) {
    if (opts.restAxios) {
      Object.assign(opts.restAxios, {
        headers: (opts.restAxios.headers) ? (
          Object.assign({}, opts.restAxios.headers, headers)
        ) : headers,
      });
    } else {
      Object.assign(
        opts,
        {
          restAxios: {
            headers,
          },
        }
      );
    }
  } else {
    opts = {
      restAxios: {
        headers,
      },
    };
  }
  return opts;
}

function debuglog (...args)  {
  if (localStorage.getItem('logTransform')) {
      console.log(...args)
  } else {
      return
  }
}

function getValueFromChangeMap(changeMap, sourcePath) {
  // 使用 URL 构造函数解析 sourcePath，提取 pathname
  try {
    const urlObj = new URL(sourcePath, 'https://xyz.abc.com'); // 使用 dummy 基础 URL
    // 获取去除参数后的请求
    const path = urlObj.pathname;
    const pagePath = window.location.pathname
    // 检查 changeMap 中是否存在该路径
    if (changeMap.hasOwnProperty(path)) {
      const item = changeMap[path].find(i => {
        // 目标页面包含当前页面路径
        return i.targetPage.includes(pagePath) || i.targetPage.includes('all')
      })
      return item;
    } else {
      return null; // 或者根据需求返回其他默认值
    }
  } catch (error) {
    console.error('无效的 URL:', sourcePath);
    return null;
  }
}

export default function getEdgeRequest(axios, mozuId) {
  const { isTbos } = business;
  const solveFunc = (result, v) => {
    // eslint-disable-next-line no-eval
    let myFunction = eval(`(${v.change})`)();
    let newData = result;
    try {
      newData = myFunction.res(result);
      myFunction.res = null;
      myFunction = null;
    } catch (error) {
      debuglog('处理返回错误', '\n', v.sourcePath, result);
      return result;
    }
    debuglog('处理返回', '\n', '旧地址', v.sourcePath, result, '\n', '新地址', v?.trueTargetPath, newData);
    const deepCopyNewData = cloneDeep(newData);
    return deepCopyNewData;
  };
  return {
    post(url, data, loadOpt = true, opts, reqParams = {}) {
      if (isForward(mozuId, url)) {
        url = getUrl(url);
        opts = getOpts(mozuId, opts);
      }
      return new Promise(async (resolve, reject) => {
        let changeMap = window.tnwebServices.changeApiMap;
        let newData = data;
        let newUrl = url;
        let apiChangeItem = '';
        if (isTbos && !changeMap) {
          const cgiResult = await axios.post('/cgi/nodeserver/common', {
            'path': 'config_tbosapi/findApi',
            'data': {
              'sourcePath': { $regex: 'cgi' },
            },
          });
          changeMap = mapValues(groupBy(cgiResult, 'sourcePath'), group => group[0]);
        }
        const pathName = window.location.pathname.replace('.html', '');
        let pathCheck = false;
        const changeApiItem = getValueFromChangeMap(changeMap, url);
        const targetPage = changeApiItem?.targetPage;
        if (targetPage && targetPage instanceof Array && (targetPage.includes(pathName) || targetPage.includes('all'))) {
          pathCheck = true;
        }
        const transformCheck = isTbos && changeMap && changeApiItem
          && changeApiItem?.trueTargetPath && pathCheck;
        if (transformCheck) {
          apiChangeItem = changeApiItem;
          // eslint-disable-next-line no-eval
          let myFunction = eval(`(${apiChangeItem.change})`)();
          if (isTbos && has(myFunction, 'req')) {
            try {
              newData = myFunction.req({ ...data, ...reqParams });
              myFunction.req = null;
              myFunction = null;
            } catch (error) {
              debuglog('处理请求出错', error, '\n', url, data);
            }
            debuglog('处理请求', '\n', '旧地址', url, data, '\n', '新地址', apiChangeItem?.trueTargetPath, newData);
          }
          newUrl = apiChangeItem?.trueTargetPath;
        }

        axios.post(newUrl, newData, loadOpt, opts).then((result) => {
          if (transformCheck) {
            resolve(solveFunc(result, apiChangeItem));
          } else {
            resolve(result);
          }
        })
          .catch((e) => {
            reject(e);
          });
      });
      // return axios.post(url, data, loadOpt, opts);
    },
    get(url, data, loadOpt = true, opts) {
      if (isForward(mozuId, url)) {
        url = getUrl(url);
        opts = getOpts(mozuId, opts);
      }
      return axios.get(url, data, loadOpt, opts);
    },
    download(url, data, loadOpt = true, opts) {
      if (isForward(mozuId, url)) {
        url = getUrl(url);
        opts = getOpts(mozuId, opts);
      }
      return new Promise(async () => {
        let changeMap = window.tnwebServices.changeApiMap;
        let newData = data;
        let newUrl = url;
        let apiChangeItem = '';
        if (isTbos && !changeMap) {
          const cgiResult = await axios.post('/cgi/nodeserver/common', {
            'path': 'config_tbosapi/findApi',
            'data': {
              'sourcePath': { $regex: 'cgi' },
            },
          });
          changeMap = mapValues(groupBy(cgiResult, 'sourcePath'), group => group[0]);
        }
        debuglog(changeMap, 'changeMap', window.location.pathname, url);
        const pathName = window.location.pathname.replace('.html', '');
        let pathCheck = false;
        const changeApiItem = getValueFromChangeMap(changeMap, url);
        const targetPage = changeApiItem?.targetPage;
        if (targetPage && targetPage instanceof Array && (targetPage.includes(pathName) || targetPage.includes('all'))) {
          pathCheck = true;
        }
        const transformCheck = isTbos && changeMap && changeApiItem
        && changeApiItem?.trueTargetPath && pathCheck;
        if (transformCheck) {
          apiChangeItem = changeApiItem;
          // eslint-disable-next-line no-eval
          let myFunction = eval(`(${apiChangeItem.change})`)();
          if (isTbos && has(myFunction, 'req')) {
            try {
              newData = myFunction.req({ ...data });
              myFunction.req = null;
              myFunction = null;
            } catch (error) {
              debuglog('处理请求出错', error, '\n', url, data);
            }
            debuglog('处理请求', '\n', '旧地址', url, data, '\n', '新地址', apiChangeItem?.trueTargetPath, newData);
          }
          newUrl = apiChangeItem?.trueTargetPath;
        }
        axios.download(newUrl, newData, loadOpt, opts);
      });
      // return axios.download(url, data, loadOpt, opts);
    },
  };
}

export function transformQueryRsp(data) {
  return {
    count: data.total !== undefined ? data.total : 0,
    list: data.list.map(item => ({
      gid: item.device_gid,
      attrId: item.point_name_en,
      attrName: item.point_name_zh,
      deviceNumber: item.device_number,
      categoryEn: item.device_type_en,
      categoryZh: item.device_type_zh,
      applicationTypeEn: item.application_type_en,
      applicationTypeZh: item.application_type_zh,
      id: item.point_key,
      templatePointId: item.point_id,
      updateTime: item.update_time,
      value: item.latest_value,
      enumValue: item.enum_value,
      unit: item.unit,
      q: item.q,
      signType: item.sign_type,
      status: item.status,
      readAndWrite: item.read_and_write,
      simulation: item.simulation,
    })),
  };
}
