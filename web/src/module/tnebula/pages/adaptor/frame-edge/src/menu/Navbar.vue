<template>
  <div class="header-container">
    <div
      class="brand-container"
    >
      <!-- <span
        class="brand-title"
        @click="btnClickHandler"
      >{{ mainTitle }}</span> -->
      <template v-if="!isPad">
        <span
          class="brand-title"
          @click="btnClickHandler"
        >{{ mozuInfo.blockId || mozuInfo.mozu || mainTitle }}</span>

        <el-select
          v-if="isTbos"
          v-model="tbosMozuId"
          class="mozu-select"
          filterable
          placeholder=""
          :no-input="showNoInput"
          @change="mozuSelectChange"
          @visible-change="mozuSelectVisibleChange"
        >
          <el-option
            v-for="item in tbosMozuOptions"
            :key="item.mozuId"
            :label="item.mozu"
            :value="item.mozuId"
          />
        </el-select>
      </template>

      <img
        v-if="isPad"
        src="/static/pad/images/site-logo.png"
        class="pad-logo"
      >

      <!-- <span class="brand-mozu">{{ mozuInfo.mozu }}</span> -->
    </div>
    <ul>
      <template v-for="item in tabList">
        <li
          v-if="item.n_showtype"
          :key="item.n_id"
          :class="item.n_href===activeIndex ? 'active' : ''"
          @click="switchTabHandler(item)"
        >
          <div class="tab-title-wrapper">
            <span class="tab-title-text">
              <template v-if="item.n_name === '监控告警' || item.n_name === '告警记录'">
                <template v-if="totalAlarm > 0">
                  <el-badge
                    :value="totalAlarm"
                    :max="99"
                    class="item"
                  >
                    <div style="width:80px">{{ item.n_name }}</div>
                  </el-badge>
                </template>
                <template v-else>
                  {{ item.n_name }}
                </template>
              </template>
              <template v-else>
                {{ item.n_name }}
              </template>
            </span>
          </div>
        </li>
      </template>
    </ul>

    <div
      :class="{ pad: isPad }"
      class="nav-right"
    >
      <el-dropdown
        v-if="!isPad"
        class="header-dropdown ml-2"
        @command="handleUserLoginCommand"
      >
        <span class="el-dropdown-link">
          {{ userName }} <i class="tn-icon-arrow-down" />
        </span>
        <el-dropdown-menu slot="dropdown">
          <el-dropdown-item command="logout">
            退出
          </el-dropdown-item>
        </el-dropdown-menu>
      </el-dropdown>
    </div>
    <log-modal :visible.sync="modalVisible" />
    <system-dialog
      v-if="!isPad"
      :visible.sync="dialogVisible"
    />

    <reset-password-modal
      ref="resetPasswordModal"
    />
  </div>
</template>
<script>
import LogModal from '../components/LogModal';
import SystemDialog from '../components/systemDialog';
import ResetPasswordModal from '../components/reset-password-modal.vue';
import PadBell from '../components/pad-bell.vue';
import Cookies from 'js-cookie';
import dayjs from 'dayjs'

