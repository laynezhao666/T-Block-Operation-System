<template>
  <div
    :style="{ cursor: vertical ? 'row-resize' : 'col-resize' }"
    class="resizable-divider"
    @mousedown="handleMouseDown"
  />
</template>
<script>
export default {
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Number,
      required: true,
    },
    persistenceKey: {
      type: String,
      default() {
        return null;
      },
    },
    vertical: {
      type: Boolean,
      default() {
        return false;
      },
    },
  },
  data() {
    return {
      moving: false,
      startValue: null,
      startPos: null,
    };
  },
  watch: {
    persistenceKey: {
      immediate: true,
      handler() {
        if (this.persistenceKey) {
          this.restore();
        }
      },
    },
  },
  computed: {
    fullPersistenceKey() {
      if (!this.persistenceKey) return;
      return `resizable-divider[${this.persistenceKey}]`;
    },
  },
  mounted() {
    window.addEventListener('mouseup', this.handleMouseUp);
    window.addEventListener('mousemove', this.handleMouseMove);
  },
  beforeDestroy() {
    window.removeEventListener('mouseup', this.handleMouseUp);
    window.removeEventListener('mousemove', this.handleMouseMove);
  },
  methods: {
    restore() {
      const { fullPersistenceKey } = this;
      if (!fullPersistenceKey) return;

      const storedValue = localStorage.getItem(fullPersistenceKey);
      if (!storedValue) return;

      this.$emit('change', Number(storedValue) || this.value);
    },
    handleMouseDown(evt) {
      this.moving = true;
      this.startValue = this.value;
      this.startPos = this.vertical ? evt.clientY : evt.clientX;
    },
    handleMouseUp() {
      this.moving = false;
    },
    handleMouseMove(evt) {
      if (!this.moving) return;

      evt.preventDefault();

      const newPos = this.vertical ? evt.clientY : evt.clientX;
      const diffPos = newPos - this.startPos;
      const newValue = this.startValue + diffPos;

      this.$emit('change', newValue);

      if (this.fullPersistenceKey) {
        localStorage.setItem(this.fullPersistenceKey, newValue);
      }
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
