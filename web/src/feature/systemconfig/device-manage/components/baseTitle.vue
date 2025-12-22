
<template>
  <div class="title">
    <div class="label" />
    <el-title :title="title" />
  </div>
</template>

<script>
export default {
  props: {
    title: {
      type: String,
      default: '',
    },
  },
};
</script>

<style lang="scss" scoped>
.title {
  display: flex;
  align-items: center;
  .label {
    width: 10px;
    height: 16px;
    background: #1470cc;
    margin-right: 18px;
  }
}
</style>
