<template functional>
  <el-table-column
    :label="props.tableContext.radioRowSelection.title"
    fixed="left"
    width="80"
  >
    <template
      #default="{ row }"
    >
      <el-radio
        :value="$options.getRowIsChecked(props.tableContext, row)"
        :label="true"
        class="table-radio"
        @change="$options.handleChange(props.tableContext, row)"
      />
    </template>
  </el-table-column>
</template>

<script>
export default {
  props: {
    tableContext: {
      type: Object,
      required: true,
    },
  },
  getRadioRowSelection(tableContext) {
    return tableContext.radioRowSelection;
  },
  getRowIsChecked(tableContext, row) {
    const {
      identify,
      value,
    } = this.getRadioRowSelection(tableContext);

    return identify(row) === value;
  },
  handleChange(tableContext, row) {
    const radioRowSelection = this.getRadioRowSelection(tableContext);
    const {
      identify,
      onChange,
    } = radioRowSelection;

    radioRowSelection.value = identify(row);
    onChange(radioRowSelection.value, row);
  },
};
</script>

<style lang="scss" scoped>
.table-radio /deep/ {
  .el-radio__label {
    display: none;
  }
}
</style>
