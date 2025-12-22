<template>
  <span :class="effect">
    请登录管理员身份后操作
    &nbsp;
    <el-button
      type="text"
      size="small"
      @click="loginStatusService.login()"
    >
      登录
    </el-button>
  </span>
</template>

<script>
export default {
  props: {
    effect: {
      type: String,
      default() {
        return 'light';
      },
    },
  },
  data() {
    return {
      loginStatusService: window.tnwebServices.loginStatusService,
    };
  },
};
</script>

<style lang="scss" scoped>
.dark /deep/ {
  .el-button {
    color: #31a8f9 !important;
  }
}
</style>
