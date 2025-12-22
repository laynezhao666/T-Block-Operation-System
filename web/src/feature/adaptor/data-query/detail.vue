<template>
  <div>
    <el-title>监控数据详情</el-title>

    <el-block
      no-padding
    >
      <div style="width:580px;padding-left:24px;line-height:56px">
        <el-date-picker
          v-model="selDateTime"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          :picker-options="pickerOptions"
          type="datetimerange"
          style="width:'400px'"
          @change="selectDT"
        />
      </div>

      <tn-line-chart
        :x-axis="{data: chartData.time}"
        :series="chartData.series"
        :y-axis="chartData.yAxis"
        :tooltip="{
          ignoreNil: true
        }"
      />
    </el-block>
    <el-block
      no-padding
      style="margin-top:16px"
    >
      <el-table-toolbar hide-search>
        <template slot="extra">
          <el-button
            plain
            @click="exportList"
          >
            导出
          </el-button>
        </template>
      </el-table-toolbar>
      <el-table
        :data="tableData"
        style="width: 100%"
        @sort-change="sortChange"
      >
        <el-table-column
          prop="updateTime"
          label="时间"
          sortable="custom"
        />
        <el-table-column
          width="180"
          label="值"
          prop="value"
          sortable="custom"
        >
          <template v-slot:default="scope">
            <span v-html="scope.row.value" />{{ scope.row.unit }}
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        styled
        layout="total, prev, pager, next, sizes, jumper"
        background
        :pager-count="5"
        :total="totalItems"
        :current-page.sync="currentPage"
        :page-sizes="[10, 20, 30, 40, 50, 100,200,500]"
        :page-size="pageSize"
        @size-change="handleSizeChangeHistory"
        @current-change="handleCurrentChangeHistory"
      />
    </el-block>
  </div>
</template>

<script>

import { dataQuery as cgi } from '@@/config/cgi';
import business from '@@/config/business';
import getEdgeRequest from '../../utils/request';
import { getQueryString } from 'common/script/utils.js';
import * as dayjs from 'dayjs';
export default {
  components: {
  },
  data() {
    const strStartTime = getQueryString('start');
    const strEndTime = getQueryString('end');
    let startTime;
    let endTime;
    if (strStartTime) {
      startTime = new Date(strStartTime);
    }
    if (strEndTime) {
      endTime = new Date(strEndTime);
    }
    if (startTime && endTime && startTime < endTime) {
      startTime = +startTime;
      endTime = +endTime;
    } else {
      startTime = (+new Date()) - (3600000 * 1);
      endTime = +new Date();
    }

    let mozuId = 0;
    const strMozuId = getQueryString('mozuId');
    if (strMozuId && !isNaN(strMozuId)) {
      mozuId = parseInt(strMozuId);
    }

    return {
      links: [
        { name: '监控数据', url: `/${business.moduleName}/data-query-index` },
        { name: '数据查询', url: `/${business.moduleName}/data-query-index` },
        { name: getQueryString('devName') },
      ],
      options: [
      ],
      mozuId,
      selected: {},
      selDateTime: [
        startTime,
        endTime,
      ],
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
      },
      chartData: {
        time: [],
        series: [],
        yAxis: {},
      },
      orgData: [],
      tableData: [],
      totalItems: 0,
      currentPage: 1,
      pageSize: 10,
    };
  },
  watch: {
    selected() {
      this.showHistory();
    },
  },
  mounted() {
    this.getData();
  },
  methods: {
    getData() {
      const devId = getQueryString('devId');
      const id = getQueryString('id');
      const cgiUrl = cgi.queryPointDetailInfoWithCurrentValueByGid;
      getEdgeRequest(this.$axios, this.mozuId).get(`${cgiUrl}/${devId}`, { rnd: +new Date() },)
        .then((data) => {
          this.options = data.map((item) => {
            const opName = item.name;
            const op = {
              label: opName,
              value: opName,
              id: item.id,
            };
            if (item.id === id) {
              this.selected = op;
            }
            return op;
          });
        });
    },
    showHistory() {
      const { id } = this.selected;
      if (!this.selDateTime) {
        this.$message.warning('请选择时间范围');
        return;
      }
      const params = {
        id,
        startSecondKey: dayjs(this.selDateTime[0]).format('YYYYMMDDHHmmss'),
        endSecondKey: dayjs(this.selDateTime[1]).format('YYYYMMDDHHmmss'),
        isAll: true,
      };
      const cgiUrl = cgi.queryHistoryPointInfoByTimeRangeAndPageAndOrder;
      getEdgeRequest(this.$axios, this.mozuId)
        .post(cgiUrl, params, undefined, {
          isJson: true,
        })
        .then((data) => {
          const orgData = data.list;
          this.chartData.time = orgData.map(item => item.updateTime);
          this.chartData.series = [{
            name,
            data: orgData.map(item => item.value),
            areaStyle: true,
            unit: (orgData && orgData.length > 0) ? orgData[0].unit : '',
          }];
          const yAxis = { };
          this.chartData.yAxis = yAxis;
          this.orgData = orgData;
          this.totalItems = orgData.length;
          this.currentPage = 1;
          const start = (this.currentPage - 1) * this.pageSize;
          this.tableData = orgData.slice(start, this.pageSize);
        });
    },
    handleSizeChangeHistory(val) {
      this.currentPage = 1;
      this.pageSize = val;

      const start = (this.currentPage - 1) * this.pageSize;
      this.tableData = this.orgData.slice(start, start + this.pageSize);
    },
    handleCurrentChangeHistory(val) {
      this.currentPage = val;
      const start = (this.currentPage - 1) * this.pageSize;
      this.tableData = this.orgData.slice(start, start + this.pageSize);
    },
    selectDT() {
      this.showHistory();
    },
    handleRoute(r) {
      if (r && r.url) {
        TNBL.redirectUrl(r.url);
        // window.open(r.url);
      }
    },
    sortChange({ prop, order }) {
      if (prop) {
        this.orgData.sort((a, b) => {
          const av = a[prop];
          const bv = b[prop];
          if (av === bv) {
            return 0;
          }
          if (av > bv) {
            return order === 'descending' ? -1 : 1;
          }

          return order === 'descending' ? 1 : -1;
        });
        this.handleSizeChangeHistory(this.pageSize);
      }
    },
    exportList() {
      const { id } = this.selected;
      if (id) {
        const startSecondKey = dayjs(this.selDateTime[0]).format('YYYYMMDDHHmmss');
        const endSecondKey = dayjs(this.selDateTime[1]).format('YYYYMMDDHHmmss');
        const url = `${cgi.exportHistoryPointInfoByTimeRangeAndPageAndOrder}/${id}/${startSecondKey}/${endSecondKey}`;
        getEdgeRequest(this.$axios, this.mozuId).download(url);
      }
    },
  },
};
</script>
<style scoped>
.el-date-editor.el-input{
  width:auto
}
</style>
