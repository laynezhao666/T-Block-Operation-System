
<template>
  <div
    v-loading="vloading"
    class="content-container"
  >
    <base-title title="连接信息" />
    <div class="form">
      <div
        v-for="(info, index) in connectForm"
        :key="index"
        class="info-item"
      >
        <span class="label">{{ info.label }}: </span>
        <span class="value">{{ baseInfo[info.key] || "--" }}</span>
      </div>
    </div>
    <base-title title="设备信息" />
    <div class="form">
      <div
        v-for="(info, index) in deviceForm"
        :key="index"
        class="info-item"
      >
        <span class="label">{{ info.label }}: </span>
        <span class="value">{{
          Array.isArray(info.key)
            ? baseInfo[info.key[0]] + "/" + baseInfo[info.key[1]]
            : baseInfo[info.key] || "--"
        }}</span>
      </div>
    </div>
    <base-title title="接口监控" />
    <el-table
      :data="points"
      style="width: 100%"
    >
      <el-table-column
        prop="name"
        label="接口指标"
      />
      <el-table-column
        prop="status"
        label="值"
      >
        <template
          v-if="deviceStatus[scope.row.id]"
          slot-scope="scope"
        >
          <div v-if="collector && collector.collector && !collector.collector.isOnline">
            --
          </div>

          <span
            v-else
          >
            {{ deviceStatus[scope.row.id].pv }}
          </span>
        </template>
      </el-table-column>
      <el-table-column
        prop="updateTime"
        label="刷新时间"
      >
        <template
          v-if="deviceStatus[scope.row.id]"
          slot-scope="scope"
        >
          <div v-if="collector && collector.collector && !collector.collector.isOnline">
            --
          </div>
          <span v-else>{{ deviceStatus[scope.row.id].updateTime || "--" }}</span>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import BaseTitle from './baseTitle';
import { collectorApi } from '@@/config/cgi';
import { devicebaseInfo } from '../mockData.js';
import moment from 'moment';
import getEdgeRequest from 'feature/utils/request';

export default {
  components: {
    BaseTitle,
  },
  props: {
    collector: {
      type: Object,
      default: () => null,
    },
  },
  data() {
    return {
      vloading: false,
      connectForm: [
        { label: '通讯协议', key: 'profile.thing_template.protocol' },
        { label: '端口', key: 'link_channel.chid' },
        { label: '通讯地址', key: 'link_channel.addr' },
        { label: '通讯参数', key: 'link_channel.chparams' },
        { label: '超时时间', key: 'link_channel.max_fail_time' },
        { label: '中断判定次数', key: 'link_channel.max_fail_count' },
        { label: '采集模板', key: 'profile.thing_template.tplnm' },
        { label: '总测点', key: 'profile.thing_template.total_point' },
      ],
      pointForm: [
        { label: '采集模板', key: 'profile.thing_template.tplnm' },
        { label: '总测点', key: 'profile.thing_template.total_point' },
        // { label: '应采集测点', key: 'hardware_version' },
        // { label: '必要测点', key: 'sn' },
        // { label: '可选测点', key: 'kernel_version' },
        // { label: '最近采集时间', key: 'software_version' },
      ],
      deviceForm: [
        { label: '安装位置', key: ['position.room', 'position.block'] },
        { label: '厂商', key: 'profile.vendor' },
        { label: '型号', key: 'profile.model' },
        { label: '设备序列号', key: 'profile.sn' },
        { label: '描述', key: 'desc' },
      ],
      baseInfo: {},
      timerId: null,
      points: [],
      devicePoints: [],
      deviceStatus: {

      },
    };
  },
  watch: {
    collector: {
      handler(v) {
        if (v) {
          this.queryDetail();
          this.refreshData();
        }
      },
      immediate: true,
    },
  },
  beforeDestroy() {
    clearInterval(this.timerId);
  },
  methods: {
    queryDetail() {
      getEdgeRequest(this.$axios)
        .post(collectorApi.queryDeviceDetail, {
          id: this.collector.id,
          assigned: !this.collector.isUnassigned,
        })
        .then((res) => {
          this.points = res?.state?.metrics || [];
          this.devicePoints = this.points.map(point => point.id);
          this.baseInfo = this.traverseObject(res);
          this.refreshData();
        })
        .catch((err) => {
          // this.baseInfo = this.traverseObject(devicebaseInfo);
          console.log(err);
        });
    },
    traverseObject(obj) {
      const flatObj = {};
      _.forEach(obj, (value, key) => {
        if (value instanceof Object) {
          const childObj = this.traverseObject(value);
          _.forEach(childObj, (v, k) => {
            flatObj[`${key}.${k}`] = v;
          });
        } else {
          flatObj[key] = value;
        }
      });
      return flatObj;
    },
    queryPointData() {
      if (!this.devicePoints.length || !this.collector) return;
      console.log('设备-基本信息');
      getEdgeRequest(this.$axios).post(collectorApi.queryPointData, {
        ids: [...this.devicePoints],
        assigned: !this.collector.isUnassigned,
      }, false).then((res) => {
        _.forEach(res, (val) => {
          val.updateTime = moment(+val.tms * 1000).format('yyyy-MM-DD HH:mm:ss');
        });
        this.deviceStatus = res;
      })
        .catch((err) => {
          console.log(err);
        });
    },
    refreshData() {
      this.queryPointData();
      clearInterval(this.timerId);
      this.timerId = setInterval(() => {
        this.queryPointData();
      }, 5000);
    },
  },
};
</script>

<style lang="scss" scoped>
.content-container {
  padding: 0 32px;
  .form {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    .info-item {
      padding: 5px 0 15px 0;
      .label {
        width: 100px;
        text-align: right;
        display: inline-block;
        margin-right: 12px;
        color: #bfbcbc;
        font-weight: 800;
      }
      .value {
        color: #333;
      }
    }
  }
}
</style>
