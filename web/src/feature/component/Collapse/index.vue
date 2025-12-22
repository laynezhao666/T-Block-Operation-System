<template>
  <div>
    <div
      v-if="finalState==='open'"
      :class="`foldicon theme-${themeMode}`"
      @click="handleClick('close')"
    >
      <i class="tn-icon-arrow-left" />
    </div>
    <div
      v-if="finalState==='close'"
      :class="`foldicon theme-${themeMode}`"
      @click="handleClick('open')"
    >
      <i class="tn-icon-arrow-right" />
    </div>
  </div>
</template>
<script>
export default {
  props: {
    themeMode: {
      type: String,
      default: 'raw',
    },
    controlledValue: {
      type: String,
      default() {
        return null;
      },
    },
  },
  data() {
    return {
      state: 'close',
    };
  },
  computed: {
    finalState() {
      return this.controlledValue || this.state;
    },
  },
  methods: {
    handleClick(state) {
      this.$emit('handleClick', state);
      this.state = state;
    },
  },
};
</script>
<style scoped>
  .foldicon{
    /* box-shadow: 0 0 10px rgb(0 0 0 / 20%); */
    background-color: #FFFFFF;
    left:0;
    top: 1px;
    width: 28px;
    height: 46px;
    z-index: 1;
    display: flex;
    position: absolute;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: transform .5s, opacity .5s, visibility .5s;
  }
.theme-raw{
  background-color: white;
}
.theme-dark{
  background-color: #00123d;
  color: white;
}
</style>
