
<template>
  <div
    v-loading="vloading"
    class="content-container"
  >
    <div class="info">
      <base-title title="连接信息" />
      <el-row
        type="flex"
        class="flex-wrap"
      >
        <el-col :span="24">
          <div class="grid">
            <div
              v-for="(info, index) in connectInfo"
              :key="index"
              class="info-item"
            >
              <span class="label">{{ info.label }}: </span>
              <span v-if="!info.render" class="value">{{
                Array.isArray(info.key)
                  ? baseInfo[info.key[0]] + "/" + baseInfo[info.key[1]]
                  : baseInfo[info.key] || "--"
              }}</span>
              <component
                v-if="info.render"
                :is="info.render"
                :data="baseInfo"
              />
            </div>
          </div>
        </el-col>
        <!-- <el-col :span="1">
          <el-button type="text" @click="handleEdit"> 重新分配 </el-button>
        </el-col> -->
      </el-row>
    </div>
    <div class="model">
      <base-title title="实物模型" />
      <div
        class="operate-buttons"
      >
        <el-button
          v-if="isVirtualBox"
          size="small"
          @click="replaceAsRealTbox"
        >
          替换为新上线TBOX
        </el-button>
        <el-button
          v-show="!isSNMP"
          size="small"
          @click="restartService"
        >
          重启服务
        </el-button>
        <el-button
          v-show="linkType !== 'SNMP'"
          size="small"
          @click="restartHost"
        >
          重启主机
        </el-button>
        <el-button
          v-show="!isSNMP"
          size="small"
          @click="remoteLogin"
        >
          远程登录
        </el-button>
      </div>
      <model-previewer
        model-url="./static/models/caijiqi.glb"
        :data="collectorPointsData"
        :ip="deviceIp"
        :is-s-n-m-p="isSNMP"
      />
    </div>
    <div
      ref="monitor"
      class="monitor"
    >
      <base-title title="接口监控" />
      <!-- <div
        v-if="isSNMP"
        class="total"
      >
        接入采集模板 <span style="color: #036fd3">{{ linkTemplate.length }}个</span>， 启用
        <span style="color: #3ecc46">{{ onlineTemplates }}个</span> ， 未启用
        <span style="color: #ff9200">{{ linkTemplate.length - onlineTemplates }}个</span>，涉及采集设备
        <span style="color: #036fd3">{{ collectorDeviceLength }}台</span>
      </div> -->
      <div class="total">
        接入采集设备 <span style="color: #036fd3">{{ devices.length }}台</span>， 在线
        <span style="color: #3ecc46">{{ onlineNum }}台</span> ， 离线
        <span style="color: #ff9200">{{ devices.length - onlineNum }}台</span>
      </div>
      <el-table
        v-if="maxTableHeight"
        :data="devices"
        :max-height="maxTableHeight"
        style="width: 100%"
      >
        <el-table-column
          prop="link_channel.chid"
          label="通道"
          width="180"
        />
        <el-table-column
          prop="name"
          label="采集设备"
        />
        <el-table-column
          prop="position"
          label="安装位置"
        >
          <template slot-scope="scope">
            {{ scope.row.position.room + "/" + scope.row.position.block }}
          </template>
        </el-table-column>
        <el-table-column
          prop="link_channel.addr"
          label="通讯地址"
          width="100"
        />
        <el-table-column
          prop="status"
          label="状态"
          width="80"
        >
          <template
            v-if="deviceStatus[scope.row.pointId]"
            slot-scope="scope"
          >
            <div v-if="collector && !collector.isOnline">
              --
            </div>

            <div
              v-else
              :class="{ online: deviceStatus[scope.row.pointId].pv }"
              class="alarm-button"
            >
              {{ deviceStatus[scope.row.pointId].pv ? '在线' : '离线' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column
          prop="updateTime"
          label="最近数据上报"
          width="180"
        >
          <template
            v-if="deviceStatus[scope.row.pointId]"
            slot-scope="scope"
          >
            <div v-if="collector && !collector.isOnline">
              --
            </div>

            <span v-else>{{ deviceStatus[scope.row.pointId].updateTime || "--" }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="150"
        >
          <template slot-scope="scope">
            <el-button
              type="text"
              @click="showDetail(scope.row)"
            >
              查看数据
            </el-button>
            <!-- <el-button type="text"> 关联到业务设备 </el-button> -->
          </template>
        </el-table-column>
      </el-table>
      <!-- <el-table
        v-if="maxTableHeight && isSNMP"
        :data="linkTemplate"
        :max-height="maxTableHeight"
        style="width: 100%"
      >
        <el-table-column
          prop="addr"
          label="从机地址"
        />
        <el-table-column
          prop="mapping"
          label="采集模板"
        />
        <el-table-column
          prop="desc"
          label="描述"
        />
        <el-table-column
          prop="enable"
          label="状态"
          width="80"
        >
          <template slot-scope="scope">
            <div
              :class="{ online: +scope.row.enable === 1 }"
              class="alarm-button"
            >
              {{ +scope.row.enable === 1 ? '启用' : '未启用' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="150"
        >
          <template slot-scope="scope">
            <el-button
              type="text"
              @click="showDetail(scope.row)"
            >
              查看数据
            </el-button>
          </template>
        </el-table-column>
      </el-table> -->
    </div>
    <point-modal
      :visible.sync="modalVisible"
      :device-id="collecorDeviceId"
      :assigned="assigned"
    />
  </div>
</template>

<script>
import BaseTitle from './baseTitle';
import { collectorApi } from '@@/config/cgi';
import PointModal from './pointModal';
import ModelPreviewer from './model-previewer';
import getEdgeRequest from 'feature/utils/request';

export default {
  components: {
    BaseTitle,
    PointModal,
    ModelPreviewer,
  },
  props: {
    collector: {
      type: Object,
      default: () => null,
    },
    deviceStatus: {
      type: Object,
      default: () => null,
    },
  },
  data() {
    return {
      vloading: false,
      connectInfo: [// 连接信息
        { label: '采集器IP', key: 'link_channel.chid' },
        { label: '设备SN', key: 'profile.sn' },
        { label: '安装位置', key: ['position.room', 'position.block'] },
        {
          label: '硬件版本',
          key: 'extend.hardware_version',
          render: {
            functional: true,
            render: (h) => {
              const {
                isVirtualBox,
                baseInfo,
                collector,
              } = this;

              const version = _.get(baseInfo, 'extend.hardware_version');

              const children = [];

              if (!isVirtualBox || collector?.isUnassigned) {
                children.push(version || '--');
              }

              if (!collector?.isUnassigned) {
                if (isVirtualBox) {
                  children.push(
                    <el-tag
                      type="warning"
                      size="small"
                    >虚拟TBOX</el-tag>
                  );
                }

                if (collector?.unssignedTbox) {
                  children.push(
                    <el-tag
                      type="success"
                      size="small"
                      class="blink-animate"
                    >
                      发现新上线TBOX
                    </el-tag>
                  );
                }
              }

              return (<div class="hardware-version">
                { children }
              </div>);
            }
          }
        },
        { label: '内核版本', key: 'extend.kernel_version' },
        { label: '软件版本', key: 'extend.software_version' },
      ],
      baseInfo: {}, // 基本信息
      devices: [],
      devicePoints: [],
      maxTableHeight: 0,
      // deviceStatus: {},
      timerId: null,
      onlineNum: 0,
      modalVisible: false,
      collecorDeviceId: '',
      collectorPoints: [],
      collectorPointsData: {},
      assigned: false,
      stateId: '',
      deviceIp: '',
      isSNMP: false,
      linkTemplate: [], // 第三方采集器采集模板
      onlineTemplates: 0,
      collectorDeviceLength: 0,
      linkType: '',
    };
  },
  computed: {
    isVirtualBox() {
      return _.get(this.baseInfo, 'extend.hardware_version') === 'virtual';
    },
  },
  watch: {
    collector: {
      handler(v) {
        if (v) {
          const { type, id, isUnassigned, comm_state_id: stateId, ip, link_type: linkType } = v;
          this.deviceIp = ip;
          this.isDevice = type === 'device';
          this.deviceId = id;
          this.assigned = !isUnassigned;
          this.stateId = stateId;
          this.linkType = linkType;
          this.isSNMP = Boolean(linkType?.includes('SNMP'));
          this.collectorDeviceLength = v.children?.length || 0;
          this.getBaseInfo();
        }
      },
      immediate: true,
    },
    deviceStatus: {
      handler() {
        this.computeNumber();
      },
      deep: true,
    },
    devices: {
      handler() {
        this.computeNumber();
      },
      deep: true,
    },
  },
  mounted() {
    this.$nextTick(() => {
      this.observer = new ResizeObserver(() => {
        const { top } = this.$refs.monitor.getBoundingClientRect();
        this.maxTableHeight = Math.max(window.innerHeight - top - 94 - 56, 200);
      });
      this.observer.observe(this.$refs.monitor);
    });
  },
  beforeDestroy() {
    clearInterval(this.timerId);
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
  },
  methods: {
    computeNumber() {
      this.onlineNum = 0;
      _.forEach(this.devices, (device) => {
        if (this.deviceStatus[device.pointId] && this.deviceStatus[device.pointId].pv) {
          this.onlineNum = this.onlineNum + 1;
        }
      });
    },
    getBaseInfo() {
      this.vloading = true;
      const params = this.assigned ? {
        id: this.deviceId,
        assigned: this.assigned,
        link_type: this.linkType,
      } : {
        id: this.deviceId,
        assigned: this.assigned,
      };
      getEdgeRequest(this.$axios)
        .post(collectorApi.queryCollectorDetail, params, false)
        .then((res) => {
          this.vloading = false;
          this.baseInfo = this.traverseObject(res);
          // if (this.isSNMP) {
          //   this.linkTemplate = this.baseInfo.link_template;
          //   this.linkTemplate.sort((a, b) => +a.addr - +b.addr);
          //   this.onlineTemplates = this.linkTemplate.filter(item => +item.enable === 1).length;
          // } else {
          this.devices = (res.devices || []).map((device) => {
            device.pointId = device.comm_state_id;
            delete device.comm_state_id;
            return device;
          });
          this.devices.sort((a, b) => {
            const { chid: chidA, addr: addrA } = a.link_channel;
            const { chid: chidB, addr: addrB } = b.link_channel;
            if (!chidA) return -1;
            if (!chidB) return 1;
            if (chidA.includes('COM') && chidB.includes('COM')) { // 都是网口
              if (chidA === chidB) {
                return (+addrA) - (+addrB);
              }
              return (+chidA.slice(3)) - (+chidB.slice(3));
            }
            if (!chidA.includes('COM') && !chidB.includes('COM')) { // 都是串口
              const ipA = chidA.split(':')[0].split('.').map(el => el.padStart(3, '0'))
                .join('') + chidA.split(':')[1];
              const ipB = chidB.split(':')[0].split('.').map(el => el.padStart(3, '0'))
                .join('') + chidB.split(':')[1]; ;
              return ipA - ipB;
            }
            return chidA.includes('COM') ? 1 : -1;
          });
          // }
          const { di, do: dom, power } = res.state_id;
          this.collectorPoints = [...di, ...dom, ...power];
          this.refreshCollectorData();
        })
        .catch((err) => {
          this.vloading = false;
          console.log(err);
        });
    },

    refreshCollectorData() {
      if (this.isSNMP) {
        this.collectorPointsData = {};
      }
      this.queryCollectorData();
      clearInterval(this.timerId);
      this.timerId = setInterval(() => {
        this.queryCollectorData();
      }, 5000);
    },
    queryCollectorData() {
      if (this.isSNMP) return;
      if (!this.collectorPoints.length || !this.collector) return;
      getEdgeRequest(this.$axios).post(collectorApi.queryPointData, {
        ids: [...this.collectorPoints],
        assigned: this.assigned,
      }, false).then((res) => {
        const object = {};
        _.forEach(res, (val, key) => {
          object[key.split('.').slice(-1)[0]] = val;
        });
        this.collectorPointsData = {
          ...object,
          online: this.deviceStatus[this.stateId],
        };
      })
        .catch((err) => {
          console.log(err);
        });
    },
    restartService() {
      this.loading = true;
      this.$axios
        .post(collectorApi.restartCollector, {
          id: this.deviceId,
        })
        .then(() => {
          this.loading = false;
          this.$message.success('重启服务成功');
        })
        .catch(() => {
          this.loading = false;
          this.$message.error('重启服务失败');
        });
    },
    restartHost() {
      this.loading = true;
      this.$axios
        .post(collectorApi.restartOS, {
          id: this.deviceId,
        })
        .then(() => {
          this.loading = false;
          this.$message.success('重启主机成功');
        })
        .catch(() => {
          this.loading = false;
          this.$message.error('重启主机失败');
        });
    },
    replaceAsRealTbox() {
      const { unssignedTbox } = this.collector || {};

      if (!unssignedTbox) {
        this.$message.error('没有新上线的TBOX，或新上线TBOX已下线');
        return;
      }

      this.loading = true;
      this.$axios
        .post('/api/dcos/tboxmonitor-cgi/collector/swap', {
          assigned_id: this.deviceId,
          unassigned_id: unssignedTbox?.id,
        })
        .then(() => {
          this.loading = false;
          this.$message.success('替换为新上线TBOX成功');
          this.collector.unssignedTbox = null;
          this.getBaseInfo();
        })
        .catch(() => {
          this.loading = false;
          // this.$message.error('替换为新上线TBOX失败');
        });
    },
    remoteLogin() {
      window.open(`http://${this.baseInfo['link_channel.chid']}`);
    },
    showDetail(row) {
      this.collecorDeviceId = row.id;
      this.modalVisible = true;
    },

    traverseObject(obj) {
      const flatObj = {};
      _.forEach(obj, (value, key) => {
        if (value instanceof Object && !Array.isArray(value)) {
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

    handleEdit() {
      const testId = 3;
      this.$emit('edit', testId);
    },
  },
};
</script>

<style lang="scss" scoped>
.content-container {
  padding: 0 32px;
  display: flex;
  flex-direction: column;
  .flex-wrap {
    align-items: start;
    z-index: 99;
  }
  .grid {
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
.model {
  position: relative;
  height: 30vh;
  .operate-buttons {
    position: absolute;
    top: 16px;
    right: 0;
    z-index: 99;
    /deep/ .el-button {
      border-radius: 4px;
    }
  }
}
.monitor {
  flex: 1;
  .total {
    margin-bottom: 16px;
    font-family: TencentSansW3;
    color: #333;
    font-weight: 800;
  }
}
.alarm-button {
  width: 40px;
  border: 1px solid;
  border-radius: 5px;
  text-align: center;
  border-color: #ff9200;
  color: #ff9200;
  display: flex;
  align-items: center;
  justify-content: space-around;
  &.online {
    border-color: #3ecc46;
    color: #3ecc46;
  }
}

.hardware-version {
  display: inline-block;
  & > * {
    margin-left: 8px;
  }
}

.blink-animate {
  animation: blink-animate 1s infinite;
}

@keyframes blink-animate {
  from {
    opacity: 1;
  }
  50% {
    opacity: 0.2;
  }
  to {
    opacity: 1;
  }
}

</style>
