import io from 'socket.io-client';
import { RequestProxyConfig, RequestConfig } from './request-config';
import { uniqueId } from 'lodash';
import Axios, { AxiosResponse } from 'axios';
import { POLLING_PROXY_EVENT } from './message-types';
import { PollingProxySwitcher } from './polling-proxy-switcher';
import { PollingPorxyPluginManager } from './plugin/manager';
import _ = require('lodash');
import { BACKEND_SERVICE_NAMES_MAP, BACKEND_SVC_CGI_NAMES, POLORIS_BACKEND_SERVICE_NAMES } from './backend-service-names.const';
import { transformMap } from './transformMap';

export class ClientProxy<T> {
  constructor(config: RequestProxyConfig, onData: ClientProxy<T>['onData']) {
    this.config = config;
    this.onData = onData;
  }

  config: RequestProxyConfig;
  onData: ((data: T) => void);
  status: 'joining' | 'joined' | 'exit' = 'joining';

  serverProxyId?: string;

  devData?: { status: string; data: any; };
  updatedAt?: Date;
};

export class PollingProxyAgentService {
  constructor(public moduleDomain: string, defaultMode: PollingProxyAgentService['runningMode'] = 'websocket') {
    this._bindSocketEvents();
    this.runningMode = defaultMode || 'websocket';
    if (this.runningMode === 'websocket') {
      this.socket.connect();
    }
  }

  static PollingProxySwitcher = PollingProxySwitcher;
  PollingProxySwitcher = PollingProxySwitcher;

  socket = io('', {
    path: '/cgi/tedge-bff/socket.io',
    transports: ['websocket'],
    autoConnect: false,
  });

  runningMode: 'websocket' | 'http' = 'websocket';
  devMode: boolean = localStorage.getItem('_polling-proxy-dev-mode') === 'on';

  localPollingRequestAdaptor = new LocalPollingRequestAdaptor();

  /** { [clientId]: ClientProxy } */
  clientIdToProxyMap = new Map<string, ClientProxy<any>>();
  /** { [serverId]: { [clientId]: ClientProxy } } */
  serverIdToProxyMap = new Map<string, Map<string, ClientProxy<any>>>();
  /** 最近一次数据缓存 */
  serverProxyIdToLastDataMap = new Map<string, any>();

  toggleDevMode(devMode: boolean) {
    this.socket.emit('toggleClientDevMode', devMode);
    this.devMode = devMode;
  }

  /** 更新开发者工具回调 */
  onUpdateDevTool: () => void = () => void (0);

