<template>
  <div class="grid">
    <div
      class="grid-tabs"
    >
      <el-button
        type="icon"
        :icon="`tn-icon-arrow-${treeCollapsed ? 'right' : 'left'}`"
        @click="toggleTreeVisible"
      />
      <el-button-group style="padding: 12px 0">
        <el-button
          :type="viewType === 'list' ? 'primary': 'plain'"
          @click="changeView('list')"
        >
          列表
        </el-button>
        <el-button
          :type="viewType === 'grid' ? 'primary': 'plain'"
          @click="changeView('grid')"
        >
          网格
        </el-button>
        <el-button
          :type="viewType === 'chart' ? 'primary': 'plain'"
          @click="changeView('chart')"
        >
          图表
        </el-button>
      </el-button-group>
    </div>
    <el-table-toolbar
      v-model="searchValue"
      filter-placeholder="搜索"
      class="grid-toolbar"
      :filter-width="226"
      :actions="[
        {
          text: `导出`,
          icon: 'tn-icon-import',
          action: exportList,
        },
      ]"
      dropdown-width="160"
      @search="filterList"
    >
      <template slot="extra">
        <div class="extra">
          <!-- <el-button @click="addMultiplePoint">
              批量添加
            </el-button> -->

          <!-- <el-button
            class="extra-export"
            @click="exportList"
          >
            导出全部
          </el-button> -->
        </div>
      </template>
    </el-table-toolbar>
    <div
      class="grid-extra-button"
      style="border-bottom: solid 1px #f0f0f0;margin-right: 20px"
    >
      <el-button
        v-if="multipleSelection.length"
        type="text"
        @click="removeInterestIndicatorList"
      >
        移除
      </el-button>
      <el-button
        v-if="multipleSelection.length"
        type="text"
        @click="jumpHistory"
      >
        查看历史
      </el-button>
    </div>
    <bar-chart
      v-if="viewType === 'chart'"
      v-show="showChart"
      ref="barChart"
      class="grid-table"
      :data="{ itps: pointYaxis,pointXaxis: pointXaxis ,pointYaxisUnit: pointYaxisUnit }"
    />
    <div
      v-if="viewType === 'list'"
      class="grid-table"
      style=""
    >
      <el-table
        ref="table"
        :data="tableData"
        style="width: 100%;overflow:overlay;max-height: calc(100vh - 167px);"
        row-key="id"
        height="1500"
        @selection-change="handleSelectionChange"
      >
        <el-table-column
          type="selection"
          width="70"
          reserve-selection
        />
        <el-table-column
          v-if="enableDeviceNumberV2 !== '1'"
          label="设备类型"
          prop="deviceTypesName"
          width="150"
          show-overflow-tooltip
        />
        <el-table-column
          v-if="enableDeviceNumberV2 === '1'"
          label="应用类型"
          prop="applicationTypeZh"
          width="150"
          show-overflow-tooltip
        />
        <el-table-column
          v-if="enableDeviceNumberV2 === '1'"
          prop="categoryZh"
          label="设备种类"
          :width="isPad ? 110 : 150"
          column-key="categoryZh"
          show-overflow-tooltip
        />
        <el-table-column
          :formatter="formatDeviceNumber"
          prop="deviceNumber"
          label="设备编号"
          min-width="310"
          show-overflow-tooltip
        />
        <!-- <el-table-column
          prop="attrId"
          label="测点标识符"
          min-width="160"
          show-overflow-tooltip
        /> -->
        <el-table-column
          prop="attrName"
          label="测点名称"
          min-width="200"
          show-overflow-tooltip
        />

        <el-table-column
          label="当前值"
          width="140"
          show-overflow-tooltip
        >
          <template v-slot="{ row }">
            <el-tag
              :type="row.correct ? 'danger': 'success'"
              style="font-weight: 800;height:20px"
            >
              <span v-if="row.status"> {{ row.enumValue }} </span>
              <span v-else> {{ row.value }} {{ row.unit }}</span>
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="updateTime"
          label="刷新时间"
          width="190"
        />
        <el-table-column
          label="操作"
          width="130"
          fixed="right"
        >
          <template v-slot="{ row }">
            <!-- <el-button
              type="text"
              @click="addToFocus(row)"
            >
              关注
            </el-button>
                        <el-button
              type="text"
              @click="checkSinglePoint(row)"
            >
              查看
            </el-button> -->
            <el-button-group>
              <el-button
                type="text"
                style="margin-right:8px"
                @click="removeToFocus(row)"
              >
                <!-- <i
                  class="tn-icon-star-hollow"
                /> -->
                移除
              </el-button>
              <el-button
                type="text"
                @click="checkSinglePoint(row)"
              >
                <!-- <i
                  class="tn-icon-history"
                /> -->
                历史
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <div
      v-if="viewType === 'grid'"
      class="grid-table"
      style="padding-left: 10px;overflow:scroll;height: calc(100vh - 170px);background-color:#f5f8f9"
    >
      <div class="info-main-wrapper">
        <el-card
          v-for="i in tableData"
          :key="i.id"
          style="margin: 0 10px 10px 0;min-width:200px;box-shadow:0px 1px 3px 0 #cbcbcb80;"
        >
          <div
            v-if="i.status"
            style="height:38px;font-size: 22px;font-weight: 550;"
          >
            {{ i.enumValue }}
          </div>
          <el-data-patch
            v-else
            :value="i.value"
            style="margin-right: 36px;"
            :suffix="i.unit"
            :class="i.correct ? 'correct-rule': ''"
          />
          <div style="font-size: 12px">
            <div style="margin: 8px 0">
              {{ i.attrName }}
            </div>
            <div>
              {{ i.deviceNumber }}
            </div>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script>
