<template>
  <el-modal :visible.sync="visible">
    <template slot="title">
      {{ isCreate ? '新增' : '编辑' }} 门禁控制器
    </template>

    <el-form
      v-if="control"
      ref="form"
      :model="control"
      :rules="rules"
      label-width="120px"
    >
      <el-form-item
        label="设备名称"
        prop="name"
      >
        <el-input
          v-model="control.name"
          placeholder="请输入设备名称"
        />
      </el-form-item>

      <el-form-item
        label="厂商/型号"
        prop="profile.vendor"
      >
        <div class="flex-form-item-content">
          <el-input
            v-model="control.profile.vendor"
            placeholder="厂商"
          />

          <el-input
            v-model="control.profile.model"
            placeholder="型号"
          />
        </div>
      </el-form-item>

      <el-form-item
        label="安装位置"
        prop="position"
      >
        <div class="flex-form-item-content">
          <el-select
            v-model="control.position.room"
            placeholder="选择房间"
            @change="handleRoomChange"
          >
            <el-option
              v-for="(children, roomKey) in roomsMap"
              :key="roomKey"
              :value="roomKey"
              :label="roomKey"
            />
          </el-select>

          <el-select
            v-model="control.position.block"
            placeholder="选择方仓"
          >
            <el-option
              v-for="(blockKey, i) in roomsMap[control.position.room]"
              :key="i"
              :value="blockKey"
              :label="blockKey"
            />
          </el-select>
        </div>
      </el-form-item>

      <div class="control-form-split-header-bar">
        <split-header-bar
          title="通讯参数"
          no-padding
        >
          <el-button
            type="text"
            @click="testLink"
          >
            连接测试
          </el-button>
        </split-header-bar>
      </div>

      <el-form-item
        label="IP地址"
        prop="channel.chid"
      >
        <div class="flex-form-item-content">
          <el-input
            v-model="control.channel.chid"
            placeholder="IP地址"
          />

          <el-input
            v-model="control.channel.chidPort"
            placeholder="端口"
            style="width: 160px;"
          />

          <el-input
            v-model="control.protocol.version"
            placeholder="协议版本"
            style="width: 240px;"
          />
        </div>
      </el-form-item>

      <el-form-item
        label="序列号"
        prop="profile.sn"
      >
        <el-input
          v-model="control.profile.sn"
          placeholder="序列号"
        />
      </el-form-item>

      <el-form-item
        label="账号密码"
        prop="position.vendor"
      >
        <div class="flex-form-item-content">
          <!-- 对接接口 -->
          <el-input
            v-model="control.account"
            placeholder="通信账号"
          />
          <!-- 对接接口 -->
          <el-input
            v-model="control.password"
            type="password"
            show-password
            placeholder="通信密码"
          />
        </div>
      </el-form-item>

      <el-form-item
        label="超时时间"
        prop="channel.timeout"
      >
        <el-input
          v-model="control.channel.timeout"
          placeholder="连接超时时间，单位ms"
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
import SplitHeaderBar from '../../../component/tedge-components/split-header-bar.vue';
import getEdgeRequest from 'feature/utils/request';

export default {
  components: {
    SplitHeaderBar,
  },
  data() {
    return {
      control: null,
      callback: null,

      allDoors: [],
      roomsMap: {},

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
        return Boolean(this.control);
      },
      set(visible) {
        if (!visible) {
          this.close();
        }
      },
    },
    isCreate() {
      return Boolean(!this.control?.id);
    },
  },
  methods: {
    async loadAllDoors() {
      const controlList = await getEdgeRequest(this.$axios).get('/api/dcos/tdac-cgi/controllers');

      this.allDoors = _.chain(controlList)
        .map('doors')
        .flatten()
        .value();
    },
    async loadRooms() {
      this.roomsMap = await getEdgeRequest(this.$axios).get('/api/dcos/tboxmonitor-cgi/rooms');
    },
    transformFilter(query, item) {
      if (!query?.trim()) return true;
      return item.name.includes(query.tirm());
    },
    edit(control, callback) {
      this.control = this.normalizeForEdit(_.cloneDeep(control));
      this.callback = callback;

      this.loadAllDoors();
      this.loadRooms();
    },
    close() {
      this.control = null;
      this.callback = null;
    },
    async testLink() {
      const {
        control,
      } = this;

      const {
        channel,
        protocol,
      } = control;

      const result = await getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/test', {
        host: `${channel.chid}:${channel.chidPort}`,
        protocol_name: protocol.name,
        protocol_version: protocol.version,
        timeout: channel.timeout,
        account: control.account,
        password: control.password,
      });
      this.$message.success(result);
    },
    normalizeForEdit(control) {
      const ipSplited = (control.channel.chid || '').split(':');
      return {
        ...control,
        channel: {
          ...control.channel,
          chid: ipSplited[0] || '',
          chidPort: ipSplited[1] || '80',
        },
      };
    },
    normalizeForSave(control) {
      const ipJoined = `${control.channel.chid || ''}:${control.channel.chidPort || ''}`;
      return {
        ...control,
        channel: {
          ...control.channel,
          chid: ipJoined,
          chidPort: undefined,
        },
      };
    },
    async saveToServer() {
      const controlForSave = this.normalizeForSave(this.control);
      if (this.control.id) {
        await axiosPut('/api/dcos/tdac-cgi/controller', controlForSave);
      } else {
        await getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/controller', controlForSave);
      }
      // 修改门的分组信息
      this.$message.success('保存成功');
    },
    async submit() {
      if (!(await this.$refs.form.validate())) return;

      await this.saveToServer();

      if (this.callback) {
        this.callback(this.control);
      }
      this.close();
    },
    handleRoomChange() {
      this.control.position.block = null;
    },
  },
};
</script>

<style lang="scss" scoped>
.flex-form-item-content {
  display: flex;
  gap: 32px;
  align-items: center;

  /deep/ {
    .el-input {
      margin: 12px 0;
    }
  }
}

.control-form-split-header-bar {
  padding: 24px 24px 16px 0 !important;
}
</style>
