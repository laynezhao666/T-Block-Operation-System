<template>
  <div>
    <el-title
      v-if="showTitle"
      style="width:200px"
    >
      历史告警
      <el-help-tip icon="tn-icon-question">
        <div style="line-height: 40px">
          <p>当前告警：告警未恢复且未挂起</p>
          <p>挂起告警：告警未恢复且已挂起</p>
          <p v-if="!business.isTedge">
            活动告警：告警未恢复且未挂起且未转单
          </p>
          <p v-if="!business.isTedge">
            转单告警：告警未恢复且未挂起且已转单且未结单
          </p>
          <p>历史告警：告警已恢复</p>
          <p v-if="!business.isTedge">
            当前告警 = 活动告警 + 转单告警
          </p>
        </div>
      </el-help-tip>
    </el-title>
    <el-block>
      <common-table
        :columns="columns"
        :table-config="tableConfig"
        :config-cgi="cgi"
      >
        <template slot="extraButtons">
          <el-button
            v-if="business.isTedge"
            type="primary"
            @click="toggleAnalysis"
          >
            告警统计
          </el-button>
        </template>
      </common-table>
    </el-block>
    <analysis-modal
      v-if="handleVisible"
      :visible.sync="handleVisible"
      :data="modalData"
      @toggleAnalysisDownload="toggleAnalysisDownload"
    />
    <detail-modal
      v-if="detailVisible"
      :visible.sync="detailVisible"
      :data="detailData"
    />
  </div>
</template>

<script>
import business from '@@/config/business';
import commonTable from '../component/commonTable/ConfigPanel/index.vue';
import config from './config';
import tedgeConfig from './tedgeConfig';
import { warning as cgi } from '@@/config/cgi';
import { eventBus } from '../component/commonTable/script/eventBus';
import { getQueryString } from 'common/script/utils';
import { getMozuId } from '../../utils/business';
import mixin from 'feature/utils/mixin';
import analysisModal from './analysis-modal.vue';
import detailModal from './detail-modal.vue';

