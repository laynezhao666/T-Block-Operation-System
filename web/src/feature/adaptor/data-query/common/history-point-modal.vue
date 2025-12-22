<template>
  <el-modal
    title="历史数据查询"
    :visible.sync="logVisible"
    :width="900"
  >
    <el-block style="padding: 10px 10px">
      <div class="advanced-search-container">
        <div class="advanced-search-main">
          <div
            class="advanced-search-toolbar"
            style="padding :10px 20px 10px 10px"
          >
            <div style="display: flex;">
              <el-space
                :size="24"
                style="flex: 1;"
              >
                <el-date-picker
                  v-model="filter.timerange"
                  style="width:400px"
                  class="advanced-search-toolbar-date-picker"
                  type="datetimerange"
                  range-separator="至"
                  :format="{
                    date: 'yyyy-MM-dd',
                    time: 'HH:mm:ss'
                  }"
                  value-format="yyyy-MM-dd HH:mm:ss"
                  start-placeholder="开始时间"
                  end-placeholder="结束时间"
                  :picker-options="pickerOptions"
                  @change="timerangeChangeHandler"
                />

                <div
                  v-if="!isPad"
                  class="advanced-search-toolbar-time"
                >
                  每
                  <el-input
                    v-model="filter.duration"
                    style="width:80px"
                    class="advanced-search-toolbar-time-value"
                    clearable
                  />
                  <el-select
                    v-model="filter.unit"
                    style="width:100px"
                    class="advanced-search-toolbar-time-unit"
                    placeholder="时间"
                    no-input
                  >
                    <el-option
                      v-for="item in unitOptions"
                      :key="item.value"
                      :label="item.label"
                      :value="item.value"
                    />
                  </el-select>
                </div>
              </el-space>

              <el-button
                type="primary"
                :disabled="!checkedPointList.length"
                @click="getData"
              >
                查询
              </el-button>

              <el-button
                v-if="!isPad"
                type="primary"
                :disabled="!checkedPointList.length"
                @click="exportHistory"
              >
                导出
              </el-button>
            </div>

            <div class="advanced-search-toolbar-template">
            </div>
          </div>

          <div
            class="advanced-search-main-content"
          >
            <el-block
              inner
              border
            >
              <div
                ref="chart"
                :style="{ height: isPad ? 'calc(100vh - 300px)' : '800px' }"
                class="advanced-search-chart"
              />
            </el-block>
          </div>
        </div>
      </div>
    </el-block>
  </el-modal>
</template>

<script>
import moment from 'moment';
import qs from 'qs';
import dayjs from 'dayjs';
import getEdgeRequest from '../../../utils/request';
import { dataQuery as cgi } from '@@/config/cgi';
const color = [
  '#1470CC',
  '#FFB20D',
  '#0ACCCC',
  '#FF7A0D',
  '#5939D6',
  '#0ACC78',
  '#DA5FAC',
  '#96C0EB',
  '#FFE09C',
  '#9BE8E8',
  '#FFCA9E',
  '#AE9EEB',
  '#9BEDCA',
  '#E4A5D4',
];
/**
 * 找到一个大约等于 x 的“好”数。
 */
