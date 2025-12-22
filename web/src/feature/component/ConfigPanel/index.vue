<template>
  <div>
    <form-modal
      v-if="bizColumns.length"
      :id="id"
      :visible.sync="showModal"
      :data="curNotified"
      :pre-data="preData"
      :columns="bizColumns"
      :is-edit="isEdit"
      :text="text"
      :table="table"
      :options="opts"
      :conditions="conditions"
      @success="refresh"
    />
    <data-modal
      :columns="tableColumns"
      :visible.sync="showData"
      :data="curData.data"
      :main="curData.main"
    />
    <div class="config-panel-wrap">
      <div
        v-if="$slots.default"
        class="config-panel-sidebar"
      >
        <slot />
      </div>
      <div class="config-panel-body">
        <el-block>
          <el-block
            inner
            border
          >
            <table-toolbar
              ref="toolbar"
              :codes="codes"
              :roles="roles"
              :table="remoteTable"
              :text="text"
              @export="download"
              @refresh="refresh"
            >
              <div slot="buttons">
                <el-button
                  v-if="hasRights('canAdd')"
                  :auth-roles="roles.write"
                  :auth-right-code="codes.xz"
                  :disabled="!canAdd"
                  type="primary"
                  @click="add"
                >
                  新增{{ text }}
                </el-button>
              </div>

              <!-- 统计类别 -->
              <div slot="statistics">
                <statistics
                  v-if="hasRights('showStatistics')"
                  :table="table"
                />
              </div>

              <div
                slot="handlers"
                :auth-roles="roles.write"
                :auth-right-code="codes.zdyzs"
              >
                <show-setting
                  :table-col="bizColumns"
                  :search-col="searchData"
                  :table="table"
                  :show-search-setting="showSearchSetting"
                  :show-table-setting="showTableSetting"
                  :is-customize-local="isCustomizeLocal"
                  :is-table-setting-merge="isTableSettingMerge"
                  @change="changeTableColumns"
                  @save="saveTableColumns"
                  @changeFilter="changeFilterColumns"
                  @saveFilter="saveFilterColumns"
                />
              </div>
            </table-toolbar>
          </el-block>
          <el-block
            inner
            collapsible
          >
            <template slot="header">
              高级筛选
            </template>
            <advanced-search
              ref="advancedSearch"
              :columns="advancedSearchColumns"
              :has-condition="hasExtraCondition"
              :show-search-setting="showSearchSetting"
              :is-customize-local="isCustomizeLocal"
              :table="table"
              :codes="codes"
              :roles="roles"
              :preset-form-data="presetFormData"
              @search="search"
            />
          </el-block>
          <data-table
            :id="id"
            ref="table"
            :columns="tableColumns"
            :border-prop="borderProp"
            :query="query"
            :table="table"
            :codes="codes"
            :roles="roles"
            @edit="edit"
            @expand="expand"
            @export="download"
            @success="getTreeData"
          >
            <template v-slot:customizeBtn="{data}">
              <slot
                name="customizeBtn"
                :data="data"
              />
            </template>
          </data-table>
        </el-block>
      </div>
    </div>
  </div>
</template>
<script>
import { find, map, pick, cloneDeep, each, isPlainObject } from 'lodash';
import { flatten, filter, map as mapFp, get as getFp, omit, last } from 'lodash/fp';
import mixin from 'component/script/mixin';
import { get, set } from 'component/script/storage';
import { createTableConfig, getNameById, getTextByTable, getRemoteTable } from 'component/script/configHelper';
import TableToolbar from 'component/ConfigTool/Toolbar';
import AdvancedSearch from './AdvancedSearch';
import ShowSetting from './ShowSetting';
import FormModal from './FormModal';
import DataModal from './DataModal';
import DataTable from './DataTable';
import Statistics from './statistics';
import configMixin from './mixin';

