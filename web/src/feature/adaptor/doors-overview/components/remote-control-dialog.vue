<template>
  <el-dialog
    :visible.sync="visible"
    title="远程控制"
  >
    <el-form
      ref="form"
      :model="formData"
      :rules="rules"
      label-width="120px"
    >
      <el-form-item
        label="设备名称"
        prop="opr"
      >
        <el-radio-group
          v-model="formData.opr"
          variant="default-filled"
        >
          <el-radio label="1">
            远程开门
          </el-radio>
          <el-radio label="2">
            设置门常开
          </el-radio>
          <el-radio label="3">
            设置门常闭
          </el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>

    <span
      slot="footer"
      class="dialog-footer"
    >
      <el-button @click="close">
        取消
      </el-button>

      <el-button
        type="primary"
        @click="submit"
      >
        确定
      </el-button>
    </span>
  </el-dialog>
</template>

<script>
import getEdgeRequest from 'feature/utils/request';

export default {
  data() {
    return {
      doorIds: null,
      formData: {
        opr: '',
      },

      rules: {
        rules: [
          { required: true, message: '请选择操作类型' },
        ],
      },
    };
  },
  computed: {
    visible: {
      get() {
        return Boolean(this.doorIds?.length);
      },
      set(v) {
        if (!v) {
          this.close();
        }
      },
    },
  },
  methods: {
    show(doorIds) {
      this.doorIds = doorIds;
    },
    close() {
      this.doorIds = null;
    },
    async submit() {
      if (!(await this.$refs.form.validate())) return;

      await this.toggleDoorsStatusByIds(this.doorIds, this.formData.opr);
      this.$message.success('远程控制成功');

      this.close();
    },
    async toggleDoorsStatusByIds(ids, opr) {
      if (!ids?.length) return;

      return getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/door/state', {
        ids,
        state: Number(opr),
      });
    },
  },
};
</script>

<style lang="scss" scoped>

</style>
