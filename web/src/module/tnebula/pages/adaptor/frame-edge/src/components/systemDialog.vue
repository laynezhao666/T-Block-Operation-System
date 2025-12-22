<template>
  <el-dialog
    :visible.sync="dialogVisible"
    width="400px"
    title="关于系统"
  >
    <div
      v-for="(group, i) in infoGroups"
      :key="i"
      class="about-system-group"
    >
      <split-header-bar
        :title="group.name"
        no-padding
        class="header-bar"
      />
      <div
        v-for="item in group.info"
        :key="item.label"
        class="form"
      >
        <span class="label">{{ item.label }}:</span>
        <span class="value">{{ item.key }}</span>
      </div>
    </div>
  </el-dialog>
</template>

<script>
import SplitHeaderBar from 'feature/component/tedge-components/split-header-bar.vue';
import { latestVersion } from '../const/change-log';

export default {
  components: {
    SplitHeaderBar,
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      systemInfo: [
        { key: 'TB-2.0.5', label: '模组架构版本' },
        { key: latestVersion, label: '软件版本' },
        { key: 'V2.0.1', label: '北向版本' },
        { key: 'V2.3', label: '告警策略版本' },
      ],
      configInfoList: [],
    };
  },
  computed: {
    dialogVisible: {
      set(v) {
        this.$emit('update:visible', v);
      },
      get() {
        return this.visible;
      },
    },
    infoGroups() {
      return [{
        name: '版本信息',
        info: this.systemInfo,
      }, ...this.configInfoList];
    },
  },
  created() {
    this.loadContactInfo();
  },
  methods: {
    async loadContactInfo() {
      this.configInfoList = (await window.tnwebServices.customConfigService.loadConfig('about_system_info')) || [];
    },
  },
};
</script>

<style lang="scss" scoped>
.form {
  padding-bottom: 15px;
  .label {
    display: inline-block;
    width: 100px;
    text-align: right;
    margin-right: 10px;
    font-weight: 600;
    color: #999;
  }

  .value {
    color: #333;
  }
}
/deep/ .el-dialog {
  border-radius: 5px;
}
/deep/ .el-dialog__header {
  background: #c4b1b12e;
  padding: 15px 24px;
}

.about-system-group {
  margin-top: 24px;

  &:first-child {
    margin-top: 0;
  }

  /deep/ .header-bar  {
    padding-bottom: 16px;
  }
}
</style>
