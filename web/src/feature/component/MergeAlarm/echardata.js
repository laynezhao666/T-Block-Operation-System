import * as echarts from 'echarts';
export default {};
export const initChart = function (el) {
  const myChart = echarts.init(el);
  let option = {};

  myChart.showLoading();

  const data = {
    name: '告警维度（20：14）',
    children: [
      {
        name: 'data',
        children: [
          {
            name: 'converters',
            children: [
              { name: 'Converters', value: 721 },
              { name: 'DelimitedTextConverter', value: 4294 },
            ],
          },
          {
            name: 'DataUtil',
            value: 3322,
          },
        ],
      },
      {
        name: 'display',
        children: [
          { name: 'DirtySprite', value: 8833 },
          { name: 'LineSprite', value: 1732 },
          { name: 'RectSprite', value: 3623 },
        ],
      },
      {
        name: 'flex',
        children: [
          { name: 'FlareVis', value: 4116 },
        ],
      },
      {
        name: 'query',
        children: [
          { name: 'AggregateExpression', value: 1616 },
          { name: 'And', value: 1027 },
          { name: 'Arithmetic', value: 3891 },
          { name: 'Average', value: 891 },
          { name: 'BinaryExpression', value: 2893 },
          { name: 'Comparison', value: 5103 },
          { name: 'CompositeExpression', value: 3677 },
          { name: 'Count', value: 781 },
          { name: 'DateUtil', value: 4141 },
          { name: 'Distinct', value: 933 },
          { name: 'Expression', value: 5130 },
          { name: 'ExpressionIterator', value: 3617 },
          { name: 'Fn', value: 3240 },
          { name: 'If', value: 2732 },
          { name: 'IsA', value: 2039 },
          { name: 'Literal', value: 1214 },
          { name: 'Match', value: 3748 },
          { name: 'Maximum', value: 843 },
          {
            name: 'methods',
            children: [
              { name: 'add', value: 593 },
              { name: 'and', value: 330 },
              { name: 'average', value: 287 },
              { name: 'count', value: 277 },
              { name: 'distinct', value: 292 },
              { name: 'div', value: 595 },
              { name: 'eq', value: 594 },
              { name: 'fn', value: 460 },
              { name: 'mul', value: 603 },
              { name: 'neq', value: 599 },
              { name: 'not', value: 386 },
              { name: 'or', value: 323 },
              { name: 'orderby', value: 307 },
              { name: 'range', value: 772 },
              { name: 'select', value: 296 },
              { name: 'stddev', value: 363 },
              { name: 'sub', value: 600 },
              { name: 'sum', value: 280 },
              { name: 'update', value: 307 },
              { name: 'variance', value: 335 },
              { name: 'where', value: 299 },
              { name: 'xor', value: 354 },
              { name: '_', value: 264 },
            ],
          },
          { name: 'Minimum', value: 843 },
          { name: 'Not', value: 1554 },
          { name: 'Or', value: 970 },
          { name: 'Query', value: 13896 },
          { name: 'Range', value: 1594 },
          { name: 'StringUtil', value: 4130 },
          { name: 'Sum', value: 791 },
          { name: 'Variable', value: 1124 },
          { name: 'Variance', value: 1876 },
          { name: 'Xor', value: 1101 },
        ],
      },
      {
        name: 'scale',
        children: [
          { name: 'IScaleMap', value: 2105 },
          { name: 'LinearScale', value: 1316 },
          { name: 'LogScale', value: 3151 },
          { name: 'OrdinalScale', value: 3770 },
          { name: 'QuantileScale', value: 2435 },
          { name: 'QuantitativeScale', value: 4839 },
          { name: 'RootScale', value: 1756 },
          { name: 'Scale', value: 4268 },
          { name: 'ScaleType', value: 1821 },
          { name: 'TimeScale', value: 5833 },
        ],
      },
    ],
  };

  const data2 = {
    name: '设备维度（20：14）',
    children: [
      {
        name: 'flex',
        children: [
          { name: 'FlareVis', value: 4116 },
        ],
      },
      {
        name: 'scale',
        children: [
          { name: 'IScaleMap', value: 2105 },
          { name: 'LinearScale', value: 1316 },
          { name: 'LogScale', value: 3151 },
          { name: 'OrdinalScale', value: 3770 },
          { name: 'QuantileScale', value: 2435 },
          { name: 'QuantitativeScale', value: 4839 },
          { name: 'RootScale', value: 1756 },
          { name: 'Scale', value: 4268 },
          { name: 'ScaleType', value: 1821 },
          { name: 'TimeScale', value: 5833 },
        ],
      },
      {
        name: 'display',
        children: [
          { name: 'DirtySprite', value: 8833 },
        ],
      },
    ],
  };

  myChart.hideLoading();

  myChart.setOption(option = {
    tooltip: {
      trigger: 'item',
      triggerOn: 'mousemove',
    },
    legend: {
      top: '2%',
      left: '3%',
      orient: 'vertical',
      data: [{
        name: '告警维度',
        icon: 'rectangle',
      },
      {
        name: '设备维度',
        icon: 'rectangle',
      }],
      borderColor: '#c23531',
    },
    series: [
      {
        type: 'tree',

        name: '告警维度',

        data: [data],

        top: '5%',
        left: '7%',
        bottom: '2%',
        right: '60%',

        symbolSize: 7,

        label: {
          position: 'left',
          verticalAlign: 'middle',
          align: 'right',
        },

        leaves: {
          label: {
            position: 'right',
            verticalAlign: 'middle',
            align: 'left',
          },
        },

        emphasis: {
          focus: 'descendant',
        },

        expandAndCollapse: true,

        animationDuration: 550,
        animationDurationUpdate: 750,

      },
      {
        type: 'tree',
        name: '设备维度',
        data: [data2],

        top: '20%',
        left: '60%',
        bottom: '22%',
        right: '18%',

        symbolSize: 7,

        label: {
          position: 'left',
          verticalAlign: 'middle',
          align: 'right',
        },

        leaves: {
          label: {
            position: 'right',
            verticalAlign: 'middle',
            align: 'left',
          },
        },

        expandAndCollapse: true,

        emphasis: {
          focus: 'descendant',
        },

        animationDuration: 550,
        animationDurationUpdate: 750,
      },
    ],
  });

  option && myChart.setOption(option);
};

