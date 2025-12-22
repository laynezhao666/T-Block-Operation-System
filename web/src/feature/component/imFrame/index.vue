<template>
  <div>
    <el-title>
      {{ title }}
    </el-title>
    <el-block>
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
          <keep-alive>
            <im-frame
              v-if="visible"
              :key="activeTab+'1'"
              :curtab="activeTab"
              style="height:100%;"
              :handlers="handlers"
              :url="url + ''"
            />
          </keep-alive>
        </el-tabs>
      </div>
      <div v-else>
        <keep-alive>
          <im-frame
            v-if="visible"
            :curtab="activeTab"
            style="height:100%;"
            :handlers="handlers"
            :url="url + ''"
          />
        </keep-alive>
      </div>
    </el-block>
  </div>
</template>
<script>
import imFrame from './iframeComponent';
import { ENV_NAME } from 'common/script/passport_login';
import 'feature/utils/business';

// const hostName = location.hostname;

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
    // treeComponent,
  },
  props: {
    urlTabs: {
      type: Array,
      default: () => [],
      require: true,
    },
    title: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
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
    // const name = sessionStorage.getItem('currentTab');
    // 判断是否存在currentTab，即tab页之前是否被点击切换到别的页面
    // if (name) {
    //   this.activeTab = name;
    // } else {
    this.activeTab = this.urlTabs[0].value;
    // }
    this.url = host + this.urlTabs.find(i => i.value === this.activeTab).url;

    if (window.TNBL.eventBus && window.TNBL.eventBus.addEventListener) {
      window.TNBL.eventBus.addEventListener('mfe-start', this.initReport);
    }
    this.initReport('init', location.href);
  },
  // beforeRouteLeave(to, from, next) {
  //   // 在离开此路由之后清除保存的状态（我的需求是只需要在当前tab页操作刷新保存状态，路由切换之后不需要保存）
  //   // 根据个人需求决定清除的时间
  //   sessionStorage.removeItem('currentTab');
  //   next();
  // },
  beforeDestroy() {
    // sessionStorage.removeItem('currentTab');
    if (window.TNBL.eventBus && window.TNBL.eventBus.addEventListener) {
      window.TNBL.eventBus.removeEventListener('mfe-start', this.initReport);
    }
  },
  methods: {
    handleClick() {
      // sessionStorage.setItem('currentTab', tab.name);
      window.stop();
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
