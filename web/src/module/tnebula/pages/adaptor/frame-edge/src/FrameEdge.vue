
<template>
  <div
    :class="{ pad: isPad, 'dashboard-park': isPark && !isPad,'frame-edge': !isPark }"
  >
    <Navbar
      v-show="mainNavStatus&&fixed"
      :style="commonStyle"
      :mozu-info="mozuInfo"
      :tab-list="noDataScope ? [] : menuL1"
      :alarm-total="subAlarmTotal"
      :active-index="onePathActive"
      :main-title="mainTitle"
      :is-pad="isPad"
      :is-tbos="isTbos"
      :tbos-mozu-options="tbosMozuOptions"
      :user-name="userName"
      :no-data-scope="noDataScope"
      @onSwitchTabs="menuClick"
      @onFixedModeClick="exitFixedMode"
      @mozuSelectChanged="mozuSelectChanged"
    />
    <Header
      v-if="mainNavStatus&&!fixed"
      :style="commonStyle"
      :logo-url="logoUrl"
      :mozu-info="mozuInfo"
      :main-title="mainTitle"
      :background-url="backgroundUrl"
    >
      <!-- <template v-slot:footer-button>
        <el-button
          type="primary"
          size="small"
        >
          <a
            target="_blank"
            @click="jumpScreen"
          >运营管理中心</a>
        </el-button>
      </template> -->
    </Header>
    <SwitchTabs
      v-show="mainNavStatus&&!fixed"
      :active-index="onePathActive"
      :tab-list="menuL1"
      class="default-tabs"
      :alarm-total="subAlarmTotal"
      @onTabSwitched="menuClick"
    />

    <div
      :class="{ 'frame-edge-main-container': true, 'frame-edge-no-header-container': !mainNavStatus }"
      :style="mainContainerStyle"
    >
      <Sidebar
        v-show="mainNavStatus&&showSidebar"
        :menu-list="menuL2"
        :active-index="thrPathActive"
        :collapsed="collapsed"
        style="margin-right:16px"
        @onMenuSelect="menuClick"
      />
      <el-main
        id="app_main"
        ref="main"
        v-loading.lock="appLoading"
        style="padding:0"
        element-loading-text="加载中"
      >
        <!-- 改为全局监听，不受微前端切换影响，需要考虑切换菜单生效 -->
        <template v-if="isBuilding">
          <building :desc="buildingObj" />
        </template>
        <div
          v-else
          id="container"
          ref="container"
        >
          <div
            id="app"
            ref="app"
          >
            <slot name="content" />
          </div>
        </div>
      </el-main>
    </div>
    <audio
      id="au-alarm-alarm0"
      ref="au-alarm-alarm0"
      muted
      src="/static/audio/alarm0.mp3"
      preload="auto"
    />
    <audio
      id="au-alarm-alarm1to2"
      ref="au-alarm-alarm1to2"
      muted
      src="/static/audio/alarm1to3Edge.wav"
      preload="auto"
    />
    <audio
      id="au-alarm-alarm3to4"
      ref="au-alarm-alarm3to4"
      muted
      src="/static/audio/alarm3to4.mp3"
      preload="auto"
    />

    <polling-proxy-dev-tool />
    <alarms-notifies
      v-if="!noAlarmsNotifies"
      ref="alarmsNotifies"
    />
  </div>
</template>

<script>
import Navbar from './menu/Navbar';
import Sidebar from './menu/Sidebar';
import SwitchTabs from './menu/SwitchTabs.vue';
import noaccess from './built-in/noaccess.vue';
import Header from './components/Header.vue';
import { alarmList, newAlarmList, new1AlarmList } from './mock/mockAm';
import { wsScreen } from '@@/config/api';
import webSocket from 'feature/utils/websocket';
import Cookie from 'js-cookie';
import qs from 'qs';
import PollingProxyDevTool from 'services/polling-request-proxy/dev-tool/index.vue';
import dayjs from 'dayjs';
import AlarmsNotifies from 'feature/adaptor/alarms-notifies/index.vue';
import { cloneDeep, has, keys, omit } from 'lodash';

