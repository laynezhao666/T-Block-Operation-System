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
      label: '进出记录',
      component: () => import('./in-out'),
    }, {
      label: '门禁告警记录',
      component: () => import('./alarms'),
    }];

    return {
      tabs,
    };
  },
};
</script>

<style lang="scss" scoped>

</style>
