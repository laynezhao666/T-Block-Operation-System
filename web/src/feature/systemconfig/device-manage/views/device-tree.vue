<template>
  <div>
    <div style="display: flex; padding-left:8px;">
      <div class="search-wrap">
        <el-select
          v-model="typeOption"
          border-type="bordered"
        >
          <el-option
            v-for="item in typeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-select
          v-model="statusOption"
          border-type="bordered"
        >
          <el-option
            v-for="item in statusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>

        <el-input
          v-model="filterText"
          class="c-tree-tools__input"
          placeholder="请输入"
          suffix-icon="tn-icon-search"
          clearable
          border-type="bordered"
        />

        <el-dropdown
          @command="handleBatchAction"
        >
          <el-space :size="4">
            <i class="tn-icon-more" />
          </el-space>
          <el-dropdown-menu slot="dropdown">
            <el-dropdown-item
              v-for="(action, i) in batchActions"
              :key="i"
              :command="action.key"
            >
              {{ action.label }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown>
      </div>
    </div>
    <el-tree
      ref="tree"
      :key="treeVersion"
      :data="treeData"
      :props="defaultProps"
      :filter-node-method="filterNode"
      node-key="id"
      :default-expanded-keys="expandedkeys"
      :current-node-key="currentNodekey"
      highlight-current
      :expand-on-click-node="false"
      style="overflow: scroll;"
      :style="{height: treeHeight+'px'}"
      @node-click="handleNodeClick"
      @node-expand="handleNodeExpand"
      @node-collapse="handleNodeCollapse"
    >
      <span
        slot-scope="{ node, data }"
        class="custom-tree-node"
      >
        <el-popover
          placement="top"
          width="150"
          trigger="hover"
        >
          <p>{{ data.chid || data.ip }}</p>

          <tbox-icon
            v-if="data.type === 'collector'"
            slot="reference"
            :is-snmp="agentLinkTypes.includes(data.link_type)"
            :color="resolveTboxIconColor(data)"
            :blink-animate="data.unssignedTbox"
            style="margin-right: 6px; height: 18px"
          />

          <img
            v-if="data.type === 'device' && data.chtype"
            slot="reference"
            :src="require(`../../assets/${(data.collector && data.collector.isOnline) && pointValue[data.comm_state_id]
              ? `${data.chtype}` : `${data.chtype}-offline`}.svg`)"
            style="margin-right: 6px; height: 18px"
          >
          <img
            v-if="data.type === 'device' && !data.chtype"
            slot="reference"
            :src="require(`../../assets/${(data.collector && data.collector.isOnline) && pointValue[data.comm_state_id]
              ? 'socket.svg' : 'socket-offline.svg'}`)"
            style="margin-right: 6px; height: 18px"
          >

        </el-popover>

        <span v-if="rootTreeType.includes(data.type)">
          {{ data.name }} | {{ data.count }} 台
        </span>
        <el-badge
          :value="data.new_discovered ? 'new' : ''"
          class="item"
        >
          <span v-if="data.type === 'collector'">
            {{ data.name }} | {{ data.ip }}
          </span>
          <span v-if="data.type === 'device'">
            {{ data.device_type }} | {{ data.name }}
          </span>
        </el-badge>
        <!-- 红点提示 -->
        <div
          v-if="!data.new_discovered && (data.hasNewDiscovered || data.has_new_discovered)"
          class="red"
        />

        <el-button
          v-if="data.isUnassigned && ['collector'].includes(data.type)"
          type="text"
          style="margin-left: 12px;"
          @click="(e) => handleEdit(e,data)"
        >
          <i class="el-icon-edit" />
        </el-button>
      </span>
    </el-tree>

    <import-config-modal
      ref="importConfigModal"
      :tree-data="treeData"
    />
  </div>
</template>

<script>
import { collectorApi, tbosCollectorApi } from '@@/config/cgi';
// import { assignedCollectorList } from '../mockData.js';
import moment from 'moment';
import { forEachTreeNode } from '../../../../utils/tree';
import TboxIcon from '../components/tbox-icon.vue';
import ImportConfigModal from './import-config-modal.vue';
import { has } from 'lodash';
import getEdgeRequest from 'feature/utils/request';

export default {
  components: {
    TboxIcon,
    ImportConfigModal,
  },
  props: {
    height: {
      type: Number,
      default: 0,
    },
  },
  data() {
    return {
      agentLinkTypes: ['SNMP', 'SNMP_AGENT'],
      rootTreeType: ['mozu', 'room', 'block', 'unassinged'],
      defaultProps: {
        children: 'children',
        label: 'name',
        name: 'name',
        type: 'type',
        number: 'number',
      },
      treeData: [],
      currentNodekey: '',
      expandedkeys: [],
      currentCollector: null,
      collectorPointMap: [], // 采集器-状态测点表
      timerId: null,
      refreshTimerId: null,
      filterText: '',
      statusOption: 'all',
      statusOptions: [
        {
          value: 'all',
          label: '全部',
        },
        {
          value: '在线',
          label: '在线',
        },
        {
          value: '离线',
          label: '离线',
        },
        {
          value: '未分配',
          label: '未分配',
        },
        {
          value: '新发现',
          label: '新发现',
        },
        {
          value: '未上线过',
          label: '未上线过',
        },
      ],
      typeOption: 'all',
      typeOptions: [
        {
          value: 'all',
          label: '全部',
        },
        {
          value: 'TLINK',
          label: '自研',
        },
        {
          value: 'SNMP',
          label: '三方',
        },
      ],
      assignedIds: [],
      unassingedIds: [],
      pointValue: {},
      deviceStatus: {},

      treeVersion: 0,

      batchActions: [{
        key: 'importCollectConfig',
        label: '更新采集配置',
        func: () => {
          this.$refs.importConfigModal.open('CollectConfig');
        },
      }, {
        key: 'importDriverConfig',
        label: '更新驱动模板',
        func: () => {
          this.$refs.importConfigModal.open('DriverConfig');
        },
      }],
    };
  },
  computed: {
    treeHeight() {
      return this.height - 108;
    },
  },
  watch: {
    currentCollector(v) {
      if (v) {
        this.$emit('checkDevice', v);
      }
    },
    statusOption(v) {
      this.$refs.tree.filter({ type: 'status', val: v });
    },
    typeOption(v) {
      this.$refs.tree.filter({ type: 'type', val: v });
    },
    filterText(v) {
      this.$refs.tree.filter({ type: 'name', val: v });
    },
    deviceStatus(v) {
      this.$emit('updateStatus', v);
    },
    // treeData() {
    //   const {
    //     treeData,
    //   } = this;

    //   let currentCollector = null;

    //   forEachTreeNode(treeData, (node) => {
    //     if (node.type === 'collector') {
    //       currentCollector = node;
    //       this.$set(node, 'collector', node);
    //     } else {
    //       this.$set(node, 'collector', currentCollector);
    //     }
    //   });
    // },
    pointValue() {
      const {
        treeData,
        pointValue,
      } = this;

      forEachTreeNode(treeData, (node) => {
        if (node.type !== 'collector') return;

        this.$set(node, 'isOnline', pointValue[node.comm_state_id]);
      });
    },
  },
  mounted() {
    this.refreshCollector();
  },
  beforeDestroy() {
    clearInterval(this.timerId);
    this.timerId = null;
    clearInterval(this.refreshTimerId);
    this.refreshTimerId = null;
  },
  methods: {
    refreshCollector() {
      this.queryAssignedCollectorList();
      this.queryUnassignedCollectorList();
    },
    queryAssignedCollectorList() { // 查询已分配设备
      const url = window.tnwebServices.isTbos ? tbosCollectorApi.GetCollectorStatusTree : collectorApi.queryCollectorAssignedListNew;
      getEdgeRequest(this.$axios).get(url, null, false, false)
        .then((res) => {
          const { _lastQueryAssignedCollectorListResponse } = this;
          const isResChange = !_.isEqual(res, _lastQueryAssignedCollectorListResponse);
          this._lastQueryAssignedCollectorListResponse = _.cloneDeep(res);

          // 设置设备所属的采集器对象/节点，子设备的在线离线状态需要通过采集器的状态判断（前端判断采集器离线，子设备也要设置为离线）
          let belongCollector = null;
          forEachTreeNode([res], (node) => {
            const keyMap = {
              type: 'collector_type',
              name: 'device_name',
              device_type: 'device_type_en',
              machine_type: '',
              ip: 'channel_id',
              id: 'device_gid',
              count: 'device_count',
              comm_state_id: 'comm_state_id',
              chid: 'channel_id',
              has_been_collected: 'has_been_collected',
              collected_time: 'collected_time',
              chtype: 'channel_type',
            };
            for (const key in keyMap) {
              if (has(node, keyMap[key])) {
                if (key === 'type') {
                  node[key] = node[keyMap[key]] === 1 ? 'collector' : 'device';
                } else {
                  node[key] = node[keyMap[key]];
                }
              }
            }
            node.link_type = 'TLINK';
            if (node.type === 'collector') {
              belongCollector = node;
              this.$set(node, 'collector', node);
            } else {
              this.$set(node, 'collector', belongCollector);
            }
          });

          // 打补丁先吧
          if (this.treeData.length && this.treeData[0].name !== '未分配') { // 已存在
            if (!isResChange) return;
            this.treeVersion += 1;
            const treeScrollTop = this.$refs.tree.$el.scrollTop;
            setTimeout(() => {
              this.$refs.tree.$el.scrollTop = treeScrollTop;
            }, 0);
            const oldTreeMap = {};
            let { currentCollector } = this;

            forEachTreeNode([this.treeData[0]], (node) => {
              oldTreeMap[node.id] = node;
            });
            forEachTreeNode([res], (node) => {
              if (node.id === currentCollector.id) {
                currentCollector = node;
              }

              const oldNode = oldTreeMap[node.id];

              if (!oldNode) { // new node
                return;
              }

              Object.assign(oldNode);
              _.forEach(oldNode, (v, k) => {
                if (node.hasOwnProperty(k)) return;
                node[k] = v;
              });
            });

            this.treeData[0] = res;

            if (currentCollector) { // 更新当前选中采集器
              this.currentCollector = currentCollector;
            }
          } else { // 未存在
            this.treeData.unshift(res);
            this.assignedIds = this.traverseCollector(this.treeData[0].children);
            if (!this.currentCollector) {
              this.currentCollector = res.children[0];
              this.currentNodekey = this.currentCollector?.id;
              this.expandedkeys.push(res.id);
              this.expandedkeys.push(res.children[0].id);
            }
          }
        })
        .catch((err) => {
          console.log(err);
        });
    },
    traverseCollector(data) {
      const arr = [];
      _.forEach(data, (d) => {
        const { comm_state_id: stateId, link_type: linkType } = d;
        if (stateId) {
          arr.push(stateId);
        }
        if (linkType && d.children) {
          d.children.forEach((child) => {
            child.link_type = linkType;
          });
        }
        if (d.children) {
          arr.push(...this.traverseCollector(d.children));
        }
      });
      return arr;
    },

    // 如何更新未分配设备
    refreshUnassignedCollector() {
      clearInterval(this.refreshTimerId);
      this.refreshTimerId = setInterval(() => {
        // this.queryUnassignedCollectorList();
        this.refreshCollector();
      }, 2000);
    },
    queryUnassignedCollectorList() { // 查询未分配设备
        const res = []
        const total = res.length || 0;
        res.forEach((collect) => {
          collect.isUnassigned = true;
          collect.children.forEach((child) => {
            child.isUnassigned = true;
          });
        });
        const { length } = this.treeData;
        if (length && this.treeData[length - 1].name === '未分配') {
          this.updateCollector(res);
          const hasNewDiscovered = Boolean(res.filter(r => r.has_new_discovered || r.new_discovered).length);
          this.treeData[this.treeData.length - 1].hasNewDiscovered = hasNewDiscovered;
        } else {
          const hasNewDiscovered = Boolean(res.filter(r => r.has_new_discovered || r.new_discovered).length);
          // 初始化
          this.treeData.push({
            id: 'xx',
            name: '未分配',
            type: 'unassinged',
            count: total,
            children: res,
            isUnassigned: true,
            hasNewDiscovered,
          });
          this.expandedkeys.push('xx');
          this.refreshStatus();
          this.refreshUnassignedCollector();
        }

        this.unassingedIds = this.traverseCollector(this.treeData[this.treeData.length - 1].children);

        this.updateCollectorHasUnssignStatus(res);
    },
    updateCollector(newValue) {
      this.treeData[this.treeData.length - 1].count = newValue.length;
      const unassingedCollectorList = this.treeData[this.treeData.length - 1].children;

      newValue.forEach((collector, index) => {
        let oldCollector = unassingedCollectorList[index];
        if (!oldCollector) { // 新增的采集器
          unassingedCollectorList.push(collector);
        } else if (collector.id !== oldCollector.id) {
          oldCollector = collector;
          unassingedCollectorList[index] = collector;
        } else {
          oldCollector.new_discovered = collector.new_discovered;
          oldCollector.has_new_discovered = collector.has_new_discovered;
          collector.children.forEach((device, index) => {
            let oldDevice = oldCollector.children[index];
            if (!oldDevice) {
              oldCollector.children.push(device);
            } else if (device.id !== oldDevice.id) {
              oldDevice = device;
            } else {
              oldDevice.new_discovered = device.new_discovered;
            }
          });
          if (collector.children.length !== oldCollector.children.length) {
            oldCollector.children.slice(collector.children.length);
          }
        }
      });
      if (unassingedCollectorList.length !== newValue.length) {
        unassingedCollectorList.splice(newValue.length);
      }
    },
    updateCollectorHasUnssignStatus(unssignedTboxList) {
      const registedTboxList = this.treeData?.[0]?.children;
      const unssignedTboxMap = _.chain(unssignedTboxList)
        .filter(item => !!this.pointValue[item.comm_state_id])
        .mapKeys('ip')
        .value();
      registedTboxList.forEach((tbox) => {
        this.$set(tbox, 'unssignedTbox', unssignedTboxMap[tbox.ip]);
      });
    },
    queryPoints(params) { // 查询采集器/设备在线、离线状态
      getEdgeRequest(this.$axios).post(collectorApi.queryPointData, params, false, false)
        .then((res) => {
          const status = {};
          _.forEach(res, (val) => {
            val.pv = !+val.pv;
            status[val.id] = val.pv && val.qua === '0';
            val.updateTime = moment(+val.tms * 1000).format('yyyy-MM-DD HH:mm:ss');
          });
          this.pointValue = {
            ...this.pointValue,
            ...status,
          };
          this.deviceStatus = {
            ...this.deviceStatus,
            ...res,
          };
        })
        .catch((err) => {
          console.log(err);
        });
    },
    queryStatus() {
      if (this.assignedIds.length) {
        this.queryPoints({ ids: this.assignedIds, assigned: true });
      }
      if (this.unassingedIds.length) {
        this.queryPoints({ ids: this.unassingedIds, assigned: false });
      }
    },
    refreshStatus() {
      this.queryStatus();
      clearInterval(this.timerId);
      this.timerId = null;
      this.timerId = setInterval(() => {
        this.queryStatus();
      }, 5000);
    },
    handleNodeClick(data) {
      if (data.type === 'mozu') {
        this.currentCollector = data;
        this.currentNodekey = data.id;
        return;
      }

      if (!this.rootTreeType.includes(data.type)) {
        this.currentCollector = data;
        this.currentNodekey = data.id;
      }
      if (data.new_discovered && data.type === 'collector') {
        this.$axios.post(collectorApi.resetStatus, {
          id: data.id,
        }, false, false).then(() => {
          data.new_discovered = false;
          this.queryUnassignedCollectorList();
        })
          .catch((err) => {
            console.log(err);
          });
      } else if (data.new_discovered && data.type === 'device') {
        this.$axios.post(collectorApi.resetDeviceStatus, {
          id: data.id,
        }, false, false).then(() => {
          data.new_discovered = false;
          this.queryUnassignedCollectorList();
        })
          .catch((err) => {
            console.log(err);
          });
      }
    },
    handleNodeExpand(data) {
      this.expandedkeys.push(data.id);
    },
    handleNodeCollapse(data) {
      const index = this.expandedkeys.indexOf(data.id);
      this.expandedkeys.splice(index, 1);
    },
    filterNode(value, data) {
      if (!value) return true;
      if (value.type === 'status') {
        if (value.val === '未分配') {
          return data.isUnassigned;
        } if (value.val === '在线') {
          return this.pointValue[data.comm_state_id];
        } if (value.val === '离线') {
          return !this.pointValue[data.comm_state_id];
        } if (value.val === '新发现') {
          return data.new_discovered;
        } if (value.val === '未上线过') {
          return data.has_been_collected === 0;
        }
        return true;
      } if (value.type === 'type') {
        if (value.val === 'SNMP') {
          return this.agentLinkTypes.includes(data.link_type);
        } if (value.val === 'all') {
          return true;
        } if (value.val === 'TLINK') {
          return !this.agentLinkTypes.includes(data.link_type);
        }
      }
      return data.name?.includes(value.val);
    },
    resolveTboxIconColor(data) {
      if (!this.pointValue[data.comm_state_id]) return '#cccccc';

      if (data.unssignedTbox) return '#3ecc46';

      return Number(data.machine_type) === 2 ? '#ed7b2f' : '#1296db';
    },
    handleEdit(e, data) {
      e.stopPropagation();
      this.$emit('edit', data);
    },
    handleBatchAction(command) {
      const {
        batchActions,
      } = this;

      return _.find(batchActions, {
        key: command,
      })?.func?.();
    },
  },
};
</script>

<style lang="scss" scoped>
.custom-tree-node {
  padding-right: 12px;
  display: flex;
  align-items: center;
}
/deep/ .el-tree > .el-tree-node {
  min-width: 100%;
  display: inline-block;
}
.search-wrap {
  // border-bottom: solid 1px #c0c0c0;
  display: flex;
  align-items: center;
  padding: 5px 10px 5px 5px;
  margin: 5px 0;
  flex: 1;
  .el-select {
    width: 130px;
    margin-right: 10px;
  }
  /deep/ .el-select .el-input__inner {
    border-radius: 2px;
  }
  .el-input {
    margin-right: 5px;
  }
  /deep/ .el-input .el-input__inner {
    border-radius: 2px;
  }
  .add {
    padding: 0px 5px;
  }
  .more {
    padding: 0px 5px;
    cursor: pointer;
  }
}
/deep/ .el-popper {
  min-width: 50px;
  border-radius: 8px;
}

.item {
  margin-right: 10px;
  /deep/ .el-badge__content.is-fixed {
    right: 4px;
    top: -3px;
  }
}
.red {
  border: 5px solid red;
  border-radius: 50%;
}
</style>

<!-- <el-button
          type="text"
          class="add"
          @click="addCollector"
        >
          <i class="el-icon-plus" />添加
        </el-button>
        <el-dropdown
          class="more"
          @command="handleCommand"
        >
          <span class="el-dropdown-link">
            <i class="el-icon-more" />
          </span>
          <el-dropdown-menu slot="dropdown">
            <el-dropdown-item command="自动发现">
              自动发现
            </el-dropdown-item>
            <el-dropdown-item command="下载到采集器">
              下载到采集器
            </el-dropdown-item>
            <el-dropdown-item command="从采集器上传">
              从采集器上传
            </el-dropdown-item>
            <el-dropdown-item command="批量导入">
              批量导入
            </el-dropdown-item>
            <el-dropdown-item command="批量导出">
              批量导出
            </el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown> -->
<!-- <span v-if="!node.disabled">
          <el-button
            type="text"
            style="margin-left: 12px;"
            @click="() => handleEdit(data)"
          >
            <i class="el-icon-edit" />
          </el-button>
          <el-button
            type="text"
            style="transform: rotate(90deg)"
          >
            <i class="tn-icon-view-module" />
          </el-button>
        </span> -->
