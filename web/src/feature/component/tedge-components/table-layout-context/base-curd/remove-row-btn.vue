<template>
  <admin-limit-tooltips
    :disabled="!tableContext.curd.remove.adminRight"
  >
    <span slot-scope="{ hasRight }">
      <el-popconfirm
        :disabled="tableContext.curd.remove.confirm === false"
        :title="typeof tableContext.curd.remove.confirm === 'string'
          ? tableContext.curd.remove.confirm
          : '是否确认删除？'
        "
        @onConfirm="handleConfirm()"
      >
        <el-button
          slot="reference"
          type="text"
          :disabled="(tableContext.curd.remove.adminRight && !hasRight)
            || tableContext.curd.remove.disabled && tableContext.curd.remove.disabled()"
        >
          {{ tableContext.curd.remove.label || '删除' }}
        </el-button>
      </el-popconfirm>
    </span>
  </admin-limit-tooltips>
</template>

<script>
import AdminLimitTooltips from 'feature/component/tedge-components/admin-limit-tooltips.vue';

export default {
  components: {
    AdminLimitTooltips,
  },
  props: {
    tableContext: {
      type: Object,
      required: true,
    },
    row: {
      type: Object,
      required: true,
    },
    index: {
      type: Number,
      required: true,
    },
  },
  methods: {
    async handleConfirm() {
      const {
        tableContext,
        row,
        index,
      } = this;

      // tableContext.curd.remove.confirm是之前bug勿用了confirm字段做回调，但是由于多个地方使用，先适配
      const removeFunc = tableContext.curd.remove.remove || tableContext.curd.remove.confirm;
      const result = await removeFunc?.(row, index);
      if (result === false) return;

      tableContext.loadData();
    },
  },
};
</script>
