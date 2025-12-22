import { has } from 'lodash';
import { PollingProxyPluginData } from 'services/polling-request-proxy/plugin/types';
import { ClientProxy } from "services/polling-request-proxy/polling-proxy.service";
import { RequestConfig, RequestProxyConfig } from "services/polling-request-proxy/request-config";
import * as _ from 'lodash';
export abstract class TedgeRemoteWatcher<T, P> {
  constructor(private interval: number) {
    if (typeof this.interval !== 'number') {
      throw new Error('interval参数错误');
    }
  }

  pollingProxyAgentService = window.tnwebServices.pollingProxyAgentService;

  lastClientProxy: ClientProxy<T> | null = null;

  plugins?: RequestProxyConfig['plugins'];

  changeApiMap = window.tnwebServices.changeApiMap;

  changeApiItem = null;

  private _cacheMockRequest: {
    lastParamsJson: string;
    lastResultDataPromise?: Promise<T>;
    lastResultDataPromiseReject?: () => void;
    lastResultData?: T;
  } | null = null;

  private watchVersion = 0;

  watch(params: P, onData: (data: T) => void) {
    // 支持链式调用，延后触发
    setTimeout(() => {
      this.cancel();
      const watchVersion = this.watchVersion += 1;
      this.lastClientProxy = this.pollingProxyAgentService.proxy<any>({
        interval: this.interval,
        plugins: this.plugins,
        request: this.transformRequestData(this.resolveRequestConfig(params)),
      }, (data) => {
        if (watchVersion !== this.watchVersion) return;
        onData(this.resolveResponseData(this.transformResponseData(data)));
      });
    }, 0);

    return this;
  }

  /** 模拟轮询调用，为了低成本迁移，若是新开发代码建议使用watch方法 */
  async mockRequest(params: P): Promise<T> {
    const paramsJson = JSON.stringify(params);

    if (paramsJson === this._cacheMockRequest?.lastParamsJson) {
      return this._cacheMockRequest.lastResultDataPromise
        ? this._cacheMockRequest.lastResultDataPromise
        : Promise.resolve(this._cacheMockRequest.lastResultData);
    } else if (this._cacheMockRequest?.lastResultDataPromise) {
      this._cacheMockRequest.lastResultDataPromiseReject();
      this._cacheMockRequest.lastResultDataPromise = null;
      this._cacheMockRequest.lastResultDataPromiseReject = null;
    }

    if (!this._cacheMockRequest) {
      this._cacheMockRequest = {
        lastParamsJson: paramsJson,
        lastResultDataPromise: null,
      };
    } else {
      this._cacheMockRequest.lastParamsJson = paramsJson;
      this._cacheMockRequest.lastResultData = null;
    }

    this._cacheMockRequest.lastResultDataPromise = new Promise<T>((resolve, reject) => {
      let resolved = false;
      this._cacheMockRequest.lastResultDataPromiseReject = () => {
        resolved = true;
        reject('cancel');
      };
      this.watch(params, (data) => {
        this._cacheMockRequest.lastResultData = data;

        if (resolved) return;
        resolve(data);
        this._cacheMockRequest.lastResultDataPromiseReject = null;
        resolved = true;
        this._cacheMockRequest.lastResultDataPromise = null;
      });
    });

    return this._cacheMockRequest.lastResultDataPromise;
  }

  cancel() {
    if (!this.lastClientProxy) return;
    this.pollingProxyAgentService.exit([this.lastClientProxy]);
  }

  bindVueVm(vm: any) {
    vm.$once('hook:beforeDestroy', () => {
      this.cancel();
    });
    return this;
  }

