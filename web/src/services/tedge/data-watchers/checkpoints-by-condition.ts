import { RequestConfig } from "services/polling-request-proxy/request-config";
import { TedgeRemoteWatcher } from "../remote-data-watcher";
import * as _ from 'lodash';

export interface TedgeFetchCheckpointsParams {
  conditions?: {
    name: string;
    value: string[];
  }[];
  keyword?: string;
  limit?: number;
  notCascade?: boolean;
  operator?: string;
  start?: number;
};

export interface CheckPointData {
  applicationTypeEn: string;
  applicationTypeZh: string;
  attrId: string;
  attrName: string;
  categoryEn: string;
  categoryZh: string;
  deviceNumber: string;
  deviceTypesName: string;
  enumValue: string;
  gid: string;
  id: string;
  readAndWrite: boolean;
  simulation: boolean;
  status: boolean;
  templatePointId: string;
  unit: string;
  updateTime: string;
  value: string;
}

export interface CheckPointDataWithCount {
  count: number;
  list: CheckPointData[];
}

/**
 * 动环测点监听器，按查询条件
 * 用法：
 * const watcher = new CheckpointsByConditionsWatcher([{...}], (data) => console.log(data))
 * watcher.cancel();
 */
export class CheckpointsByConditionsWatcher extends TedgeRemoteWatcher<CheckPointDataWithCount, TedgeFetchCheckpointsParams> {
  constructor(interval: number, private autoCache: boolean = true) {
    super(interval);
  }

  private _cacheInfo: {
    gids: string[];
    attrs: string[];
  } | null = null;

  private _cachedCount: number | null = null;

  private watchTimes = 0;

  watch(paramsInput: TedgeFetchCheckpointsParams, onData: (data: CheckPointDataWithCount) => void) {
    const params = JSON.parse(JSON.stringify(paramsInput));

    this.watchTimes += 1;

    const { watchTimes } = this;

    // 支持链式调用，延后
    setTimeout(async () => {
      if (this.autoCache) {
        const { data, cacheInfo } = await this._loadCacheInfo(params);

        // 过期的请求抛弃
        if (this.watchTimes !== watchTimes) return;

        onData(data);
        this._cacheInfo = cacheInfo;
        super.watch(params, onData);
        return;
      }

      super.watch(params, onData);
    }, 0);

    return this;
  }

  resolveRequestConfig(params: TedgeFetchCheckpointsParams): RequestConfig {
    const {
      autoCache,
      _cacheInfo,
    } = this;

    return this.transformRequestData({
      url: autoCache
        ? '/cgi/dataQuery/edge/getGidAndAttrListValueMapWithoutCache'
        : '/cgi/dataQuery/edge/getCurrentBizGidAttrsWithValueByConditions',
      method: 'POST',
      data: autoCache ? {
        gidWithAttrListMap: [_cacheInfo],
        limit: 999999,
      } : params,
    });
  }

  resolveResponseData(data: any) {
    if (this.autoCache) {
      const res = (newResult) => {
        return {
          "count": this._cachedCount,
          "list": newResult.list.map(item => {
            return {
              "gid": item.device_gid,
              "attrId": item.point_name_en,
              "attrName": item.point_name_zh,
              "deviceNumber": item.device_number,
              "categoryEn": item.device_type_en,
              "categoryZh": item.device_type_zh,
              "applicationTypeEn": item.application_type_en,
              "applicationTypeZh": item.application_type_zh,
              "id": item.point_key,
              "templatePointId": item.point_id,
              "updateTime": item?.update_time,
              "value": item?.latest_value,
              "enumValue": item.enum_value,
              "unit": item.unit,
              "q": item.q,
              "signType": item.sign_type,
              "status": item.status,
              "readAndWrite": item.read_and_write,
              "simulation": item.simulation,
              "isAutoCache": true
            };
          })
        }
      }
      return res(data.data);
    } else {
      return this.transformResponseData({
        count: this._cachedCount,
        list: data.data.list,
      });
    }

  }