export default {
  components: {
    commonTable,
    analysisModal,
    detailModal,
  },
  mixins: [mixin],
  data() {
    return {
      detailData: [],
      detailVisible: false,
      modalData: {},
      handleVisible: false,
      business,
      mozuId: 326,
      mozuloaded: false,
      mozuName: [],
      options: [],
      columns: business.isTedge ? tedgeConfig : config,
      overview: {},
      modalVisible: false,
      loading: false,
      count: 0,
      current: 1,
      limit: 10,
      searchValue: '',
      cgi: {
        queryCgi: cgi.getWarningHistory,
        exportCgi: cgi.exportWarningHistory,
        detailUrl: '/tedge/warning-detail',
        extraCgi: '/cgi/alarm/history/exportStat',
      },
      tableConfig: {
        rights: 0b10000,
        showSetting: false,
        showSearch: true,
        refreshNow: false,
        analysis: true,
        placeHolder: '搜索告警内容',
        occurTime: 'warning',
        datePickerClearable: false,
        searchParams: { mozuId: 326 },
        searchNameMap: { mozuName: 'mozuId', roomName: 'roomId', deviceNumber: 'deviceGid', closeOperatorName: 'closeOperatorUid' },
      },
    };
  },
  mounted() {
    // 是否大园区
    this.fromIm = TNBL.getCurrModule().source === 1;
    if (!business.showModuleSelected) {
      this.initEdgePage();
    } else {
      this.mozuId = parseInt(TNBL.getCurModuleId()) || parseInt(getQueryString('mozuId')) || 326;
      this.initPage();
    }
  },
  beforeDestroy() {
    eventBus.$off('showModal');
  },
  methods: {
    toggleAnalysisDownload() {
      eventBus.$emit('toggleAnalysisDownload');
    },
    toggleAnalysis() {
      eventBus.$emit('toggleSearch');
    },
    initPage() {
      this.tableConfig.searchParams.mozuId = this.mozuId;
      eventBus.$on('showModal', ({ type, data }) => {
        if (type === '详情') {
          if (data.alarm_id_string) { // alarm_id_string存在跳转1.0
            window.open(`/${business.moduleName}/warning-detail?id=${data.id}&mozuId=${data.mozuId}&alarm_id_string=${data.alarm_id_string}`);
          } else {
            window.open(`/${business.moduleName}/warning-detail?id=${data.id}&mozuId=${data.mozuId}`);
          }
        }
      });
    },
    initEdgePage() {
      this.mozuId = getMozuId();
      this.tableConfig.searchParams.mozuId = this.mozuId;
      this.mozuloaded = true;
      eventBus.$on('showModal', ({ type, data }) => {
        if (type === '详情') {
          if (data.alarm_id_string) { // alarm_id_string存在跳转1.0
            window.open(`/${business.moduleName}/warning-detail?id=${data.id}&mozuId=${data.mozuId}&alarm_id_string=${data.alarm_id_string}`);
          } else {
            window.open(`/${business.moduleName}/warning-detail?id=${data.id}&mozuId=${data.mozuId}`);
          }
        }
        if (type === 'alarmType') {
          this.detailVisible = true;
          this.detailData = data;
        }
        if (type === 'analysis' || type.type === 'analysis') {
          this.handleVisible = true;
          this.modalData = data;
        }
      });
    },
    // scopeloaded(val) {
    //   this.mozuloaded = true;
    //   this.mozuId = val.id;
    //   this.initPage();
    // },
    // changeMozu(val) {
    //   this.$set(this.tableConfig.searchParams, 'mozuId', val.id);
    //   this.tableConfig.refreshNow = !this.tableConfig.refreshNow;
    // },
    handleChange() {

    },
    headerClick(column, event) {
      console.log(column, event);
    },
    clearSelection() {
      this.$refs.strategyTable.clearSelection();
    },
    exportData() {},
    handleSizeChange(limit) {
      this.limit = limit;
      this.current = 1;
      this.fetchDatas();
    },
    handleCurrentChange(num) {
      this.current = num;
      this.fetchDatas();
    },
    search() {
      this.current = 1;
      this.fetchDatas();
    },
    fetchDatas() {},
    filterChange(val) {
      console.log(val);
      // eslint-disable-next-line no-restricted-syntax
      for (const item in val) {
        this.tableQueryParams[item] = val[item].toString();
      }
      this.current = 1;
      this.fetchDatas(0);
    },
    // sortChange({ column, prop, order }) {

    // },
    importData() {},
  },
};
</script>

<style>
</style>
<style lang="scss">
.selection-header {
  $height: 64px;

  position: absolute;
  top: -$height;
  z-index: 90;
  height: $height;
  width: 100%;
  background-color: white;
}
.overview {
  display: flex;
  width: 100%;
  min-width: 960px;
  background-color: white;
  margin-bottom: 20px;

  &-left {
    width: 200px;
    border-right: 1px solid #f0f0f0;
  }

  &-right {
    flex: 1;
    display: flex;
    justify-content: space-between;
  }

  .el-data-patch {
    height: 104px;
    padding: 0 24px;
    flex: 1;
  }
}
.selected-div {
  height: 68px;
  width: 100%;
  position: absolute;
  z-index: 90;
  background-color: white;
  box-shadow: 0px 3px 5px rgba(192, 192, 192, 0.8);
  display: flex;
}
#mozu-cascader {
  margin-top:16px;
  float:right;
  width:389px;
  .el-input__inner {
  height: 35px;
  padding-left: 5px ;
}
}
.el-table tbody tr:hover>td {
    background-color:#ffffff!important
}
// .el-table tbody tr{
//     pointer-events:none;
// }
#mozu-cascader .el-input::before{
  content:none;
}
#mozu-cascader .el-input::after{
  content:none;
}
</style>
