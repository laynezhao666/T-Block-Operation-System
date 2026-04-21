<template>
  <div class="tree-view">
    <header class="tree-view-header">
      <el-select
        v-model="filters.status"
        border-type="bordered"
        size="small"
        class="status-filter"
      >
        <el-option
          v-for="(opt, i) in statusFilterOptions"
          :key="i"
          :value="opt.value"
          :label="opt.label"
        />
      </el-select>

      <el-input
        v-model="filters.keywords"
        placeholder="请输入关键词"
        border-type="bordered"
        size="small"
        class="keywords-input"
      />

      <admin-limit-tooltips>
        <div
          slot-scope="{ hasRight }"
          class="lineHeight32"
        >
          <el-button
            :disabled="!hasRight"
            type="text"
            class="add-btn"
            size="small"
            @click="handleAddClicked"
          >
            +添加
          </el-button>
        </div>
      </admin-limit-tooltips>

      <el-dropdown
        @command="handleMoreCommand"
      >
        <el-space :size="4">
          <i class="tn-icon-more more-btn" />
        </el-space>
        <el-dropdown-menu slot="dropdown">
          <el-dropdown-item
            v-for="(command, i) in mode === 'control' ? commandList : commandList.slice(-2)"
            :key="i"
            :command="command.label"
          >
            {{ command.label }}
          </el-dropdown-item>
        </el-dropdown-menu>
      </el-dropdown>
    </header>

    <main class="tree-view-main">
      <el-tree
        ref="tree"
        :data="treeData"
        :props="treeProps"
        :default-expanded-keys="defaultExpandedKeys"
        :current-node-key="currentNodeKey"
        :filter-node-method="filterTreeNode"
        :expand-on-click-node="false"
        check-on-click-node
        highlight-current
        node-key="nodeKey"
        @node-click="handleTreeNodeClick"
      >
        <template
          #default="{ data }"
        >
          <div class="node">
            <div class="node-info">
              <span v-if="data.type === 'control'">
                【
                <span
                  :class="{
                    'control-status': true,
                    online: controlStatusMap[data.id] !== '1',
                    offline: controlStatusMap[data.id] === '1',
                  }"
                >
                  {{ controlStatusMap[data.id] === '1' ? '离线' : '在线' }}
                </span>
                】
              </span>

              <span v-if="data.type === 'control'">
                {{ data.name }} | {{ data.channel && data.channel.chid }}
              </span>

              <div v-if="data.type === 'group'">
                {{ data.name }} | {{ getGroupdDoorsCount(data) }}门
              </div>

              <div v-if="data.type === 'door'">
                {{ data.name }}
              </div>

              <div v-if="data.type === 'root'">
                全部
              </div>
            </div>

            <div class="oprs">
              <admin-limit-tooltips
                v-if="mapOfTreeNodeOprsCode[data.type] & nodeOprCodeMap.edit"
              >
                <span slot-scope="{ hasRight }">
                  <el-button
                    :disabled="!hasRight"
                    type="text"
                    icon="tn-icon-edit"
                    size="small"
                    @click.stop="editNode(data)"
                  />
                </span>
              </admin-limit-tooltips>

              <span
                v-if="mapOfTreeNodeOprsCode[data.type] & nodeOprCodeMap.remove"
                @click.stop
              >
                <admin-limit-tooltips>
                  <span slot-scope="{ hasRight }">
                    <el-popconfirm
                      title="确定是否删除？"
                      @onConfirm="removeNode(data)"
                    >
                      <el-button
                        slot="reference"
                        :disabled="!hasRight"
                        type="text"
                        icon="tn-icon-circle-remove"
                        size="small"
                      />
                    </el-popconfirm>
                  </span>
                </admin-limit-tooltips>
              </span>

              <!-- 暂时无用 -->
              <el-button
                v-if="mapOfTreeNodeOprsCode[data.type] & nodeOprCodeMap.add"
                type="text"
                icon="tn-icon-circle-add"
                size="small"
                @click.stop="addNode(data)"
              />

              <span @click.stop>
                <el-dropdown
                  v-if="mapOfTreeNodeOprsCode[data.type] & nodeOprCodeMap.more"
                  @command="handleNodeCommand"
                >
                  <el-button
                    type="text"
                    icon="tn-icon-list"
                    size="small"
                  />
                  <el-dropdown-menu slot="dropdown">
                    <el-dropdown-item
                      v-for="(command, key) in nodeCommandsMap[data.nodeKey]"
                      :key="key"
                      :command="`${data.nodeKey}-${key}`"
                    >
                      {{ command.label }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </el-dropdown>
              </span>
            </div>
          </div>
        </template>
      </el-tree>
    </main>

    <group-form-modal
      ref="groupFormModal"
      @reloadTree="loadData"
    />

    <control-form-modal
      ref="controlFormModal"
      @reloadTree="loadData"
    />

    <sync-time-groups
      ref="syncTimeGroups"
    />

    <remote-control-modal
      ref="remoteControlModal"
    />

    <doors-relations-import-modal
      ref="doorsRelationsImportModal"
    />
  </div>
</template>

<script>
import _ from 'lodash';
import GroupFormModal from './group-form-modal.vue';
import ControlFormModal from './control-form-modal.vue';
import RemoteControlModal from './remote-control-dialog.vue';
import DoorsRelationsImportModal from './doors-relations-import-modal.vue';
import SyncTimeGroups from '../../security-time-period-setting/time-period/sync-time-groups.vue';
import { loadControlAndGroupsTree } from '../utils/fetch-group-control-door-trees';
import { downloadByUrl } from '../../../../utils/download';
import { axiosDelete, axiosUploadFile } from '../../../../utils/axios-methods';
import { chunkSeriesPromise } from '../../../../utils/promise-utils';
import { forEachTreeNode } from 'utils/tree';
import AdminLimitTooltips from 'feature/component/tedge-components/admin-limit-tooltips.vue';
import { DcosRtdWatcher } from 'services/tedge/data-watchers/dcos-rtd.ts';
import getEdgeRequest from 'feature/utils/request';

const mapOfGetDoors = {
  root: data => _.chain(data.doors)
    .map('doors')
    .flatten()
    .value(),
  door: data => [data],
  control: data => data.doors,
  group: data => data.doors,
};

const nodeOprCodeMap = {
  edit: 0b1,
  // add: 0b10,
  remove: 0b100,
  more: 0b1000,
};

/** 按位运算：计算 */
const mapOfTreeNodeOprsCode = {
  root: 0,
  door: nodeOprCodeMap.edit | nodeOprCodeMap.more,
  control: nodeOprCodeMap.edit | nodeOprCodeMap.add | nodeOprCodeMap.remove | nodeOprCodeMap.more,
  group: nodeOprCodeMap.edit | nodeOprCodeMap.add | nodeOprCodeMap.remove | nodeOprCodeMap.more,
};

export default {
  components: {
    GroupFormModal,
    ControlFormModal,
    SyncTimeGroups,
    RemoteControlModal,
    DoorsRelationsImportModal,
    AdminLimitTooltips,
  },
  props: {
    mode: {
      type: String,
      default() {
        return 'control';
      },
    },
  },
  data() {
    return {
      controlsTree: [],
      groupsTree: [],
      controlStatusMap: {},
      nodeCommandsMap: {},

      currentNodeKey: 'root',
      defaultExpandedKeys: ['root'],

      nodeOprCodeMap,
      mapOfTreeNodeOprsCode,

      filters: {
        status: 'all',
        keywords: '',
      },

      treeProps: {
        children: 'doors',
        label: 'name',
      },

      statusFilterOptions: [{
        value: 'all',
        label: '全部',
      }, {
        value: 'online',
        label: '在线',
      }, {
        value: 'offline',
        label: '离线',
      }],

      commandList: [{
        label: '清空时间组',
        confirm: {
          title: '请确认是否执行清空时间组吗？',
        },
        action: this.clearTimeGroups.bind(this),
      }, {
        label: '同步时钟',
        confirm: {
          title: '请确认是否执行同步时钟吗？',
        },
        action: this.syncTimeClock.bind(this),
      }, {
        label: '消防复位',
        confirm: {
          title: '请确认是否执行消防复位吗？',
        },
        action: this.resetFire.bind(this),
      }, {
        label: '导入门映射',
        action: this.batchImportDoorRelations.bind(this),
      }, {
        label: '导入控制器',
        action: this.batchImport.bind(this),
      }, {
        label: '导出控制器',
        action: this.batchExport.bind(this),
      }],

      rtdWatcher: new DcosRtdWatcher(3000)
        .withDiffPlugin()
        .bindVueVm(this),
    };
  },
  computed: {
    treeData() {
      return this.mode === 'control' ? this.controlsTree : this.groupsTree;
    },
  },
  watch: {
    filters: {
      deep: true,
      handler() {
        this.$refs.tree.filter(this.filters);

        // 等待树组件渲染
        setTimeout(() => {
          this.handleTreeNodeClick(this.treeData[0]);
        }, 20);
      },
    },
    mode() {
      this.filters.keywords = '';
      this.filters.status = 'all';

      // 等待树组件渲染
      setTimeout(() => {
        this.handleTreeNodeClick(this.treeData[0]);
      }, 20);
    },
  },
  created() {
    this.loadControlStatus = this.$intervalFunction(this.loadControlStatus, 1000, true);
    this.loadData();
  },
  methods: {
    async loadData() {
      await this.loadTrees();
      this.loadControlStatus();
    },
    async loadTrees() {
      const {
        controlsTree,
        groupsTree,
      } = await loadControlAndGroupsTree();

      this.controlsTree = [{
        type: 'root',
        nodeKey: 'root',
        name: '全部',
        doors: controlsTree,
      }];
      this.groupsTree = [{
        type: 'root',
        nodeKey: 'root',
        name: '全部',
        doors: groupsTree,
      }];

      this.generateControlNodeCommandActions();
      this.generateGroupNodeCommandActions();
      this.generateDoorNodeCommandActions();

      setTimeout(() => {
        if (!this.$refs) return;
        this.$refs.tree.filter(this.filters);

        // 等待树组件渲染
        setTimeout(() => {
          this.handleTreeNodeClick(this.treeData[0]);
        }, 20);
      }, 20);
    },
    async loadControlStatus() {
      const controls = this.controlsTree?.[0]?.doors;
      if (!controls?.length) return;

      const statusPointIdToControlIdMap = _.chain(controls)
        .mapKeys('comm_id')
        .mapValues('id')
        .value();

      const ids = _.keys(statusPointIdToControlIdMap);
      const statusData = await this.rtdWatcher.mockRequest({
        ids,
      });

      this.controlStatusMap = _.chain(statusData)
        .map((checkPointValue) => {
          const { id: pointId, pv } = checkPointValue;
          return [statusPointIdToControlIdMap[pointId], pv];
        })
        .fromPairs()
        .forEach()
        .value();
    },

    generateControlNodeCommandActions() {
      _.forEach(this.controlsTree[0].doors, (control) => {
        const actions = {
          clearTimeGroups: {
            label: '清空时间组',
            confirm: {
              title: '请确认是否执行清空时间组吗？',
            },
            action: () => this.clearTimeGroups(control),
          },
          syncTimeClock: {
            label: '同步时钟',
            confirm: {
              title: '请确认是否执行同步时钟吗？',
            },
            action: () => this.syncTimeClock(control),
          },
          resetFire: {
            label: '消防复位',
            confirm: {
              title: '请确认是否执行消防复位吗？',
            },
            action: () => this.resetFire(control),
          },
          formatControl: {
            label: '格式化',
            confirm: {
              title: '请确认是否执行格式化吗？',
            },
            action: () => this.formatControl(control),
          },
        };

        this.$set(this.nodeCommandsMap, control.nodeKey, actions);
      });
    },

    generateGroupNodeCommandActions() {
      _.forEach(this.groupsTree[0].doors, (group) => {
        const actions = {
          remoteControl: {
            label: '远程控制',
            action: () => this.remoteControl(group),
          },
          setDoorParams: {
            label: '设置门参数',
            action: () => this.setDoorParams(group),
          },
          showRecords: {
            label: '刷卡记录',
            action: () => this.showRecords(group),
          },
        };

        this.$set(this.nodeCommandsMap, group.nodeKey, actions);
      });
    },

    generateDoorNodeCommandActions() {
      const doors = _.chain(this.controlsTree[0].doors)
        .map('doors')
        .flatten()
        .value();

      _.forEach(doors, (door) => {
        const actions = {
          remoteControl: {
            label: '远程控制',
            action: () => this.remoteControl(door),
          },
          setDoorParams: {
            label: '设置门参数',
            action: () => this.setDoorParams(door),
          },
          showRecords: {
            label: '刷卡记录',
            action: () => this.showRecords(door),
          },
        };

        this.$set(this.nodeCommandsMap, door.nodeKey, actions);
      });
    },

    filterTreeNode(filters, data, node) {
      if (this.mode === 'control' && (!filters || data.type === 'root' || (node.level > 2 && node.parent.visible))) return true;

      if (filters.keywords) {
        const isKeywordsUnmatch = !data.name?.includes(filters.keywords)
          && !data.code?.includes(filters.keywords);
        if (isKeywordsUnmatch) {
          return false;
        }
      }

      if (filters.status === 'all') return true;

      const controlId = data.type === 'control' ? data.id : data.controlId;

      if (!controlId) return true;

      const { controlStatusMap } = this;

      return (filters.status === 'online' && controlStatusMap[controlId] === '0')
        || (filters.status === 'offline' && controlStatusMap[controlId] === '1');
    },

    getGroupdDoorsCount(group) {
      const {
        controlStatusMap,
        filters: {
          keywords,
          status,
        },
      } = this;

      if (!keywords && status === 'all') return group.doors?.length || 0;

      const isGroupVisible = (!keywords || group.name?.includes(keywords))
        && (
          (status === 'all')
          || (status === 'online' && controlStatusMap[group.id] !== 'false')
          || (status === 'offline' && controlStatusMap[group.id] === 'false')
        );

      return isGroupVisible
        ? group.doors.length
        : _.chain(group.doors)
          .filter(door => !keywords || door.name?.includes(keywords))
          .filter((door) => {
            const { controlId } = door;
            return (status === 'all')
            || (status === 'online' && controlStatusMap[controlId] !== 'false')
            || (status === 'offline' && controlStatusMap[controlId] === 'false');
          })
          .size()
          .value();
    },

    editNode(data) {
      if (data.type === 'door') {
        this.$emit('setDoorParams', data);
      } else {
        this.$refs[`${data.type}FormModal`].edit(data, () => {
          this.loadData();
        });
      }
    },
    removeNode(data) {
      ({
        control: this.removeControl,
        group: this.removeGroup,
      })[data.type].call(this, data);
    },
    async removeControl(data) {
      await axiosDelete('/api/dcos/tdac-cgi/controller', {
        id: data.id,
      });
      this.$message.success('删除控制器成功');
      this.loadData();
    },
    async removeGroup(data) {
      await axiosDelete('/api/dcos/tdac-cgi/group', {
        id: data.id,
      });
      this.$message.success('删除分组成功');
      this.loadData();
    },

    // TODO: 实现或对接下列功能
    async syncTimeGroups(control) {
      this.$refs.syncTimeGroups.startSync(control ? [control.id] : []);
    },
    clearTimeGroups(control) {
      console.log('TODO: clearTimeGroups', control);
    },
    async syncTimeClock(control) {
      if (!control) {
        await getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/controllers/sync-time');
      } else {
        await getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/controller/sync-time', {
          id: control.id,
        });
      }

      this.$message.success('同步成功');
    },
    async resetFire(control) {
      const controlIds = [];

      if (control) {
        controlIds.push(control.id);
      } else {
        this.controlsTree[0].doors.forEach((item) => {
          controlIds.push(item.id);
        });
      }

      await chunkSeriesPromise(controlIds, 4, id => getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/controller/reset', {
        id,
      }));

      this.$message.success('消防复位完成');
    },
    async batchImportDoorRelations() {
      this.$refs.doorsRelationsImportModal.open();
    },
    async batchImport() {
      await axiosUploadFile('/api/dcos/tdac-cgi/controllers/import', {
        file: axiosUploadFile.fileSelectSymbol,
      });
      this.loadData();
      this.$message('导入完成');
    },
    batchExport() {
      downloadByUrl('/api/dcos/tdac-cgi/controllers/export', 'aaaa.xlsx');
    },
    async formatControl(control) {
      await getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/controller/clean', {
        id: control.id,
      });
      this.$message.success('格式化完成。');
    },
    setDoorParams(door) {
      this.$emit('setDoorParams', door);
    },
    remoteControl(data) {
      this.$refs.remoteControlModal.show(data.type === 'door' ? [data.id] : _.map(data.doors, 'id'));
    },
    showRecords(data) {
      const doors = data.type === 'group'
        ? data.doors
        : [data];
      this.$emit('showRecords', doors);
    },

    handleAddClicked() {
      const {
        mode,
      } = this;

      if (mode === 'group') {
        this.$refs.groupFormModal.edit({
          name: '',
          doors: [],
        }, () => {
          this.loadData();
        });
        return;
      }

      if (mode === 'control') {
        this.$refs.controlFormModal.edit({
          name: '',
          profile: {},
          position: {},
          channel: {},
          protocol: {},
        }, () => {
          this.loadData();
        });
      }
    },
    async handleMoreCommand(commandKey) {
      if (!window.tnwebServices.loginStatusService.hasRight(true)) {
        return;
      }

      const command = _.find(this.commandList, { label: commandKey });
      if (!command) return;

      if (command.confirm && !(await this.$confirm(command.confirm.title, '是否确认操作'))) return;

      command.action();
    },
    handleTreeNodeClick(data) {
      this.currentNodeKey = data.nodeKey;

      const doors = [];
      const doorsValidMap = _.mapKeys(mapOfGetDoors[data.type](data), 'id');

      const rootNode = this.$refs.tree.getNode('root');

      if (rootNode) {
        forEachTreeNode([rootNode], (node) => {
          if (!node.visible || node.data?.type !== 'door') return;

          if (doorsValidMap[node.data.id]) {
            doors.push(node.data);
          }
        }, {
          childrenField: 'childNodes',
        });
      }

      this.$emit('doorsChange', doors);
    },
    async handleNodeCommand(commandKey) {
      if (!window.tnwebServices.loginStatusService.hasRight(true)) {
        return;
      }

      const {
        nodeCommandsMap,
      } = this;

      const commandKeyParties = commandKey.split('-');
      const path = commandKeyParties.slice(0, -1).join('-');
      const actionName = commandKeyParties[commandKeyParties.length - 1];

      const command = nodeCommandsMap[path][actionName];

      if (command.confirm && !(await this.$confirm(command.confirm.title, '是否确认操作'))) return;

      command.action();
    },
  },
};
</script>

<style lang="scss" scoped>
.tree-view {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.tree-view-header {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  box-sizing: border-box;
  border-bottom: silver;
  border-bottom: 1px solid #ededed;
}

.tree-view-main {
  flex: 1;
  overflow: auto;
}

.status-filter {
  width: 120px;
}

.node {
  position: relative;
  width: 100%;

  &:hover {
    .oprs {
      visibility: visible;
    }
  }
}

.control-status {
  font-weight: 600;

  &.online {
    color: var(--tn-color-success);
  }
  &.offline {
    color: var(--tn-color-danger);
  }
}

.more-btn  {
  position: relative;
  top: 4px;
}

.oprs {
  visibility: hidden;

  position: absolute;
  right: 0;
  top: -10px;
  line-height: 36px;
  background-color: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(5px);
  padding-left: 4px;

  & > * {
    margin-left: 0;
  }
}

.lineHeight32 {
  line-height: 32px;
}
</style>
