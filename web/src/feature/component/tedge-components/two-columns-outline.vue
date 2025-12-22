<template>
  <div
    :class="{
      resizable: resizable,
      'left-drawer-mode': drawer,
    }"
    class="two-columns-outline"
    @mouseup="moveEnd"
    @mousemove="move"
  >
    <div
      ref="leftPanel"
      :class="{
        collapsed: isCollapsed,
        drawer: drawer,
        inner: !drawer,
      }"
      class="left-panel"
    >
      <div
        class="left-panel-content"
      >
        <slot
          name="left"
        />
      </div>

      <div
        class="control-bar"
        @mousedown="moveStart"
      >
        <div
          class="el-block-collapse"
          @mousedown.capture.stop
          @click="toggleCollapse"
        >
          <i :class="`tn-icon-arrow-${isCollapsed ? 'right' : 'left'}`" />
        </div>
      </div>
    </div>

    <div class="right-panel">
      <slot name="right" />
    </div>
  </div>
</template>

<script>
export default {
  props: {
    resizable: {
      type: Boolean,
      default() {
        return true;
      },
    },
    drawer: {
      type: Boolean,
      default() {
        return true;
      },
    },
  },
  data() {
    return {
      leftWidth: null,
      isMoving: false,
      isCollapsed: this.drawer,
    };
  },
  computed: {
    leftStyle() {
      const {
        leftWidth,
        drawer,
      } = this;

      const width = leftWidth === null
        ? null
        : `${leftWidth}px`;

      return {
        width,
      };
    },
  },
  watch: {
    drawer() {
      this.isCollapsed = this.drawer;
    },
    isCollapsed() {
      this.$emit('collapsedChange', this.isCollapsed);
    },
  },
  methods: {
    toggleCollapse() {
      this.isCollapsed = !this.isCollapsed;
      this.moveEnd();
    },
    moveStart() {
      if (!this.resizable) return;
      this.isMoving = true;
    },
    moveEnd() {
      this.isMoving = false;
    },
    move(evt) {
      if (!this.isMoving || this.isCollapsed) return;

      const {
        clientX,
      } = evt;

      const {
        leftPanel,
      } = this.$refs;

      const [leftRect] = leftPanel.getClientRects();

      if (!leftRect) return;

      const leftWidth = Number(clientX - leftRect.left);

      if (_.isNaN(leftWidth)) return;

      this.leftWidth = leftWidth;
    },
  },
};
</script>

<style lang="scss" scoped>
.two-columns-outline {
  display: flex;
  background-color: #ffffff;
}

.left-drawer-mode {
  .left-panel {
    position: absolute;
    left: 0;
    top: 0;
    height: calc(100% - 4px);
    background-color: #FFFFFF;
    z-index: 2;

    border: 1px solid #ddd;
  }
}

.control-bar {
  position: absolute;
  top: 0;
  right: -2px;

  width: 1px;
  height: 100%;
  background-color: #ededed;
  z-index: 99;

  &.resizable {
    cursor: move;
  }
}

.left-panel {
  position: relative;
  flex: 0;

  transition: translate 0.2s;
  transform-origin: left;

  &.collapsed {
    &.drawer {
      translate: -100%;
    }

    &.inner {
      width: 0;
    }
  }
}

.left-panel-content {
  height: 100%;
  overflow: hidden;
}

.right-panel {
  flex: 1;
  overflow: auto;
}

.el-block-collapse {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  -webkit-box-align: center;
  align-items: center;
  -webkit-box-pack: center;
  justify-content: center;
  width: 16px;
  height: 112px;
  user-select: none;
  background: #f1f2f5;
  z-index: 999;
  border-radius: 0px 8px 8px 0px;
  cursor: pointer;

  box-shadow: 2px 0 3px -1px #e0e0e0;
}
</style>
