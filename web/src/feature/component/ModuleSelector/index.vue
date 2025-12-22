<template>
  <el-select
    :value="curModule"
    v-bind="$attrs"
    @change="change"
  >
    <el-option
      v-for="item in modules"
      :key="item.value"
      :label="item.label"
      :value="item.value"
    />
  </el-select>
</template>

<script>
import { find } from 'lodash';
// import qs from 'qs';
// import api from '../../config/cgi';

export default {
  props: {
    init: {
      type: Boolean,
      default: true,
    },
    initBySearch: {
      type: Boolean,
      default() {
        return true;
      },
    },
  },

  data() {
    return {
      modules: [],
      curModule: void 0,
      // curModule: this.initBySearch ? qs.parse(location.search.replace(/^\?/, '').moduleId || (void 0)) : (void 0)
    };
  },
  computed: {
    curModuleName() {
      let curModuleName = '';
      this.modules.map((item) => {
        if (item.value === this.curModule) {
          curModuleName = item.label;
        }
        return true;
      });
      return curModuleName;
    },
  },

  mounted() {
    this.getModules();
  },

  methods: {
    getModules() {
      // this.$axios.post(api.getModules).then((data) => {
      const data = [{
        moduleId: TNBL.getCurrModule().id,
        moduleName: TNBL.getCurrModule().name,
      }];
      this.modules = data.map(item => ({
        value: `${item.moduleId}`,
        label: item.moduleName,
      }));
      const curModule = localStorage.getItem('curModule');
      if (curModule && find(this.modules, {
        value: curModule,
      })) {
        this.change(curModule);
      } else if (this.init && this.modules.length) {
        this.change(this.modules[0].value);
      }
      // });
    },
    change(v) {
      this.curModule = v;
      localStorage.setItem('curModule', v);

      this.$emit('change', v, this.curModuleName);
    },
  },
};
</script>

<style lang="scss" scoped>
.el-select {
  /deep/ .el-input {

    input {
      padding-left: 8px;
    }

    &:before, &:after {
      display: none;
    }
  }
}
</style>
