import * as _ from 'lodash';
import http from 'common/script/http2';
import * as Cookies from 'js-cookie';
import * as yaml from 'yaml';

export interface IPageConfigInfo {
  url: string;
  scope?: string;
  content: {
    type: 'Yaml';
    defaultContent: string;
  };
  docs: {
    type: 'markdown';
    content: string;
  };
  CustomContentInputComp?: Vue.Component;
  ExtraContentComp?: Vue.Component;
}

export interface IPageConfigVmBindOptions {
  vm: Vue;
  scopeElement?: HTMLElement;
  callback: (config: any) => void;
}

/**
 * 自定义配置
 * 页面自定义配置，同时支持子范围配置，如：
 *   xxxx: 页面配置
 *   _scopes:
 *    [scopeId]: 子范围配置
 */
export class CustomConfigService {
  static globalInjectFieldPath: string = '__initInjectCustomConfig';

  moduleId: string = Cookies.get('tnebula_cu_moduleid') || '';

  loadedConfigKeySet: Set<string> = new Set();
  configMap: Map<string, any> = new Map();

  eventBus = new window.Vue();

  continuousClickTimes = 0;
  pageConfigTriggerClickTimes = 10;
  clickClearTimeout?: ReturnType<typeof setTimeout>;

  pageConfigInfo?: IPageConfigInfo;

  constructor() {
    this.initFromGlobal();
    this.initPageConfigurationEvent();
  }

  setModuleId(moduleId) {
    this.moduleId = moduleId;
  }

  initFromGlobal() {
    const injectList: any[] | undefined | null = _.get(window, CustomConfigService.globalInjectFieldPath);
    if (!injectList) return;

    injectList.forEach(item => this.setConfig(item));
  }

  async preload() {
    const list: any[] = await http.get('/cgi/tedge-bff/user-custom-config/preload');
    list.forEach(item => {
      this.setConfig(item);
    });
  }

  get(key: string, scope?: string) {
    const config = this.configMap.get(key);

    if (!config) return config;

    return scope
      ? _.get(config, ['_scopes', scope])
      : config;
  }

  resetConfigValue(key: string, newValue: any) {
    this.configMap.set(key, newValue);
  }

  async loadConfig(key: string) {
    if (this.loadedConfigKeySet.has(key)) return this.get(key);

    // const data = await http.get('/cgi/tedge-bff/user-custom-config/getByKey', {
    //   key,
    //   moduleId: this.moduleId,
    //   url: window.location.pathname,
    // });
    const data = [{
      "id": "EnableDeviceNumberV2__992",
      "key": "EnableDeviceNumberV2",
      "label": "新版设备编号",
      "desc": "",
      "category": "system",
      "url": "",
      "moduleId": "992",
      "preload": true,
      "enable": true,
      "contentType": "Text",
      "createdAt": "2025-07-14T06:51:15.241Z",
      "updatedAt": "2025-07-14T06:51:15.241Z",
      "content": {
        "configId": "EnableDeviceNumberV2__992",
        "type": "Text",
        "content": "1",
        "version": 0,
        "moduleId": "992"
      }
    }, {
      "id": "polling-proxy-mode__992",
      "key": "polling-proxy-mode",
      "label": "轮询代理模式",
      "desc": "",
      "category": "system",
      "url": "",
      "moduleId": "992",
      "preload": true,
      "enable": true,
      "contentType": "Text",
      "createdAt": "2025-07-14T06:51:15.241Z",
      "updatedAt": "2025-07-14T06:51:15.241Z",
      "content": {
        "configId": "polling-proxy-mode__992",
        "type": "Text",
        "content": "http",
        "version": 0,
        "moduleId": "992"
      }
    }].find(item => item.key === key);
    if (!data) return undefined;

    this.setConfig(data);
    return this.get(key);
  }

  async initCurrentPageConfig(pageConfigInfo: IPageConfigInfo, vmBind?: IPageConfigVmBindOptions) {
    if (!pageConfigInfo?.scope) {
      // 如果作用于限定在某个元素内，则无需设置当前页面全局配置
      this.pageConfigInfo = pageConfigInfo;
    }
    const configData = this.resolvePageConfigByInfo(pageConfigInfo);

    await this.loadConfig(configData.key);

    const currentConfigData = this.get(configData.key, pageConfigInfo.scope);

    if (!currentConfigData) {
      this.setConfig(configData, pageConfigInfo.scope);
    }

    if (vmBind) {
      const configChangeEventKey = `config-changed:${configData.key}`;

      if (vmBind.scopeElement) {
        this.bindPageConfigScopeElementEvents(pageConfigInfo, vmBind)
      }

      if (vmBind.callback) {
        this.eventBus.$on(configChangeEventKey, vmBind.callback);
        vmBind.vm.$on('hook:beforeDestroy', () => {
          this.eventBus.$off(configChangeEventKey, vmBind.callback);
        });
      }
    }

    return this.get(configData.key, pageConfigInfo.scope);
  }

