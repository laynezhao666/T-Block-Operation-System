<template>
  <el-modal
    :visible.sync="logVisible"
    @close="close"
    @opened="openChart"
  >
    <template slot="title">
      告警详情
    </template>
    <el-block
      padding
      header-border
    >
      <template slot="header">
        <el-tag
          type="warning"
          style="padding:0 16px 0px 16px;margin-right:10px;"
        >
          <span style="margin-bottom:0px!important">{{
            levelMap[data.alarmLevel] }}</span>
        </el-tag>
        {{ data.alarmType }}
      </template>
      <el-form label-width="100px">
        <el-form-item label="告警源：">
          <el-button
            type="text"
            @click="JumpOrigin(data)"
          >
            {{ data.deviceType }}【{{ formatDeviceNumber(data.deviceNumber) }}】
          </el-button>
        </el-form-item>
        <el-form-item label="定位：">
          <el-button type="text">
            {{ data.position }}
          </el-button>
        </el-form-item>
        <el-form-item label="触发原因">
          {{ data.alarmContent }}
        </el-form-item>
        <el-form-item
          label="状态"
          class="custom-el-form-item"
        >
          <el-tag
            :type="data.eventStatus === 2 ? 'success' : 'danger'"
            effect="plain"
          >
            {{ data.eventStatus === 2 ? '已转单' : '未转单' }}
          </el-tag>
        </el-form-item>
        <el-form-item label="处理建议">
          {{ dtl.alarm.DealSuggestion || '无' }}
        </el-form-item>
      </el-form>
    </el-block>
    <br>
    <el-block
      no-padding
      header-border
    >
      <el-title slot="header">
        运行数据
        <span slot="extra">
          <el-date-picker
            v-model="selDateTime"
            style="width: 430px;"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            :picker-options="pickerOptions"
            type="datetimerange"
            @change="selectDT"
          />
        </span>
      </el-title>
      <div
        v-if="typeList.length > 1"
        style="padding: 16px 0 0"
      >
        <el-radio-group
          v-model="type"
          size="small"
        >
          <el-radio-button
            v-for="item in typeList"
            :key="item"
            :label="item"
          >
            {{ item }}
          </el-radio-button>
        </el-radio-group>
      </div>
      <!-- <div
        ref="eChart"
        class="run-data"
      />  -->
      <div style="height:300px;margin:0px 20px;overflow:hidden">
        <customize-line-chart
          v-if="showChart"
          ref="mychart"
          :x-axis="xAxis"
          :series="chartSeries"
          :legend="legend"
          :tooltip="{
            ignoreNil: false
          }"
        />
      </div>
    </el-block>
    <br>
    <el-block
      no-padding
      header-border
    >
      <template slot="header">
        相关性告警
      </template>
      <el-tabs
        v-model="activeRelation"
        @tab-click="clickRelation"
      >
        <el-tab-pane
          label="该告警相关"
          name="alarmTypeRelation"
        />
        <el-tab-pane
          label="该设备相关"
          name="deviceRelation"
        />
        <el-table
          :data="tableData"
          style="width: 100%;overflow:scroll"
          @filter-change="filterHandler"
        >
          <el-table-column
            label="告警等级"
            width="100"
            prop="alarmLevel"
          >
            <template slot-scope="scope">
              <span :style="getStyle(scope.row.alarmLevel)">
                {{ scope.row.alarmLevel | LevelMap }}
              </span>
            </template>
          </el-table-column>

          <el-table-column
            label="告警类型"
            column-key="alarmType"
            prop="alarmType"
            width="150"
          />

          <el-table-column
            label="告警源"
            width="300"
          >
            <template slot-scope="scope">
              <span>
                {{ scope.row.deviceType }}【{{ formatDeviceNumber(scope.row.deviceNumber) }}】
              </span>
            </template>
          </el-table-column>
          <el-table-column
            label="告警内容"
            prop="alarmContent"
            width="200"
          />
          <el-table-column
            label="定位信息"
            width="200"
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
                popper-class="tpopover"
                trigger="click"
              >
                <el-date-picker
                  v-model="timeRange"
                  type="datetimerange"
                  value-format="yyyy-MM-dd HH:mm:ss"
                  range-separator="至"
                  start-placeholder="开始日期"
                  end-placeholder="结束日期"
                  align="right"
                  @change="v => filterHandler({ 'occurTimeStart': v && v[0], 'occurTimeEnd': v && v[1] })"
                />
                <span slot="reference">
                  <span :style="{ color: timeRange ? '#1470cc' : '' }">{{ scope.label || '触发时间' }}</span>
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
            label="恢复时间"
            prop="restoreTime"
            width="180"
          >
            <template
              slot="header"
              slot-scope="scope"
            >
              <el-popover
                v-model="restore_popoverVisible"
                placement="bottom"
                width="400"
                popper-class="tpopover"
                trigger="click"
              >
                <el-date-picker
                  v-model="restore_timeRange"
                  type="datetimerange"
                  value-format="yyyy-MM-dd HH:mm:ss"
                  range-separator="至"
                  start-placeholder="开始日期"
                  end-placeholder="结束日期"
                  align="right"
                  @change="v => filterHandler({ 'restoreTimeStart': v && v[0], 'restoreTimeEnd': v && v[1] })"
                />
                <span slot="reference">
                  <span :style="{ color: restore_timeRange ? '#1470cc' : '' }">{{ scope.label || '恢复时间' }}</span>
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
            label="关闭时间"
            prop="closeTime"
          >
            <template
              slot="header"
              slot-scope="scope"
            >
              <el-popover
                v-model="closeTime_popoverVisible"
                placement="bottom"
                width="400"
                popper-class="tpopover"
                trigger="click"
              >
                <el-date-picker
                  v-model="closeTime_timeRange"
                  type="datetimerange"
                  value-format="yyyy-MM-dd HH:mm:ss"
                  range-separator="至"
                  start-placeholder="开始日期"
                  end-placeholder="结束日期"
                  align="right"
                  @change="v => filterHandler({ 'closeTimeStart': v && v[0], 'closeTimeEnd': v && v[1] })"
                />
                <span slot="reference">
                  <span :style="{ color: closeTime_timeRange ? '#1470cc' : '' }">{{ scope.label || '关闭时间' }}</span>
                  <!--这里改结构到和其他filter的th一样-->
                  <span class="el-table__column-filter-trigger">
                    <!--无法判断状态和修改class：https://github.com/ElemeFE/element/issues?page=2&q=table+date+picker&utf8=%E2%9C%93-->
                    <i class="el-icon-caret-bottom" />
                  </span>
                </span>
              </el-popover>
            </template>
          </el-table-column>
          <el-table-column label="持续时长">
            <template slot-scope="scope">
              {{ calDuraTime(scope.row) }}
            </template>
          </el-table-column>
          <el-table-column
            label="关闭人"
            column-key="closeOperatorName"
            prop="closeOperatorName"
          >
            <template slot-scope="scope">
              {{ formatCloseOperatorName(scope.row.closeOperatorName) }}
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
      </el-tabs>
    </el-block>
  </el-modal>
