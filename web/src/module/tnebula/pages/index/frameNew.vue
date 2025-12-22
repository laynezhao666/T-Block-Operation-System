<template>
  <div v-if="showFrame">
    <el-frame-business-new
      v-if="hasAppAuth || hasAppAuthOldData"
      mozu-select-style="cascader"
      :app-container-style="{ 'max-width': '2560px' }"
      :system_flag="systemFlag"
      :mozu_data="mozu_data"
      :menu_data="menu_data"
      :msg_data="msg_data"
      :badge_status="isdot"
      :task_data="task_data"
      :mozu-select-visible="mozuSelectVisible"
      :badge-visible="true"
      :user-dropdown-visible="false"
      :debug="debug"
      :env="env"
      :user_name="user_name"
      :is_micfrontend="is_micfrontend"
      :cur_mozuid="curMozuId"
      :home_url="home_url"
      :todo_type="todoTypeList"
      :fairy_todo_type="todoTypeList"
      switch-user-placeholder="请输入用户名搜索"
      :switch-user-visivle="switchUserVisivle"
      :switch-user-list="switchUserList"
      :update-content="updateContent"
      :top-menu-data="topMenuData"
      :avatar-src="avatarSrc"
      :show-download-app="showDownloadApp"
      :download-progress-visible="true"
      :download-progress-data="downloadProgressData"
      :i-s-y-y-j-z="hasAppAuth"
      :show-side-bar="showSideBar"
      :tnebula-url="tnebulaUrl"
      :tt-url="ttUrl"
      @showUpdate="getUpdateContent"
      @versionUpdateRead="versionUpdateRead"
      @initComplete="initComplete"
      @taskListHandle="taskListHandle"
      @msgListHandle="msgListHandle"
      @updMsgStatusHandle="updMsgStatusHandle"
      @modalClose="modalClose"
      @logoutHandle="logout"
      @mozuChange="mozuChange"
      @menuChange="menuChange"
      @switchUserHandle="switchUserHandle"
      @searchUserHandle="searchUserHandle"
      @downloadBigFile="downloadBigFile"
      @clearDownloadedData="clearDownloadedData"
      @noaccessBack="noaccessBack"
      @addAuth="addAuth"
    >
      <template slot="content" />
      <template slot="customTopLeftBlock">
      </template>
      <template slot="customTopMenu">
        <el-popover
          width="500"
          trigger="click"
        >
          <div class="menu-wrapper">
            <div class="menu-wrapper-top">
              <div class="menu-wrapper-top-title">
                星云1.0应用
              </div>
              <div class="menu-wrapper-top-href">
                {{ nebula1CurHref }}
              </div>
              <div class="menu-wrapper-top-decorate decorate1" />
              <div class="menu-wrapper-top-decorate decorate2" />
            </div>
            <ul class="menu-wrapper-list">
              <li
                v-for="(menu, index) in nebula1Menu"
                :key="index"
                class="menu-item-wrapper"
              >
                <p class="menu-item-title">
                  {{ menu.title }}
                </p>
                <span
                  v-for="(subMenu, subMenuIndex) in menu.subMenu"
                  :key="subMenuIndex"
                  class="menu-item-content"
                  @click="jump(subMenu.href)"
                  @mouseenter="nebula1CurHref = subMenu.href"
                >
                  {{ subMenu.title }}
                </span>
              </li>
            </ul>
          </div>
          <template slot="reference">
            <div
              v-if="showSideBar"
              class="custom-menu-item nebula-one-item"
            >
              星云1.0应用
            </div>
          </template>
        </el-popover>
      </template>
    </el-frame-business-new>
    <el-frame-business
      v-else
      mozu-select-style="cascader"
      :app-container-style="{ 'max-width': '2248px' }"
      :system_flag="systemFlag"
      :mozu_data="mozu_data"
      :menu_data="menu_data"
      :msg_data="msg_data"
      :badge_status="isdot"
      :tt_badge_status="isTTDot"
      :task_data="task_data"
      :show-side-bar="showSideBar"
      :mozu-select-visible="mozuSelectVisible"
      :badge-visible="true"
      :user-dropdown-visible="false"
      :debug="debug"
      :env="env"
      :user_name="user_name"
      :is_micfrontend="is_micfrontend"
      :cur_mozuid="curMozuId"
      :home_url="home_url"
      :todo_type="todoTypeList"
      :fairy_todo_type="todoTypeList"
      switch-user-placeholder="请输入用户名搜索"
      :switch-user-visivle="switchUserVisivle"
      :switch-user-list="switchUserList"
      :show-download-app="showDownloadApp"
      :download-progress-visible="showDownloadProgress"
      :download-progress-data="downloadProgressData"
      :show-t-t="showTT"
      :tnebula-url="tnebulaUrl"
      :tt-url="ttUrl"
      :show-todo-center="showTodoCenter"
      @initComplete="initComplete"
      @taskListHandle="taskListHandle"
      @msgListHandle="msgListHandle"
      @updMsgStatusHandle="updMsgStatusHandle"
      @modalClose="modalClose"
      @logoutHandle="logout"
      @mozuChange="mozuChange"
      @menuChange="menuChange"
      @switchUserHandle="switchUserHandle"
      @searchUserHandle="searchUserHandle"
      @downloadBigFile="downloadBigFile"
      @clearDownloadedData="clearDownloadedData"
    >
      <template #topMenuIcon>
        <!-- <div class="tt-memo-button">
          <el-button
            v-if="showTroubleTicketMemo"
            size="small"
            :round="true"
            type="primary"
            style=""
            @click="jumpToTTMemo"
          >
            <span style="font-size: 16px !important;"> 使用说明 </span>
          </el-button>
        </div> -->
      </template>
      <template slot="content" />
    </el-frame-business>
  </div>
