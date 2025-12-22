
<template>
  <div class="grid">
    <div class="grid-tabs">
      <el-button
        type="icon"
        :icon="`tn-icon-arrow-${treeCollapsed ? 'right' : 'left'}`"
        @click="toggleTreeVisible"
      />
      <el-tabs v-model="activeName">
        <el-tab-pane
          label="基本信息"
          name="base"
        />
        <el-tab-pane
          label="测点信息"
          name="point"
        />
        <el-tab-pane
          v-if="showBackup"
          label="备份还原"
          name="backup"
        />
        <!-- <el-tab-pane label="端口调试" name="port" /> -->
      </el-tabs>
    </div>
    <div
      class="grid-extra-button"
      style="border-bottom: solid 1px #f0f0f0"
    >
      <el-input
        v-show="activeName === 'point'"
        v-model="filterText"
        class="c-tree-tools__input"
        placeholder="请输入测点标识符或名称"
        suffix-icon="tn-icon-search"
        clearable
        border-type="bordered"
      />
    </div>
    <div class="grid-content">
      <base-info-new
        v-show="activeName === 'base'"
        :device-status="deviceStatus"
        :collector="collector"
      />
      <point-info
        v-show="activeName === 'point'"
        ref="point"
        :visible="activeName === 'point'"
        :collector="collector"
        @edit="handleEdit"
      />
      <back-up
        v-show="activeName === 'backup'"
        :visible="activeName === 'backup'"
        :collector="collector"
      />
    </div>
  </div>
</template>

<script>
import PointInfo from './pointInfo';
import BaseInfoNew from './baseInfo-new';
import BackUp from './backup';
export default {
  components: {
    PointInfo,
    BaseInfoNew,
    BackUp,
  },
  props: {
    collector: {
      type: Object,
      default: null,
    },
    deviceStatus: {
      type: Object,
      default: null,
    },
  },
  data() {
    return {
      treeCollapsed: false,
      activeName: 'base',
      filterText: '',
      showBackup: false,
    };
  },
  computed: {
    collectorId() {
      return this.collector?.id || null;
    },
  },
  watch: {
    filterText(v) {
      this.$refs.point.filterData(v);
    },
    collector(v) {
      if (!v) return;
      // 自研采集器设备
      if (v.type === 'collector' && v.link_type === 'TLINK') {
        this.showBackup = true;
      } else {
        this.showBackup = false;
        this.activeName = this.activeName === 'backup' ? 'base' : this.activeName;
      }
    },
  },
  methods: {
    toggleTreeVisible() {
      this.$emit('tree-visible-change', this.treeCollapsed);
      this.treeCollapsed = !this.treeCollapsed;
    },
    handleEdit(id) {
      this.$emit('edit', id);
    },
  },
};
</script>

<style lang="scss" scoped>
.grid {
  height: 100%;
  display: grid;
  grid-template-areas:
    "tabs extra"
    "table table";
  grid-template-columns: minmax(0, 1fr) 320px;
  grid-template-rows: auto auto 1fr;
  &-tabs {
    grid-area: tabs;
    display: flex;
    border-bottom: solid 1px #f0f0f0;
    padding-right: 12px;
    .el-tabs {
      flex: 1;
    }
  }
  &-content {
    grid-area: table;
  }
  &-extra-button {
    grid-area: extra;
    display: flex;
    justify-content: right;
    padding-right: 16px;
    align-items: center;
    > button {
      height: 32px;
      border-radius: 4px;
    }
    .el-button--primary {
      width: 100px;
    }
  }
}
/deep/ .el-input .el-input__inner {
  border-radius: 4px;
}
</style>

<!-- <div v-show="activeName === 'point'">
        <el-button type="text">
          批量移除
        </el-button>
        <el-button type="text">
          批量导出
        </el-button>
        <el-button type="text">
          导入
        </el-button>
        <el-button
          type="primary"
          icon="tn-icon-add"
        >
          批量添加
        </el-button>
      </div> -->
