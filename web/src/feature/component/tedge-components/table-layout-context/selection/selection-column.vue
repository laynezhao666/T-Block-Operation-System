<template functional>
  <el-table-column
    fixed="left"
    type="selection"
    reserve-selection
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
};
</script>
