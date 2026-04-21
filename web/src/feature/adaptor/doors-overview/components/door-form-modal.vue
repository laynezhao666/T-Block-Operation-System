<template>
  <el-modal :visible.sync="visible">
    <template slot="title">
      设置 门参数
    </template>

    <el-form
      v-if="door"
      ref="form"
      :model="door"
      :rules="rules"
      label-width="140px"
    >
      <el-form-item
        v-if="!isBatch"
        label="门编号"
        prop="number"
      >
        <el-input
          v-model.trim="door.number"
          disabled
          placeholder="请输入门编号"
        />
      </el-form-item>

      <el-form-item
        v-if="!isBatch"
        label="门名称"
        prop="name"
      >
        <el-input
          v-model.trim="door.name"
          placeholder="请输入门名称"
        />
      </el-form-item>

      <el-form-item
        v-if="!isBatch"
        label="关联设备"
        prop="idcdbCode"
      >
        <el-select
          v-model="door.idcdbCode"
          allow-create
          filterable
          clearable
          placeholder="关联idcdb中门对象，支持检索"
        >
          <el-option
            v-for="(idcdbDoor, i) in idcdbDoorOptions"
            :key="i"
            :label="idcdbDoor.device_number"
            :value="idcdbDoor.device_number"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        label="门密码"
        prop="password"
      >
        <el-input
          v-model.trim="door.password"
          placeholder="请输入密码"
          type="password"
          show-password
        />
      </el-form-item>

      <el-form-item
        label="门密码确认"
        prop="passwordConfirm"
      >
        <el-input
          v-model.trim="door.passwordConfirm"
          placeholder="请再次输入密码"
          type="password"
          show-password
        />
      </el-form-item>

      <el-form-item
        label="门开保持时间"
        prop="keep_open_timeout"
      >
        <template slot="label">
          门开保持时间
          <el-help-tip width="200">
            单位:秒， 缺省值:5， 可设置范围：1~600
          </el-help-tip>
        </template>

        <el-input
          v-model.number="door.keep_open_timeout"
          placeholder="请输入门开保持时间"
          type="number"
          class="short-input-1"
        >
          <span slot="suffix">
            秒
          </span>
        </el-input>
      </el-form-item>

      <el-form-item
        label="门开超时时间"
        prop="open_timeout"
        help="123"
      >
        <template slot="label">
          门开超时时间
          <el-help-tip width="200">
            单位:秒， 缺省值:15， 可设置范围：1~600
          </el-help-tip>
        </template>

        <el-input
          v-model.number="door.open_timeout"
          placeholder="请输入超时时间"
          type="number"
          class="short-input-1"
        >
          <span slot="suffix">
            秒
          </span>
        </el-input>
      </el-form-item>

      <el-form-item
        label="非法卡刷卡间隔"
        prop="verifv_interval"
      >
        <template slot="label">
          非法卡刷卡间隔
          <el-help-tip width="200">
            单位:秒， 缺省值:60， 可设置范围：3~120
          </el-help-tip>
        </template>

        <el-input
          v-model.number="door.verifv_interval"
          placeholder="请输入非法卡刷卡间隔"
          type="number"
          class="short-input-1"
        >
          <span slot="suffix">
            秒
          </span>
        </el-input>
      </el-form-item>

      <!-- <el-form-item
        label="非法卡刷卡间隔"
        prop="lock_time"
      >
        <template slot="label">
          非法卡刷卡间隔
          <el-help-tip width="200">
            单位:秒， 缺省值:60， 可设置范围：3~120
          </el-help-tip>
        </template>

        <el-input
          v-model.number="door.lock_time"
          placeholder="请输入非法卡刷卡间隔"
          type="number"
          class="short-input-1"
        >
          <span slot="suffix">
            秒
          </span>
        </el-input>
      </el-form-item> -->

      <el-form-item
        label="卡封锁错误次数"
        prop="lock_count"
      >
        <template slot="label">
          卡封锁错误次数
          <el-help-tip width="200">
            连续刷多少次非法卡后门封卡， 缺省值:5， 可设置范围：3~100
          </el-help-tip>
        </template>

        <el-input
          v-model.number="door.lock_count"
          placeholder="请输入卡封锁错误次数"
          type="number"
          class="short-input-1"
        >
          <span slot="suffix">
            次
          </span>
        </el-input>
      </el-form-item>

      <el-form-item
        label="验证方式"
        prop="open_mode"
      >
        <el-select
          v-model="door.open_mode"
          placeholder="请选择验证方式"
        >
          <el-option
            v-for="(opt, i) in open_modeOptions"
            :key="i"
            :value="opt.value"
            :label="opt.label"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        label="火警信号"
        prop="fire_signal_mode"
      >
        <el-radio-group
          v-model="door.fire_signal_mode"
          size="small"
        >
          <el-radio
            v-for="opt in fireAlarmSignOptions"
            :key="opt.value"
            :label="opt.value"
          >
            {{ opt.label }}
          </el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item
        v-if="!isBatch"
        label="关联摄像头"
        prop="relatedCameras"
      >
        <el-select
          v-model.trim="door.relatedCameras"
          allow-create
          multiple
          filterable
          placeholder="关联idcdb中门对象，支持检索"
        >
          <el-option
            v-for="(idcdbDoor, i) in idcdbCameraOptions"
            :key="i"
            :label="idcdbDoor.device_number"
            :value="idcdbDoor.device_number"
          />
        </el-select>
      </el-form-item>
    </el-form>

    <template slot="footer">
      <el-button
        type="primary"
        @click="submit"
      >
        提交
      </el-button>

      <el-button @click="close">
        取消
      </el-button>
    </template>
  </el-modal>
