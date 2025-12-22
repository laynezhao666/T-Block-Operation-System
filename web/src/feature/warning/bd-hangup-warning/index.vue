<template>
  <div>
    <div style="display:flex;width:100%">
      <el-title
        v-if="showTitle"
        style="width:200px"
      >
        已挂起告警
        <el-help-tip icon="tn-icon-question">
          <div style="line-height: 40px">
            <p>当前告警：告警未恢复且未挂起</p>
            <p>挂起告警：告警未恢复且已挂起</p>
            <p>历史告警：告警已恢复</p>
          </div>
        </el-help-tip>
      </el-title>
    </div>
    <div class="overview">
      <div class="overview-left">
        <el-data-patch
          title="总告警"
          :value="overview.all"
        />
      </div>
      <div class="overview-right">
        <el-data-patch
          title="零级"
          :value="overview.L0"
        />
        <el-data-patch
          title="一级"
          :value="overview.L1"
        />
        <el-data-patch
          title="二级"
          :value="overview.L2"
        />
        <el-data-patch
          title="三级"
          :value="overview.L3"
        />
        <el-data-patch
          title="四级"
          :value="overview.L4"
        />
        <el-data-patch
          title="五级"
          :value="overview.L5"
        />
      </div>
    </div>
    <el-block no-padding>
      <el-table-toolbar
        v-show="!isRowsSelected"
        :hide-search="true"
        class="active-toolbar"
        :actions="[
          {
            text: `导出`,
            icon: 'tn-icon-import',
            action: exportList,
          },
        ]"
        dropdown-width="160"
        filter-placeholder="请输入您需要检索的告警内容"
      >
        <!-- <template slot="extra">
          <el-button
            plain
            @click="exportList"
          >
            导出
          </el-button>
        </template> -->
      </el-table-toolbar>

      <selection-toolbar
        v-show="isRowsSelected"
        :selected-list="selectedList"
      >
        <el-button
          type="primary"
          @click="batchCancelHangup"
        >
          批量挂起
        </el-button>
      </selection-toolbar>

      <el-table
        ref="table"
        :data="tableData"
        :row-style="rowStyle"
        style="width: 100%"
        row-key="id"
        @filter-change="filterHandler"
        @selection-change="handleSelectionChange"
      >
        <el-table-column
          v-if="isHangUpEnable"
          type="selection"
          reserve-selection
          width="70"
        />

        <!-- <el-table-column
          label="告警等级"
          width="100"
          prop="Level"
          column-key="Level"
          :filter-multiple="true"
          :filters="levelList"
        /> -->
        <el-table-column
          label="触发时间"
          prop="occurTime"
          width="170"
        >
          <template
            slot="header"
            slot-scope="scope"
          >
            <el-popover
              ref="date-popover"
              placement="bottom"
              width="400"
              trigger="click"
              popper-class="tpopover"
            >
              <el-date-picker
                v-model="timeRange"
                type="datetimerange"
                value-format="yyyy-MM-dd HH:mm:ss"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                align="right"
                @change="v => filterHandler({ 'OccurTimeStart': v && v[0],'OccurTimeEnd':v && v[1] })"
              />
              <span slot="reference">
                <span :style="{color:timeRange?'#1470cc':''}">{{ scope.label || '触发时间' }}</span>
                <!--这里改结构到和其他filter的th一样-->
                <span class="el-table__column-filter-trigger">
                  <!--无法判断状态和修改class：https://github.com/ElemeFE/element/issues?page=2&q=table+date+picker&utf8=%E2%9C%93-->
                  <i class="el-icon-caret-bottom" />
                </span>
              </span>
            </el-popover>
          </template>
        </el-table-column>
        <el-table-column
          label="告警等级"
          width="100"
          prop="level"
          column-key="level"
          :filter-multiple="true"
          :filters="levelList"
        >
          <template slot-scope="scope">
            <span :style="getStyle(scope.row.level)">
              {{ scope.row.level | LevelMap }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          label="设备编号"
          prop="deviceNumber"
          width="350"
        />
        <el-table-column
          label="告警类型"
          column-key="alarmType"
          :filter-multiple="true"
          :filters="alarmTypeList"
          prop="alarmType"
          width="150"
        />
        <el-table-column
          label="告警内容"
          prop="content"
          width="150"
        >
          <template
            slot="header"
            slot-scope="scope"
          >
            <span
              :style="{
                color: hangupReasonContent ? '#1470cc' : ''
              }"
            >{{ scope.column.label }}
              <el-popover
                ref="content-popover"
                placement="bottom"
                transition="el-zoom-in-top"
                popper-class="el-table-time-filter"
                trigger="click"
                @show="popoverContentShow"
                @hide="popoverContentHide"
              >
                <el-input
                  v-model="hangupReasonContent"
                  size="small"
                  placeholder="请输入内容"
                />
                <div class="el-table-filter__bottom">
                  <button
                    :disabled="!hasSearchContent"
                    :class="{ 'is-disabled': !hasSearchContent }"
                    @click="contentFilter({Content: [hangupReasonContent]})"
                  >筛选</button>
                  <button @click="clearSearchContent">重置</button>
                </div>
                <i
                  slot="reference"
                  :class="iconContentControl ? 'el-table__column-filter-trigger el-icon-caret-bottom'
                    : 'el-table__column-filter-trigger el-icon-caret-top'"
                  style="color: #c0c0c0;"
                />
              </el-popover>
            </span>
          </template>
        </el-table-column>
        <el-table-column
          label="触发值"
          width="150"
        >
          <template slot-scope="scope">
            <div
              v-for="(item) in scope.row.occurPointList"
              :key="item.zhName"
            >
              {{ item.zhName + '：' + item.value + item.unit }}
            </div>
          </template>
        </el-table-column>
        <el-table-column
          label="模组"
          prop="mozuName"
          width="150"
        />
        <el-table-column
          label="房间"
          prop="roomName"
          width="150"
        />
        <el-table-column
          label="挂起原因"
          prop="hangupReason"
          width="150"
        >
          <template
            slot="header"
            slot-scope="scope"
          >
            <span
              :style="{
                color: hangupReasonSearch ? '#1470cc' : ''
              }"
            >{{ scope.column.label }}
              <el-popover
                ref="reason-popover"
                placement="bottom"
                transition="el-zoom-in-top"
                popper-class="el-table-time-filter"
                trigger="click"
                @show="popoverShow"
                @hide="popoverHide"
              >
                <el-input
                  v-model="hangupReasonSearch"
                  size="small"
                  placeholder="请输入内容"
                />
                <div class="el-table-filter__bottom">
                  <button
                    :disabled="!hasSearchValue"
                    :class="{ 'is-disabled': !hasSearchValue }"
                    @click="contentFilter({hangupReason: hangupReasonSearch})"
                  >筛选</button>
                  <button @click="clearSearch">重置</button>
                </div>
                <i
                  slot="reference"
                  :class="iconControl ? 'el-table__column-filter-trigger el-icon-caret-bottom'
                    : 'el-table__column-filter-trigger el-icon-caret-top'"
                  style="color: #c0c0c0;"
                />
              </el-popover>
            </span>
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="150"
          fixed="right"
        >
          <template slot-scope="scope">
            <el-button
              type="text"
              size="small"
              @click="handleClick(scope.row)"
            >
              详情
            </el-button>

            <el-button
              v-if="isHangUpEnable"
              type="text"
              size="small"
              @click="cancelHangUp(scope.row)"
            >
              解除挂起
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        layout="total, prev, pager, next, sizes, jumper"
        styled
        background
        :pager-count="5"
        :total="totalItems"
        :current-page.sync="currentPage"
        :page-sizes="[10, 20, 30, 40, 50, 100]"
        :page-size="limit"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-block>
  </div>
