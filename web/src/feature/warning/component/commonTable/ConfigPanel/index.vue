<template>
  <div class="el-common-table">
    <div class="config-panel-wrap">
      <div
        v-if="$slots.default"
        class="config-panel-sidebar"
      >
        <slot />
      </div>
      <div class="config-panel-body">
        <advanced-search
          v-if="showAdvance"
          :advanced-columns="bizColumns"
          :show-customize-search="showCustomizeSearch"
          :play="play"
          :codes="codes"
          :roles="roles"
          @search="search"
          @doelse="doelse"
        />
        <table-toolbar
          v-if="showTabletoolbar"
          ref="toolbar"
          style="height: 32px;"
          :show-search="tableConfig.showSearch"
          :placeholder="tableConfig.placeHolder"
          :codes="codes"
          :roles="roles"
          @search="v => keyWordSearch({keyword: v })"
        >
          <div slot="buttons">
            <slot name="extraButtons" />
            <el-button
              v-if="hasRights('canAdd')"
              v-appmatrixauth="roles.write"

              type="primary"
              @click="add"
            >
              新增{{ text }}
            </el-button>
          </div>
          <div
            slot="handlers"
          >
            <show-setting
              v-if="tableConfig.showSetting"
              :columns="bizColumns"
              @change="setColumns"
              @save="saveChange"
            />
          </div>
          <template>
            <div
              v-if="hasRights('canExport') || hasRights('canImport')"
              slot="more"
            >
              <config-popover
                :codes="codes"
                :roles="roles"
                :text="text"
                @uploadSuccess="refresh"
                @export="download"
              />
            </div>
          </template>
        </table-toolbar>
        <data-table
          ref="table"
          :columns="tableColumns"
          :local-data="localData"
          :play="play"
          :query="query"
          :codes="codes"
          :roles="roles"
          :show-table-select="showTableSelect"
          @edit="edit"
          @expand="expand"
          @export="download"
        />
      </div>
    </div>
  </div>
</template>
<script>
import { map, pick, cloneDeep, find } from 'lodash';
import mixin from '../script/mixin';
import { set, get } from '../script/storage';
import TableToolbar from '../Table/Toolbar';
import AdvancedSearch from './AdvancedSearch';
import ShowSetting from './ShowSetting';
// import FormModal from './FormModal';
import DataTable from './DataTable';
import ConfigPopover from './ConfigPopover';
import getEdgeRequest from '../../../../utils/request';
import { eventBus } from '../script/eventBus';

