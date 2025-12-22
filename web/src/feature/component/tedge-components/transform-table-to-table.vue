<template>
  <div class="transform-table-to-table">
    <div class="table-panel">
      <header class="header">
        <span class="header-title">
          待选列表
        </span>

        <span class="header-check-container">
          <span class="count-total-info">
            {{ paddingTableContext.selection.selectedLength }} / {{ paddingTableContext.tableData.length }}项
          </span>
        </span>
      </header>

      <tedge-table-layout
        :context="paddingTableContext"
        class="table-layout"
      >
        <template #columns>
          <slot
            name="table-columns"
          />

          <slot
            name="padding-table"
          />
        </template>
      </tedge-table-layout>
    </div>

    <div class="transform-oprs">
      <el-button
        :disabled="!selectedTableContext.selection.selectedLength"
        type="primary"
        icon="tn-icon-arrow-left"
        size="small"
        @click="triggerRemove"
      />

      <el-button
        :disabled="!paddingTableContext.selection.selectedLength"
        type="primary"
        icon="tn-icon-arrow-right"
        size="small"
        @click="triggerAppend"
      />
    </div>

    <div class="table-panel">
      <header class="header">
        <span class="header-title">
          已选列表
        </span>

        <span class="header-check-container">
          <span class="count-total-info">
            {{ selectedTableContext.selection.selectedLength }} / {{ selectedTableContext.tableData.length }}项
          </span>
        </span>
      </header>

      <tedge-table-layout
        :context="selectedTableContext"
        class="table-layout"
      >
        <template #columns>
          <slot
            name="table-columns"
          />

          <slot
            name="padding-table"
          />
        </template>
      </tedge-table-layout>
    </div>
  </div>
</template>

<script>
import TedgeTableLayout from './tedge-table-layout.vue';
import {
  chainTableLayout,
  normalizeIdentity,
} from './table-layout-context/table-layout-context';

export default {
  components: {
    TedgeTableLayout,
  },
  props: {
    rowIdentity: {
      type: [String, Function],
      required: true,
    },
    allOptions: {
      type: Array,
      required: true,
    },
    selectedKeys: {
      type: Array,
      required: true,
    },
  },
  data() {
    window.tf = this;
    return {
      paddingTableContext: this.createTableContext('padding'),
      selectedTableContext: this.createTableContext('selected'),
    };
  },
  computed: {
    normalizedRowIdentity() {
      return normalizeIdentity(this.rowIdentity);
    },
    allOptionsMap() {
      return _.fromPairs(_.map(this.allOptions, opt => [this.normalizedRowIdentity(opt), opt]));
    },
  },
  watch: {
    selectedKeys: {
      deeo: true,
      handler() {
        this.reloadTables();
      },
    },
  },
  methods: {
    createTableContext(type) {
      return chainTableLayout(type === 'padding' ? this.getPaddingTableData : this.getSelectedTableData)
        .hideToolbar()
        .tableStyle({
          size: 'small',
          stripe: true,
          height: 300,
          style: 'width: 100%',
        })
        .indexColumn({
          width: 80,
        })
        .selection({
          identity: this.rowIdentity,
          hideToolbar: true,
        })
        .done();
    },
    getPaddingTableData() {
      const {
        allOptions,
        selectedKeys,
        normalizedRowIdentity,
      } = this;

      return _.filter(allOptions, item => !selectedKeys.includes(normalizedRowIdentity(item)));
    },
    getSelectedTableData() {
      const {
        selectedKeys,
        allOptionsMap,
      } = this;

      return _.chain(selectedKeys)
        .map(key => allOptionsMap[key])
        .filter(_.identity)
        .value();
    },
    triggerAppend() {
      const { paddingTableContext } = this;
      const {
        selection,
      } = paddingTableContext;

      const checkedRows = selection.getSelectedRows();
      const keys = _.map(checkedRows, this.normalizedRowIdentity);
      this.selectedKeys.push(...keys);

      selection.cancel(paddingTableContext);
    },
    triggerRemove() {
      const {
        selectedTableContext,
        selectedKeys: oldSelectedKeys,
      } = this;
      const {
        selection,
      } = selectedTableContext;

      const checkedRows = selection.getSelectedRows();
      const checkedKeysSet = new Set(_.map(checkedRows, this.normalizedRowIdentity));

      // 从后往前删，避免因为删除导致位置偏移
      for (let i = this.selectedKeys.length - 1; i >= 0; i--) {
        const key = this.selectedKeys[i];
        if (!checkedKeysSet.has(key)) continue;

        this.selectedKeys.splice(i, 1);
      }

      selection.cancel(selectedTableContext);
    },
    reloadTables() {
      this.reloadTable(this.paddingTableContext);
      this.reloadTable(this.selectedTableContext);
    },
    reloadTable(tableContext) {
      tableContext.loadData();
    },
    getSelectedOptions() {
      const {
        allOptionsMap,
        selectedKeys,
      } = this;

      return selectedKeys.map(key => allOptionsMap[key]);
    },
  },
};
</script>

<style lang="scss" scoped>
.transform-table-to-table {
  display: flex;
}

.table-panel {
  flex: 1;
  border: 1px solid #ebebeb;
  overflow: hidden;
}

.header {
  display: flex;
  align-items: center;
  line-height: 48px;
  margin: 0 16px;
}

.header-check-container {
  display: block;
  flex: 1;
  align-items: center;
  text-align: right;
}

.header-title {
  font-size: 14px;
}

.select-all {
  margin-right: 8px;
}

.count-total-info {
  font-size: 14px;
  color: #666;
  font-weight: 400;
}

.transform-oprs {
  padding: 0 24px;
  display: inline-block;
  vertical-align: middle;
  align-self: center;

  & > * {
    display: block;
    margin: 5px 0;
    padding: 4px;
    width: 34px;
    height: 34px;
  }
}

.table-layout {
  box-shadow: none;
}
</style>
