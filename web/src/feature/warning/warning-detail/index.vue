<template>
  <div>
    <el-title>
      告警详情
    </el-title>

    <el-space
      direction="vertical"
      size="middle"
    >
      <el-block>
        <el-block
          inner
          header="告警等级"
        >
          <el-descriptions>
            <el-descriptions-item
              label=""
              span="2"
            >
              {{ dtl.ruleInfo.alarmLevelZh }}
            </el-descriptions-item>
          </el-descriptions>
        </el-block>
        <el-block
          inner
          header="告警描述"
        >
          <el-descriptions>
            <el-descriptions-item
              label="告警描述"
              span="2"
            >
              {{ desc }}
            </el-descriptions-item>
          </el-descriptions>
        </el-block>

        <el-block
          inner
          header="告警影响"
        >
          <el-descriptions>
            <el-descriptions-item
              label="影响分析"
              span="2"
            >
              {{ dtl.alarm.InfluenceAnalyze }}
            </el-descriptions-item>
          </el-descriptions>
        </el-block>

        <el-block
          inner
          header="处理建议"
        >
          <el-descriptions>
            <el-descriptions-item
              label=""
              span="2"
            >
              {{ dtl.alarm.DealSuggestion || '无' }}
            </el-descriptions-item>
          </el-descriptions>
        </el-block>
      </el-block>

      <el-block
        header="设备信息"
      >
        <el-descriptions>
          <el-descriptions-item label="设备编号">
            <span v-if="fromMon || isTEdge">{{ dtl.deviceInfo.DeviceNumber }}</span>
            <a
              v-else
              style="color:rgb(20, 112, 204);cursor: pointer;"
              :href="dealJumpUrl()"
              target="_blank"
              class="f-blue"
            >{{ dtl.deviceInfo.DeviceNumber }}</a>
          </el-descriptions-item>
          <el-descriptions-item label="设备名称">
            {{ dtl.deviceInfo.DeviceName }}
          </el-descriptions-item>
          <el-descriptions-item label="维保厂商">
            {{ dtl.deviceInfo.DeviceMaintenanceCompany }}
          </el-descriptions-item>
          <el-descriptions-item label="生产厂商">
            {{ dtl.deviceInfo.DeviceManufacturer }}
          </el-descriptions-item>
          <el-descriptions-item label="启用时间">
            {{ dtl.deviceInfo.DeviceActivationTime }}
          </el-descriptions-item>
          <el-descriptions-item label="运维团队">
            {{ dtl.deviceInfo.MozuInfrastructureCompany }}
          </el-descriptions-item>
        </el-descriptions>
      </el-block>

      <el-block
        header="告警阈值"
      >
        <el-descriptions>
          <el-descriptions-item label="告警触发表达式">
            {{ dtl.ruleInfo.alarmExpressionStr }}
          </el-descriptions-item>
          <el-descriptions-item label="告警恢复表达式">
            {{ dtl.ruleInfo.restoreExpressionStr }}
          </el-descriptions-item>
        </el-descriptions>
      </el-block>

      <el-block
        v-if="fromMon"
        header="运行数据"
      >
        <el-descriptions>
          <el-descriptions-item>
            <el-button
              type="text"
              @click="jumpToMon"
            >
              查看告警点数据
            </el-button>
          </el-descriptions-item>
        </el-descriptions>
      </el-block>

      <el-block
        v-else
        padding
        header="运行数据"
      >
        <span
          slot="extra"
        >
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
        <customize-line-chart
          ref="mychart"
          :x-axis="xAxis"
          :series="chartSeries"
          :legend="legend"
          :tooltip="{
            ignoreNil: false
          }"
        />
      </el-block>

      <el-block header="告警历史记录">
        <el-table
          :data="tableData"
          style="width: 100%"
          @filter-change="filterHandler"
        >
          <el-table-column
            label="触发时间"
            prop="occurTime"
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
            label="恢复时间"
            prop="restoreTime"
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
                  @change="v => filterHandler({ 'restoreTimeStart': v && v[0],'restoreTimeEnd':v && v[1] })"
                />
                <span slot="reference">
                  <span :style="{color:restore_timeRange?'#1470cc':''}">{{ scope.label || '恢复时间' }}</span>
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
                  @change="v => filterHandler({ 'closeTimeStart': v && v[0],'closeTimeEnd':v && v[1] })"
                />
                <span slot="reference">
                  <span :style="{color:closeTime_timeRange?'#1470cc':''}">{{ scope.label || '关闭时间' }}</span>
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
          />
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
    </el-space>
  </div>
</template>
<script>
import qs from 'qs';
import dayjs from 'dayjs';
import duration from 'dayjs/plugin/duration';
import { has, isEmpty } from 'lodash';
import { warning as cgi, tbosWarning as tbosCgi } from '@@/config/cgi';
import getEdgeRequest from '../../utils/request';
import business from '@@/config/business';
import CustomizeLineChart from 'feature/component/chart/line-chart';

