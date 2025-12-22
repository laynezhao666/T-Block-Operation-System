<template>
  <el-modal
    :visible.sync="visibleData"
    custom-layout
  >
    <template slot="title">
      导入{{ text }}
    </template>

    <config-uploader
      v-if="visibleData"
      :urls="urls"
      passed-text="校验通过，点击确定后生效"
      :params="{}"
      :t-columns="['errType', 'record_index', 'message']"
      :type-text="text"
      @success="onSuccess"
      @error="curFile = void 0"
    />

    <template slot="footer">
      <el-button
        type="primary"
        :disabled="!curFile"
        @click="save"
      >
        确定
      </el-button>
    </template>
  </el-modal>
</template>
<script>
import ConfigUploader from '../ConfigTool/ConfigUploader';

export default {
  components: {
    ConfigUploader,
  },
  inject: ['configCgi', 'tableConfig'],
  props: {
    visible: Boolean,
    text: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      visibleData: this.visible,
      curFile: void 0,
    };
  },
  computed: {
    urls() {
      return {
        upload: this.configCgi.uploadCgi,
        download: this.configCgi.downloadCgi,
      };
    },
  },
  watch: {
    visible(v) {
      if (this.visibleData !== v) {
        this.visibleData = v;
      }
    },
    visibleData(v) {
      if (!v) {
        this.curFile = void 0;
      }
      this.$emit('update:visible', v);
    },
  },
  methods: {
    save() {
      this.$axios.post(this.configCgi.applyImportTpl, this.curFile).then(() => {
        this.visibleData = false;
        this.$message({
          type: 'success',
          message: '导入成功',
        });
        this.$emit('success');
      });
    },
    onSuccess(data) {
      this.curFile = {
        skey: data.skey,
        spass: data.spass,
      };
    },
  },
};
</script>
