import Vue from 'vue';
import Cookies from 'js-cookie';
import init from 'common/script/bootstrap2';
import { create } from 'common/script/http2';
// import { plugin, redirectToLogin } from 'common/script/passport_login'
// import passportPlugin from 'common/script/passport_plugin';
import main from './main.vue';
import 'common/style/ui/_reset.scss';
import 'common/style/reset.css';
import './assets/local.css';
// import * as reqhel from '../common/reqhel.js';
import { getAllOpCode } from '../common/auth.js';
import passportPlugin from 'common/script/passport_plugin';
import { CustomConfigService } from 'services/custom-config.service';
import { V2DeviceNumberTransformerService } from 'services/v2-device-number-transformer.service';
import { CheckPointRealtimeDataService } from 'services/tedge/check-point-realtime-data.service';
import { LoginStatusService } from 'services/login-status.service';
import { PollingProxyAgentService } from 'services/polling-request-proxy/polling-proxy.service';
import CustomConfigValue from '@/feature/component/custom-config-value.vue';
import { groupBy, has, isEmpty, keys, mapValues, omit } from 'lodash';
// import tbosTransform from './assets/tbosTransform.json';
// 在tbos-transform的funcs文件夹下有各个接口的处理函数，修改后运行 transFuncToObj.js 更新 transformMap.json
import newTbosTransform from './assets/tbos-transform/transformMap.json';
import { getQueryString } from 'common/script/utils.js';

Vue.component(
  'CustomConfigValue',
  CustomConfigValue,
);

if (!window.TNBL) {
  window.TNBL = {};
  TNBL.redirectUrl = function () { console.log('dev空函数:redirectUrl', arguments); };
}

