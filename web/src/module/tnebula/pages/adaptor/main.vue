<template>
  <FrameEdge
    v-if="homeUrl"
    :menu-data="menu_data"
    :site-title="urlQuery.title"
    :fixed="isFixed"
    :is-micfrontend="is_micfrontend"
    :main-title="mainTitle"
    :home-url="homeUrl"
    :mozu-info="mozuInfo"
    :alarm-total="alarmTotal"
    :is-pad="isPad"
    :is-park="isPark"
    :is-tbos="isTbos"
    :is-dev="isDev"
    collapsed
    @onExitFixedClick="exitFixedClickHandler"
    @menuChange="menuChange"
    @initComplete="initComplete"
  />
</template>

<script>
import _ from 'lodash';
import qs from 'qs';
import FrameEdge from '@/module/tnebula/pages/adaptor/frame-edge/index.js';
import mockMenu from './mockMenu.js';
import * as reqhel from '../common/reqhel.js';
import { pageConfig } from '@@/config/page';
import { ENV_NAME } from 'common/script/passport_login';
import getEdgeRequest from 'feature/utils/request';
// import { memoriedFetchBlockDevices } from 'feature/adaptor/pad/block/utils/block-devices.ts';
import { AlarmsCountWatcher } from 'services/tedge/data-watchers/alarms';
import { TboxModeAlarmsCountWatcher } from 'services/tedge/data-watchers/tbox-mode-alarms';


// 由于tnfusion不支持 npm run dev tedge pad/block/xxxx 这种方式，只能改为 pad-block-xxx，这是历史原因
// 更合适的是 pad/block/xxx 这种组织方式，有机会需要推送tnfusion支持然后改掉