</template>

<script>
// import frameBusinessNew from './component/frame-business-new/src/main.vue';
// import frameBusiness from './component/frame-business/src/main.vue';
import * as reqhel from '../common/reqhel.js';
import * as helper from '../common/helper.js';
import { pageConfig } from '@@/config/page';
import { ENV_NAME } from 'common/script/passport_login';
// import _ from 'lodash';
import ttmenu from './ttmenu.json';
import ttmenuNew from './ttmenu-new.json';
import nebula1Menu from './nebula1menu.js';
import webSocket from 'feature/utils/websocket';
// import ttmenu from './ttmenu.json';
// import appmatrixWhiteList from './appmatrixWhiteList.json';
import { observeNew, checkAllNew } from '../common/newAuth.js';
import { observe, checkAll } from '../common/auth';
import Cookies from 'js-cookie';

export default {
  mixins: [webSocket],
  props: {
    cgi: {
      type: Object,
      default: () => ({}),
    },
    url: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      // useMatrix: 'nouse',
      nebula1Menu,
      nebula1CurHref: nebula1Menu[0].subMenu[0].href,
      topMenuData: [],
      updateContent: [],
      switchUserVisivle: false,
      pageTitle: '',
      switchUserList: [],
      env: ENV_NAME,
      mozuSelectVisible: false,
      home_url: '',
      user_name: this.$storage.account.name,
      systemFlag: 'nebula_fairy',
      curUserMozuData: [],
      mozu_data: [],
      menu_data: [],
      task_data: {},
      msg_data: {},
      debug: false,
      msgStatus: false,
      ttTodoStatus: false,
      taskStatus: false,
      is_micfrontend: true,
      curMozuId: '',
      todoTypeList: [],
      avatarSrc: ``,
      hasAppAuth: false,
      hasAppAuthOldData: false,
      showFrame: false,
      customData: ttmenu,
      customDataNew: ttmenuNew,
      showDownloadApp: false,
      downloadProgressData: [],
      webSocketConfigs: {
        [this.cgi.getDownloadFileList]: {
          dataProcess: ({ data }) => {
            this.downloadProgressData = data;
          },
        },
      },
      showTT: false,
      showSideBar: true,
      logContent: '回到首页',
      tnebulaUrl: '',
      ttUrl: '',
      appmatrixWhiteList: {},
      roleList: [],
      showTodoCenter: true,
      showTroubleTicketMemo: false,
      showDownloadProgress: true,
    };
  },
  computed: {
    isdot() {
      return this.msgStatus || this.taskStatus;
    },
    // curMozuData() {
    //   // eslint-disable-next-line radix
    //   return this.mozu_data.filter(item => item.id === this.curMozuId)[0];
    // },
    isTTDot() {
      return this.ttTodoStatus;
    },
    /**
     * 白名单列表，可根据用户角色判断是否显示
     * 白名单三级菜单无权限，删除三级菜单，二级菜单下无三级菜单，删除二级菜单，一级菜单下无二级菜单，删除一级菜单
     * 白名单三级菜单无权限，删除三级菜单，重置二级菜单href，重置一级菜单href
     */
    whiteMenuList() {
      const { topMenu, customData } = this.appmatrixWhiteList;
      const threeList = customData.filter(e => e.n_level === 3 && e.n_type !== 'all');
      if (threeList.length === 0) {
        return topMenu.filter((e) => {
          if (e.n_id > 2) {
            e.customData = customData;
            // 如果data为空，则不显示
            const { length } = customData.filter(v => v.n_top_menu_id === e.n_id);
            return length > 0;
          }
          return true;
        });
      }
      let curUser;
      if (threeList.some(e => e.n_type === 'user')) {
        curUser = Cookies.get('tnebula_username');
      }
      const menuIds = threeList.filter((item) => {
        if (item.n_type === 'user') {
          // eslint-disable-next-line no-new-func
          const fn = new Function('userList', 'user', `return ${item.n_expression}`);
          return !fn(item.n_show_by_userList, curUser);
        }
        if (item.n_type === 'role') {
          // eslint-disable-next-line no-new-func
          const fn = new Function('appId', 'curAppid', 'roleId', 'curRoleId', `return ${item.n_expression}`);
          return !this.roleList.some(e => item.n_show_by_roleList.some(v => fn(e.appId, v.appId, e.roleId, v.roleId)));
        }
        return true;
      }).map(e => e.n_id);
      // 删除三级应用
      const newCustomData = customData.filter(e => !menuIds.includes(e.n_id));
      const treeList = helper.dataToTree(newCustomData);
      const twoList = treeList.map(e => e.children).flat()
        .filter(e => e.children && e.children.length);
      // 一级应用n_id对应children
      const twoMap = {};
      twoList.forEach((e) => {
        // 重设n_href以及 n_top_menu_id
        e.n_href = e.children[0].n_href;
        e.n_top_menu_id = e.children[0].n_top_menu_id;
        if (twoMap[e.n_pid]) {
          twoMap[e.n_pid].push(e);
        } else {
          twoMap[e.n_pid] = [e];
        }
      });

      const newTreeList = treeList.filter(e => Object.keys(twoMap).includes(`${e.n_id}`));

      newTreeList.forEach((e) => {
        if (twoMap[e.n_id]) {
          e.children = twoMap[e.n_id];
          e.n_href = e.children[0].n_href;
          e.n_top_menu_id = e.children[0].n_top_menu_id;
        }
      });
      const data = helper.flatten(newTreeList);

      topMenu.forEach((e) => {
        if (e.n_id > 2) {
          e.customData = data;
          const row = data.find(v => v.n_level === 3 && e.n_id === v.n_top_menu_id);
          e.n_href = row?.n_href;
        }
      });
      return topMenu.filter((e) => {
        if (e.n_id > 2) {
          // 如果data为空，则不显示
          const { length } = data.filter(v => v.n_top_menu_id === e.n_id);
          return length > 0;
        }
        return true;
      });
    },
  },
  watch: {
    pageTitle(v) {
      if (v) {
        document.title = `${v}-${pageConfig.title}`;
      }
    },
    showTT(v) {
      if (v && this.systemFlag === 'nebula_fairy') this.initTTTodoCnt();
    },
  },
  beforeCreate() {
    window.timePoints = {
      'frame.beforeCreate': Date.now(),
    };
    if (window?.TNBL?.eventBus) {
      // window.TNBL.eventBus.addGlobalEventListener('mfe-loading', () => {
      //   window.timePoints['frame.mfeLoading'] = Date.now();
      // });
      // window.TNBL.eventBus.addGlobalEventListener('mfe-bootstrap', () => {
      //   window.timePoints['frame.mfeLoaded'] = Date.now();
      // });

      window.TNBL.eventBus.addGlobalEventListener('mfe-mount', () => {
        if (window.timePoints['mfe.mount']) {
          // 非第一次就直接重新开始计时
          // 注意：微前端内存管理原因。只能改timePoints属性值，不能重置对象
          Object.entries(window.timePoints).forEach(([key]) => {
            window.timePoints[key] = 0;
          });
        }
        window.timePoints['mfe.mount'] = Date.now();

        setTimeout(() => {
          this.setDomMinHeight();
        }, 10);
      });
      // window.TNBL.eventBus.addGlobalEventListener('mfe-unmount', () => {
      //   window.timePoints['mfe.unmount'] = Date.now();
      // });
    }
  },
  created() {
    window.timePoints['frame.created'] = Date.now();
    this.debugFunc();
    this.initMenu();
    this.initTodoCnt();
    // this.initMsgCnt();
    this.getTaskType();

    if (this.env === 'publish' && window.initAnalysis) {
      try {
        // eslint-disable-next-line no-undef
        initAnalysis(
          {},
          {
            site: 'tidc-om',
          }
        );
        console.log('created数据上报', location.pathname);
      } catch (e) {
        console.error(e);
      }
    }
    // }
    if (this.systemFlag === 'nebula_fairy') {
      this.taskStatusHandle();
    }
    setTimeout(() => {
      // 获取切换权限
      reqhel.hasPrivSwitch().then((data) => {
        if (data && data?.hasPrivSwitch) {
          this.switchUserVisivle = data.hasPrivSwitch;
          this.searchUserHandle();
        }
      });
      // 给logo添加点击跳转首页
      // try {
      //   const btn = document.querySelector('.el-sidebar-header__logo');
      //   btn.onclick = function () {
      //     window.open('/home/#/');
      //   };
      // } catch (error) {
      //   console.log(error);
      // }
    }, 1000);

    // window.setMatrix = function (data) {
    //   localStorage.setItem('useMatrix', data);
    //   location.reload();
    // };
  },
  beforeMount() {
    window.timePoints['frame.beforeMount'] = Date.now();
  },
  mounted() {
    // this.useMatrix = localStorage.getItem('useMatrix');
    window.timePoints['frame.mounted'] = Date.now();

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
  methods: {
    jumpToTTMemo() {
      window.open('/tt/manual');
    },
    jumpToMain() {
      if (this.hasAppAuth) {
        const path = '/home/#/';
        if (this.showSideBar) {
          window.open(path, '_blank');
        } else {
          let prefix = 'http://';
          if (location.protocol === 'https:') {
            prefix = 'https://';
          }
          if (this.env !== 'publish') {
          } else {
          }
        }
      }
    },
    noaccessBack() {
      reqhel.getByUrl({ url: location.pathname }).then((res) => {
        const name = res.n_name;
        TNBL.redirectUrl(`/hr/app-auth-apply?appTitle=${name}`);
      });
    },
    addAuth() {
      TNBL.redirectUrl('/hr/app-auth-apply?type=1');
    },
    jump(href) {
      window.open(href, '_blank');
    },
    getOpcode() {
      return window.TNBL.getAllOpCode();
    },
    versionUpdateRead(data) {
      reqhel.readUpdate({ ids: data }).then(() => {
        this.getUpdateContent({ loading: false });
      });
    },
    getUpdateContent(httpConfig = {}) {
      reqhel.getUpdateContent(httpConfig).then((result) => {
        this.updateContent = result;
      });
    },
    downloadBigFile(fileKey) {
      window.open(`${this.cgi.downloadFile}?file_key=${fileKey}`, '_blank');
    },
    clearDownloadedData(ids) {
      reqhel.clearDownloadedData(ids).then(() => {
        this.$message.success('清除成功');
      })
        .catch(({ message }) => {
          this.$message.error(message);
        });
    },
    async searchUserHandle(keywords = '') {
      await reqhel.getSimpleUserList(keywords, this).then(({ list }) => {
        this.switchUserList = list.map(v => ({
          lable: v.userUid,
          value: `${v.userName}(${v.userRealName})`,
        }));
      });
    },
    switchUserHandle(v) {
      reqhel.switchUser(v.split('(')[0], this);
    },
    debugFunc() {
      this.debug = reqhel.ls('debug') === 'true';
    },
    getTaskType(status) {
      if (!status) { return; };
      reqhel
        .getTaskType({ taskStatus: status }, this.systemFlag)
        .then((r) => {
          if (this.systemFlag === 'nebula_fairy') {
            if (r) this.todoTypeList = r.filter(item => item.type !== '全部');
          } else {
            this.todoTypeList = Object.keys(r).map(e => r[e]);
          }
        });
    },
    setFrameDataByKey(key, val) {
      this.$data[key] = val;
    },

    setPageTitle(keyPath) {
      let menuList = this.menu_data;
      if (this.appmatrixWhiteList.customData) {
        menuList = [...this.appmatrixWhiteList.customData, ...this.menu_data];
      }
      const menuRow = menuList.find(e => e.n_href === keyPath);
      this.pageTitle = menuRow?.n_name || '无权访问';
    },

    menuChange(keyPath, menu) {
      this.setPageTitle(keyPath);
      if (this.hasAppAuth) {
        if (helper.checkMozuSelectVisible(keyPath)) {
          this.initDataScope(keyPath.split('?')[0]);
          return;
        }
        this.initAllOpcodes(keyPath);
      } else {
        reqhel.deleteCookie('tnebula_roleId');
      }
      // 切换菜单变换模组
      helper.initMozuByUrl(keyPath, this);
      TNBL.redirectUrl(keyPath, (menu && menu.$attrs && menu.$attrs.target));
      if (this.env === 'publish') {
        try {
          // eslint-disable-next-line no-undef
          reportPV();
          console.log('切换菜单后的数据上报', location.pathname);
        } catch (e) {
          console.error(e);
        }
      }
    },
    mozuChange(data) {
      helper.setCurMozuCookie(data);
      helper.refreshUrlMozuId(data.id, true);
    },
    initTodoCnt() {
      reqhel.getTaskList({
        taskStatus: 'todo',
        start: 0,
        limit: 1,
      }, this.systemFlag).then((r) => {
        let state = false;
        if (r && r?.count) {
          state = true;
        }
        this.taskStatus = state;
      });
    },
    modalClose() {
      // this.initTodoCnt();
      // this.initMsgCnt();
    },
    initMsgCnt() {
      reqhel.getMsgList({
        status: 0,
        start: 0,
        limit: 1,
      }).then((r) => {
        let state = false;
        if (r && r?.count) {
          state = true;
        }
        this.msgStatus = state;
      });
    },

    initTTTodoCnt() {
      this.getTTStatus();
      setInterval(() => {
        this.getTTStatus();
      }, 10000);
    },

    getTTStatus() {
      reqhel.getTTTodo().then((r) => {
        let state = false;
        if (r && r?.list.length > 0) {
          state = true;
        }
        this.ttTodoStatus = state;
      });
    },

    async initMenu() {
      await this.getOpcode().then((res) => {
        const { opcodeList } = res;
        // this.hasAppAuth = opcodeList.indexOf('TNYYJZ-YYJZKJ-CK') > -1; // 完全用应用矩阵的新ui，同时调用新的权限/数据域接口
        this.hasAppAuthOldData = opcodeList.indexOf('TNYYJZ-YYJZKJ-CK') > -1;// 用应用矩阵的新ui，同时调用老的权限/数据域接口
        this.showDownloadApp = opcodeList.indexOf('TNAPPXZ-EWM-CK') > -1;
        this.showTT = opcodeList.indexOf('TNTT-TTC-CK') > -1;
        this.showFrame = true;
        if (opcodeList.indexOf('TNYYJZ-YYJZTY-CK') > -1) {
          this.hasAppAuthOldData = false;
          this.hasAppAuth = true;
        }
        window.hasAppAuthOldData = this.hasAppAuthOldData;
        window.hasAppAuth = this.hasAppAuth;
        if (window.hasAppAuth) {
          console.log('********开始新auth check');
          observeNew();
          checkAllNew();
        } else {
          console.log('********开始老auth check');
          observe();
          checkAll();
        };
        this.getUpdateContent();
      })
        .catch((e) => {
          this.showFrame = true;
          console.log(e);
        });

        let prefix = 'http://';
      if (location.protocol === 'https:') {
        prefix = 'https://';
      }
      if (this.env !== 'publish') {
        this.tnebulaUrl = ``;
        this.ttUrl = ``;
      } else {
        this.tnebulaUrl = ``;
        this.ttUrl = ``;
      }

      if (location.origin.includes('tt.tidc')) {
        if (this.hasAppAuth) {
          this.menu_data = this.customDataNew;
        } else {
          this.menu_data = this.customData;
        }
        const list = [...this.menu_data].filter(v => v.n_pid === 0)
          .sort((a, b) => a.n_order - b.n_order);
        this.$set(this, 'home_url', list[0].n_href);
        this.showSideBar = false;
        this.showDownloadProgress = false;
        this.showTodoCenter = false;
        this.showTroubleTicketMemo = true;
        this.showDownloadApp = false;
        this.logContent = '返回';
        this.topMenuData = [
          {
            name: 'Trouble Ticket',
            n_href: '/tt/list',
            n_target: '_self',
            Wrap: true,
            origin: 'topMenu',
            child: [],
            customData: this.menu_data,
          },
        ];
        this.initDataScope();
        return;
      }

      reqhel.getMenu({ hasAppAuth: this.hasAppAuth }, this.$storage.account.name)
        .then(async (r) => {
          if (r && r.length) {
            if (this.hasAppAuth || this.hasAppAuthOldData) {
              const list = [...r].filter(v => v.n_pid === 0)
                .sort((a, b) => a.n_order - b.n_order);
              const index = list.findIndex(e => e.n_showtype);
              if (this.hasAppAuth) {
                const data = await reqhel.getMenuWhiteList();
                this.appmatrixWhiteList = data;
                await this.getRoleList();

                const homeUrl = this.whiteMenuList.find(e => e.n_id === 1).n_href;
                this.$set(this, 'home_url', homeUrl);
                const row = this.whiteMenuList.find(e => e.n_id === 2);
                row.n_href = list[index].n_href;
                this.topMenuData = [...this.whiteMenuList];
                this.menu_data = [...r, ...this.appmatrixWhiteList.appList];

                const keyPath = this.formatPath(location.pathname);
                // 没有模组选择器
                if (!helper.checkMozuSelectVisible(keyPath)) {
                  this.initAllOpcodes(keyPath);
                  this.setPageTitle(keyPath);
                  TNBL.redirectUrl(keyPath + location.search);
                  return;
                }
              } else {
                this.$set(this, 'home_url', '/tompage/personal-desktop-new');
                this.menu_data = r;
                this.topMenuData = [
                  { name: '工作台', n_href: '/tompage/personal-desktop-new', n_id: 1, isdot: false, n_target: '_self', origin: 'topMenu', child: [], customData: [] },
                  { name: '应用中心', n_href: list[index].n_href, n_target: '_self', n_id: 2, Wrap: true, origin: 'topMenu', child: [], customData: [] },
                ];
              }
            } else {
              this.menu_data = r.filter(v => !v.n_href.includes('/tt/'));
              const list = [...r].filter(v => v.n_pid === 0)
                .sort((a, b) => a.n_order - b.n_order);
              this.$set(this, 'home_url', list[0].n_href);
            }
            this.initDataScope();
          } else {
          // 沒有任何应用权限，走白名单逻辑
            if (this.hasAppAuth) {
              const data = await reqhel.getMenuWhiteList();
              this.appmatrixWhiteList = data;
              await this.getRoleList();

              const homeUrl = this.whiteMenuList.find(e => e.n_id === 1).n_href;
              this.$set(this, 'home_url', homeUrl);
              this.topMenuData = [...this.whiteMenuList];
              this.menu_data = [...this.appmatrixWhiteList.appList];
            }
            const keyPath = this.formatPath(location.pathname);
            this.setPageTitle(keyPath);
            TNBL.redirectUrl(keyPath);
          }
        });
    },

    async getRoleList() {
      return new Promise((resolve) => {
        const threeList = this.appmatrixWhiteList.customData.filter(e => e.n_level === 3);
        if (threeList.some(e => e.n_type === 'role')) {
          return TNBL.getAllRoleList().then((res) => {
            this.roleList = res;
          });
        };
        resolve();
      });
    },

    findFirstAppId(keyPath) {
      const thirdApp = this.menu_data.find(item => item.n_href === keyPath && item.n_level === 3);
      if (thirdApp) {
        const secondAppId = thirdApp.n_pid;
        const secondApp = this.menu_data.find(item => item.n_id === secondAppId && item.n_level === 2);
        if (secondApp) {
          return secondApp.n_pid;
        }
        return -1;
      }
      return -1;
    },

    async initAllOpcodes(newPath) {
      const pathUrl = newPath || location.pathname || this.home_url;
      const keyPath = this.formatPath(pathUrl);
      this.setPageTitle(keyPath);

      let secondAppId = '';
      const secondApp = this.menu_data.find(item => item.n_href === keyPath && item.n_level === 3);
      if (secondApp) {
        secondAppId = secondApp.n_id;
      }
      // 代表没有权限
      if (secondAppId === '') {
        return TNBL.redirectUrl(newPath || (keyPath + location.search));
      }

      const params = { appID: secondAppId };
      // const dimValueID = reqhel.cs('tnebula_cu_moduleid');
      // if (dimValueID !== '全部' && dimValueID) {
      //   params.dimValueID = parseInt(dimValueID, 10);
      // } else {
      params.dimValueID = 0;
      // }

      reqhel.cs('tnebula_appId', this.findFirstAppId(keyPath));
      localStorage.setItem('getToleAuthTypesParams', JSON.stringify(params));
      const appmatrixWhiteListUrls = this.appmatrixWhiteList.appList.map(i => i.n_href);
      if (!appmatrixWhiteListUrls.includes(keyPath)) {
        await reqhel.GetUserPrivilegByApp(params).then((result) => {
          console.log('**********initAllOpcodes 请求**********');
          reqhel.cs('tnebula_roleId', result.role.roleId);
          this.currentRole = result.role.roleType;
          window.currentRole = result.role.roleType;
        });
      }
    },

    async initDataScope(newPath) {
      const pathUrl = newPath || location.pathname || this.home_url;
      const keyPath = this.formatPath(pathUrl);
      this.setPageTitle(keyPath);

      if (this.hasAppAuth) {
        let secondAppId = '';
        const secondApp = this.menu_data.find(item => item.n_href === keyPath && item.n_level === 3);
        if (secondApp) {
          secondAppId = secondApp.n_id;
        }
        // 代表没有权限
        if (secondAppId === '') {
          return TNBL.redirectUrl(newPath || (keyPath + location.search));
        }

        // 获取模组维度的模组列表
        // let mozuId;
        // await reqhel.getUserDmiMozu({ appID: secondAppId }).then((r) => {
        //   if (r?.module_groups?.length > 0) {
        //     this.curUserMozuData = r.module_groups;
        //     helper.initMozuByUrl(keyPath, this);
        //     mozuId = this.curUserMozuData[0].children[0].children[0].id;
        //     if (process.env.NODE_ENV !== 'production') {
        //       console.log(`initDataSwitchMenu:${keyPath}`);
        //     } else {
        //       console.log(`initDataSwitchMenu:${keyPath}`);
        //       TNBL.eventBus.dispatch('main:CurModuleChange');
        //       TNBL.redirectUrl(newPath || (keyPath + location.search));
        //     }
        //   } else {
        //     TNBL.redirectUrl(newPath || (keyPath + location.search));
        //   }
        // });

        const params = { appID: secondAppId };
        // const dimValueID = reqhel.cs('tnebula_cu_moduleid');
        // if (dimValueID !== '全部') {
        //   params.dimValueID = parseInt(dimValueID, 10);
        // } else {
        //   params.dimValueID = mozuId;
        // }
        reqhel.cs('tnebula_appId', this.findFirstAppId(keyPath));
        localStorage.setItem('getToleAuthTypesParams', JSON.stringify(params));
        const appmatrixWhiteListUrls = this.appmatrixWhiteList.appList.map(i => i.n_href);
        if (!appmatrixWhiteListUrls.includes(keyPath)) {
          await reqhel.GetUserPrivilegByApp(params).then((result) => {
            console.log('**********initDataScope 请求**********');

            if (result?.value?.length > 0) {
              this.curUserMozuData = result.value;
              helper.initMozuByUrl(keyPath, this);
              // mozuId = this.curUserMozuData[0].children[0].children[0].id;
              if (process.env.NODE_ENV !== 'production') {
                console.log(`initDataSwitchMenu:${keyPath}`);
              } else {
                console.log(`initDataSwitchMenu:${keyPath}`);
                TNBL.eventBus.dispatch('main:CurModuleChange');
                TNBL.redirectUrl(newPath || (keyPath + location.search));
              }
            } else {
              TNBL.redirectUrl(newPath || (keyPath + location.search));
            }
            reqhel.cs('tnebula_roleId', result.role.roleId);
            this.currentRole = result.role.roleType;
            window.currentRole = result.role.roleType;
          });
        }
      } else {
        reqhel.deleteCookie('tnebula_roleId');
        TNBL.getScopeModules().then((r) => {
          if (r?.module_groups?.length > 0) {
            this.curUserMozuData = r.module_groups;
            helper.initMozuByUrl(keyPath, this);

            if (process.env.NODE_ENV !== 'production') {
              console.log(`initDataSwitchMenu:${keyPath}`);
            } else {
              console.log(`initDataSwitchMenu:${keyPath}`);
              TNBL.eventBus.dispatch('main:CurModuleChange');
              TNBL.redirectUrl(keyPath + location.search);
            }
          } else {
            TNBL.redirectUrl(keyPath + location.search);
          }
        });
      }
    },
    formatPath(keyPath) {
      // eslint-disable-next-line no-param-reassign
      keyPath = keyPath.trim();
      return (keyPath === '' || keyPath === '/') ? this.home_url : keyPath;
    },
    initComplete() {
      console.log('初始化完成');
    },
    msgListHandle(param) {
      reqhel.getMsgList(param).then((r) => {
        this.msg_data = r;
      });
    },
    taskStatusHandle() {
      setInterval(() => {
        reqhel.taskStatusHandle().then((r) => {
          let state = false;
          if (r && r?.count) {
            state = true;
          }
          this.taskStatus = state;
        });
      }, 10000);
    },
    taskListHandle(param, status) {
      this.getTaskType(param.taskStatus);
      reqhel.getTaskList(param, status).then((r) => {
        if (r.list === null) { r.list = []; };
        this.task_data = r;
      });
    },
    updMsgStatusHandle(param) {
      reqhel.changeStatus(param).then((r) => {
        if (r.succ) {
          reqhel.getMsgList({
            status: 0,
            start: 0,
            limit: 10,
          }).then((r) => {
            this.msg_data = r;
          });
        }
      });
    },

    // 递归计算，设置元素的最小高度
    // 全部用jquery处理
    setDomMinHeight() {
      // const that = this;
      // 是否有非透明背景色
      function hasColor($main) {
        const color = $main.css('backgroundColor');

        if (!color.length || !color.trim() === 'transparent') return false;

        if (color.indexOf('rgba') > -1) {
          return color.replace(/ /g, '').indexOf(',0)') === -1;
        }

        return true;
      }

      // 是否在文档流
      function hasDom($main) {
        if ($main.is('#app_main')) return true;
        const { left, top } = $main.offset();
        return left || top;
      }

      function canSetHeight($main) {
        if ($main.is('.el-title')) return false;
        const { left, top } = $main.offset();
        const height = $main.height();
        const width = $main.width();
        return (left || top) && height && width;
      }

      function findBottomChildrenAndSet($main) {
        if (!hasDom($main)) {
          // console.log('元素定位为0，不再计算', $main);
          return;
        }
        if (hasColor($main)) {
          // console.log('元素有背景色，不再计算', $main);
          return;
        }
        let last = [];
        const { left } = $main.offset();
        const $children = $main.children();
        if ($children.length === 1) {
          const $child = $children.eq(0);
          if (!canSetHeight($child)) {
            // console.log('唯一元素不可设置', $child);
            return;
          }
          if (setMinHeight($main, $child)) {
            findBottomChildrenAndSet($child);
          } else {
            // console.log('父子元素高度不符合设置要求，不再计算(1)', $main, $child);
          }
        } else {
          // 先过滤可见和占区域的元素
          $children.filter((index, child) => canSetHeight($(child))).each((index, child) => {
            // 从中选取最后一个/一排元素作为撑开的元素
            const $child = $(child);
            const childLeft = $child.offset().left;
            if (childLeft <= left) {
              last = [$child];
            } else {
              last.push($child);
            }
          });
          if (!last.length) {
            // console.log('没有满足条件的子元素，不再计算', $main);
          }
          last.forEach(($child) => {
            if (setMinHeight($main, $child)) {
              findBottomChildrenAndSet($child);
            } else {
              // console.log('父子元素高度不符合设置要求，不再计算(2)', $main, $child);
            }
          });
        }
      }

      function setMinHeight($main, $child) {
        const totalHeight = $main.height();
        // const totalHeight = getContentHeight($main);
        let topDistance = $child.offset().top + scrollTop - $main.offset().top;
        topDistance = topDistance < 0 ? 0 : topDistance;
        const childPadding = $child.outerHeight() - $child.height();
        let minChildHeight = totalHeight - topDistance - childPadding;

        // 原先每个页面有标题，不需要padding-top；应用矩阵去掉标题后，需要增加paddng-top: 16px,导致minChildHeight多
        // 减去了16px，所以这里处理下
        if ($('.el-frame-business-new').length && $main.is('#app_main') && !$('.is-hide-main-nav').length) {
          minChildHeight = minChildHeight + 16;
        }

        if ($child.height() <= minChildHeight) {
          $child.css('minHeight', `${minChildHeight}px`);
          $allChildren = $allChildren.add($child);
          // console.log('设置最小高度', $main, $child);
          return true;
        }
        return false;
      }

      // 获取content-box高度
      // function getContentHeight($el) {
      //   const boxSizing = $el.css('boxSizing');
      //   const height = $el.height();
      //   if (boxSizing === 'border-box') {
      //     // border-box的时候，jquery的innerHeight是多算了一次border和padding，刚好用来算差值
      //     return height - ($el.innerHeight() - $el.height());
      //   }
      //   return height;
      // }

      // const removeEltitle = function () {
      //   // if ($('.app-container .el-title').first().length) {
      //   //   $('.app-container .el-title').first()
      //   //     .css('opacity', 0);
      //   // }
      //   // 应用首页不处理
      //   if (that.topMenuData.map(item => item.n_href).includes(location.pathname)) {
      //     return;
      //   }
      //   if ($('.app-container .el-title').first().length) {
      //     console.log($('.el-container #app_main').css('margin-top'), '1');
      //     $('.el-container #app_main').css('margin-top', '-48px');
      //   } else {
      //     _.debounce(() => {
      //       if ($('.el-container #app_main').offset().top < 100) {
      //         $('.el-container #app_main').css('margin-top', '0');
      //       }
      //     }, 1000)();
      //   }
      // };

      function runTimer() {
        if (timer) {
          clearTimeout(timer);
        }
        timer = setTimeout(() => {
          $allChildren.css('minHeight', '');
          findBottomChildrenAndSet($container);
          // removeEltitle();
        }, 100);
      }

      const $container = $('#app_main');
      let $allChildren = $();
      let timer = null;
      let scrollTop = 0;

      const observer = new MutationObserver(() => {
        runTimer();
      });

      observer.observe($container[0], {
        attributes: true,
        // 这里不能有style，不然会导致死循环
        attributeFilter: ['class'],
        attributeOldValue: false,
        characterData: true,
        characterDataOldValue: false,
        childList: true,
        subtree: true,
      });

      // const lateTime = localStorage.getItem('lateTime') || 600;
      // 顶部菜单
      if ($('.top-bar') && $('.top-bar').length) {
        const observerTopMenu = new MutationObserver(() => {
          setTimeout(() => {
            runTimer();
          }, 600);
        });
        const topmenu = $('.top-bar');
        observerTopMenu.observe(topmenu[0], {
          attributes: true,
          attributeFilter: ['class'],
          attributeOldValue: false,
          characterData: true,
          characterDataOldValue: false,
          childList: true,
          subtree: true,
        });
      }

      // 全屏切换
      (new MutationObserver(() => {
        runTimer();
      })).observe($('.app')[0], {
        attributes: true,
        attributeFilter: ['class'],
      });

      window.addEventListener('resize', () => {
        runTimer();
      });

      $container.on('scroll', () => {
        scrollTop = $container.scrollTop();
      });

      runTimer();
    },
  },
};
</script>