</template>
<script>
import { warning as cgi } from '@@/config/cgi';
import { getQueryString } from 'common/script/utils';
import { orderBy } from 'lodash';
import getEdgeRequest from '../../utils/request';
import { getMozuId } from '../../utils/business';
import business from '@@/config/business';
import mixin from 'feature/utils/mixin';
import SelectionToolbar from 'feature/component/tedge-components/selection-toolbar.vue';

export default {
  components: {
    SelectionToolbar,
  },
  filters: {
    LevelMap(value) {
      if (value === 'L0') {
        return '零级';
      }
      if (value === 'L1') {
        return '一级';
      }
      if (value === 'L2') {
        return '二级';
      }
      if (value === 'L3') {
        return '三级';
      }
      if (value === 'L4') {
        return '四级';
      }
      if (value === 'L5') {
        return '五级';
      }
      return value;
    },
  },
  mixins: [mixin],
  data() {
    return {
      iconControl: true,
      iconContentControl: true,
      hangupReasonSearch: '',
      hangupReasonContent: '',
      totalItems: 0,
      currentPage: 1,
      limit: 10,
      tableData: [],
      mozuId: 0,
      overview: {},
      prevSearchValue: '',
      modalVisible: false,
      levelList: [{
        text: '零级',
        value: 'L0',
      }, {
        text: '一级',
        value: 'L1',
      }, {
        text: '二级',
        value: 'L2',
      }, {
        text: '三级',
        value: 'L3',
      }, {
        text: '四级',
        value: 'L4',
      }, {
        text: '五级',
        value: 'L5',
      }],
      alarmTypeList: [],
      filtered: {

      },
      timer: null,
      popoverVisible: false,
      timeRange: null,
      searchResultType: 0, // 1表示搜索是

      selectedList: [],
      isHangUpEnable: false,
    };
  },
  computed: {
    hasSearchValue() {
      return (this.hangupReasonSearch !== '');
    },
    hasSearchContent() {
      return (this.hangupReasonContent !== '');
    },
    isRowsSelected() {
      return !!this.selectedList?.length;
    },
  },
  watch: {
    timeRange(v) {
      console.log(v);
    },
  },
  mounted() {
    const deviceNo = getQueryString('deviceNo');
    if (deviceNo) {
      this.searchValue = deviceNo;
    }
    if (business.showModuleSelected) {
      // eslint-disable-next-line radix
      this.mozuId = parseInt(TNBL.getCurModuleId()) || parseInt(getQueryString('mozuId')) || 326;
    }
    this.getOverviewData();
    this.queryWarning();
    this.loadIsHangUpEnable();
  },
  beforeDestroy() {
    clearTimeout(this.timer);
  },
  methods: {
    popoverShow() {
      this.iconControl = false;
    },
    popoverHide() {
      this.iconControl = true;
    },
    popoverContentShow() {
      this.iconContentControl = false;
    },
    popoverContentHide() {
      this.iconContentControl = true;
    },
    clearSearch() {
      this.hangupReasonSearch = '';
      // 清空搜索值在拉取新的数据
      this.filterHandler({ hangupReason: this.hangupReasonSearch });
      this.$refs['reason-popover'].doClose();
    },
    clearSearchContent() {
      this.hangupReasonContent = '';
      this.filterHandler({ Content: [] });
      this.$refs['content-popover'].doClose();
    },
    filterReason() {

    },
    getOverviewData() {
      getEdgeRequest(this.$axios, this.mozuId).post(
        cgi.getAlarmType,
        { start: 0, limit: 0, alarmStatus: 2, mozuId: getMozuId(), eventStatus: -1 }
      )
        // this.$axios.post(cgi.getAlarmType, { start: 0, limit: 0 })
        .then((data) => {
          console.log('data', data);
          this.alarmTypeList = Object.keys(data).map(item => ({
            text: item,
            value: item,
          }));
        });
    },
    getStyle(level) {
      const colors = {
        L0: '#ff3e00',
        L1: '#ff3e00',
        L2: '#ff9200',
        L3: '#fbd743',
        L4: '#008adc',
        L5: '#8acbf2',
      };

      return {
        color: colors[level],
        border: `1px solid${colors[level]}`,
        padding: '0 8px',
        'border-radius': '6px',
      };
    },
    rowStyle({ row }) {
      if (row.Level === 'L0') {
        return 'background:  #ffff00; animation: twinkling 2s infinite ease-in-out; color: rgb(248, 61, 124)';
      }
    },
    getParam() {
      const start = (this.currentPage - 1) * this.limit;
      const parm = {
        start,
        limit: this.limit,
        mozuId: getMozuId(),
        alarmStatus: 2,
        eventStatus: -1,
      };
      Object.assign(parm, this.filtered);

      return parm;
    },
    queryWarning() {
      if (this.timer) {
        clearTimeout(this.timer);
        this.timer = null;
      }
      const parm = this.getParam();
      getEdgeRequest(this.$axios, this.mozuId).post(
        // this.$axios.post(
        cgi.getActivedWarning, parm,
        false
      )
        .then((data) => {
          this.totalItems = data.count;
          this.tableData = data.list;
          this.tableData.forEach((item) => {
            this.$set(item, 'OccurPointList', orderBy(item.OccurPointList, 'enName'));
            this.$set(item, 'occurPointList', orderBy(item.occurPointList, 'enName'));
          });
          if (data.count === 0) {
            if (this.searchResultType === 1) {
              this.searchResultType = 2;
              this.queryWarning();
              return;
            }
          }

          this.timer = setTimeout(() => {
            this.queryWarning();
          }, 3000);
        });

      // overview
      getEdgeRequest(this.$axios, this.mozuId).post(
        // this.$axios.post(
        cgi.getActivedOverview,
        { alarmStatus: 2, mozuId: getMozuId(), eventStatus: -1 },
        false
      )
        .then((data) => {
          this.overview = data.list;
        });
    },
    async loadIsHangUpEnable() {
      const data = await window.tnwebServices.customConfigService.loadConfig('AlarmHangUpEnable');
      this.isHangUpEnable = data === '1';
    },
    handleClick(info) {
      // TNBL.redirectUrl(`/timpage/warning-detail?id=${info.AlarmId}`);
      const mozuId = getMozuId();
      window.open(`/tedge/warning-detail?id=${info.alarmId}&mozuId=${mozuId}`);
    },
    handleSizeChange(value) {
      this.limit = value;
      this.queryWarning();
    },
    handleCurrentChange() {
      this.queryWarning();
    },
    contentFilter(kv) {
      Object.keys(kv).forEach((k) => {
        if (!kv[k] || kv[k].length === 0) {
          this.filtered[k] = undefined;
        } else {
          this.filtered[k] = kv[k];
        }
      });
      this.currentPage = 1;
      this.queryWarning();
      this.$refs['content-popover'].doClose();
      this.$refs['reason-popover'].doClose();
    },
    cancelHangUp(row) {
      this.requestCancelHangup([row.alarmId]);
    },
    batchCancelHangup() {
      const alarmIdList = _.map(this.selectedList, 'alarmId');
      this.requestCancelHangup(alarmIdList);
    },
    requestCancelHangup(alarmIdList) {
      this.$confirm('确认要解除告警挂起吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        const userID = String('1000000001');

        // 解除挂起告警不加转发 使用原来封装的axios请求
        this.$axios.post(
          '/cgi/alarm/active/unHangup',
          {
            alarmIdList,
            mozuId: this.mozuId,
            userID,
          },
          false
        )
          .then((data, code) => {
            if (!code) {
              this.$message({
                message: '解除挂起成功',
                type: 'success',
              });
              this.$refs.table.clearSelection();
              this.getOverviewData();
              this.queryWarning();
            }
          });
      })
        .catch(() => {
        });
    },
    filterHandler(kv) {
      Object.keys(kv).forEach((k) => {
        if (!kv[k] || kv[k].length === 0) {
          this.filtered[k] = undefined;
        } else {
          this.filtered[k] = kv[k];
        }
      });
      this.currentPage = 1;
      this.queryWarning();
      // this.popoverVisible = false;
      this.$refs['date-popover'].doClose();
    },
    handleSelectionChange(rows) {
      this.selectedList = rows;
    },
    // search() {
    //   this.searchResultType = 0;
    //   this.currentPage = 1;
    //   this.queryWarning();
    // },
    exportList() {
      const params = this.getParam();
      params.limit = 0;
      params.start = 0;
      getEdgeRequest(this.$axios, this.mozuId).download(`${cgi.exportList}?limit=0&start=0`, params);
      // this.$axios.download(`${cgi.exportList}?limit=0&start=0`, params);
      // window.open(`${cgi.exportList}?limit=0&start=0`);
    },
  },

};
</script>
<style lang="scss">
.overview {
  display: flex;
  width: 100%;
  min-width: 960px;
  background-color: white;
  margin-bottom: 18px;

  &-left {
    width: 200px;
    border-right: 1px solid #f0f0f0;
  }

  &-right {
    flex: 1;
    display: flex;
    // justify-content: space-between;
  }

  .el-data-patch {
    height: 84px;
    padding: 0 24px;
    flex: 1;
    .el-data-patch__value {
      color: #666;
    }
  }

  .overview-right {
    .el-data-patch {
      .el-data-patch__title {
        &:before {
          content: "";
          width: 14px;
          height: 14px;
          border-radius: 50%;
          display: inline-block;
          vertical-align: -14%;
        }
      }
      &:nth-child(1),
      &:nth-child(2) {
        .el-data-patch__title:before {
          background-color: #ff3e00;
        }
      }
      &:nth-child(3) {
        .el-data-patch__title:before {
          background-color: #ff9200;
        }
      }
      &:nth-child(4) {
        .el-data-patch__title:before {
          // background-color: #ffff00;
          background-color: #fbd743;
        }
      }
      &:nth-child(5) {
        .el-data-patch__title:before {
          background-color: #008adc;
        }
      }
      &:nth-child(6) {
        .el-data-patch__title:before {
          background-color: #8acbf2;
        }
      }
    }
  }
}
.tpopover {
  padding-right: 0 !important;
}

@keyframes twinkling {
  /* 透明度由0到1 */
  0% {
    // opacity: 0.2; /* 透明度为0 */
    background-color: #fff;
  }

  100% {
    // opacity: 1; /* 透明度为1 */
    background-color: #ffff00;
  }
}
</style>
