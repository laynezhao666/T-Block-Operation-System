<template>
  <div>
    <el-steps
      :active="active"
      finish-status="success"
    >
      <el-step
        v-for="(node,index) in stepNames"
        :key="index"
        :title="node.item"
      />
    </el-steps>
  </div>
</template>

<script>
export default {
  props: {
    stepNames: {
      type: Array,
      required: true,
    },
    active: {
      type: Number,
      default: 0,
    },
  },
  data() {
    return {};
  },
};
</script>
<style lang="scss" scoped>
.el-steps--simple {
  background: #fff;
}
</style>
