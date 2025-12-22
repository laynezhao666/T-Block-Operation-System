<template functional>
  <div
    v-if="props.tableContext.selection.selectedLength && !props.tableContext.selection.hideToolbar"
    class="selection-header"
  >
    <i
      class="selection-close-btn tn-icon tn-icon-close"
      @click="props.tableContext.selection.cancel(props.tableContext)"
    />
    <span class="selection-text"> 已选择 {{ props.tableContext.selection.selectedLength }} 项 </span>
    <span
      class="right-btns"
    >
      <component
        :is="opr"
        v-for="(opr, i) in props.tableContext.selection.oprs"
        :key="i"
        :table-context="props.tableContext"
        :selection="props.tableContext.selection"
        :get-selected-rows="() => props.tableContext.selection.getSelectedRows()"
      />
    </span>
  </div>
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

<style lang="scss" scoped>
.selection-header {
  position: absolute;
  top: 0;
  z-index: 90;
  height: 64px;
  width: 100%;
  background-color: rgb(255, 255, 255);
  box-shadow: rgba(203, 203, 203, 0.5) 0px 3px 5px 0px;

  /deep/ {
    span {
      vertical-align: middle;
    }
  }
}

.selection-close-btn {
  padding: 20px 16px;
  cursor: pointer;
  vertical-align: middle;
}

.selection-text {
  font-size: 20px;
}

.right-btns {
  float: right;
  padding: 19px 0px;
  margin-right: 24px;

  & > * {
    margin-right: 16px;
  }
}
</style>
