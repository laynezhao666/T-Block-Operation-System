<template>
  <div>
    <div style="display:flex;width:100%">
      <el-title
        v-if="showTitle"
        style="width:200px"
      >
        当前告警
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
    </div>
    <!-- <el-block
      v-if="business.isTedge"
      no-padding
      header-border
    > -->
    <el-collapse
      v-if="business.isTedge"
      v-model="activeName"
      accordion
      @change="changeCollapse"
    >
      <el-collapse-item
        title="一致性 Consistency"
        name="1"
      >
        <template slot="title">
          <div class="chart-title">
            告警时间线
            <el-help-tip
              icon="tn-icon-question"
              style="margin-top: -3px"
            >
              <div style="line-height: 40px">
                <p>统计最近24小时内告警的发生情况（含已恢复）</p>
              </div>
            </el-help-tip>
          </div>
        </template>

        <line-chart
          v-if="showLineChart"
          :data="trendData"
          :collapse-item="activeName"
          height="220px"
        />
      </el-collapse-item>
    </el-collapse>
    <!-- </el-block> -->
    <br>
    <div
      v-if="!business.isTedge"
      class="overview"
    >
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
      <div
        v-if="business.isTedge"
        style="display:flex"
      >
        <template v-if="!isRowsSelected">
          <el-tabs
            v-if="business.isTedge"
            v-model="activeTab"
            class="active-warning-tabs"
            style="height: 56px;flex: 1"
            @tab-click="tabClick"
          >
            <el-tab-pane
              v-for="item in alarmLevelMap"
              :key="item.key"
              :label="levelMapObj[item.text]+'('+item.value+')'"
              :name="item.text.toString()"
            />
          </el-tabs>
          <!-- <img
          src="./setting.svg"
          alt=""
        > -->
          <div style="height: 56px;line-height:56px">
            <!-- <el-tooltip
            class="item"
            effect="dark"
            placement="top-start"
          >
            <div slot="content">
              排序规则设置：
              <br>
              关闭：先等级后时间(默认)
              <br>
              开启：先时间后等级
            </div>

            <el-switch v-model="sortSwitch" />
          </el-tooltip> -->
            <div style="display:flex;align-items: center;">
              <el-button-group class="sort-button-group">
                <el-button
                  size="small"
                  :type="sortSwitch ? 'primary' : 'plain'"
                  @click="changeSort(false)"
                >
                  按时间排序
                </el-button>
                <el-button
                  size="small"
                  :type="!sortSwitch ? 'primary' : 'plain'"
                  @click="changeSort(true)"
                >
                  按等级排序
                </el-button>
              </el-button-group>
              <div
                style="cursor:pointer"
                @click="setModalVisible"
              >
                <el-tooltip
                  placement="top"
                  trigger="hover"
                  content="消息提醒设置"
                >
                  <img
                    src="./assets/setting.svg"
                    alt=""
                    style="transform: translateY(5px);margin-left:20px;"
                  >
                </el-tooltip>
              </div>
            </div>
          </div>

          <el-table-toolbar
            v-model="searchValue"
            style="height:56px;padding-left: 0px"
            class="active-toolbar"
            :actions="[
              {
                text: `导出`,
                icon: 'tn-icon-import',
                action: exportList,
              }
            ]"
            dropdown-width="160"
            filter-placeholder="请输入您需要检索的告警内容"
            :filter-width="226"
            @search="search"
          />
        </template>

        <template
          v-else
        >
          <selection-toolbar
            :selected-list="selectedList"
          >
            <el-button
              type="primary"
              @click="batchHangup"
            >
              批量挂起
            </el-button>
          </selection-toolbar>
        </template>
      </div>
      <el-table-toolbar
        v-else
        v-model="searchValue"
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
        @search="search"
      />

      <!-- 表格 -->
      <el-table
        ref="table"
        :key="isHangUpEnable ? 1 : 0"
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
        <template v-if="business.isTedge">
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
            label="告警类型"
            column-key="alarmType"
            :filter-multiple="true"
            :filters="alarmTypeList"
            prop="alarmType"
            width="200"
          >
            <template slot-scope="scope">
              <el-button
                type="text"
                class="el-button__text-ellipsis"
                @click="handleClick(scope.row)"
              >
                {{ scope.row.alarmType }}
              </el-button>
            </template>
          </el-table-column>
          <el-table-column
            label="告警源"
            min-width="250"
          >
            <template slot-scope="scope">
              <span>
                {{ scope.row.deviceType }}【{{ formatDeviceNumber(scope.row.deviceNumber) }}】
              </span>
            </template>
          </el-table-column>

          <el-table-column
            label="告警原因"
            prop="content"
            min-width="200"
          />
          <el-table-column
            label="定位信息"
            width="250"
          >
            <template slot-scope="scope">
              {{ scope.row.position }}
            </template>
          </el-table-column>

          <el-table-column
            label="触发时间"
            prop="occurTime"
            width="180"
          >
            <template
              slot="header"
              slot-scope="scope"
            >
              <el-popover
                v-model="popoverVisible"
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
                  @change="v => filterHandler({ 'occurTimeStart': v && v[0],'occurTimeEnd':v && v[1] })"
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
            label="持续时间"
            width="220"
          >
            <template slot-scope="scope">
              {{ getRamainTime(scope.row.occurTime) }}
            </template>
          </el-table-column>
          <el-table-column
            v-if="activeTab === '当前所有'"
            label="状态"
          >
            <template slot-scope="scope">
              <el-span
                v-if="scope.row.eventStatus === 1"
                type="danger"
              >
                未转单
              </el-span>
              <!-- <el-tooltip
                v-if="scope.row.eventStatus=== 2"
                class="item"
                effect="dark"
                :content="`于【${scope.row.occurTime}】由服务台转单`"
                placement="left-start"
              > -->
              <el-span
                v-if="scope.row.eventStatus=== 2"
                type="success"
              >
                已转单
              </el-span>
              <!-- </el-tooltip> -->
            </template>
          </el-table-column>
        </template>
        <template v-else>
          <el-table-column
            label="触发时间"
            prop="OccurTime"
            width="170"
          >
            <template
              slot="header"
              slot-scope="scope"
            >
              <el-popover
                v-model="popoverVisible"
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
            prop="Level"
            column-key="Level"
            :filter-multiple="true"
            :filters="levelList"
          >
            <template slot-scope="scope">
              <span :style="getStyle(scope.row.Level)">
                {{ scope.row.Level | LevelMap }}
              </span>
            </template>
          </el-table-column>
          <el-table-column
            label="设备编号"
            prop="DeviceNumber"
            width="350"
            :formatter="formatTableCellDeviceNumber"
          />
          <el-table-column
            label="告警类型"
            column-key="AlarmType"
            :filter-multiple="true"
            :filters="alarmTypeList"
            prop="AlarmType"
          />
          <el-table-column
            label="告警内容"
            prop="Content"
          />
          <el-table-column
            label="触发值"
          >
            <template slot-scope="scope">
              <div
                v-for="(item) in scope.row.OccurPointList"
                :key="item.zhName"
              >
                {{ item.zhName + '：' + item.value + item.unit }}
              </div>
            </template>
          </el-table-column>
          <el-table-column
            label="模组"
            prop="MozuName"
          />
          <el-table-column
            label="房间"
            prop="RoomName"
          />
        </template>

        <el-table-column
          v-if="!business.isTedge || isHangUpEnable"
          label="操作"
          width="100"
        >
          <template slot-scope="scope">
            <el-button
              v-if="business.isTedge && isHangUpEnable"
              type="text"
              size="small"
              @click="hangUpAlarm(scope.row)"
            >
              挂起
            </el-button>

            <el-button
              v-if="!business.isTedge"
              type="text"
              size="small"
              @click="handleClick(scope.row)"
            >
              详情
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
      <audio
        id="au-alarm-alarm0"
        ref="au-alarm-alarm0"
        muted
        src="/static/audio/alarm0.mp3"
        preload="auto"
      />
      <audio
        id="au-alarm-alarm1to2"
        ref="au-alarm-alarm1to2"
        muted
        src="/static/audio/alarm1to2.mp3"
        preload="auto"
      />
      <audio
        id="au-alarm-alarm3to4"
        ref="au-alarm-alarm3to4"
        muted
        src="/static/audio/alarm3to4.mp3"
        preload="auto"
      />

      <detail-modal
        v-if="handleVisible"
        :visible.sync="handleVisible"
        :data="modalData"
      />
      <set-modal
        v-if="setVisible"
        :checked-levels="checkedLevels"
        :levels="levels"
        :visible.sync="setVisible"
        @confirm="confirmSet"
      />
    </el-block>

    <hang-up-modal
      :visible="hangupModalVisible"
      @submit="hangupAlarms"
      @close="closeHangup"
    />
  </div>
