<template>
  <el-form
    class="filters-form"
    size="small"
    label-width="8em"
  >
    <el-row :gutter="16">
      <el-col :span="8">
        <el-form-item
          label="人员组"
        >
          <el-input
            v-model="filters.company"
            placeholder="请输入人员组"
            clearable
          />
        </el-form-item>
      </el-col>
    </el-row>
  </el-form>
</template>

<script>
export default {
  props: {
    tableContext: {
      type: Object,
      required: true,
    },
  },
  computed: {
    filters() {
      return this.tableContext.filters;
    },
  },
};
</script>

<style lang="scss" scoped>
.filters-form {
  border-top: 1px solid #efefef;
}
</style>
