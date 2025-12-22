<template>
  <div
    style="position: relative;"
  >
    <div
      ref="itCharts"
      :style="{height:height}"
    />
  </div>
</template>
<script>
export default {
  props: {
    height: {
      type: String,
      default: '300px',
    },
    data: {
      type: Array,
      default: () => [],
    },
    collapseItem: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      lineOption: {},
      itCharts: null,
      itOption: {},
    };
  },
  watch: {
    data: {
      handler() {
        this.init();
      },
      deep: true,
    },
    collapseItem() {
      this.itCharts.resize();
    },
  },
  mounted() {
    this.init();
    window.addEventListener('resize', () => {
      if (this.collapseItem) {
        this.itCharts.resize();
      }
    });
  },
  destroyed() {
    window.removeEventListener('resize', this.itCharts.resize);
  },
  methods: {
    init() {
      this.itCharts = echarts.init(this.$refs.itCharts);
      // this.itCharts.clear();
      this.generateLinedata('lineOption', '告警', this.data, { 告警: '#1470CC' });
      this.itOption = {
        tooltip: {
          trigger: 'axis',
          formatter: (params) => {
            let html = `${params[0].name}<br>`;
            for (let i = 0; i < params.length; i++) {
              html += `<span style="display:inline-block;margin-right:5px;border-radius:10px;width:10px;height:10px;background-color:${params[i].color};"></span>`;
              html += `${params[i].seriesName.split('（')[0]}: <span">${params[i].value}</span><br>`;
            }
            return html;
          },
        },
        title: {
          show: false, // 显示策略，默认值true,可选为：true（显示） | false（隐藏）
          text: '主标题', // 主标题文本，'\n'指定换行
        },
        legend: {
          show: !(this.lineOption.showLegend && this.lineOption.showLegend === 'false'),
          itemHeight: 10,
          itemWidth: 10,
          icon: 'rect',
          data: this.lineOption.legendData || [],
          // right: '20',
          x: 'right',
          // top: '1%',
          y: 'top',

        },

        grid: {
          top: '10%',
          left: '32px',
          right: '17',
          bottom: '10%',
          containLabel: true,
        },
        xAxis: {
          type: 'category',
          data: this.lineOption.xData,
          axisLine: { show: true, lineStyle: { color: '#ccc' } },
          axisPointer: { show: true, type: 'none' },
        },
        yAxis: {
          show: true,
          type: 'value',
          minInterval: 1,
          position: 'left',
          axisLine: { show: true,
            lineStyle: {
              color: '#F0F0F0',
            } },
          axisTick: { show: true,
            lineStyle: {
              color: '#F0F0F0',
            } },
          splitLine: {
            show: true,
            lineStyle: {
              color: '#F0F0F0',
            },
          },
          axisLabel: {
            margin: 10,
            color: '#666666',
            fontSize: 12,
            verticalAlign: 'middle',
            align: 'right',
          },
        },
        series: this.lineOption.series || [],
      };
      this.itCharts.setOption(this.itOption);
    },
    generateLinedata(optionName, name, data, colorMap) {
      this[optionName].series = [];
      this[optionName].xData = data.map(item => item.date);
      this[optionName].legendData = [name];
      this[optionName].showLegend = 'false';
      this[optionName].series.push(this.formatTrendData(data, name, colorMap));
    },
    formatTrendData(data, name, colorMap = {}) {
      const baseSeries = {
        name,
        type: 'line',
        // data: [],
        symbol: 'circle',
        symbolSize: 7,
        smooth: false,
        itemStyle: {
          normal: {
            color: '#5bd9a2',
            label: { show: false,
              position: 'top',
              color: '#1470CC',
              formatter(params) {
                if (params.value == 0) { // 为0时不显示
                  return '';
                }
                return params.value;
              },
            },
          },
        },
        barWidth: 18, // 柱图宽度
        barMinHeight: 2,
        yAxisIndex: 0,
        areaStyle: {
          normal: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              {
                offset: 0,
                color: '#5bd9a226',
              },
              {
                offset: 1,
                color: '#fff',
              }]),
          } },
      };
      if (colorMap[name]) {
        baseSeries.itemStyle.normal.color = colorMap[name];
        baseSeries.areaStyle.normal.color = new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          {
            offset: 0,
            color: `${colorMap[name]}26` || '#5bd9a226',
          },
          {
            offset: 1,
            color: '#fff',
          }]);
      }
      baseSeries.data = data && data.length && data.map((item) => {
        if (item.value === '-') return 0;
        return item.value;
      });
      return baseSeries;
    },
  },
};
</script>
<style lang="scss" scoped>

</style>
