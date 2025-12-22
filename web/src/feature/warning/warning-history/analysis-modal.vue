<template>
  <el-modal
    class="analysis-history-warning-modal"
    :visible.sync="logVisible"
    @close="close"
    @opened="openChart"
  >
    <template
      slot="title"
    >
      统计分析
    </template>
    <template
      slot="actions"
    >
      <el-button
        type="text"
        @click="download"
      >
        <i class="tn-icon-download" /> 导出分析结果
      </el-button>
    </template>
    <el-block
      padding
      header-border
    >
      <template slot="header">
        等级分布
      </template>
      <div style="display:flex">
        <div style="min-width: 250px;flex:1">
          <pie-chart
            v-if="showLevelChart"
            ref="levelChart"
            :data="levelComposition"
            :series="levelSeries"
          />
        </div>

        <div style="margin-left: 40px;width:280px">
          <el-data-patch
            title="总数"
            :value="alarmTotal"
            style="min-width:120px"
          />
          <el-data-patch
            :value="levelData['零级'] || 0"
            style="margin-left: 36px;"
          >
            <template slot="title">
              零级 <span class="level-title">{{ levelPercent['零级'] }}</span>
            </template>
          </el-data-patch>
          <el-data-patch
            title="一级"
            :value="levelData['一级'] || 0"
            style="min-width:120px"
          >
            <template slot="title">
              一级 <span class="level-title">{{ levelPercent['一级'] }}</span>
            </template>
          </el-data-patch>

          <el-data-patch
            title="二级"
            :value="levelData['二级'] || 0"
            style="margin-left: 36px;"
          >
            <template slot="title">
              二级 <span class="level-title">{{ levelPercent['二级'] }}</span>
            </template>
          </el-data-patch>
          <el-data-patch
            title="三级"
            :value="levelData['三级'] || 0"
            style="min-width:120px"
          >
            <template slot="title">
              三级 <span class="level-title">{{ levelPercent['三级'] }}</span>
            </template>
          </el-data-patch>
          <el-data-patch
            title="四级"
            :value="levelData['四级'] || 0"
            style="margin-left: 36px;"
          >
            <template slot="title">
              四级 <span class="level-title">{{ levelPercent['四级'] }}</span>
            </template>
          </el-data-patch>
          <el-data-patch
            title="五级"
            :value="levelData['五级'] || 0"
            style="min-width:120px"
          >
            <template slot="title">
              五级 <span class="level-title">{{ levelPercent['五级'] }}</span>
            </template>
          </el-data-patch>
        </div>
      </div>
    </el-block>
    <br>
    <el-block
      padding
      header-border
    >
      <template slot="header">
        <div
          class="topten"
        >
          <div>
            TOP10分布
          </div>
          <el-button-group
            class="topten-buttongroup"
          >
            <el-button
              v-for="item in buttonOptions"
              :key="item.name"
              :type="item.name === defaultButton ? 'primary' : 'plain'"
              @click="clickButton(item)"
            >
              {{ item.name }}
            </el-button>
          </el-button-group>
        </div>
      </template>
      <el-table :data="tableData">
        <el-table-column
          label="排名"
          prop="seqId"
          width="100"
        />
        <el-table-column
          label="类型"
          prop="name"
        />
        <el-table-column
          label="数量"
          prop="count"
          width="100"
        />
        <el-table-column
          label="占比"
          width="200"
        >
          <template
            slot-scope="scope"
          >
            <el-progress :percentage="scope.row.percent" />
          </template>
        </el-table-column>
      </el-table>
    </el-block>
  </el-modal>
</template>

<script>
import { required } from 'common/script/form_rules';
import pieChart from './pie-chart';
import getEdgeRequest from '../../utils/request';
import moment from 'moment';

const timeBeforeOneMonth = moment().add('year', 0)
  .month(moment().month() - 1)
  .format('YYYY-MM-DD HH:MM:SS');
