import { redirectToLogin } from 'common/script/passport_login';
import Vue from 'vue';
import qs from 'qs';

let interval = 0;
const watchScrollInterval = 0;

// 这里的this是./index.vue
export default {
  _message({ method, args }) {
    const result = this[method](...args);
    // 可能会返回vue实例，vue实例不能给到postMessage，会报错
    if (!(result instanceof Vue)) {
      return result;
    }
  },
  flowKey() {
    // this.callback({
    //   event: 'setStyle',
    //   result: getStyleByFlowKey(v),
    // });
  },
  // 需要跳转登录
  login() {
    redirectToLogin();
  },
  // 已加载成功（bpm内时机还不好，待调整）
  loaded() {
    this.setStyle({
      visibility: 'visible',
    });
  },
  addTimePoints(times) {
    if (window.timePoints) {
      this.$nextTick(() => {
        Object.entries(times).forEach(([name, t]) => {
          window.timePoints[name] = t;
        });
      });
    }
  },
  addTimePoint(name, t) {
    if (window.timePoints) {
      this.$nextTick(() => {
        window.timePoints.points[name] = t;
      });
    }
  },
  // bpm页面高度变化监听
  // 改名成下划线的，这个后期废弃
  resize({ height }) {
    // 通信协议问题，先在重复这里调用，保证一定能可见
    // this.setStyle({
    //   visibility: 'visible',
    // });
    this.setStyle({
      // 增加一点高度，防止出滚动条，挤压页面宽度，反复触发高度变化，导致页面反复抖动
      height: `${parseInt(height, 10) + 35}px`,
    });
  },
  // 用于BPM内部获取url的参数
  // 方便表单脚本做一些自定义的处理功能
  getParams() {
    return new Promise((resolve) => {
      try {
        const { params } = qs.parse(location.search, { ignoreQueryPrefix: true });
        resolve(qs.parse(params || ''));
      } catch (e) {
        console.log(e);
        resolve({});
      }
    });
  },
  _scrollTo({ top }) {
    this.scrollContainer.scrollTop = top;
  },
  _iframeBtns(btns) {
    this.iframeBtns = btns;
  },
  _iframeDialogVisible(disabled) {
    this.iframeDialogVisible = disabled;
  },
  // 将外部滚动高度传到内部使用
  _watchScroll() {
    if (watchScrollInterval) {
      window.clearInterval(watchScrollInterval);
    }
    let lastScrollTop;
    window.setInterval(() => {
      if (lastScrollTop === this.scrollTop) return;
      lastScrollTop = this.scrollTop;

      this.callback({
        event: '_watchScroll',
        result: {
          scrollTop: lastScrollTop,
        },
      });
    }, 500);
  },
  _loading({ loading }) {
    this.loading = loading;
  },
  // bpm流程完成（跳转时）
  // 废弃 ，已改用redirectFlowPage直接强制跳转
  finish({ type }) {
    this.$emit('finish', { type });
    // 500s不处理直接跳转详情页
    // 使用om内跳转url，防止刷新报错
    setTimeout(() => {
      if (this.instId) {
        location.search = `?instId=${this.instId}`;
      }
    }, 500);
  },
  // bpm流程id
  // 废弃，改用redirectFlowPage，保证om侧的跳转可刷新
  instId(instId) {
    if (!this.instId) {
      this.$emit('update:instId', instId);
    }
  },
  // 废弃，改用redirectFlowPage，保证om侧的跳转可刷新
  draftId(draftId) {
    if (!this.instId) {
      this.$emit('update:instId', draftId);
    }
    if (!this.draftId) {
      this.$emit('update:draftId', draftId);
    }
  },
  // 废弃，统一用nodeInfo
  nodeId(nodeId) {
    this.$emit('update:nodeId', nodeId);
  },
  /**
   * 包含（可能空）：instId/taskId/nodeId/nodeName/nodeType
   */
  nodeInfo(nodeInfo) {
    this.$emit('update:nodeName', nodeInfo.nodeName);
    this.$emit('update:nodeId', nodeInfo.nodeId);
  },
  // 发请求
  request({ method, url, data }) {
    if (!method || !url) {
      console.warn('缺少参数');
      return;
    }
    return this.$axios[method.toLowerCase()](url, data);
  },
  // 跳转链接
  link({ url, target = 0 }) {
    if (target) {
      window.open(url);
    } else {
      TNBL.redirectUrl(url);
    }
  },
  redirectFlowPage({ type, id }) {
    let newSearch = '';
    // 不靠谱，会无效
    // TNBL.redirectUrl(`?${type}Id=${id}`);
    if (type !== 'start') {
      newSearch = `?${type}Id=${id}`;
    }
    window.location.search = newSearch;
  },
  // 长效通信测试
  connectTest(data) {
    console.log('receive:connect', data);
    const that = this;

    let index = 0;
    if (data && data.clear) {
      clearInterval(interval);
    }
    if (!interval) {
      interval = setInterval(() => {
        that.callback({
          event: 'connectTest',
          result: {
            index,
            data,
          },
        });
        index += 1;
      }, 1000);
    }
  },
  /**
   * 单次通信测试
   * data: {type: 1，2，3，code: number}
   */
  msgTest(data, callbackId) {
    console.log('receive:msg', data);
    if (data && data.type) {
      if (data.type === 1) {
      // 直接返回
        return data;
      } if (data.type === 2) {
      // 一定时间后返回
        return new Promise((resolve) => {
          setTimeout(() => {
            resolve(data);
          }, 1000);
        });
      } if (data.type === 3) {
      // 一定时间后返回错误
        setTimeout(() => {
          this.dispatchCallback(callbackId, {
            code: data.code || -1,
            data,
          });
        }, 1000);
      }
    }
  },
};
