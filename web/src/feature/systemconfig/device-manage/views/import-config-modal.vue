<template>
  <el-modal
    :visible.sync="visible"
    :title="title"
    :width="1080"
  >
    <el-form
      v-if="formData"
      ref="form"
      :model="formData"
    >
      <el-form-item
        label="文件"
        required
        prop="files"
      >
        <input
          :multiple="currentImportTargetConfig && currentImportTargetConfig.multipleFiles"
          type="file"
          @change="handleFilesInputChange"
        >
      </el-form-item>

      <el-form-item
        label="应用范围"
        required
        prop="tboxList"
      >
        <span class="gray-text">
          已选 <em>{{ formData.tboxList.length }}</em> 个
        </span>
      </el-form-item>
    </el-form>

    <el-table
      v-if="visible"
      ref="table"
      :data="tableData"
      row-key="id"
      @selection-change="handleSelectionChange"
    >
      <el-table-column
        type="selection"
        reserve-selection
      />

      <el-table-column
        label="Tbox编号"
        prop="name"
      />

      <el-table-column
        label="IP"
        prop="ip"
      />

      <el-table-column
        label="状态"
        prop="status"
      >
        <template #default="{ row }">
          <el-tag
            v-if="statusMap[row.isOnline ? 'online' : 'offline']"
            :type="statusMap[row.isOnline ? 'online' : 'offline'].tagType || 'info'"
          >
            {{ statusMap[row.isOnline ? 'online' : 'offline'].label || '--' }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>

    <template #footer>
      <div class="footer">
        <el-button
          @click="cancel"
        >
          取消
        </el-button>

        <el-button
          type="primary"
          @click="submit"
        >
          提交
        </el-button>
      </div>
    </template>
  </el-modal>
</template>

<script>
export default {
  props: {
    treeData: {
      type: Array,
      default() {
        return [];
      },
    },
  },
  data() {
    return {
      visible: false,
      importTarget: null,
      formData: {
        files: null,
        tboxList: [],
      },
      formVersion: 0,

      statusMap: {
        online: {
          label: '在线',
          tagType: 'success',
        },
        offline: {
          label: '离线',
          tagType: 'danger',
        },
      },

      importTagetConfigsMap: {
        CollectConfig: {
          title: '采集配置',
          url: '/api/dcos/tboxmonitor-cgi/collector/forward/batch/restful/api/tbcm/devices/tpl',
          fileField: 'file',
          multipleFiles: false,
        },
        DriverConfig: {
          title: '驱动模板',
          url: '/api/dcos/tboxmonitor-cgi/collector/forward/batch/restful/api/tbcm/tpls',
          fileField: 'files',
          multipleFiles: true,
        },
      },
    };
  },
  computed: {
    currentImportTargetConfig() {
      return this.importTagetConfigsMap[this.importTarget];
    },
    title() {
      return `导入【${this.currentImportTargetConfig?.title}】`;
    },
    tableData() {
      const { treeData } = this;
      return (treeData?.[0]?.children || []).map(item => ({
        ...item,
        children: undefined,
      }));
    },
  },
  methods: {
    open(importTarget) {
      this.importTarget = importTarget;
      this.visible = true;
      this.formData = {
        files: null,
        tboxList: [],
      };
      this.formVersion += 1;
    },
    cancel() {
      this.visible = false;
      this.formData = null;
    },
    async submit() {
      const {
        formData,
        // importTarget,
        currentImportTargetConfig: postConfig,
      } = this;

      await this.$refs.form.validate();

      const formDataToPost = new FormData();

      if (formData.files) {
        formData.files.forEach((file) => {
          formDataToPost.append(postConfig.fileField, file);
        });
      }

      const tboxIds = formData.tboxList.map(item => item.id);

      await this.$axios.ins({
        url: postConfig.url,
        method: 'POST',
        data: formDataToPost,
        headers: {
          'X-TboxMonitor-Payload': JSON.stringify({
            ids: tboxIds,
          }),
        },
      });

      this.$message.success('导入完成');
    },
    updateTableSelection() {
      const { tableData, formData } = this;
      const tableMap = _.mapKeys(tableData, 'id');
      _.chain(formData.tboxList)
        .map(item => tableMap[item.id])
        .forEach((item) => {
          this.$refs.table.toggleRowSelection(item, true);
        })
        .value();
    },
    handleFilesInputChange(event) {
      const files = event.target?.files;
      if (!files?.length) {
        this.formData.files = [];
        return;
      }
      this.formData.files = this.currentImportTargetConfig?.multipleFiles
        ? [...files]
        : [files[0]];
    },
    handleSelectionChange(selectedList) {
      this.formData.tboxList = selectedList;
    },
  },
};
</script>

<style lang="scss" scoped>
.gray-text {
  color: #a0a0a0;
}

.footer {
  text-align: right;
  padding: 0 32px;
}
</style>