const currentTime = moment(Date.now()).format('YYYY-MM-DD HH:MM:SS');
export default {
  components: {
    pieChart,
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
    return {
      levelData: {},
      levelPercent: {
        一级: '0%',
        二级: '0%',
        三级: '0%',
        四级: '0%',
        五级: '0%',
        零级: '0%',
      },
      alarmTotal: 0,
      showLevelChart: false,
      colorMap: {
        一级: '#ff9200',
        二级: '#ffb20c',
        三级: '#09cccb',
        四级: '#156fcc',
        五级: '#8acbf2',
        零级: 'red',
      },
      tableData: [

      ],
      defaultButton: '按设备类型',
      buttonOptions: [
        { name: '按设备类型', value: 'device' },
        { name: '按告警设备', value: 'alarmDevice' },
        { name: '按告警类型', value: 'alarm' },
        { name: '按告警实例', value: 'alarmFp' },
      ],
      levelSeries: { title: '', maintitleLeft: 0, titleLeft: 40, legendX: 260 },
      levelComposition: [],
      form: {
        onResult: '',
        failureReason: '',
      },
      showFailureReason: false,
      rules: {
        onResult: [required()],
        failureReason: [required()] },
      tabData: {},
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
    'form.onResult'(v) {
      if (v !== '上架失败') {
        this.showFailureReason = false;
        this.$set(this.form, 'failureReason', '');
      } else {
        this.showFailureReason = true;
      }
    },
  },
  mounted() {
    // this.formatLevelData();
  },
  provide() {
    return {
      commonCgi: this.cgi.commonCgi,
    };
  },
  methods: {
    download() {
      this.$emit('toggleAnalysisDownload');
    },
    clickButton(item) {
      this.defaultButton = item.name;
      this.tableData = this.tabData[item.value];
    },
    // 格式化小数点
    formatDecimal(num, decimal) {
      const index = num.indexOf('.');
      if (index !== -1) {
        num = num.substring(0, decimal + index + 1);
        console.log(num);
      } else {
        num = num.substring(0);
      }
      console.log(parseFloat(num).toFixed(decimal) * 100);
      return parseFloat(num).toFixed(decimal);
    },
    formatPercent(data, total) {
      if (data === '0') return '0.00';
      const result = this.formatDecimal((parseFloat(data) / total).toLocaleString(), 4) * 100;
      if (result === 0) return '0.00';
      return `${result.toFixed(2)}`;
    },

    openChart() {
      console.log(this.$refs.levelChart);
      this.getData();
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
    // /cgi/alarm/history/getStat
    getData() {
      if (!this.data.occurTimeStart) {
        this.data.occurTimeStart = timeBeforeOneMonth;
        this.data.occurTimeEnd = currentTime;
      }
      getEdgeRequest(this.$axios, this.mozuId).post('/cgi/alarm/history/getStat', { ...this.data,
        offset: 0,
        limit: 100000 },)
        .then((data) => {
          if (data.total) {
            this.levelComposition = data.levelList.map((item) => {
              item.value = item.count;
              item.itemStyle = {
                color: this.colorMap[item.name] };
              this.levelData[item.name] = item.count;
              this.levelPercent[item.name] = item.percent;
              return item;
            });
            this.alarmTotal = data.total;

            this.tabData.device = data.topTenByDeviceType.map((i) => {
              i.percent = parseInt(i.percent.replace('%', ''));
              return i;
            });
            this.tabData.alarmDevice = data.topTenByDeviceNumber.map((i) => {
              i.percent = parseInt(i.percent.replace('%', ''));
              return i;
            }); ;
            this.tabData.alarm = data.topTenByAlarmType.map((i) => {
              i.percent = parseInt(i.percent.replace('%', ''));
              return i;
            }); ;
            this.tabData.alarmFp = data.topTenByAlarmFp.map((i) => {
              i.percent = parseInt(i.percent.replace('%', ''));
              return i;
            });
            this.tableData = this.tabData.device;
          }
          this.showLevelChart = true;
        });
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

<style lang="scss" scoped>
.topten {
  position:relative;
  display:flex;
  width:100%;
  &-buttongroup {
    position:absolute;
    right:16px;
    margin-top:16px
  }
}
</style>

<style lang="scss" >
.analysis-history-warning-modal .level-title {
  margin-left:16px;
  color:red
}
</style>
