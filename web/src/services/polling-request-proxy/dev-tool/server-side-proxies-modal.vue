<template>
  <el-modal
    :visible.sync="modalVisible"
    :width="1080"
    @click="handleClose"
  >
    <template #title>
      轮询代理调试工具--服务端代理实例
    </template>

    <el-table
      :data="tableData"
      row-key="id"
    >
      <el-table-column
        type="expand"
      >
        <template #default="{ row }">
          <div class="two-cols">
            <pre v-html="row.config.request.data" />
            <div>
              <el-divider v-if="row.config.plugins">
                插件配置：
              </el-divider>
              <pre
                v-if="row.config.plugins"
                v-html="row.config.plugins"
              />
            </div>
          </div>
        </template>
      </el-table-column>

      <el-table-column
        label="服务端ID"
        prop="id"
        width="180px"
      />

      <el-table-column
        label="客户端端ID"
        prop="config.clientProxyIds"
        width="120px"
      />

      <el-table-column
        label="请求配置"
        prop="config.request.url"
      >
        <template #default="{ row }">
          【{{ row.config.request.method }}】{{ row.config.request.url }}
        </template>
      </el-table-column>

      <el-table-column
        label="请求间隔"
        prop="interval"
        width="120px"
      >
        <template #default="{ row }">
          {{ row.config.interval }}毫秒
        </template>
      </el-table-column>

      <el-table-column
        label="状态"
        prop="status"
        width="120px"
      />

      <el-table-column
        label="客户端数"
        prop="clientCount"
        width="120px"
      />

      <el-table-column
        label="创建时间"
        prop="createTime"
        width="120px"
      />
    </el-table>
  </el-modal>
</template>

<script>
import highlightJs from 'highlight.js/lib/core';
import javascript from 'highlight.js/lib/languages/javascript';
import json from 'highlight.js/lib/languages/json';
import 'highlight.js/styles/atom-one-light.css';

highlightJs.registerLanguage('javascript', javascript);
highlightJs.registerLanguage('json', json);

const MSG_TYPE = 'serverPollingProxyList';

export default {
  data() {
    return {
      modalVisible: false,
      tableData: [],
    };
  },
  beforeDestroy() {
    this.offWatcher();
  },
  methods: {
    open() {
      this.modalVisible = true;
      this.initServerProxiesListWatcher();
    },
    initServerProxiesListWatcher() {
      const {
        pollingProxyAgentService: {
          socket,
        },
      } = window.tnwebServices;

      socket.on(MSG_TYPE, this.handleMsg);

      this.timer = setInterval(() => {
        socket.emit(MSG_TYPE);
      }, 1000);
    },
    offWatcher() {
      const {
        pollingProxyAgentService: {
          socket,
        },
      } = window.tnwebServices;

      socket.off(MSG_TYPE, this.handleMsg);

      if (this.timer) {
        clearInterval(this.timer);
      }
    },
    handleMsg(data) {
      this.tableData = data;
    },
    handleClose() {
      this.offWatcher();
    },
  },
};
</script>

<style lang="scss" scoped>
.two-cols {
  display: flex;
  padding-left: 36px;

  & > * {
    flex: 1;
    overflow: auto;
    padding: 8px;
    max-height: 200px;

    &:first-child {
      border-right: 1px solid #e0e0e0;
    }
  }
}

.dev-switcher {
  position: absolute;
  right: 24px;

  font-size: 12px;
  top: 24px;
}
</style>
