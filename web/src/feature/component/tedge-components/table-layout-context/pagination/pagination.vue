<template functional>
  <el-pagination
    v-if="props.tableContext.pagination"
    :total="props.tableContext.pagination.total"
    :current-page="props.tableContext.pagination.current"
    :page-size="props.tableContext.pagination.size"
    :page-sizes="props.tableContext.pagination.pageSizes || [10, 20, 30, 40, 50, 100]"
    styled
    layout="total, prev, pager, next, sizes, jumper"
    @size-change="$options.handleSizeChange(props, $event)"
    @current-change="$options.handleCurrentChange(props, $event)"
  />
</template>

<script>
export default {
  props: {
    tableContext: {
      type: Object,
      required: true,
    },
  },
  handleCurrentChange(props, newPage) {
    // eslint-disable-next-line no-param-reassign
    props.tableContext.pagination.current = newPage;
  },
  handleSizeChange(props, newSize) {
    // eslint-disable-next-line no-param-reassign
    props.tableContext.pagination.size = newSize;
  },
};
</script>
