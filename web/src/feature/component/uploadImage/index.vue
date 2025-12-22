<template>
  <el-upload
    ref="upload"
    :action="commonCgi.uploadImage"
    list-type="picture-card"
    :on-success="onSuccess"
    :on-remove="onRemove"
    :before-upload="beforeUpload"
    :file-list="fileList"
  >
    <i class="el-icon-plus" />
  </el-upload>
</template>

<script>
export default {
  props: {
    pic: {
      type: Object,
      default: () => {},
    },
  },
  inject: ['commonCgi', 'configCgi'],
  data() {
    return {
      limit: 5,
      separator: ';',
      fileList: [],
      fileListKey: this.pic.picList.split(';').filter(Boolean),
      fieldName: this.pic.fieldName,
    };
  },
  watch: {
    fileListKey(v) {
      this.$emit('updatePic', v.join(';'), this.fieldName);
    },
  },
  mounted() {
    this.initFileList();
  },
  methods: {
    initFileList() {
      this.fileListKey.forEach((key) => {
        this.fileList.push({
          key,
          url: `${this.commonCgi.downloadImage}?key=${key}`,
        });
      });
    },
    onSuccess({ code, data, message }) {
      if (code) {
        this.$message.error(message || '上传失败');
      } else {
        this.fileListKey.push(data);
      }
    },
    beforeUpload() {
      if (this.fileListKey.length >= this.limit) {
        this.$message.error(`最多只能上传${this.limit}张照片`);
        return false;
      }
    },
    onRemove(file) {
      const key = file.response ? file.response.data : file.key;
      const ind = this.fileListKey.indexOf(key);
      if (ind > -1) {
        this.fileListKey.splice(ind, 1);
      } else {
        this.$message.error('删除失败');
      }
    },
  },
};
</script>
