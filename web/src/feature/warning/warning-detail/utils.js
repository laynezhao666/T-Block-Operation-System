
export function getOptions(data, extData) {
  const colors = ['#3F9BF9', '#37d1d1', '#ffc751'];
  const series = [];
  const legends = [];
  for (let i = 0; i < data.length; i++) {
    console.log(data[i].serie.length);
    series.push({
      name: data[i].name,
      type: 'line',
      showAllSymbol: false,
      color: [colors[i]],
      data: data[i].serie.map(item => item.value),
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [{
            offset: 0, color: 'rgba(63,155,249,0.15)', // 0% 处的颜色
          }, {
            offset: 0.5, color: 'rgba(63,155,249,0.01)', // 100% 处的颜色
          }, {
            offset: 1, color: 'rgba(63,155,249,0)', // 100% 处的颜色
          }],
          global: false, // 缺省为 false
        },
      },
    });
    legends.push(data[i].name);
  }

  const options = {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(51, 51, 51, .8)',
      padding: 16,
      extraCssText: 'border-radius: 0; min-width: 200px;',
      textStyle: {
        fontSize: 12,
      },
      axisPointer: {
        type: 'cross',
        label: {
          backgroundColor: '#6a7985',
        },
      },
    },
    legend: {
      top: '-3px',
      data: legends,
      selected: { 功率法: true, 电量法: false },
    },
    xAxis: {
      data: data[0].serie.map(item => item.time),
      axisLabel: {
        textStyle: {
          color: '#999',
          lineHeight: 18,
        },
        formatter(value) {
          const str = value;
          const [date, time] = str.split(' ');
          return [`${date}`, `${time}`].join('\n');
        },
      },
    },
    yAxis: {
      textStyle: {
        color: '#999999',
      },
      splitLine: {
        show: false,
      },
    },
    grid: {
      left: '16px',
      right: '16px',
      top: '16px',
      containLabel: true,
      bottom: '24px',
    },
    brush: {
      xAxisIndex: 'all',
      brushLink: 'all',
      transformable: true,
      throttleType: 'debounce',
      throttleDelay: 300,
      outOfBrush: {
        colorAlpha: 0.1,
      },
    },
    toolbox: {
      show: false,
    },
    series,
  };
  if (extData.title) {
    options.title = {
      left: 'center',
      text: extData.title,
    };
  }
  if (extData.yMin) {
    const { yMin } = extData;
    // options.yAxis.min = Math.trunc(yMin)
    options.yAxis.min = Math.trunc(yMin * 100) / 100;
  }
  console.log(extData);
  if (extData.yMax) {
    const { yMax } = extData;
    // options.yAxis.max = Math.ceil(yMax)
    options.yAxis.max = Math.ceil(yMax * 100) / 100;
  }
  return options;
}

export function floor(i) {
  return i;
}