</template>
<script>
import { warning as cgi, ba } from '@@/config/cgi';
import { getQueryString } from 'common/script/utils';
import { orderBy } from 'lodash';
import getEdgeRequest from '../../utils/request';
import { getMozuId } from '../../utils/business';
import business from '@@/config/business';
import mixin from 'feature/utils/mixin';
import moment from 'moment';
import lineChart from './line-chart.vue';
import detailModal from './detail-modal.vue';
import setModal from './set-modal.vue';
import HangUpModal from './hang-up-modal.vue';
import SelectionToolbar from 'feature/component/tedge-components/selection-toolbar.vue';
import { NewAlarmsTrendWatcher, AlarmsCountByLevelWatcher, AlarmsWithTotalWatcher } from 'services/tedge/data-watchers/alarms.ts';

export default {
  components: {
    lineChart,
    detailModal,
    setModal,
    HangUpModal,
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
      levels: [
        {
          label: '零级',
          value: 'L0',
        },
        {
          label: '一级',
          value: 'L1',
        },
        {
          label: '二级',
          value: 'L2',
        },
        {
          label: '三级',
          value: 'L3',
        },
        {
          label: '四级',
          value: 'L4',
        },
        // {
        //   label: '五级',
        //   value: 'L5',
        // },

      ],
      setVisible: false,
      eventStatus: 1,
      firstLoad: true,
      latestAlarmTime: null,
      audioList: [],
      business,
      totalItems: 0,
      currentPage: 1,
      limit: 10,
      tableData: [],
      mozuId: 0,
      overview: {},
      prevSearchValue: '',
      searchValue: '',
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
      noticeTimer: null,
      lastToAudioList: [],
      trendData: [],
      lineOption: {},
      showLineChart: false,
      activeTab: '活动',
      handleVisible: false,
      modalData: {},
      activeName: '1',
      levelMapObj: {
        L0: '零级',
        L1: '一级',
        L2: '二级',
        L3: '三级',
        L4: '四级',
        L5: '五级',
        当前所有: '当前所有',
        活动: '活动' },
      dialogFormVisible: false,
      sortSwitch: true,
      checkedLevels: 'L0;L1;L2;L3;L4;',

      isHangUpEnable: false,
      doingHangupAlarm: null,

      selectedList: [],
      hangupModalVisible: false,

      newAlarmsTrendWatcher: new NewAlarmsTrendWatcher(3000)
        .watch({
          mozuId: this.$moduleInfo.mozuId,
        }, (data) => {
          // this.count = data.count;
          this.trendData = Object.keys(data.map).map(key => ({
            // date: key.split(' ')[1].slice(0, -3),
            date: key,
            value: data.map[key],
            // value: data.map[key] + Math.random(1) * 100,
          }));
          this.showLineChart = true;
        })
        .withDiffPlugin()
        .bindVueVm(this),

      // old
      alarmLevelList: [],
      alarmLevelMap: {
        active: {
          text: '活动',
          value: 0,
          key: 0,
        },
        L0: {
          text: 'L0',
          value: 0,
          key: 1,
        },
        L1: {
          text: 'L1',
          value: 0,
          key: 2,
        },
        L2: {
          text: 'L2',
          value: 0,
          key: 3,
        },
        L3: {
          text: 'L3',
          value: 0,
          key: 4,
        },
        L4: {
          text: 'L4',
          value: 0,
          key: 5,
        },
        all: {
          text: '当前所有',
          value: 0,
          key: 6,
        },
      },
      // 不含转单
      alarmsCountByLevelWatcherActive: new AlarmsCountByLevelWatcher(3000)
        .watch({
          mozuId: this.$moduleInfo.mozuId,
          eventStatus: 1,
        }, (data) => {
          this.alarmLevelMap.active.value = data.list.all;

          _.forEach(data.list, (v, k) => {
            if (!this.alarmLevelMap[k] || (k === 'all')) return;
            this.alarmLevelMap[k].value = v;
          });
        })
        .withDiffPlugin()
        .bindVueVm(this),
      // 含转单
      alarmsCountByLevelWatcherActiveAll: new AlarmsCountByLevelWatcher(3000)
        .watch({
          mozuId: this.$moduleInfo.mozuId,
          eventStatus: -1,
        }, (data) => {
          this.alarmLevelMap.all.value = data.list.all;
        })
        .withDiffPlugin()
        .bindVueVm(this),
      // 告警列表
      alarmsWithTotalWatcher: new AlarmsWithTotalWatcher(3000)
        .bindVueVm(this)
        .withDiffPlugin(),
    };
  },
  computed: {
    sortedTypeParam() {
      return this.sortSwitch ? 1 : 0;
    },
    isRowsSelected() {
      return !!this.selectedList?.length;
    },
  },
  watch: {
    sortSwitch(val) {
      const activeSortTypeValue = val ? '1' : '0';
      localStorage.setItem('activeSortType', activeSortTypeValue);
    },
    audioList(list) {
      const that = this;
      if (!list || list.length === 0) return;
      const audio = new Audio(`/static/audio/${list[0]}.mp3`);
      if (audio.paused) {
        audio.muted = true;
        const promise = audio.play();
        // TODO 待产品策略
        if (promise) {
          console.log('调用播放');
          promise.catch(() => {
            console.log('未交互无法播放声音');
          });
        }
        audio.muted = false;
      }
      if (!audio.onended) {
        audio.addEventListener('ended', () => {
          that.audioList.shift();
        });
      }
    },
  },
  mounted() {
    this.getLocalStorage();
    const deviceNo = getQueryString('deviceNo');
    if (deviceNo) {
      this.searchValue = deviceNo;
    }
    if (business.showModuleSelected) {
      this.mozuId = parseInt(TNBL.getCurModuleId()) || parseInt(getQueryString('mozuId'));
    }
    this.getOverviewData();
    this.queryWarning();
    this.loadIsHangUpEnable();
    // if (business.isTedge) this.queryWarningNotice();
  },
  beforeDestroy() {
    clearTimeout(this.timer);
    // (new BroadcastChannel('alarmBoradCast')).postMessage({
    //   event: 'beginWarning',
    // });s
    if (business.isTedge) clearTimeout(this.noticeTimer);
  },
  methods: {
    formatDeviceNumber(deviceNumber) {
      return window.tnwebServices.v2DeviceNumberTransformerService.get(deviceNumber);
    },
    formatTableCellDeviceNumber(row, col, deviceNumber) {
      return window.tnwebServices.v2DeviceNumberTransformerService.get(deviceNumber, true);
    },

    reloadData() {
      this.getOverviewData();
      this.queryWarning();
    },
    async loadIsHangUpEnable() {
      const data = await window.tnwebServices.customConfigService.loadConfig('AlarmHangUpEnable');
      this.isHangUpEnable = data === '1';
    },

    setModalVisible() {
      const notifyLevelsInLocalStorage = localStorage.getItem('notifyLevels');
      this.checkedLevels = _.isNil(notifyLevelsInLocalStorage)
        ? 'L0;L1;L2;L3;L4'
        : notifyLevelsInLocalStorage;
      this.setVisible = true;
    },
    confirmSet(data) {
      localStorage.setItem('notifyLevels', data.notifyLevels);
      this.$message.success('设置成功');
      setTimeout(() => {
        location.reload();
      }, 0);
    },
    changeSort(val) {
      this.sortSwitch = !val;
      this.queryWarning();
    },
    changeCollapse(activeNames) {
      localStorage.setItem('collapseStatus', activeNames);
    },
    getLocalStorage() {
      if (localStorage.getItem('activeSortType')) this.sortSwitch = localStorage.getItem('activeSortType') !== '0';
      if (!_.isNil(localStorage.getItem('notifyLevels'))) {
        this.checkedLevels = localStorage.getItem('notifyLevels');
      }

      this.activeName = localStorage.getItem('collapseStatus');
    },
    getRamainTime(time) {
      function formatDuring(mss) {
        const days = parseInt(mss / (1000 * 60 * 60 * 24));
        const hours = parseInt((mss % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        const minutes = parseInt((mss % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = (mss % (1000 * 60)) / 1000;
        return `${days} 天 ${hours} 小时 ${minutes} 分钟 ${seconds.toFixed(0)} 秒 `;
      }

      const nowTime = parseInt(new Date().getTime());
      const currTime = parseInt(new Date(time).getTime());
      return formatDuring(nowTime - currTime);
    },
    tabClick(val) {
      this.eventStatus = val.name === '当前所有' ? -1 : 1;
      const level = ['当前所有', '活动'].includes(val.name) ? { level: [] } : { level: [val.name] };
      this.getOverviewData();
      this.filterHandler(level);
    },
    getHoursAlarmTrend(occurTimeStart, occurTimeEnd) {
      this.$axios
        .post(ba.getHoursAlarmTrend, { mozuId: getMozuId(), occurTimeStart, occurTimeEnd }, false)
        .then((data) => {
          // this.count = data.count;
          this.trendData = Object.keys(data.map).map(key => ({
            // date: key.split(' ')[1].slice(0, -3),
            date: key,
            value: data.map[key],
            // value: data.map[key] + Math.random(1) * 100,
          }));
          this.showLineChart = true;
        });
    },
    getOverviewData() {
      const mozuId = getMozuId() || this.mozuId;
      getEdgeRequest(this.$axios, mozuId).post(
        cgi.getAlarmType,
        { start: 0, limit: 0, mozuId, eventStatus: this.eventStatus }
      )
        .then((data) => {
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
    getParam(mozuId) {
      const start = (this.currentPage - 1) * this.limit;
      const parm = {
        start,
        limit: this.limit,
        mozuId,
        eventStatus: this.eventStatus,
      };
      Object.assign(parm, this.filtered);
      if (this.searchValue) {
        if (this.searchResultType === 0 || this.searchResultType === 1) {
          parm.DeviceNumber = [this.searchValue];
          this.searchResultType = 1;
        } else {
          parm.Content = [];
          parm.Content[0] = this.searchValue;
        }
      }

      return parm;
    },
    queryWarningNotice() {
      this.noticeTimer = setInterval(() => {
        const mozuId = getMozuId() || this.mozuId;
        const nowBefore3Second = moment(Date.now()).subtract(3, 'second')
          .format('YYYY-MM-DD HH:mm:ss');
        const nowBefore6Second = moment(Date.now()).subtract(6, 'second')
          .format('YYYY-MM-DD HH:mm:ss');
        getEdgeRequest(this.$axios, mozuId).post(
          cgi.getActivedWarning,
          {
            start: 0,
            limit: 100,
            mozuId,
            eventStatus: this.eventStatus,
            OccurTimeEnd: nowBefore3Second,
            OccurTimeStart: nowBefore6Second,
          },
          false
        )
          .then((data) => {
            this.addToAudioList(data);
          });
      }, 3000);
    },
    queryWarning() {
      if (this.timer) {
        clearTimeout(this.timer);
        this.timer = null;
      }
      const mozuId = getMozuId() || this.mozuId;
      const parm = this.getParam(mozuId);
      parm.sortedType = this.sortedTypeParam;

      // console.log('parm', parm);
      // getEdgeRequest(this.$axios, mozuId).post(
      //   cgi.getActivedWarning, parm,
      //   false
      // )
      this.alarmsWithTotalWatcher.mockRequest(parm)
        .then((data) => {
          this.totalItems = data.count;
          this.tableData = data.list;
          this.tableData.forEach((item) => {
            this.$set(item, 'OccurPointList', orderBy(item.OccurPointList, 'enName'));
          });
          // (new BroadcastChannel('alarmBoradCast')).postMessage({
          //   event: 'activeWarning',
          //   total: this.totalItems,
          // });
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

      // 获取24小时告警趋势
      if (business.isTedge) {
        // this.getOverview(mozuId);
        // 改为watcher
        // this.getHoursAlarmTrend();
      } else {
        // overview
        getEdgeRequest(this.$axios, mozuId).post(
          cgi.getActivedOverview, { mozuId, eventStatus: -1 },
          '',
          false
        )
          .then((data) => {
            this.overview = data.list;
          });
      }
    },
    getOverview(mozuId) {
      const fn = (eventStatus, mozuId) => getEdgeRequest(this.$axios, mozuId).post(
        cgi.getActivedOverview, { mozuId, eventStatus },
        '',
        false
      );
      Promise.all([fn(1, mozuId), fn(-1, mozuId)]).then((values) => {
        const [a, b] = values;
        // 活动 一级  二级 三级 四级 这几个都拉未转单的告警(1)  当前所有加到最后面 拉所有的告警包含转单的(-1)
        // All EventStatus = -1  未转事件 EventStatus = 1  已经转事件 EventStatus = 2  事件已经结单 EventStatus = 3
        a.list.active = a.list.all;
        const levelArray = { active: '活动', L0: 'L0', L1: 'L1', L2: 'L2', L3: 'L3', L4: 'L4', L5: 'L5' };
        const levelList = [];
        Object.keys(levelArray).forEach((i) => {
          if (a.list[i] > 0 || levelArray[i] === '活动') {
            levelList.push({ text: levelArray[i],
              value: a.list[i] || 0,
              key: i });
          }
        });
        if (b.list.all !== 0) { // 当前所有
          levelList.push({ text: '当前所有',
            value: b.list.all,
            key: 6 });
        }
        this.alarmLevelList = levelList;
      });
    },
    addToAudioList(data) {
      const levelMap = {
        L0: 0,
        L1: 1,
        L2: 2,
        L3: 3,
        L4: 4,
        L5: 5,
      };
      const lastAlarmId = this.lastToAudioList.map(item => item.AlarmId);

      const toAudioList = data.list.sort((a, b) => levelMap[a.Level] - levelMap[b.Level])
        .filter(item => !lastAlarmId.includes(item.AlarmId));
      function countByItem(data, item) {
        const result = data.reduce((acc, value) => {
          if (!acc[value[item]]) {
            acc[value[item]] = 1;
          } else {
            acc[value[item]] = acc[value[item]] + 1;
          }
          return acc;
        }, {});
        return result;
      }
      this.lastToAudioList = toAudioList;
      const groupByLevel = countByItem(toAudioList, 'Level');
      Object.keys(groupByLevel).forEach((item) => {
        let num = 0;
        const level = item;
        if (level === 'L0') {
          num = 0;
        } else if (level === 'L1' || level === 'L2') {
          num = 1;
        } else if (level === 'L3' || level === 'L4') {
          num = 2;
        }
        this.audioPlay(num);
      });
    },
    async hangUpAlarm(alarm) {
      this.doingHangupAlarm = alarm;
      this.hangupModalVisible = true;
    },
    handleClick(info) {
      if (business.isTedge) {
        this.modalData = info;
        this.modalData.mozuId = getMozuId();
        this.handleVisible = true;
      } else {
        const mozuId = getMozuId() || this.mozuId;
        if (info.alarm_id_string) { // alarm_id_string存在增加1.0标记
          window.open(`/${business.moduleName}/warning-detail?id=${info.AlarmId}&mozuId=${mozuId}&alarm_id_string=${info.alarm_id_string}`);
        } else {
          const mozuId = getMozuId() || this.mozuId;
          if (info.alarm_id_string) { // alarm_id_string存在跳转1.0
            const url = ``;
            window.open(url, '_blank');
          } else {
            window.open(`/${business.moduleName}/warning-detail?id=${info.AlarmId}&mozuId=${mozuId}`);
          }
        }
      }
    },
    handleSizeChange(value) {
      this.limit = value;
      this.queryWarning();
    },
    handleCurrentChange() {
      this.queryWarning();
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
      this.popoverVisible = false;
    },
    search() {
      this.searchResultType = 0;
      this.currentPage = 1;
      this.queryWarning();
    },
    exportList() {
      const mozuId = getMozuId() || this.mozuId;
      const params = this.getParam(mozuId);
      params.limit = 0;
      params.start = 0;
      getEdgeRequest(this.$axios, mozuId).download(`${cgi.exportList}?limit=0&start=0`, params);
      // window.open(`${cgi.exportList}?limit=0&start=0`);
    },
    audioPlay(num) {
      const focAudioMap = { // FOC广播类型
        0: 'alarm0',
        1: 'alarm1to2',
        2: 'alarm3to4',
      };
      const src = focAudioMap[num];
      if (this.audioList.indexOf(src) === -1) {
        this.audioList.push(src);
      };
    },
    async hangupAlarms(formData) {
      const alarmIdList = this.doingHangupAlarm
        ? [String(this.doingHangupAlarm.alarmId)]
        : _.map(this.selectedList, item => String(item.alarmId));
      if (!alarmIdList.length) return;

      const userID = String('1000000001');

      await this.$axios.post('/cgi/alarm/active/hangup', {
        mozuId: TNBL.getCurrModule().id,
        alarmIdList,
        userID,
        hangupReason: formData.hangupReason,
      });
      this.$message.success('挂起成功');
      this.reloadData();

      this.closeHangup();
    },
    batchHangup() {
      this.hangupModalVisible = true;
    },
    closeHangup() {
      this.hangupModalVisible = false;
      this.doingHangupAlarm = null;
      this.selectedList = [];
      this.$refs.table.clearSelection();
    },
    handleSelectionChange(rows) {
      this.selectedList = rows;
    },
  },

};
</script>
<style lang="scss" scoped>
.sort-button-group{
  /deep/ .el-button{
    padding: 3px 10px;
  }
}
.chart-title {
  display:flex;
  padding:0px 0 0 16px;
  color:#000;
  font-weight:600;
  font-size:16px;
}

.el-button__text-ellipsis {
  width: 100%;
  /deep/ span {
    display: block !important;
    width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}
</style>

<style lang="scss">
.overview {
  display: flex;
  width: 100%;
  min-width: 960px;
  background-color: white;
  margin-bottom: 18px;
  box-shadow:   3px 1px 8px 2px rgba(218, 218, 218, 0.5);
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
   /deep/ .el-data-patch {
      .el-data-patch__title {
        &:before {
          content: '';
          width: 14px;
          height: 14px;
          border-radius: 50%;
          display: inline-block;
          vertical-align: -14%;
          margin-right:5px;
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
    }
  }
}

.active-warning-tabs {
  // .el-tabs__item {
  //   height: 66px;
  //   line-height: 66px
  // }
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
