<template>
  <div class="defaultPage">
    {{ title }}
  </div>
</template>
<script>
export default {
  props: {
    title: {
      type: String,
      default: '暂无内容',
    },
  },
};
</script>
<style lang="scss" scoped>
.defaultPage {
  width: 100%;
  height: 50vh;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color:#999;
  background: white;
}
</style>
