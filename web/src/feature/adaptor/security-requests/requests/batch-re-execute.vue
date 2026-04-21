<template>
  <el-button
    type="text"
    icon="tn-icon-refresh"
    @click="batchReExecute()"
  >
    重新执行
  </el-button>
  
</template>



<script>

export default {
  props: {
    tableContext: {
      type: Object,
      required: true,
    },
    selection: {
      type: Object,
      required: true,
    },
    getSelectedRows: {
      type: Function,
      required: true,
    },
  },


  methods: {
    async batchReExecute() {
      const { getSelectedRows } = this;
      const selectedRows = getSelectedRows();
      const idList = selectedRows.map(row => row.id);

      try {
        await this.$axios.post('/api/dcos/tdac-cgi/requests/batch-re-execute', {
          ids: idList
        });
        this.$message.success(`已触发 ${idList.length} 条记录重新执行`);
      } catch (error) {
        this.$message.error('操作失败: ' + error.message);
      }
    }
  }
};
</script>
