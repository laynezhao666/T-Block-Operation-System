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
      label: '异步消息',
      component: () => import('./requests'),
    }];

    return {
      tabs,
    };
  },
};
</script>

<style lang="scss" scoped>

</style>
