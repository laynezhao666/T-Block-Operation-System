<template>
  <div>
    <div
      v-if="tabFrom === '有效策略实例'"
      style="margin:20px 0 20px 25px;display:flex"
    >
      <el-button
        plain
        :class="triggerWarning ? 'plain-button' : ''"
        @click="hasWarning"
      >
        已触发告警（{{ extraData.fired }}）
      </el-button>
      <el-button
        plain
        :class="!triggerWarning ? 'plain-button' : ''"
        style="margin-left: 0px"
        @click="haventWarning"
      >
        未触发告警（{{ extraData.unfired }}）
      </el-button>
    </div>
    <commonTable
      :columns="columns"
      :table-config="tableConfig"
      :config-cgi="configCgi"
    />
  </div>
</template>

<script>
import commonTable from '../component/commonTable/ConfigPanel/index';

export default {
  components: {
    commonTable,
  },
  props: {
    extraData: {
      type: Object,
      default() {
        return {};
      },
    },
    tabName: {
      type: String,
      default: '',
    },
    mozuId: {
      type: Number,
      default: 326,
    },
    modalVisible: {
      type: Boolean,
      default: false,
    },
    showExtra: {
      type: Boolean,
      default: false,
    },
    refreshTable: {
      type: Boolean,
      default: false,
    },
    total: {
      type: Number,
      default: 0,
    },
    expireCount: {
      type: Number,
      default: 0,
    },
    buttonItems: {
      type: Array,
      default() {
        return [];
      },
    },
    config: {
      type: Array,
      default() {
        return {};
      },
    },
    configCgi: {
      type: Object,
      default() {
        return {};
      },
    },
  },
  // inject: ['configCgi'],
  provide() {
    return {
      tableConfig: this.tableConfig,
      configCgi: this.configCgi,
    };
  },

  data() {
    return {
      triggerWarning: true,
      columns: this.config,
      tableConfig: {
        showTableSelect: true,
        rights: 0b10100,
        showSetting: false,
        showSearch: true,
        placeHolder: '搜索告警内容',
        method: 'post',
        refreshNow: false,
        searchParams: { mozuId: this.mozuId },
        searchNameMap: { lastRuntimeStart: 'occurTimeStart', lastRuntimeEnd: 'occurTimeEnd', deviceNumber: 'deviceGid', ruleType: 'isStandard', validateTypeStr: 'validateType' },
      },
      codes: {},
      tabFrom: this.tabName,

    };
  },
  watch: {
    refreshTable() {
      this.refresh();
    },
    mozuId(val) {
      this.$set(this.tableConfig.searchParams, 'mozuId', val);
      this.tableConfig.refreshNow = !this.tableConfig.refreshNow;
    },
  },

  beforeMount() {
    if (this.tabName === '有效策略实例') {
      this.$set(this.tableConfig.searchParams, 'fired', true);
    }
  },
  mounted() {
    console.log(this.configCgi, 'this.configCgi');
  },

  methods: {
    hasWarning() {
      this.triggerWarning = true;
      this.$set(this.tableConfig.searchParams, 'fired', true);
      // this.tableConfig.refreshNow = !this.tableConfig.refreshNow;
      this.$emit('updatedata');
    },
    haventWarning() {
      this.triggerWarning = false;
      this.$set(this.tableConfig.searchParams, 'fired', false);
      this.$emit('updatedata');
      // this.tableConfig.refreshNow = !this.tableConfig.refreshNow;
    },
    refresh() {
      this.tableConfig.refreshNow = !this.tableConfig.refreshNow;
    },
  },
};
</script>

<style lang="scss" scoped>
  .plain-button {
    color: #1470CC;
    border-color: #1470CC !important;
  }
</style>