export const initLine = function (el) {
  const myChart = echarts.init(el);
  let option = {};
  const dimensions = [
    'name', 'Price', 'Prime cost', 'Prime cost min', 'Prime cost max', 'Price min', 'Price max',
  ];
  const data = [
    ['Blouse "Blue Viola"', 101.88, 99.75, 76.75, 116.75, 69.88, 119.88],
    ['Dress "Daisy"', 155.8, 144.03, 126.03, 156.03, 129.8, 188.8],
    ['Trousers "Cutesy Classic"', 203.25, 173.56, 151.56, 187.56, 183.25, 249.25],
    ['Dress "Morning Dew"', 256, 120.5, 98.5, 136.5, 236, 279],
    ['Turtleneck "Dark Chocolate"', 408.89, 294.75, 276.75, 316.75, 385.89, 427.89],
    ['Jumper "Early Spring"', 427.36, 430.24, 407.24, 452.24, 399.36, 461.36],
    ['Breeches "Summer Mood"', 356, 135.5, 123.5, 151.5, 333, 387],
    ['Dress "Mauve Chamomile"', 406, 95.5, 73.5, 111.5, 366, 429],
    ['Dress "Flying Tits"', 527.36, 503.24, 488.24, 525.24, 485.36, 551.36],
    ['Dress "Singing Nightingales"', 587.36, 543.24, 518.24, 555.24, 559.36, 624.36],
    ['Sundress "Cloudy weather"', 603.36, 407.24, 392.24, 419.24, 581.36, 627.36],
    ['Sundress "East motives"', 633.36, 477.24, 445.24, 487.24, 594.36, 652.36],
    ['Sweater "Cold morning"', 517.36, 437.24, 416.24, 454.24, 488.36, 565.36],
    ['Trousers "Lavender Fields"', 443.36, 387.24, 370.24, 413.24, 412.36, 484.36],
    ['Jumper "Coffee with Milk"', 543.36, 307.24, 288.24, 317.24, 509.36, 574.36],
    ['Blouse "Blooming Cactus"', 790.36, 277.24, 254.24, 295.24, 764.36, 818.36],
    ['Sweater "Fluffy Comfort"', 790.34, 678.34, 660.34, 690.34, 762.34, 824.34],
  ];

  function renderItem(params, api) {
    const children = [];
    const coordDims = ['x', 'y'];

    for (let baseDimIdx = 0; baseDimIdx < 2; baseDimIdx++) {
      const otherDimIdx = 1 - baseDimIdx;
      const { encode } = params;
      const baseValue = api.value(encode[coordDims[baseDimIdx]][0]);
      const param = [];
      param[baseDimIdx] = baseValue;
      param[otherDimIdx] = api.value(encode[coordDims[otherDimIdx]][1]);
      const highPoint = api.coord(param);
      param[otherDimIdx] = api.value(encode[coordDims[otherDimIdx]][2]);
      const lowPoint = api.coord(param);
      const halfWidth = 5;

      const style = api.style({
        stroke: api.visual('color'),
        fill: null,
      });

      children.push({
        type: 'line',
        transition: ['shape'],
        shape: makeShape(
          baseDimIdx,
          highPoint[baseDimIdx] - halfWidth, highPoint[otherDimIdx],
          highPoint[baseDimIdx] + halfWidth, highPoint[otherDimIdx]
        ),
        style,
      }, {
        type: 'line',
        transition: ['shape'],
        shape: makeShape(
          baseDimIdx,
          highPoint[baseDimIdx], highPoint[otherDimIdx],
          lowPoint[baseDimIdx], lowPoint[otherDimIdx]
        ),
        style,
      }, {
        type: 'line',
        transition: ['shape'],
        shape: makeShape(
          baseDimIdx,
          lowPoint[baseDimIdx] - halfWidth, lowPoint[otherDimIdx],
          lowPoint[baseDimIdx] + halfWidth, lowPoint[otherDimIdx]
        ),
        style,
      });
    }

    function makeShape(baseDimIdx, base1, value1, base2, value2) {
      const shape = {};
      shape[`${coordDims[baseDimIdx]}1`] = base1;
      shape[`${coordDims[1 - baseDimIdx]}1`] = value1;
      shape[`${coordDims[baseDimIdx]}2`] = base2;
      shape[`${coordDims[1 - baseDimIdx]}2`] = value2;
      return shape;
    }

    return {
      type: 'group',
      children,
    };
  }

  option = {
    tooltip: {
    },
    legend: {
      data: ['bar', '收敛时间线'],
    },
    dataZoom: [{
      type: 'slider',
    }, {
      type: 'inside',
    }],
    grid: {
      bottom: 80,
    },
    xAxis: {},
    yAxis: {},
    series: [{
      type: 'scatter',
      name: '收敛时间线',
      data,
      dimensions,
      encode: {
        x: 2,
        y: 1,
        tooltip: [2, 1, 3, 4, 5, 6],
        itemName: 0,
      },
      itemStyle: {
        color: '#77bef7',
      },
    }, {
      type: 'custom',
      name: '收敛时间线',
      renderItem,
      dimensions,
      encode: {
        x: [2, 3, 4],
        y: [1, 5, 6],
        tooltip: [2, 1, 3, 4, 5, 6],
        itemName: 0,
      },
      data,
      z: 100,
    }],
  };

  option && myChart.setOption(option);
};
