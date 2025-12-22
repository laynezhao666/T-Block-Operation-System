<template>
  <div class="strategy-history-block">
    <div style="display:flex;height:48px;margin-bottom:10px">
      <el-title style="width:400px">
        <el-breadcrumb separator-class="el-icon-arrow-right">
          <el-breadcrumb-item
            :to="{ path: '/timpage/strategy-effective-verify' }"
          >
            <span
              style="color:#1470CC"
              @click="jumpUrl('/timpage/strategy-effective-verify')"
            >策略生效验证</span>
          </el-breadcrumb-item>
          <el-breadcrumb-item>查看无效策略历史数据</el-breadcrumb-item>
        </el-breadcrumb>
      </el-title>
    </div>
    <el-block
      no-padding
    >
      <common-table
        :columns="columns"
        :table-config="tableConfig"
        :config-cgi="configCgi"
        style="border-top:1px solid #f0f0f0"
      />
    </el-block>
  </div>
</template>

<script>
import commonTable from '../component/commonTable/ConfigPanel/index.vue';
import config from './config';
import { warning as cgi } from '@@/config/cgi';
import { eventBus } from '../component/commonTable/script/eventBus';
import { getQueryString } from 'common/script/utils';

export default {
  components: {
    commonTable,
  },
  data() {
    return {
      mozuId: 326,
      mozuloaded: false,
      showMoreDetail: false,
      activeName: 'invalid',
      columns: config,
      configCgi: {
        queryCgi: cgi.getHistoryList,
        exportCgi: cgi.exportInvalidHistory,
      },
      tableConfig: {
        showTableSelect: true,
        rights: 0b10100,
        // hasDetail: true,
        showSetting: false,
        showSearch: true,
        deleteCgi: cgi.deleteCustom,
        placeHolder: '搜索告警内容',
        refreshNow: false,
        searchParams: { mozuId: this.mozuId },
        searchNameMap: { lastRuntimeStart: 'occurTimeStart', lastRuntimeEnd: 'occurTimeEnd', deviceNumber: 'deviceGid', ruleType: 'isStandard', validateTypeStr: 'validateType' },

      },
      form: {},
      mozuName: [],
      options: [],
    };
  },
  mounted() {
    this.mozuId = parseInt(TNBL.getCurModuleId()) || parseInt(getQueryString('mozuId')) || 326; ;
    this.initPage();
    this.$set(this.tableConfig.searchParams, 'mozuId', this.mozuId);
  },
  beforeDestroy() {
    eventBus.$off('showModal');
  },
  methods: {
    initPage() {
      eventBus.$on('showModal', ({ type, data }) => {
        if (type === '详情') {
          window.open(`/timpage/warning-strategy-detail?validateId=${data.validateId}&mozuId=${this.mozuId}`);
        }
      });
    },
    jumpUrl(url) {
      window.open(url);
    },
  },
};
</script>
<style lang="scss">
  #mozu-cascader {
  margin-top:16px;
  float:right;
  width:389px;
  .el-input__inner {
    height: 32px;
    padding-left: 5px ;
  }
   .el-input::before{
      content:none;
    }
    .el-input::after{
      content:none;
    }
  }
.strategy-history-block{
  .el-icon-arrow-right {
    // content: "\e604";
    transform: translate(0px,5px);
  }
}
  </style>