  getEdgeRequest(axios) {
    const isTbos = window.tnwebServices.isTbos;
    const getValueFromChangeMap = (changeMap, sourcePath) => {
      // 使用 URL 构造函数解析 sourcePath，提取 pathname
      try {
        const urlObj = new URL(sourcePath, 'https://abc.def.com'); // 使用 dummy 基础 URL
        const path = urlObj.pathname;
        const pagePath = window.location.pathname

        // 检查 changeMap 中是否存在该路径
        if (changeMap.hasOwnProperty(path)) {
          const item = changeMap[path].find(i => {
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
      const deepCopyNewData = _.cloneDeep(newData);
      return deepCopyNewData;
    };
    const debuglog = (...args) => {
      if (localStorage.getItem('logTransform')) {
        console.log(...args)
      } else {
        return
      }
    }
    return {
      post(url, data, loadOpt = true, opts = {}, reqParams = {}) {
        return new Promise(async (resolve, reject) => {
          let changeMap = window.tnwebServices.changeApiMap;
          let newData = data;
          let newUrl = url;
          let apiChangeItem: any = {};
          if (!changeMap) {
            const cgiResult = await axios.post('/cgi/nodeserver/common', {
              'path': 'config_tbosapi/findApi',
              'data': {
                'sourcePath': { $regex: 'cgi' },
              },
            });
            changeMap = _.mapValues(_.groupBy(cgiResult, 'sourcePath'), group => group[0]);
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
            debuglog(changeMap, 'changeMap');
            apiChangeItem = changeApiItem;
            // eslint-disable-next-line no-eval
            let myFunction = eval(`(${apiChangeItem?.change})`)();
            if (_.has(myFunction, 'req')) {
              try {
                newData = myFunction.req({ ...data, ...reqParams });
                myFunction.req = null;
                myFunction = null;
              } catch (error) {
                debuglog('处理请求出错', error, '\n', url, data);
              }
              debuglog('处理请求', '\n', '旧地址', url, data, '\n', '新地址', apiChangeItem?.trueTargetPath, newData);
            }
            newUrl = apiChangeItem?.trueTargetPath
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
      },
      get(url, data, loadOpt = true, opts = {}, reqParams = {}) {
        return new Promise(async (resolve, reject) => {
          let changeMap = window.tnwebServices.changeApiMap;
          let newData = data;
          let newUrl = url;
          let apiChangeItem: any = {};
          if (isTbos && !changeMap) {
            const cgiResult = await axios.post('/cgi/nodeserver/common', {
              'path': 'config_tbosapi/findApi',
              'data': {
                'sourcePath': { $regex: 'cgi' },
              },
            });
            changeMap = _.mapValues(_.groupBy(cgiResult, 'sourcePath'), group => group[0]);
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
            debuglog(changeMap, 'changeMap');
            apiChangeItem = changeApiItem;
            // eslint-disable-next-line no-eval
            let myFunction = eval(`(${apiChangeItem?.change})`)();
            if (_.has(myFunction, 'req')) {
              try {
                newData = myFunction.req({ ...data, ...reqParams });
                myFunction.req = null;
                myFunction = null;
              } catch (error) {
                debuglog('处理请求出错', error, '\n', url, data);
              }
              debuglog('处理请求', '\n', '旧地址', url, data, '\n', '新地址', apiChangeItem?.trueTargetPath, newData);
            }
            newUrl = apiChangeItem?.trueTargetPath
          }

          axios.get(newUrl, newData, loadOpt, opts).then((result) => {
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
      }
    };
  }

  async _loadCacheInfo(params: TedgeFetchCheckpointsParams) {
    const url = '/cgi/dataQuery/edge/getCurrentBizGidAttrsWithValueByConditions';
    // const data = await (window.Vue as any).prototype.$axios.post(url, {
    //   ...params,
    // }, false, false);
    const data: any = await this.getEdgeRequest((window.Vue as any).prototype.$axios).post(url, {
      ...params,
    }, false, false);
    const devicePointIdCache: typeof this._cacheInfo = {
      gids: [],
      attrs: [],
    };

    data.list.forEach((item) => {
      devicePointIdCache.gids.push(item.gid);
      devicePointIdCache.attrs.push(item.attrId);
    });

    devicePointIdCache.gids = _.union(devicePointIdCache.gids);
    devicePointIdCache.attrs = _.union(devicePointIdCache.attrs);

    this._cachedCount = data.count;

    return {
      cacheInfo: devicePointIdCache,
      data,
    };
  }
}
