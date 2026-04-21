<template>
  <tabs-page
    :tabs="tabs"
  />
</template>

<script>
import TabsPage from '../../component/tedge-components/tabs-page.vue';

export default {
  components: {
    TabsPage,
  },
  data() {
    const tabs = [{
      label: '门禁卡',
      component: () => import('./cards'),
    }, {
      label: '领卡人员',
      component: () => import('./card-holders'),
    }, {
      label: '权限组',
      component: () => import('./permission-groups'),
    }];

    return {
      tabs,
    };
  },
};
</script>

<style lang="scss" scoped>

</style>
