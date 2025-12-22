<template>
  <div
    slot="actions"
    class="toolbar"
  >
    <div
      class="toolbar-unit"
      :class="{dark: themeMode === 'dark'}"
    >
      <div
        class="reset"
        @click="fit"
      >
        <i
          class="tn-icon-scan"
          :style="{color: themeMode === 'raw' ? '' : '#1470cc'}"
        />
      </div>
    </div>
    <!-- <div
      v-if="themeMode === 'raw'"
      class="toolbar-unit"
    >
      <div
        :class="{ active: isShowTemp }"
        @click="toggleTempMode"
      >
        <i class="tn-icon-cold-temp" />
      </div>
    </div> -->
  </div>
</template>
<script>

export default {
  props: {
    themeMode: {
      type: String,
      default: 'raw',
    },
    isShowTemp: {
      type: Boolean,
      default: true,
    },
  },
  methods: {
    fit() {
      this.$emit('fit');
    },
    // toggleTempMode() {
    //   this.$emit('toggleTempMode');
    // },
    // fullscreen() {
    //   this.$emit('fullscreen');
    // },
    // zoomChange(v) {
    //   this.$emit('zoomChange', v);
    // },
  },
};
</script>
<style lang="scss" scoped>
@import "~feature/style/reference";
.toolbar {
  position: absolute;
  right: 0;
  // top: 52px;
  z-index: 999;
  margin-top: 12px;

  .toolbar-unit {
    width: 48px;
    text-align: center;
    background-color: white;
    box-shadow: $box-shadow;
    border: 1px solid $border-color;
    &.dark {
      background-color: #001529;
      box-shadow: none;
      border: 1px solid #1470CC;
    }
    &:hover {
      border: 1px solid #1470CC;
    }
    + .toolbar-unit {
      margin-top: 12px;
    }

    &:nth-child(2) {
      > div {
        &.active {
          border: 1px solid $primary;
          color: $primary !important;
        }
      }
    }

    > div {
      height: 48px;
      line-height: 48px;
      color: #999;
      cursor: pointer;
      user-select: none;

      + div {
        border-top: 1px solid $border-color;
        &.dark {
          border-top: 1px solid #1470CC;
        }
      }

      i {
        position: relative;
        top: 6px;
        font-size: 30px;
      }
    }
  }
}
</style>
