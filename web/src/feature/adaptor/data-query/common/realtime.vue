<template>
  <div class="realtime">
    <el-tabs
      v-model="activeRootTabKey"
      class="full-w el-tabs-full-w sign-alarm-switch"
    >
      <el-tab-pane
        :label="`信号(${totalItems})`"
        name="signs"
      >
        <div
          :class="{
            pad: isPad
          }"
          class="grid-tabs"
        >
          <el-button
            v-show="hideCollapseBtn"
            type="icon"
            :icon="`tn-icon-arrow-${treeCollapsed ? 'right' : 'left'}`"
            @click="toggleTreeVisible"
          />

          <el-tabs
            v-model="activeName"
            type="border-card"
            class="el-tabs-small el-tabs-full-w el-tabs-empty"
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
        </div>

        <div
          class="toolbar"
        >
          <el-input
            v-model="searchValue"
            border-type="shadow"
            placeholder="搜索测点名称"
            size="mini"
            class="no-padding"
            @change="getData"
          />

          <template>
            <div
              v-if="!tboxObsEnable"
              class="divider-h-sm"
            />

            <el-button
              type="text"
              :disabled="!multipleSelection.length"
              @click="jumpHistory"
            >
              查看历史
            </el-button>

            <div
              v-if="!tboxObsEnable"
              class="divider-h-sm"
            />

            <el-button
              v-show="!hideForceBtn && !tboxObsEnable"
              type="text"
              :disabled="!multipleSelection.length"
              @click="addInterestIndicatorList"
            >
              添加到关注
            </el-button>

            <div
              v-if="!tboxObsEnable"
              class="divider-h-sm"
            />

            <!-- <el-button type="text">
              导出
            </el-button> -->
            <el-button
              v-if="!tboxObsEnable"
              type="text"
              @click="exportList"
            >
              导出全部
            </el-button>
          </template>
        </div>

        <el-table-toolbar
          v-if="false"
          v-model="searchValue"
          :filter-placeholder="tboxObsEnable ? '搜索测点名称' : '搜索测点名称'"
          :filter-width="226"
          class="grid-toolbar"
          @search="getData"
        />
        <div
          v-if="!tboxObsEnable && false"
          class="grid-extra-button"
          style="border-bottom: solid 1px #f0f0f0;"
        >
          <el-button
            type="text"
            :disabled="!multipleSelection.length"
            @click="jumpHistory"
          >
            查看历史
          </el-button>
          <el-button
            v-show="!hideForceBtn"
            type="text"
            :disabled="!multipleSelection.length"
            @click="addInterestIndicatorList"
          >
            添加到关注
          </el-button>
          <!-- <el-button type="text">
            导出
          </el-button> -->
          <el-button
            type="text"
            @click="exportList"
          >
            导出全部
          </el-button>
        </div>
        <div class="grid-table">
          <el-table
            ref="table"
            v-loading="rtLoading"
            :data="tableData"
            :height="tableHeight"
            row-key="id"
            class="points-table"
            stripe
            style="width: 100%;"
            @selection-change="handleSelectionChange"
            @filter-change="filterChange"
          >
            <el-table-column
              type="selection"
              :width="isPad ? 46 : 70"
              reserve-selection
            />
            <el-table-column
              prop="deviceNumber"
              label="设备编号"
              :min-width="isPad ? 120 : 310"
              :formatter="formatDeviceNumber"
            />
            <el-table-column
              prop="deviceName"
              label="设备名称"
              :min-width="isPad ? 120 : 310"
            />
            <el-table-column
              v-if="enableDeviceNumberV2 !== '1' && !isPad"
              :key="columnKey"
              prop="deviceTypesName"
              label="设备类型"
              column-key="deviceTypesName"
              :width="isPad ? 110 : 150"
              :filter-multiple="true"
              :filters="activeTab === 'type' ? null : deviceTypeList"
              show-overflow-tooltip
            />
            <el-table-column
              v-if="enableDeviceNumberV2 === '1'"
              :key="columnKey"
              prop="applicationTypeZh"
              label="应用类型"
              column-key="applicationTypeZh"
              :width="isPad ? 110 : 150"
              :filter-multiple="true"
              :filters="activeTab === 'type' ? null : deviceTypeList"
              show-overflow-tooltip
            />
            <el-table-column
              v-if="enableDeviceNumberV2 === '1'"
              prop="categoryZh"
              label="设备种类"
              :width="isPad ? 110 : 150"
              column-key="categoryZh"
              show-overflow-tooltip
            />
            <el-table-column
              v-if="showAttr"
              prop="attrId"
              label="测点标识符"
              :fixed="isPad ? false : 'right'"
              :min-width="isPad ? 100 : 160"
              show-overflow-tooltip
            />
            <el-table-column
              prop="attrName"
              label="测点名称"
              :fixed="isPad ? false : 'right'"
              :min-width="isPad ? 100 : 200"
              :show-overflow-tooltip="!isPad"
            />

            <el-table-column
              label="当前值"
              :fixed="isPad ? false : 'right'"
              :width="isPad ? 120 : 190"
              show-overflow-tooltip
            >
              <template v-slot="{ row }">
                <el-tooltip
                  v-if="row.q !== undefined && row.q !== '0' && row.q !== 0"
                  :content="getPointErrorDesc(row.q)"
                >
                  <i class="el-icon-warning color-danger" />
                </el-tooltip>

                <el-tag
                  :type="warningAttrList.includes(row.attrId) ? 'danger' : 'success'"
                  style="font-weight: 800;"
                >
                  <span v-if="row.status"> {{ row.enumValue }} </span>
                  <span v-else> {{ row.value }} {{ row.unit }}</span>
                </el-tag>

                <admin-limit-tooltips
                  v-if="enableControlSetting && row.readAndWrite && !isPad"
                  container-comp="span"
                >
                  <template
                    slot-scope="{ hasRight }"
                  >
                    <span>
                      <el-button
                        :disabled="!hasRight"
                        type="text"
                        icon="tn-icon-edit"
                        @click="updateControlValue(row)"
                      >
                        设置
                      </el-button>
                    </span>
                  </template>
                </admin-limit-tooltips>
              </template>
            </el-table-column>
            <el-table-column
              prop="updateTime"
              label="刷新时间"
              :fixed="isPad ? false : 'right'"
              :width="isPad ? 160 : 190"
            />
            <el-table-column
              label="操作"
              :width="isPad ? 120 : 160"
              :fixed="isPad ? false : 'right'"
            >
              <template v-slot="{ row }">
                <el-button
                  v-show="!tboxObsEnable"
                  type="text"
                  @click="traceSource(row)"
                >
                  溯源
                </el-button>

                <el-button
                  v-show="!hideForceBtn && !tboxObsEnable"
                  type="text"
                  @click="addToFocus(row)"
                >
                  <!-- <i
                    class="tn-icon-star-hollow"
                  /> -->
                  关注
                </el-button>
                <el-button
                  type="text"
                  @click="checkSinglePoint(row)"
                >
                  <!-- <i
                    class="tn-icon-history"
                  /> -->
                  {{ tboxObsEnable ? '历史曲线' : '历史' }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <div>
            <el-pagination
              layout="total, prev, pager, next, sizes, jumper"
              styled
              background
              :pager-count="5"
              :total="totalItems"
              :current-page.sync="currentPage"
              :page-sizes="[10, 15, 20, 30, 40, 50, 100]"
              :page-size="pageSize"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane
        name="alarms"
        lazy
      >
        <template #label>
          告警 (<span :class="alarmList && alarmList.length ? 'error-text' : ''">{{
            alarmList ? alarmList.length : '...'
          }}</span>)
        </template>

        <el-tabs
          v-model="alarmActiveKey"
          class="el-tabs-small el-tabs-full-w border-bottom"
        >
          <el-tab-pane
            label="当前告警"
            name="active"
          />
          <el-tab-pane
            label="历史告警"
            name="history"
          />
        </el-tabs>

        <alarm-list
          :node-data="getSelNodeData()"
          :mode="alarmActiveKey"
          :alarms="alarmActiveKey === 'history' ? historyAlarmList : alarmList"
          :height="tableHeight"
          :fetch-history-alarms="fetchHistoryAlarms"
          :history-total="historyAlarmCount"
        />
      </el-tab-pane>
    </el-tabs>

    <history-point-modal
      v-if="modalVisible"
      :visible.sync="modalVisible"
      :point-list-prop="pointlist"
      :is-pad="isPad"
    />

    <update-control-value-modal
      :row.sync="updatingControlValueRow"
      @saved="handleUpdateControlValueRowSaved"
    />

    <trace-source-tbos-modal v-if="isTbos"  :point.sync="pointTracingSource">
    </trace-source-tbos-modal>

    <trace-source-modal
      v-else
      :point.sync="pointTracingSource"
    />
  </div>
</template>

<script>
import { dataQuery as cgi } from '@@/config/cgi';
import business from '@@/config/business';
import getEdgeRequest from '../../../utils/request';
import eventBus from '../eventBus';
import { uniq, cloneDeep, isEmpty, isArray } from 'lodash';
import historyPointModal from './history-point-modal.vue';
import { bindVueResize } from '@/utils/dom-resize';
import UpdateControlValueModal from './update-control-value-modal.vue';
import AdminLimitTooltips from 'feature/component/tedge-components/admin-limit-tooltips.vue';
import dayjs from 'dayjs';
import { CheckpointsByConditionsWatcher } from 'services/tedge/data-watchers/checkpoints-by-condition.ts';
// import { TboxModeCheckpointstWatcher } from 'services/tedge/data-watchers/tbox-mode-checkpoints.ts';
import { AlarmsWatcher, HistoryAlarmsWatcher } from 'services/tedge/data-watchers/alarms';
import AlarmList from './alarm-list.vue';
import TraceSourceModal from './trace-source-modal.vue';
import TraceSourceTbosModal from './trace-source-tbos-modal.vue';

const MAX_ITEMS_COUNT = 10;

export default {
  inject: ['getSelNodeData'],
  components: {
    historyPointModal,
    UpdateControlValueModal,
    AdminLimitTooltips,
    AlarmList,
    TraceSourceModal,
    TraceSourceTbosModal,
  },
  props: {
    mozuloaded: Boolean,
    mozuId: Number,
    activeTab: String,
    nodePathArray: {
      type: Array,
      default: () => [],
    },
    // warningAttrList: {
    //   type: Array,
    //   default: () => [],
    // },
    treeOption: {
      type: String,
      default: '',
    },
    hideCollapseBtn: {
      type: Boolean,
      default() {
        return false;
      },
    },
    hideForceBtn: {
      type: Boolean,
      default() {
        return false;
      },
    },
    enableControlSetting: {
      type: Boolean,
      default() {
        return false;
      },
    },
    isPad: {
      type: Boolean,
      default() {
        return false;
      },
    },
    enableDeviceNumberV2: {
      type: String,
      default() {
        return '0';
      },
    },
    devicesMap: {
      type: Object,
      default() {
        return {};
      },
    },
    treeType: {
      type: String,
      default: 'position',
    },
  },
  data() {
    return {
      notCascade: false,
      modalVisible: false,
      maxItemsCount: MAX_ITEMS_COUNT,
      activeName: 'all',
      loading: false,
      filter: {
        roomCode: [],
        deviceTypesName: [],
        deviceNumber: '',
        attrName: '',
        attrId: '',
      },
      roomList: [],
      deviceTypeList: [],
      tableData: [],
      tableHeight: 600,
      timer: null,
      rtLoading: true,
      selection: [],
      searchValue: '',
      currentPage: 1,
      totalItems: 0,
      pageSize: 15,

      multipleSelection: [],
      treeCollapsed: false,
      globalParams: {},
      nodePathArrayProp: [],
      showAttr: true,
      pointlist: [],
      columnKey: Math.random(),

      updatingControlValueRow: null,

      tboxObsEnable: false,

      checkpointsByConditionsWatcher: new CheckpointsByConditionsWatcher(3000),
      alarmsWatcher: new AlarmsWatcher(3000).withDiffPlugin(),
      historyAlarmsWatcher: new HistoryAlarmsWatcher(3000),

      warningAttrList: [],

      activeRootTabKey: 'signs',

      alarmList: null,
      historyAlarmList: null,
      historyAlarmCount: 0,
      alarmActiveKey: 'active',

      pointCodeDir: {},

      pointTracingSource: null,
      isTbos: window.tnwebServices.isTbos,
      
      lastTableDataCondition: null,
    };
  },
  watch: {
    activeName() {
      this.getData();
    },
    mozuId() {
      this.getData();
    },
    treeOption() {
      // if (val === 'alarm') {
      //   clearTimeout(this.timer);
      //   this.timer = null;
      // }
    },
    searchValue() {
      this.currentPage = 1;
    },

    // multipleSelection(val) {
    //   if (val.length) {
    //     clearTimeout(this.timer);
    //     this.timer = null;
    //   } else {
    //     this.getData();
    //   }
    // },
  },
  mounted() {
    bindVueResize(this, () => {
      this.tableHeight = window.innerHeight - (this.isPad ? 180 : 268);
    });

    eventBus.$on('notCascade', (val) => {
      this.notCascade = val;
    });

    this.$nextTick(() => {
      // this.calcTableHeight();
    });
    window.setShowAttr = () => {
      localStorage.setItem('showAttr', 'true');
    };
    this.showAttr = localStorage.getItem('showAttr') !== 'false';

    this.loadErrorCodeDir();
  },
  beforeDestroy() {
    clearTimeout(this.timer);
    this.timer = null;

    eventBus.$off('notCascade');

    this.checkpointsByConditionsWatcher.cancel();
    this.alarmsWatcher.cancel();
    this.historyAlarmsWatcher.cancel();
  },
  methods: {
    formatDeviceNumber(row, col, deviceNumber) {
      return window.tnwebServices.v2DeviceNumberTransformerService.get(deviceNumber, true);
    },
    getPointErrorDesc(q) {
      const pointCodeDescMsg = this.pointCodeDir[q]?.desc || '';
      return `【故障码：${q}】${pointCodeDescMsg}`;
    },

    async loadErrorCodeDir() {
      if (this.isPad) return;
      this.pointCodeDir = _.mapKeys(await window.tnwebServices.customConfigService.loadConfig('point_code_dir'), 'code');
    },
    traceSource(point) {
      this.pointTracingSource = point;
    },

    maunClearTimeout() {
      clearTimeout(this.timer);
      this.timer = null;
    },
    addInterestIndicatorList() {
      this.$emit('mutifocused', this.multipleSelection);
    },
    addToFocus(row) {
      this.$emit('focused', row);
    },
    filterChange(v) {
      this.filter.deviceTypesName = v.deviceTypesName || v.applicationTypeZh;
    },
    jumpHistory() {
      this.addMultiplePoint();
      this.checkSelectedPoints();
    },
    handleSelectionChange(val) {
      this.multipleSelection = val;

      // if (this.multipleSelection.length >= 10) {
      //   this.multipleSelection = val.slice(0, 10);

      //   const others = val.slice(10);

      //   if (others.length) {
      //     others.forEach((e) => {
      //       this.$refs.table.toggleRowSelection(e, false);
      //     });
      //   }
      // }
    },

    /**
     * 批量添加测点
     */
    addMultiplePoint() {
      for (let i = 0; i < this.multipleSelection.length; i++) {
        const e = this.multipleSelection[i];
        if (this.selection.length < 50) {
          if (!this.selection.find(item => item.id === e.id)) {
            this.selection.push(e);
          }
        } else {
          this.$message('最多可选择50个测点');
          return;
        }
      }

      this.$refs.table.clearSelection();
    },

    toggleTreeVisible() {
      this.$emit('tree-visible-change', this.treeCollapsed);
      this.treeCollapsed = !this.treeCollapsed;
    },

    /**
     * 拉取高级筛选的房间候选项
     */
    getRoomList() {
      getEdgeRequest(this.$axios, this.mozuId)
        .post(cgi.getDistinctByFieldName, { fieldName: 'roomCode' }, false)
        .then((data) => {
          this.roomList = data.map(e => ({
            value: e,
            label: e,
          }));
        });
    },

    /**
     * 拉取高级筛选房间对应的设备类型
     */
    getDeviceTypesName() {
      this.globalParams.conditions.shift();
      const { filter: { deviceTypesName } } = this;
      if(this.globalParams?.conditions.length === 0 || ((this.globalParams?.conditions || []).filter(i=> isEmpty(i.value)).length) || this.treeType === 'type'){
        return
      }
      getEdgeRequest(this.$axios, this.mozuId)
        .post('/cgi/dataQuery/edge/getDistinctByFieldNameByConditions', {
          ...this.globalParams,
          fieldName: this.enableDeviceNumberV2 === '1'
            ? 'applicationTypeZh'
            : 'deviceTypesName',
        }, false)
        .then((data) => {
          const deviceTypeList = data.sort(a => deviceTypesName.includes(a) ? -1 : 0).map(e => ({
            value: e,
            text: e,
          }));
          this.deviceTypeList = deviceTypeList;
          this.$emit('getDeviceTypeList', this.deviceTypeList);
        });
    },

    /**
     * 重置高级筛选
     */
    reset() {
      this.$set(this, 'filter', {
        roomCode: [],
        deviceTypesName: [],
        deviceNumber: '',
        attrName: '',
      });
      this.$refs.filter.resetFields();
    },

    createConditions() {
      const conditions = [];

      Object.keys(this.filter).forEach((key) => {
        const val = _.cloneDeep(this.filter[key]);

        if (Array.isArray(val)) {
          if (val.length) {
            conditions.push({
              name: key,
              value: val,
            });
          }
        } else {
          if (val !== '') {
            conditions.push({
              name: key,
              value: [val],
            });
          }
        }
      });

      if (this.activeName === 'status') {
        conditions.push({ name: 'status', value: ['true'] });
      }
      if (this.activeName === 'simulation') {
        conditions.push({ name: 'status', value: ['false'] });
      }
      if (this.activeName === 'control') {
        conditions.push({ name: 'rw', value: ['true'] });
      }

      return conditions;
    },

    getData() {
      if (business.showModuleSelected && !this.mozuloaded) {
        return;
      }
      // const cgiUrl = cgi.getCurrentBizGidAttrsWithValueByConditions;
      const selNode = this.getSelNodeData();
      clearTimeout(this.timer);

      if (selNode) {
        this.selNode = selNode;
        const keyword = this.searchValue;

        let gidValueList = [selNode.id];

        const createConditions = this.createConditions();
        if (this.activeTab === 'type' && selNode.level < 3 && this.nodePathArrayProp.length) {
          if (selNode.id.indexOf('room') > -1) {
            gidValueList = [];
          }
          const nodePathArrayReverse = cloneDeep(this.nodePathArrayProp).reverse();
          if (nodePathArrayReverse.find(i => i.indexOf('room') > -1)) {
            const room = nodePathArrayReverse.find(i => i.indexOf('room') > -1).split(':')[1];
            createConditions.push({ name: 'roomCode', value: [room] });
          }
          if (selNode.id.indexOf('room') > -1) {
            const selNodeRoom = selNode.id.split(':')[1];
            createConditions.push({ name: 'roomCode', value: [selNodeRoom] });
          }
          const deviceTypeNamesCondition = createConditions.find(i => i.name === 'deviceTypesName');
          if (deviceTypeNamesCondition) {
            nodePathArrayReverse.forEach((item) => {
              deviceTypeNamesCondition.value.push(item);
            });
          } else {
            createConditions.push({ name: 'deviceTypesName', value: [nodePathArrayReverse[0]] });
          }
          createConditions.find(i => i.name === 'deviceTypesName').value = uniq(createConditions.find(i => i.name === 'deviceTypesName').value);
        }

        const initCondition = [
          ...createConditions,
        ];
        if(this.treeType === 'type'){
          if(selNode.level >= 3) {
            initCondition.push({ name: 'deviceGid', value: gidValueList });
          } 
          if(selNode.level == 2 && (selNode.deviceTypeName.indexOf('房间') > -1 || selNode.deviceTypeName.indexOf('区') > -1)) {
            initCondition.push({ name: 'idc_area', value: gidValueList });
          }
        } else {
          initCondition.push({ name: 'deviceGid', value: gidValueList });
        }
        // 确认每个项的value数组里不包含空值
        initCondition.forEach(i=>{
          if(isArray(i.value)){
            i.value = i.value.filter(j=>j);
          }
        })

        const params = {
          conditions: initCondition,
          start: (this.currentPage - 1) * this.pageSize,
          limit: this.pageSize,
          keyword,
          operator: 'like',
          notCascade: this.notCascade,
        };

        const enableDeviceNumberV2 = this.enableDeviceNumberV2 === '1';
        if (enableDeviceNumberV2 && params.conditions) {
          params.conditions.forEach((item) => {
            if (item.name !== 'deviceTypesName') return;

            // eslint-disable-next-line no-param-reassign
            item.name = 'applicationTypeZh';
          });
        }

        const deviceTypeNamesCondition = createConditions.find(i => (i.name === 'deviceTypesName' || i.name === 'applicationTypeZh'));

        if (this.activeTab === 'type') {
          params.type = 'true';
          deviceTypeNamesCondition.value = deviceTypeNamesCondition.value.filter(item => item !== '房间');
        }

        this.lastTableDataCondition = cloneDeep(initCondition);
        
        this.globalParams = params;

        this.checkpointsByConditionsWatcher.mockRequest(params)
          .then((data) => {
            this.totalItems = data.count;
            this.tableData = data.list;
            this.rtLoading = false;
            this.timer = setTimeout(() => {
              this.getData();
            }, 3000);
          });

        if (this.getSelNodeData()?.name) {
          this.alarmsWatcher.mockRequest({
            eventStatus: -1,
            DeviceNumber: [
              this.getSelNodeData()?.name,
            ],
          }).then((list) => {
            this.alarmList = Object.freeze(list);
            this.warningAttrList = _.chain(list)
              .map('occurPointList')
              .flatten()
              .map('enName')
              .value();
          });
        }
      }
    },

    async fetchHistoryAlarms(pagination) {
      const nodeData = this.getSelNodeData();

      if (!nodeData) {
        this.historyAlarmList = [];
        this.historyAlarmCount = 0;
        return;
      }

      this.historyAlarmsWatcher.mockRequest({
        deviceGids: [
          nodeData.id,
        ],
        mozuId: window.Vue.prototype.$moduleInfo.mozuId,
        limit: pagination.size,
        offset: pagination.size * (pagination.current - 1),
      }).then((data) => {
        this.historyAlarmList = data.list;
        this.historyAlarmCount = data.count;
      });
    },

    /**
     * 改条件重新高级筛选后，将当前页改为1
     */
    searchDataByFilter() {
      this.$refs.table.clearSelection();
      this.currentPage = 1;
      this.notCascade = false;
      this.getData();
    },

    calcTableHeight() {
      this.tableHeight = this.$refs.tags.$el.offsetHeight - 56;
    },

    visibleChangeHandler(val) {
      this.$emit('collapse-change', val);
    },

    exportList() {
      const { selNode } = this;
      if (selNode) {
        const deviceGid = this.activeTab === 'type'
          ? selNode.id.replace(/^[^-]+-/, '')
          : selNode.id;

        const params = {
          conditions: [
            { name: 'deviceGid', value: [deviceGid] },
            ...this.createConditions(),
          ],
          start: (this.currentPage - 1) * this.pageSize,
          limit: this.totalItems,
          keyword: this.searchValue,
          operator: 'like',
          notCascade: this.notCascade,
        };

        // 先打补丁fix bug，后续重构为更合理的实现方式
        if (this.treeType === 'type') {
          if (!this.lastTableDataCondition.length) {
            console.warn('lastTableDataCondition 为空，可能数据还未加载完成');
            this.$message.warning('数据加载中，请稍后再试');
            return;
          }
          params.conditions = this.lastTableDataCondition;
        }

        const enableDeviceNumberV2 = this.enableDeviceNumberV2 === '1';
        if (enableDeviceNumberV2) {
          params.conditions.forEach((item) => {
            if (item.name !== 'deviceTypesName') return;

            item.name = 'applicationTypeZh';
          });
        }

        getEdgeRequest(this.$axios, this.mozuId)
          .download(cgi.exportCurrentBizGidAttrsWithValueByConditions, params);
      }
    },

    /**
     * 切换不同节点后，重新拉取
     */
    refresh() {
      this.rtLoading = true;
      this.$refs.table.clearSelection();
      this.currentPage = 1;
      const selNode = this.getSelNodeData();
      if (!selNode) return;

      this.alarmList = null;
      this.warningAttrList = [];

      const { deviceTypeName } = selNode;
      this.filter.deviceTypesName = [selNode.applicationTypeZh || selNode.deviceTypeName];
      this.deviceTypeList = [];
      this.columnKey = Math.random();
      this.getData();
      this.getDeviceTypesName();
    },

    /**
     * 增加测点到已选测点列表中
     * @param {Object} row - 选择的列
     */
    addPoint(row) {
      if (!this.selection.find(e => e.id === row.id)) {
        if (this.selection.length === 50) {
          this.$message('最多可选择50个测点');
          return false;
        }

        this.selection.push(row);
      }
    },

    /**
     * 删除测点
     */
    handleClose(tag) {
      this.selection.splice(this.selection.findIndex(e => e.id === tag.id), 1);
    },

    /**
     * 清空已选测点列表
     */
    clearSelection() {
      this.selection = [];
    },

    /**
     * 查看单个测定的历史数据
     */
    checkSinglePoint(row) {
      this.pointlist = this.isPad ? row.id : row.templatePointId;
      this.modalVisible = true;
      // this.goToAdvancedSearch([row]);
    },

    /**
     * 查看已选测点列表的历史数据
     */
    checkSelectedPoints() {
      this.goToAdvancedSearch(this.selection);
    },

    /**
     * 跳转至历史数据高级查询页面
     */
    goToAdvancedSearch(list) {
      // const { moduleName } = business;
      const result = list.map(e => this.isPad ? e.id : e.templatePointId).join(',');
      this.pointlist = result;
      this.modalVisible = true;
      this.selection = [];

      // const href = `/${moduleName}/advanced-search?pointlist=${result}`;
      // window.open(href);
    },

    updateControlValue(row) {
      this.updatingControlValueRow = row;
    },

    /**
     * 分页 pageSize处理
     */
    handleSizeChange(val) {
      this.$refs.table.clearSelection();
      this.pageSize = val;
      this.currentPage = 1;
      this.getData();
    },

    /**
     * 分页 当前页处理
     */
    handleCurrentChange(val) {
      this.$refs.table.clearSelection();
      this.currentPage = val;
      this.getData();
    },

    handleUpdateControlValueRowSaved() {
      this.getData();
    },
  },
};
</script>

<style lang="scss" scoped>
.grid-tabs.pad {
  /deep/ {
    .el-tabs__item {
      min-width: 80px;
    }
  }
}

.selection-enter-active, .selection-leave-active {
  transition: all .3s;
}
.selection-enter, .selection-leave-to {
  opacity: 0;
  width: 0;
}
.table-link {
  color: #1470cc;
}

/deep/ .el-table__body-wrapper {
  // height: calc(100vh - 290px);
  overflow: hidden;
  &:hover {
    overflow: auto;
  }
}

 
    // 禁用固定列的背景色过渡动画
  /deep/ .el-table--enable-row-transition .el-table__body td {
    transition: none;
  }
    

// /deep/ .el-table-toolbar__extra {
//   flex: 1;
// }

// .extra {
//   display: flex;
//   align-items: center;
//   padding-right: 24px;

//   &-status {
//     font-size: 18px;
//     margin-right: 16px;
//   }

//   &-export {
//     margin-left: auto;
//   }
// }

.filter {
  &-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));

    .el-select {
      /deep/ input {
        height: 26px !important;
      }
    }
  }

  &-footer {
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: flex-end;

    &-reset {
      color: #333;
    }
  }
}