export default {
  name: 'ElCommonTable',
  components: {
    // FormModal,
    DataTable,
    ShowSetting,
    TableToolbar,
    ConfigPopover,
    AdvancedSearch,
  },
  provide() {
    return {
      tableConfig: this.tableConfig,
      configCgi: this.configCgi,
      commonCgi: this.commonCgi,
    };
  },
  mixins: [mixin],
  props: {
    codes: {
      type: Object,
      default: () => ({}),
    },
    roles: {
      type: Object,
      default: () => ({}),
    },
    conditions: {
      type: Object,
      default: () => ({}),
    },
    columns: {
      type: Array,
      default: () => ([]),
    },
    showAdvance: {
      type: Boolean,
      default: true,
    },
    showTabletoolbar: {
      type: Boolean,
      default: true,
    },
    localData: {
      type: Array,
      required: false,
      default: () => ([]),
    },
    play: {
      type: Boolean,
      default: false,
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
    showTableSelect: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      show: false,
      showModal: false,
      showData: false,
      showFilter: false,
      bizColumns: [],
      tableColumns: [],
      mutiFields: [],
      curNotified: void 0,
      query: {},
      opts: {},
      dic: {},
      curData: {},
      searchData: [],
      tableConfigs: this.tableConfig,
    };
  },
  computed: {
    text() {
      return this.tableConfig.text || '';
    },
    rights() {
      return this.tableConfig.rights;
    },
    showCustomizeSearch() {
      return this.tableConfig.showCustomizeSearch || true;
    },
  },
  watch: {
    'tableConfigs.refreshNow': {
      handler() {
        this.search(this.query);
      },
      deep: true,
      immediate: true,
    },
  },
  mounted() {
    this.loadFields();
  },
  methods: {
    loadFields() {
      const localData = this.handleConfig(this.columns);
      const savedData = get(this.tableConfig.table);
      const columns = map(localData, (item) => {
        const savedItem = find(savedData, {
          name: item.name,
        });
        return {
          ...item,
          ...savedItem,
        };
      });
      this.setColumns(columns);
      this.search();
    },
    keyWordSearch(v) {
      this.query = { ...this.query, ...v };
      const key = this.tableConfig?.inputSearchKey;
      if (key) {
        this.query[key] = [v.keyword][0].length === 0 ? undefined : [v.keyword];
      }
    },
    search(conditions = {}) {
      const searchParams = this.tableConfig?.searchParams || {};
      if (Object.keys(searchParams).length !== 0) {
        this.query = {
          ...conditions,
          ...searchParams,
          ...this.conditions,
        };
      } else {
        this.query = {
          ...conditions,
          ...this.conditions,
        };
      }
    },
    refresh(axiosLoading) {
      if (this.init) {
        this.$refs.table.refresh(axiosLoading);
      } else {
        this.query = {
          table: this.table,
          fields: this.showFields,
        };
        this.init = true;
      }
    },
    saveChange(data) {
      set(this.tableConfig.table, map(data, item => pick(item, ['name', 'showInTable', 'showInSearch'])));
    },
    sortBySeqNum(arr) { // 表格字段排序
      return [...arr.sort((v1, v2) => v1.seqNumber - v2.seqNumber)];
    },
    setColumns(data) {
      this.bizColumns = data;
      const tableContent = this.sortBySeqNum(cloneDeep(data)).filter(item => item.showInTable);
      const tableColumns = [...tableContent.map((column) => {
        if (column.type === 'bool' && !column.formatter) {
          return {
            ...column,
            formatter(row, column, v) {
              return v === '1' ? '是' : '否';
            },
          };
        }
        return {
          ...column,
        };
      })];
      if (this.hasRights('canEdit') || this.tableConfig.hasDetail) {
        tableColumns.push({
          name: 'actions',
          label: '操作',
          width: 160,
          manualFixed: 'right',
          fixed: 'right',
        });
      }
      this.tableColumns = tableColumns;
    },
    filterByMuti(arr) {
      return arr.filter(item => item.type === 'mutiInt');
    },
    download(ids, tag) {
      const params = ids && ids.length ? { ids } : { ...this.query, ...this.tableConfig.defaultParams }; // 按筛选条件导出
      Object.keys(params).forEach((item) => {
        if (params[item] === 'true') {
          params[item] = true;
        }
        if (params[item] === 'false') {
          params[item] = false;
        }
      });
      if (this.play) {
        this.playSuccess();
      } else {
        const pageUrl = window.location.href;
        if (pageUrl.includes('timpage/warning-history')) {
          params.mozuId = this.tableConfig.searchParams.mozuId;
        }
        if (this.configCgi.extraCgi && tag === 'analysis') {
          getEdgeRequest(this.$axios, this.tableConfig.searchParams.mozuId).download(this.configCgi.extraCgi, params);
        } else {
          getEdgeRequest(this.$axios, this.tableConfig.searchParams.mozuId).download(this.configCgi.exportCgi, params);
        }
      }
    },
    add() {
      this.curNotified = void 0;
      this.opts = {};
      this.showModal = true;
    },
    expand(data, main) {
      this.showData = true;
      this.curData = {
        data,
        main,
      };
    },
    edit(data) {
      this.curNotified = data;
      this.showModal = true;
    },
    doelse(data) {
      eventBus.$emit('showModal', { type: data, data: { ...this.query, ...this.tableConfig.defaultParams } });
    },
  },
};
</script>
<style lang="scss" scoped>
  .config-panel-wrap {
    display: flex;

    .config-panel-sidebar {
      position: relative;
      width: 200px;
    }

    .config-panel-body {
      overflow: auto;
      flex: 1;
    }
  }
</style>