  getValueFromChangeMap(changeMap, sourcePath) {
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

  debuglog(...args) {
    if (localStorage.getItem('logTransform')) {
      console.log(...args)
    } else {
      return
    }
  }

  // 解析请求配置，用于对请求有特殊需求的场景
  protected abstract resolveRequestConfig(params: P): RequestConfig;
  // 返回响应数据，传入经过transformResponseData处理的数据，返回结果用于监听器回调
  protected abstract resolveResponseData(data: any): T;

  // 请求转换，用于一些数据格式处理
  protected transformRequestData(params: RequestConfig): RequestConfig {
    const { url } = params
    if (!url) return params;
    this.changeApiItem = this.getValueFromChangeMap(this.changeApiMap, url);
    let newParams = params.data
    const pathName = window.location.pathname.replace('.html', '');
    let pathCheck = false;
    const targetPage = this.changeApiItem?.targetPage;
    if (targetPage && targetPage instanceof Array && (targetPage.includes(pathName) || targetPage.includes('all'))) {
      pathCheck = true;
    }
    const transformCheck = this.changeApiItem && this.changeApiItem?.change && this.changeApiItem?.trueTargetPath && pathCheck;
    console.log('transformCheck', transformCheck)
    if (transformCheck) {
      const myFunction = eval(`(${this.changeApiItem.change})`)();
      if (has(myFunction, 'req')) {
        try {
          newParams = myFunction.req({ ...params.data })
        } catch (error) {
          this.debuglog('处理请求出错', error, '\n', `${url}`, params)
          return params;
        }
      }
      this.debuglog('处理请求', '\n', `旧地址${url}`, params, '\n', `新地址${this.getValueFromChangeMap(this.changeApiMap, url)}`, {
        ...params,
        data: { ...newParams },
        url: this.getValueFromChangeMap(this.changeApiMap, url)?.trueTargetPath ? this.getValueFromChangeMap(this.changeApiMap, url)?.trueTargetPath : url
      })
      return {
        ...params,
        data: { ...newParams },
        url: this.getValueFromChangeMap(this.changeApiMap, url)?.trueTargetPath ? this.getValueFromChangeMap(this.changeApiMap, url)?.trueTargetPath : url
      }
    } else {
      return params;
    }
  };


  // 响应转换，用于一些数据格式处理
  protected transformResponseData(data: any): any {
    // 如果已经转换过，直接返回
    if (data.data?.hasTransformed) return data;
    const pathName = window.location.pathname.replace('.html', '');
    let pathCheck = false;
    const targetPage = this.changeApiItem?.targetPage;
    if (targetPage && targetPage instanceof Array && (targetPage.includes(pathName) || targetPage.includes('all'))) {
      pathCheck = true;
    }
    const transformCheck = this.changeApiItem && this.changeApiItem?.change && this.changeApiItem?.trueTargetPath && pathCheck;
    if (transformCheck) {
      const myFunction = eval(`(${this.changeApiItem.change})`)();
      if (has(myFunction, 'res')) {
        let newData = null
        try {
          newData = myFunction.res({ ...data.data })
        } catch (error) {
          this.debuglog('处理返回出错', error, '\n', this.changeApiItem.sourcePath, data)
          this.changeApiItem = null;
          return data
        }
        this.debuglog('处理返回', '\n', '旧地址', this.changeApiItem.sourcePath, data, '\n', '新地址', this.changeApiItem.trueTargetPath, newData)
        return {
          data: newData
        }
      } else {
        return data
      }
    } else {
      return data
    }
  };

  withPlugin(pluginData: PollingProxyPluginData) {
    if (!this.plugins) {
      this.plugins = [];
    }
    this.plugins.push(pluginData);
    return this;
  }

  withDiffPlugin() {
    if (!this.plugins) {
      this.plugins = [];
    }
    this.plugins.push({
      id: 'diff',
      config: {},
    });
    return this;
  }

  /** 获取时间参数，距离当前时间差距的值作为参数 */
  withTimeParamsPlugin(params: {
    [key: string]: {
      paramType: 'body' | 'query',
      diff: number,
    }
  }) {
    if (!this.plugins) {
      this.plugins = [];
    }
    this.plugins.push({
      id: 'timeParams',
      config: params,
    });
    return this;
  }
}