<style scoped lang="scss">

body {
  min-width: 1608px;
  min-height: 600px;
}

.el-block .el-collapse-item__header {
  box-sizing: content-box !important;
}

.have-read {
  float: right;
  margin-right: 40px;
  margin-top: 6px;
  font-size: 14px;
  color: #1470CC;
  white-space:nowrap;
  cursor: pointer;
}
.top-bar-icon {
  width: 200px;
  display: flex;
  align-items: center;
  padding-left: 24px;
  box-sizing: content-box;
  cursor: pointer;
}
/deep/.el-sidebar{
    width: 248px;
}
/deep/.el-sidebar--collapsed{
    width: 88px;
}
/deep/ .el-frame-business-new .el-sidebar--collapsed{
     width: 68px;
}
/deep/ .el-sidebar-header__logo {
  // cursor: pointer;
}
/deep/ .el-container #app_main {
  overflow-y: overlay;

  }
/deep/ .el-frame-business-new .el-container #app_main {
  padding: 16px;
  }
/deep/ .el-frame-business-new .is-hide-main-nav #app_main {
  padding: 0 ;
  }
/deep/.el-frame-business-new .app-container .el-title.frame-business-new-title{
  display:flex !important;
  }
/deep/.el-frame-business-new .app-container .el-title.frame-business-new-title.cockpit {
  display: flex !important;
  justify-content: flex-end;
  margin: 0;
  padding: 0;

  .el-title__content {
    display: none;
  }
}
 /deep/ .el-frame-business-new .app-container .el-title {
  display: none;
  }
  .controlshowtopMenu {
      display: none !important;
  }

