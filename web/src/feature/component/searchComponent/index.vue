<template>
  <el-block
    no-padding
    inner
  >
    <div v-if="showTab">
      <el-tabs
        v-model="flowKeyIndex"
        @tab-click="handleClick"
      >
        <el-tab-pane
          v-for="tab in orderTypeList"
          :key="tab.index"
          :auth-roles="tab.authRoles"
          :label="tab.title"
          :name="tab.index.toString()"
        />
      </el-tabs>
    </div>
    <div class="el-common-table">
      <div class="config-panel-wrap">
        <div
          v-show="(columns || []).length > 0"
          class="config-panel-body"
        >
          <advanced-search
            :columns="columns"
            :flow-key="flowKey"
            :conditions="conditions"
            :show-export-btn="showExportBtn"
            :search-collapsed="searchCollapsed"
            :init-from="initFrom"
            @search="search"
            @export="download"
            @refreshTable="refreshTable"
          >
            <template
              v-for="(slot, name) in $scopedSlots"
              :slot="name"
              slot-scope="scope"
            >
              <slot
                :name="name"
                :row="scope.row"
              />
            </template>
          </advanced-search>
          <data-table
            ref="table"
            :columns="columns"
            :query="query"
            :hide-check-box="hideCheckBox"
            :show-export-btn="showExportBtn"
            :export-key="exportKey"
            :extra-params="extraParams"
            :actions-label-width="actionsLabelWidth"
            :custom-fixed="customFixed"
            :loading-switch="loadingSwitch"
            @export="download"
            @change="change"
          >
            <template
              v-for="(slot, name) in $scopedSlots"
              :slot="name"
              slot-scope="scope"
            >
              <slot
                :name="name"
                :row="scope.row"
              />
            </template>
          </data-table>
        </div>
      </div>
    </div>
  </el-block>
</template>
<script>
import { cloneDeep, throttle } from 'lodash';
import AdvancedSearch from './AdvancedSearch';
import DataTable from './Table';
import ResizeObserver from 'resize-observer-polyfill';

