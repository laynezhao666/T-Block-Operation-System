<template>
  <div>
    <div style="display:flex;width:100%">
      <el-title
        v-if="showTitle"
        style="width:200px"
      >
        告警配置
      </el-title>
    </div>
    <el-block
      v-loading="!pageConfig"
      no-padding
      class="strategy-list"
    >
      <el-tabs
        v-if="pageConfig"
        v-model="activeTab"
      >
        <el-tab-pane
          name="strategy"
          label="设备告警策略"
        >
          <common-table
            :columns="columns"
            :table-config="tableConfig"
            :config-cgi="cgi"
            style="border-top:1px solid #f0f0f0"
          >
            <template v-slot:extraButtons>
              <el-button
                v-if="business.showModuleSelected"
                type="primary"
                auth-right-code="TNJKGL-GJPZ-GL"
                @click="add"
              >
                新增
              </el-button>
            </template>
          </common-table>
        </el-tab-pane>
        <el-tab-pane
          v-if="!business.isTedge"
          name="instanceDiff"
          label="告警策略基线对比"
        >
          <instance-diff :mozu-id="mozuId" />
        </el-tab-pane>
        <el-tab-pane
          v-if="business.showModuleSelected"
          name="operation"
          label="操作日志"
        >
          <div style="border-top:1px solid #f0f0f0;position:relative">
            <div
              v-if="showSelectionHeader"
              class="export-selected-div"
            >
              <div style="width:100px">
                <el-button>导出所选</el-button>
              </div>
              <div style="flex:1;">
                <span style="float:right;margin:5px 20px 0 0;font-size:16px">已选择       <i
                  class="tn-icon tn-icon-close"
                  style="transform:translate(0px,3px);cursor:pointer"
                  @click="showSelectionHeader = false"
                /></span>
              </div>
            </div>
            <el-table-toolbar
              v-model="searchValue"
              placeholder="搜索操作内容"
              @search="searchOperation"
            >
              <template slot="extra" />
            </el-table-toolbar>
            <el-table
              ref="table"
              class="warning-strategy-optable"
              :cell-height="49"
              element-loading-text="正在获取数据"
              :data="operationList"
              style="width: 100%; text-align: left"
              @selection-change="handleSelectionChange"
            >
              <el-table-column
                type="index"
                label="序号"
                width="80"
              />

              <el-table-column
                v-for="item in columnsOperation"
                :key="item.id"
                :label="item.label"
                :prop="item.prop"
                :sortable="item.sortable"
                align="left"
                :width="item.width"
                :formatter="formatter"
              >
                <template
                  v-if="item.label === '操作人'"
                  slot="header"
                >
                  <span
                    :style="operator ? 'color:#1470CC' : 'color:#333'"
                  >  {{ item.label }}</span>
                  <user-selector
                    v-if="item.label === '操作人'"
                    :filter="filter"
                    @input="v => operatorSearch(v)"
                  />
                </template>
                <template
                  v-else-if="item.label === '操作名称'"
                  slot="header"
                >
                  <span
                    :style="operation ? 'color:#1470CC' : 'color:#333'"
                  >  {{ item.label }}</span>
                  <head-selector
                    v-if="item.label === '操作名称'"
                    type="logOperation"
                    :filter="filter"
                    @change="(v) => operationSearch(v)"
                  />
                </template>
                <template
                  v-else-if="item.label === '操作时间'"
                  slot="header"
                >
                  <span>  {{ item.label }}</span>
                  <date-filter
                    :value="time"
                    @input="v => chooseDate(v)"
                  />
                </template>
                <template
                  v-else
                  slot="header"
                >
                  {{ item.label }}
                </template>
                <template />
              </el-table-column>
            </el-table>
            <el-pagination
              styled
              layout="total, sizes, prev, pager, next, jumper"
              :pager-count="5"
              :page-sizes="[10, 20, 50, 100, 500]"
              :total="formData.total"
              :current-page="formData.offset"
              :page-size.sync="formData.limit"
              @size-change="v => filterHandler({limit: v})"
              @current-change="v => filterHandler({currentPage: v})"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-block>
    <detail-modal
      :mozu-id="mozuId"
      :visible.sync="modalVisible"
      :modal-data="modalData"
    />
    <edit-modal
      :mozu-id="mozuId"
      :visible.sync="editModalVisible"
      :modal-data="modalData"
      @successchange="refresh"
    />
    <add-modal
      :mozu-id="mozuId"
      :visible.sync="addModalVisible"
      @successchange="refresh"
    />
  </div>
</template>

<script>
import business from '@@/config/business';
import commonTable from '../component/commonTable/ConfigPanel/index.vue';
import config from './config';
import { warning as cgi } from '@@/config/cgi';
import detailModal from './detailModal.vue';
import editModal from './editModal.vue';
import addModal from './addModal.vue';
import { eventBus } from '../component/commonTable/script/eventBus';
import headSelector from './head-selector';
import userSelector from './user-selector';
import instanceDiff from './instanceDiff';
import dateFilter from './date-filter';
import getEdgeRequest from '../../utils/request';
import { getQueryString } from 'common/script/utils';
import mixin from 'feature/utils/mixin';

