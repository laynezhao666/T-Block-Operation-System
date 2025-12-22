<template>
  <el-block
    no-padding
    class="tedge-table-layout"
  >
    <el-table-toolbar
      v-if="!context.hideToolbar"
      v-model.trim="tableSearch.value"
      :filter-placeholder="tableSearch.placeholder || '输入关键字'"
      :hide-search="tableSearch.isHide"
      :actions="context.toolbarActions"
      @search="tableSearch.doSearch(context)"
    >
      <template
        slot="extra"
      >
        <div class="extra-items">
          <component
            :is="extra"
            v-for="(extra, i) in context.extras"
            :key="i"
            :table-context="context"
          />
        </div>

        <slot
          name="toolbar-extra"
          v-bind="context"
        />
      </template>
    </el-table-toolbar>

    <component
      :is="bar"
      v-for="(bar, i) in context.topBars"
      :key="`top-bars-${i}`"
      :table-context="context"
    />

    <el-table
      ref="table"
      :data="context.tableData"
      style="width: 100%"
      v-bind="context.tableProps"
      v-on="tableListeners"
    >
      <component
        :is="col"
        v-for="(col, i) in context.prefixColumns"
        :key="i"
        :table-context="context"
      />

      <el-table-column
        v-if="context.indexColumn"
        :label="context.indexColumn.label"
        :width="context.indexColumn.width || 80"
        :fixed="context.indexColumn.fixed || 'left'"
        type="index"
      >
        <template #default="{ $index }">
          {{ computeIndex($index) }}
        </template>
      </el-table-column>

      <slot name="columns" />

      <el-table-column
        v-if="context.oprsColumnOprs.length"
        :width="getByPath(context, 'curd.rowEditColumnWidth') || 120"
        label="操作"
        fixed="right"
      >
        <template #default="{ row, $index }">
          <template
            v-for="(opr, i) in context.oprsColumnOprs"
          >
            <span
              v-if="i !== 0"
              :key="`split-${i}`"
              class="oprs-split"
            >
              |
            </span>

            <component
              :is="opr"
              :key="i"
              :row="row"
              :index="$index"
              :table-context="context"
            />
          </template>
        </template>
      </el-table-column>
    </el-table>

    <component
      :is="bar"
      v-for="(bar, i) in context.footerBars"
      :key="i"
      :table-context="context"
    />

    <component
      :is="m"
      v-for="(m, i) in context.modals"
      :key="i"
      :table-context="context"
    />

    <slot name="outer-modals" />
  </el-block>
</template>

<script>
import _ from 'lodash';

export default {
  provide() {
    return {
      context: this.context,
    };
  },
  props: {
    context: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {};
  },
  computed: {
    tableSearch() {
      return this.context.search || {
        isHide: true,
      };
    },
    tableListeners() {
      return _.mapValues(this.context.tableListeners, handler => (...args) => handler(this.context, ...args));
    },
  },
  watch: {
    context: {
      immediate: true,
      handler() {
        this.context.getTableRef = () => this.$refs.table;
      },
    },
  },
  created() {
    this.loadData();
    if (this.context.watches) {
      this.context.watches.forEach((w) => {
        const expOrFn = typeof w.expOrFn === 'string'
          ? `context.${w.expOrFn}`
          : w.expOrFn;
        this.$watch(expOrFn, (...args) => w.callback.call(this, this.context, ...args), w.options);
      });
    }
  },
  methods: {
    getByPath: _.get,
    loadData() {
      this.context.loadData();
    },
    computeIndex(index) {
      const {
        pagination,
      } = this.context;

      if (!pagination) return index + 1;

      return index + 1 + (pagination.size * (pagination.current - 1));
    },
  },
};
</script>

<style lang="scss" scoped>
.tedge-table-layout {
  position: relative;
}

.oprs-split {
  color: var(--tn-color-primary);
}

.extra-items {
  /deep/ {
    & > * {
      margin-right: 8px;
    }
  }
}
</style>