.custom-menu-item:hover {
  background-color: #437ce7 !important;
  color: rgba($color: #ffffff, $alpha: 1);
}

.nebula-one-item {
    width: 110px !important;
    // padding-left: 18px !important;
    cursor:pointer;
    padding:12px 0;
    text-align:center;
    border-radius: 3px;
    font-family: PingFang SC;
    font-style: normal;
    font-weight: 500;
    color: rgba($color: #ffffff, $alpha: 0.55);
    font-size: 16px;
    margin-left: 5px;
}

.menu-wrapper {
  margin: -12px;

  &-top {
    padding: 24px;
    background: linear-gradient(274.96deg, #ECF2FE 0.85%, rgba(236, 242, 254, 0) 98.28%);
    overflow: hidden;
    position: relative;

    &-title {
      font-size: 20px;
      font-weight: 600;
      color: rgba($color: #000000, $alpha: 0.9);
      margin-bottom: 8px;
    }
    &-href {
      position: relative;
      color: rgba($color: #000000, $alpha: 0.4);
      font-size: 14px;
      height: 24px;
      z-index: 2;
    }

    &-decorate {
      position: absolute;
      width: 140px;
      height: 40px;
      border-radius: 30px;
      transform: rotate(-60deg);
      z-index: 1;

      &.decorate1 {
        top: -8px;
        right: -12px;
        background: linear-gradient(90.84deg, #B7CEFE -14.67%, rgba(148, 184, 255, 0) 119.8%);
      }
      &.decorate2 {
        top: 60px;
        right: 0;
        background: linear-gradient(268.87deg, #FFFFFF -33.84%, rgba(228, 237, 255, 0) 88.32%);
        backdrop-filter: blur(8px);
      }
    }
  }

  &-list {
    padding: 24px;
    display: grid;
    grid-template-columns: 150px 250px;
    grid-column-gap: 50px;
  }

  li.menu-item-wrapper{
    .menu-item-title {
      height: 50px;
      line-height: 50px;
      font-size: 16px;
      font-weight: 600;
      color: rgba($color: #000000, $alpha: 0.9);
    }
    .menu-item-content {
      display: inline-block;
      width: 120px;
      height: 35px;
      line-height: 35px;
      font-size: 16px;
      cursor: pointer;
      color: rgba($color: #000000, $alpha: 0.6);

      &:hover {
        color: #266FE8;
      }
    }
  }
  li:nth-child(2) {
    span:nth-child(odd) {
      text-align: right;
    }
  }
  li:nth-child(3) {
    margin-top: -136px;
  }
}
.tt-memo-button {
  margin-left: 6px;
  /deep/ .el-button {
    padding: 0 8px;
  }
}

</style>

<style lang="scss">
.el-frame-business-new {
  // ::-webkit-scrollbar {
  //   width: 0px !important;
  // }
  .app {
    height: 100vh;
    background: #f6f6f6;
    display:flex;

    &-container {
      min-width: 968px;
      max-width: 1304px;
      margin: 0 auto;
    }

    &-mozu-cascader{
      min-width: 300px;
    }
  }
}
</style>