  /**
   * 启动/加入代理
   */
  proxy<T>(params: Omit<RequestProxyConfig, 'clientProxyId' | 'moduleDomain'>, onData: (data: T) => void) {
    const clientProxyId = uniqueId();

    const config = {
      clientProxyId,
      moduleDomain: this.moduleDomain,
      ...params,
      request: {
        ...params.request,
        headers: {
          ...params.request.headers,
          Cookie: document.cookie,
          mozuid: (window as any)?.Vue?.prototype?.$moduleInfo?.mozuId,
          "platform": "tbos",
          'x-client-id': 'tedge.web.zt-views',
        },
      },
    };

    const proxy = new ClientProxy(config, onData);
    this.clientIdToProxyMap.set(clientProxyId, proxy);
    if (this.runningMode === 'http') {
      this.localPollingRequestAdaptor.runProxy(proxy);
    } else {
      let svcUrl = config.request.url;
      console.log(config.request.url, 'svcurl')
      let toBase64 = false
      if (config.request.url.includes('getGidAndAttrListValueMapWithoutCache')) {
        // svcUrl = 'http://idc-public-gateway:8080/cgi/idc-tbos-cgi/Data/Query'
        svcUrl = 'aHR0cDovL2lkYy1wdWJsaWMtZ2F0ZXdheTo4MDgwL2NnaS9pZGMtdGJvcy1jZ2kvRGF0YS9RdWVyeQ=='
        toBase64 = true
      }
      if (config.request.url.includes('/cgi/idc-tbos-cgi/Data/Query')) {
        // svcUrl = 'http://idc-public-gateway:8080/cgi/idc-tbos-cgi/Data/Query'
        svcUrl = 'aHR0cDovL2lkYy1wdWJsaWMtZ2F0ZXdheTo4MDgwL2NnaS9pZGMtdGJvcy1jZ2kvRGF0YS9RdWVyeQ=='
        toBase64 = true
      }
      if (config.request.url.includes('/active/getList')) {
        // svcUrl = 'http://idc-public-gateway:8080/cgi/idc-tbos-cgi/alarm/server/GetAlarmList'
        svcUrl = 'aHR0cDovL2lkYy1wdWJsaWMtZ2F0ZXdheTo4MDgwL2NnaS9pZGMtdGJvcy1jZ2kvYWxhcm0vc2VydmVyL0dldEFsYXJtTGlzdA=='
        toBase64 = true
      }
      if (config.request.url.includes('/cgi/alarm/active/getActiveOverview')) {
        // svcUrl = 'http://idc-public-gateway:8080/cgi/idc-tbos-cgi/alarm/server/GetAlarmList'
        svcUrl = 'aHR0cDovL2lkYy1wdWJsaWMtZ2F0ZXdheTo4MDgwL2NnaS9pZGMtdGJvcy1jZ2kvYWxhcm0vc2VydmVyL0dldEFsYXJtTGlzdA=='
        toBase64 = true
      }
      if (config.request.url.includes('/cgi/alarm/ba/getHoursAlarmTrend')) {
        // svcUrl = 'http://idc-public-gateway:8080/cgi/idc-tbos-cgi/alarm/server/GetAlarmCntTrend'
        svcUrl = 'aHR0cDovL2lkYy1wdWJsaWMtZ2F0ZXdheTo4MDgwL2NnaS9pZGMtdGJvcy1jZ2kvYWxhcm0vc2VydmVyL0dldEFsYXJtQ250VHJlbmQ='
        toBase64 = true
      }
      if (config.request.url.includes('/cgi/dataQuery/edge/getHistoryBizGidAttrValues')) {
        // svcUrl = 'http://idc-public-gateway:8080/cgi/idc-tbos-cgi/Data/Query'
        svcUrl = 'aHR0cDovL2lkYy1wdWJsaWMtZ2F0ZXdheTo4MDgwL2NnaS9pZGMtdGJvcy1jZ2kvRGF0YS9RdWVyeQ=='
        toBase64 = true
      }

      (config as any).request.sourceUrl = config.request.url;
      config.request.url = toBase64 ? atob(svcUrl) : svcUrl;
      console.log(config.request, 'polling config.request')
      config.request.data = this.localPollingRequestAdaptor.transformRequestData({
        url: (config.request as any)?.sourceUrl,
        data: config.request.data
      })?.data
      this.socket.emit(POLLING_PROXY_EVENT.join, {
        url: toBase64 ? atob(svcUrl) : svcUrl,
        ...config,
        request: {
          url: toBase64 ? atob(svcUrl) : svcUrl,
          ..._.omit(config.request, ['url']),
        },
      });
    }

    this.onUpdateDevTool();

    return proxy;
  }

  /** 退出代理 */
  exit(proxyList: ClientProxy<any>[]) {
    const {
      clientIdToProxyMap,
      serverIdToProxyMap,
    } = this;

    const serverIds: string[] = [];

    proxyList.forEach(proxy => {
      clientIdToProxyMap.delete(proxy.config.clientProxyId);

      serverIdToProxyMap.forEach(clientMap => {
        clientMap.delete(proxy.config.clientProxyId);

        if (clientMap.size === 0) {
          serverIds.push(proxy.serverProxyId);
        }
      });

      proxy.status = 'exit';
    });

    if (this.runningMode === 'http') {
      proxyList.forEach(proxy => {
        this.localPollingRequestAdaptor.cancelProxy(proxy);
      });
    } else {
      this._exitServerProxy(serverIds);
    }

    serverIds.forEach(serverId => {
      serverIdToProxyMap.delete(serverId);
    });

    this.onUpdateDevTool();
  }

