<template>
  <div>
    <el-input-number
      v-model="inputValue[0]"
      :min="min"
      :max="inputValue[1]"
      :step="step"
      size="mini"
    />
    -
    <el-input-number
      v-model="inputValue[1]"
      :min="inputValue[0]"
      :max="max"
      :step="step"
      size="mini"
    />
  </div>
</template>
<script>
/**
 * 计数范围组件
 * @param {Array} value
 * @param {Number} min 非必须，计数器的最小值，默认为0
 * @param {Number} min 非必须，计数器的最大值，默认为9999
 * @param {Number} step 非必须，计数器的步长，默认为1
 */
export default {
  props: {
    value: {
      type: Array,
      required: true,
      default: () => ([]),
    },
    min: {
      type: Number,
      default: 0,
    },
    max: {
      type: Number,
      default: 9999,
    },
    step: {
      type: Number,
      default: 1,
    },
  },
  data() {
    return {
      inputValue: [],
    };
  },
  watch: {
    inputValue(v) {
      this.$emit('input', v);
    },
    value(v) {
      this.inputValue = v;
    },
  },
};
</script>
