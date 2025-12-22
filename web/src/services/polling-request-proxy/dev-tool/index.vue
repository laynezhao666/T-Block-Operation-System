<template>
  <el-modal
    :visible.sync="modalVisible"
    :width="1080"
    @close="handleClose"
  >
    <template #title>
      轮询代理调试工具【{{ mode }}】

      <!-- <div class="dev-switcher">
        启用：
        <el-switch
          v-model="enableDevMode"
          size="small"
        />
      </div> -->

      <el-button
        type="text"
        @click="openServerSideDevTool"
      >
        服务端实例
      </el-button>
    </template>

    <el-table
      :data="tableDataGroupedByServerId"
      row-key="serverId"
    >
      <el-table-column
        type="expand"
      >
        <template #default="{ row }">
          <div class="two-cols">
            <pre v-html="row.request" />
            <pre v-html="row.devData.data" />
          </div>

          <el-divider v-if="row.plugins">
            插件配置：
          </el-divider>
          <pre
            v-if="row.plugins"
            v-html="row.plugins"
          />
        </template>
      </el-table-column>

      <el-table-column
        label="服务端ID"
        prop="serverId"
        width="180px"
      />

      <el-table-column
        label="客户端端ID"
        prop="clientIds"
        width="120px"
      />

      <el-table-column
        label="请求配置"
        prop="request.url"
      >
        <template #default="{ row }">
          【{{ row.request.method }}】{{ row.request.url }}
        </template>
      </el-table-column>

      <el-table-column
        label="请求间隔"
        prop="interval"
        width="120px"
      >
        <template #default="{ row }">
          {{ row.interval }}毫秒
        </template>
      </el-table-column>

      <el-table-column
        label="更新时间"
        prop="updatedAt"
        width="130px"
      />
    </el-table>

    <server-side-proxies-modal
      ref="serverSideModal"
    />
  </el-modal>
</template>

<script>
import highlightJs from 'highlight.js/lib/core';
import javascript from 'highlight.js/lib/languages/javascript';
import json from 'highlight.js/lib/languages/json';
import 'highlight.js/styles/atom-one-light.css';
import dayjs from 'dayjs';
import ServerSideProxiesModal from './server-side-proxies-modal.vue';

highlightJs.registerLanguage('javascript', javascript);
highlightJs.registerLanguage('json', json);

const MAX_SHOW_JSON_LENGTH = 1e5;

const formatObject = (obj) => {
  const json = JSON.stringify(obj, 4, 4);
  if (json.length > MAX_SHOW_JSON_LENGTH) return `${json.substring(0, MAX_SHOW_JSON_LENGTH)}...`;
  return highlightJs.highlight(json, {
    language: 'json',
  }).value;
};

export default {
  components: {
    ServerSideProxiesModal,
  },
  data() {
    return {
      modalVisible: false,
      tableDataGroupedByServerId: [],
      mode: '',
    };
  },
  mounted() {
    let clickTimes = 0;
    let clickTimeout;

    let locked = false;

    window.addEventListener('click', (evt) => {
      if (locked) {
        evt.preventDefault();
        evt.stopImmediatePropagation();
        evt.stopPropagation();
        return;
      }
      if (!evt.altKey) return;

      clickTimes += 1;

      if (clickTimes >= 3) {
        this.open();
        clickTimes = 0;
        locked = true;
        setTimeout(() => {
          locked = false;
        }, 300);
      }

      if (!clickTimeout) {
        clickTimeout = setTimeout(() => {
          clickTimes = 0;
          clickTimeout = undefined;
        }, 10000);
      }
    }, true);
  },
  methods: {
    open() {
      const { pollingProxyAgentService } = window.tnwebServices;
      this.modalVisible = true;
      pollingProxyAgentService.toggleDevMode(true);
      pollingProxyAgentService.onUpdateDevTool = () => {
        this.updateTableDataGroupedByServerId();
      };
      this.mode = pollingProxyAgentService.runningMode;
    },
    updateTableDataGroupedByServerId() {
      if (!this.modalVisible) return;

      const { pollingProxyAgentService } = window.tnwebServices;
      const { serverIdToProxyMap } = pollingProxyAgentService;
      const list = [];

      if (serverIdToProxyMap) {
        serverIdToProxyMap.forEach((proxiesMap, serverId) => {
          const clientIds = [];
          let request;
          let interval;
          let devData;
          let updatedAt;
          let plugins;

          if (!proxiesMap.size) return;

          proxiesMap.forEach((proxy) => {
            clientIds.push(proxy.config.clientProxyId);
            // eslint-disable-next-line prefer-destructuring
            request = proxy.config.request;
            // eslint-disable-next-line prefer-destructuring
            interval = proxy.config.interval;
            // eslint-disable-next-line prefer-destructuring
            devData = proxy.devData;
            // eslint-disable-next-line prefer-destructuring
            updatedAt = proxy.updatedAt;
            // eslint-disable-next-line prefer-destructuring
            plugins = proxy.config.plugins;
          });

          list.push({
            serverId,
            clientIds: clientIds.join(','),
            request: {
              ...request,
              data: formatObject(request.data),
            },
            plugins: plugins && formatObject(plugins),
            devData: {
              ...(devData || {}),
              data: devData?.data && formatObject(devData.data),
            },
            interval,
            updatedAt: dayjs(updatedAt).format('HH:mm:ss.SSS'),
          });
        });
      }

      this.tableDataGroupedByServerId = list;

      this.mode = pollingProxyAgentService.runningMode;
    },
    initServerProxiesListWatcher() {
      const {
        pollingProxyAgentService: {
          socket,
        },
      } = window.tnwebServices;
      const MSG_TYPE = 'serverPollingProxyList';
      socket.emit(MSG_TYPE);
      socket.on(MSG_TYPE, (evt) => {
        console.log(evt);
      });
    },
    openServerSideDevTool() {
      this.$refs.serverSideModal.open();
    },
    handleClose() {
      const { pollingProxyAgentService } = window.tnwebServices;
      pollingProxyAgentService.toggleDevMode(false);
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
