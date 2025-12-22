/**
 * 把拉取用户数据单独拎出来
 * 这样可以提供给不同的组件使用
 * @provide cgi
 * @data map
 */
export default {
  data() {
    return {
      defaultParams: {
        start: 0,
        limit: 5,
      },
    };
  },
  computed: {
    cgiUrl() {
      return this.commonCgi.getSimpleUserList;
    },
  },
  methods: {
    getList(params) {
      return new Promise((resolve, reject) => {
        this.$axios.post(this.cgiUrl, {
          ...this.defaultParams,
          ...params,
        }, undefined, {
          isJson: true,
        }).then(({ list }) => {
          resolve(list || []);
        })
          .catch(() => {
            reject(new Error('fail'));
          });
      });
    },
  },
};