export default {
  components: {
    commonTable,
    detailModal,
    editModal,
    headSelector,
    userSelector,
    addModal,
    dateFilter,
    instanceDiff,
  },
  mixins: [mixin],
  data() {
    return {
      activeTab: 'strategy',
      business,
      mozuloaded: false,
      showHeaderDate: false,
      operationTime: [],
      mozuId: null,
      operationList: [],
      filter: {},
      value: '',
      mozuName: [],
      options: [],
      operation: '',
      operator: '',
      time: [],
      showSelectionHeader: false,
      activeNamesLog: [1],
      activeNames: [1],
      modalData: {},
      columns: window.location.search !== '?ivq7nnvmof' ? config : (business.isTedge ? [
        ...config.slice(0, -1),
        {
          ...config[config.length - 1],
          // 临时放开编辑权限
          operationUrl: [
            { operation: '详情', url: 'c', authCode: 'TNJKGL-GJPZ-CK' },
            { operation: '编辑', url: 'c' },
            { operation: '删除', url: 'c', authCode: 'TNJKGL-GJPZ-GL', cloudFlag: !business.showModuleSelected },
          ],
        },
      ] : config),
      overview: {},
      modalVisible: false,
      editModalVisible: false,
      addModalVisible: false,
      loading: false,
      count: 0,
      current: 1,
      limit: 10,
      searchValue: '',
      formData: {
        keyword: '',
        offset: 1,
        limit: 10,
        limits: 10,
        total: 0,
      },
      columnsOperation: [
        { label: '操作人', prop: 'operatorName', width: 120 },
        {
          label: '操作时间',
          prop: 'time',
          width: 180,
        },
        { label: '操作名称', prop: 'action', width: 120 },
        { label: '操作内容', prop: 'description' },
      ],
      cgi: {
        queryCgi: cgi.getStrategyList,
        exportCgi: cgi.exportStrategyList,
      },
      tableData1: [],
      tableConfig: {
        rights: 0b10100,
        showSetting: false,
        showSearch: true,
        deleteCgi: cgi.deleteCustom,
        placeHolder: '搜索告警内容',
        refreshNow: false,
        searchParams: {
          // eslint-disable-next-line no-underscore-dangle
          mozuId: this.$moduleInfo.mozuId || 326,
        },
        rowClass({ rowIndex }) {
          if (rowIndex % 2 === 0) {
            return { 'background-color': '#f5f5f5' };
          }
          // if (row.row.isStandard === false) {
          //  return { 'background-color': '#fdf5e6' };
          // }
          // return { 'background-color': '#f0f9eb' };
        },
        searchNameMap: { createdByName: 'createdBy', updatedByName: 'updatedBy', deviceNumberList: 'deviceNumber' },
        useNoEdge: !business.isTedge,
      },

      pageConfig: null,
    };
  },
  created() {
    this.initConfig();
  },
  mounted() {
    if (!this.business.showModuleSelected) {
      this.initEdgePage();
    } else {
      this.mozuId = parseInt(TNBL.getCurModuleId()) || parseInt(getQueryString('mozuId')) || 326;
      this.activeTab = getQueryString('tab') || 'strategy';
      this.initPage();
    }
  },
  beforeDestroy() {
    eventBus.$off('showModal');
  },
  provide() {
    return {
      tableConfig: this.tableConfig,
    };
  },
  methods: {
    async initConfig() {
      this.pageConfig = await window.tnwebServices.customConfigService.initCurrentPageConfig({
        url: window.location.href,
        content: {
          type: 'Yaml',
          // eslint-disable-next-line import/no-webpack-loader-syntax
          defaultContent: require('!raw-loader!./default-config.yaml').default,
        },
        docs: {
          type: 'markdown',
          // eslint-disable-next-line import/no-webpack-loader-syntax
          content: require('!raw-loader!./config.md').default,
        },
      });

      this.columns = !this.pageConfig.editable ? config : (business.isTedge ? [
        ...config.slice(0, -1),
        {
          ...config[config.length - 1],
          // 临时放开编辑权限
          operationUrl: [
            { operation: '详情', url: 'c', authCode: 'TNJKGL-GJPZ-CK' },
            { operation: '编辑', url: 'c' },
            { operation: '删除', url: 'c', authCode: 'TNJKGL-GJPZ-GL', cloudFlag: !business.showModuleSelected },
          ],
        },
      ] : config);
    },

    initPage() {
      this.tableConfig.searchParams.mozuId = this.mozuId;
      eventBus.$on('showModal', ({ type, data }) => {
        if (type === '详情') {
          this.modalVisible = true;
          this.modalData = data;
        }
        if (type === '编辑') {
          this.editModalVisible = true;
          this.modalData = data;
        }
        if (type === '删除') {
          this.deleteRow(data);
        }
      });
      this.filterHandler();
      getEdgeRequest(this.$axios, this.mozuId).post(cgi.getStrategyStatistics)
        .then((data) => {
          this.overview = data;
        });
    },
    initEdgePage() {
      this.mozuId = this.$moduleInfo.mozuId;
      this.tableConfig.searchParams.mozuId = this.mozuId;
      this.mozuloaded = true;
      eventBus.$on('showModal', ({ type, data }) => {
        if (type === '详情') {
          this.modalVisible = true;
          this.modalData = data;
        }
        if (type === '编辑') {
          this.editModalVisible = true;
          this.modalData = data;
        }
        if (type === '删除') {
          this.deleteRow(data);
        }
      });
    },
    scopeloaded(val) {
      this.mozuloaded = true;
      this.mozuId = val.id;
      this.initPage();
    },
    changeMozu(val) {
      this.$set(this.tableConfig.searchParams, 'mozuId', val.id);
      this.mozuId = val.id;
      this.tableConfig.refreshNow = !this.tableConfig.refreshNow;
      this.getOperationList();
    },
    toggleVisible(show) {
      if (show) {
        this.showHeaderDate = true;
      } else {
        this.showHeaderDate = false;
      }
    },
    getOperationList(params) {
      const mozuId = this.mozuId || Number(window.__GetFrameDataByKey('curMozuData')?.id);
      getEdgeRequest(this.$axios, mozuId)
        .post(cgi.getOperationLog, {
          mozuId,
          ...params,
        })
        .then((data) => {
          this.formData.total = data.count;
          this.operationList = data.list;
        });
    },
    handleSelectionChange() {

    },
    changeSize1() {

    },
    handleChange(data) {
      console.log(data);
    },
    formatter(row, column, val) {
      if (val.length > 200) {
        return `${val.substr(0, 200)}......`;
      }
      return val;
    },
    deleteRow(data) {
      this.$confirm('确认删除吗?', '提示', { type: 'warning' }).then(() => {
        getEdgeRequest(this.$axios, this.mozuId).post(cgi.deleteCustom, { id: parseInt(data.id) })
          .then(() => {
            this.$message.success('删除成功');
            this.tableConfig.refreshNow = !this.tableConfig.refreshNow;
            this.getOperationList();
          });
      });
    },
    refresh() {
      this.getOperationList();
    },
    chooseDate(data) {
      this.time = data;
      this.filterHandler();
    },
    add() {
      this.addModalVisible = true;
    },
    headerClick(column, event) {
      console.log(column, event);
    },
    clearSelection() {
      this.$refs.strategyTable.clearSelection();
    },
    exportData() {
      this.showSelectionHeader = true;
    },
    customSort1() {

    },
    changePage1() {

    },
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
      // eslint-disable-next-line no-restricted-syntax
      for (const item in val) {
        this.tableQueryParams[item] = val[item].toString();
      }
      this.current = 1;
      this.fetchDatas(0);
    },
    // sortChange({ column, prop, order }) {

    // },
    operatorSearch(v) {
      this.operator = v;
      // if (v !== '') {
      // }
      this.filterHandler();
    },
    operationSearch(v) {
      this.operation = v;
      this.filterHandler();
    },
    searchOperation(v) {
      const param = { keyword: v };
      this.formData.offset = 1;
      this.filterHandler(param);
    },
    filterHandler(v = {}) {
      if (!v.currentPage) {
        v.offset = 1;
        this.offset = 1;
      }

      const time = v.time || this.time || [];
      const operator = v.operator || this.operator;
      const opeartion = v.operation || this.operation;
      const params = {
        limit: this.formData.limit,
        ...v,
        timeStart: time && time[0],
        timeEnd: time && time[1],
      };
      params.offset = (v.currentPage - 1) * params.limit;
      if (operator) {
        params.operator = operator;
      }
      params.action = opeartion;
      this.getOperationList(params);
    },
    importData() {},
  },
};
</script>
<style lang="scss">
.el-cascader-menu {
  min-width: 120px;
}
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
.strategy-list {
  margin-top: 0;
}
.el-collapse-item__content {
  padding-bottom: 0;
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
.el-table tbody tr:hover>td {
    background-color:#ffffff!important
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
.export-selected-div {
  display:flex;
  height:44px;
  position:absolute;
  z-index:999;
  background-color:#fff;
  padding:12px 10px 0px 10px;
  width:100%;
  box-shadow: -1px 3px 5px rgba(192, 192, 192, 0.8);
}
#mozu-cascader .el-input::before{
  content:none;
}
#mozu-cascader .el-input::after{
  content:none;
}
.warning-strategy-optable.el-table th>.cell {
    display: flex !important;
  }
</style>
