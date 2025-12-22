<template>
  <div>
    <el-title>
      业务数据查询
    </el-title>
    <el-block
      inner
      no-padding
    >
      <div v-if="urlTabs.length > 1">
        <el-tabs
          v-model="activeTab"
          @tab-click="handleClick"
        >
          <el-tab-pane
            v-for="item in urlTabs"
            :key="item.value"
            :label="item.value"
            :name="item.value"
          />
          <im-frame
            v-if="visible && activeTab !== '服务器告警配置'"
            :key="activeTab+'1'"
            :curtab="activeTab"
            style="height:100%;"
            :handlers="handlers"
            :url="url + ''"
          />
          <page-nebula
            v-if="visible && activeTab === '服务器告警配置'"
            :show-title="false"
          />
        </el-tabs>
      </div>
      <div v-else>
        <im-frame
          v-if="visible"
          :curtab="activeTab"
          style="height:100%;"
          :handlers="handlers"
          :url="url + ''"
        />
      </div>
    </el-block>
  </div>
</template>
<script>
import imFrame from 'feature/component/imFrame/iframeComponent';
import { ENV_NAME } from 'common/script/passport_login';
import 'feature/utils/business';
import pageNebula from 'feature/warning/server-warning-config/main.vue';

const BIDOMAIN = {
  local: {
    PC_ORIGIN: '',
  },
  dev: {
    PC_ORIGIN: '',
  },
  test: {
    PC_ORIGIN: '',
  },
  pre: {
    PC_ORIGIN: '',
  },
  publish: {
    PC_ORIGIN: '',
  },
};
const host = BIDOMAIN && BIDOMAIN[ENV_NAME] && BIDOMAIN[ENV_NAME].PC_ORIGIN;
export default {
  components: {
    imFrame,
    pageNebula,
  },
  props: {
    title: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      urlTabs: [
        { label: '基础设施告警配置',
          url: '/web/monitor/alarm/alarmconfig/index',
          value: '基础设施告警配置',
        },
        { label: '服务器告警配置',
          url: '',
          value: '服务器告警配置',
        },
      ],
      visible: false,
      url: '',
      handlers: {
        test(data) {
          console.log(data);
        },
      },
      activeTab: '',
    };
  },
  mounted() {
    this.activeTab = this.urlTabs[0].value;
    this.url = host + this.urlTabs.find(i => i.value === this.activeTab).url;

    if (window.TNBL.eventBus && window.TNBL.eventBus.addEventListener) {
      window.TNBL.eventBus.addEventListener('mfe-start', this.initReport);
    }
    this.initReport('init', location.href);
  },

  beforeDestroy() {
    if (window.TNBL.eventBus && window.TNBL.eventBus.addEventListener) {
      window.TNBL.eventBus.removeEventListener('mfe-start', this.initReport);
    }
  },
  methods: {
    handleClick() {
      this.url = host + this.urlTabs.find(i => i.value === this.activeTab).url;
    },

    refresh() {

    },
    initReport() {
      const testUrl = localStorage.getItem('testUrl');
      this.url = testUrl || `${this.url}`;
      this.visible = true;
    },
  },
};
</script>
