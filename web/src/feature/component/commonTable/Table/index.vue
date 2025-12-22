<template>
  <div class="table-wrap">
    <div
      v-if="showSelectionHeader"
      class="selection-header"
    >
      <i
        class="tn-icon tn-icon-close"
        @click="showSelectionHeader = false"
      />
      <span class="selection-text">
        已选择 {{ selection.length }} 项
      </span>
      <span class="right-btns">
        <slot
          name="buttons"
          :data="selection"
        />
      </span>
    </div>
    <el-table
      ref="table"
      v-loading="loading"
      :cell-height="49"
      element-loading-text="正在获取数据"
      :data="list"
      style="width: 100%;"
      @selection-change="select"
      @sort-change="sortChange"
    >
      <el-table-column
        v-if="tableConfig.selectable !== false && showCheckBox"
        type="selection"
        width="64"
        fixed
      />

      <el-table-column
        v-if="tableConfig.expandable"
        type="expand"
        fixed
      >
        <template slot-scope="props">
          <slot
            name="expand"
            :row="props.row"
          >
            No Data
          </slot>
        </template>
      </el-table-column>
      <el-table-column
        v-for="item in curColumns"
        :key="item[columnKey]"
        :column-key="item[columnKey]"
        :prop="item[columnKey]"
        v-bind="item"
        :show-overflow-tooltip="true"
      >
        <template
          v-if="isTableDef(item)"
          v-slot="scope"
        >
          <slot
            :name="item[columnKey]"
            :data="scope"
            :row="scope.row"
          >
            <template v-if="item.formatter">
              {{ item.formatter(scope.row, item, scope.row[item[columnKey]]) }}
            </template>
            <template v-else-if="['user', 'mutiUser'].includes(item.type)">
              {{ formatUserData(scope.row[item[columnKey]]) }}
            </template>
            <template v-else-if="item.jump">
              <a
                href="javascript:void(0)"
                @click="item.jump(item, scope.row, table)"
              >
                {{ scope.row[item[columnKey]] }}
              </a>
            </template>
            <template v-else-if="item.mutiRow && scope.row[item[columnKey]]">
              <div
                v-for="(l,ind) in scope.row[item[columnKey]].split(';')"
                :key="ind"
              >
                <span>{{ l }}</span>
              </div>
            </template>
            <template v-else-if="item.type === 'textarea'">
              <el-input
                v-model="scope.row[item[columnKey]]"
                type="textarea"
                readonly
                resize="none"
                :autosize="{ maxRows: 5 }"
                border-type="plain"
              />
            </template>
            <template v-else-if="item.mutiRow && scope.row[item[columnKey]]">
              <div
                v-for="(l,ind) in scope.row[item[columnKey]].split(';')"
                :key="ind"
              >
                <span>{{ l }}</span>
              </div>
            </template>
            <template v-else>
              {{ scope.row[item[columnKey]] }}
            </template>
          </slot>
        </template>
        <template #header>
          <span
            class="middle-wrap"
          >
            <span>{{ item.label }}</span>
            <span v-if="item.manualFixed">
              <i
                class="pointer tn-icon-pin-inclined text-primary"
                @click="toggleFixed(item[columnKey])"
              />
              <i
                class="pointer tn-icon-pin text-light"
                @click="toggleFixed(item[columnKey])"
              />
            </span>
          </span>
        </template>
      </el-table-column>
    </el-table>
    <div
      class="pagination el-row is-justify-end el-row--flex"
      :style="{margin: 0}"
    >
      <el-pagination
        :current-page.sync="currentPage"
        :page-size.sync="limit"
        layout="total, prev, pager, next, sizes, jumper"
        :total="total"
        :page-sizes="[10, 20, 30, 40, 50, 100, 200, 500]"
        styled
        background
        @size-change="v => filterHandler({limit: v})"
        @current-change="v => filterHandler({currentPage: v})"
      />
    </div>
  </div>
</template>
<script>
import { find, every, reduce, throttle } from 'lodash';
import mixin from '../script/mixin';

