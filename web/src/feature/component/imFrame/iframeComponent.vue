<template>
  <!-- loading:该loading给report内部事件使用，保证能覆盖到整个页面 -->
  <div
    class="im-frame"
  >
    <!--
      initLoading：
      此loading仅加载时使用，保证页面空白时dom的loading效果
      局部loading放到container的内部，它会生成到元素内部最后一个子元素
      以保证container能作为外层div的最后一个子元素
      可以被正确设置最小撑开高度
    -->
    <!-- v-loading="initLoading" -->

    <div
      id="im-container"
      ref="imContainer"
      v-loading="initLoading"
      element-loading-text="加载中"
      element-loading-spinner="el-icon-loading"
    />
  </div>
</template>
<script>
import { map, each } from 'lodash';
import mixin from './mixin';
// import { ENV_NAME } from 'common/script/passport_login';

/**
 * 通过props传函数绑定回调
 * 具体见mixin
 */
export default {
  mixins: [mixin],
  props: {
    url: {
      type: [String],
      default: '',
    },
    // 注册事件
    handlers: {
      type: Object,
      default: () => ({}),
    },
    curtab: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      scrollContainer: null,
      // 禁止滚动
      iframeDialogVisible: false,
      iframeBtns: [],
      initLoading: true,
      // 用于内部触发，block整个页面
      loading: false,
    };
  },
  computed: {
    style() {
      return {
        width: '100%',
      };
    },
  },
  watch: {
    url() {
      this.initFrame();
    },
  },
  mounted() {
    this.initFrame();
  },
  methods: {
    initFrame() {
      if (!this.url) {
        return;
      }
      this.initLoading = true;
      setTimeout(() => {
        this.initLoading = false;
      }, 400);
      this.initHandshake({
      // Element to inject frame into
        container: document.getElementById('im-container'),
        // Page to load, must have postmate.js. This will also be the origin used for communication.
        url: this.url,
        // Set Iframe name attribute. Useful to get `window.name` in the child.
        name: 'imIframe',
        // Classes to add to the iframe via classList, useful for styling.
        classListArray: ['nebula-fairy-iframe'],
      }).then((child) => {
        this.initLoading = false;

        child.on('idcloaded', () => {
          this.initLoading = false;
        });
        this.addHandler('_system', () => {
          this.bindSystemEvent();
        });
        map(this.handlers, (fn, eventName) => {
          this.addHandler(eventName, fn);
        });
        each(this.style, (val, key) => {
        // eslint-disable-next-line no-param-reassign
          child.frame.style[key] = val;
        });
      })
        .catch((e) => {
          this.initLoading = false;
          this.$confirm('页面加载失败，请刷新重试').then(() => {
            window.location.reload();
          });
          console.log('握手失败', e);
        });
    },
    setStyle(styles) {
      this.handshake.then((child) => {
        each(styles, (val, key) => {
        // eslint-disable-next-line no-param-reassign
          child.frame.style[key] = val;
        });
      });
    },
    bindSystemEvent() {
      window.addEventListener('click', (e) => {
        if (!e.target.closest('#im-container')) {
          this.callback({
            event: '_system',
            result: {
              type: 'blur',
            },
          });
        }
      });
      this.initScrollEvent();
    },
    initScrollEvent() {
      this.scrollContainer = document.getElementById('app_main');
      this.scrollContainer.addEventListener('scroll', (e) => {
        const { scrollTop } = e.target;
        this.callback({
          event: '_watchScroll',
          result: scrollTop,
        });
        // 上面要废弃，统一为下面这个
        // 先并存，后续去掉_watchScroll事件
        this.callback({
          event: '_system',
          result: {
            type: 'scroll',
            scrollTop,
          },
        });
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.im-frame {
  box-shadow: rgba(218,218,218,0.5) 3px 3px 8px 2px;
  background-color: #fff;
  min-height: 400px;
  // overflow: hidden;

  /deep/ iframe {
    width: 100%;
    min-height:calc(100vh - 260px);
    // margin-top: -51px;
    // margin-top: -35px;
  }
}
</style>