function nice(span, round) {
  let val = span;
  const exponent = Math.floor(Math.log(val) / Math.LN10);
  const exp10 = 10 ** exponent;
  const f = val / exp10;

  let nf;

  if (round) {
    if (f < 1.5) {
      nf = 1;
    } else if (f < 2.5) {
      nf = 2;
    } else if (f < 4) {
      nf = 3;
    } else if (f < 7) {
      nf = 5;
    } else {
      nf = 10;
    }
  } else {
    if (f < 1) {
      nf = 1;
    } else if (f < 2) {
      nf = 2;
    } else if (f < 3) {
      nf = 3;
    } else if (f < 5) {
      nf = 5;
    } else {
      nf = 10;
    }
  }

  val = nf * exp10;

  return exponent >= -20 ? +val.toFixed(exponent < 0 ? -exponent : 0) : val;
}
export default {
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    pointListProp: {
      type: String,
      default: '',
    },
    isPad: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    const defaultTimeRange = [
      moment().subtract(1, 'hour')
        .format('YYYY-MM-DD HH:mm:ss'),
      moment().format('YYYY-MM-DD HH:mm:ss'),
    ];

    return {
      mozuId: null,

      /**
       * 测点列表
       */
      isIndeterminate: false,
      checkAll: false,
      pointList: [], // 从 url 获取的测点列表
      checkedPointList: [], // checkbox group绑定用的测点列表
      fullPointList: [], // 记录了Y轴设置的完整测点列表

      /**
       * 查询条件
       */
      filter: {
        timerange: defaultTimeRange,
        duration: 1,
        unit: 60,
        stats: [],
      },
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
        ],
        disabledDate: time => time.getTime() > Date.now(),
      },
      unitOptions: [
        {
          value: 1,
          label: '秒',
        },
        {
          value: 60,
          label: '分钟',
        },
        {
          value: 60 * 60,
          label: '小时',
        },
        {
          value: 60 * 60 * 24,
          label: '天',
        },
      ],
      statOption: [
        {
          value: 'avg',
          label: '平均值',
        },
        {
          value: 'max',
          label: '最大值',
        },
        {
          value: 'min',
          label: '最小值',
        },
      ],

      /**
       * table
       */
      tableColumns: [],
      tableData: [],
      originTableData: [],
      statTableData: [],
      currentPage: 1,
      totalItems: 0,
      pageSize: 10,

      /**
       * y轴选项
       */
      currentPoint: {},
      yAxisOptionDialogVisible: false,
      form: {
        max: '',
        min: '',
        ownYaxis: false,
      },

      /**
       * 加载模板
       */
      loadTemplateModalVisible: false,
      templateList: [],
      currentTemplate: {},
      templateModalType: 'load',
      templatePagination: {
        totalItems: 0,
        currentPage: 1,
        pageSize: 10,
      },

      /**
       * 保存模板
       */
      saveTemplateModalVisible: false,
      templateForm: {
        templateName: '',
        templateDetail: '',
      },
      templateFormRules: {
        templateName: [
          { required: true, message: '请输入模板名称', trigger: 'blur' },
        ],
      },
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

  },
  mounted() {
    const query = qs.parse(location.search.slice(1));
    const mozuId = query.mozuId ? query.mozuId : 0;
    this.mozuId = mozuId || TNBL.getCurrModule?.().id || window.__GetFrameDataByKey('curMozuData')?.id;

    if (this.pointListProp) {
      this.getPointList();
    }
  },
  provide() {
    return {
      // commonCgi: this.cgi.commonCgi,
    };
  },
  methods: {
    /**
     * 从 url 参数中获取测点
     */
    getPointList() {
      this.pointList = this.pointListProp.split(',');
      this.checkedPointList = this.pointList;
      this.checkAll = true;

      this.fullPointList = this.pointList.map(e => ({
        id: e,
        deviceNo: e.split('.')[0],
        attrName: e.split('.')[1],
        yaxisMax: '',
        yaxisMin: '',
        ownYaxis: false,
      }));

      this.getData();
    },

    /**
     * 全选处理
     */
    handleCheckAllChange(val) {
      if (val) {
        this.checkedPointList = this.pointList;
      } else {
        this.checkedPointList = [];
      }

      this.isIndeterminate = false;
    },

    /**
     * 单选处理
     */
    handleCheckedPointListChange(val) {
      const checkedCount = val.length;
      this.checkAll = checkedCount === 10;
      this.isIndeterminate = checkedCount > 0 && checkedCount < 10;
    },

    /**
     * 根据时间跨度，设置时间单位
     * 时间跨度小于10分钟时，默认1秒
     * 大于等于10分钟，小于1天，默认1分钟
     * 大于等于1天，小于30天，默认1小时
     * 大于等于30天，默认1天
     * @param {Array} val - 时间跨度
     */
    timerangeChangeHandler(val) {
      const [startTime, endTime] = val;
      const diff = moment(endTime).diff(moment(startTime));

      if (diff >= 30 * 24 * 60 * 60 * 1000) {
        this.filter.unit = 60 * 60 * 24;
      } else if (diff >= 24 * 60 * 60 * 1000) {
        this.filter.unit = 60 * 60;
      } else if (diff >= 10 * 60 * 1000) {
        this.filter.unit = 60;
      } else {
        this.filter.unit = 1;
      }

      this.filter.duration = 1;
    },

    /**
     * 根据模板测点列表查询历史数据
     */
    getData() {
      const [startTime, endTime] = this.filter.timerange;

      if (this.isPad) {
        this.$axios.post('/cgi/standard/history/points/sampling', {
          ids: this.pointListProp.split(','),
          begin_time: dayjs(startTime).toDate()
            .getTime() / 1000,
          end_time: dayjs(endTime).toDate()
            .getTime() / 1000,
        }).then((data) => {
          this.createPadChart(data);
        });
      } else {
        const params = {
          templatePointList: this.checkedPointList,
          startTime,
          endTime,
          interval: this.filter.duration * this.filter.unit,
          stats: ['avg', 'min', 'max'],
        };

        getEdgeRequest(this.$axios, this.mozuId)
          .post(cgi.getHistoryBizGidAttrValuesByTemplate, params)
          .then((data) => {
            // const sortedData = _.sortBy(
            //   data,
            //   o => this.checkedPointList.findIndex(e => e === o.id)
            // );

            this.createTable(data);
            this.createChart(data);
          });
      }
    },

    /**
     * 根据模板测定列表导出历史数据
     */
    exportHistory() {
      const [startTime, endTime] = this.filter.timerange;
      const params = {
        templatePointList: this.checkedPointList,
        startTime,
        endTime,
        interval: this.filter.duration * this.filter.unit,
        stats: this.filter.stats,
      };

      getEdgeRequest(this.$axios, this.mozuId)
        .download(cgi.exportHistoryBizGidAttrValuesByTemplate, params);
    },

    /**
     * 将获取到的历史数据转换为表格支持的数据
     * @param {Array} data - 历史数据
     */
    createTable(data) {
      this.currentPage = 1;
      this.pageSize = 10;

      const statMap = {
        avg: '平均值',
        max: '最大值',
        min: '最小值',
      };

      this.tableColumns = data.map(e => ({
        prop: e.deviceNumber + e.attrName,
        label: e.id,
        sortable: true,
      }));

      this.originTableData = data[0].data.map((e, i) => {
        const columnData = {
          updateTime: e.updateTime,
        };

        data.forEach((item) => {
          const id = item.deviceNumber + item.attrName;
          columnData[id] = item.data[i]?.value;
        });

        return columnData;
      });
      this.totalItems = this.originTableData.length;
      this.sliceTableData();

      this.statTableData = this.filter.stats.map((e) => {
        const columnData = {
          stat: statMap[e],
        };

        data.forEach((item) => {
          const id = item.deviceNumber + item.attrName;
          columnData[id] = item.stats.find(s => s.name === e).value;
        });

        return columnData;
      });
    },

    sliceTableData() {
      const start = (this.currentPage - 1) * this.pageSize;
      const limit = this.pageSize;

      this.tableData = this.originTableData.slice(start, start + limit);
    },

    handleSizeChange(val) {
      this.pageSize = val;
      this.currentPage = 1;

      this.sliceTableData();
    },

    handleCurrentChange(val) {
      this.currentPage = val;

      this.sliceTableData();
    },

    /**
     * 创建PAD端折线图
     * @param {Array} data - 历史数据
     */
    createPadChart(data) {
      this.chart = echarts.init(this.$refs.chart);
      this.chart.clear();

      const option = {
        color,
        tooltip: {
          trigger: 'axis',
        },
        legend: {
          data: _.keys(data),
          top: 10,
          left: 32,
          selectedMode: false,
          textStyle: {
            fontSize: 11,
            color: '#FFF',
          },
        },
        grid: {
          top: 80,
          bottom: 80,
          left: 64,
          right: 64,
        },
        xAxis: {
          type: 'time',
        },
        yAxis: {
          type: 'value',
          splitLine: {
            lineStyle: {
              color: '#1470cc99',
            },
          },
        },
        series: _.map(data, (valueList, pointId) => ({
          name: pointId,
          type: 'line',
          data: valueList.map(item => ([
            Number(item.tms + 1000),
            Number(item.pv),
          ])),
        })),
      };

      console.log(option);

      this.chart.setOption(option);
    },

    /**
     * 创建折线图
     * @param {Array} data - 历史数据
     */
    createChart(data) {
      this.chart = echarts.init(this.$refs.chart);
      this.chart.clear();

      const noYAxisPoints = data.filter(e => this.fullPointList.find(a => a.id === e.id && !a.ownYaxis));
      const yAxisPoints = data.filter(e => this.fullPointList.find(a => a.id === e.id && a.ownYaxis));
      const yAxisCount = (noYAxisPoints.length ? 1 : 0) + yAxisPoints.length;

      const option = {
        color,
        toolbox: {
          show: true,
          feature: {
            saveAsImage: {},
          },
          top: 4,
          right: 24,
        },
        tooltip: {
          trigger: 'axis',
          backgroundColor: '#fff',
          padding: 16,
          extraCssText: 'border-radius: 0; min-width: 200px; box-shadow: 0 3px 5px 0 rgba(203,203,203,0.50);',
          textStyle: {
            color: '#666',
            fontSize: 12,
          },
          formatter(params) {
            const arr = params.map((param) => {
              const { unit } = data.find(e => e.id === param.seriesName);
              return `${param.marker}${param.seriesName}：${param.data} ${unit}`;
            });

            return `
              ${params[0].name}<br>
              ${arr.join('<br>')}
            `;
          },
        },
        legend: {
          data: data.map(e => e.id),
          top: 30,
          left: 24,
          selectedMode: false,
          textStyle: {
            fontSize: 11,
          },
        },
        grid: {
          top: 120,
          bottom: 40,
          left: 80,
          right: yAxisCount <= 2
            ? 80
            : ((yAxisCount - 1) * 80),
        },
        xAxis: {
          type: 'category',
          data: data[0].data.map(e => e.updateTime),
          axisTick: { show: false },
          axisLine: {
            lineStyle: {
              color: '#f0f0f0',
            },
          },
          axisLabel: {
            color: '#999',
            fontSize: 10,
            formatter(value) {
              return value.split(' ').join('\n');
            },
          },
        },
        yAxis: this.createYAxis(data),
        series: this.createSeries(data),
      };

      this.chart.setOption(option);
    },

    createYAxis(data) {
      const yAxisList = [];

      // 最左侧的主轴
      const noYAxisPoints = data.filter(e => this.fullPointList.find(a => a.id === e.id && !a.ownYaxis));
      const yAxisPoints = data.filter(e => this.fullPointList.find(a => a.id === e.id && a.ownYaxis));
      let max = nice(
        Math.max(...noYAxisPoints.map(e => e.stats.find(item => item.name === 'max').value)),
        false
      );
      if (max === 0) {
        max = 1;
      }

      let mainAxisCount = 0;

      if (noYAxisPoints.length) {
        yAxisList.push({
          type: 'value',
          axisTick: { show: false },
          axisLine: { show: false },
          axisLabel: {
            color: '#999',
          },
          splitLine: {
            lineStyle: {
              color: '#f0f0f0',
            },
          },
          splitNumber: 5,
          interval: Math.round(max / 5 * 100) / 100,
          // min: 0,
          max,
        });

        mainAxisCount = 1;
      }

      yAxisPoints.forEach((e, i) => {
        const index = data.findIndex(item => item.id === e.id);
        const { yaxisMax, yaxisMin } = this.fullPointList.find(p => p.id === e.id);
        const maxValue = e.stats.find(item => item.name === 'max').value;

        let min = yaxisMin === '' ? 0 : yaxisMin;
        let max = yaxisMax === '' ? nice(maxValue, false) : yaxisMax;

        if (max === 0) {
          max = 1;
        }
        if (min === 1) {
          min = 0;
        }

        let offset = 0;
        if (mainAxisCount === 1) {
          offset = 80 * i;
        } else {
          if (i === 0) {
            offset = 0;
          } else {
            offset = 80 * (i - 1);
          }
        }

        yAxisList.push({
          type: 'value',
          axisTick: { show: false },
          axisLine: { show: false },
          axisLabel: {
            color: color[index],
          },
          splitLine: {
            lineStyle: {
              color: '#f0f0f0',
            },
          },
          splitNumber: 5,
          interval: Math.round((max - min) / 5 * 100) / 100,
          min,
          max,
          offset,
        });
      });

      return yAxisList;
    },

    createSeries(data) {
      const noYAxisPoints = data.filter(e => this.fullPointList.find(a => a.id === e.id && !a.ownYaxis));

      let yAxisCount = 0;
      if (noYAxisPoints.length) {
        yAxisCount = 1;
      }
      console.log(this.fullPointList);
      return data.map((e) => {
        const fullPoint = this.fullPointList.find(p => p.id === e.id);
        const { ownYaxis } = fullPoint;

        let yAxisIndex = 0;

        if (ownYaxis) {
          yAxisIndex = yAxisCount;
          yAxisCount += 1;
        }

        return {
          name: e.id,
          data: e.data.map(e => e.value),
          yAxisIndex: ownYaxis ? yAxisIndex : 0,
          type: 'line',
          markPoint: {
            symbolSize: 40,
            data: [
              this.filter.stats.includes('max')
                ? { type: 'max', name: '最大值' } : {},
              this.filter.stats.includes('min')
                ? { type: 'min', name: '最小值' } : {},
            ],
          },
          markLine: this.filter.stats.includes('avg') ? {
            symbol: 'none',
            data: [
              { type: 'average', name: '平均值' },
            ],
            label: {
              show: false,
            },
          } : {},
          symbol: 'none',
        };
      });
    },

    /**
     * 打开Y轴设置编辑弹框
     * @param {string} pointId - 测点的id
     */
    openYAxisOptionDialog(pointId) {
      this.yAxisOptionDialogVisible = true;

      this.currentPoint = this.fullPointList.find(e => e.id === pointId);

      const { yaxisMax, yaxisMin, ownYaxis } = this.currentPoint;
      this.form = {
        max: yaxisMax,
        min: yaxisMin,
        ownYaxis: Number(ownYaxis) === 1,
      };
    },

    /**
     * Y轴选项恢复为默认值
     */
    resetYAxisOption() {
      // 恢复为空值，在图表设置中会自动计算
      this.form = {
        max: '',
        min: '',
        ownYaxis: false,
      };
    },

    /**
     * 保存测点的Y轴选项
     */
    saveYAxisOption() {
      const { form, currentPoint } = this;
      const { max, min, ownYaxis } = form;

      currentPoint.yaxisMax = max;
      currentPoint.yaxisMin = min;
      currentPoint.ownYaxis = ownYaxis;

      this.yAxisOptionDialogVisible = false;
    },

    /**
     * 模板管理
     */
    handleCommand(command) {
      switch (command) {
        case 'loadTemplate': {
          this.openTemplateListModal({ type: 'load' });
          break;
        }
        case 'manageTemplate': {
          this.openTemplateListModal({ type: 'manage' });
          break;
        }
        case 'saveTemplate': {
          this.openSaveTemplateModal();
          break;
        }
        case 'editTemplate': {
          this.overrideTemplate();
          break;
        }
      }
    },

    /**
     * 打开加载模板弹框
     */
    openTemplateListModal(options) {
      this.templateModalType = options.type;
      this.loadTemplateModalVisible = true;

      this.getTemplateList();
    },

    /**
     * 加载模板列表
     */
    getTemplateList() {
      const { currentPage, pageSize } = this.templatePagination;
      const params = {
        fieldWithValueMap: {},
        start: (currentPage - 1) * pageSize,
        limit: pageSize,
      };

      getEdgeRequest(this.$axios, this.mozuId)
        .post(cgi.selectTemplateByCondition, params, '', false)
        .then((data) => {
          this.templateList = data.list;
          this.templatePagination.totalItems = data.count;
        });
    },

    /**
     * 点击模板名称加载模板
     */
    loadTemplate(row) {
      this.currentTemplate = { ...row };

      this.fullPointList = row.data.map(e => ({
        ...e,
        id: `${e.deviceNo}.${e.attrName}`,
      }));
      this.pointList = this.fullPointList.map(e => e.id);

      this.checkedPointList = this.fullPointList
        .filter(e => Number(e.checked) === 1)
        .map(e => e.id);

      this.checkAll = false;
      if (this.checkedPointList.length) {
        this.isIndeterminate = true;
      }
      if (this.checkedPointList.length === this.fullPointList) {
        this.checkAll = true;
        this.isIndeterminate = false;
      }

      const { startTime, endTime, step, unit, stats } = row;

      this.filter = {
        timerange: [startTime, endTime],
        duration: step,
        unit,
        stats,
      };

      this.loadTemplateModalVisible = false;
      this.getData();
    },

    /**
     * 删除模板
     * @param {Object} row - 需要删除的模板列
     */
    deleteTemplate(row) {
      this
        .$confirm(`确定删除模板“${row.templateName}”吗？`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        })
        .then(() => {
          getEdgeRequest(this.$axios, this.mozuId)
            .post(`${cgi.deleteTemplateById}/${row.id}`, {}, '', false)
            .then(() => {
              this.$message.success('已删除模板');

              this.getTemplateList();
            });
        })
        .catch(() => {});
    },

    /**
     * 打开保存模板弹框
     */
    openSaveTemplateModal() {
      this.templateForm = {
        templateName: '',
        templateDetail: '',
      };

      this.saveTemplateModalVisible = true;
    },

    /**
     * 打开编辑模板弹框
     * @param {Object} row - 需要编辑的模板列
     */
    openEditTemplateModal(row) {
      const { id, templateName, templateDetail } = row;

      this.templateForm = {
        id,
        templateName,
        templateDetail,
      };
      this.saveTemplateModalVisible = true;
    },

    createParams() {
      const data = this.fullPointList.map(e => ({
        ...e,
        ownYaxis: e.ownYaxis ? 1 : 0,
        checked: this.checkedPointList.includes(e.id) ? 1 : 0,
      }));
      const { duration, unit, timerange, stats } = this.filter;
      const step = duration;
      const [startTime, endTime] = timerange;

      return {
        data,
        step,
        unit,
        stats,
        startTime,
        endTime,
      };
    },

    /**
     * 发送保存模板请求
     */
    saveTemplate() {
      this.$refs.templateForm.validate((valid) => {
        if (valid) {
          const { templateName, templateDetail } = this.templateForm;

          const params = {
            ...this.createParams(),
            templateName,
            templateDetail,
          };

          // 如果带有 id 属性，为编辑模板
          if ('id' in this.templateForm) {
            params.id = this.templateForm.id;
          }

          getEdgeRequest(this.$axios, this.mozuId)
            .post(cgi.insertOrUpdateTemplate, params, '', false)
            .then(() => {
              this.$message.success('保存模板成功');
              this.saveTemplateModalVisible = false;
            });
        }
      });
    },

    /**
     * （修改测点的Y轴设置后）覆盖模板
     */
    overrideTemplate() {
      const { templateName } = this.currentTemplate;
      this
        .$confirm(`确定覆盖模板“${templateName}”吗？`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        })
        .then(() => {
          const { id, templateDetail } = this.currentTemplate;

          const params = {
            ...this.createParams(),
            id,
            templateName,
            templateDetail,
          };

          getEdgeRequest(this.$axios, this.mozuId)
            .post(cgi.insertOrUpdateTemplate, params, '', false)
            .then(() => {
              this.$message.success('编辑模板成功');
              this.saveTemplateModalVisible = false;
            });
        })
        .catch(() => {});
    },

    handleTemplatePaginationSizeChange(val) {
      this.templatePagination.pageSize = val;
      this.templatePagination.currentPage = 1;

      this.getTemplateList();
    },

    handleTemplatePaginationCurrentChange(val) {
      this.templatePagination.currentPage = val;

      this.getTemplateList();
    },
  },
};
</script>
<style>

</style>