</template>

<script>
import { required } from 'common/script/form_rules';
import dayjs from 'dayjs';
import getEdgeRequest from '../../utils/request';
import { warning as cgi, tbosWarning as tbosCgi } from '@@/config/cgi';
import business from '@@/config/business';
import CustomizeLineChart from 'feature/component/chart/line-chart';
import duration from 'dayjs/plugin/duration';
import { has } from 'lodash';

dayjs.extend(duration);
function floor(v) {
  return Math.floor(v.toFixed(2));
}
export default {
  components: {
    CustomizeLineChart,
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
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    cgi: {
      type: Object,
      default: () => ({}),
    },
    data: {
      type: Object,
      default: () => ({}),
    },
    view: {
      type: String,
      default: '',
    },
    status: {
      type: String,
      default: 'add',
    },
  },
  data() {
    const endDate = (+new Date());
    const startDate = endDate - (3600 * 1000);
    return {
      levelMap: {
        L0: '零级',
        L1: '一级',
        L2: '二级',
        L3: '三级',
        L4: '四级',
        L5: '五级',
      },
      defaultButton: '按设备类型',
      buttonOptions: [
        { name: '按设备类型', value: '' },
        { name: '按告警设备', value: '' },
        { name: '按告警类型', value: '' },
        { name: '按时间段', value: '' },
      ],
      emissionSeries: { title: '', maintitleLeft: 0, titleLeft: 40, legendX: 260 },
      emissionComposition: [{ value: 2, name: '零级' },
        { value: 7, name: '一级' },
        { value: 4, name: '二级' },
        { value: 5, name: '三级' },
        { value: 6, name: '四级' }],
      form: {
        onResult: '',
        failureReason: '',
      },
      showFailureReason: false,
      rules: {
        onResult: [required()],
        failureReason: [required()],
      },
      selDateTime: [startDate, endDate],
      mychart: null,
      pickerOptions: {
        shortcuts: [
          {
            text: '最近半小时',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (1800 * 1000));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '最近一小时',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (3600 * 1000));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '最近八小时',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (3600 * 1000 * 8));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '最近一天',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (3600 * 1000 * 24));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '默认',
            onClick: (picker) => {
              const end = dayjs(this.dtl.alarm.OccurTime).add(10, 'm');
              const start = dayjs(this.dtl.alarm.OccurTime).subtract(10, 'm');
              picker.$emit('pick', [start, end]);
            },
          },
        ],
      },
      type: '',
      typeList: [],
      id: '',
      chartSeries: [],
      chartTimes: [],
      yAxis: {},
      xAxis: {},
      legend: {
        top: '5%',
        left: 'left',
      },
      chartPointData: {},
      dtl: {
        deviceInfo: {},
        alarm: {},
        ruleInfo: {},
      },
      showChart: false,
      totalItems: 0,
      currentPage: 1,
      limit: 10,
      tableData: [],
      closeTime_popoverVisible: false,
      closeTime_timeRange: null,
      restore_popoverVisible: false,
      restore_timeRange: null,
      popoverVisible: false,
      timeRange: null,
      filtered: {
      },
      activeRelation: 'alarmTypeRelation',
    };
  },
  computed: {
    logVisible: {
      set(v) {
        this.$emit('update:visible', v);
      },
      get() {
        return this.visible;
      },
    },
  },
  watch: {
    type(val, oldval) {
      if (oldval !== '') {
        this.fetchChartData();
      }
    },
  },
  mounted() {

  },
  provide() {
    return {
      commonCgi: this.cgi.commonCgi,
    };
  },
  beforeDestroy() {
    this.mychart = this.$refs.mychart;
    const echart = this.mychart.chart;
    echart.off('legendselectchanged');
    this.mychart = null;
  },
  methods: {
    formatDeviceNumber(deviceNumber) {
      return window.tnwebServices.v2DeviceNumberTransformerService.get(deviceNumber, true);
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
    formatCloseOperatorName(val) {
      return val === '0' ? '' : val;
    },
    JumpOrigin(data) {
      window.open(`/tedge/data-query-index?devId=${data.deviceGid}`);
    },
    clickRelation() {
      this.fetchHistory();
    },
    calDuraTime(row) {
      const occurTime = dayjs(row.occurTime);
      let endTime;
      if (row.restoreTime) {
        endTime = dayjs(row.restoreTime);
      } else if (row.closeTime) {
        endTime = dayjs(row.closeTime);
      } else {
        endTime = dayjs();
      }
      let seconds = endTime.diff(occurTime) / 1000;
      const days = floor(seconds / (24 * 3600));
      seconds = seconds % (24 * 3600);
      const hours = floor(seconds / 3600);
      seconds = seconds % 3600;
      const mins = floor(seconds / 60);
      seconds = seconds % 60;
      return `${days ? `${days}天` : ''}${hours ? `${hours}小时` : ''}${mins ? `${mins}分钟` : ''}${seconds ? `${seconds}秒` : ''}`;
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
      this.fetchHistory();
      this.popoverVisible = false;
    },
    handleSizeChange(value) {
      this.limit = value;
      this.fetchHistory();
    },
    handleCurrentChange() {
      this.fetchHistory();
    },
    fetchHistory() {
      const offset = (this.currentPage - 1) * this.limit;
      const parm = {
        offset,
        limit: this.limit,
        // alarmType: this.dtl.alarm.AlarmType,
        // deviceGid: this.dtl.alarm.DeviceGid,
        // fingerprint: this.dtl.alarm.Fingerprint,
      };
      Object.assign(parm, this.filtered);
      if (this.activeRelation === 'deviceRelation') {
        parm.deviceGid = this.dtl.alarm.DeviceGid;
      } else {
        parm.fingerprint = this.dtl.alarm.Fingerprint;
      }
      getEdgeRequest(this.$axios, this.mozuId).post(cgi.getWarningHistory, parm,)
        .then((data) => {
          this.totalItems = data.count;
          this.tableData = data.list;
        });
    },
    openChart() {
      this.id = this.data.alarmId;
      this.mozuId = this.data.mozuId;
      this.fetchInfo();
    },
    fetchInfo() {
      getEdgeRequest(this.$axios, this.mozuId).post(cgi.getWarningDetail, {
        AlarmId: this.id,
        MozuId: this.mozuId,
      })
        .then((data) => {
          if (data && data.detail) {
            this.dtl = data.detail;
            this.fetchHistory();
            const occurTime = data.detail.alarm.OccurTime;
            if (occurTime) {
              this.selDateTime = [dayjs(occurTime).subtract(10, 'm')
                .format('YYYY-MM-DD HH:mm:ss'), dayjs(occurTime).add(10, 'm')
                .format('YYYY-MM-DD HH:mm:ss')];
              if (this.$moduleInfo?.isTbos) {
                this.fetchChartData();
              } else {
                getEdgeRequest(this.$axios, this.mozuId).post(cgi.getPointDataType, {
                  AlarmId: this.id,
                  MozuId: this.mozuId,
                })
                  .then((data) => {
                    if (data && data.list.length > 0) {
                      console.log('pointDatatype不存在');
                      // eslint-disable-next-line prefer-destructuring
                      this.type = data.list[0];
                      this.typeList = data.list;
                      this.fetchChartData();
                    }
                  });
              }
            };
          }
        });
    },
    async fetchChartData() {
      let data = null;
      if (this.$moduleInfo?.isTbos) {
        const result = await getEdgeRequest(this.$axios, this.mozuId).post(tbosCgi.pointQuery, {
          start_time: dayjs(this.selDateTime[0]).format('YYYY-MM-DD HH:mm:ss'),
          end_time: dayjs(this.selDateTime[1]).format('YYYY-MM-DD HH:mm:ss'),
          conditions: [{
            name: 'point_key',
            value: this.dtl?.points || [],
          }],
          data_type: 1,
          interval: 1,
        });
        const keyMap = {
          collectorPoint: 'collector_point',
          deviceGID: 'device_gid',
          deviceName: 'device_type_zh',
          point: 'point_key',
          // pointDataList: 'point_data',
          pointName: 'point_name_en',
          pointZhName: 'point_name_zh',
        };
        const list = result?.list.map((i) => {
          const item = {};
          // eslint-disable-next-line no-restricted-syntax
          for (const key in keyMap) {
            if (has(i, keyMap[key])) {
              item[key] = i[keyMap[key]];
            }
          }
          item.pointDataList = i?.point_data.map(j => ({
            time: j.update_time,
            val: j.value,
          })) || [];
          return {
            ...item,
          };
        });
        data = { list };
        console.log(data, 'isTbos');
      } else {
        getEdgeRequest(this.$axios, this.mozuId).post(cgi.getPointData, {
          AlarmId: this.id,
          MozuId: this.mozuId,
          PointType: this.type,
          StartTime: dayjs(this.selDateTime[0]).format('YYYY-MM-DD HH:mm:ss'),
          EndTime: dayjs(this.selDateTime[1]).format('YYYY-MM-DD HH:mm:ss'),
        });
      }
      const times = [];
      if (data.list.length > 0) {
        const chartSeries = [];
        // eslint-disable-next-line no-unused-expressions, babel/no-unused-expressions
        data.list[0].pointDataList?.forEach((item) => {
          times.push(item.time);
        });
        const categorySeries = {};
        let ischeck; let min;
        let max;
        if (this.type === '状态量') {
          this.yAxis = {
            type: 'category',
            boundaryGap: false,
            axisLine: {
              show: true,
              lineStyle: {
                color: '#ccc',
              },
            },
          };
          categorySeries.step = 'end';
          categorySeries.smooth = false;
          this.xAxis.axisLine = {
            show: true,
            lineStyle: {
              color: '#ccc',
            },
          };
        } else {
          ischeck = true;
          this.yAxis = {
            type: 'value',
          };
        }
        data.list.forEach((pointdata) => {
          const series = [];
          // eslint-disable-next-line no-unused-expressions, babel/no-unused-expressions
          pointdata.pointDataList?.forEach((item) => {
            series.push(item.val);
            if (ischeck) {
              if (min === undefined) {
                min = item.val;
              } else {
                min = item.val < min ? item.val : min;
              }
              if (max === undefined) {
                max = item.val;
              } else {
                max = item.val > max ? item.val : max;
              }
            }
          });
          const markLine = {
            symbol: 'none',
            data: [
              {
                silent: true,
                lineStyle: {
                  type: 'dashed',
                  width: 2,
                  color: '#ff9200',
                },
                xAxis: this.dtl.alarm.OccurTime,
              },
            ],
          };

          chartSeries.push({ name: pointdata.pointZhName, data: series, ...categorySeries, markLine });
          this.chartPointData[pointdata.pointZhName] = pointdata;
        });
        if (ischeck) {
          this.yAxis = {
            ...this.yAxis,
            min: min * 0.9,
            max: max * 1.1,
          };
        }
        this.chartSeries = chartSeries;
        this.xAxis.data = times;
      } else {
        this.chartSeries = [];
        this.xAxis.data = [];
      }
      this.showChart = true;
      this.$nextTick(() => {
        this.mychart = this.$refs.mychart;
        const echart = this.mychart.chart;
        echart.on('legendselectchanged', (params) => {
          const pd = this.chartPointData[params.name];
          if (pd) { // 打开新的页面
            const start = dayjs(this.selDateTime[0]).format('YYYY-MM-DD HH:mm:ss');
            const end = dayjs(this.selDateTime[1]).format('YYYY-MM-DD HH:mm:ss');
            const url = `/${business.moduleName}/data-query-detail?id=${pd.point}&devName=${pd.deviceName}&devId=${pd.deviceGID}&start=${start}&end=${end}&mozuId=${this.mozuId}`;
            window.open(url);
          }
          echart.dispatchAction({
            type: 'legendAllSelect',
          });
        });
      });
    },
    selectDT() {
      this.fetchChartData();
    },
    close() {
      this.logVisible = false;
      // this.$refs.form.resetFields();
      this.form = {};
      this.$emit('close');
    },
    logout() {
      console.log(this.tag);
    },
    confirm() {
      this.$refs.form.validate((valid) => {
        if (valid) {
          this.$emit('confirm', { result: this.form.onResult, failureReason: this.form.failureReason });
          this.$message.success('设置成功');
          this.close();
        }
      });
    },
  },
};
</script>
<style></style>
