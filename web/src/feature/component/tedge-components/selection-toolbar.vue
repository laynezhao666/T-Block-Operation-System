<template>
  <div
    class="selection-header"
  >
    <i
      class="selection-close-btn tn-icon tn-icon-close"
      @click="cancel"
    />
    <span class="selection-text"> 已选择 {{ selectedList.length }} 项 </span>
    <span
      class="right-btns"
    >
      <slot />
    </span>
  </div>
</template>

<script>
export default {
  props: {
    selectedList: {
      type: Object,
      required: true,
    },
  },
  methods: {
    cancel() {
      this.$emit('cancel');
    },
  },
};
</script>

<style lang="scss" scoped>
.selection-header {
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
}
</style>
