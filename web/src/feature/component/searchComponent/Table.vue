<template>
  <el-block inner>
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
        <span
          v-if="showExportBtn"
          class="right-btns"
        >
          <slot
            name="extraOperator"
            :row="selection"
          />
          <el-button
            type="text"
            :auth-right-code="codes.dc"
            @click="expBatch(selection)"
          >
            <i class="tn-icon-import" />
            <span>导出所选</span>
          </el-button>
        </span>
      </div>
      <el-table
        v-show="(curColumns || []).length > 0"
        ref="table"
        v-loading="loading"
        :cell-height="49"
        element-loading-text="正在获取数据"
        :data="list"
        style="width: 100%;"
        @selection-change="select"
      >
        <el-table-column
          v-if="$scopedSlots.expand"
          type="expand"
          fixed
        >
          <template slot-scope="props">
            <slot
              name="expand"
              :row="props.row"
            />
          </template>
        </el-table-column>
        <el-table-column
          v-if="!hideCheckBox"
          type="selection"
          width="64"
          fixed
        />
        <template
          v-for="(item,index) in curColumns"
        >
          <el-table-column
            v-if="item.show"
            :key="index"
            :column-key="item.name"
            :prop="item.name"
            v-bind="item"
            :show-overflow-tooltip="true"
            :fixed="item.fixDirection || false"
          >
            <template
              slot-scope="scope"
            >
              <slot
                :name="item.name"
                :row="scope.row"
              >
                <template v-if="item.jump">
                  <a
                    href="javascript:void(0)"
                    @click="jump(item.jump, item.jumpScript, scope.row)"
                  >
                    {{ scope.row[item.name] }}
                  </a>
                </template>
                <template v-else>
                  {{ scope.row[item.name] }}
                </template>
              </slot>
            </template>
            <template
              slot="header"
            >
              <span
                class="middle-wrap"
              >
                <span>{{ item.label || item.name }}</span>
              </span>
            </template>
          </el-table-column>
        </template>
        <el-table-column
          v-if="$scopedSlots.operateBtn"
          :width="actionsLabelWidth"
          :fixed="operateBtnFixed || operateCustomFixed"
        >
          <template
            slot="header"
          >
            <span
              class="middle-wrap"
            >
              <span>操作</span>
              <span v-if="customFixed">
                <i
                  class="pointer tn-icon-pin-inclined text-primary"
                  @click="toggleFixed(false)"
                />
                <i
                  class="pointer tn-icon-pin text-light"
                  @click="toggleFixed('right')"
                />
              </span>
            </span>
          </template>
          <template slot-scope="props">
            <slot
              name="operateBtn"
              :row="props.row"
            />
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
  </el-block>
</template>
<script>
import { every, reduce, throttle } from 'lodash';
import Cookies from 'js-cookie';

export default {
  inject: ['configCgi', 'tableConfig', 'codes', 'roles'],
  props: {
    columns: {
      type: Array,
      required: true,
    },
    query: {
      type: Object,
      default: () => ({}),
    },
    exportKey: {
      type: String,
      default: 'id',
    },
    hideCheckBox: {
      type: Boolean,
      required: true,
    },
    actionsLabelWidth: {
      type: Number,
      default: 120,
    },
    showExportBtn: {
      type: Boolean,
      required: true,
    },
    extraParams: {
      type: Object,
      default: () => ({}),
    },
    customFixed: {
      type: Boolean,
      default: false,
    },

  },
  data() {
    return {
      rst: [],
      loading: false,
      currentPage: 1,
      limit: 10,
      total: 0,
      selection: [],
      showSelectionHeader: false,
      operateCustomFixed: false,
      currentUser: Cookies.get('tnebula_username'),
    };
  },
  computed: {
    list() {
      return this.rst;
    },
    curColumns() {
      const rst = [...this.columns]
        .sort((item1, item2) => item2.fieldOrder - item1.fieldOrder)
        .map(column => ({ fixed: column.type === 'selection', ...column }));
      const isFixWidth = every(rst, column => !!column.width);
      if (isFixWidth) {
        const totalWidth = reduce(rst, (memo, column) => memo + column.width, 0);
        if (this.$refs.table) {
          if (this.$refs.table.$el.offsetWidth > totalWidth) {
            try {
              delete rst[rst.length - 1].width;
            } catch {}
          }
        }
      }
      return rst;
    },
    operateBtnFixed() { // 列表项的fixDirection为right时，操作列必须右侧固定
      return this.curColumns.filter(v => v.fixDirection === 'right').length > 0 ? 'right' : false;
    },
  },
  watch: {
    list() {
      this.$emit('change', this.list);
    },
    query() {
      this.refresh();
    },
    selection(v) {
      this.showSelectionHeader = !!v.length;
    },
  },
  created() {
    this.styleTag = document.createElement('style');
    this.styleTag.innerHTML = `.el-tooltip__popper {
      max-width: 500px;
    }`;
    // eslint-disable-next-line prefer-destructuring
    const head = document.getElementsByTagName('head')[0];
    head.append(this.styleTag);
    this.$once('hook:beforeDestroy', () => {
      if (!this.styleTag) return;
      this.styleTag.remove();
      this.styleTag = null;
    });
  },
  mounted() {
    this.operateCustomFixed = this.customFixed ? 'right' : false;
  },
  methods: {
    toggleFixed(type) {
      this.operateCustomFixed = type;
    },
    clearSelection() {
      this.$refs.table.clearSelection();
    },
    jump(jumpObj, jumpScript, row) {
      let url = '';
      // 后端配置项无法满足需求，增加了自定义脚本字段
      if (jumpScript.length) {
        const { currentUser } = this;
        // eslint-disable-next-line no-eval
        const fn = eval(`(${jumpScript})`);
        url = fn(row, currentUser);
      } else {
        // 老的逻辑
        const data = JSON.parse(jumpObj);
        if (data.length === 1) { // 单一跳转
          const [{ path, query }] = data;
          const arr = [];
          Object.entries(query).forEach(([k, v]) => {
            arr.push([k, row[v]]);
          });
          url = `${path}?${arr.map(v => v.join('=')).join('&')}`;
        } else { // 多类跳转
          for (let i = 0; i < data.length; i++) {
            const { key, value, path, query } = data[i];
            if (!row[key].includes(value)) continue;
            if (Object.keys(query).length) {
              const arr = [];
              Object.entries(query).forEach(([k, v]) => {
                arr.push([k, row[v]]);
              });
              url = `${path}?${arr.map(v => v.join('=')).join('&')}`;
              break;
            }
          }
        }
      }
      window.open(url);
    },
    select(selection) {
      this.selection = selection;
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
        ...this.query,
        ...v,
        ...this.extraParams,
      };
      if (params.conditions && params.conditions.default) {
        params.conditions.default = params.conditions.default.filter(v => v.field !== 'id');
      }
      delete params.currentPage;
      params.start = (v.currentPage - 1) * params.limit;
      this.getData(params);
    },
    getData: throttle(function (params) {
      this.$axios.post(this.configCgi.queryCgi, params).then((data) => {
        this.total = data.count;
        this.rst = data.list;
      })
        .catch(({ message }) => {
          this.$message.error(message);
          this.total = 0;
          this.rst = [];
        });
    }, 1000, { leading: true, trailing: true }),
    expBatch(selection) {
      this.$emit('export', selection.map(item => item[this.exportKey]));
    },
  },
};
</script>
<style lang="scss">
@import './Table';
</style>
