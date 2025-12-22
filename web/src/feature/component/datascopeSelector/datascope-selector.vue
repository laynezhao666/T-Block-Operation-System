<template>
  <el-cascader
    v-model="mozuName"
    class="mozu-cascader"
    :options="dataScopes"
    size="medium"
    filterable
    clearable=""
    :props="{ expandTrigger: 'hover',value: 'id' ,label:'name'}"
    :show-all-levels="true"
    @change="handleChange"
  />
</template>

<script>
import Cookies from 'js-cookie';

export default {
  props: {
    dataScope: {
      type: Array,
      default() {
        return [];
      },
    },
    withAllScope: {
      type: Boolean,
      default: false,
    },
    useModel: {
      type: String,
      default: 'localstorage',
    },
    setLocal: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      mozuName: [],
      dataScopes: this.dataScope,
      options: [
        {
          id: '全部',
          name: '全部',
          type: 'region',
          children: [{
            id: '全部',
            name: '全部',
            type: 'campus',
            children: [
              {
                id: '全部',
                name: '全部模组',
                type: 'mozu',
              },
            ],
          }],
        },
      ],
    };
  },
  created() {
  },
  mounted() {
    TNBL.getScopeModules().then((r) => {
      console.log(r);
      this.dataScopes = r.module_groups;
      this.initDatascope();
    });
  },
  methods: {
    getCurrentMozu() {
      if (localStorage.getItem('datascope') && JSON.parse(localStorage.getItem('datascope')).length !== 0) {
        return JSON.parse(localStorage.getItem('datascope'))[2];
      }
    },
    getCurrentCampus() {
      return this.getNode('id', JSON.parse(localStorage.getItem('datascope'))[1]);
    },
    initDatascope() {
      const tempScope = this.dataScopes;
      if (localStorage.getItem('datascope') && JSON.parse(localStorage.getItem('datascope')).length !== 0) {
        if (!JSON.parse(localStorage.getItem('datascope'))[0]) localStorage.removeItem('datascope');
      }
      if (this.withAllScope) {
        this.options[0].children[0].children[0].id = this.getAllMozu().join(',');
        this.dataScopes = this.options.concat(this.dataScopes);
      }

      if (localStorage.getItem('datascope') && JSON.parse(localStorage.getItem('datascope')).length !== 0 && JSON.parse(localStorage.getItem('datascope'))[0]) {
        if (this.getCurrentMozu() && !isNaN(Number(this.getCurrentMozu()))) {
          this.$emit('mozuloaded', this.getNode('id', this.getCurrentMozu()));
          this.mozuName = JSON.parse(localStorage.getItem('datascope'));
        } else {
          if (this.withAllScope) {
            this.$emit('mozuloaded', {
              id: this.getAllMozu().join(','),
              name: '全部模组',
              type: 'mozu',
            });
          } else {
            this.$emit('mozuloaded', this.getNode('id', tempScope[0]?.children[0]?.children[0].id));
            this.mozuName = [tempScope[0]?.id, tempScope[0]?.children[0]?.id,
        tempScope[0]?.children[0]?.children[0].id];
            localStorage.setItem('datascope', JSON.stringify(this.mozuName));
          }
        }
      } else {
        if (this.withAllScope) {
          this.mozuName = ['全部', '全部', this.getAllMozu().join(',')];
        } else {
          this.mozuName = [this.dataScopes[0]?.id,
          this.dataScopes[0]?.children[0]?.id,
          this.dataScopes[0]?.children[0]?.children[0].id];
        }
        localStorage.setItem('datascope', JSON.stringify(this.mozuName));
        this.$emit('mozuloaded', this.getNode('id', this.getCurrentMozu()));
      }
    },
    getNode(key, value) {
      let stark = [];
      stark = stark.concat(this.dataScopes);
      while (stark.length) {
        const temp = stark.shift();
        if (temp.children) {
          stark = stark.concat(temp.children);
        }
        if (temp[key] === value) {
          return temp;
        }
      }
    },
    setCookie(key, val, expires = 1) {
      console.log('init:cookie', key, val);

      if (val === undefined) {
        return Cookies.get(key);
      }

      const domainArr = location.hostname.split('.');
      const domain = domainArr.filter((item, i) => i > (domainArr.length - 4)).join('.');

      return Cookies.set(key, val, { expires, domain, path: '/' });
    },
    getAllMozu() {
      const result = [];
      let stark = [];
      stark = stark.concat(this.dataScopes);
      while (stark.length) {
        const temp = stark.shift();
        if (temp.children) {
          stark = stark.concat(temp.children);
        }
        if (temp.type === 'mozu' && temp.name !== '全部模组') result.push(temp.name);
      }
      return result;
    },
    handleChange(val) {
      const mozuObj = this.getNode('id', val[2]);
      if (!this.setLocal) {
        this.$emit('change', this.getNode('id', val[2]));
      } else {
        if (this.useModel === 'cookie') {
          this.setCookie('tnebula_cu_moduleid', mozuObj.id);
          this.setCookie('tnebula_cu_modulename', mozuObj.name);
          this.setCookie('tnebula_cu_modulealias', mozuObj.name);
          localStorage.setItem('datascope', JSON.stringify(val));
          location.reload();
        } else {
          this.setCookie('tnebula_cu_moduleid', mozuObj.id);
          this.setCookie('tnebula_cu_modulename', mozuObj.name);
          this.setCookie('tnebula_cu_modulealias', mozuObj.name);
          localStorage.setItem('datascope', JSON.stringify(val));
          location.reload();
        // this.$emit('change', this.getNode('id', val[2]));
        }
      }
    },
  },
};
</script>

<style lang="scss">
.mozu-cascader {
    margin-top:16px;
    float:right;
    width:389px;
  .el-input__inner {
    height: 35px;
    padding-left: 5px ;
  }
  .el-input::before{
    content:none;
  }
  .el-input::after{
    content:none;
  }
  .el-input__suffix-inner {
      margin-right:10px
  }
}
</style>