export default {
  components: {
    DataModal,
    FormModal,
    DataTable,
    ShowSetting,
    TableToolbar,
    Statistics,
    AdvancedSearch,
  },
  inject: ['configCgi', 'commonCgi'],
  mixins: [mixin, configMixin],
  props: {
    table: {
      type: String,
      required: true,
    },
    conditions: {
      type: [Array, undefined],
      default: () => undefined,
    },
    tableText: {
      type: String,
      default: '',
    },
    codes: {
      type: Object,
      default: () => ({}),
    },
    canAdd: {
      type: Boolean,
      default: true,
    },
    showSearchSetting: {
      type: Boolean,
      default: true,
    },
    showTableSetting: {
      type: Boolean,
      default: true,
    },
    isCustomizeLocal: {
      type: Boolean,
      default: false,
    },
    isTableSettingMerge: {
      type: Boolean,
      default: false,
    },
    preData: {
      type: Object,
      default: () => ({}),
    },
    options: {
      type: Object,
      default: () => ({}),
    },
    /**
     * 预设的“高级筛选”表单数据，传递给 advanced-search 组件
     */
    presetFormData: {
      type: Object,
      default: () => ({}),
    },
    borderProp: {
      type: Boolean,
      default: false,
    },
    roles: { // 增加：应用矩阵的权限
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      show: false,
      showModal: false,
      showData: false,
      showFilter: false,
      isEdit: void 0,
      clientWidth: 800,
      bizColumns: [],
      tableColumns: [],
      mutiFields: [],
      curNotified: void 0,
      query: {},
      opts: {},
      dic: {},
      curData: {},
      remoteTable: void 0,
      searchData: [], // 传递给自定义查询条件tab数据
    };
  },
  computed: {
    id() {
      return `${this.table}_id`;
    },
    text() {
      return getTextByTable(this.table) || this.tableText;
    },
    showFields() {
      let fields = this.bizColumns |> mapFp('fields') |> flatten |> filter(item => ((item.showContent && item.show) || this.isKey(item.name))) |> mapFp('name');
      fields = fields.filter(field => (!this.isKey(field) || fields.includes(getNameById(field))));
      let list = [
        ...new Set(fields),
        this.id,
      ];
      // 兼容资源管理-服务器资源接口fields字段不需要以下参数
      try {
        if (this.table === 'server' && window.location.href.indexOf('resourcechart/setting-manager-querydata') !== -1) {
          list = list.filter(item => ['mozu_name', 'mozu_alias', 'mozu_number', 'building_name', 'room_name',
            'building_shortName', 'building_code', 'room_No', 'room_code'].indexOf(item) === -1);
        }
        return list;
      } catch (error) {
        return [
          ...new Set(fields),
          this.id,
        ];
      }
    },
    hasExtraCondition() {
      return !!this.conditions;
    },
    advancedSearchColumns() {
      return this.showSearchSetting ? this.searchData : this.bizColumns;
    },
  },
  watch: {
    conditions() {
      this.loadFields();
    },
    table() {
      this.loadFields();
    },
  },
  mounted() {
    this.$nextTick(() => {
      this.clientWidth = $('.app-container')[0].clientWidth;
      this.loadFields();
    });
  },
  methods: {
    loadFields() {
      this.init = false;
      this.$axios.post(this.configCgi.setMgrFields, {
        table: this.table,
      }).then((remoteData) => {
        const remoteTables = map(remoteData, 'table');
        this.remoteTable = getRemoteTable(this.table, remoteTables);
        const localData = createTableConfig(this.table, remoteTables, find(remoteData, {
          table: this.table,
        })?.fields);
        const savedData = get(this.table);
        const tableColumns = map(localData, (fields, table) => {
          const remoteItem = find(remoteData, {
            table,
          });
          const savedItem = find(savedData, {
            table,
          }) || {};
          return {
            table,
            fields: fields.map((field) => {
              const match1 = find(remoteItem?.fields, {
                name: field.name,
              }) || {};
              const match2 = find(savedItem?.fields, {
                name: field.name,
              }) || {};
              let showIndex; let showContent;

              if (this.isCustomizeLocal) {
                showIndex = match2.showIndex === undefined ? true : match2.showIndex;
                showContent = match2.showContent === undefined ? true : match2.showContent;
              } else {
                showIndex = match1.showIndex;
                showContent = match1.showContent;
              }

              return {
                ...match1,
                ...field,
                ...match2,
                showIndex,
                showContent,
                type: !field.type ? match1.type : (isPlainObject(field.type) ? field.type.type : field.type),
                isMulti: isPlainObject(field.type) && field.type ? field.type.isMulti : false,
              };
            }),
          };
        });
        const searchColumns = map(tableColumns, ({ fields, table }) => ({
          table,
          fields: fields.filter((fields) => {
            const isFilter = Array.isArray(fields.isFilter) ? fields.isFilter.includes(this.table) : fields.isFilter;
            return isFilter;
          }),
        })).filter(({ fields }) => fields.length > 0);
        this.searchData = searchColumns;
        this.changeTableColumns(tableColumns);
        this.getMultiColumn(tableColumns);
        this.search();
      });
    },
    search(conditions = []) {
      if (!this.init) {
        this.init = true;
      }
      this.load({ conditions: [...conditions, ...this.conditions || []] });
    },
    load(query) {
      if (this.init) {
        this.query = {
          table: this.table,
          fields: this.showFields,
          conditions: this.query?.conditions,
          ...query,
        };
      }
    },
    refresh() {
      this.getTreeData();
      if (this.init) {
        this.$refs.table.refresh();
      } else {
        this.query = {
          table: this.table,
          fields: this.showFields,
        };
        this.init = true;
      }
    },
    getTreeData() {
      this.$emit('refresh');
    },
    saveFilterAndTableColumns(tag, data) {
      if (this.isCustomizeLocal) {
        this.setCustomizeLocal(data);
      } else {
        const value = map(data, item => ({ ...item,
          fields: item.fields.map(field => pick(field, ['name', 'showContent', 'showIndex'])),
        }));
        const params = { table: this.table, type: tag, data: value };
        this.$axios.post(this.configCgi.updatePreference, params).then(() => {
          if (tag === 'index') {
            this.$refs.advancedSearch.reset();
          }
        });
      }
    },
    saveTableColumns(data) {
      return this.saveFilterAndTableColumns('show', data);
    },
    saveFilterColumns(data) {
      return this.saveFilterAndTableColumns('index', data);
    },
    setCustomizeLocal(data) {
      try {
        const value = map(data, item => ({
          ...item,
          fields: item.fields.map(field => pick(field, ['name', 'show', 'showContent', 'showIndex'])),
        }));
        set(this.table, value);
      } catch {
        this.$message.error('自定义筛选保存失败');
      }
    },
    changeFilterColumns(data) {
      this.searchData = data;
    },
    sortArrBySeqNum(arrList) { // 表格字段排序
      const ret = [...arrList.map(arr => this.sortBySeqNum(arr))];
      return ret;
    },
    sortBySeqNum(arr) { // 编辑modal字段排序
      return [...arr.sort((item1, item2) => item1.seqNumber - item2.seqNumber)];
    },
    filterFields(arr) {
      return [...arr.filter(v => !v.isFrontEndField)];
    },
    formatWidth(columns) {
      const list = cloneDeep(columns) |> mapFp('fields') |> flatten |> filter(item => (item.showContent && item.show && !item.notShowItem)) |> mapFp('width');
      const len = (this.hasRights('canEdit') || this.hasRights('canDel')) ? list.length + 1 : list.length;
      const sumLen = list.reduce((prev, cur) => prev + cur);
      const num = this.clientWidth > 1600 ? 10 : this.clientWidth > 1000 ? 8 : 6;

      if (len <= num) {
        const wid = Math.floor((this.clientWidth - 64 - sumLen) / len);
        return cloneDeep(columns).map(({ table, fields }) => ({
          table,
          fields: fields.map(item => ({
            ...item,
            width: item.showContent && item.show && !item.notShowItem ? item.width + wid : item.width,
          }
          )),
        }));
      }
      return columns;
    },
    changeTableColumns(data) {
      this.bizColumns = data;
      const tableContent = cloneDeep(data).reverse() |> this.formatWidth |> mapFp('fields') |> this.sortArrBySeqNum |> flatten |> filter(item => item.showContent && item.show && !item.notShowItem);
      const tableColumns = [...tableContent.map((column) => {
        if (column.type === 'bool' && !column.formatter) {
          return {
            ...column,
            formatter(row, column, v) {
              return {
                1: '是',
                0: '否',
              }[v] || v;
            },
          };
        }
        return {
          ...column,
        };
      })];
      tableColumns.unshift({
        type: 'selection',
        width: 64,
      });
      const authRightCode = (!this.codes.bj && !this.codes.sc) ? '' : `${this.codes.bj},${this.codes.sc}`;
      const authRoles = (!this.roles.write) ? '' : `${this.roles.write}`;
      if (this.hasRights('canEdit') || this.hasRights('canDel') || this.hasRights('customizeBtn')) {
        tableColumns.push({
          name: 'actions',
          label: '操作',
          width: 125,
          manualFixed: 'right',
          authRightCode,
          authRoles,
        });
      }
      this.tableColumns = tableColumns;
      this.load();
    },
    getMultiColumn(data) {
      this.mutiFields = [...data |> last |> getFp('fields') |> this.filterByMuti |> mapFp('name')];
    },
    filterByMuti(arr) {
      return arr.filter(item => item.type === 'mutiInt');
    },
    download(ids) {
      let fields = this.tableColumns.filter(item => item.name && item.name !== 'actions').map(item => item.name);
      if (!ids && this.table === 'uPos' && fields.length > 15) {
        this.$alert('U位的全量导出最大列数不能超过15条');
        return;
      }
      const conditions = [];
      if (ids) {
        // 兼容资源管理-服务器资源接口fields字段不需要以下参数
        try {
          if (this.table === 'server' && this.id === 'server_id'
        && window.location.href.indexOf('resourcechart/setting-manager-querydata') !== -1) {
            ids = ids.map(item => String(item));
          }
          conditions.push({
            name: this.id,
            value: ids,
          });
        } catch (error) {
          conditions.push({
            name: this.id,
            value: ids,
          });
        }
      }
      // 兼容资源管理-服务器资源接口fields字段不需要以下参数
      if (this.table === 'server' && window.location.href.indexOf('resourcechart/setting-manager-querydata') !== -1) {
        fields = fields.filter(item => ['mozu_name', 'mozu_alias', 'mozu_number', 'building_name', 'room_name',
          'building_shortName', 'building_code', 'room_No', 'room_code'].indexOf(item) === -1);
      }
      const params = {
        table: this.table,
        fields,
        conditions: [...conditions, ...this.query.conditions], // 按筛选条件导出
      };
      this.$axios.download(this.configCgi.exportData, params, true, { restAxios: { timeout: 60000 * 2 } });
    },
    add() {
      const path = this.getPath();
      if (path) {
        this.jump(path);
      } else {
        const fields = [...this.bizColumns |> last |> getFp('fields')].map(field => pick(field, ['name', 'defaultValue']));
        const curNotified = {};
        fields.forEach((v) => {
          const { name, defaultValue } = v;
          if (!defaultValue || !defaultValue.length) return;
          curNotified[name] = defaultValue;
        });
        this.curNotified = {
          ...curNotified,
        };
        this.opts = this.options;
        this.showModal = true;
        this.isEdit = false;
      }
    },
    expand(data, main) {
      this.showData = true;
      this.curData = {
        data,
        main,
      };
    },
    edit(data) {
      const fields = [...this.bizColumns |> last |> getFp('fields') |> this.sortBySeqNum |> this.filterFields |> mapFp('name'), this.id];
      // 外键需要拿对应的 name，用于 select 的显示
      fields.forEach((field) => {
        if (this.isFK(field)) {
          fields.push(getNameById(field));
        }
      });
      const query = {
        table: this.table,
        conditions: [{
          name: this.id,
          value: [data[this.id]],
        }],
        fields,
      };
      this.$axios.post(this.configCgi.getMgrList, {
        ...query,
        start: 0,
        limit: 1,
      }).then((data) => {
        // eslint-disable-next-line prefer-destructuring
        const row = data.list[0];
        const omits = [];
        const opts = {};
        each(row, (value, key) => {
          if (this.isFK(key)) {
            const keys = [...this.bizColumns |> last |> getFp('fields')];
            const match = find(keys, {
              name: key,
            });
            const name = getNameById(key);
            omits.push(name);
            if (value) {
              if (this.mutiFields.includes(key)) {
                const tempArr = value.split(';');
                const tempVal = row[name].split(';');

                const item = {};
                for (let i = 0; i < tempArr.length; i++) {
                  item[tempArr[i]] = tempVal[i];
                }
                opts[key] = item;
              } else if (match.type === 'cascader') {
                opts[key] = [{
                  value,
                  label: row[name],
                }];
              } else {
                opts[key] = {
                  [value]: row[name],
                };
              }
            }
          }
        });
        const rst = omit(omits)(row);
        const multiRst = {};
        each(rst, (value, key) => {
          multiRst[key] = this.mutiFields.includes(key) && value.length ? value.split(';') : value;
        });
        this.curNotified = multiRst;
        this.opts = opts;
        this.showModal = true;
        this.isEdit = true;
      });
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
