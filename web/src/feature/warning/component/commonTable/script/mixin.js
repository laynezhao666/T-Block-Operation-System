import qs from 'qs';

export default {
  methods: {
    hasRights(opr) {
      const rights = this.rights || 0b00000;
      const oprName = { canExport: 0b10000, canImport: 0b01000, canDel: 0b00100, canEdit: 0b00010, canAdd: 0b00001 };
      return rights & oprName[opr];
    },
    parseEnum(enums) {
      return enums.map((enu) => {
        const arr = enu.split('|');
        if (arr.length === 1) {
          return {
            value: arr[0],
            label: arr[0],
          };
        }
        return {
          value: arr[0],
          label: arr[1],
        };
      });
    },
    reverseData(data) {
      const result = {};
      Object.keys(data).forEach((item) => {
        result[data[item]] = item;
      });
      return result;
    },
    length(text) {
      const reg = /[\x21-\x7E]/g;
      const match = text.match(reg);
      if (match) {
        return text.length - (match.length / 2);
      }
      return text.length;
    },
    calcWidth(label, { count = 0, type = 'text' }) {
      const padding = 24 + 24;
      const border = 1;
      let body;
      if (type === 'date') {
        body = 72;
      } else if (type === 'num') {
        body = count * 12 / 2;
      } else if (type === 'char') {
        body = count * 12 / 1.5;
      } else {
        body = count * 12;
      }
      const header = this.length(label) * 14;
      const max = Math.max(header, body);
      return max + padding + border;
    },
    handleConfig(config) {
      return config.map(item => ({
        ...item,
        fixed: item.fixed,
        show: item.show,
        width: this.calcWidth(item.label, { count: item.size }),
      }));
    },
    jump(path, query = {}, row = {}) {
      const ind = path.indexOf('?');
      const url = ind > -1 ? path.slice(0, ind) : path;
      const params = {};
      Object.keys(query).forEach((k) => {
        params[k] = row[query[k]];
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
      window.open(`${url}${queryStr ? `?${queryStr}` : ''}`);
    },
    playSuccess() {
      setTimeout(() => { this.$message.success('操作成功'); });
    },
  },
};