  bindPageConfigScopeElementEvents(pageConfigInfo: IPageConfigInfo, vmBind: IPageConfigVmBindOptions) {
    const elt = vmBind.scopeElement;
    if (!elt) return;

    let clickTimes = 0;
    let clickClearTimeout: ReturnType<typeof setTimeout> | null = null;

    elt.addEventListener('click', (evt) => {
      if (!evt.metaKey && !evt.ctrlKey) return;

      (evt as any).scopedPageConfigTriggered = true;

      clickTimes += 1;

      if (clickClearTimeout) {
        clearTimeout(clickClearTimeout);
      }

      if (clickTimes > this.pageConfigTriggerClickTimes) {
        this.showPageEditConfigModal(pageConfigInfo);
        clickTimes = 0;
      }

      clickClearTimeout = setTimeout(() => {
        clickTimes = 0;
      }, 300);
    });
  }

  private setConfig(config: any, scope?: string) {
    if (!config.content) {
      console.warn(`配置没有内容：${config.key}`);
      config.content = '';
    }

    const {
      key,
      enable,
      content: {
        type,
        content,
      },
    } = config;

    if (!enable) return;

    let resultContent = content;

    if (type === 'Yaml' && content) {
      try {
        resultContent = yaml.parse(content);
      } catch (err) {
        console.error('解析自定义yaml配置错误，请检查配置：', content);
        return;
      }
    }

    if (scope) {
      const rootConfig = this.get(key);
      _.set(rootConfig, ['_scopes', scope], resultContent);
      resultContent = rootConfig;
    }

    this.configMap.set(key, resultContent);
    this.loadedConfigKeySet.add(key);

    this.eventBus.$emit('config-change', {
      key,
      content: resultContent,
    });
    this.eventBus.$emit(`config-changed:${key}`, resultContent);
  }

  private initPageConfigurationEvent() {
    window.addEventListener('click', this.handleWindowClicked);
  }

  private handleWindowClicked = (evt) => {
    if (!evt.metaKey && !evt.ctrlKey) return;

    if ((evt as any).scopedPageConfigTriggered) return

    this.continuousClickTimes += 1;

    if (this.clickClearTimeout) {
      clearTimeout(this.clickClearTimeout);
    }

    if (this.continuousClickTimes >= this.pageConfigTriggerClickTimes) {
      this.showPageEditConfigModal();
      this.continuousClickTimes = 0;
    } else {
      this.clickClearTimeout = setTimeout(() => {
        this.continuousClickTimes = 0;
      }, 300);
    }
  }

  private showPageEditConfigModal(pageConfigInfo: IPageConfigInfo = this.pageConfigInfo) {
    if (!pageConfigInfo || !window.tnwebServices.loginStatusService.adminLogined) return;

    if (!checkIsUrlInclude(window.location.href, pageConfigInfo.url)) {
      return;
    }

    const defaultConfig = this.resolvePageConfigByInfo(pageConfigInfo);

    this.eventBus.$emit('editPageConfig', {
      configKey: defaultConfig.key,
      rootConfig: pageConfigInfo.scope
        ? this.resolvePageConfigByInfo(this.pageConfigInfo)
        : defaultConfig,
      defaultConfig,
      info: pageConfigInfo,
    });
  }

  resolvePageConfigKey(url: string) {
    let key = `page-configs[${url}]`;

    return key;
  }

  private resolvePageConfigByInfo(pageConfigInfo: IPageConfigInfo): IPageConfigData {
    const {
      url,
      scope,
      content,
    } = pageConfigInfo;

    const simplifiedUrl = _.last(url.split(window.location.host));

    return {
      key: this.resolvePageConfigKey(simplifiedUrl),
      label: scope ? `【${scope}】--页面配置【${simplifiedUrl}】` : `页面配置【${simplifiedUrl}】`,
      desc: `页面专用配置`,
      category: 'page-configs',
      contentType: content.type,
      content: {
        type: content.type,
        content: content.defaultContent,
      },

      url: simplifiedUrl,
      moduleId: this.moduleId,

      preload: false,
      enable: true,
    };
  }
}

const checkIsUrlInclude = (href: string, url: string) => {
  if (window.decodeURI(href).includes(url)) return true;

  if (href.includes(url)) return true;

  const [hrefPath, hrefSearch] = href.split('?');
  const [urlPath, urlSearch] = url.split('?');

  if (urlSearch && hrefSearch) {
    return hrefPath.includes(urlPath) && hrefSearch.includes(urlSearch);
  }

  return false;
}

export interface IPageConfigData {
  key: string,
  label: string,
  desc: string,
  category: string,
  contentType: string,
  content: {
    type: string,
    content: string,
  },

  url: string,
  moduleId: string,

  preload: boolean,
  enable: boolean,
}
