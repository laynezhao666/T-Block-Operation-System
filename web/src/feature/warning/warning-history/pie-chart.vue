<template>
  <div>
    <div
      id="echarts"
      ref="itCharts"
      :style="{height:height}"
    />
  </div>
</template>
<script>
// import { numberWithCommas } from '@@/utils/utils.js';
export default {

  props: {
    data: {
      type: Array,
      default: () => [],
    },
    height: {
      type: String,
      default: '250px',
    },
    series: {
      type: Object,
      default: () => ({
        center: [150, 125],
        radius: ['45%', '50%'],
      }),
    },
  },
  data() {
    const that = this;
    return {
      optionData: [],
      totalNum: '',
      itCharts: null,
      itOption: {
        tooltip: {
          trigger: 'item',
        },
        title: {
          text: this.series.title,
          left: this.series.maintitleLeft || 90,
          top: this.series.titleTop || '50%',
          textStyle: {
            color: '#333333',
            fontSize: 18,
            align: 'center',
          },
        },
        graphic: {
          type: 'text',
          left: this.series.titleLeft,
          top: this.series.graphicTop || '45%',
          style: {
            text: this.series.graphicTitle,
            textAlign: 'center',
            fill: 'gray',
            fontSize: '100em',
            fontWeight: 700,
          },
        },
        // color: [
        //   '#ff9200',
        //   '#ffb20c',
        //   '#09cccb',
        //   '#156fcc',
        //   'red',
        // ],
        legend: {
          icon: 'circle',
          x: 'right',
          y: 'center',
          orient: 'vertical',
          align: 'left',
          itemGap: 15,
          type: 'scroll',
          pageIconColor: '#75b7fa',
          pageIconSize: 12,
          formatter(name) {
            const { data } = that;
            let total = 0;
            let tarValue = 0;
            for (let i = 0; i < data.length; i++) {
              total += parseFloat(data[i].value);
              if (data[i].name == name) {
                tarValue = parseFloat(data[i].value);
              }
            }
            const v = tarValue;
            const p = total ? Math.round(((tarValue / total) * 100)) : 0;
            const result = `${name}  ${v}   (${p}%)`;
            // if (result.length > 24) {
            //   result = `${result.slice(0, 22)}\n${result.slice(22, result.length)}`;
            // }
            if (that.series.showLegendLabel) {
              return `${result}`;
            }
            return `${name}`;
          },
          tooltip: {
            show: true,
          },
        },
        series: [
          {
            center: ['40%', '50%'],
            type: 'pie',
            radius: this.series.radius || ['40%', '60%'],
            avoidLabelOverlap: true,
            label: {
              show: true,
              textStyle: {
                fontSize: 8,
                color: 'inherit',
              },
              formatter(param) {
                return `${param.name}: ${param.percent}%`;
              },
            },
            labelLine: {
              length: 8,
              length2: 10,
            },
            data: this.data,
          },
        ],
      },
    };
  },
  watch: {
  },
  mounted() {
    this.setOp();
  },
  destroyed() {
    window.removeEventListener('resize', this.itCharts.resize);
  },
  methods: {
    setOp() {
      // this.optionData = this.data;
      this.itCharts = echarts.init(this.$refs.itCharts);
      if (!this.checkData(this.data)) {
        this.itOption.color = ['#ffffff'];
      }
      this.$nextTick(() => {
        this.itCharts.setOption(this.itOption);
      });
      window.addEventListener('resize', this.itCharts.resize);
    },
    checkData(data) {
      if (data && data.length) {
        const hasValueLength = data.filter(item => (item.value !== '0' && item.value !== '0.00')).length;
        if (hasValueLength > 0) {
          return true;
        }
        return false;
      }
    },
  },
};
</script>
<style lang="scss" scoped>

</style>