(async () => {
  const $axios = create({
    isJson: true,
  });

    // 登录的逻辑
    const psPlugin = passportPlugin({
      opts: {
        isJson: true,
      },
    });
    init({
      plugins: [psPlugin],
      customHttp: true,
    });

  // eslint-disable-next-line prefer-arrow-callback
  Vue.prototype.$axios.ins.interceptors.request.use(function (config) {
    const { pathname } = window.location;

    let clientId;
    if (pathname.includes('zt')) {
      clientId = 'tedge.web.zt-view';
    } else if (pathname.includes('warning')) {
      clientId = 'tedge.web.alarm';
    } else if (pathname.includes('electric-system')) {
      clientId = 'tedge.web.electric';
    } else {
      clientId = 'tedge.web.default';
    }

    if (!config.headers) {
      // eslint-disable-next-line no-param-reassign
      config.headers = {};
    }

    // eslint-disable-next-line no-param-reassign
    config.headers['X-CLIENT-ID'] = clientId;

    return config;
  });

  try {
    const customConfigService = new CustomConfigService();
    const v2DeviceNumberTransformerService = new V2DeviceNumberTransformerService();
    const checkPointRealtimeDataService = new CheckPointRealtimeDataService();
    const loginStatusService = new LoginStatusService();

    v2DeviceNumberTransformerService.parseCustomConfigContent(customConfigService.get('DeviceNumberV2Mapping'));

    // 框架和modules里各个模块是分批次构建的，import通用service是不相通的，所以只能通过全局来注册和传递
    window.tnwebServices = {
      customConfigService,
      v2DeviceNumberTransformerService,
      checkPointRealtimeDataService,
      loginStatusService
    };
    const isTbos = true;
    // const isTbos = window.tnwebServices.customConfigService.get('OpenTbosMode') === '1';
    window.tnwebServices.isTbos = isTbos;
    try {
      const result = newTbosTransform;
      window.tnwebServices.changeApiMap = mapValues(groupBy(result, 'sourcePath'), group => group);
      window.changeApiMap = mapValues(groupBy(result, 'sourcePath'), group => group);
      console.log(window.tnwebServices.changeApiMap, '初始化api map');
    } catch (error) {
      console.log('获取方法map失败');
      window.tnwebServices.changeApiMap = mapValues(groupBy(newTbosTransform, 'sourcePath'), group => group);
      window.changeApiMap = mapValues(groupBy(result, 'sourcePath'), group => group);
      console.log(window.tnwebServices.changeApiMap, '使用本地初始化api map');
    }
      const cacheKey = '__TedgeCacheModuleInfoKey';
      let urlMozuId = '';
      try {
        let moduleInfo = null;
        if (isTbos) {
          let cacheStatus = false;
          try {
            if (localStorage.getItem(cacheKey)) {
              const cacheValue = JSON.parse(localStorage.getItem(cacheKey));
              cacheStatus = keys(cacheValue).length > 0;
            }
          } catch (error) {
            console.log('解析__TedgeCacheModuleInfoKey出错', localStorage.getItem(cacheKey));
            cacheStatus = false;
          }
          urlMozuId = getQueryString('mozuId');
          const mozuInfoParams = {
            "access_type": [
              1
            ],
          }
          if (window.location.hostname.includes('lab')) {
            Object.assign(mozuInfoParams, {
              "set_name_cn": "实验室"
            })
          }
          const result = await Vue.prototype.$axios.post('/cgi/idc-tbos-cgi/Cmdb/GetMozuInfo', {});
          const keyMap = {
            mozu_id: 'mozuId',
            mozu_name: 'mozu',
            mozu_code: 'mozuNumber',
            belong_building: 'building',
            belong_campus: 'park',
          };
          const options = result.list.map((i) => {
            keys(i).forEach((key) => {
              if (has(keyMap, key)) {
                i[keyMap[key]] = i[key];
              }
            });
            return i;
          }).map(i => omit(i, ['id']));
          // 如果url存在mozuId
          if (urlMozuId) {
            const mozuOptions = options.find(i => i.mozuId === Number(urlMozuId));
            console.log(mozuOptions, 'mozuOptions');
            if (mozuOptions) {
              moduleInfo = mozuOptions;
            } else {
              if (options?.length) {
                moduleInfo = options[0];
              } else {
                moduleInfo = {};
              }
            }
          } else {
            if (cacheStatus) {
              moduleInfo = JSON.parse(localStorage.getItem(cacheKey));
              const { mozu_id: cacheMozuId } = moduleInfo;
              const mozuOptions = options.find(i => i.mozuId === Number(cacheMozuId));
              // 如果模组列表和缓存都有则使用缓存
              if (mozuOptions) {
                moduleInfo = JSON.parse(localStorage.getItem(cacheKey));
              } else {
                moduleInfo = {}
              }
            } else {
              if (options?.length) {
                moduleInfo = options[0];
              } else {
                moduleInfo = {};
              }
            }
          }
        } else {
          moduleInfo = await Vue.prototype.$axios.get('/cgi/dataQuery/edge/getEdgeLocation');
        }
        if (!isEmpty(moduleInfo)) {
          const pollingProxyAgentService = new PollingProxyAgentService(
            localStorage.getItem('pollingProxyAgent.moduleDomain') || location.host,
            customConfigService.get('polling-proxy-mode') || 'http',
          );
          window.tnwebServices.pollingProxyAgentService = pollingProxyAgentService;
        }
        Vue.prototype.$moduleInfo = { ...moduleInfo, isTbos };
        localStorage.setItem(cacheKey, JSON.stringify(moduleInfo));

        window.tnwebServices.customConfigService.setModuleId(moduleInfo.mozuId);
        Cookies.set('mozuid', moduleInfo.mozuId);
        if (urlMozuId) {
          const searchParams = new URLSearchParams(window.location.search);
          const params = Array.from(searchParams.keys());
          // 只有模组Id时才替换
          if (params.length === 1 && params[0] === 'mozuId') {
            history.replaceState(null, '', location.pathname);
          }
        }
      } catch (err) {
        Vue.prototype.$moduleInfo = JSON.parse(localStorage.getItem(cacheKey));
      }
    
  } catch (err) {
    console.error(err);
  }

  // const psPlugin = passportPlugin();
  // $(document).ajaxSuccess((event, request, settings, data) => {
  //   try {
  //     if (typeof data === 'string') {
  //       // eslint-disable-next-line no-param-reassign
  //       data = JSON.parse(data);
  //     }
  //     if (data && data.code) {
  //       psPlugin.args.codeHandler.call(vue, data.code);
  //     }
  //   } catch (e) { }
  // });

  window.TNBL.getAllOpCode = getAllOpCode;

  const vue = new Vue({
    el: '#__TNBL__',
    render: h => h(main),
  });

  Vue.config.errorHandler = function (err /* , vm, info */) {
    if (err.code) {
      vue.$message.error(err.message || '发生错误了');
    }
    throw err;
  };

  // 全局捕获promise的reject;
  window.addEventListener('unhandledrejection', (event) => {
    if (event.reason?.code) {
      vue.$message.error(event.reason.message || '发生错误了');

      event.preventDefault();
    }
  }, false);

  // 全局获取opcodeList

  // mock 登录
  // const isCheckLogin = false;
  // function logoutRedirect() {
  //   location.href = `/index.html#/login?r_url=${encodeURIComponent(location.href)}`;
  // }

  // Vue.prototype.$axios = create({
  //   isJson: true,
  //   codeHandler(code) {
  //     if (isCheckLogin && code === 40101) {
  //       localStorage.removeItem('token');
  //       console.log('token 过期，重新登录');
  //       logoutRedirect();
  //       return false;
  //     }
  //   },
  // }, undefined, [
  //   (config) => {
  //     const token = localStorage.getItem('token');
  //     if (token) {
  //       config.headers.Authorization = `Bearer ${token}`;
  //     }
  //     return config;
  //   },
  // ]);
  // let loginInit;
  // const storage = {
  //   pageStatus: 'ok',
  //   loginStatus: '',
  //   account: {
  //     name: Cookies.get('nickname') || '', // 登录名
  //     status: !!Cookies.get('nickname'), // 是否登录
  //   },
  // };
  // Vue.mixin({
  //   beforeCreate() {
  //     Vue.util.defineReactive(this, '$storage', storage);
  //     if (!loginInit) {
  //       const token = localStorage.getItem('token');
  //       if (isCheckLogin && (!token)) {
  //         logoutRedirect();
  //       }

  //       this.$storage.loginStatus = 'ok';
  //       loginInit = true;
  //     }
  //   },
  //   methods: {
  //     logout() {
  //       localStorage.removeItem('token');
  //       logoutRedirect();
  //     },
  //   },
  // });
  // end mock

  // init({
  //   customHttp: true,
  // });

  if (!TNBL.eventBus) {
    TNBL.eventBus = {};
    TNBL.eventBus.dispatch = function () { console.log('dev空函数:eventBus.dispatch', arguments); };
    TNBL.eventBus.addGlobalEventListener = function () { console.log('dev空函数:eventBus.addGlobalEventListener', arguments); };
  }

  /*
按钮级权限开始
// */
  // observe();
  // checkAll();
  /*
按钮级权限结束
*/
  window.__isPadPage = true;
})();

// export default vue;