export default {
  components: {
    Header,
    Navbar,
    Sidebar,
    SwitchTabs,
    noaccess,
    PollingProxyDevTool,
    AlarmsNotifies,
  },
  mixins: [webSocket],
  inheritAttrs: false,

  props: {
    /**
     * 侧边栏只有一个菜单时是否显示，默认不显示
     */
    aloneSideShow: {
      default: false,
      type: Boolean,
    },

    // header背景图片
    backgroundUrl: {
      default: () => '#',
      type: [String, Object],
    },

    collapsed: {
      default: false,
      type: Boolean,
    },
    alarmTotal: {
      default: 0,
      type: Number,
    },
    debug: {
      default: false,
      type: Boolean,
    },
    fixed: {
      default: false,
      type: Boolean,
    },
    homeUrl: {
      default: '',
      type: String,
    },
    isDev: {
      default: false,
      type: Boolean,
    },
    isMicfrontend: {
      default: false,
      type: Boolean,
    },
    logoUrl: {
      default: () => '#',
      type: [String, Object],
    },
    mainTitle: {
      default: '',
      type: String,
    },
    menuData: {
      default: () => [{
        name: '',
        index: '',
        icon: '',
        href: '',
      }],
      type: Array,
    },
    mozuInfo: {
      default: () => ({
        park: '',
        mozu: '',
      }),
      type: Object,
    },
    initComplete: Function,

    isPad: {
      type: Boolean,
      default() {
        return false;
      },
    },
    isPark: {
      type: Boolean,
      default() {
        return false;
      },
    },
    isTbos: {
      type: Boolean,
      default() {
        return false;
      },
    },
  },

  data() {
    return {
      tbosMozuOptions: [],
      levelMapObj: {
        L0: '零级',
        L1: '一级',
        L2: '二级',
        L3: '三级',
        L4: '四级',
      },
      webSocketConfigs: this.isTbos
        ? {
          [wsScreen.tbosAlarm]: {
            dataProcess: (data) => {
              const respData = cloneDeep(data);
              const keyMap = {
                alarm_id: 'AlarmId',
                level: 'Level',
                alarm_name: 'AlarmType',
                device_gid: '3458875016081309696',
                device_number: 'DeviceNumber',
                device_type_zh: 'DeviceType',
                box: 'BoxName',
                room: 'RoomName',
                mozu_name: 'MozuName',
                alarm_content: 'Content',
                occur_time: 'OccurTime',
              };
              respData.data.warn = data.data.warn.map((i) => {
                const item = i;
                Object.keys(i).forEach((j) => {
                  if (Object.prototype.hasOwnProperty.call(keyMap, j)) {
                    item[keyMap[j]] = i[j];
                  }
                });
                return item;
              });
              this.recieveData(respData);
            },
            onConnected: (url) => {
              console.log(`${url} had connected!`);
              const ws = this.webSocketInstances[url].ins;
              const data = {
                cmd: 'getInitDetailList',
                reqid: 0,
                timsstamp: new Date() / 1000,
              };
              setTimeout(() => {
                ws.send(JSON.stringify(data));
              }, 5000);
            },
          },
        }
        : {
          [wsScreen.alarm]: {
            dataProcess: (data) => { this.recieveData(data); },
            onConnected: (url) => {
              console.log(`${url} had connected!`);
              const ws = this.webSocketInstances[url].ins;
              const data = {
                cmd: 'getInitDetailList',
                reqid: 0,
                timsstamp: new Date() / 1000,
              };
              setTimeout(() => {
                ws.send(JSON.stringify(data));
              }, 5000);
            },
          },
        },
      appLoading: false,
      curmenu: {},
      defaultPath: this.homeUrl,
      innerHeight: window.innerHeight,
      isBuilding: false,
      mainNavStatus: true,
      menu: {},
      noaccess: false,
      subAlarmTotal: 0,
      isMock: false, // 是否使用mock数据
      userName: Cookie.get('tnebula_username') || 'common',
      notifiedList: [],
      audioList: [],

      noAlarmsNotifies: !!this.getQueryVariable('no_alarms_notifies') || this.isPark,
      noDataScope: false,
    };
  },

  computed: {
    commonStyle() {
      return { minWidth: this.isPad ? 'unset' : '1280px' };
    },
    menuL1() {
      return this.menu && this.menu[1] && Object.keys(this.menu[1]).map(key => this.menu[1][key]);
    },

    menuL2() {
      const menu2 = this.getChild(this.oneIdActive);
      return Object.keys(menu2).map((key) => {
        if (this.getChildLength(menu2[key].n_id) > 1) {
          const menu = menu2[key];
          menu.children = this.getChild(menu2[key].n_id);
          return menu;
        }
        return menu2[key];
      });
    },

    showSidebar() {
      if (this.aloneSideShow) return true;
      if (!this.menuL2 || this.menuL2.length === 0) return false;
      // 二级菜单多于1个时显示侧边栏
      if (this.menuL2.length > 1) return true;
      // 二级菜单只有1个，但其下有三级子菜单时也显示侧边栏
      return this.menuL2[0] && this.menuL2[0].children && Object.keys(this.menuL2[0].children).length > 0;
    },

    oneIdActive() {
      return this.getCurMenuPropHel('1', 'n_id');
    },
    oneLevelActive() {
      return this.getCurMenuPropHel('1', 'n_level');
    },
    onePathActive() {
      return this.getCurMenuPropHel('1', 'n_href');
    },
    thrPathActive() {
      return this.getCurMenuPropHel('3', 'n_href');
    },
    thrActivePid() {
      return this.getCurMenuPropHel('3', 'pid');
    },
    /**
     * 设置比例尺 原型按宽度按1920尺寸下进行换算
     */
    ratio() {
      return window.innerWidth / 1920;
    },

    /**
     * 窗口高度-减去头部和底部的一些间距
     */
    defaultViewHeight() {
      // return this.innerHeight - (this.ratio * 260) - 80 - 12 - 16;
      return this.innerHeight - (this.ratio * 260) - 80 - 12 - 16;
    },

    /**
     *  窗口高度-减去头部和底部的一些间距
     * */
    fullViewHeight() {
      // return this.innerHeight - 64 - 12 - 32;
      return this.innerHeight - 64 - 12 - 16;
    },

    mainContainerStyle() {
      const { isPad, mainNavStatus } = this;
      let height = (this.fixed ? this.fullViewHeight : this.defaultViewHeight);

      if (!mainNavStatus) {
        height = height + 70;
      }

      return {
        height: isPad ? `${window.innerHeight - 50}px` : `${height}px`,
        width: isPad ? 'auto' : `${this.fixed ? 98 : 95}vw`,
        marginBottom: isPad ? '0' : `${this.fixed ? 0 : 200}px`,
        marginTop: isPad ? '0' : '12px',
        minWidth: isPad ? 'unset' : '1280px',
      };
    },

  },
  watch: {
    isPad: {
      immediate: true,
      handler() {
        document.body.classList[this.isPad ? 'add' : 'remove']('pad');
      },
    },
    isPark: {
      immediate: true,
      handler() {
        document.body.classList[this.isPark ? 'add' : 'remove']('dashboard-park');
      },
    },
    alarmTotal(val) {
      this.subAlarmTotal = val;
    },
    menuData() {
      this.initMenu();
    },
    audioList(list) {
      const that = this;
      if (!list || list.length === 0) return;
      // const [audio] = $(`#au-alarm-${list[0]}`);
      const audio = this.$refs[`au-alarm-${list[0]}`];
      if (audio.paused) {
        [that.currentAudio] = list;
        audio.muted = true;
        const promise = audio.play();
        // TODO 待产品策略
        if (promise) {
          console.log('调用播放');
          promise.catch(() => {
            console.log('未交互无法播放声音');
          });
        }
        audio.muted = false;
      }
      if (!audio.onended) {
        audio.addEventListener('ended', () => {
          that.audioList.shift();
        });
      }
    },
  },

  created() {
    this.checkUrl();
    if (this.isTbos) {
      this.initMozuSelectOptions();
    }
    this.initMenu();
    this.setCurMenu();
    window.addEventListener('resize', this.windowResize);
  },

  mounted() {
    window.__SwitchDebug = () => {
      if (localStorage.getItem('debug') === 'true') {
        localStorage.setItem('debug', false);
      } else {
        localStorage.setItem('debug', true);
      }
    };

    // __resetCurMenu方法绑定到window下面，提供给外部框架调用
    window.__ResetCurMenu = path => this.setCurMenu(path);
    window.__LoadingApp = flag => this.isLoadingApp(flag);
    window.__SwitchMainNavStatus = (status) => {
      this.mainNavStatus = !!status;
    };
    window.__GetFrameDataByKey = key => this.getFrameDataByKey(key);
    window.__SetFrameDataByKey = (key, val) => this.setFrameDataByKey(key, val);

    // 微前端应用加载完成，取消lodaing
    this.$refs.app && this.$refs.app.addEventListener('DOMNodeInserted', () => {
      this.isLoadingApp(false);
    }, false);
    // setInterval(() => {
    //   this.mock();
    //   console.log('mock');
    // }, 5000);
  },

  beforeDestroy() {
    window.removeEventListener('resize', this.windowResize);
  },

  updated() {
    if (this.isMicfrontend) {
      if (document.getElementById('container')) {
        const nHtml = document.getElementById('container').innerHTML;
        if (nHtml === '<div id="app"></div>' && nHtml !== this.div_innerHTML) {
          this._log('reset app html');
          document.getElementById('container').innerHTML = this.div_innerHTML;
        }
      }
    }
  },

  beforeUpdate() {
    if (this.isMicfrontend) {
      if (document.getElementById('container')) {
        this.div_innerHTML = document.getElementById('container').innerHTML;
      }
    }
  },

  methods: {
    async initMozuSelectOptions() {
      try {
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
        const result = await this.$axios.post('/cgi/idc-tbos-cgi/Cmdb/GetMozuInfo', {});
        const keyMap = {
          mozu_id: 'mozuId',
          mozu_name: 'mozu',
          mozu_code: 'mozuNumber',
          belong_building: 'building',
          belong_campus: 'park',
        };
        const tbosMozuOptions = result.list.map((i) => {
          keys(i).forEach((key) => {
            if (has(keyMap, key)) {
              i[keyMap[key]] = i[key];
            }
          });
          return i;
        }).map(i => omit(i, ['id']));
        this.tbosMozuOptions = tbosMozuOptions;
      } catch (error) {
        this.tbosMozuOptions = [];
      } finally {
        // if(!this.tbosMozuOptions.length && !this.isPark && !this.isDev) {
        //   this.noDataScope = true;
        //   this.webSocketConfigs = {};
        //   if (window.location.pathname !== '/tedge/no-datascope') {
        //     window.location.href = '/tedge/no-datascope';
        //   }
        // }
      }
    },
    mozuSelectChanged(mozuId) {
      localStorage.setItem('__TedgeCacheModuleInfoKey', JSON.stringify(this.tbosMozuOptions.find(i => i.mozuId === mozuId) || {}));
      window.tnwebServices.customConfigService.setModuleId(mozuId);
      Cookie.set('mozuid', mozuId);
      window.location.reload();
    },
    // 播放提示声音
    audioPlay(num) {
      const focAudioMap = {
        0: 'alarm0',
        1: 'alarm1to2',
        2: 'alarm3to4',
      };
      const src = focAudioMap[num];
      if (this.audioList.indexOf(src) === -1) {
        this.audioList.push(src);
      };
    },
    mock() {
      this.isMock = true;
      this.initAlarmList(alarmList);
      setTimeout(() => {
        this.initAlarmList(newAlarmList);
      }, 1000);
      setTimeout(() => {
        this.initAlarmList(new1AlarmList);
      }, 2000);
      setTimeout(() => {
        this.initAlarmList(newAlarmList);
      }, 16000);
    },
    recieveData(resData) {
      if (!this.isMock) {
        switch (resData.cmd) {
          case 'getInitDetailList':// 告警初始化
            this.initAlarmList(resData);
            break;
          default:
            // console.log(`other command:${resData.cmd}`);
            break;
        }
      }
    },
    initAlarmList(resData) {
      let canNotifyLevels = ['L0', 'L1', 'L2', 'L3', 'L4'];
      if (!_.isNil(localStorage.getItem('notifyLevels'))) {
        canNotifyLevels = localStorage.getItem('notifyLevels').split(';');
      }
      // let confirmedAlarmIds = [];
      // if (localStorage.getItem(`totalConfirmedAlarmIds${this.userName}`)) {
      //   confirmedAlarmIds = localStorage.getItem(`totalConfirmedAlarmIds${this.userName}`).split(';')
      //     .map(i => i);
      // }
      const notifyPath = ['/tedge/actived-warning'];
      // 播放告警声音
      function countByItem(data, item) {
        const result = data.reduce((acc, value) => {
          if (!acc[value[item]]) {
            acc[value[item]] = 1;
          } else {
            acc[value[item]] = acc[value[item]] + 1;
          }
          return acc;
        }, {});
        return result;
      }
      const levelMap = {
        L0: 0,
        L1: 1,
        L2: 2,
        L3: 3,
        L4: 4,
        L5: 5,
      };

      const audioList = resData.data.warn.sort((a, b) => levelMap[a.Level] - levelMap[b.Level])
        .filter(i => !this.notifiedList.includes(i.AlarmId));
      if (audioList.length) {
        const groupByLevel = countByItem(audioList, 'Level');
        const groupByLevelList = Object.keys(groupByLevel);
        const item = groupByLevelList[0];

        let num = 0;
        const level = item;
        if (level === 'L0') {
          num = 0;
        } else if (level === 'L1' || level === 'L2') {
          num = 1;
        } else if (level === 'L3' || level === 'L4') {
          num = 2;
        }
        this.audioPlay(num);
      }

      // const { v2DeviceNumberTransformerService } = window.tnwebServices;

      const { alarmsNotifies } = this.$refs;

      if (!alarmsNotifies) return;

      const alarmsToNotify = _.orderBy(
        resData.data.warn,
        item => dayjs(item.OccurTime)
          .toDate()
          .getTime()
      ).filter(i => canNotifyLevels.includes(i.Level)
        // if (canNotifyLevels.includes(i.Level)) {
        //   alarmsNotifies.push(i);
        // }

        // const customClass = `tedge-warning-notify-${i.Level}`;
        // if (canNotifyLevels.includes(i.Level) && !confirmedAlarmIds.includes(i.AlarmId.toString())
        // && !this.notifiedList.includes(i.AlarmId) && notifyPath.includes(location.pathname)) {
        //   setTimeout(() => {

        //     this.$notify({
        //       title: `【${this.levelMapObj[i.Level]}】${i.Content}`,
        //       dangerouslyUseHTMLString: true,
        //       message: `<p class="device-area">${v2DeviceNumberTransformerService.get(i.DeviceNumber)}</p>
        //           <p class="time-area">${i.OccurTime}</p>`,
        //       // <p class="time-area">${i.OccurTime.substr(10, i.OccurTime.length - 1)}</p>`,
        //       duration: 0,
        //       customClass,
        //       // type: 'warning',
        //       onClick: () => {
        //         if (location.pathname !== '/tedge/actived-warning') { window.open('/tedge/actived-warning'); }
        //       },
        //       onClose: () => {
        //         const localString = sessionStorage.getItem(`totalConfirmedAlarmIds${this.userName}`) ? sessionStorage.getItem(`totalConfirmedAlarmIds${this.userName}`) : '';
        //         if (localString.indexOf(i.AlarmId) === -1) {
        //           const totalConfirmedAlarmIds = `${i.AlarmId};${localString}`;
        //           sessionStorage.setItem(`totalConfirmedAlarmIds${this.userName}`, totalConfirmedAlarmIds);
        //         }
        //       },
        //     });
        //   }, 0);
        // }
      );

      alarmsNotifies.pushAll(alarmsToNotify);
      this.notifiedList = resData.data.warn.map(i => i.AlarmId);
    },
    jumpScreen() {
      location.href = '/tshows/am?showJump';
    },
    windowResize() {
      this.innerHeight = window.innerHeight;
    },

    isLoadingApp(flag) {
      if (this.isBuilding || this.noaccess) {
        this.appLoading = false;
      } else {
        this.appLoading = flag;
      }
    },

    checkUrl() {
      // 根据url参数隐藏主导航  &_pn_
      this.mainNavStatus = true;
      if (this.getQueryVariable('_pn_') !== false) {
        this.mainNavStatus = false;
        // 添加组合按键退出全屏
        document.onkeydown = (e) => {
          const keyCode = e.keyCode || e.which || e.charCode;
          const { ctrlKey } = e;
          // ctrl+q退出全屏
          if (ctrlKey && keyCode == 81) {
            this.mainNavStatus = !this.mainNavStatus;
            return false;
          }
        };
      }
    },

    // 清理数据
    cleanData() {
      this.curmenu = {};
      this.noaccess = false;
    },

    getFrameDataByKey(key) {
      let rtn;
      switch (key) {
        case 'curMozuData':
          // rtn = { alias: this.mozuInfo.mozu, id: `${this.mozuInfo.mozuId}`, name: this.mozuInfo.mozu };
          rtn = this.mozuInfo;
          break;
        default:
          rtn = this.$data[key];
          break;
      }
      return rtn;
    },

    setFrameDataByKey(key, val) {
      this.$data[key] = val;
      return this.$data[key];
    },

    getQueryVariable(variable) {
      const query = window.location.search.substring(1);
      const vars = query.split('&');
      for (let i = 0; i < vars.length; i++) {
        const pair = vars[i].split('=');
        if (pair[0] == variable) { return pair[1] || ''; }
      }
      return false;
    },

    menuClick(keyPath) {
      this.isBuilding = false;
      if (!this.$refs.container) {
        this.$nextTick(() => {
          this.setCurMenu(keyPath);
          if (!this.reload) {
            console.log('container is not ready');
            this.$emit('menuChange', keyPath);
          }
        });
      } else {
        this.setCurMenu(keyPath);
        if (!this.reload) {
          this.$emit('menuChange', keyPath);
        }
      }
    },

    exitFixedMode() {
      this.$emit('onExitFixedClick');
    },

    formatPath(keyPath) {
      const newKeyPath = (keyPath && keyPath.trim()) || '';
      return (newKeyPath === '' || newKeyPath === '/') ? this.defaultPath : newKeyPath;
    },

    setCurMenu(pathname) {
      this.checkUrl();
      this.reload = false;
      pathname = this.formatPath(pathname);

      // 特殊规则：页内路径、无需参与菜单路径匹配，以##开头
      pathname = pathname.replace(/#inner.*$/, '');

      // 解析参数   菜单配置的链接本身带有参数
      const [newpathname, ...searchParts] = pathname.split('?');
      const searchQueryObj = qs.parse(location.search.replace('?', ''));
      let pagehitflag = false;
      const reg = /^(https|http):\/\//;
      if ((this.menu && this.menu[3] && newpathname !== '')) {
        this.cleanData();

        let lastMatchScope = -1;

        Object.keys(this.menu[3]).forEach((nid) => {
          const n = this.menu[3][nid];
          // 命中菜单
          if (n.n_href.trim() === newpathname || (n.n_href.trim()).indexOf(newpathname) === 0) {
            const splitStr = searchParts.join('?') === '' ? '' : `?${searchParts.join('?')}`;
            const jumpUrl = n.base_url + newpathname + splitStr;

            if (n.base_url && n.base_url.replace(reg, '') !== location.origin.replace(reg, '') && !this.isDev) {
              if (!this.isPad) {
                this.reload = true;
              }

              if (n.n_target === '_self') {
                window.location.href = jumpUrl;
              } else if (n.n_target === '_blank') {
                window.open(jumpUrl);
              }

              return false;
            }

            const menuHrefSearchStr = _.last(n.n_href.split('?'))?.trim();
            const menuHrefSearchQueryObject = menuHrefSearchStr ? qs.parse(menuHrefSearchStr) : null;

            // 跳转更精确匹配查询参数的菜单
            const matchScope = !menuHrefSearchQueryObject
              ? 0
              : _.sumBy(_.toPairs(menuHrefSearchQueryObject), ([k, v]) => {
                const scope = searchQueryObj[k] === v ? 1 : 0;
                return scope;
              });

            if (matchScope <= lastMatchScope) return;

            lastMatchScope = matchScope;

            pagehitflag = true;
            this.curmenu[3] = n;
            this.curmenu[2] = this.menu[2][`id_${n.n_pid}`];
            this.curmenu[1] = this.menu[1][`id_${this.menu[2][`id_${n.n_pid}`].n_pid}`];

            return true;
          }
          // console.log('未命中', n.n_href.trim(), newpathname);
        });
      }
      // 没有命中
      if (!pagehitflag) {
        this.noaccess = true;
      }
      return false;
    },

    // 获取子节点长度
    getChildLength(npid) {
      return Object.keys(this.getChild(npid)).length;
    },

    // 处理二级子节点
    getChild(npid) {
      const childNode = (this.menu.childData && npid && this.menu.childData[npid]) || {};
      const child = {};
      // 过滤是否展示
      Object.keys(childNode).forEach((i) => {
        if (this.isShow(childNode[i])) {
          child[i] = childNode[i];
        }
      });
      return child;
    },

    getCurMenuPropHel(level, propName) {
      return (this.curmenu && this.curmenu[level] && this.curmenu[level][propName]) || '';
    },

    _log(msg, data, fn) {
      if (this.debug || localStorage.getItem('debug') === 'true') {
        if (data) {
          console.log(`[Tframe ${new Date().getTime()}] ${msg}`, JSON.stringify(data, null, 2));
        } else {
          console.log(`[Tframe ${new Date().getTime()}] ${msg}`);
        }

        if (fn && typeof fn === 'function') {
          fn();
        }
      }
    },

    initMenu() {
      const r = this.menuData;
      const menu = {};
      const childData = {};
      r.forEach((node) => {
        // 按照父id归类2，3级儿子
        if (node.n_level === 4 || node.n_level === 0) {
          return true;
        }

        // 按照级别归类
        if (menu[node.n_level]) {
          menu[node.n_level][`id_${node.n_id}`] = node;
        } else {
          menu[node.n_level] = { [`id_${node.n_id}`]: node };
        }

        // 按照pid 归类2级和3级菜单
        if (node.n_level === 1) {
          return true;
        }

        if (childData[node.n_pid]) {
          childData[node.n_pid][`id_${node.n_id}`] = node;
        } else {
          childData[node.n_pid] = { [`id_${node.n_id}`]: node };
        }
      });

      // 重置二级菜单的showtype字段
      if (menu && menu[2]) {
        Object.keys(menu[2]).forEach((i) => {
          const item = menu[2][i];

          item.n_showtype = false;
          childData[item.n_pid][`id_${item.n_id}`].n_showtype = false;

          const child = childData[item.n_id] || {};
          Object.keys(child).forEach((k) => {
            if (this.isShow(child[k])) {
              item.n_showtype = true;
              childData[item.n_pid][`id_${item.n_id}`].n_showtype = true;
            }
          });
        });
      }

      // 重置一级菜单的showtype
      if (menu && menu[1]) {
        Object.keys(menu[1]).forEach((key) => {
          const item = menu[1][key];
          // item.n_showtype = false;
          const child = childData[item.n_id] || {};
          Object.keys(child).forEach((k) => {
            if (this.isShow(child[k])) {
              item.n_showtype = true;
            }
          });
        });
      }

      this.menu = menu;
      this.menu.childData = childData;
      this._initMozu();
    },

    _initMozu() {
      const keyPath = this.formatPath(location.pathname);
      console.log(`初始化开始，执行菜单${keyPath}`);
      this.setCurMenu(keyPath);
      // 初始化完成回调
      this.$emit('initComplete');
    },

    isShow(node) {
      return node && node.n_showtype && this.isHasAccess(node.n_opcode) && !node.n_scope;
    },

    isHasAccess(opcode) {
      if (opcode) { return true; }
      return true;
    },
  },

};
</script>

<style lang='scss' scoped>
  @import "./style/index";
</style>

<style lang="scss">
.frame-edge.pad {
  height: 100vh;
  display: flex;
  flex-direction: column;

  .__dev-side-show-btn {
    display: none;
  }

  .header-container {
    background-color: #142141;
    border-bottom: 1px solid rgba(3,34,77,1);
    box-shadow: 0px 2px 6px 0px rgba(0,0,0,0.4);
    font-size: 20px;
    z-index: 2;

    .header-container ul li {
      font-size: 20px;
    }

    ul {
      height: 50px;
      color: #EDEFF2;
      background-color: #142141;
    }

    .active {
      color: white;
      background-color: #223461;
      font-weight: 600;
    }
  }

  .frame-edge-main-container {
    flex: 1;
    margin: 0;
    background-image: url('/static/pad/images/bg.png');
    background-size: 100% 100%;
  }
}

body.pad {
  overflow: hidden;

  * {
    scrollbar-color: #223460 orange !important;
  }

  .el-block, .el-table, .el-table tr, .el-table th, .el-pagination {
    background: transparent;
    background-color: transparent;
    box-shadow: none;
  }

  .el-table {
    color: white;
    font-size: 12px;

    &::before, .el-table__fixed::before {
      display: none;
    }

    thead {
      background-color: rgba(137,224,255,0.15);
      border-bottom: 1px solid #3B5267;
    }

    th {
      color: white;
      font-weight: 600;
      border: none !important;
    }

    th, td {
      padding: 9px 0;
    }
  }

  .el-table--striped .el-table__body tr.el-table__row--striped td,
  .el-table__body tr.hover-row>td {
    background-color: #223460;
  }

  .el-table td, .el-table th.is-leaf {
    border-color: #18273D;
  }

  .el-table__fixed-right {
    &::before {
      display: none;
    }
  }

  .el-pagination.is-background .btn-next,
  .el-pagination.is-background .btn-next:disabled,
  .el-pagination.is-background .btn-prev,
  .el-pagination.is-background .btn-prev:disabled {
    background-color: #23364C;
    border: none;
    color: white;
  }

  .el-pager {
    .number {
      color: white !important;
      background-color: #23364C !important;
      border: none;

      &.active {
        background-color: #6D7FAB !important;
        border: none;
      }
    }
  }

  .el-pagination .el-select .el-input .el-input__inner,
  .el-input.is-bordered .el-input__inner {
    color: white;
    border: none;
    background-color: #23364C;
  }

  .el-tabs {
    .el-tabs__nav-wrap:after {
      display: none;
    }

    .el-tabs__active-bar {
      background-color: white;
    }

    .el-tabs__item {
      color: #EDEFF2;
      height: 40px;
      line-height: 40px;
      font-size: 14px;

      &.is-active {
        font-weight: 600;
        color: white;
        background-color: #223461;
      }
    }

    .el-tabs--border-card>.el-tabs__header, .el-tabs--border-card {
      background-color: transparent;
      border-bottom: none;
    }
  }

  .el-button--primary {
    background-color: #223460;
    border-color: #234375;
    padding: 4px 24px;
  }

  .el-popover {
    background-color: #041832;
    border-radius: 10px;
    padding: 10 30px;
  }

  .el-checkbox__inner {
    background-color: #23364C;
    border-color: white;
    border-radius: 2px;
  }

  .el-checkbox__input.is-indeterminate .el-checkbox__inner {
    background-color: #23364C;
    border-color: #61CED2;

    &::before {
      background-color: #61CED2;
    }
  }

  .el-checkbox.is-checked {
    vertical-align: top;
    border: 1px solid #61CED2;
    height: 14px;
    width: 14px;
    box-sizing: border-box;
    border-radius: 2px;
    display: block;

    .el-checkbox__input {
      width: 100%;
      height: 100%;
      border: 1.5px solid #23364C;
      display: block;
      box-sizing: border-box;
      margin: 0;
      font-size: 0;
      border-radius: 2px;

      .el-checkbox__inner {
        width: 100%;
        height: 100%;
        background-color: #61CED2;
        border-radius: 1px;
        box-sizing: border-box;
        border: none;

        &::after {
          display: none;
        }
      }
    }
  }

  .el-button--text {
    color: #409EFF;
  }

  .el-button.is-disabled.el-button--icon,
  .el-button.is-disabled.el-button--text {
    color: rgba(204,204,204,0.68);
  }

  .el-table__body-wrapper {
    scrollbar-color: #223460 orange;
  }

  .el-radio-button__inner {
    border-color: #234375;
    background-color: transparent;
  }

  .el-radio-button__orig-radio:checked+.el-radio-button__inner {
    border-color: #234375;
    background-color: #223460;
  }

  .custom-sidebar {
    display: none;
  }

  .el-loading-mask {
    opacity: 0.3;
  }

  .el-modal-body, .el-modal-body__top {
    background-color: #10182a;
    color: white;
    box-shadow: none;
  }

  .el-modal-body {
    box-shadow: 0 0 4px -1px #234375;
  }

  .el-modal-body__top {
    // border-bottom: 1px solid rgba(12,98,161,0.3);
    background-color: #142141;
    box-shadow: 0 2px 6px 0 rgba(0, 0, 0, .4);
  }

  .el-input__inner {
    background-color: #041832;
  }

  .el-modal-body__title {
    color: white;
    font-weight: 600;
  }

  .el-block--border {
    // border-bottom: 1px solid rgba(12,98,161,0.3);
    border-bottom: none;
  }

  .el-button--default {
    background-color: #223460;
    color: white;
    border-color: #234375;
  }

  .el-picker-panel,
  .el-picker-panel__sidebar,
  .el-picker-panel__footer,
  .el-picker-panel .el-input.no-border .el-input__inner {
    background-color: #172B50;
  }

  .el-date-table {
    td.disabled > div {
      background-color: #2C3750;
    }
  }

  // 选中范围中间颜色
  .el-date-table td.in-range div, .el-date-table td.in-range div:hover {
    background-color: #104471;
  }

  // 选中颜色
  .el-date-table td.end-date span, .el-date-table td.start-date span {
    background-color: #0E61A1;
  }

  // 当天颜色
  .el-date-table td.today span {
    color: #61CED2;
  }

  .el-picker-panel,
  .el-picker-panel__sidebar,
  .el-date-range-picker__time-header,
  .el-date-range-picker__header,
  .el-date-range-picker__content.is-left,
  .el-picker-panel__footer,
  .el-time-panel,
  .el-time-panel__footer {
    border-color: #0E61A0;
  }

  .el-time-panel {
    background-color: #101f3b;
    color: white;

    .el-time-panel__content.has-seconds:before,
    .el-time-panel__content.has-seconds:after {
      border-color: #0E61A0;
    }

    .el-time-spinner__item {
      color: white;

      &.active {
        color: #61CED2;
      }
    }

    .el-time-spinner__item:hover:not(.disabled):not(.active) {
      background: transparent;
    }

    .el-time-panel__btn {
      color: #aaaaaa;

      &.confirm {
        color: #409eff;
      }
    }
  }

  .el-range-input, .el-date-editor .el-range-separator {
    color: #CCC;
  }

  .el-date-editor {
    border: 1px solid #234375;
    border-radius: 5px;
    padding: 0 8px;

    &.el-input::before {
      display: none;
    }
  }

  .el-tag--plain {
    background-color: transparent;
  }

  .nav-right {
    padding-right: 8px;
  }

  .el-table .cell,
  .el-table th>.cell {
    padding-left: 16px;
    padding-right: 16px;
  }
}

.el-notification {
  min-height: 80px;
  height: 90px;
  .device-area {
    font-size: 14px;
    margin: 10px 0 0 17px;
    color: #000;
  }
  .time-area {
    font-size: 12px;
    width: 320px;
    text-align: right;
    margin: 3px 15px 0 0;
    color: #000;
  }
}
.el-notification.tedge-warning-notify-L0 {
    padding: 16px 0 0 0px;
  .el-notification__group {
    margin-left: 0px;
  }
  .el-notification__title {
    margin-left: 8px;
    color: #ff3e00;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    width: 300px;
  }
  // background-color: #ff3e00;
  cursor: pointer;
}
.el-notification.tedge-warning-notify-L1 {
    padding: 16px 0 0 0px;
  .el-notification__group {
    margin-left: 0px;
  }
  .el-notification__title {
    margin-left: 8px;
    color: #ff3e00;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    width: 300px;
  }
  // background-color: #ff3e00;
    // background-color: #878787;
  // background-color: #878787;

  cursor: pointer;

}
.el-notification.tedge-warning-notify-L2 {
    padding: 16px 0 0 0px;
  .el-notification__group {
    margin-left: 0px;
  }
  .el-notification__title {
    margin-left: 8px;
    color: #ff9200;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    width: 300px;
  }
  // background-color: #878787;
    cursor: pointer;

}
.el-notification.tedge-warning-notify-L3 {
    padding: 16px 0 0 0px;
  .el-notification__group {
    margin-left: 0px;
  }
  .el-notification__title {
    margin-left: 8px;
    color: #fbd743;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    width: 300px;
  }
    // background-color: #878787;

  // background-color: #fbd743;
    cursor: pointer;

}
.el-notification.tedge-warning-notify-L4 {
    padding: 16px 0 0 0px;
  .el-notification__group {
    margin-left: 0px;
  }
  .el-notification__title {
    margin-left: 8px;
    color:#008adc;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    width: 300px;
  }
  // background-color: #008adc;
    cursor: pointer;

}

#app_main {
  position: relative;

  &.dark-marks {
    &::before {
      content: '';
      display: block;
      width: 100%;
      height: 100%;
      position: absolute;
      top: 0;
      left: 0;
      background-color: rgba(0,0,0,0.6);
      z-index: 1001;
    }
  }
}
</style>
