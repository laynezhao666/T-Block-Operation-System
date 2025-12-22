<template>
  <custom-config-value
    class="data-query"
    name="EnableDeviceNumberV2"
  >
    <template #default="{ value: enableDeviceNumberV2 }">
      <el-block>
        <div
          class="data-query-container"
        >
          <transition name="tree">
            <el-tabs
              v-show="treeVisible"
              v-model="activeName"
              :style="{
                width: `${leftPanelWidth}px`
              }"
              class="data-query-container-tabs"
              @tab-click="tabClick"
            >
              <el-tab-pane
                label="按位置"
                name="position"
                :lazy="true"
              >
                <div v-if="activeName === 'position'">
                  <div style="display:flex;">
                    <div class="search-wrap">
                      <el-select
                        v-model="treeOption"
                        border-type="no-border"
                      >
                        <el-option
                          label="全部"
                          value="all"
                        >
                          全部
                        </el-option>
                        <el-option
                          label="告警"
                          value="alarm"
                        >
                          告警
                        </el-option>
                      </el-select>

                      <el-input
                        v-model="filterText"
                        class="c-tree-tools__input"
                        placeholder="搜索设备"
                        prefix-icon="tn-icon-search"
                        clearable
                        border-type="no-border"
                      />
                      <!-- <img
                        :src="seletedAllColor"
                        :style="{cursor:'pointer', color: '#1470CC'}"
                        @click="showSetModal"
                      > -->
                      <i
                        :style="{cursor:'pointer', color: seletedAllColor ? 'gray' : '#1470CC'}"
                        class="tn-icon-filter"
                        @click="showSetModal"
                      />
                    </div>
                  </div>
                  <tree-component
                    ref="treeContainerPosition"
                    data-source-type="byBiz"
                    class="data-query-tree"
                    :rooms="rooms"
                    :tree-type="activeName"
                    :tree-option="treeOption"
                    :tree-filter-text="filterText"
                    :show-device-search="showDeviceSearch"
                    :mozuloaded="mozuloaded"
                    :tree-data-prop="filteredTree"
                    :mozu-id="mozuId"
                    :warning-device-gid-list="warningDeviceGidList"
                    :enable-device-number-v2="enableDeviceNumberV2"
                    @node-click="handleClick"
                    @refresh="refresh"
                    @getTreeData="getTreeData"
                    @getAllTreeData="getAllTreeData"
                  />
                </div>
              </el-tab-pane>
              <el-tab-pane
                label="按类型"
                name="type"
                :lazy="true"
              >
                <div v-if="activeName === 'type'">
                  <div style="display:flex;">
                    <div class="search-wrap">
                      <el-select
                        v-model="treeOption"
                        border-type="no-border"
                      >
                        <el-option
                          label="全部"
                          value="all"
                        >
                          全部
                        </el-option>
                        <el-option
                          label="告警"
                          value="alarm"
                        >
                          告警
                        </el-option>
                      </el-select>

                      <el-input
                        v-model="filterText"
                        class="c-tree-tools__input"
                        placeholder="搜索设备"
                        prefix-icon="tn-icon-search"
                        clearable
                        border-type="no-border"
                      />
                      <i
                        :style="{cursor:'pointer', color: seletedAllColor ? 'gray' : '#1470CC'}"
                        class="tn-icon-filter"
                        @click="showSetModal"
                      />
                    </div>
                  </div>
                  <tree-component
                    data-source-type="byBiz"
                    class="data-query-tree"
                    :rooms="rooms"
                    :tree-option="treeOption"
                    :tree-type="activeName"
                    :tree-filter-text="filterText"
                    :show-device-search="showDeviceSearch"
                    :warning-device-gid-list="warningDeviceGidList"
                    :mozuloaded="mozuloaded"
                    :tree-data-prop="filteredTree"
                    :mozu-id="mozuId"
                    :enable-device-number-v2="enableDeviceNumberV2"
                    @node-click="handleClick"
                    @refresh="refresh"
                    @getTreeData="getTypeTreeData"
                  />
                </div>
              </el-tab-pane>
            </el-tabs>
          </transition>

          <resizable-divider
            v-model="leftPanelWidth"
            persistence-key="/tedge/data-query-index【main-layout】"
          />

          <realtime-focus-component
            v-if="activeName === 'focus'"
            ref="focusData"
            :view-type="viewType"
            class="data-query-realtime"
            :mozuloaded="mozuloaded"
            :mozu-id="mozuId"
            :card-obj="cardObj"
            :selected-data="selectedData"
            :enable-device-number-v2="enableDeviceNumberV2"
            :devices-map="devicesMap"
            @tree-visible-change="toggleTreeVisible"
            @collapse-change="changeTreeHeight"
            @getDeviceTypeList="getDeviceTypeList"
            @focused="focused"
            @mutifocused="mutifocused"
            @removeFocus="removeFocus"
          />
          <realtime-component
            v-else
            ref="realtime"
            class="data-query-realtime mid"
            :mozuloaded="mozuloaded"
            :mozu-id="mozuId"
            :tree-option="treeOption"
            :node-path-array="nodePathArray"
            :active-tab="activeName"
            :warning-attr-list="warningAttrList"
            :enable-control-setting="pageConfig.enableControlSetting"
            :enable-device-number-v2="enableDeviceNumberV2"
            :devices-map="devicesMap"
            @tree-visible-change="toggleTreeVisible"
            @collapse-change="changeTreeHeight"
            @getDeviceTypeList="getDeviceTypeList"
            @focused="focused"
            @mutifocused="mutifocused"
          />
        </div>
        <add-focus
          v-if="addVisible"
          :visible.sync="addVisible"
          :data="focusData"
          @confirm="focusConfirm"
        />
        <set-modal
          v-if="setVisible"
          :checked-rooms-prop="checkedRooms"
          :tree-data="treeDataProp"
          :visible.sync="setVisible"
          @confirm="confirm"
        />
        <set-type-modal
          v-if="setTypeVisible"
          :checked-rooms-prop="checkedRooms"
          :tree-data="treeDataProp"
          :type-data="deviceTypeList"
          :visible.sync="setTypeVisible"
          @confirm="typeModalConfirm"
        />
        <el-dialog
          title="添加关注"
          width="500px"
          :visible.sync="seleteFocusVisible"
        >
          <el-form label-width="130px">
            <el-form-item label="关注组">
              <el-select
                v-model="seleteFocusId"
                placeholder="请选择关注组"
              >
                <el-option
                  v-for="i in focusOptions"
                  :key="i.id"
                  :label="i.title"
                  :value="i.id"
                />
              </el-select>
            </el-form-item>
          </el-form>

          <span
            slot="footer"
            class="dialog-footer"
          >
            <el-button
              type="text"
              @click="seleteFocusVisible = false"
            >取消</el-button>
            <el-button
              type="text"
              @click="addFocusConfirm"
            >确定</el-button>
          </span>
        </el-dialog>
      </el-block>
    </template>
  </custom-config-value>
