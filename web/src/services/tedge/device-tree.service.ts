import * as _ from 'lodash';

const axios = (window.Vue as any).prototype.$axios;

export class DeviceTreeService {
  static instance = new DeviceTreeService();

  get isEnableV2() {
    return window.tnwebServices.customConfigService.get('EnableDeviceNumberV2') === '1';
  }

  fetchTreeData(isTypeMode = false) {
    return this.isEnableV2 ? this.fetchTreeDataV2(isTypeMode) : this.fetchTreeDataV1(isTypeMode);
  }

  getEdgeRequest(axios) {
    const isTbos = window.tnwebServices.isTbos;
    const getValueFromChangeMap = (changeMap, sourcePath) => {
      // 使用 URL 构造函数解析 sourcePath，提取 pathname
      try {
        const urlObj = new URL(sourcePath, 'https://test.xyz.com'); // 使用 dummy 基础 URL
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
          const panelPaths = []
          if (targetPage && targetPage instanceof Array && (targetPage.includes(pathName) || targetPage.includes('all')) && !panelPaths.includes(pathName)) {
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
          const panelPaths = []
          if (targetPage && targetPage instanceof Array && (targetPage.includes(pathName) || targetPage.includes('all')) && !panelPaths.includes(pathName)) {
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

  fetchTreeDataV2(isTypeMode = false) {
    return this.getEdgeRequest(axios).post('/cgi/gidmapping/getV2DeviceTree', {
      tree_type: isTypeMode ? 1 : 0,
      is_old: false
    });
  }

  fetchTreeDataV1(isTypeMode = false) {
    return this.getEdgeRequest(axios).post('/cgi/dataQuery/edge/getBizDeviceLevelTree', {
      tree_type: isTypeMode ? 1 : 0,
      is_old: true
    });
  }

  fetchCollectDeviceTree() {
    return axios.get('/cgi/dataQuery/edge/getCollectDeviceTree');
  }
}