  /** 加入代理结果处理 */
  private handleJoinResult({ clientProxyId, serverProxyId }) {
    const proxy = this.clientIdToProxyMap.get(clientProxyId);

    if (!proxy) {
      console.warn('没有找到客户端代理ID：', clientProxyId);
      return;
    }

    proxy.serverProxyId = serverProxyId;
    proxy.status = 'joined';

    let clientMap = this.serverIdToProxyMap.get(serverProxyId);

    if (!clientMap) {
      clientMap = new Map();
      this.serverIdToProxyMap.set(serverProxyId, clientMap);
    }

    clientMap.set(clientProxyId, proxy);

    if (this.serverProxyIdToLastDataMap.has(serverProxyId)) {
      const lastResp = this.serverProxyIdToLastDataMap.get(serverProxyId);

      proxy.onData(lastResp);
      proxy.devData = {
        status: 'success',
        data: lastResp.data,
      };
    }

    this.onUpdateDevTool();
  }

  /** 请求结果消息处理 */
  private handleRequestData({ serverProxyId, data }, cb: (msg?: any) => void) {
    try {
      let resp: AxiosResponse = typeof data === 'string' ? JSON.parse(data) : data;
      const clientMap = this.serverIdToProxyMap.get(serverProxyId);
      if (!clientMap) return;
      clientMap.forEach((proxy) => {
        const sourceUrl = (proxy.config.request as any)?.sourceUrl
        if (!proxy.config.plugins || proxy.config.plugins?.length <= 1) {
          resp = this.localPollingRequestAdaptor.transformResponseData(resp, sourceUrl)
        }
        proxy.onData(resp.data);
        proxy.updatedAt = new Date();
      });

      if (this.devMode) {
        clientMap.forEach((proxy) => {
          proxy.devData = {
            status: 'success',
            data: resp,
          };
        });
      }

      this.serverProxyIdToLastDataMap.set(serverProxyId, resp);

      this.onUpdateDevTool();
    } finally {
      if (cb) {
        cb();
      }
    }
  }

  private handleRequestDataError({ serverProxyId, data }) {
    const clientMap = this.serverIdToProxyMap.get(serverProxyId);
    if (!clientMap) return;

    const item = clientMap.get(clientMap.keys().next().value);
    console.error('轮询代理--远程请求错误：', data, item.config);

    if (this.devMode) {
      clientMap.forEach((proxy) => {
        proxy.devData = {
          status: 'error',
          data,
        };
        proxy.updatedAt = new Date();
      });
    }

    this.onUpdateDevTool();
  }

  /** 绑定websocket事件 */
  private _bindSocketEvents() {
    const { socket } = this;

    // 加入代理结果事件
    socket.on(POLLING_PROXY_EVENT.joinResult, this.handleJoinResult.bind(this));

    // 代理请求数据事件
    socket.on(POLLING_PROXY_EVENT.requestData, this.handleRequestData.bind(this));

    // 代理请求数据错误事件
    socket.on(POLLING_PROXY_EVENT.requestDataError, this.handleRequestDataError.bind(this));

    socket.on('disconnect', (reason) => {
      this._switchRunningMode('http');
    });

    socket.on('connect', () => {
      this._switchRunningMode('websocket');

      this.toggleDevMode(this.devMode);
    });

    socket.on('connect_error', () => {
      this._switchRunningMode('http');
    });
  }

  private _exitServerProxy(serverProxyIds: string[]) {
    this.socket.emit(POLLING_PROXY_EVENT.exit, { serverProxyIds: serverProxyIds });
    serverProxyIds.forEach(id => {
      this.serverProxyIdToLastDataMap.delete(id);
    });
  }

  private _switchRunningMode(mode: typeof this.runningMode) {
    if (this.runningMode === mode) return;

    if (mode === 'http') {
      this._switchHttpMode();
    } else if (mode === 'websocket') {
      this._switchWebsocketMode();
    }

    this.runningMode = mode;
  }

