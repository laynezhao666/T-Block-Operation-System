<template>
  <component :is="containerComp">
    <slot />

    <!-- <el-page-holder
      v-else
      type="no-permission"
      subtitle="您还没有访问的权限，请先登录管理员身份"
      class="page-holder"
      :button-list="[
        { text: '登录', type: 'primary', onclick: login },
      ]"
    /> -->
  </component>
</template>

<script>
export default {
  props: {
    containerComp: {
      type: [String, Object],
      default() {
        return 'div';
      },
    },
  },
  data() {
    return {
      loginStatusService: window.tnwebServices.loginStatusService,
    };
  },
  computed: {
    adminLogined() {
      return this.loginStatusService.adminLogined;
    },
  },
  watch: {
    adminLogined(v) {
      this.$emit('adminLoginedChange', v);
    },
  },
  methods: {
    login() {
      this.loginStatusService.login();
    },
  },
};
</script>

<style lang="scss" scoped>
.page-holder {
  margin: auto;
}
</style>
