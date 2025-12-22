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
import dateUtil from 'element-ui/src/utils/date';

import { tooltipRest, gridRest, xAxisLabelRest, yAxisLabelRest, optionRest, styles } from './props';
import { getText } from './utils';
import mixin from './mixin';

const colors = ['#1470cc', '#37d1d1', '#ffc751'];

const xAxisRest = {
  type: 'category',
  boundaryGap: false,
  axisTick: { show: false },
  axisLine: { show: false },
  axisPointer: {
    lineStyle: {
      type: 'dashed',
    },
    z: -1,
  },
};

const yAxisRest = {
  type: 'value',
  axisTick: { show: false },
  axisLine: { show: false },
};

const seriesRest = {
  type: 'line',
  showSymbol: false,
  symbol: 'circle',
  smooth: true,
};

function formatDate(date, format) {
  date = new Date(date);
  if (isNaN(date.getTime())) {
    return date;
  }
  return dateUtil.format(date, format);
}

export default {
  name: 'CustomizeLineChart',
  mixins: [mixin],
  props: {
    height: {
      type: Number,
      default: 274,
    },
    xAxis: {
      type: Object,
      required: true,
    },
    yAxis: {
      type: [Object, Array],
    },
    series: {
      type: Array,
      default: () => [],
      required: true,
    },
    tooltip: {
      type: Object,
      default: () => ({}),
    },
    legend: {
      type: Object,
      default: () => ({}),
    },
  },
  watch: {
    series() {
      this.render();
    },
    xAxis() {
      this.render();
    },
  },
  methods: {
    render() {
      const { xAxis, yAxis = {}, series, tooltip } = this;
      const { mode = 'datetime' } = xAxis;

      const fSeries = series.map((item, i) => {
        const rst = {
          color: [colors[i]],
          ...seriesRest,
          ...item,
        };
        if (item.areaStyle === true) {
          rst.areaStyle = {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
              offset: 0,
              color: color(rst.color[0]).alpha(0.15)
                .toString(),
            }, {
              offset: 1,
              color: color(rst.color[0]).alpha(0)
                .toString(),
            }]),
          };
        }
        return rst;
      });
      const options = {
        ...optionRest,
        tooltip: {
          ...tooltipRest,
          position: tooltip.position,
          formatter(params) {
            let obj; let detail;
            const units = Array.isArray(yAxis) ? yAxis.map(i => i.unit) : yAxis.unit;
            if (Array.isArray(params)) {
              obj = params[0];
              if (tooltip.ignoreNil) {
                params = params.filter(item => item.data);
              }
              detail = params.map((item) => {
                let unit;
                if (Array.isArray(units)) {
                  const curSeries = series[item.seriesIndex];
                  const yIndex = curSeries.yAxisIndex || 0;
                  unit = units[yIndex] || '';
                } else {
                  unit = units || '';
                }
                if (!unit) {
                  const curSeries = series[item.seriesIndex];
                  unit = curSeries.unit || '';
                }
                let value;
                if (tooltip.formatter) {
                  value = tooltip.formatter(item.data, item);
                } else {
                  value = item.data;
                }
                return `<div>${item.marker} <span>${item.seriesName}: ${getText(value, unit)}</span></div>`;
              }).join('');
            } else {
              const unit = units || '';
              obj = params;
              let value;
              if (tooltip.formatter) {
                value = tooltip.formatter(params.data);
              } else {
                value = params.data;
              }
              detail = `<div>${params.marker} <span>${params.seriesName}: ${getText(value, unit)}</span></div>`;
            }
            let title;
            if (mode === 'datetime') {
              title = formatDate(obj.name, 'MM-dd HH:mm:ss');
            } else if (mode === 'date') {
              title = formatDate(obj.name, 'MM-dd');
            } else if (mode === 'time') {
              title = formatDate(obj.name, 'HH:mm');
            }
            return `
            <div style="${styles.tooltipTitle}">${title}</div>
            ${detail}
            `;
          },
        },
        grid: {
          ...gridRest,
          bottom: mode === 'datetime' ? '24px' : '12px',
        },
        legend: {
          ...this.legend,
        },
        xAxis: {
          ...xAxisRest,
          ...this.xAxis,
          axisLabel: {
            ...xAxisLabelRest,
            formatter(value) {
              const str = formatDate(value, 'MM-dd HH:mm:ss');
              const [date, time] = str.split(' ');
              if (mode === 'datetime') {
                return [`{value|${date}}`, `{value|${time}}`].join('\n');
              } if (mode === 'date') {
                return `{value|${date}}`;
              } if (mode === 'time') {
                return `{value|${time}}`;
              }
              return str;
            },
          },
        },
        yAxis: Array.isArray(yAxis) ? yAxis.map(i => ({
          ...yAxisRest,
          ...i,
          axisLabel: {
            ...yAxisLabelRest,
            formatter(value) {
              return `${value}${i.unit || ''}`;
            },
          },
        })) : {
          ...yAxisRest,
          ...yAxis,
          axisLabel: {
            ...yAxisLabelRest,
            formatter(value) {
              return `${value}${yAxis.unit || ''}`;
            },
          },
        },
        series: fSeries,
      };

      this.chart.setOption({
        ...options,
      });
    },
  },
};
</script>
