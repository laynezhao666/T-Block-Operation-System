<template>
  <div
    ref="chart"
    class="chart-standard-wrap"
    :style="{height: `${height}px`}"
  />
</template>

<script>
import * as echarts from 'echarts';
import color from 'color';
import mixin from './src/mixin';
import * as props from './src/props';

const barWidth = 16;
const lineStyle = {
  color: '#f0f0f0',
};
const colors = ['#1470CC', '#67A4E0', '#C9DFF5'];

export default {
  name: 'TnBarChartTedge',
  mixins: [mixin],
  props: {
    max: Number,
    shadow: Boolean,
    gradients: Boolean,
    series: {
      type: Array,
      default: () => [],
    },
    cates: {
      type: Array,
      default: () => [],
    },
    direction: {
      type: [String, Number],
      default: 'horizontal',
    },
    title: String,
    legend: Boolean,
    labelRotate: {
      type: Array,
      default: () => [45, 45],
    },
  },
  computed: {
    height() {
      if (this.direction === 'horizontal') {
        if (this.series.length) {
          return (((this.series.length * this.series[0].data.length) * barWidth) * 2) + 60;
        }
        return 0;
      } if (this.direction === 'vertical') {
        return 300;
      }
      return this.direction;
    },
  },
  watch: {
    data() {
      this.render();
    },
    max() {
      this.render();
    },
    series() {
      this.render();
    },
  },
  methods: {
    render() {
      const { cates } = this;
      const { labelRotate } = this;

      const valueAxis = {
        axisLine: {
          show: true,
          lineStyle,
        },
        splitLine: {
          lineStyle,
        },
        axisTick: {
          show: false,
        },
        axisLabel: {
          ...props.xAxisLabelRest,
          rotate: labelRotate[0],
        },
        max: this.max,
      };

      const cateAxis = {
        data: cates,
        axisLabel: {
          ...props.yAxisLabelRest,
          rotate: labelRotate[1],
        },
        axisTick: {
          show: false,
        },
        axisLine: {
          show: true,
          lineStyle: {
            color: '#f0f0f0',
          },
        },
        z: 9,
      };

      const option = {
        grid: {
          ...props.gridRest,
        },
        tooltip: {
          ...props.tooltipRest,
          axisPointer: {
            lineStyle: {
              color: 'rgba(0, 0, 0, 0)',
            },
          },
          formatter(params) {
            const lines = params.map((param) => {
              const series = option.series[param.seriesIndex];
              return `
                <div>
                  <span style="display:inline-block; margin-right:5px; border-radius:10px; width:9px; height:9px; background-color:${series.color}"></span>
                  <span>${param.seriesName}: ${param.data}${series.unit || ''}</span>
                </div>
              `;
            });
            return `
              <div style="${props.styles.tooltipTitle}">${params[0].name}</div>
              ${lines.join('')}
            `;
          },
        },
        series: this.series.map((series, i) => ({
          color: colors[i % 3],
          ...series,
        })),
      };

      if (this.title) {
        option.title = {
          text: this.title,
          textStyle: {
            fontWeight: 'normal',
            fontSize: '14',
          },
        };
      }

      if (this.legend) {
        option.legend = {
          data: option.series.map(series => series.name),
          right: '16px',
          icon: 'circle',
          textStyle: {
            color: '#999999',
          },
        };
      }

      if (this.title || this.legend) {
        option.grid.top = '60px';
      }

      if (this.direction === 'horizontal') {
        option.xAxis = valueAxis;
        option.yAxis = cateAxis;
        // 平行柱状图要控制柱宽
        option.series = option.series.map(series => ({
          ...series,
          barWidth,
        }));
      } else {
        option.yAxis = valueAxis;
        option.xAxis = cateAxis;
      }

      if (this.gradients) {
        option.series = option.series.map((series) => {
          const baseColor = color(series.color);
          return {
            ...series,
            type: 'bar',
            itemStyle: {
              normal: {
                color: new echarts.graphic.LinearGradient(
                  1, 0, 0, 0,
                  [
                    { offset: 0, color: color({ h: baseColor.hue(), s: 100, l: 94 }).toString() },
                    { offset: 1, color: color({ h: baseColor.hue(), s: 100, l: 98 }).toString() },
                  ]
                ),
                borderWidth: 1,
                borderColor: color({ h: baseColor.hue(), s: 67, l: 73 }).toString(),
              },
            },
          };
        });
      } else {
        option.series = option.series.map((series) => {
          const baseColor = series.color;
          return {
            ...series,
            type: 'bar',
            itemStyle: {
              normal: {
                color: baseColor,
              },
            },
          };
        });
      }

      if (this.shadow) {
        if (option.series.length > 1) {
          throw new Error('shadow 模式下 series 的长度不能超过 1');
        }
        const { data } = option.series[0];
        const max = this.max || Math.max(...data);
        const unit = option.series[0].unit || '';
        option.series = [
          { // For shadow
            type: 'bar',
            itemStyle: {
              color: '#f5f5f5',
              emphasis: { color: '#f5f5f5' },
            },
            barGap: '-100%',
            barCategoryGap: '40%',
            data: new Array(data.length).fill(max),
            animation: false,
            barWidth,
            z: 2,
          },
          {
            ...option.series[0],
            z: 3,
          },
        ];
        option.tooltip.formatter = (params) => {
          const data = params[1];
          return `
              <div>
                <span style="display:inline-block; margin-right:5px; border-radius:10px; width:9px; height:9px; background-color:#8CBAE8;"></span>
                <span>${data.name}: ${data.data}${unit}</span>
              </div>
            `;
        };
      }

      option.tooltip.textStyle.color = '#ffffff';

      this.chart.setOption(option);
    },
  },
};
</script>
