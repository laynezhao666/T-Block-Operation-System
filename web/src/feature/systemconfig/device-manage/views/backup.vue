
<template>
  <div v-loading="vLoading">
    <div class="operate">
      <el-button
        type="primary"
        icon="tn-icon-add"
        class="button-round"
        @click="handleConstruct"
      >
        构建备份
      </el-button>
    </div>

    <el-table
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column
        prop="version"
        label="版本号"
      />
      <el-table-column
        prop="size"
        label="文件大小"
        width="180"
      >
        <template slot-scope="scope">
          {{ (scope.row.size / (1024 * 1024)).toFixed(2) }}MiB
        </template>
      </el-table-column>
      <el-table-column
        prop="backup_time"
        label="备份时间"
      />
      <el-table-column
        prop="operator"
        label="备份人"
        :filters="nameList"
        :filter-method="filterName"
      />
      <el-table-column
        prop="description"
        label="描述"
      >
        <template slot-scope="scope">
          <el-tooltip
            class="item"
            :content="scope.row.description"
            placement="top"
          >
            <div class="table-cell-remark">
              {{ scope.row.description }}
            </div>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column
        fixed="right"
        label="操作"
        width="200"
      >
        <template slot-scope="scope">
          <el-button
            type="text"
            size="small"
            @click="handleBackup(scope.row)"
          >
            还原
          </el-button>
          <el-button
            type="text"
            size="small"
            @click="downloadBackup(scope.row)"
          >
            下载
          </el-button>
          <el-button
            type="text"
            size="small"
            @click="deleteBackup(scope.row)"
          >
            移除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog
      title="构建备份"
      :visible.sync="showDialog"
      width="700px"
    >
      <el-form
        :model="form"
        label-width="120px"
      >
        <el-form-item label="版本号：">
          <el-row
            :gutter="24"
            type="flex"
          >
            <el-col :span="20">
              <el-input
                v-model="form.version"
                autocomplete="off"
                border-type="bordered"
              />
            </el-col>
            <!-- <el-col :span="4">
              <el-checkbox v-model="autoFill">
                自动
              </el-checkbox>
            </el-col> -->
          </el-row>
        </el-form-item>
        <el-form-item label="版本名称">
          <el-row
            :gutter="24"
            type="flex"
          >
            <el-col :span="20">
              <el-input
                v-model="form.name"
                autocomplete="off"
                border-type="bordered"
              />
            </el-col>
          </el-row>
        </el-form-item>
        <el-form-item label="描述">
          <el-row
            :gutter="24"
            type="flex"
          >
            <el-col :span="20">
              <el-input
                v-model="form.description"
                type="textarea"
                :autosize="{ minRows: 2, maxRows: 4}"
                placeholder="请输入内容"
                border-type="bordered"
              />
            </el-col>
          </el-row>
        </el-form-item>
      </el-form>
      <div
        slot="footer"
        class="dialog-footer"
      >
        <el-button
          type="text"
          @click="showDialog = false"
        >
          取消
        </el-button>
        <el-button
          type="text"
          @click="createBackup"
        >
          确定
        </el-button>
      </div>
    </el-dialog>

    <el-dialog
      title="操作确认"
      :visible.sync="backupConfirmVisible"
      width="400px"
    >
      <h2
        style="font-family:'PingFangSC-Semibold','PingFang SC Semibold','PingFang SC',
        sans-serif;font-weight:650;color:#666666;"
      >
        还原配置将采集器配置将被覆盖，是否继续？
      </h2>
      <span>
        采集器导入还原配置后将自动进行服务重启，重启后还原配置生效
      </span>
      <span
        slot="footer"
        class="dialog-footer"
      >
        <el-button
          type="text"
          @click="backupConfirmVisible = false"
        >取消</el-button>
        <el-button
          type="text"
          @click="restoreBackup"
        >确定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { collectorApi } from '@@/config/cgi';
import axios from 'axios';
export default {
  props: {
    collector: {
      type: Object,
      default: () => null,
    },
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      vLoading: false,
      autoFill: false,
      showDialog: false,
      backupConfirmVisible: false,
      tableData: [],
      nameList: [
        { text: 'airyao', value: 'airyao' },
        { text: 'mingfengli', value: 'mingfengli' },
      ],
      form: {
        version: '', // 备份版本
        description: '', // 备份描述
        name: '', // 备份名称
      },
      restoreData: null,
    };
  },
  watch: {
    collector(v) {
      if (!v) return;
      this.getBackupdata();
    },
  },
  methods: {
    getBackupdata() { // 获取备份还原数据
      this.vLoading = true;
      this.$axios.post(collectorApi.getCollectorBackupList, {
        id: this.collector.id,
      }).then((res) => {
        this.tableData = res;
        this.vLoading = false;
      })
        .catch((err) => {
          this.vLoading = true;
          this.tableData = [];
          console.log(err);
        });
    },
    handleClick() {

    },
    filterName(value, row) {
      return row.name === value;
    },
    handleConstruct() {
      this.showDialog = true;
    },
    handleBackup(row) {
      this.restoreData = row;
      this.backupConfirmVisible = true;
    },
    createBackup() {
      const params = {
        ...this.form,
        id: this.collector.id,
      };
      this.$axios.post(
        collectorApi.createCollectorBackup,
        params,
        true,
        { restAxios: { timeout: 60000 } }
      ).then(() => {
        this.$message.success('备份成功');
        this.showDialog = false;
        this.getBackupdata();
      })
        .catch((err) => {
          this.showDialog = false;
          console.log(err);
        });
    },
    downloadBackup(row) {
      window.open(`${collectorApi.downloadBackup}/${row.id}`);
    },
    deleteBackup(row) {
      this.$confirm(`确认删除${row.version}备份吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        axios.delete(collectorApi.deleteBackup, {
          data: {
            id: row.id,
          },
        }).then(() => {
          this.getBackupdata();
          this.$message.success('删除成功');
        });
      })
        .catch(() => {});
    },
    restoreBackup() {
      this.$axios.post(collectorApi.restoreBackup, {
        id: this.restoreData.id,
      }).then(() => {
        this.getBackupdata();
        this.backupConfirmVisible = false;
        this.$message.success('还原成功');
      })
        .catch((err) => {
          console.log(err);
        });
    },

  },
};
</script>

<style lang="scss" scoped>
.operate {
  padding: 15px;
}
.table-cell-remark {
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
}
.button-round {
  border-radius: 5px;
  padding: 6px;

}
/deep/ .el-form-item__content .el-textarea {
  // border: 1px solid silver;
  // border-radius: 4px;
}
/deep/ .el-textarea:before {
  // border: none;
}
/deep/ .el-dialog__body {
    padding: 18px 40px 30px;
}

</style>
