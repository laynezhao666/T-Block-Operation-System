<template>
  <el-dialog
    title="导入门映射"
    :visible.sync="dialogVisible"
    width="600px"
  >
    <el-form inline>
      <el-form-item
        label="模板文件"
        prop="file"
      >
        <input
          type="file"
          @change="handleFileChange"
        >

        <br>

        <el-button
          type="text"
          @click="handleDownload"
        >
          下载门信息表
        </el-button>
      </el-form-item>
    </el-form>

    <span
      slot="footer"
      class="dialog-footer"
    >
      <el-button
        type="text"
        @click="dialogVisible = false"
      >取消</el-button>
      <el-button
        :disabled="!formData.file"
        type="text"
        @click="submit"
      >确定</el-button>
    </span>
  </el-dialog>
</template>

<script>
import { axiosUploadFile } from '../../../../utils/axios-methods';

export default {
  data() {
    return {
      dialogVisible: false,
      formData: {
        file: null,
      },
    };
  },
  methods: {
    open() {
      this.dialogVisible = true;
      this.formData = {
        file: null,
      };
    },
    async submit() {
      const {
        file,
      } = this.formData;

      if (!file) return;

      await axiosUploadFile('/api/dcos/tdac-cgi/doors/import/code', {
        file,
      });

      this.$message.success('导入门映射成功');
      this.dialogVisible = false;
      this.formData = {
        file: null,
      };
    },
    handleFileChange(evt) {
      [this.formData.file] = evt.target.files || [];
    },
    handleDownload() {
      window.open('/api/dcos/tdac-cgi/doors/export/code', '_blank');
    },
  },
};
</script>
