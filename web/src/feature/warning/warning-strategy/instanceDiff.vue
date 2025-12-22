<template>
  <div>
    <div class="data-patch">
      <el-data-patch
        title="基线策略总数"
        :value="instanceDiffOverview.standTotal || 0"
      />
      <el-data-patch
        title="已部署基线策略总数"
        type="success"
        :value="instanceDiffOverview.deployedStandCnt || 0"
      />
      <el-data-patch
        title="未部署基线策略数"
        type="danger"
        :value="instanceDiffOverview.undeployedStandCnt || 0"
      />
      <el-data-patch
        title="有修改基线策略数"
        type="warning"
        :value="instanceDiffOverview.modifyStandCnt || 0"
      />
      <el-data-patch
        title="部署策略总数"
        :value="instanceDiffOverview.deployedTotal || 0"
      />
      <el-data-patch
        title="自定义策略数"
        :value="instanceDiffOverview.customCnt || 0"
      />
      <el-data-patch
        title="基线匹配度"
        :value="deployStandPercent"
        type="danger"
        :suffix="deployStandPercent ? '%' : ''"
      />
    </div>
    <el-table
      :data="tableData"
      row-key="configId"
      :cell-style="cellStyle"
      :row-style="rowStyle"
      @filter-change="filterChange"
    >
      <el-table-column
        label="差异类型"
        prop="type"
        width="160"
        :filters="diffTypeList"
        :filter-multiple="false"
        column-key="diffType"
      >
        <template slot-scope="scope">
          <el-tag
            v-if="diffTypeList.map(e => e.value).includes(scope.row.type)"
            :type="getDiffType(scope.row.type) === '未部署' ? 'danger' : 'warning'"
          >
            {{ getDiffType(scope.row.type) }}
          </el-tag>
          <template v-else>
            {{ getDiffType(scope.row.type) }}
          </template>
        </template>
      </el-table-column>
      <el-table-column
        label="告警类型"
        prop="alarmType"
        width="240"
        :filters="alarmTypeList"
        :filter-multiple="false"
        column-key="alarmType"
      />
      <el-table-column
        label="设备类型"
        prop="deviceProtocolType"
        width="160"
        :filters="deviceTypeList"
        :filter-multiple="false"
        column-key="deviceProtocolType"
      />
      <el-table-column
        label="告警级别"
        prop="alarmLevel"
        width="120"
      >
        <template slot-scope="scope">
          {{ scope.row.alarmLevel }}
        </template>
      </el-table-column>
      <el-table-column
        label="触发表达式"
        prop="alarmExpressionStr"
        min-width="240"
      />
      <el-table-column
        label="恢复表达式"
        prop="restoreExpressionStr"
        min-width="240"
      />
      <el-table-column
        label="告警内容"
        prop="contentTemplate"
        min-width="240"
      />
      <!-- <el-table-column
        label="影响分析"
        prop="influenceAnalyze"
        show-overflow-tooltip
      />
      <el-table-column
        label="处理建议"
        prop="dealSuggestion"
        show-overflow-tooltip
      /> -->
    </el-table>

    <el-pagination
      layout="total, prev, pager, next, sizes, jumper"
      styled
      background
      :pager-count="5"
      :total="diffResult.length"
      :current-page.sync="currentPage"
      :page-sizes="[10, 20, 30, 40, 50, 100]"
      :page-size="limit"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
    />
  </div>
</template>

