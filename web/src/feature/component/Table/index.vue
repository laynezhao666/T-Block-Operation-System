<template>
  <div class="table-wrap">
    <div
      v-if="showSelectionHeader"
      class="selection-header"
    >
      <i
        class="tn-icon tn-icon-close"
        @click="showSelectionHeader = false;clearSelection()"
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
      :border="borderProp"
      :big-data="bigData"
      :cell-height="49"
      element-loading-text="正在获取数据"
      :data="list"
      style="width: 100%;"
      :max-height="tableHeight"
      :row-key="rowKey"
      @filter-change="filterChange"
      @selection-change="select"
    >
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
          >
            <template v-if="item.jump">
              <a
                href="javascript:void(0)"
                @click="item.jump(item, scope.row, curTable)"
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
            <template
              v-else-if="item.statusColor"
            >
              <div :style="{'color': item.statusColor(scope.row[item[columnKey]]) ? '#37c629':''}">
                <span>{{ scope.row[item[columnKey]] }}</span>
              </div>
            </template>
            <template v-else>
              {{ scope.row[item[columnKey]] }}
            </template>
          </slot>
        </template>
        <template
          v-if="item.manualFixed"
          slot="header"
        >
          <span
            class="middle-wrap"
            :auth-right-code="item.authRightCode"
          >
            <span>{{ item.label }}</span>
            <i
              class="pointer tn-icon-pin-inclined text-primary"
              @click="toggleFixed(item[columnKey])"
            />
            <i
              class="pointer tn-icon-pin text-light"
              @click="toggleFixed(item[columnKey])"
            />
          </span>
        </template>
      </el-table-column>
    </el-table>
    <div
      v-if="paging !== 'none'"
      class="pagination el-row is-justify-end el-row--flex"
      :style="{margin: 0}"
    >
      <el-pagination
        :current-page.sync="pagination.curpage"
        :page-size.sync="pagination.pagesize"
        layout="total, prev, pager, next, sizes, jumper"
        :total="total"
        :page-sizes="[10, 20, 30, 40, 50, 100, 200, 500]"
        styled
        background
      />
    </div>
  </div>
</template>
<script>
import { merge, find, every, reduce, mapValues, filter } from 'lodash';
import mixin, { arrToBin } from 'component/script/mixin';

export default {
  mixins: [mixin],
  props: {
    columns: {
      type: Array,
      default: () => [],
    },
    columnKey: {
      type: String,
      default: 'prop',
    },
    cgi: {
      type: String,
      default: '',
    },
    imgUrl: {
      type: String,
      default: '',
    },
    paging: {
      type: String,
      default: 'none',
    },
    method: {
      type: String,
      default: 'get',
    },
    query: {
      type: Object,
      default: () => ({}),
    },
    search: {
      type: Object,
      default: () => ({}),
    },
    // eslint-disable-next-line vue/require-default-prop
    data: Array,
    listField: {
      type: String,
      default: '',
    },
    rowKey: {
      type: String,
      default: 'id',
    },
    globalLoading: {
      type: Boolean,
      default: false,
    },
    manualInit: {
      type: Boolean,
      default: false,
    },
    curTable: {
      type: String,
      default: '',
    },
    borderProp: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    const curColumns = this.calcCurColumns();
    return {
      rawRst: [],
      rst: [],
      pagination: {
        curpage: 1,
        pagesize: 10,
      },
      bigData: false,
      loading: false,
      total: 0,
      filterQuery: {},
      selection: [],
      showSelectionHeader: false,
      curColumns,
      dialogVisible: {},
      tableHeight: 555,
    };
  },
  computed: {
    list() {
      return this.data || this.rst;
    },
  },
  watch: {
    pagination: {
      handler() {
        if (this.pagination.pagesize >= 200) {
          this.bigData = true;
        } else {
          this.bigData = false;
        }
      },
      deep: true,
    },
    'pagination.pagesize'() { // pagesize变化后，curpage必须置为1
      this.pagination.curpage = 1;
      this.pageChangeEvent();
    },
    'pagination.curpage'() {
      this.pageChangeEvent();
    },
    search: {
      handler(v) {
        this.filterRst = filter(this.rawRst, item => item[v.field].indexOf(v.keyword) > -1);
        this.calcLocalList();
      },
      deep: true,
    },
    query() {
      this.refresh(true);
    },
    filterQuery() {
      this.reload(true);
    },
    selection(v) {
      this.showSelectionHeader = !!v.length;
    },
    columns() {
      this.curColumns = this.calcCurColumns();
    },
  },
  mounted() {
    if (!this.manualInit) {
      this.refresh();
    }
  },
  methods: {
    clearSelection() {
      this.$refs.table.clearSelection();
    },
    pageChangeEvent() {
      if (this.paging === 'remote') {
        // 远程分页需要重新拉去数据
        this.reload();
      } else if (this.paging === 'local') {
        // 本地分页只要切换内容就行
        this.calcLocalList();
      }
    },
    isTableDef(item) {
      return item.type !== 'selection' && item.type !== 'index' && item.type !== 'expand' && !item.formatter;
    },
    calcLocalList() {
      const start = (this.pagination.curpage - 1) * this.pagination.pagesize;
      const end = start + this.pagination.pagesize;
      this.total = this.filterRst.length;
      this.rst = this.filterRst.slice(start, end);
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
    filterChange(v) {
      this.filterQuery = mapValues(v, arrToBin);
    },
    select(selection) {
      this.selection = selection;
    },
    reload(flag) {
      if (this.hasInit) {
        this.refresh(flag);
      }
    },
    refresh(reload) {
      if (this.cgi) {
        if (reload) {
          if (this.pagination.curpage === 1) {
            this.getData();
          } else {
            this.pagination.curpage = 1;
          }
        } else {
          this.getData();
        }
        this.hasInit = true;
      }
    },
    getData() {
      let query;
      if (this.paging === 'remote') {
        query = merge({}, this.query, this.filterQuery, this.convertPage());
      } else {
        query = { ...this.query };
      }
      if (!this.globalLoading) {
        this.loading = true;
      }
      this.$axios[this.method](this.cgi, query, this.globalLoading).then((data) => {
        if (this.paging === 'remote') {
          this.total = data.count;
          this.rst = data.list || [];
        } else if (this.paging === 'local') {
          let list;
          if (this.listField) {
            list = data[this.listField];
          } else {
            list = data;
          }
          this.rawRst = list;
          this.filterRst = list;
          this.calcLocalList();
        } else {
          if (this.listField) {
            this.rst = data[this.listField];
          } else {
            this.rst = data;
          }
        }
      })
        .finally(() => {
          if (!this.globalLoading) {
            this.loading = false;
          }
        });
    },
    convertPage() {
      const { curpage, pagesize } = this.pagination;
      return {
        start: (curpage - 1) * pagesize,
        limit: pagesize,
      };
    },
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
  },
};
</script>
<style lang="scss">
@import './style';
</style>
