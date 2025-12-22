<template>
  <div>
    <el-block
      class="upload-wrap"
    >
      <el-form
        ref="form"
        label-width="160px"
      >
        <el-form-item>
          <template slot="label">
            <span>下载模板</span>
          </template>
          <a
            class="text-primary"
            :href="urls.download"
            target="_self"
          >
            {{ typeText || '模组' }}配置模版 {{ version }}
          </a>
        </el-form-item>
        <el-form-item>
          <template slot="label">
            <span>上传模板</span>
          </template>
          <el-upload
            :action="urls.upload"
            :multiple="false"
            :on-success="uploadSuccess"
            :show-file-list="false"
            :data="params"
            :before-upload="beforeUpload"
          >
            <el-button>
              <i class="tn-icon-upload" />上传文件
            </el-button>
          </el-upload>
        </el-form-item>
      </el-form>
    </el-block>
    <el-block
      v-if="errors.length"
      no-padding
    >
      <div
        class="error-text"
      >
        校验不通过，请参考以下内容修改配置文件：
        <span v-if="errEllipsis && errLen>0">
          <el-button
            type="text"
            @click="downloadErrInfo"
          >
            点击此处下载文件查看错误信息
          </el-button>
        </span>
      </div>
      <custom-table
        :columns="columns"
        :data="errors"
      />
    </el-block>
    <el-block
      no-padding
    >
      <div
        v-if="rst"
        class="error-text"
      >
        <slot
          name="passedText"
          v-bind="rst"
        >
          {{ passedText }}
        </slot>
      </div>
      <div
        v-if="errEllipsis && errLen>50"
        class="error-text red-text"
      >
        <slot
          name="failText"
        >
          {{ errText }}
        </slot>
      </div>
    </el-block>
  </div>
</template>
<script>
import { find } from 'lodash';
import CustomTable from '../Table';

const allColumns = [{
  prop: 'sheet_name',
  label: '工作表',
  width: '200',
}, {
  prop: 'errType',
  label: '错误类型',
  width: '200',
}, {
  prop: 'record_index',
  label: '行/列',
  width: '120',
}, {
  prop: 'message',
  label: '不通过原因',
}];

export default {
  components: {
    CustomTable,
  },
  props: {
    urls: {
      type: Object,
      default: () => ({}),
    },
    method: {
      type: String,
      default: 'download',
    },
    version: {
      type: String,
      default: '',
    },
    typeText: {
      type: String,
      default: '',
    },
    passedText: {
      type: String,
      default: '',
    },
    params: {
      type: Object,
      default: () => ({}),
    },
    tColumns: {
      type: Array,
      default: () => ['sheet_name', 'record_index', 'message'],
    },
  },
  data() {
    const columns = this.tColumns.map(col => find(allColumns, {
      prop: col,
    }));
    return {
      columns,
      rst: void 0,
      errors: [],
      errLen: 0,
      errKey: void 0,
      errText: '错误过多，请下载文件查看错误信息',
    };
  },
  computed: {
    errEllipsis() {
      //  idcdb的导入文件只截取前50行
      return window.location.href.includes('idcdb');
    },
  },
  methods: {
    downloadTpl() {
      this.$axios[this.method](this.urls.download);
    },
    downloadErrInfo() {
      this.$axios[this.method](this.urls.downloadErrInfo, {
        errkey: this.errKey,
        errfailname: this.errFailName,
      });
    },
    uploadSuccess({ code, data, message }) {
      if (code !== 0) {
        this.$message({
          type: 'error',
          message,
        });
      } else {
        if (data?.result === 'passed') {
          this.rst = data;
          this.errors = [];
          this.$emit('success', data);
        } else {
          this.rst = void 0;
          this.errKey = data.errKey || '';
          this.errLen = data.errLen || 0;
          this.errFailName = data.errFailName || '';
          this.errors = data.invalidations || [];
          this.$emit('error', data);
        }
      }
    },
    beforeUpload(file) {
      if (!file.name.endsWith('.xlsx')) {
        this.$message({
          type: 'error',
          message: '配置文件后缀应为“.xlsx”。',
        });
        return false;
      }
    },
  },
};
</script>
<style lang="scss" scoped>
@import "~common/style/mixin";

.upload-wrap {
  padding: $space-l;
  background-color: white;
}

.error-text {
  padding: $space-l;
  font-size: $font-size-m;
}

.red-text {
  color: red;
}

</style>