import { dataQuery as cgi } from '@@/config/cgi';
import business from '@@/config/business';
import getEdgeRequest from '../../../utils/request';
import eventBus from '../eventBus';
import { cloneDeep } from 'lodash';
import barChart from '../components/bar-chart.vue';
import dayjs from 'dayjs';

const MAX_ITEMS_COUNT = 10;

export default {
  inject: ['getSelNodeData'],
  components: {
    barChart,
  },
  props: {
    mozuloaded: Boolean,
    mozuId: Number,
    selectedData: {
      type: Object,
      default: () => {},
    },
    cardObj: {
      type: Object,
      default: () => {},
    },
    enableDeviceNumberV2: {
      type: String,
      default() {
        return '0';
      },
    },
    devicesMap: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {
      notCascade: false,

      maxItemsCount: MAX_ITEMS_COUNT,
      activeName: 'all',
      loading: false,
      filter: {
        roomCode: [],
        deviceTypesName: [],
        deviceNumber: '',
        attrName: '',
        attrId: '',
      },
      roomList: [],
      deviceTypeList: [],
      tableData: [],
      tableHeight: 600,
      timer: null,
      rtLoading: true,
      selection: [],
      searchValue: '',
      currentPage: 1,
      totalItems: 0,
      pageSize: 10,

      multipleSelection: [],
      treeCollapsed: false,
      viewType: 'list',
      originData: [],
      pointYaxis: [],
      pointXaxis: [],
      pointYaxisUnit: [],
      showChart: false,
    };
  },
  watch: {
    cardObj() {
      if (!this.$refs?.table) return;
      this.$refs.table.clearSelection();
    },
    // viewType(val) {
    //   this.changeView(val);
    // },
    // activeName() {
    //   this.getData();
    // },
    // mozuId() {
    //   this.getData();
    // },
    selectedData: {
      handler(val) {
        this.tableData = val.list;
        this.tableData = this.checkRules(this.tableData);
        this.totalItems = val.count;
        this.originData = cloneDeep(this.tableData);
        this.pointXaxis = this.originData.map(i => `${i.attrName}\n${i.deviceNumber}`);
        this.pointYaxis = this.originData.map(i => i.value);
        this.pointYaxisUnit = this.originData.map(i => i.unit);
        if (this.searchValue) {
          this.filterList();
        }
      },
      deep: true,
    },
  },
  mounted() {
    eventBus.$on('notCascade', (val) => {
      this.notCascade = val;
    });

    this.$nextTick(() => {
      // this.calcTableHeight();
    });
  },
  beforeDestroy() {
    clearTimeout(this.timer);
    this.timer = null;
    eventBus.$off('notCascade');
  },
  methods: {
    formatDeviceNumber(row, col, deviceNumber) {
      return window.tnwebServices.v2DeviceNumberTransformerService.get(deviceNumber, true);
    },
    jumpHistory() {
      this.addMultiplePoint();
      this.checkSelectedPoints();
    },
    checkRules(data) {
      if (this.cardObj.operator === '>') {
        data.forEach((i) => {
          i.correct = parseFloat(i.value) > parseFloat(this.cardObj.value);
        });
      } else if (this.cardObj.operator === '<') {
        data.forEach((i) => {
          i.correct = parseFloat(i.value) < parseFloat(this.cardObj.value);
        });
      } else if (this.cardObj.operator === '=') {
        data.forEach((i) => {
          i.correct = parseFloat(i.value) !== parseFloat(this.cardObj.value);
        });
      } else {
        data.forEach((i) => {
          i.correct = parseFloat(i.value) === parseFloat(this.cardObj.value);
        });
      }
      if (this.cardObj.hasRule === false) {
        data.forEach((i) => {
          i.correct = false;
        });
      }
      return data;
    },
    removeInterestIndicatorList() {
      this.$emit('removeFocus', this.multipleSelection.map(i => i.focusId));
    },
    removeToFocus(row) {
      this.$emit('removeFocus', [row.focusId]);
    },
    changeView(viewType) {
      this.viewType = viewType;
      if (viewType === 'chart') {
        this.showChart = true;
      }
    },
    handleSelectionChange(val) {
      this.multipleSelection = val;

      // if (this.multipleSelection.length >= 10) {
      //   this.multipleSelection = val.slice(0, 10);

      //   const others = val.slice(10);

      //   if (others.length) {
      //     others.forEach((e) => {
      //       this.$refs.table.toggleRowSelection(e, false);
      //     });
      //   }
      // }
    },

    /**
     * 批量添加测点
     */
    addMultiplePoint() {
      for (let i = 0; i < this.multipleSelection.length; i++) {
        const e = this.multipleSelection[i];
        if (this.selection.length < 100) {
          if (!this.selection.find(item => item.id === e.id)) {
            this.selection.push(e);
          }
        } else {
          this.$message('最多可选择10个测点');
          return;
        }
      }

      this.$refs.table.clearSelection();
    },

    toggleTreeVisible() {
      this.$emit('tree-visible-change', this.treeCollapsed);
      this.treeCollapsed = !this.treeCollapsed;
      setTimeout(() => {
        this.$refs.barChart.resizeChart();
      }, 1000);
    },

    /**
     * 拉取高级筛选的房间候选项
     */
    getRoomList() {
      getEdgeRequest(this.$axios, this.mozuId)
        .post(cgi.getDistinctByFieldName, { fieldName: 'roomCode' }, false)
        .then((data) => {
          this.roomList = data.map(e => ({
            value: e,
            label: e,
          }));
        });
    },

    /**
     * 拉取高级筛选房间对应的设备类型
     */
    getDeviceTypesName() {
      const params = {
        conditions: {
          name: 'roomCode',
          value: this.filter.roomCode,
        },
        fieldName: 'deviceTypesName',
      };

      getEdgeRequest(this.$axios, this.mozuId)
        .post(cgi.getDistinctByFieldName, params, false)
        .then((data) => {
          this.deviceTypeList = data.map(e => ({
            value: e,
            label: e,
          }));
          this.$emit('getDeviceTypeList', this.deviceTypeList);
        });
    },

    /**
     * 重置高级筛选
     */
    reset() {
      this.$set(this, 'filter', {
        roomCode: [],
        deviceTypesName: [],
        deviceNumber: '',
        attrName: '',
      });
      this.$refs.filter.resetFields();
    },

    createConditions() {
      const conditions = [];

      Object.keys(this.filter).forEach((key) => {
        const val = this.filter[key];

        if (Array.isArray(val)) {
          if (val.length) {
            conditions.push({
              name: key,
              value: val,
            });
          }
        } else {
          if (val !== '') {
            conditions.push({
              name: key,
              value: [val],
            });
          }
        }
      });

      if (this.activeName === 'status') {
        conditions.push({ name: 'status', value: ['true'] });
      }
      if (this.activeName === 'simulation') {
        conditions.push({ name: 'status', value: ['false'] });
      }
      if (this.activeName === 'control') {
        conditions.push({ name: 'rw', value: ['true'] });
      }

      return conditions;
    },

    filterList() {
      if (this.searchValue) {
        this.tableData = this.originData
          .filter(i => i.deviceNumber.indexOf(this.searchValue) > -1 || i.attrName.indexOf(this.searchValue) > -1);
      } else {
        this.tableData = this.originData;
      }
    },

    getData() {
      if (business.showModuleSelected && !this.mozuloaded) {

      }
      const cgiUrl = cgi.getCurrentBizGidAttrsWithValueByConditions;
      const selNode = this.getSelNodeData();
      clearTimeout(this.timer);

      if (selNode) {
        this.selNode = selNode;

        const params = {
          conditions: [
            { name: 'deviceGid', value: [selNode.id] },
            ...this.createConditions(),
          ],
          start: (this.currentPage - 1) * this.pageSize,
          limit: this.pageSize,
          keyword: this.searchValue,
          operator: 'like',
          notCascade: this.notCascade,
        };

        getEdgeRequest(this.$axios, this.mozuId)
          .post(cgiUrl, params, '', false)
          .then((data) => {
            this.totalItems = data.count;
            this.tableData = data.list;
            this.timer = setTimeout(() => {
              this.getData();
            }, 3000);
          })
          .finally(() => {
            this.rtLoading = false;
          });
      }
    },

    /**
     * 改条件重新高级筛选后，将当前页改为1
     */
    searchDataByFilter() {
      this.$refs.table.clearSelection();
      this.currentPage = 1;
      this.notCascade = false;
      this.getData();
    },

    calcTableHeight() {
      this.tableHeight = this.$refs.tags.$el.offsetHeight - 56;
    },

    visibleChangeHandler(val) {
      this.$emit('collapse-change', val);
    },

    exportList() {
      const params = this.tableData.map(i => ({ gids: [i.gid], attrs: [i.attrId] }));

      const { title } = this.cardObj;

      const getTsString = () => dayjs().format('YYYYMMDDHHmmss');

      this.$axios.download('/cgi/dataQuery/edge/exportGidAndAttrListValueMapWithoutCache', { gidWithAttrListMap: params }, true, title ? {
        fileName: `${title}-${getTsString()}.xlsx`,
      } : {});
    },

    /**
     * 切换不同节点后，重新拉取
     */
    refresh() {
      this.$refs.table.clearSelection();
      this.currentPage = 1;
      this.getData();
    },

    /**
     * 增加测点到已选测点列表中
     * @param {Object} row - 选择的列
     */
    addPoint(row) {
      if (!this.selection.find(e => e.id === row.id)) {
        if (this.selection.length === 10) {
          this.$message('最多可选择10个测点');
          return false;
        }

        this.selection.push(row);
      }
    },

    /**
     * 删除测点
     */
    handleClose(tag) {
      this.selection.splice(this.selection.findIndex(e => e.id === tag.id), 1);
    },

    /**
     * 清空已选测点列表
     */
    clearSelection() {
      this.selection = [];
    },

    /**
     * 查看单个测定的历史数据
     */
    checkSinglePoint(row) {
      this.goToAdvancedSearch([row]);
    },

    /**
     * 查看已选测点列表的历史数据
     */
    checkSelectedPoints() {
      this.goToAdvancedSearch(this.selection);
    },

    /**
     * 跳转至历史数据高级查询页面
     */
    goToAdvancedSearch(list) {
      const { moduleName } = business;
      const result = list.map(e => e.templatePointId).join(',');
      const href = `/${moduleName}/advanced-search?title=${this.cardObj.title}&pointlist=${result}`;
      window.open(href);
      this.selection = [];
    },

    /**
     * 分页 pageSize处理
     */
    handleSizeChange(val) {
      this.$refs.table.clearSelection();
      this.pageSize = val;
      this.currentPage = 1;
      this.getData();
    },

    /**
     * 分页 当前页处理
     */
    handleCurrentChange(val) {
      this.$refs.table.clearSelection();
      this.currentPage = val;
      this.getData();
    },
  },
};
</script>

