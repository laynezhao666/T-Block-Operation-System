<template>
  <el-card
    style="position: relative;"
  >
    <div
      ref="itCharts"
      class="echarts"
      :style="{height:height}"
      style="min-height:500px;height:80vh;width:100%"
    />
  </el-card>
</template>
<script>

export default {
  props: {
    height: {
      type: String,
      default: '100%',
    },
    data: {
      type: Object,
      default: () => ({
        itps: [],
      }),
    },
    nearWarning: {
      type: Number,
      default: 100000,
    },
    overWarning: {
      type: Number,
      default: 100000,
    },
  },
  data() {
    return {
      chartData: this.data,
      itCharts: null,
      itOption: {
        tooltip: {
          trigger: 'axis',
          axisPointer: { 
            type: 'shadow', 
          },
        },
        barWidth: 17,
        legend: {
          itemHeight: 10,
          itemWidth: 10,
          icon: 'rect',
          data: [],
          y: 'top',
          x: 'right',
          itemGap: 50,
        },
        label: {
        },
        grid: {
          top: '10%',
          left: '0%',
          right: '5%',
          bottom: '1%',
          containLabel: true,
        },
        yAxis: {
          type: 'value',
          axisLabel: {
            color: '#666666',
          },
          axisLine: {
            lineStyle: {
              color: '#cccccc',

            },
          },
        },
        xAxis: {
          type: 'category',
          data: [],

          axisLine: {

            lineStyle: {
              color: '#cccccc',
            },
          },
          axisLabel: {
            interval: 0,
            rotate: 40,
            margin: 20,
            color: '#666666',
          },
          realtimeSort: true,
        },
        series: [
          {
            name: '总功率',
            type: 'bar',
            stack: 'total',
            label: {
              show: true,
            },
            emphasis: {
              focus: 'series',
            },
            data: [320, 302, 301, 334, 390, 330, 320, 320, 302, 301, 334, 390, 330, 320, 320, 302, 301, 334],
          },
        ],
      },
    };
  },
  watch: {
    data: {
      handler() {
        this.setOp();
      },
      deep: true,
    },
  },
  mounted() {
    this.setOp();
    window.addEventListener('resize', this.itCharts.resize);
  },
  destroyed() {
    window.removeEventListener('resize', this.itCharts.resize);
  },
  methods: {
    resizeChart() {
      this.itCharts.resize();
    },
    setOp() {
      this.itOption.xAxis.data = this.data.pointXaxis || [];
      this.itOption.series[0].data = this.data.itps || [];
      this.itOption.tooltip.formatter = (params) => {
        let relVal = params[0].name.split('\n')[1];
        for (let i = 0, l = params.length; i < l; i++) {
          relVal += `<br/>${params[i].marker}${params[i].name.split('\n')[0]} : ${params[i].value} ${this.data.pointYaxisUnit[params[i].dataIndex]}`;
        }
        return relVal;
      };
      this.itOption.series[0].itemStyle = {
        color() { 
          return '#1470CC';
        },
      };
      this.itCharts = echarts.init(this.$refs.itCharts);
      this.itCharts.setOption(this.itOption);
    },    
  },
};
</script>
<style lang="scss" scoped>

</style>
