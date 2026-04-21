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
      label: '通行时段',
      component: () => import('./time-period/index.vue'),
    }];

    return {
      tabs,
    };
  },
};
</script>

<style lang="scss" scoped>

</style>
