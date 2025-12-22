<template>
  <el-tooltip
    effect="light"
    placement="top"
    :disabled="!!(disabled || loginStatusService.adminLogined)"
  >
    <slot
      v-if="loginStatusService.adminLogined || mode !== 'hide'"
      :loginStatusService="loginStatusService"
      :hasRight="loginStatusService.adminLogined"
    />

    <admin-limit-login-alert slot="content" />
  </el-tooltip>
</template>

<script>
import AdminLimitLoginAlert from './admin-limit-login-alert.vue';

export default {
  components: {
    AdminLimitLoginAlert,
  },
  props: {
    disabled: {
      type: Boolean,
      default() {
        return false;
      },
    },
    mode: {
      type: String,
      default() {
        return 'disabled';
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