</template>

<script>
import business from '@@/config/business';
import treeComponent from './common/tree.vue';
import realtimeComponent from './common/realtime.vue';
import realtimeFocusComponent from './common/realtime-focus.vue';
import mixin from 'feature/utils/mixin';
import 'feature/utils/business';
import addFocus from './components/add-focus.vue';
import setModal from './components/set-modal.vue';
import setTypeModal from './components/set-type-modal.vue';
import { cloneDeep } from 'lodash';
import http from 'common/script/http2';
import getEdgeRequest from '../../utils/request';
import { warning as cgi } from '@@/config/cgi';
import dayjs from 'dayjs';
import ResizableDivider from '@/feature/component/resizable-divider.vue';
import CustomConfigValue from '@/feature/component/custom-config-value.vue';
import { forEachTreeNode } from '../../../utils/tree';
import { AlarmsCountByDeviceWatcher } from 'services/tedge/data-watchers/alarms.ts';

export default {
  components: {
    treeComponent,
    realtimeComponent,
    addFocus,
    setModal,
    realtimeFocusComponent,
    setTypeModal,
    ResizableDivider,
    CustomConfigValue,
  },
  mixins: [mixin],
  provide() {
    const that = this;
    return {
      getSelNodeData() {
        return that.nodeData;
      },
    };
  },
  props: {
    // showTitle: {
    //   type: Boolean,
    //   default: true,
    // },
  },
  data() {
    return {
      display: 'grid',
      treeWidth: window.localStorage.getItem('dataQueryIndexTreeWidth') || 240,
      business,
      nodeData: null,
      mozuloaded: false,
      mozuId: 0,
      treeVisible: true,
      activeName: 'position',
      addVisible: false,
      setVisible: false,
      treeOption: 'all',
      showDeviceSearch: false,
      filterText: '',
      treeData: [],
      treeDataProp: [],
      rooms: [],
      checkedRooms: [],
      filteredTree: [],
      seletedAllColor: true,
      treeLoaded: false,
      deviceTypeList: [],
      setTypeVisible: false,
      warningDeviceGidList: [],
      timer: null,
      focusGroups: [],
      focusOptions: [],
      focusData: {},
      seleteFocusId: null,
      seleteFocusVisible: false,
      selectedFocused: null,
      selectedData: {},
      selectedGroupId: null,
      cardObj: null,
      changeTimer: null,
      viewType: 'list',
      showMore: false,
      nodePathArray: [],
      warningList: [],
      warningAttrList: [],
      currentAlarmId: null,

      pageConfig: {},

      leftPanelWidth: 300,

      devicesMap: {},

      // alarmsWatcher: new AlarmsWatcher(3000).withDiffPlugin(),
      alarmsCountByDeviceWatcher: new AlarmsCountByDeviceWatcher(3000).withDiffPlugin(),
    };
  },
  mounted() {
    this.initPageConfig();

    this.changeMozu();
    // this.getActiveWarning();

    // this.timer = setInterval(() => {
    //   this.getActiveWarning();
    // }, 5000);
    this.getFocusGroup();
    this.watchActiveWarning();
  },
  beforeDestroy() {
    // clearInterval(this.timer);
    clearInterval(this.changeTimer);
    // this.alarmsWatcher.cancel();
    this.alarmsCountByDeviceWatcher.cancel();
  },
  methods: {
    async initPageConfig() {
      this.pageConfig = await window.tnwebServices.customConfigService.initCurrentPageConfig({
        url: window.location.pathname,
        content: {
          type: 'Yaml',
          // eslint-disable-next-line import/no-webpack-loader-syntax
          defaultContent: require('!raw-loader!./default-config.yaml').default,
        },
        docs: {
          type: 'markdown',
          // eslint-disable-next-line import/no-webpack-loader-syntax
          content: require('!raw-loader!./config.md').default,
        },
      });
    },

    mouseenterCard(data, e) {
      if (`${data.title}【${data.interestCount}项】` === e.target.innerText) {
        this.showMore = true;
      }
    },
    changeView(type) {
      this.viewType = type;
    },
    removeFocus(row) {
      this.$confirm('确认要取消关注吗？', '系统提示', { type: 'warning' }).then(() => {
        this.$axios.post('/cgi/dashboardaux/deleteInterestIndicator', { ids: row }).then(() => {
          this.$message.success('取消关注成功');
          this.getFocusGroup();
          this.changeCard(this.selectedGroupId);
        });
      });
    },
    mutifocused(rows) {
      this.selectedFocused = rows;
      this.seleteFocusVisible = true;
    },
    focused(row) {
      this.selectedFocused = row;
      this.seleteFocusVisible = true;
    },
    addFocusConfirm() {
      if (this.selectedFocused instanceof Array) {
        const selectedList = this.selectedFocused.map(i => ({ gid: i.gid,
          attr: i.attrId,
          group_id: this.seleteFocusId }));
        this.$axios.post(
          '/cgi/dashboardaux/addInterestIndicatorList',
          { InterestIndicators: selectedList },
        ).then(() => {
          this.getFocusGroup();
          this.seleteFocusVisible = false;
          this.$message.success('关注成功');
        });
      } else {
        this.$axios.post(
          '/cgi/dashboardaux/addInterestIndicator',
          { group_id: this.seleteFocusId, gid: this.selectedFocused.gid, attr: this.selectedFocused.attrId }
        ).then(() => {
          this.getFocusGroup();
          this.seleteFocusVisible = false;
          this.$message.success('关注成功');
        });
      }
    },
    focusConfirm() {
      this.doRefresh();
    },
    addNewFocus() {
      this.focusData = {};
      this.addVisible = true;
    },
    editFocus(id) {
      this.$axios.post('/cgi/dashboardaux/getOneInterestGroup', { id }).then((result) => {
        this.focusData = result;
        this.addVisible = true;
      });
    },
    deleteFocus(id) {
      this.$axios.post('/cgi/dashboardaux/deleteInterestGroup', { id }).then(() => {
        this.$message.success('删除成功');
        this.doRefresh();
      });
    },
    tabClick() {
      this.seletedAllColor = true;
      if (this.activeName === 'focus') {
        this.doRefresh();
      } else {
        clearInterval(this.changeTimer);
        this.changeTimer = null;
      }
      this.treeOption = 'all';
      this.filterText = '';
    },
    async doRefresh() {
      await this.getFocusGroup();
      if (!this.selectedGroupId || !this.focusGroups.find(i => i.id === this.selectedGroupId)) {
        this.selectedGroupId = this.focusGroups.length && this.focusGroups[0].id;
        this.changeCard(this.selectedGroupId, this.focusGroups[0]);
      } else {
        this.changeCard(this.selectedGroupId, this.focusGroups.find(i => i.id === this.selectedGroupId));
      }
    },
    getFocusGroup() {
      return this.$axios.get('/cgi/dashboardaux/getAllInterestGroup').then((result) => {
        this.focusGroups = result;
        this.focusOptions = result;
      });
    },
    watchActiveWarning() {
      const mozuId = localStorage.getItem('tidc_tedge_mozuId');
      if (!mozuId) return;

      const params = {
        eventStatus: -1,
        limit: 100000,
        mozuId: parseInt(mozuId),
        start: 0,
      };

      // this.alarmsWatcher.watch(params, (list) => {
      //   this.warningList = Object.freeze(list);
      //   // this.warningDeviceGidList = Object.freeze(list.map(i => i.deviceGid));

      //   // this.updateTreeAlarmStatus();
      // });

      this.alarmsCountByDeviceWatcher.watch(params, (data) => {
        this.warningDeviceGidList = Object.freeze(_.keys(data));
        this.updateTreeAlarmStatus();
      });
    },
    updateTreeAlarmStatus() {
      const warningDeviceGidSet = new Set(this.warningDeviceGidList);
      const pathIdsToAlarmSet = new Set();

      forEachTreeNode(this.treeData, (node) => {
        if (warningDeviceGidSet.has(node.id)) {
          node.pathMap.no.forEach((gid, i) => {
            pathIdsToAlarmSet.add(node.pathMap.no.slice(0, i + 1).join('/'));
          });
        }
      });

      forEachTreeNode(this.treeData, (node) => {
        const alarming = pathIdsToAlarmSet.has(node.pathId.toLowerCase());
        if (alarming !== node.alarming) {
          // eslint-disable-next-line no-param-reassign
          node.alarming = alarming;
        }
      });
    },
    changeCardInterval() {
      this.$axios.post(
        '/cgi/dashboardaux/getAllInterestIndicatorByGroupId',
        { group_id: this.selectedGroupId }, false
      ).then((result) => {
        const gidMap = result.map(i => ({ gids: [i.gid], attrs: [i.attr] }));
        const moduleId = '326';
        http.post(`/cgi/dataQuery/edge/getGidAndAttrListValueMapWithoutCache?mozuID=${moduleId}`, {
          gidWithAttrListMap: gidMap,
          start: 0,
          limit: 1000,
        }, false, {
          isJson: true,
          restAxios: {
            headers: {
              mozuId: moduleId,
              platform: 'cloud',
            },
          },
        }).then((pointData) => {
          result.forEach((i) => {
            pointData.list.forEach((j) => {
              if (i.gid === j.gid && i.attr === j.attrId) {
                j.focusId = i.id;
              }
            });
          });
          this.selectedData = pointData;
        });
      });
    },
    changeCard(id, cardObj) {
      this.selectedGroupId = id;
      this.cardObj = cardObj;
      this.changeCardInterval();
      clearInterval(this.changeTimer);
      this.changeTimer = null;
      if (!this.changeTimer) {
        this.changeTimer = setInterval(() => {
          this.changeCardInterval();
        }, 5000);
      }
    },

    getDeviceTypeList(deviceTypeList) {
      this.deviceTypeList = deviceTypeList;
    },
    showSetModal() {
      if (this.activeName === 'position') {
        this.checkedRooms = this.treeData[0].children.map(i => i.name);
        this.setVisible = true;
      } else {
        this.checkedRooms = this.treeData.map(i => i.id);
        this.setTypeVisible = true;
      }
    },
    getTreeData(treeData) {
      this.treeData = treeData;
      this.treeDataProp = cloneDeep(treeData);
      this.updateTreeAlarmStatus();
    },
    getAllTreeData(allTreeData) {
      const devicesMap = {};
      forEachTreeNode(allTreeData, (node) => {
        devicesMap[node.id] = node;
      });
      this.devicesMap = devicesMap;
    },
    getTypeTreeData(treeData) {
      this.treeData = treeData;
      this.treeDataProp = cloneDeep(treeData);
      this.updateTreeAlarmStatus();
    },
    confirm(data) {
      const tempTree = cloneDeep(this.treeData);
      tempTree[0].children = JSON.parse(JSON.stringify(this.treeDataProp[0].children
        .filter(i => data.checkedRooms.includes(i.name))));

      this.filteredTree = tempTree;
      this.treeData = tempTree;
      this.seletedAllColor = data.seletedAll;
    },
    typeModalConfirm(data) {
      let tempTree = cloneDeep(this.treeData);
      tempTree = JSON.parse(JSON.stringify(this.treeDataProp
        .filter(i => data.checkedRooms.includes(i.name))));

      this.filteredTree = tempTree;
      this.treeData = tempTree;
      this.seletedAllColor = data.seletedAll;
    },
    changeMozu() {
      this.mozuloaded = true;
      this.mozuId = TNBL.getCurrModule().id;
      this.nodeData = null;
    },
    // 获取告警的测点
    fetchInfo() {
      getEdgeRequest(this.$axios, this.mozuId).post(cgi.getWarningDetail, {
        AlarmId: this.currentAlarmId,
        MozuId: this.mozuId,
      }, false)
        .then((data) => {
          if (data && data.detail) {
            // this.dtl = data.detail;
            const occurTime = data.detail.alarm.OccurTime;
            if (occurTime) {
              this.selDateTime = [dayjs(occurTime).subtract(10, 'm')
                .format('YYYY-MM-DD HH:mm:ss'), dayjs(occurTime).add(10, 'm')
                .format('YYYY-MM-DD HH:mm:ss')];
              getEdgeRequest(this.$axios, this.mozuId).post(cgi.getPointDataType, {
                AlarmId: this.currentAlarmId,
                MozuId: this.mozuId,
              }, false)
                .then((data) => {
                  if (data && data.list.length > 0) {
                    this.type = data.list[0];
                    this.typeList = data.list;
                    this.fetchChartData();
                  }
                });
            };
          }
        });
    },
    fetchChartData() {
      getEdgeRequest(this.$axios, this.mozuId).post(cgi.getPointData, {
        AlarmId: this.currentAlarmId,
        MozuId: this.mozuId,
        PointType: this.type,
        StartTime: dayjs(this.selDateTime[0]).format('YYYY-MM-DD HH:mm:ss'),
        EndTime: dayjs(this.selDateTime[1]).format('YYYY-MM-DD HH:mm:ss'),
      })
        .then((data) => {
          this.warningAttrList = data.list.map(i => i.pointName);
        });
    },
    handleClick(data, nodePathArray) {
      this.nodeData = data;
      this.nodePathArray = nodePathArray;

      this.warningAttrList = [];
      const currentAlarmObj = this.warningList.find(i => i.deviceGid === data.id);

      if (!data.children || data.children.length === 0) {
        if (currentAlarmObj) {
          this.currentAlarmId = currentAlarmObj.alarmId;
          this.fetchInfo();
        }
      }
      this.refresh();
    },
    refresh() {
      const $child = this.$refs.realtime;
      $child.nodePathArrayProp = this.nodePathArray;
      $child.refresh();
    },

    toggleTreeVisible(val) {
      this.treeVisible = val;
    },

    changeTreeHeight(val) {
      document.querySelector('.c-tree-wrap').style.maxHeight = val ? '826px' : '666px';
    },
  },
};

