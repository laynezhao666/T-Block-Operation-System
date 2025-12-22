<template>
  <div class="mozu-info">
    <el-tabs
      v-model="activeTabKey"
    >
      <el-tab-pane
        label="当前统计"
        name="current"
      />

      <el-tab-pane
        label="历史统计"
        name="history"
      />
    </el-tabs>

    <div
      v-for="(tableData, i) in currentTablesData"
      :key="i"
    >
      <div style="margin-top: 24px;" />

      <split-header-bar
        :title="tableData.tableName"
      />

      <table
        class="info-table"
      >
        <tr>
          <th
            v-for="(col, j) in tableData.columns"
            :key="j"
          >
            {{ col.title }}
          </th>
        </tr>

        <tr>
          <td
            v-for="(col, j) in tableData.columns"
            :key="j"
          >
            {{ col.content }}台
          </td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script>
import SplitHeaderBar from '@/feature/component/tedge-components/split-header-bar.vue';

export default {
  components: {
    SplitHeaderBar,
  },
  props: {
    mozuInfo: {
      type: Object,
      required: true,
    },
  },
  data() {
    this.loadSummaryData = this.$intervalFunction(this.loadSummaryData, 5000, true);

    return {
      activeTabKey: 'current',
      summaryData: {
        current: {},
        history: {},
      },

      tablesConfigs: {
        current: [{
          tableName: '采集器',
          columns: [{
            title: 'TBOX总数',
            field: 'collector.total',
          }, {
            title: '在线',
            field: 'collector.online',
          }, {
            title: '离线',
            field: 'collector.offline',
          }],
        }, {
          tableName: '采集设备',
          columns: [{
            title: '采集设备总数',
            field: 'device.total',
          }, {
            title: '在线',
            field: 'device.online',
          }, {
            title: '离线',
            field: 'device.offline',
          }],
        }],
        history: [{
          tableName: '采集器',
          columns: [{
            title: 'TBOX总数',
            field: 'collector.total',
          }, {
            title: '通讯成功过',
            field: 'collector.succeed',
          }, {
            title: '未通讯成功过',
            field: 'collector.failed',
          }],
        }, {
          tableName: '采集设备',
          columns: [{
            title: '采集设备总数',
            field: 'device.total',
          }, {
            title: '通讯成功过',
            field: 'device.succeed',
          }, {
            title: '未通讯成功过',
            field: 'device.failed',
          }],
        }],
      },
    };
  },
  computed: {
    currentTabData() {
      return this.summaryData[this.activeTabKey] || {};
    },
    currentTablesData() {
      const {
        tablesConfigs,
        activeTabKey,
        currentTabData,
      } = this;
      return _.map(tablesConfigs[activeTabKey] || [], tableConfig => ({
        ...tableConfig,
        columns: tableConfig.columns.map(col => ({
          ...col,
          content: _.get(currentTabData, col.field),
        })),
      }));
    },
  },
  watch: {
    activeTabKey() {
      this.loadSummaryData();
    },
  },
  mounted() {
    this.loadSummaryData();
  },
  methods: {
    async loadSummaryData() {
      const url = {
        current: '/api/dcos/tboxmonitor-cgi/comm/rt',
        history: '/api/dcos/tboxmonitor-cgi/comm/nrt',
      }[this.activeTabKey];

      const data = await this.$axios.get(url, null, false);

      if (JSON.stringify(data) === JSON.stringify(this.summaryData[this.activeTabKey])) return;

      this.summaryData[this.activeTabKey] = data;
    },
  },
};
</script>

<style lang="scss" scoped>
.mozu-info {
  border-left: 1px solid #efefef;
  flex: 1;
}

.info-table {
  margin: auto;

  &, th, td {
    border: 1px solid #a0a0a0;
    border-collapse: collapse;
    text-align: center;
  }

  th, td {
    padding: 8px 12px;
    width: 8em;
  }

  td {
    font-weight: 600;
  }
}
</style>