dayjs.extend(duration);
function floor(v) {
  return Math.floor(v.toFixed(2));
}
export default {
  components: {
    CustomizeLineChart,
  },
  data() {
    const query = qs.parse(location.search.slice(1));
    const endDate = (+new Date());
    const startDate = endDate - (3600 * 1000);
    let mozuId = 0;
    if (query.mozuId && !isNaN(query.mozuId)) {
      mozuId = parseInt(query.mozuId);
    }
    return {
      // eslint-disable-next-line no-undef
      isTEdge: IS_TEDGE,
      fromMon: query.alarm_id_string,
      monUrl: '',
      dtl: {
        deviceInfo: {},
        alarm: {},
        ruleInfo: {},
      },
      type: '',
      typeList: [],
      id: query.id,
      mozuId,
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
      chartSeries: [],
      chartTimes: [],
      yAxis: {},
      xAxis: {},
      legend: {
        top: '5%',
        left: 'left',
      },
      selDateTime: [startDate, endDate],
      chartPointData: {},
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
    };
  },
  computed: {
    desc() {
      const { alarm } = this.dtl;
      if (isEmpty(alarm)) {
        return '';
      }

      return `${dayjs(alarm.OccurTime).format('YYYY-MM-DD HH:mm:ss')} ${this.dtl.deviceInfo.DeviceNumber} ${alarm.Content}`;
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
    this.monUrl = ``;
    this.$nextTick(() => {
      if (!this.fromMon) {
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
      }

      this.fetchInfo();
    });
  },
  beforeDestroy() {
    this.mychart = this.$refs.mychart;
    const echart = this.mychart.chart;
    echart.off('legendselectchanged');
  },
  destroyed() {
    // window.removeEventListener('resize', this.eChart.resize);
  },
  methods: {
    jumpToMon() {
      window.open(this.monUrl);
    },
    // 处理过长在页面报错
    dealJumpUrl() {
      return `/equipment/device-info?deviceNumber=${this.dtl.deviceInfo.DeviceNumber}&devName=${this.dtl.deviceInfo.DeviceName}`;
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
    durationTime(restoreTime, closeTime, occurTime) {
      function formatMs(startTime, endTime) {
        const ms = dayjs(endTime) - dayjs(startTime);
        const template = `${ms >= 60 * 60 * 1000 ? 'H小时' : ''}${ms >= 60 * 1000 ? 'm分' : ''}s秒`;
        return dayjs.duration(ms).format(template);
      }

      if (restoreTime) {
        return formatMs(occurTime, restoreTime);
      }

      return formatMs(occurTime, closeTime);
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
            this.getRuleInfo({
              rid: this.dtl.alarm.rid,
              mozuId: this.mozuId,
              page: 1,
              size: 10,
            }, { ...this.dtl.alarm });
            const occurTime = data.detail.alarm.OccurTime;
            if (occurTime) {
              this.selDateTime = [dayjs(occurTime).subtract(10, 'm')
                .format('YYYY-MM-DD HH:mm:ss'), dayjs(occurTime).add(10, 'm')
                .format('YYYY-MM-DD HH:mm:ss')];
                if (this.$moduleInfo?.isTbos) {
                  this.fetchChartData();
                } else {
                  if (!this.fromMon) {
                  getEdgeRequest(this.$axios, this.mozuId).post(cgi.getPointDataType, {
                    AlarmId: this.id,
                    MozuId: this.mozuId,
                  })
                    .then((data) => {
                      if (this.$moduleInfo?.isTbos) {
                        this.fetchChartData();
                      } else {
                        if (data && data.list.length > 0) {
                          console.log('pointDatatype不存在');
                          this.type = data.list[0];
                          this.typeList = data.list;
                          this.fetchChartData();
                        }
                      }
                    });
                  }
                }
            };
          }
        });
    },
    fetchHistory() {
      const offset = (this.currentPage - 1) * this.limit;
      const parm = {
        offset,
        limit: this.limit,
        alarmType: this.dtl.alarm.AlarmType,
        deviceGid: this.dtl.alarm.DeviceGid,
        fingerprint: this.dtl.alarm.Fingerprint,
        MozuId: this.mozuId,
      };
      // 1.0的模组未使用gid
      if (this.fromMon) {
        delete parm.deviceGid;
      }
      Object.assign(parm, this.filtered);
      getEdgeRequest(this.$axios).post(cgi.getWarningHistory, parm,)
        .then((data) => {
          this.totalItems = data.count;
          this.tableData = data.list;
        });
    },
    async getRuleInfo(data, otherParams = {}) {
      const ruleInfo = await getEdgeRequest(this.$axios, this.mozuId)
        .post(tbosCgi.GetStrategy, { ...data }, undefined, { isJson: true });
      const { list } = ruleInfo;
      const [ruleItemInfo] = list;
      this.dtl.ruleInfo = {
        alarmExpressionStr: ruleItemInfo.alarm_exp,
        restoreExpressionStr: ruleItemInfo.restore_exp,
        alarmLevelZh: otherParams?.alarmLevelZh,
      };
      console.log(ruleItemInfo, 'ruleItemInfo');
      // this.strategyInfo.data.map((item) => {
      //   item.value = ruleItemInfo[item.prop];
      //   return item;
      // });
      // const deviceInfo = {
      //   deviceNumber: otherParams?.deviceNumber,
      //   protocolType: ruleItemInfo?.device_type,
      //   deviceName: '',
      // };
      // this.deviceInfo.data.map((item) => {
      //   item.value = deviceInfo[item.prop];
      //   return item;
      // });
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
        data = await getEdgeRequest(this.$axios, this.mozuId).post(cgi.getPointData, {
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
          // eslint-disable-next-line no-unused-expressions
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
            // eslint-disable-next-line no-unused-expressions
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
    },
    selectDT() {
      this.fetchChartData();
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

    formatter(v) {
      return v;
    },
    position(params) {
      return [params[0] + 10, params[1] - 10];
    },
  },
};
</script>
<style lang="scss">
.inner-title {
  color: #333;
  font-size: 16px;
  font-weight: 700;
}
.inner-title2 {
  padding-bottom: 16px;
}
.inner-content {
  padding: 24px 0 40px 0;
}
.inner-value {
  padding-left: 64px;
}
.device-info-row {
  line-height: 50px;
  padding-top: 4px;
}
.data-type {
  position: relative;
  top: -5px;
}
.tpopover {
  padding-right: 0 !important;
}
</style>
