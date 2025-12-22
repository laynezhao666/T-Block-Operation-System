<template>
  <div>
    <p
      style="font-size:18px;margin-bottom:10px;font-weight:550"
    >
      {{ info.title }}
    </p>
    <el-row>
      <el-col
        v-for="item in info.data"
        :key="item.label"
        :span="12"
        style="display: flex; line-height: 40px"
      >
        <div style="width: 160px">
          {{ item.label }}
        </div>
        <div style="flex: 1">
          {{ item.value }}
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script>
export default {
  props: {
    info: {
      type: Object,
      default: () => [],
    },
    showtitle: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {

    };
  },
};
</script>

<style>

</style>