</template>

<script>
import getEdgeRequest from 'feature/utils/request';

export default {
  data() {
    return {
      door: null,
      callback: null,
      isBatch: false,

      initedOptions: false,

      allDoors: [],

      idcdbDoorOptions: [],
      idcdbCameraOptions: [],

      open_modeOptions: [
        {
          value: 0,
          label: '刷卡',
        },
        {
          value: 1,
          label: '密码',
        },
        {
          value: 2,
          label: '卡+密码',
        },
        {
          value: 3,
          label: '卡或密码',
        },
      ],

      fireAlarmSignOptions: [
        {
          value: 0,
          label: '短路有效',
        },
        {
          value: 1,
          label: '断路有效',
        },
      ],

      rules: {
        number: [{ required: true, message: '门编号不能为空' }],
        name: [{ required: true, message: '门名称不能为空' }],
        password: [{ min: 6, max: 6, message: '长度必须为6' }],
        passwordConfirm: [
          {
            validator: (rule, value, cb) => {
              cb(value === this.door.password
                ? undefined
                : '两次输入的密码不一致');
            },
          },
        ],
        keep_open_timeout: [
          { type: 'number', min: 1, max: 600, message: '可设置范围为1~600' },
        ],
        open_timeout: [
          { type: 'number', min: 1, max: 600, message: '可设置范围为1~600' },
        ],
        verifv_interval: [
          { type: 'number', min: 3, max: 120, message: '可设置范围为3~120' },
        ],
        lock_time: [
          { type: 'number', min: 60, max: 3600, message: '可设置范围为60~3600' },
        ],
        lock_count: [
          { type: 'number', min: 3, max: 600, message: '可设置范围为3~100' },
        ],
      },
    };
  },
  computed: {
    visible: {
      get() {
        return Boolean(this.door);
      },
      set(visible) {
        if (!visible) {
          this.close();
        }
      },
    },
    isCreate() {
      return this.door?.id;
    },
  },
  methods: {
    async loadOptions() {
      const {
        list,
      } = await getEdgeRequest(this.$axios).post('/cgi/dataQuery/edge/getEdgeDevices', {
        table: 'device',
        conditions: [{
          name: 'devicetypes_enAbbreviation',
          value: ['CAM', 'GSM'],
        }],
      });

      const grouped = _.groupBy(list, 'devicetypes_enAbbreviation');

      this.idcdbCameraOptions = grouped.CAM;
      this.idcdbDoorOptions = grouped.GSM;
    },
    async loadAllDoors() {
      const doorList = await getEdgeRequest(this.$axios).get('/api/dcos/tdac-cgi/controllers');

      this.allDoors = _.chain(doorList).map('doors')
        .flatten()
        .value();
    },
    async loadDoorParameters(doorId) {
      const doorInfo = await getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/door', {
        id: doorId,
      });

      Object.assign(this.door, _.omit(doorInfo.parameters, ['password', 'name']));
      this.$set(this.door, 'idcdbCode', doorInfo.code || '');
      this.$set(this.door, 'relatedCameras', doorInfo.extend?.relatedCameras || []);
    },
    transformFilter(query, item) {
      if (!query?.trim()) return true;
      return item.name.includes(query.tirm());
    },
    edit(door, callback, isBatch) {
      this.door = {
        password: '******',
        passwordConfirm: '******',
        keep_open_timeout: 5,
        open_mode: 0,
        open_timeout: 15,
        lock_time: 300,
        verifv_interval: 60,
        lock_count: 5,
        fire_signal_mode: 0,
        ...door,
      };
      this.callback = callback;
      this.isBatch = isBatch;

      this.loadAllDoors();

      if (!this.initedOptions) {
        this.loadOptions().then(() => {
          this.initedOptions = true;
        });
      }

      if (door.id) {
        this.loadDoorParameters(door.id);
      }
    },
    close() {
      this.door = null;
      this.callback = null;
    },
    async submit() {
      if (!(await this.$refs.form.validate())) return;

      const preHookResult = (await this.callback?.(this.door, 'pre')) ?? true;

      if (preHookResult === false) return;
      if (preHookResult === 'close') {
        this.close();
        return;
      }

      if (this.callback) {
        // 由上层自行保存
        this.callback(this.door, 'post');
      }
      this.close();
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

.door-form-split-header-bar {
  padding: 24px 24px 16px 0 !important;
}

.short-input-1 {
  width: 6em;
}
</style>
