import * as echarts from 'echarts';

export default {
  mounted() {
    this.chart = echarts.init(this.$refs.chart);
    window.chart = this.chart;
    this.render();
    window.addEventListener('resize', this.chart.resize);
  },
  destroyed() {
    window.removeEventListener('resize', this.chart.resize);
  },
  watch: {
    height() {
      setTimeout(this.chart.resize);
    },
  },
};