  private _switchHttpMode() {
    const { localPollingRequestAdaptor } = this;

    this.clientIdToProxyMap.forEach(item => {
      localPollingRequestAdaptor.runProxy(item);
    });
  }

  private _switchWebsocketMode() {
    const {
      clientIdToProxyMap,
    } = this;

    clientIdToProxyMap.forEach(proxy => {
      this.socket.emit(POLLING_PROXY_EVENT.join, proxy.config);
    });

    // 清理老的服务端ID相关数据，重新join以后会产生新的id
    this.serverIdToProxyMap.clear();
    this.serverProxyIdToLastDataMap.clear();

    this.localPollingRequestAdaptor.clear();
  }
}

export class LocalPollingRequestAdaptor {
  axios = Axios.create();
  pluginManager = new PollingPorxyPluginManager();

  clientProxyMap = new Map<ClientProxy<any>, {
    cancel(): void;
  }>();

  changeApiItem: any = null;
  changeApiMap = _.mapValues(_.groupBy(transformMap, 'sourcePath'), group => group);

  clear() {
    this.clientProxyMap.forEach(item => item.cancel());
    this.clientProxyMap.clear();
  }

  cancelProxy(proxy: ClientProxy<any>) {
    this.clientProxyMap.get(proxy)?.cancel();
    this.clientProxyMap.delete(proxy);
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

  transformRequestData(params: RequestConfig): RequestConfig {
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
    if (transformCheck) {
      const myFunction = eval(`(${this.changeApiItem.change})`)();
      if (_.has(myFunction, 'req')) {
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
        // url: this.getValueFromChangeMap(this.changeApiMap, url)?.trueTargetPath ? this.getValueFromChangeMap(this.changeApiMap, url)?.trueTargetPath : url
      }
    } else {
      return params;
    }
  };

  transformResponseData(data: any, sourceUrl: any = ''): any {
    if (data.data.data?.hasTransformed) return data;
    const pathName = window.location.pathname.replace('.html', '');
    let pathCheck = false;
    let currentChangeApiItem = this.getValueFromChangeMap(this.changeApiMap, sourceUrl);
    const targetPage = currentChangeApiItem?.targetPage;
    if (targetPage && targetPage instanceof Array && (targetPage.includes(pathName) || targetPage.includes('all'))) {
      pathCheck = true;
    }
    const transformCheck = currentChangeApiItem && currentChangeApiItem?.change && currentChangeApiItem?.trueTargetPath && pathCheck;
    if (transformCheck) {
      const myFunction = eval(`(${currentChangeApiItem.change})`)();
      if (_.has(myFunction, 'res')) {
        let newData = null
        try {
          newData = myFunction.res({ ...data.data.data })
        } catch (error) {
          this.debuglog('处理返回出错', error, '\n', currentChangeApiItem.sourcePath, data)
          currentChangeApiItem = null;
          return data
        }
        this.debuglog('处理返回', '\n', '旧地址', currentChangeApiItem.sourcePath, data, '\n', '新地址', currentChangeApiItem.trueTargetPath, newData)
        return {
          ...data,
          data: {
            data: newData
          }
        }
      } else {
        return data
      }
    } else {
      return data
    }
  };

  runProxy(clientProxy: ClientProxy<any>) {
    const {
      config: {
        request: requestConfig,
        interval,
        plugins,
      },
      onData,
    } = clientProxy;

    let isRunning = true;

    const tick = async () => {
      try {
        const headers = {
          ...requestConfig.headers,
        };
        delete headers.Cookie;
        const resp = await this.axios.request(this.transformRequestData({
          ...this.pluginManager.prepareRequestConfig(requestConfig, plugins),
          url: (requestConfig as any).sourceUrl || requestConfig.url,
          headers,
        }));
        const newResp = this.pluginManager.postRequest(plugins, resp, clientProxy);
        onData(newResp.data);
      } finally {
        if (!isRunning) return;
        setTimeout(tick, interval);
      }
    };

    tick();

    this.clientProxyMap.set(clientProxy, {
      cancel: () => {
        isRunning = false;
      },
    });
  }
}