.grid {
  height: 100%;
  display: grid;
  grid-template-areas:
    'tabs extra toolbar'
    // 'filter filter'
    // 'toolbar toolbar'
    'table table table';
  grid-template-columns: minmax(0, 1fr) 260px;
  grid-template-rows: auto auto auto 1fr;

  &-tabs {
    grid-area: tabs;
    display: flex;

    .el-tabs {
      flex: 1;
    }
  }

  &-filter {
    grid-area: filter;
  }

  &-table {
    grid-area: table;

  }

  &-extra-button {
    grid-area: extra;
    display:flex;
    justify-content: right;
    padding-right : 16px;
  }

  &-toolbar {
    grid-area: toolbar;
    position: relative;
    height: 56px;
    padding:0;
    /deep/ .el-table-toolbar__search {
      padding: 0 0 0 16px !important;
    }
    /deep/ .el-table-toolbar__extra {
      padding: 0px !important;
    }

    &:after {
      content: '';
      position: absolute;
      left: 0;
      right: 0;
      bottom: -1px;
      height: 0;
      border-bottom: 1px solid #f0f0f0;
    }
  }

  &-tags {
    border-left: 1px solid #f0f0f0;

    // display: grid;
    // grid-template-rows: auto minmax(0, 1fr);

    /deep/ .el-block__body-inner {
      display: flex;
      flex-direction: column;
      height: 100%;
      box-sizing: border-box;
      width: 260px;
      padding-bottom: 16px;
    }

    &-body {
      height: 480px;
      overflow-y: auto;
      box-sizing: border-box;
      padding: 0 24px;
    }

    .el-tag {
      margin-bottom: 8px;
      height: auto;
      white-space: unset;
      width: 100%;
      padding-right: 18px;
      position: relative;

      /deep/ &__close {
        position: absolute;
        right: 4px;
        top: 50%;
        transform: translateY(-50%);
      }
    }

    &-footer {
      margin-top: auto;
      align-self: center;
    }

    &-empty-text {
      color: #999;
      text-align: center;
      margin-top: 16px;
    }
  }
}
/deep/ tr {
      .is-hidden {
        display: table-cell;
        overflow: hidden;
        .cell {
          visibility: visible;
        }
      }
    }

.full-w {
  width: 100%;
}

.el-tabs-small /deep/ {
  .el-tabs__item {
    min-width: auto;
    padding: 0 16px;
    line-height: 42px;
    height: 42px;
  }
}

.el-tabs-full-w /deep/ {
  .el-tabs__content, .el-tab-pane {
    width: 100%;
  }
}

.realtime {
  overflow: hidden;
}

.toolbar {
  position: absolute;
  top: 6px;
  right: 8px;

  display: flex;
}

.divider-h-sm::before {
  content: '|';
  margin: 0 8px;
  color: #a0a0a0;

  position: relative;
  top: 8px;
}

.error-text {
  color: var(--tn-color-danger);
  font-weight: 600;
}

.border-bottom {
  border-bottom: 1px solid #e0e0e0;
}

.color-danger {
  color: var(--tn-color-danger);
}

.el-tabs-empty {
  /deep/ .el-tabs__content {
    display: none;
  }
}

.points-table {
  width: calc(100% - 16px);
  overflow:overlay;
  margin: 8px;
  // border: 1px solid #e0e0e0;
}

.no-padding {
  padding: 0;
}
</style>
