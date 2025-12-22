<template>
  <component
    :is="comp"
    v-loading="loading"
  >
    <slot
      v-if="!loading && inited"
      v-bind:value="value"
    />
  </component>
</template>
<script>
export default {
  model: {
    prop: '≈',
    event: 'change',
  },
  props: {
    name: {
      type: String,
      required: true,
    },
    comp: {
      type: [String, Object],
      default() {
        return 'div';
      },
    },
  },
  data() {
    return {
      value: null,
      loading: false,
      inited: false,
    };
  },
  watch: {
    name: {
      immediate: true,
      async handler() {
        if (!this.name) return;

        this.loading = true;
        this.value = await window.tnwebServices.customConfigService.loadConfig(this.name);
        this.loading = false;
        this.inited = true;
      },
    },
  },
};
</script>
<style lang="scss" scoped>
.resizable-divider {
  border: 1px solid #e0e0e0;
  cursor: move;
}
</style>
