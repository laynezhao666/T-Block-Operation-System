<template functional>
  <el-tooltip
    :content="props.text"
    :placement="props.placement || 'bottom'"
    :effect="props.effect"
  >
    <div class="text-container">
      {{ props.text }}
    </div>
  </el-tooltip>
</template>

<script>
export default {
  props: {
    text: {
      type: String,
      default() {
        return null;
      },
    },
  },
};
</script>

<style lang="scss" scoped>
.text-container {
  text-overflow: ellipsis;
  overflow: hidden;
  width: 100%;
  white-space: nowrap;
}
</style>