export default {
  components: { FrameEdge },
  data() {
    const isPad = false;
    const isPark = false;
    const { isTbos } = window.tnwebServices;

    // const isPark = true;
    return {
      env: ENV_NAME,
      isDev: ENV_NAME === 'dev' || ENV_NAME === 'local',
      menu_data: [],
      is_micfrontend: true,
      isFixed: true,
      homeUrl: null,
      mozuInfo: {},
      mozuId: '',
      alarmBoradCast: null,
      alarmTotal: 0,

      urlQuery: qs.parse(window.location.search.replace(/^\?/, '')),

      isPad,
      isPark,
      isTbos,

      loginStatusService: window.tnwebServices.loginStatusService,

      mainTitle: '腾讯TBOS动环平台',

      alarmsCountWatcher: new AlarmsCountWatcher(3000),
      tboxModeAlarmsCountWatcher: new TboxModeAlarmsCountWatcher(3000),
    };
  },
  watch: {
    isFixed(value) {
      this.dispatchResize();
      if (value) {
        localStorage.setItem('NAV_FIXED', true);
      } else {
        localStorage.removeItem('NAV_FIXED');
      }
    },
  },

  created() {
    this.initMenu();
    window.addEventListener('scroll', _.debounce(this.onWindowScroll, 100, { maxWait: 300 }));
    if (!this.isPad && localStorage.getItem('NAV_FIXED')) {
      this.isFixed = true;
    }
    const beaconReportDisable = window.tnwebServices.customConfigService.get('BeaconReportDisable')?.trim();
    if (this.env === 'publish' && beaconReportDisable !== '1' && window.initAnalysis) {
      try {
      // eslint-disable-next-line no-undef
        initAnalysis({
        }, {
          site: 'tidc-b',
        });
        console.log('created数据上报', location.pathname);
      } catch (e) {
        console.error(e);
      }
    }
  },

  mounted() {
    document.title = pageConfig.title;
    // eslint-disable-next-line no-underscore-dangle
    window.__SwitchDebug = () => {
      if (reqhel.ls('debug') === 'true') {
        reqhel.ls('debug', false);
      } else {
        reqhel.ls('debug', true);
      }
      this.debugFunc();
    };
    // 切换到老框架
    // eslint-disable-next-line no-underscore-dangle
    window.__ChangeFrame = () => {
      reqhel.ls('useOldFrame', true);
      location.reload();
    };
  },
  beforeDestroy() {
    // this.alarmBoradCast.close();
    this.alarmsCountWatcher.cancel();
    this.tboxModeAlarmsCountWatcher.cancel();
  },
  methods: {
    async watchWarningNumber() {
      const alarmPayload = {
        eventStatus: 1,
        limit: 1,
        mozuId: this.mozuId,
        start: 0,
      };

      if (this.isPad && this.mozuInfo.blockId) {
        // if (!this.deviceNumbers) {
        //   const {
        //     deviceNumbers,
        //   } = await memoriedFetchBlockDevices(this.mozuInfo.blockId);
        //   this.deviceNumbers = deviceNumbers;
        // }
        // alarmPayload.DeviceNumber = this.deviceNumbers;
      }
      this.alarmsCountWatcher.watch({}, (count) => {
        this.alarmTotal = count;
      });
    },
    dispatchResize() {
      const myEvent = new Event('resize');
      window.dispatchEvent(myEvent);
    },
    initMenu() {
      const { isPad, isPark } = this;
      let menuData = mockMenu.data.menus // 接口就位前，菜单先使用mock数据
      menuData = menuData;

      if (isPad) {
        const queryCurrent = qs.parse(window.location.search.replace(/^\?/, ''));
        const mergeSearch = (href) => {
          if (!href) return href;
          const hrefSplited = href.split('?');
          const queryHref = href.includes('?')
            ? qs.parse(_.last(hrefSplited))
            : {};

          return `${hrefSplited[0]}?${qs.stringify({
            ...queryCurrent,
            ...queryHref,
          })}`;
        };
        menuData = _.map(menuData, item => ({
          ...item,
          n_href: mergeSearch(item.n_href),
        }));

        this.homeUrl = menuData[0].n_href;
      } else {
        this.homeUrl = window.tnwebServices.customConfigService.get('home_page') || '/tedge/overview';
      }

      this.menu_data = menuData;
      // this.homeUrl = this.menu_data[0].n_href;
      this.initDataScope();
      this.getEdgeLocation();
    },

    initGlobalWarningNumber() {
      this.watchWarningNumber();
    },
    async getEdgeLocation() {
      const data = this.$moduleInfo;

      this.mozuInfo = {
        id: data.mozuId.toString(),
        blockId: this.urlQuery.blockId,
        name: data.mozu,
        alias: data.mozu,
        ...data,
      };
      this.mozuId = data.mozuId;
      localStorage.setItem('tidc_tedge_mozuId', data.mozuId);
      this.initGlobalWarningNumber();
    },

    initComplete() {
      console.log('初始化完成');
    },

    formatPath(keyPath) {
      // eslint-disable-next-line no-param-reassign
      keyPath = keyPath.trim();
      return (keyPath === '' || keyPath === '/') ? this.homeUrl : keyPath;
    },

    initDataScope() {
      const keyPath = this.formatPath(location.pathname);
      // if (process.env.NODE_ENV !== 'production') {
      //   console.log(`initDataSwitchMenu:${keyPath}`);
      //   console.log('非生产环境');
      // } else {
      //   console.log('生产环境');
      //   console.log(`initDataSwitchMenu:${keyPath}`);
      //   TNBL.eventBus.dispatch('main:CurModuleChange');
      //   TNBL.redirectUrl(keyPath + location.search);
      // }
      console.log(`initDataSwitchMenu:${keyPath}`);
      // eslint-disable-next-line babel/no-unused-expressions
      TNBL.eventBus && TNBL.eventBus.dispatch && TNBL.eventBus.dispatch('main:CurModuleChange');
      // eslint-disable-next-line babel/no-unused-expressions
      TNBL.redirectUrl && TNBL.redirectUrl(keyPath + location.search);
    },

    lowEnough() {
      const pageHeight = Math.max(document.body.scrollHeight, document.documentElement.offsetHeight);
      const viewportHeight = window.innerHeight
                || document.documentElement.clientHeight
                || document.body.clientHeight || 0;
      const scrollHeight = window.pageYOffset
                || document.documentElement.scrollTop
                || document.body.scrollTop || 0;
      return pageHeight - viewportHeight - scrollHeight < 10; // 通过 真实内容高度 - 视窗高度 - 上面隐藏的高度 < 20，作为加载的触发条件
    },

    onWindowScroll() {
      if (this.isFixed) return;
      if (this.lowEnough()) {
        this.isFixed = true;
      };
    },

    exitFixedClickHandler() {
      if (this.isPad) {
        this.menuChange(this.homeUrl);
        return;
      }

      this.isFixed = false;
    },

    menuChange(keyPath, menu) {
      console.log(window.location.origin + keyPath);
      TNBL.redirectUrl(keyPath, (menu && menu.$attrs && menu.$attrs.target));
      if (this.env === 'publish') {
        try {
          // eslint-disable-next-line no-undef
          if (window.reportPV) {
            window.reportPV();
          }
          console.log('切换菜单后的数据上报', location.pathname);
        } catch (e) {
          console.error(e);
        }
      }
    },
  },
};
</script>

<style >
</style>