</script>

<style lang="scss" scoped>
.box {
    width: 100%;
    height: 100%;
    margin:  0px;
    overflow: hidden;
    box-shadow: -1px 9px 10px 3px rgba(0, 0, 0, 0.11);
    /*左侧div样式*/
    .left {
        // width: calc(32% - 10px);  /*左侧初始化宽度*/
        height: 100%;
        background: #FFFFFF;
        float: left;
    }
    /*拖拽区div样式*/
    .resize {
      cursor: col-resize;
      float: left;
      position: relative;
      background-color: #d6d6d6;
      width: 5px;
      height: 56px;
      background-size: cover;
      background-position: center;
    }
    /*拖拽区鼠标悬停样式*/
    .resize:hover {
        color: #444444;
    }
    /*右侧div'样式*/
    .mid {
        float: left;
        // width: 68%;   /*右侧初始化宽度*/
        height: 100%;
        background: #fff;
        box-shadow: -1px 4px 5px 3px rgba(0, 0, 0, 0.11);
    }
}
/deep/ .el-tabs--card > .el-tabs__header .el-tabs__nav{
  border: none;
}
.data-query-container {
  margin-top: 3px;
}

.data-query-container-tabs {
  /deep/ .el-tab-pane {
    .data-query-tree {
      height: calc(100vh - 220px);
      padding: 0;
      overflow: hidden;
      // overflow: overlay;
      &:hover {
        overflow: auto;
      }
    }

  }
}