export default {
  name: 'Navbar',
  components: {
    LogModal,
    SystemDialog,
    ResetPasswordModal,
    PadBell,
  },
  props: {
    mozuInfo: {
      type: Object,
      default() {
        return {};
      },
    },
    tabList: {
      type: [Object, Array],
      default() {
        return {};
      },
    },
    activeIndex: {
      type: String,
      default: '',
    },
    alarmTotal: {
      type: Number,
      default: 0,
    },
    mainTitle: { type: String, default: '腾讯TBOS动环平台' },
    isPad: {
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
    tbosMozuOptions: {
      type: Array,
      default() {
        return [];
      },
    },
    userName: {
      type: String,
      default: '',
    },
    noDataScope: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      showNoInput: true,
      tbosMozuId: this.$moduleInfo?.mozuId || 464,
      alarmBoradCast: null,
      totalAlarm: 0,
      alarmTimer: null,
      modalVisible: false,
      dialogVisible: false,

      loginStatus: window.tnwebServices.loginStatusService,
      isFullScreen: false,
      checkTimer: null
    };
  },
  computed: {
  },
  watch: {
    alarmTotal(val) {
      this.totalAlarm = val;
    },
  },
  mounted() {
    this.checkTimer = setInterval(()=>{
      this.checkCookieExpire()
    }, 10000)
  },
  beforeDestroy(){
    clearInterval(this.checkTimer)
    this.checkTimer = null
  },
  methods: {
    async loginAdmin() {
      this.loginStatus.login();
    },
    logoutAdmin() {
      this.loginStatus.logout();
    },

    switchTabHandler(item) {
      this.$emit('onSwitchTabs', item.n_href);
    },
    btnClickHandler() {
      this.$emit('onFixedModeClick');
    },
    toggleFullScreen() {
      if (this.isFullScreen) {
        document.exitFullscreen();
        this.isFullScreen = false;
      } else {
        document.body.requestFullscreen();
        this.isFullScreen = true;
      }
    },
    handleCommand(command) {
      if (command === 'log') {
        this.modalVisible = true;
      } else if (command === 'system') {
        this.dialogVisible = true;
      } else if (command === 'help') {
      } else if (command === 'feedback') {
      }
    },
    checkCookieExpire(){
      const targetTime = Cookies.get('tnebula_expire')
      const currentTime = dayjs();
      const target = dayjs(targetTime);
      const diffMs = target.diff(currentTime);
      if (diffMs < 0) {
        this.logout()
      } 
    },
    handleUserLoginCommand(command) {
      const func = ({
        logout: () => this.logout(),
      })[command];

      if (func) {
        func();
      }
    },
    handleAdminDropdownCommand(command) {
      const func = ({
        logout: () => this.loginStatus.logout(),
        resetPassword: () => {
          this.$refs.resetPasswordModal.show();
        },
      })[command];

      if (func) {
        func();
      }
    },
    mozuSelectChange() {
      this.$emit('mozuSelectChanged', this.tbosMozuId);
    },
    mozuSelectVisibleChange(val) {
      this.showNoInput = !val;
    },
  },
};
</script>

<style lang="scss" scoped>
.header-container {
  display: flex;
  background: #1470cc;
  justify-content: space-between;
  .brand-container {
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
    color: #fff;
    margin-right: 48px;
    cursor: pointer;
    .brand-title {
      font-weight: bold;
      font-size: 24px;
      white-space: nowrap;
      margin: 0 24px;
    }
    .brand-mozu {
      font-size: 14px;
      white-space: nowrap;
    }
  }
  ul {
    display: flex;
    height: 64px;
    background: #1470cc;
    margin: 0;
    justify-content: flex-start;
    list-style: none;
    align-items: center;
    li {
      display: flex;
      min-width:96px;
      height: 100%;
      font-size: 16px;
      color: rgba(#fff, 0.87);
      text-align: center;
      cursor: pointer;
      flex: 1;
      align-items: center;
      &:hover{
        background: #036bd3;
      }
    }
    .active {
      color: #fff;
      border-bottom: 4px solid #fff;
    }
  }

  .tab-item {
    overflow: hidden;
  }

  .tab-title-wrapper {
    display: flex;
    margin: auto;
    align-items: center;
  }

  .nav-right{
    display: flex;
    flex: 1;
    min-width: 128px;
    align-items: center;
    justify-content: flex-end;
    font-size: 14px;
    padding:0 16px 0 48px ;
    color: rgba(#fff,.87);
    cursor: pointer;
    flex-wrap: nowrap;
    white-space: nowrap;
    &-exitBtn{
      margin: auto 0;
      display: flex;
      align-items: center;
    }

    &.pad {
      padding-right: 12px;
    }

    .icon{
      padding: 2px;
      opacity: 0.87;
    }
     & :hover{
        background: #1b5ea1;
      }
  }
  .log-icon {
    margin-left: 12px;
    margin-right: 12px;
    .el-dropdown-link {
      color: #fff;
    }
  }
}

.brand-container {
  .mozu-select {
    /deep/ .el-select__text span {
      display: none;
    }
  }

}

.tn-icon-logout {
  color: #fff;
  opacity: 1;
}

.tn-icon-login {
  color: #fff;
  opacity: 0.3;
}

.header-dropdown {
  color: #FFFFFF !important;
  text-align: right;
  &:hover {
    background: none !important;
  }
}

.fullscreen-btn {
  /deep/ {
    i {
      font-size: 10px;
      color: #e0e0e0;
      border: 1.5px solid #e0e0e0;
      border-radius: 100px;
      width: 18px;
      height: 18px;
      line-height: 18px;
      font-weight: 600;
    }
  }
}

.pad-logo {
  height: 27px;
  margin-left: 24px;
}
</style>