<style lang="scss" scoped>

.info-main-wrapper {
  padding: 10px 0 0 0;
  display: inline-grid;
  grid-template-columns: repeat(auto-fit,minmax(200px,1fr));
  grid-column-gap: 5px;
  grid-auto-flow: row;
  margin-bottom:16px;
  width:100%;
  color:#000

}

/deep/ .el-table__body-wrapper {
  overflow-x: hidden;
  &:hover {
    overflow: auto;
  }
}
.table-link {
  color: #1470cc;
}

.correct-rule {
  /deep/ .el-data-patch__value {
    color: red
  }
}

/deep/ .el-card__body {
  padding: 16px 16px 8px 16px;
  .el-data-patch__title {
    display: none
  }
}

.filter {
  &-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));

    .el-select {
      /deep/ input {
        height: 26px !important;
      }
    }
  }

  &-footer {
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: flex-end;

    &-reset {
      color: #333;
    }
  }
}

.grid {
  height: 100%;
  // overflow-x: hidden;
  display: grid;
  grid-template-areas:
    'tabs extra toolbar'
    // 'filter filter'
    // 'toolbar toolbar'
    'table table table';
  grid-template-columns: minmax(0, 1fr) 300px;
  grid-template-rows: auto auto auto 1fr;

  &-tabs {
    grid-area: tabs;
    display: flex;

    .el-tabs {
      flex: 1;
    }
  }

  &-filter {
    grid-area: filter;
  }

  &-table {
    grid-area: table;
    // background-color: #f5f8f9;
  }

  &-extra-button {
    grid-area: extra;
    display:flex;
    align-items: center;
    justify-content: right;
  }

  &-toolbar {
    grid-area: toolbar;
    position: relative;
    height: 56px;
    padding:0;
    /deep/ .el-table-toolbar__search {
      padding: 0 0 0 16px !important;
    }
    /deep/ .el-table-toolbar__extra {
      padding: 0px !important;
    }

    &:after {
      content: '';
      position: absolute;
      left: 0;
      right: 0;
      bottom: -1px;
      height: 0;
      border-bottom: 1px solid #f0f0f0;
    }
  }

  &-tags {
    border-left: 1px solid #f0f0f0;

    display: grid;
    grid-template-rows: auto minmax(0, 1fr);

    /deep/ .el-block__body-inner {
      display: flex;
      flex-direction: column;
      height: 100%;
      box-sizing: border-box;
      width: 260px;
      padding-bottom: 16px;
    }

    &-body {
      height: 480px;
      overflow-y: auto;
      box-sizing: border-box;
      padding: 0 24px;
    }

    .el-tag {
      margin-bottom: 8px;
      height: auto;
      white-space: unset;
      width: 100%;
      padding-right: 18px;
      position: relative;

      /deep/ &__close {
        position: absolute;
        right: 4px;
        top: 50%;
        transform: translateY(-50%);
      }
    }

    &-footer {
      margin-top: auto;
      align-self: center;
    }

    &-empty-text {
      color: #999;
      text-align: center;
      margin-top: 16px;
    }
  }
}
</style>
