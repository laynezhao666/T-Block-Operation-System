
<template>
  <el-dialog
    :visible.sync="addVisible"
    width="800px"
    :title="`${device ? '分配' : '添加'}采集器`"
    @close="addVisible = false"
  >
    <div class="dialog-container">
      <div class="form">
        <el-form
          ref="form"
          :model="defaultForm"
          label-width="120px"
          :rules="formRules"
        >
          <el-form-item
            label="TBOX名称"
            prop="name"
          >
            {{ defaultForm.name }}
          </el-form-item>
          <el-form-item
            label="TBOX地址"
            prop="ip"
          >
            <el-row
              :gutter="15"
              type="flex"
            >
              {{ defaultForm.ip }}
            </el-row>
          </el-form-item>
          <el-form-item
            label="分配方式"
            prop="ip"
          >
            <el-row
              :gutter="15"
              type="flex"
            >
              <el-radio-group v-model="defaultForm.allocateWay">
                <el-radio-button label="重新分配" />
                <el-radio-button label="从备份还原" />
              </el-radio-group>
            </el-row>
          </el-form-item>
          <el-form-item
            v-if="defaultForm.allocateWay === '重新分配'"
            label="安装位置"
            prop="room"
          >
            <el-row
              :gutter="15"
              type="flex"
            >
              <el-col :span="12">
                <el-select
                  v-model="defaultForm.room"
                  placeholder="请选择房间"
                  border-type="bordered"
                  clearable
                >
                  <el-option
                    v-for="item in roomsOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-col>
              <el-col :span="12">
                <el-select
                  v-model="defaultForm.block"
                  placeholder="请选择方仓"
                  border-type="bordered"
                  clearable
                >
                  <el-option
                    v-for="item in blockOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-col>
            </el-row>
          </el-form-item>
          <el-form-item
            v-if="defaultForm.allocateWay === '重新分配'"
            label="采集器编号"
            prop="no"
          >
            <el-input
              v-model="defaultForm.no"
              placeholder="默认为1"
              border-type="bordered"
            />
          </el-form-item>
          <el-form-item
            v-if="defaultForm.allocateWay === '从备份还原'"
            label="选择备份"
            prop="collectorId"
          >
            <el-row
              :gutter="15"
              type="flex"
            >
              <el-col :span="11">
                <el-select
                  v-model="defaultForm.collectorId"
                  placeholder="采集器名｜IP"
                  border-type="bordered"
                  clearable
                >
                  <el-option
                    v-for="item in collectorOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-col>
              <el-col :span="13">
                <el-select
                  v-model="defaultForm.backupId"
                  placeholder="选择备份版本，默认最新"
                  border-type="bordered"
                  clearable
                >
                  <el-option
                    v-for="item in backupOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-col>
            </el-row>
          </el-form-item>
        </el-form>
      </div>
    </div>
    <div
      slot="footer"
      class="dialog-footer"
    >
      <el-button
        type="text"
        @click="addVisible = false"
      >
        取消
      </el-button>
      <el-button
        type="text"
        @click="handleConfirm"
      >
        确定
      </el-button>
    </div>
  </el-dialog>
</template>

