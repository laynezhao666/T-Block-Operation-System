export default {
  data() {
    return {

    };
  },
  mounted() {
  },
  methods: {
    solveRights() {
      const currentRole = window.currentRole || -1;
      let authRightsMap = {
        0b10000: [0, 5], // 导出
        0b01000: [0], // 导入
        0b00100: [0], // 删除
        0b00010: [0], // 编辑
        0b00001: [0], // 新增
      };
      if (this.authRightsMap) {
        authRightsMap = this.authRightsMap;
      }
      let result = 0b00000;
      Object.keys(authRightsMap).forEach((item) => {
        if (authRightsMap[item].includes(currentRole)) {
          result = result | item;
        }
      });
      if (window.hasAppAuth) {
        return result;
      }
      return this.rights;
    },
  },
};
