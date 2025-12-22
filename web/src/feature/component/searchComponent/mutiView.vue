<template>
  <div class="change-calendar">
    <advanced-search
      :columns="columns"
      :flow-key="flowKey"
      :conditions="conditions"
      :show-export-btn="showExportBtn"
      :search-collapsed="true"
      @search="search"
      @export="download"
    />
    <el-block
      inner
    >
      <el-tabs
        v-model="viewType"
        :stretch="true"
        type="card"
      >
        <el-tab-pane
          v-for="(v, label) in views"
          :key="label"
          :name="label"
          :label="viewConfig[label].name"
        />
      </el-tabs>
    </el-block>
    <data-calendar
      v-if="views.calendar && viewType === 'calendar'"
      ref="calendar"
      :query="query"
      :config="views.calendar"
    >
      <template #legend>
        <slot name="calendarLegend" />
      </template>
    </data-calendar>
    <data-table
      v-if="views.table && viewType === 'table'"
      ref="table"
      :columns="columns"
      :query="query"
      :hide-check-box="hideCheckBox"
      :show-export-btn="showExportBtn"
      @export="download"
    />
  </div>
</template>

<script>
import AdvancedSearch from './AdvancedSearch.vue';
import DataTable from './Table';
import DataCalendar from './Calendar';

export default {
  components: {
    DataTable,
    AdvancedSearch,
    DataCalendar,
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
    views: {
      type: Object,
      required: true,
    },
    codes: {
      type: Object,
      default: () => ({}),
    },
    roles: {
      type: Object,
      default: () => ({}),
    },
    showExportBtn: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      viewType: '',
      // 这里是所有视图类型的列表，后续可继续添加
      // 通过传入的views来决定是否展示和顺序
      viewConfig: {
        calendar: { name: '日历视图' },
        table: { name: '列表视图' },
      },
      columns: [],
      saveColumns: {},
      query: {},
      showTab: false,
      flowKey: '',
      orderTypeList: {},
    };
  },
  watch: {
    viewType(val) {
      this.$nextTick(() => {
        // 每个tab需要都有refresh可以调用，重新拉取列表
        this.$refs[val] && this.$refs[val].refresh();
      });
    },
  },
  mounted() {
    this.init().then(() => {
      this.query = {
        flowKey: this.flowKey,
        fields: this.columns.map(v2 => v2.name),
        conditions: {},
      };
      [this.viewType] = Object.keys(this.views);
    });
  },
  methods: {
    init() {
      return this.$axios.get(this.configCgi.instanceDict).then((res) => {
        if (res) {
          this.orderTypeList = res;
          this.showTab = Object.keys(res).length > 1;
          [this.flowKey] = Object.keys(res);
        }
        return this.getConfig();
      });
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
      const origin = config.map(v => ({
        ...v,
        show: v.show || false,
        width: this.calcWidth(v.label, { count: v.size > 0 ? v.size : 8 }),
      }));
      if (this.views.table && this.viewType === 'table') {
        const [{ clientWidth }] = $('.table-wrap');
        const target = origin.filter(v => v.show);
        const totalWidth = target.map(v => v.width).reduce((prev, cur) => prev + cur);
        if (clientWidth > totalWidth) {
          const addWidth = (clientWidth - totalWidth) / target.length;
          return origin.map(item => ({
            ...item,
            width: item.width + addWidth,
          }));
        }
      }
      this.saveColumns[this.flowKey] = origin;
      return origin;
    },
    getConfig() {
      if (this.saveColumns[this.flowKey]) {
        this.columns = this.saveColumns[this.flowKey];
        return this.search();
      }
      return this.$axios.get(this.configCgi.getConfigCgi, {
        flowKey: this.flowKey,
      }).then((res) => {
        this.columns = res.length ? this.handleConfig(res) : [];
        this.search();
      });
    },
    search(conditions = {}) {
      this.query = {
        ...this.query,
        conditions,
      };
    },
    download(ids) {
      if (ids) { // 导出所选
        const value = { field: 'id', relation: 'AND', operator: 'IN', value: ids };
        if (!this.query.conditions.default) {
          this.query.conditions.default = [value];
        } else {
          const fields = this.query.conditions.default.map(v => v.field);
          const index = fields.indexOf('id');
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
      this.$axios.download(this.configCgi.exportCgi, {
        ...this.query,
      });
    },
  },
};
</script>