.focus-card {
  margin-bottom:6px;
  box-shadow:0px 1px 3px 0 #cbcbcb80;
  min-height: 90px;
  &:hover {
    background-color: #f3f2f2;
      // box-shadow: 2px 5px 8px 1px #cbcbcb80;
  }
}

.active-group {
  color: #1470CC;
}

.data-query {
  &-container {
    display: flex;
  }

  &-tree {
    // width: 240px;
    height: 100%;

    /deep/ .c-tree-wrap {
    // height: calc(100vh - 300px);
    // overflow:scroll;
    }
  }

  &-realtime {
    flex: 1;
    // border-left: 1px solid #f0f0f0;
  }
  .search-wrap {
      border-bottom: solid 1px #c0c0c0;
      display:flex;
      padding: 5px 16px 5px 12px;
      margin:5px 0;
      flex: 1;
      .el-select {
        width:92px;
        border-right:solid 1px #c0c0c0;
        padding-right: 10px;
        margin-right: 10px
      }
  }
  .focus-wrap {
    text-align:center;
    padding: 10px 10px;
    .el-button {
      width:100%;
      border-radius: 3px;
      border-color:#c0c0c082
    }
  }
}

.tree-enter-active, .tree-leave-active {
  transition: all .3s;
}
.tree-enter, .tree-leave-to {
  opacity: 0;
  width: 0;
}
  .card-title-wrap {
    padding: 0 10px 0 24px;
  }
  .card-title {
    height: 24px;
    line-height: 24px;
    padding: 16px 0;
    color: #333;
    display: flex;
    &-main {
      flex: 1;font-family: 'SimHei';
      font-size: 15px
    }
  }

  .card-content {
    font-size: 12px;
    color: #666;
    text-align: justify;
    line-height: 24px;
    overflow: hidden;
    // white-space: wrap;
    // text-overflow: ellipsis;
    .time {
      -webkit-line-clamp: 1;
      display: -webkit-box;
      -webkit-box-orient: vertical;
    }
  }
</style>

<style lang="scss">
.data-query {
  &-container {
    .el-tabs__item {
      padding: 0px;
      min-width: 100px;
    }
  }
}
</style>
