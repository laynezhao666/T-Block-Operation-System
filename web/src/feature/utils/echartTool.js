import * as echarts from 'echarts';
export default function getOptions(data, options) {
  const objData = data.map((ele) => {
    const keys = Object.keys(ele);
    let value = ele[keys[0]];
    if (options.fixTimes) {
      value = value * options.fixTimes;
    }
    return {
      time: keys[0] && keys[0].split(' ')[1],
      value,
    };
  });
  const serie = {
    type: 'line',
    data: objData.map(item => item.value),
    // symbol: 'none',
    areaStyle: '',
  };
  if (options.series) {
    if (options.series[0].color[0] === '#0acccc') {
      serie.areaStyle = {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [{
            offset: 0, color: 'rgba(10,204,204,0.3)', // 0% 处的颜色
          }, {
            offset: 0.5, color: 'rgba(10,204,204,0.2)', // 100% 处的颜色
          }, {
            offset: 1, color: 'rgba(10,204,204,0)', // 100% 处的颜色
          }],
          global: false, // 缺省为 false
        },
      };
    };
    if (options.series[0].color[0] === '#0acc78') {
      serie.areaStyle = {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [{
            offset: 0, color: 'rgba(10,204,120,0.3)', // 0% 处的颜色
          }, {
            offset: 0.5, color: 'rgba(10,204,120,.2)', // 100% 处的颜色
          }, {
            offset: 1, color: 'rgba(10,204,120,0)', // 100% 处的颜色
          }],
          global: false, // 缺省为 false
        },
      };
    };
    if (options.series[0].color[0] === '#0080FF') {
      serie.areaStyle = {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [{
            offset: 0, color: 'rgba(0,128,255,0.3)', // 0% 处的颜色
          }, {
            offset: 0.5, color: 'rgba(0,128,255,0.2)', // 100% 处的颜色
          }, {
            offset: 1, color: 'rgba(0,128,255,0)', // 100% 处的颜色
          }],
          global: false, // 缺省为 false
        },
      };
    }
    Object.assign(serie, options.series[0]);
    // eslint-disable-next-line no-param-reassign
    delete options.series;
  }
  const opts = {
    xAxis: {
      data: objData.map(item => item.time),
      splitLine: {
        show: false,
      },
      axisLabel: {
        show: true,
        textStyle: {
          color: '#0acccc',
        },
        formatter: options.xAxisFormatter,
      },
      axisLine: {
        lineStyle: {
          color: '#0acccc',
          width: 1,
        },
      },

    },
    yAxis: {
      type: 'value',
      splitLine: {
        show: false,
      },
      axisLabel: {
        show: true,
        textStyle: {
          color: '#0acccc',
        },
        formatter: options.yAxisFormatter,
      },
      axisLine: {
        lineStyle: {
          color: '#0acccc',
          width: 1, // 这里是为了突出显示加上的
        },
      },
      min: options.min,
      max: options.max,
    },
    grid: options.grid || {
      left: '15%',
      top: '20%',
    },
    series: [serie],
  };
  Object.assign(opts, options);
  return opts;
}

const getpanelOpts = (data, options) => {
  const option = {
    series: [
      {
        type: 'gauge',
        radius: '85%',
        startAngle: '220',
        endAngle: '-40',
        min: options.min,
        max: options.max,
        silent: true,
        // 图表的刻度分隔段数
        splitNumber: options.splitNumber || 10,
        // 图表的轴线相关
        axisLine: {
          show: true,
          lineStyle: {
            color: options.axisLinecolorObj || [
              [
                0.6,
                new echarts.graphic.LinearGradient(0, 1, 0, 0, [{
                  offset: 0,
                  color: 'rgba(122,255,236,0)',
                },
                {
                  offset: 1,
                  color: 'rgba(0,164,204,1.0)',
                },
                ]),
              ],
              options.axislineColor || [1, '#c50000'],
            ],
            width: 0.4211 * window.innerWidth / 100,
          },
        },
        // 图表的刻度及样式
        axisTick: {
          show: false,
        },
        // 图表的刻度标签(20、40、60等等)
        axisLabel: {
          distance: 0.2632 * window.innerWidth / 100,
          textStyle: {
            color: '#9E9E9E',
          },
          // 仪表盘刻度的字体大小
          fontSize: 0.6316 * window.innerWidth / 100,
          // 使用函数模板，函数参数分别为刻度数值
          formatter: options.formatter,
        },
        // 图表的分割线
        splitLine: {
          show: true,
          length: 0.6316 * window.innerWidth / 100,
          lineStyle: {
            color: '#fff',
            width: 0.8,
          },
        },
        // 图表的指针
        pointer: {
          show: true,
          length: '75%',
        },
        // 指针样式
        itemStyle: {
          color: '#fff',
          // opacity: 0.8,
          // shadowColor: '#0acccc',
          // shadowBlur: 8,
        },
        // 图表的数据详情
        detail: {
          formatter(params) {
            return `{score|${params}}` + `{unit|${options.unit}}` + '\n' + `{title|${data.name}}`;
          },
          offsetCenter: [0, '75%'],
          rich: {
            title: {
              fontSize: 0.8421 * window.innerWidth / 100,
              color: '#0ACCCC',
              lineHeight: 28,
              fontFamily: 'TencentSansW3',
            },
            unit: {
              fontSize: 0.6316 * window.innerWidth / 100,
              color: '#fff',
            },
            score: {
              fontSize: 1.5789 * window.innerWidth / 100,
              color: '#fff',
              fontFamily: 'DINAlternate-Bold',
            },
          },
        },
        data: [{
          name: '',
          value: data.value,
        }],
      },
    ],
  };
  return option;
};

export {
  getpanelOpts,
};
