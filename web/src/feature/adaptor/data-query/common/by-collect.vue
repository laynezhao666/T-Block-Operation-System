<template>
  <div
    class="grid"
  >
    <el-tabs
      v-model="activeName"
    >
      <el-tab-pane
        label="全部"
        name="all"
      />
      <el-tab-pane
        label="模拟量"
        name="simulation"
      />
      <el-tab-pane
        label="状态量"
        name="status"
      />
      <el-tab-pane
        label="控制量"
        name="control"
      />
    </el-tabs>

    <el-table-toolbar
      v-model="searchValue"
      :filter-value="searchValue"
      filter-placeholder="搜索"
      hide-search
      @search="search"
    >
      <template slot="extra">
        <div class="extra">
          <el-button
            class="extra-export"
            @click="exportList"
          >
            导出全部
          </el-button>
        </div>
      </template>
    </el-table-toolbar>

    <el-table
      ref="table"
      v-loading="rtLoading"
      :data="tableData"
      :max-height="tableHeight"
    >
      <el-table-column
        prop="name"
        label="测点名称"
      />
      <el-table-column
        label="当前值"
        width="150px"
      >
        <template v-slot:default="scope">
          <span v-html="scope.row.value" /> {{ scope.row.unit }}
        </template>
      </el-table-column>
      <el-table-column
        prop="updateTime"
        label="刷新时间"
        width="180px"
      />
      <el-table-column
        label="历史数据"
        width="130px"
      >
        <template v-slot:default="scope">
          <el-button
            type="text"
            size="mini"
            @click="showHistory(scope.row.id)"
          >
            查看
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import { dataQuery as cgi } from '@@/config/cgi';
import business from '@@/config/business';
import { getQueryString } from 'common/script/utils.js';
import getEdgeRequest from '../../../utils/request';

export default {
  inject: ['getSelNodeData'],
  props: {
    mozuloaded: Boolean,
    mozuId: Number,
  },
  data() {
    const keyword = getQueryString('keyword');
    return {
      activeName: 'all',
      keyword,
      searchValue: keyword || '',
      tableData: [],
      timer: null,
      rtLoading: true,
      tableHeight: 600,
    };
  },
  watch: {
    activeName() {
      this.searchValue = (this.searchValue === this.keyword) ? this.keyword : '';
      this.getData();
    },
    mozuId() {
      this.getData();
    },
  },
  mounted() {
    this.$nextTick(() => {
      this.calcTableHeight();
    });
  },
  beforeDestroy() {
    clearTimeout(this.timer);
  },
  methods: {
    calcTableHeight() {
      const totalHeight = document.querySelector('.grid').offsetHeight;
      this.tableHeight = totalHeight - 119;
    },

    /**
     * 通过点击树节点获取对应设备下的测点数据
     */
    getData() {
      if (business.showModuleSelected && !this.mozuloaded) {
        return;
      }
      clearTimeout(this.timer);
      const selNode = this.getSelNodeData();
      if (selNode) {
        this.selNode = selNode;

        // 只查询3级及以后的节点
        if (selNode.level <= 2) {
          return false;
        }

        const params = {};
        const cgiUrl = `${cgi.queryPointDetailInfoWithCurrentValueByGid}/${selNode.id}`;

        getEdgeRequest(this.$axios, this.mozuId)
          .post(cgiUrl, params, '', false)
          .then((data) => {
            this.dataList = data;
            this.showData();
            this.timer = setTimeout(() => {
              this.getData();
            }, 3000);
          })
          .finally(() => {
            this.rtLoading = false;
          });
      }
    },

    showData() {
      let dt = this.dataList;
      if (this.activeName === 'simulation') {
        dt = dt.filter(item => item.isSimulation);
      } else if (this.activeName === 'control') {
        dt = dt.filter(item => item.isReadAndWrite);
      } else if (this.activeName === 'status') {
        dt = dt.filter(item => item.isStatus);
      }
      if (this.searchValue) {
        const ls = this.searchValue.toLowerCase();
        dt = dt.filter(item => (item.name.toLowerCase().indexOf(ls) > -1 || item.id.toLowerCase().indexOf(ls) > -1));
      }
      this.tableData = dt;
    },

    showHistory(id) {
      const { selNode } = this;
      const url = `/${business.moduleName}/data-query-detail?id=${encodeURIComponent(id)}&devName=${encodeURIComponent(selNode.name)}&devId=${encodeURIComponent(selNode.id)}&mozuId=${this.mozuId}`;
      window.open(url);
    },

    search() {
      this.getData();
    },

    exportList() {
      const { selNode } = this;
      if (selNode && selNode.level === 3) {
        getEdgeRequest(this.$axios, this.mozuId).download(`${cgi.exportExcelByCollectorGidWithCurrentData}/${selNode.id}`);
      }
    },

    /**
     * 切换不同节点后，重新拉取
     */
    refresh() {
      this.getData();
    },
  },
};
</script>

<style lang="scss" scoped>
/deep/ .el-table-toolbar__extra {
  flex: 1;
}

.extra {
  display: flex;
  align-items: center;

  &-status {
    font-size: 18px;
    margin-right: 16px;
  }

  &-export {
    margin-left: auto;
  }
}

.grid {
  height: 100%;
  display: grid;
  grid-template-rows: auto auto 1fr;
}
</style>