export default {
  inject: ['configCgi', 'tableConfig'],
  mixins: [mixin],
  props: {
    columns: {
      type: Array,
      default: () => [],
    },
    columnKey: {
      type: String,
      default: 'name',
    },
    localData: {
      type: Array,
      required: false,
      default: () => ([]),
    },
    query: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    const curColumns = this.calcCurColumns();
    return {
      rawRst: [],
      rst: null,
      loading: false,
      currentPage: 1,
      limit: 10,
      total: 0,
      selection: [],
      showSelectionHeader: false,
      curColumns,
      dialogVisible: {},
      sortKey: void 0,
      sortOrder: void 0,
      table: this.tableConfig.table || 'table',
      rights: this.tableConfig.rights,
      hasUrlParam: false,
    };
  },
  computed: {
    list() {
      return this.rst || this.localData;
    },
    showCheckBox() {
      return (this.hasRights('canDel') || this.hasRights('canExport')) && !this.tableConfig.hideCheckBox;
    },
  },
  watch: {
    query() {
      this.refresh();
    },
    selection(v) {
      this.showSelectionHeader = !!v.length;
    },
    columns() {
      this.curColumns = this.calcCurColumns();
    },
  },
  mounted() {
    if (this.tableConfig.searchParams && Object.keys(this.tableConfig.searchParams).length !== 0) {
      this.hasUrlParam = true;
    }
  },
  methods: {
    formatUserData(data) {
      return Array.isArray(data) ? data.map(v => v.userName).join(';') : data;
    },
    isTableDef(item) {
      return item.type !== 'selection' && item.type !== 'index' && item.type !== 'expand';
    },
    calcCurColumns() {
      const rst = this.columns.map((column) => {
        const rst = {};
        if (column.type === 'selection') {
          rst.fixed = true;
        }
        return {
          ...rst,
          ...column,
        };
      });
      // 隐藏列后表格最少占 100%
      const isFixWidth = every(rst, column => !!column.width);
      if (isFixWidth) {
        const totalWidth = reduce(rst, (memo, column) => memo + column.width, 0);
        if (this.$refs.table) {
          if (this.$refs.table.$el.offsetWidth > totalWidth) {
            delete rst[rst.length - 1].width;
          }
        }
      }
      return rst;
    },
    select(selection) {
      this.selection = selection;
    },
    reload(flag) {
      if (this.hasInit) {
        this.refresh(flag);
      }
    },
    refresh() {
      this.filterHandler();
    },
    filterHandler(v = {}) {
      if (!v.currentPage) {
        // eslint-disable-next-line no-param-reassign
        v.currentPage = 1;
        this.currentPage = 1;
      }
      const params = {
        limit: this.limit,
        sortKey: this.sortKey,
        sortOrder: this.sortOrder,
        ...this.query,
        ...v,
      };
      params.start = (v.currentPage - 1) * params.limit;
      if (this.hasUrlParam) {
        this.getData({ ...params, ...this.tableConfig.searchParams });
      } else {
        this.getData(params);
      }
    },
    getData: throttle(function (params) {
      if (this.tableConfig.customAjax?.query) {
        return this.tableConfig.customAjax?.query(params, this.configCgi.queryCgi)
          .then((data) => {
            this.total = data.count;
            this.rst = data.list;
            this.tableConfig.searchParams = {};
            this.hasUrlParam = false;
          });
      }
      this.$axios.post(this.configCgi.queryCgi, params).then((data) => {
        this.total = data.count;
        this.rst = data.list;
        this.tableConfig.searchParams = {};
        this.hasUrlParam = false;
      });
    }, 1500, { leading: true, trailing: true }),
    toggleFixed(v) {
      const match = find(this.curColumns, {
        [this.columnKey]: v,
      });
      const index = this.curColumns.indexOf(match);
      if (match.fixed) {
        if (match.manualFixed === 'right') {
          this.curColumns.forEach((column, i) => {
            if (i >= index) {
              this.$delete(column, 'fixed');
            }
          });
        } else {
          this.curColumns.forEach((column, i) => {
            if (i <= index) {
              this.$delete(column, 'fixed');
            }
          });
        }
      } else {
        if (match.manualFixed === 'right') {
          this.curColumns.forEach((column, i) => {
            if (i >= index) {
              this.$set(column, 'fixed', match.manualFixed);
            }
          });
        } else {
          this.curColumns.forEach((column, i) => {
            if (i <= index) {
              this.$set(column, 'fixed', match.manualFixed);
            }
          });
        }
      }
    },
    sortChange({ prop, order }) {
      this.sortKey = prop;
      this.sortOrder = prop ? (order === 'descending' ? 'desc' : 'asc') : null;
      this.filterHandler();
    },
  },
};
</script>
<style lang="scss" scoped>
@import './style';
.el-textarea:after, .el-textarea:before{
  display: none;
}
</style>