<script>
import { cloneDeep } from 'lodash';
export default {
  props: ['mozuId'],
  data() {
    return {
      instanceDiffOverview: {},
      upArrowSvg: 'M29.381772 120.037733l453.660918 453.666336c15.992267 15.994073 41.934997 15.994073 57.927264 0L994.632678 120.037733c15.000653-15.002459 15.000653-39.315979 0-54.303988-14.988009-15.000653-39.301529-15.000653-54.303988 0-142.774424 142.778036-285.55246 285.543429-428.31424 428.321465-142.778036-142.778036-285.55246-285.543429-428.330496-428.321465-15.000653-15.000653-39.301529-15.000653-54.302182 0-15.000653 14.988009-15.000653 39.303336 0 54.303988z',
      diffResult: [],
      filterData: [],

      currentPage: 1,
      limit: 10,
      getTableFilterParam: {
        diffType: '',
        alarmType: '',
        deviceProtocolType: '',
      },
    };
  },
  computed: {
    tableData() {
      return this.diffResult.slice((this.currentPage - 1) * this.limit, this.currentPage * this.limit);
    },
    diffTypeList() {
      const result = new Set(this.filterData.map(e => e.type));
      return Array.from(result).map(e => ({ text: this.getDiffType(e), value: e }));
    },
    alarmTypeList() {
      const result = new Set(this.filterData.map(e => e.alarmType));
      return Array.from(result).map(e => ({ text: e, value: e }));
    },
    deviceTypeList() {
      const result = new Set(this.filterData.map(e => e.deviceProtocolType));
      return Array.from(result).map(e => ({ text: e, value: e }));
    },
    deployStandPercent() {
      if (this.instanceDiffOverview.deployStandPercent) {
        return this.instanceDiffOverview.deployStandPercent * 100;
      }
      return 0;
    },
  },
  watch: {
    mozuId() {
      this.getInstancediffData();
    },
  },
  // created() {
  //   this.getInstancediffData();
  // },
  methods: {
    filterChange(filters) {
      if (filters) {
        const [key] = Object.keys(filters);
        [this.getTableFilterParam[key]] = filters[key];
      }
      const { diffType, alarmType, deviceProtocolType } = this.getTableFilterParam;
      const filterData = this.filterData.filter((e) => {
        if (diffType && alarmType && deviceProtocolType) {
          return e.type === diffType && e.alarmType === alarmType && e.deviceProtocolType === deviceProtocolType;
        }
        if (diffType && alarmType) {
          return e.type === diffType && e.alarmType === alarmType;
        }
        if (diffType && deviceProtocolType) {
          return e.type === diffType && e.deviceProtocolType === deviceProtocolType;
        }
        if (alarmType && deviceProtocolType) {
          return e.alarmType === alarmType && e.deviceProtocolType === deviceProtocolType;
        }

        if (diffType) return e.type === diffType;
        if (alarmType) return e.alarmType === alarmType;
        if (deviceProtocolType) return e.deviceProtocolType === deviceProtocolType;

        return true;
      });
      this.diffResult = filterData;
    },
    rowStyle({ row }) {
      if (row.isChildren) {
        return {
          backgroundColor: '#eee',
        };
      }
      return {};
    },
    cellStyle({ row, column }) {
      if (row.isParent && row[column.property] !== row.children[0][column.property]) {
        return {
          color: '#ff3e00',
        };
      }
      return {};
    },
    handleSizeChange(val) {
      this.limit = val;
    },
    handleCurrentChange(val) {
      this.currentPage = val;
    },
    getDiffType(type) {
      if (type) return type === 'undeployed' ? '未部署' : '有修改';

      return '基线';
    },
    getInstancediffData() {
      this.$axios.get('/cgi/alarm/configaudit/cover/getInstanceDiff', { mozuId: this.mozuId })
        .then((res) => {
          this.instanceDiffOverview = res.overview;
          const { diffResult } = res;
          function getObj(e) {
            return e.type === 'modify' ? e.instance : e.standard;
          };
          diffResult.forEach((e, index) => {
            diffResult[index] = {
              ...e,
              ...getObj(e),
              isParent: e.type === 'modify',
              children: e.type === 'modify' ? [{ ...e.standard, isChildren: true }] : [],
            };
          });
          this.diffResult = diffResult.filter(e => e.type !== 'custom');
          this.filterData = cloneDeep(this.diffResult);
        });
    },
  },
};
</script>
<style lang="scss" scoped>
.data-patch {
  padding: 16px 24px;
  box-sizing: border-box;
  display: flex;
  justify-content: space-between;
  // display: grid;
  // grid-template-columns: repeat(4, 25%);
  // grid-template-columns: repeat(4, minmax(0, 1fr));
}
/deep/.el-table tbody tr:hover>td {
  background-color: unset !important;
}
</style>
