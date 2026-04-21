<template>
  <el-form
    class="filters-form"
    size="small"
    label-width="8em"
  >
    <el-row :gutter="16">
      <el-col :span="8">
        <el-form-item
          label="卡类型"
        >
          <el-select
            v-model="filters.card_type"
            placeholder="请选择卡类型"
            clearable
            :fix-width="true"
          >
            <el-option
              :value="0"
              label="长期卡"
            />
            <el-option
              :value="1"
              label="临时卡"
            />
          </el-select>
        </el-form-item>
      </el-col>

      <el-col :span="8">
        <el-form-item
          label="卡状态"
        >
          <el-select
            v-model="filters.card_flag"
            placeholder="请选择卡状态"
            clearable
            :fix-width="true"
          >
            <el-option
              v-for="(item, i) in validOptions"
              :key="i"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>

      <el-col :span="8">
        <el-form-item
          label="权限组"
        >
          <el-select
            v-model="filters.access_group"
            placeholder="请选择权限组"
            clearable
            :fix-width="true"
          >
            <el-option
              v-for="(item, i) in allGroups"
              :key="i"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
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
  data() {
    return {
      validOptions: [{
        value: 0,
        label: '正常',
      }, {
        value: 1,
        label: '停用',
      }],
      allGroups: [],
    };
  },
  computed: {
    filters() {
      return this.tableContext.filters;
    },
  },
  mounted() {
    this.loadGroups();
  },
  methods: {
    async loadGroups() {
      const url = '/api/dcos/tdac-cgi/access-groups/card';
      const resp = await this.$axios.get(url, {
        offset: 0,
        limit: 100000,
      });
      this.allGroups = resp;
    },
  },
};
</script>

<style lang="scss" scoped>
.filters-form {
  border-top: 1px solid #efefef;
}
</style>
