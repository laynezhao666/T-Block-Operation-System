
<template>
  <admin-limit-content
    ref="container"
  >
    <el-block
      no-padding
      :style="{ height: `${pageHeight}px`, overflow: 'hidden' }"
    >
      <div
        class="device-manage"
        :style="{ height: `${pageHeight}px` }"
      >
        <transition name="tree">
          <el-tabs
            v-show="treeVisible"
            v-model="menuTabsName"
            class="device-manage-list"
            style="width: 350px"
          >
            <el-tab-pane
              label="采集设备"
              name="collector"
            >
              <!-- 设备树 -->
              <device-tree
                ref="tree"
                class="data-query-tree"
                :height="pageHeight"
                @updateStatus="updateStatus"
                @checkDevice="handleCheck"
                @edit="handleEdit"
              />
            </el-tab-pane>
          </el-tabs>
        </transition>

        <device-info
          v-if="currentCollector && currentCollector.type !== 'mozu'"
          class="device-manage-info"
          :height="pageHeight"
          :collector="currentCollector"
          :device-status="deviceStatus"
          @tree-visible-change="toggleTreeVisible"
          @edit="handleEdit"
        />

        <mozu-info
          v-if="currentCollector && currentCollector.type === 'mozu'"
          :mozu-info="currentCollector"
        />
      </div>
      <add-form
        v-if="addVisible"
        :visible.sync="addVisible"
        :info-form="collectorInfo"
        :device="editDevice"
        @confirm="focusConfirm"
      />
    </el-block>
  </admin-limit-content>
</template>

<script>
import deviceInfo from './views/device-info';
import addForm from './components/add';
import deviceTree from './views/device-tree';
import AdminLimitContent from 'feature/component/tedge-components/admin-limit-content.vue';
import MozuInfo from './views/mozu-info';

export default {
  components: {
    MozuInfo,
    deviceInfo,
    addForm,
    deviceTree,
    AdminLimitContent,
  },
  data() {
    return {
      menuTabsName: 'collector',
      treeVisible: true,
      addVisible: false,
      collectorInfo: null,
      pageHeight: 0,
      editDevice: null,
      currentCollector: null,
      deviceStatus: {},
    };
  },
  mounted() {
    this.pageHeight = window.innerHeight - this.$refs.container.$el.offsetTop - 16;
    this.$nextTick(() => {
      this.observer = new ResizeObserver(() => {
        this.pageHeight = Math.max(window.innerHeight - this.$refs.container.$el.offsetTop - 16, 800);
      });
      this.observer.observe(this.$refs.container.$el);
    });
  },
  beforeDestroy() {
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
  },
  methods: {
    updateStatus(data) {
      this.deviceStatus = data;
    },
    handleCheck(data) {
      this.currentCollector = data;
    },
    toggleTreeVisible(val) {
      this.treeVisible = val;
    },

    // 二期迭代
    handleCommand(command) {
      this.$message(`TODO ${command}`);
    },
    addCollector() {
      this.collectorInfo = null;
      this.addVisible = true;
      this.editDeviceId = null;
    },
    focusConfirm() {
      // todo 请求接口更新修改数据
      // 修改成功之后刷新数据
      this.$refs.tree.refreshCollector();
    },
    handleEdit(data) {
      // todo 根据采集器ID请求相关设备数据，根据数据更新
      this.collectorInfo = {
        model: '', // 采集模板
        name: '', // TBOX名称
        type: '', // TBOX 类别
        ipAddress: '', // TBOX地址 IP
        port: '', // TBOX地址 port
        roomId: '', // 安装位置：房间
        itmId: '', // 安装位置：方仓
        account: '', // 账号
        password: '', // 密码
      };
      this.editDevice = data;
      this.addVisible = true;
    },
  },
};
</script>

<style lang="scss" scoped>
.device-manage {
  display: flex;
  height: 100%;
  &-list {
    //
  }
  &-info {
    flex: 1;
    border-left: 1px solid #f0f0f0;
  }
  .search-wrap {
    // border-bottom: solid 1px #c0c0c0;
    display: flex;
    align-items: center;
    padding: 5px 10px 5px 5px;
    margin: 5px 0;
    flex: 1;
    .el-select {
      width: 130px;
      margin-right: 10px;
    }
    /deep/ .el-select .el-input__inner {
      border-radius: 2px;
    }
    .el-input {
      margin-right: 5px;
    }
    /deep/ .el-input .el-input__inner {
      border-radius: 2px;
    }
    .add {
      padding: 0px 5px;
    }
    .more {
      padding: 0px 5px;
      cursor: pointer;
    }
  }
  /deep/ .el-tab-pane {
    .data-query-tree {
      padding: 0;
      overflow: hidden;
      // overflow: overlay;
      &:hover {
        overflow: auto;
      }
    }
  }
}
</style>
