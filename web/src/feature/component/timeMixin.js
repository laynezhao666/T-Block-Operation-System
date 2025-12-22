import { ENV_NAME } from 'common/script/passport_login';
export default {
  beforeCreate() {
    const now = Date.now();
    this.$nextTick(() => {
      this.addTimePoint('page.beforeCreate', now);
    });
  },
  created() {
    const now = Date.now();
    this.$nextTick(() => {
      this.addTimePoint('page.created', now);
    });
  },
  beforeMount() {
    const now = Date.now();
    this.$nextTick(() => {
      this.addTimePoint('page.beforeMount', now);
    });
  },
  mounted() {
    const now = Date.now();
    this.$nextTick(() => {
      this.addTimePoint('page.mounted', now);
    });
    setTimeout(() => {
      this.uploadTime();
      // 不能改对象指针，不然微前端会重置
      // 下次上报的时候只有startPoint开始
      this.clearTime();
    }, 10000);
  },
  methods: {
    addTimePoint(name, t) {
      if (window.timePoints) {
        window.timePoints[name] = t;
      }
    },
    clearTime() {
      Object.entries(window.timePoints).forEach(([key]) => {
        window.timePoints[key] = 0;
      });
    },
    uploadTime() {
      if (!window.timePoints) return;

      // 开始时间:
      // 微前端框架按浏览器加载开始时间算
      // 页面按mfe-loading开始时间算
      const startPoint = 'mfe.mount';
      const framePoint = 'frame.beforeCreate';
      const startTime = window.timePoints[startPoint];
      const isRedirect = !window.timePoints[framePoint];

      if (!startTime) return;

      const timeList = [];

      // 微前端框架的时间都和页面载入时间对比
      // 其他时间和上面的开始时间对比
      Object.entries(window.timePoints).forEach(([name, t]) => {
        if (!t) return;
        if (!isRedirect) {
          // 微前端统计
          // 第一次开始:起点统计
          if (name.indexOf('frame') === 0 || name === startPoint) {
            timeList.push([name, parseInt(t - window.performance.timeOrigin)]);
            return;
          }
        }

        // 其他时间统计(一致)
        if (name !== startPoint) {
          timeList.push([name, parseInt(t - startTime)]);
        }
      });

      timeList.sort(([, duration1], [, duration2]) => duration1 - duration2);

      console.log('耗时', timeList);

      if (process.env.NODE_ENV !== 'production') {
        return;
      }
    },
  },
};
