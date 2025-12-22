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
      :row-style="rowClass"
      style="width: 100%;"
      @selection-change="select"
      @sort-change="sortChange"
    >
      <el-table-column
        v-if="!tableConfig.showTableSelect"
        type="selection"
        width="64"
        fixed
      />
      <el-table-column
        v-for="item in curColumns"
        :key="item[columnKey]"
        :column-key="item[columnKey]"
        :prop="item[columnKey]"
        v-bind="item"
        :show-overflow-tooltip="true"
        :formatter="item.formatter"
      >
        <template
          v-if="isTableDef(item)"
          slot-scope="scope"
        >
          <slot
            :name="item[columnKey]"
            :data="scope"
          >
            <template v-if="item.formatter">
              {{ item.formatter(scope.row, item, scope.row[item[columnKey]]) }}
            </template>
            <template v-else-if="item.jump">
              <a
                href="javascript:void(0)"
                @click="item.jump(item, scope.row, table)"
              >
                {{ scope.row[item[columnKey]] }}
              </a>
            </template>
            <template v-else-if="item.operationUrl && item.operationMap">
              <span
                v-for="(i,index) in item.operationUrl"
                :key="index"
              >
                <a
                  v-if="!i.cloudFlag"
                  href="javascript:void(0)"
                  style="cursor:pointer"
                  :auth-right-code="i.authCode"
                  @click="solveOperation(i.operation,scope.row)"
                >
                  <span style="margin-right:5px">
                    {{ i.operation }}
                  </span>
                </a>
              </span>
            </template>
            <template v-else-if="item.switch">
              <el-switch
                v-model="scope.row[item[columnKey]]"
                @change="solveOperation('switch',scope.row)"
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
            <template v-else-if="item.customStyle">
              <span
                v-if="item.modifyMap && item.modifyMap[scope.row[item[columnKey]]]"
                :style="item.customStyle"
              >
                {{ item.modifyMap[scope.row[item[columnKey]]] }}</span>
              <span v-else>{{ scope.row[item[columnKey]] }}</span>
            </template>
            <template v-else-if="(item.fieldEnum || item.modifyMap) && item.modifyName">
              <span v-if="!scope.row[item[columnKey]]">--</span>
              <span
                v-if="scope.row[item[columnKey]]"
                :style="item.getColumnStyle ? item.getColumnStyle(scope): {}"
              >
                {{ scope.row[item[columnKey]] &&
                  (item.fieldEnum || item.modifyMap).find(v => v.value === scope.row[item[columnKey]]).label }}
              </span>
            </template>
            <template v-else-if="item.modal">
              <el-button
                type="text"
                @click="solveOperation(item.modal,scope.row)"
              >
                {{ scope.row[item[columnKey]] }}
              </el-button>
            </template>
            <template v-else>
              <span :style="item.getColumnStyle ? item.getColumnStyle(scope): {}">
                {{ scope.row[item[columnKey]] }}
              </span>
            </template>
          </slot>
        </template>
        <template
          slot="header"
        >
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
import { find, every, reduce } from 'lodash';
import { eventBus } from '../script/eventBus';
import getEdgeRequest from '../../../../utils/request';
import moment from 'moment';

const timeBeforeOneMonth = moment().add('year', 0)
  .month(moment().month() - 1)
  .format('YYYY-MM-DD HH:mm:ss');
const currentTime = moment(Date.now()).format('YYYY-MM-DD HH:mm:ss');

export default {
  inject: ['configCgi', 'tableConfig'],
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
    play: {
      type: Boolean,
      default: false,
    },
    query: {
      type: Object,
      default: () => ({}),
    },
    showTableSelect: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    const curColumns = this.calcCurColumns();
    return {
      rawRst: [],
      rst: [],
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
      hasUrlParam: false,
      rowClass: this.tableConfig?.rowClass || {},
    };
  },
  computed: {
    list() {
      return this.play ? this.localData : this.rst;
    },
  },
  watch: {
    query() {
      this.currentPage = 1;
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
    solveOperation(data, item) {
      eventBus.$emit('showModal', {
        type: data,
        data: item,
      });
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
    refresh(axiosLoading) {
      this.filterHandler({ limit: this.limit, currentPage: this.currentPage }, axiosLoading);
    },
    filterHandler(v = {}, axiosLoading) {
      if (!v.currentPage) {
        // eslint-disable-next-line no-param-reassign
        v.currentPage = 1;
        this.currentPage = 1;
      }
      const params = {
        limit: this.limit,
        sortKey: this.sortKey,
        sortOrder: this.sortOrder,
        ...this.tableConfig.defaultParams,
        ...this.query,
        ...v,
      };

      params[this.tableConfig.paginationParam || 'offset'] = (v.currentPage - 1) * params.limit;

      if (!this.tableConfig.hasCurrentPage && params.currentPage) {
        delete params.currentPage;
      }
      if (this.tableConfig.inputSearchKey && this.tableConfig.inputSearchKey !== 'keyword') {
        delete params.keyword;
      }

      Object.keys(params).forEach((item) => {
        if (params[item] === 'true') {
          params[item] = true;
        }
        if (params[item] === 'false') {
          params[item] = false;
        }
      });
      // 查询持续时间需要带上默认起止时间
      // if (this.query.duration) {
      //   params.occurTimeStart = timeBeforeOneMonth;
      //   params.occurTimeEnd = currentTime;
      // }
      if (this.hasUrlParam) {
        if (params.occurTimeStart) {
          this.getData({ ...params, ...this.tableConfig.searchParams }, axiosLoading);
        } else {
          let extraParams = { };
          if (this.tableConfig.occurTime) {
            extraParams = { occurTimeStart: timeBeforeOneMonth, occurTimeEnd: currentTime };
          }
          this.getData({ ...params, ...extraParams, ...this.tableConfig.searchParams }, axiosLoading);
        }
      } else {
        this.getData(params, axiosLoading);
      }
    },
    getData(params, axiosLoading = true) {
      if (!this.play) {
        if (this.tableConfig.useNoEdge) {
          this.$axios.post(this.configCgi.queryCgi, params, axiosLoading)
            .then((data) => {
              this.total = data.count;
              eventBus.$emit('outputData', {
                data: this.total,
              });
              this.rst = data.list;
            // this.tableConfig.searchParams = {};
            // this.hasUrlParam = false;
            });
        } else {
          getEdgeRequest(this.$axios, params.mozuId).post(this.configCgi.queryCgi, params, axiosLoading)
            .then((data) => {
              this.total = data.count;
              eventBus.$emit('outputData', {
                data: this.total,
              });
              this.rst = data.list;
            // this.tableConfig.searchParams = {};
            // this.hasUrlParam = false;
            });
        }
      } else {
        console.log('演示模式');
      }
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
    sortChange({ prop, order }) {
      this.sortKey = prop;
      // eslint-disable-next-line no-nested-ternary
      this.sortOrder = prop ? (order === 'descending' ? 'desc' : 'asc') : null;
      this.filterHandler();
    },
  },
};
</script>
<style lang="scss">
@import './style';
.el-table--striped
.el-table__body tr.el-table__row--striped.current-row td,
.el-table__body tr.current-row>td,
.el-table__body tr.hover-row.current-row>td,
.el-table__body tr.hover-row.el-table__row--striped.current-row>td,
.el-table__body tr.hover-row.el-table__row--striped>td,
.el-table__body tr.hover-row>td {
    background-color: #fff;
  }
</style>
