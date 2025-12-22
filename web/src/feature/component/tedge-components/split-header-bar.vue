<template functional>
  <div
    :class="{
      'no-padding': props.noPadding,
      'split-header-bar': true,
      [data.staticClass || '']: true,
    }"
    v-bind="data.attrs"
  >
    <div class="split-title">
      {{ props.title }}
    </div>

    <div class="oprs">
      <slot />
    </div>
  </div>
</template>

<script>
export default {
  props: {
    title: {
      type: String,
      required: true,
    },
    noPadding: {
      type: Boolean,
      default() {
        return false;
      },
    },
  },
};
</script>

<style lang="scss" scoped>
.split-header-bar {
  display: flex;
  align-items: center;

  &:not(.no-padding) {
    padding: 16px 48px;
  }
}

.split-title {
  font-weight: 600;

  &::before {
    content: '';
    display: inline-block;
    width: 8px;
    height: 24px;
    background-color: var(--tn-color-primary);
    vertical-align: middle;
    margin-right: 8px;
  }
}

.oprs {
  flex: 1;
  text-align: right;
}

.header-like-form-item {
  /deep/ {
    .el-form-item__label {
      padding-right: 0;
      width: 100%;

      &:before {
        display: none !important;
      }
    }

    .el-form-item__error {
      position: relative;
    }
  }
}
</style>
