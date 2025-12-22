import BigNumber from 'bignumber.js';
import qs from 'qs';
import { isNumber, isNil, get, isString } from 'lodash';

export const NULL_LABEL = '--';

export default {
  directives: {
    confirm: {
      bind(el, binding, vnode) {
        el.addEventListener('click', () => {
          vnode.context.$confirm('确认要删除吗？', '系统提示').then(binding.value);
        });
      },
    },
  },
  data() {
    return {
      dateOpts: {
        disabledDate: time => time.getTime() > new Date(),
      },
    };
  },
  filters: {
    percentify(v) {
      const rst = percentify(v);
      return isNumber(rst) ? `${rst}%` : rst;
    },
    polyfill,
    numberify,
    track: data => data,
  },
  methods: {
    jump(path, query = [], row = {}) {
      const ind = path.indexOf('?');
      const url = ind > -1 ? path.slice(0, ind) : path;
      const params = {};
      query.forEach((q) => {
        params[q] = row[q];
      });

      if (ind > -1) {
        const data = path.slice(ind + 1).split('&');
        data.forEach((item) => {
          const temp = item.split('=');
          // eslint-disable-next-line prefer-destructuring
          params[temp[0]] = temp[1];
        });
      }
      const queryStr = qs.stringify(params);
      location.href = `${url}${queryStr ? `?${queryStr}` : ''}`;
    },
    getText(channel, type, level) {
      const cMap = {
        cool: '冷通道',
        hot: '热通道',
      };
      const tMap = {
        temp: '温度',
        hum: '湿度',
      };

      const lMap = {
        max: '最高',
        avg: '平均',
        min: '最低',
      };

      return `${cMap[channel]}${lMap[level]}${tMap[type]}`;
    },
    getChartData(chart) {
      const colors = ['#FF6D00', '#09DCE7', '#1470CC'];

      return {
        title: chart.title,
        type: chart.datasets[0]?.unit === '℃' ? 'temp' : 'percent',
        series: chart.datasets.map((dataset, i) => ({
          name: dataset.name,
          color: [colors[i]],
          data: dataset.values,
        })),
      };
    },
    getProp: get,
  },
};

export function percentify(v) {
  if (isNumber(v)) {
    if (v > 1) {
      // eslint-disable-next-line new-cap
      return +BigNumber(v).valueOf();
    }
    // eslint-disable-next-line new-cap
    return +BigNumber(v).times(100)
      .valueOf();
  }
  return v;
}

export function numberify(v) {
  if (isNaN(v)) {
    return v;
  }
  if (v.startsWith?.('0') & v.length > 1) {
    return v;
  }
  return +v;
}

export function polyfill(v) {
  if (isString(v)) {
    return v.length ? v : NULL_LABEL;
  }
  return isNil(v) ? NULL_LABEL : v;
}

export function composePercentify(v) {
  // eslint-disable-next-line no-confusing-arrow
  return v |> percentify |> polyfill |> (v => v === NULL_LABEL ? v : `${v}%`);
}

export function binToArr(num) {
  let value = 1;
  let bit = 0;
  const rst = [];
  while (num >= value) {
    if (num & value) {
      rst.push(bit);
    }
    bit = bit + 1;
    value = value << 1;
  }
  return rst;
}

export function arrToBin(arr) {
  let rst = 0;
  arr.forEach((item) => {
    const value = 1 << item;
    rst += value;
  });
  return rst;
}
