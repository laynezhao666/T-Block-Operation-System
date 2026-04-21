<template>
  <el-modal
    :visible.sync="visible"
    :width="1080"
  >
    <template slot="title">
      {{ isCreate ? '新增' : '编辑' }}分组
    </template>

    <el-form
      v-if="group"
      ref="form"
      :model="group"
      :rules="rules"
      label-width="120px"
    >
      <el-form-item
        label="分组名称"
        prop="name"
      >
        <el-input
          v-model="group.name"
          placeholder="请输入分组名称"
        />
      </el-form-item>

      <el-form-item
        label="关联门设备"
        prop="doors"
      >
        <el-transfer
          v-model="group.doors"
          :filter-method="transformFilter"
          :data="allDoors"
          :props="transferProps"
          :titles="['待选', '已选']"
          filterable
          filter-placeholder="请输入关键词"
          class="transfer"
        />
      </el-form-item>
    </el-form>

    <template slot="footer">
      <el-button
        type="primary"
        @click="submit"
      >
        提交
      </el-button>

      <el-button
        @click="close"
      >
        取消
      </el-button>
    </template>
  </el-modal>
</template>

<script>
import { axiosPut } from '../../../../utils/axios-methods';
import getEdgeRequest from 'feature/utils/request';

export default {
  data() {
    return {
      group: null,
      callback: null,

      allDoors: [],

      transferProps: {
        key: 'id',
        label: 'name',
      },

      rules: {
        name: [
          { required: true, message: '分组名称不能为空' },
        ],
      },
    };
  },
  computed: {
    visible: {
      get() {
        return Boolean(this.group);
      },
      set(visible) {
        if (!visible) {
          this.close();
        }
      },
    },
    isCreate() {
      return !this.group?.id;
    },
  },
  methods: {
    async loadAllDoors() {
      const controlList = await getEdgeRequest(this.$axios).get('/api/dcos/tdac-cgi/controllers');

      this.allDoors = _.chain(controlList)
        .map('doors')
        .flatten()
        .filter(Boolean)
        .value();
    },
    transformFilter(query, item) {
      if (!query?.trim()) return true;

      return item.name.includes(query.trim());
    },
    edit(group, callback) {
      this.group = {
        ...group,
        doors: _.map(group.doors, 'id'),
      };

      this.callback = callback;

      this.loadAllDoors();
    },
    close() {
      this.group = null;
      this.callback = null;
    },
    async saveToServer() {
      const { group } = this;

      // 保存分组
      if (this.isCreate) {
        await getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/group', group);
      } else {
        await axiosPut('/api/dcos/tdac-cgi/group', {
          name: group.name,
          id: group.id,
          doors: group.doors,
        });
      }
      // 修改门的分组信息
      this.$message.success('保存成功');
      this.$emit('reloadTree');
    },
    async submit() {
      if (!(await this.$refs.form.validate())) return;

      await this.saveToServer();

      if (this.callback) {
        this.callback(this.group);
      }
      this.close();
    },
  },
};
</script>

<style lang="scss" scoped>
.transfer /deep/ {
  .el-transfer-panel {
    width: 400px;
  }

  .el-transfer-panel__body {
    height: calc(100vh - 284px - 38px) !important;
  }

  .el-transfer-panel__list {
    height: calc(100vh - 284px) !important;
  }

  .el-transfer-panel__item.el-checkbox {
    width: 100%;
    box-sizing: border-box;
    margin-right: 0;
  }
}
</style>
