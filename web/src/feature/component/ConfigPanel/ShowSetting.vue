<template>
  <el-popover
    v-model="showFilter"
    popper-class="list-popover filter-popover"
    placement="bottom-end"
    width="800"
    :offset="16"
  >
    <el-tabs
      v-model="activeName"
    >
      <el-tab-pane
        v-for="tab in tabsShow"
        :key="tab.name"
        :label="`自定义${tab.label}内容`"
        :name="tab.name"
      >
        <div>
          <div class="filter-popover-content">
            <template v-if="tab.name === 'table'">
              <el-row
                v-for="column in tableColumns"
                :key="column.table"
              >
                <template v-for="field in column.fields">
                  <el-col
                    v-if="!isKey(field.name, table) && field.showInTabelSetting && !field.notShowItem"
                    :key="field.label"
                    :span="6"
                  >
                    <template v-if="isCustomizeLocal">
                      <el-checkbox
                        v-model="field.show"
                        :disabled="field.fixed"
                      />
                    </template>
                    <template v-else>
                      <el-checkbox
                        v-model="field.showContent"
                        :disabled="field.fixed"
                      />
                    </template>
                    <span :title="field.label">{{ field.label }}</span>
                  </el-col>
                </template>
              </el-row>
            </template>
            <template v-else-if="tab.name === 'search'">
              <el-row
                v-for="column in searchColumns"
                :key="column.table"
              >
                <template v-for="field in column.fields">
                  <el-col
                    :key="field.label"
                    :span="6"
                  >
                    <el-checkbox
                      v-model="field.showIndex"
                      :disabled="field.fixed"
                    />
                    <span :title="field.label">{{ field.label }}</span>
                  </el-col>
                </template>
              </el-row>
            </template>
          </div>
          <div class="popover-footer">
            <el-button
              type="text"
              class="text-dark"
              @click="reset(tab.name)"
            >
              取消
            </el-button>
            <el-button
              type="text"
              class="text-dark"
              @click="apply(tab.name)"
            >
              应用
            </el-button>
            <el-button
              type="text"
              @click="save(tab.name)"
            >
              保存
            </el-button>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
    <i
      slot="reference"
      class="tn-icon-filter"
    />
  </el-popover>
</template>
<script>
import { cloneDeep } from 'lodash';
import { flatten, map as mapFp } from 'lodash/fp';
import configMixin from './mixin';

export default {
  mixins: [configMixin],
  props: {
    table: {
      type: String,
      required: true,
    },
    tableCol: {
      type: Array,
      default: () => ([]),
    },
    searchCol: {
      type: Array,
      default: () => ([]),
    },
    showSearchSetting: {
      type: Boolean,
      required: true,
    },
    showTableSetting: {
      type: Boolean,
      required: true,
    },
    isCustomizeLocal: {
      type: Boolean,
      required: false,
    },
    isSearchSettingMerge: {
      type: Boolean,
      required: false,
      default: false,
    },
    isTableSettingMerge: {
      type: Boolean,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      showFilter: false,
      tableColumns: void 0,
      searchColumns: void 0,
      activeName: 'table',
      tabs: [{
        label: '展示',
        name: 'table',
      }, {
        label: '查询',
        name: 'search',
      }],
    };
  },
  computed: {
    tabsShow() {
      if (!this.showSearchSetting) {
        return this.tabs.filter(v => v.name !== 'search');
      }
      return this.tabs;
    },
  },
  watch: {
    tableCol() {
      this.tableReset();
    },
    searchCol() {
      this.searchReset();
    },
  },
  methods: {
    apply(name) {
      if (name === 'table') {
        this.applyTableColumns(this.tableColumns);
      } else {
        this.applySearchColumns();
      }
    },
    applyTableColumns(data) {
      this.$emit('change', data);
      this.showFilter = false;
      if (!this.isCustomizeLocal) {
        this.searchReset();
      }
    },
    applySearchColumns() {
      this.$emit('changeFilter', this.searchColumns);
      this.tableReset();
      this.showFilter = false;
    },
    arrToObj(arr) {
      const obj = {};
      arr.forEach((v) => { obj[v.name] = v; });
      return obj;
    },
    saveTableColumns() {
      let data = void 0;
      if (this.isTableSettingMerge) { // 还原合并的数据
        const fields = cloneDeep(this.tableColumns) |> mapFp('fields') |> flatten |> this.arrToObj;
        data = this.tableCol.map(v1 => ({
          table: v1.table,
          fields: v1.fields.map(v2 => fields[v2.name]),
        }));
      } else {
        data = this.tableColumns;
      }
      this.$emit('save', data);
      this.applyTableColumns(data);
    },
    saveSearchColumns() {
      this.$emit('saveFilter', this.searchColumns);
      this.applySearchColumns();
    },
    save(name) {
      if (name === 'table') {
        this.saveTableColumns();
      } else {
        this.saveSearchColumns();
      }
    },
    showItem(field) {
      return this.isCustomizeLocal ? field.show : field.showContent;
    },
    tableReset() {
      if (this.isTableSettingMerge) {
        const tableColumns = cloneDeep(this.tableCol) |> mapFp('fields') |> flatten;
        this.tableColumns = [{
          table: this.table,
          fields: tableColumns,
        }];
      } else {
        this.tableColumns = cloneDeep(this.tableCol);
      }
    },
    searchReset() {
      this.searchColumns = cloneDeep(this.searchCol);
    },
    reset(name) {
      this.showFilter = false;
      if (name === 'table') {
        this.tableReset();
      } else {
        this.searchReset();
      }
    },
  },
};
</script>
<style lang="scss">
@import "~common/style/mixin";

.filter-popover {
  padding: 0 !important;
  margin: 0 !important;

  .el-popover__title {
    padding: $space-l;
    margin: 0;
  }

  .filter-popover-content {
    max-height: 600px;
    overflow-y: auto;
  }

  .el-row {
    padding: $space-xs 0;
    margin: 0;
    border-top: 1px solid $border-color;

    .el-col {
      padding: 0 $space-l;
      height: 40px;
      line-height: 40px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .el-checkbox {
    margin-right: $space-xs;
  }

  .popover-footer {
    padding: $space-m;
    text-align: right;
    border-top: 1px solid $border-color;
  }
}
</style>
