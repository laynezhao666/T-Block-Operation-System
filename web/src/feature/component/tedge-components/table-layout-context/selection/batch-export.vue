<template functional>
  <el-button
    type="text"
    icon="tn-icon-import"
    @click="$options.doExport(props)"
  >
    导出所选
  </el-button>
</template>

<script>
import { exportElTable } from '@/utils/xlsx.ts';

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
  doExport(props) {
    exportElTable(props.tableContext.getTableRef(), {
      excludeColumnTitles: ['操作'],
    });
  },
};
</script>