export default {
  components: {
    DataTable,
    AdvancedSearch,
  },
  provide() {
    return {
      configCgi: this.configCgi,
      commonCgi: this.commonCgi,
      tableConfig: this.tableConfig,
      codes: this.codes,
      roles: this.roles,
    };
  },
  props: {
    conditions: {
      type: Array,
      default: () => ([]),
    },
    configCgi: {
      type: Object,
      default: () => ({}),
    },
    commonCgi: {
      type: Object,
      default: () => ({}),
    },
    tableConfig: {
      type: Object,
      default: () => ({}),
    },
    hideCheckBox: {
      type: Boolean,
      default: false,
    },
    exportKey: {
      type: String,
      default: 'id',
    },
    codes: {
      type: Object,
      default: () => ({}),
    },
    roles: {
      type: Object,
      default: () => ({}),
    },
    customizeExportFunc: {
      type: Function,
      default: null,
    },
    showExportBtn: {
      type: Boolean,
      default: true,
    },
    actionsLabelWidth: {
      type: Number,
      default: 120,
    },
    extraParams: {
      type: Object,
      default: () => ({}),
    },
    customFixed: {
      type: Boolean,
      default: false,
    },
    searchCollapsed: {
      type: Boolean,
      default: false,
    },
    loadingSwitch: {
      type: Boolean,
      default: false,
    },
    initFrom: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      show: false,
      orderTypeList: [],
      flowKeyIndex: 0,
      columns: [],
      saveColumns: {},
      query: {},
      advancedConditions: {},
      observer: null,
      configRes: null,
    };
  },
  computed: {
    flowKey() {
      if (!this.orderTypeList.length) return '';
      return this.orderTypeList[Number(this.flowKeyIndex)].flowKey;
    },
    showTab() {
      return this.orderTypeList.length > 1;
    },
  },
  watch: {
    conditions: {
      handler(val, oldVal) {
        if (Object.keys(val).length !== 0) {
          if (oldVal && oldVal[0].value.length !== 0) this.search();
        }
      },
      deep: true,
      immediate: true,
    },
  },
  mounted() {
    this.init().then((data) => {
      this.orderTypeList = this.formatFlowKeyObj(data);
      this.getConfig();
    });
    this.$nextTick(() => {
      this.listernSideBarRezise();
    });
  },
  beforeDestroy() {
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
  },
  methods: {
    init() {
      return new Promise((resolve) => {
        if (this.tableConfig?.flowKeyObj) { // 外部获取流程名称
          resolve(this.tableConfig.flowKeyObj);
        } else {
          this.$axios.get(this.configCgi.instanceDict).then((flowKeyObj) => {
            resolve(flowKeyObj);
          });
        }
      });
    },
    // 监听sidebar重新计算宽度
    listernSideBarRezise() {
      const element = document.querySelector('.el-sidebar');
      if (element) {
        this.observer = new ResizeObserver(() => {
          this.resetColumn();
        });
        this.observer.observe(element);
      }
    },
    resetColumn: throttle(function () {
      if (this.configRes) {
        this.columns = this.configRes.length ? this.handleConfig(this.configRes) : [];
      }
    }, 300, {
      leading: false,
      trailing: true,
    }),
    getConfig() {
      if (this.saveColumns[this.flowKey]) { // 如果列表配置数据被申请过了
        this.columns = this.saveColumns[this.flowKey];
        return this.search();
      }
      return this.$axios.get(this.configCgi.getConfigCgi, {
        flowKey: this.orderTypeList[this.flowKeyIndex].flowKey,
      }).then((res) => {
        this.$emit('getColumns', res); // 外部可能会用到这些配置数据
        const main = res.main || res;
        this.columns = main.length ? this.handleConfig(main) : [];
        this.configRes = main;
        this.search();
      });
    },
    change(list) {
      this.$emit('change', list);
    },
    handleClick() {
      this.getConfig();
      this.$emit('getFlowKeyIndex', this.flowKeyIndex);
    },
    length(text) {
      const reg = /[\x21-\x7E]/g;
      const match = text.match(reg);
      if (match) {
        return text.length - (match.length / 2); // 英文及符号长度减半
      }
      return text.length;
    },
    calcWidth(label, { count = 8, type = 'text' }) {
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
      // eslint-disable-next-line prefer-destructuring
      let { clientWidth } = $('.table-wrap')[0];
      if (!clientWidth) {
        clientWidth = $('.config-panel-wrap')[0].clientWidth;
      }
      const origin = config.map(v => ({
        ...v,
        show: v.show || false,
        width: this.calcWidth(v.label, { count: v.size > 0 ? v.size : 8 }),
      }));
      const target = origin.filter(v => v.show);
      const totalWidth = target.map(v => v.width).reduce((prev, cur) => prev + cur);
      if (clientWidth > totalWidth) {
        const addWidth = (clientWidth - totalWidth) / target.length;
        return origin.map(item => ({
          ...item,
          width: item.width + addWidth,
        }));
      }
      this.saveColumns[this.flowKey] = origin;
      return origin;
    },
    search(params) {
      // 如果没有新条件，就用旧条件
      const data = params || this.advancedConditions;
      let conditions = cloneDeep(data);
      this.$set(this, 'advancedConditions', data);
      if (this.conditions.length) { // 先处理conditions
        this.conditions.forEach((item) => {
          const { relationGroup } = item;
          if (!conditions[relationGroup]) {
            conditions[relationGroup] = [item];
          } else {
            conditions[relationGroup].push(item);
          }
        });
      }
      const currentTab = this.orderTypeList[this.flowKeyIndex];
      if (currentTab.conditions && JSON.stringify(currentTab.conditions) !== '{}') conditions = Object.assign(conditions, { ...currentTab.conditions });
      this.query = {
        conditions,
        flowKey: currentTab.flowKey,
        fields: this.columns.map(v2 => v2.name),
      };
      // 刷新tab数量
      this.$emit('refreshTabNumber');
    },
    download(ids) {
      if (ids) { // 导出所选
        const value = { field: this.exportKey, relation: 'AND', operator: 'IN', value: ids };
        if (!this.query.conditions.default) {
          this.query.conditions.default = [value];
        } else {
          const fields = this.query.conditions.default.map(v => v.field);
          const index = fields.indexOf(this.exportKey);
          if (index > -1) {
            this.query.conditions.default[index].value = ids;
          } else {
            this.query.conditions.default.push(value);
          }
        }
      } else { // 按筛选条件导出全部
        const defaultVal = this.query.conditions.default;
        if (defaultVal) {
          this.query.conditions.default = defaultVal.filter(v => v.field !== 'id');
        }
      }
      if (this.customizeExportFunc) {
        return this.customizeExportFunc({ query: this.query });
      }
      return this.$axios.download(this.configCgi.exportCgi, {
        ...this.query,
      });
    },
    refreshTable() {
      this.$refs.table.refresh();
    },
    formatFlowKeyObj(data) {
      if (!data || JSON.stringify(data) === '{}') return [];
      const ret = [];
      Object.entries(data).forEach(([k, v]) => {
        if (Array.isArray(v)) {
          v.forEach((subV) => {
            ret.push({
              index: ret.length,
              flowKey: k,
              authRoles: this.roles[k],
              title: subV.title,
              conditions: subV.conditions || [],
            });
          });
        } else {
          ret.push({ index: ret.length,
            authRoles: this.roles[k],
            flowKey: k,
            title: v });
        }
      });
      return ret;
    },
  },
};
</script>
<style lang="scss" scoped>
  .el-common-table {
    background: #fff;
  }
  .config-panel-wrap {
    display: flex;

    .config-panel-body {
      overflow: auto;
      flex: 1;
    }
  }
</style>
