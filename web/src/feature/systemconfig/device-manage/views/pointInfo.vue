
<template>
  <div
    ref="point"
    v-loading="vLoading"
  >
    <el-table
      v-if="tableHeight"
      ref="multipleTable"
      :data="tableData"
      tooltip-effect="dark"
      style="width: 100%"
      :max-height="tableHeight"
    >
      <el-table-column
        prop="no"
        label="信号标识符"
      />
      <el-table-column
        prop="name"
        label="信号名称"
      />

      <el-table-column
        prop="valtype"
        width="100"
        label="数据类型"
      >
        <template slot-scope="scope">
          {{ typeMap[scope.row.valtype] }}
        </template>
      </el-table-column>
      <el-table-column
        prop="name"
        label="采集值"
      >
        <template slot-scope="scope">
          <!-- <div v-if="collector && collector.collector && !collector.collector.isOnline"> -->
          <div v-if="false">
            --
          </div>
          <div v-else-if="pointDatas[scope.row.id]">
            <span
              :class="{normal: pointDatas[scope.row.id].qua === '0'}"
              class="button"
            >
              {{ scope.row.valtype === 'A'
                ? (pointDatas[scope.row.id].pv + (scope.row.valdef.unit ? scope.row.valdef.unit : '') || '--')
                : (scope.row.valueMap[pointDatas[scope.row.id].pv] || pointDatas[scope.row.id].pv || '--') }}
            </span>
          </div>
          <div v-else>
            --
          </div>
        </template>
      </el-table-column>
      <el-table-column
        prop="name"
        width="180"
        label="采集时间"
      >
        <template slot-scope="scope">
          <!-- <div v-if="collector && collector.collector && !collector.collector.isOnline"> -->
          <div v-if="false">
            --
          </div>

          <div v-else-if="pointDatas[scope.row.id]">
            <span>{{ pointDatas[scope.row.id].formatTms }}</span>
          </div>
          <div v-else>
            --
          </div>
        </template>
      </el-table-column>

      <el-table-column
        prop="qua"
        width="150"
        label="数据质量"
      >
        <template slot-scope="scope">
          <!-- <div v-if="collector && collector.collector && !collector.collector.isOnline"> -->
          <!-- <div v-if="false">
            --
          </div> -->
          <div v-if="pointDatas[scope.row.id]">
            <span
              slot="reference"
              :class="{normal: pointDatas[scope.row.id].qua === '0'}"
              class="button"
            >{{ pointDatas[scope.row.id].qua === '0' ? '正常': '异常('+ pointDatas[scope.row.id].qua+')' }}</span>
          </div>
          <div v-else>
            --
          </div>
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
            @click="showDetail(scope.row)"
          >
            查看测点信息
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <point-detail-modal
      :visible.sync="modalVisible"
      :point-info="currentPoint"
    />
  </div>
</template>

<script>
import PointDetailModal from '../components/pointDetailModal';
import { collectorApi } from '@@/config/cgi';
import moment from 'moment';
import getEdgeRequest from 'feature/utils/request';

export default {
  components: {
    PointDetailModal,
  },
  props: {
    collector: {
      type: Object,
      default: null,
    },
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      tableData: [],
      originTableData: [],
      vLoading: false,
      typeMap: {
        A: '浮点型',
        E: '枚举型',
        D: '状态量',
      },
      timerId: null,
      tableHeight: 0,
      currentPoint: {},
      modalVisible: false,
      pointIds: [],
      pointDatas: {},
      isDevice: false,
      deviceId: null,
      assigned: true,
      isSNMP: false,
    };
  },
  watch: {
    collector(v) {
      if (!v) return;
      const { type, id, isUnassigned, link_type: linkType } = v;
      this.isDevice = type === 'device';
      this.deviceId = id;
      this.assigned = !isUnassigned;
      this.linkType = linkType;
      this.isSNMP = linkType === 'SNMP';
      this.queryBaseInfo();
    },
    visible(v) {
      if (v) {
        this.queryPointInfo();
      }
    },
  },
  beforeDestroy() {
    clearInterval(this.timerId);
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
  },
  mounted() {
    this.tableHeight = window.innerHeight - this.$refs.point.offsetTop - 56;
    this.$nextTick(() => {
      this.observer = new ResizeObserver(() => {
        this.tableHeight = window.innerHeight - this.$refs.point.offsetTop - 56;
      });
      this.observer.observe(this.$refs.point);
    });
  },
  methods: {
    queryBaseInfo() {
      this.vLoading = true;
      const params = this.assigned ? {
        id: this.deviceId,
        assigned: this.assigned,
        device: this.isDevice,
        link_type: this.linkType,
      } : {
        id: this.deviceId,
        assigned: this.assigned,
        device: this.isDevice,
      };
      getEdgeRequest(this.$axios).post(collectorApi.queryPointInfo, params)
        .then((res) => {
          const points = this.isDevice ? res : res.filter(point => !(point.name.includes('串口状态') || point.name.includes('DI状态') || point.name.includes('DO状态')));
          this.pointIds = points.map(d => d.id);
          points.forEach((d) => {
            const { valdesc } = d.valdef;
            const valueMap = {};
            if (valdesc) {
              valdesc.split(',').forEach((item) => {
                const [key, value] = item.split('=');
                valueMap[key] = value;
              });
            }
            d.valueMap = valueMap;
          });
          this.tableData = points;
          console.log(this.tableData, 'this.tableData');
          this.originTableData = points;
          this.refreshData();
          this.vLoading = false;
        })
        .catch((err) => {
          this.vLoading = false;
          console.log(err);
        });
    },
    showDetail(row) {
      this.modalVisible = true;
      this.currentPoint = row;
    },
    queryPointInfo() {
      if (!this.visible) return;
      const params = this.assigned ? {
        ids: this.pointIds,
        assigned: this.assigned,
        link_type: this.linkType,
      } : {
        ids: this.pointIds,
        assigned: this.assigned,
      };
      getEdgeRequest(this.$axios).post(collectorApi.queryPointData, params, false)
        .then((res) => {
          _.forEach(res, (d) => {
            if (+d.tms) {
              d.formatTms = moment(+d.tms * 1000).format('yyyy-MM-DD HH:mm:ss');
            } else {
              d.formatTms = '--';
            }
            if (!Number.isInteger(+d.pv)) {
              d.pv = _.round(d.pv, 2);
            }
          });
          this.pointDatas = res;
          console.log(res, 'res');
        })
        .catch((err) => {
          console.log(err);
        });
    },
    refreshData() {
      this.queryPointInfo();
      clearInterval(this.timerId);
      this.timerId = setInterval(() => {
        this.queryPointInfo();
      }, 5000);
    },
    filterData(val) {
      this.tableData = this.originTableData.filter((data) => {
        if (val) {
          return data.no.includes(val) || data.name.includes(val);
        }
        return true;
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.button {
    display: inline-block;
    padding: 2px 5px;
    min-width: 60px;
    border: 1px solid;
    font-weight: 600;
    border-radius: 5px;
    text-align: center;
    border-color: #ffb84d;
    color: #ff9200;
    background: #fff7eb;
    &:hover {
       border-color: #ff9900;
        background: #ff9900;
        color: #fff;
    }

    &.normal {
      border-color: #19BE6B;
      color: #19BE6B;
      background: #e7f4ee;

      &:hover {
        border-color: #19be6b;
        background: #19be6b;
        color: #fff;
      }
    }

}
</style>