<script>
import { collectorApi } from '@@/config/cgi';
export default {
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    device: {
      type: Object,
      default: () => null,
    },
  },
  data() {
    const validatePosition = (rule, value, callback) => {
      if (value === '') {
        callback(new Error('请选择房间'));
      } else if (!this.defaultForm.block) {
        callback(new Error('请选择方仓'));
      } else {
        callback();
      }
    };
    const validateBackup = (rule, value, callback) => {
      if (value === '') {
        callback(new Error('请选择采集器'));
      } else if (!this.defaultForm.backupId) {
        callback(new Error('请选择备份'));
      } else if (this.defaultForm.backupId !== this.collectorConfig[value].backups[0].id) {
        callback(new Error('系统检测有相同IP采集器或更新备份，请确认是否选择最合适的版本！'));
      } else {
        callback();
      }
    };
    return {
      rules: {
        room: [
          { required: true, message: '请选择房间', trigger: 'blue' },
          { validator: validatePosition, trigger: ['blur', 'change'] },
        ],
        no: [
          { required: true, message: '请输入采集器编号', trigger: 'blur' },
        ],
      },
      backupRules: {
        collectorId: [
          { required: true, message: '请选择采集器', trigger: 'blur' },
          { validator: validateBackup, trigger: ['blur', 'change'] },
        ],
      },
      defaultForm: {
        id: '',
        name: '',
        room: '',
        block: '',
        ip: '',
        no: '1',
        port: '80',
        allocateWay: '重新分配',
        collectorId: '',
        backupId: '',
      },
      roomsOptions: [],
      blockOptions: [],
      roomsConfig: {},
      collectorOptions: [],
      backupOptions: [],
      collectorConfig: {},
    };
  },
  computed: {
    addVisible: {
      set(v) {
        this.$emit('update:visible', v);
      },
      get() {
        return this.visible;
      },
    },
    formRules() {
      return this.defaultForm.allocateWay === '重新分配' ? this.rules : this.backupRules;
    },
  },
  watch: {
    visible: {
      handler(v) {
        if (v) {
        }
      },
      immediate: true,
    },
    device: {
      handler(v) {
        if (v) {
          // 获取采集器详情
          if (!v) return;
          this.getCollectorInfo();
          this.getAllBacks();
        } else {
          // 添加采集器
        }
      },
      immediate: true,
    },
    'defaultForm.room': {
      immediate: true,
      handler(v) {
        this.updateBlockOptions();
      },
    },
    'roomsConfig': {
      handler() {
        this.updateBlockOptions();
      }
    },
    'defaultForm.collectorId': {
      immediate: true,
      handler(v) {
        if (v) {
          this.backupOptions = this.collectorConfig[v]?.backups?.map(backup => ({
            label: backup.version,
            value: backup.id,
          }));
          this.defaultForm.backupId = this.collectorConfig[v]?.backups[0]?.id;
        } else {
          this.backupOptions = [];
        }
      },
    },

  },
  mounted() {
    this.getRoomConfig();
  },
  methods: {

    getCollectorInfo() {
      this.$axios.post(collectorApi.queryCollectorDetail, {
        id: this.device.id,
        assigned: !this.device.isUnassigned,
      }).then((res) => {
        this.defaultForm.name = res.name;
        this.defaultForm.ip = res.link_channel.chid;
        this.defaultForm.room = res.position.room;
        this.defaultForm.block = res.position.block;
        this.defaultForm.id = res.id;
      })
        .catch((err) => {
          console.log(err);
        });
    },
    getAllBacks() {
      this.$axios.post(collectorApi.getAllBackups2, {
        ip: this.device.ip,
      }).then((res) => {
        this.collectorConfig = _.keyBy(res, 'id');
        if (res[0].ip === this.device.ip) {
          this.defaultForm.collectorId = `${res[0].name} | ${res[0].ip}`;
          this.defaultForm.backupId = this.collectorConfig[res[0].id].backups[0];
        }
        this.collectorOptions = res.map(item => ({
          label: `${item.name} | ${item.ip}`,
          value: item.id,
        }));
      })
        .catch((err) => {
          console.log(err);
        });
    },
    getRoomConfig() {
      this.$axios.get(collectorApi.getRoomConfig).then((res) => {
        this.roomsOptions = Object.keys(res).map(room => ({
          label: room,
          value: room,
        }));
        this.roomsConfig = res;
      })
        .catch((err) => {
          console.log(err);
        });
    },
    updateBlockOptions() {
      const v = this.defaultForm?.room;
      if (v) {
        this.defaultForm.block = '';
        this.blockOptions = this.roomsConfig[v]?.filter(block => !block.includes('IEAC')).map(block => ({
          label: block,
          value: block,
        }));
      } else {
        this.blockOptions = [];
      }
    },
    handleTest() {
      this.$message('TODO 测试');
    },
    handleConfirm() {
      // 校验
      if (this.defaultForm.allocateWay === '从备份还原') {
        // restoreBackup
        this.$refs.form.validate((valid) => {
          if (valid) {
            // eslint-disable-next-line camelcase
            const { id: collector_id, backupId: id } = this.defaultForm;
            this.$axios.post(collectorApi.restoreBackup2, {
              id,
              collector_id,
            }).then(() => {
              this.$message.success('备份还原成功');
              this.addVisible = false;
              this.$emit('confirm', { ...this.form });
            })
              .catch(() => {
                this.$message.error('备份还原失败');
              });
          } else {
            return false;
          }
        });
      } else {
        this.$refs.form.validate((valid) => {
          if (valid) {
            const { id, room, block, no } = this.defaultForm;
            this.$axios.post(collectorApi.editCollector, {
              id,
              room,
              block,
              no,
              assign: false,
            }).then(() => {
              this.$message.success('分配成功');
              this.addVisible = false;
              this.$emit('confirm', { ...this.form });
            })
              .catch(() => {
                this.$message.error('分配失败');
              });
          } else {
            console.log('error submit!!');
            return false;
          }
        });
      }
    },
  },
};
</script>

<style lang="scss" scoped>
/deep/ .el-input.is-bordered .el-input__inner {
  border-radius: 4px;
}
.dialog-container {
  width: 100%;
  display: flex;
  align-items: center;
  .form {
    flex: 1;
  }
}
</style>
