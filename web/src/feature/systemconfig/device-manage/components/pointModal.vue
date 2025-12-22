
<template>
  <el-modal
    :visible.sync="modalVisible"
    :width="600"
  >
    <template slot="title">
      设备测点数据
    </template>
    <el-table :data="pointInfos">
      <el-table-column
        type="index"
        width="80"
      />
      <el-table-column
        prop="name"
        label="测点名称"
      />
      <el-table-column
        prop="no"
        label="标识符"
      />
      <el-table-column
        prop="name"
        label="采集值"
      >
        <template
          v-if="pointDatas[scope.row.id]"
          slot-scope="scope"
        >
          <span
            :class="{normal: pointDatas[scope.row.id].qua === '0'}"
            class="button"
          >
            {{ scope.row.valtype === 'A'
              ? pointDatas[scope.row.id].pv + scope.row.valdef.unit
              : scope.row.valueMap[pointDatas[scope.row.id].pv] }}
          </span>
        </template>
      </el-table-column>
      <el-table-column
        prop="valdef.valdesc"
        label="枚举描述"
      />
      <!-- <el-table-column
        prop="channel"
        label="说明"
      /> -->
    </el-table>
  </el-modal>
</template>

<script>
import { collectorApi } from '@@/config/cgi';
import getEdgeRequest from 'feature/utils/request';

export default {
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    deviceId: {
      type: String,
      default: '',
    },
    assigned: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      pointInfos: [],
      pointIds: [],
      pointDatas: {},
    };
  },
  computed: {
    modalVisible: {
      set(v) {
        this.$emit('update:visible', v);
      },
      get() {
        return this.visible;
      },
    },
  },
  watch: {
    modalVisible(v) {
      if (v) {
        this.queryPointInfo();
      }
    },
  },
  methods: {
    queryPointInfo() {
      getEdgeRequest(this.$axios)
        .post(collectorApi.queryPointInfo, {
          id: this.deviceId,
          assigned: this.assigned,
          device: true,
        })
        .then((res) => {
          this.pointIds = res.map(d => d.id);
          res.forEach((d) => {
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
          this.pointInfos = res;
          this.queryData();
        })
        .catch((err) => {
          console.log(err);
        });
    },
    queryData() {
      getEdgeRequest(this.$axios).post(collectorApi.queryPointData, {
        assigned: this.assigned,
        ids: this.pointIds,
      }).then((res) => {
        this.pointDatas = res;
      })
        .catch((err) => {
          console.log(err);
        });
    },
    // todo 测点数值
  },
};
</script>
