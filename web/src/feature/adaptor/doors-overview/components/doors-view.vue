<template>
  <div class="doors-view">
    <div
      v-if="!doors.length"
      class="empty-text"
    >
      暂无数据
    </div>

    <template v-else>
      <header>
        <div class="header-left">
          <el-checkbox
            v-model="allChecked"
            :indeterminate="allChecked && checkedCount !== doors.length"
            :label="true"
          >
            全选
          </el-checkbox>

          <el-button
            size="small"
            @click="setAllOpen"
          >
            全开
          </el-button>

          <el-button
            size="small"
            @click="setAllClose"
          >
            全关
          </el-button>

          <el-button
            size="small"
            @click="setAllAlwaysOpen"
          >
            常开
          </el-button>

          <el-button
            :disabled="!checkedCount"
            size="small"
            @click="showCheckedVideos"
          >
            查看视频
          </el-button>

          <el-button
            :disabled="!checkedCount"
            size="small"
            @click="showRecords()"
          >
            查看刷卡记录
          </el-button>

          <admin-limit-tooltips>
            <span
              slot-scope="{ hasRight }"
              style="margin-left: 8px;"
            >
              <el-button
                :disabled="!hasRight || !checkedCount"
                size="small"
                @click="startSetDoorParams"
              >
                设置门参数
              </el-button>
            </span>
          </admin-limit-tooltips>
        </div>

        <div class="header-right">
          <div
            v-for="(summary, i) in summaryMap"
            :key="i"
            class="summary-item"
          >
            <div class="summary-item-label">
              {{ summary.label }}
            </div>
            <div
              class="summary-item-value"
              :class="summary.status"
            >
              {{ summary.value }}
            </div>
          </div>
        </div>
      </header>

      <main class="door-list-container">
        <div class="door-list">
          <el-card
            v-for="(door, i) in doors"
            :key="i"
            :body-style="{ padding: '0px' }"
            class="door-card"
            shadow="hover"
          >
            <div
              class="card-image"
              :style="{background: `url(${getDoorStatusImage(door)})`}"
            >
              <div class="door-oprs">
                <el-checkbox
                  :key="selectedDoorsMap[door.id] ? 1 : 0"
                  v-model="selectedDoorsMap[door.id]"
                  class="door-checkbox"
                />

                <el-button
                  type="text"
                  icon="tn-icon-security"
                  class="video-btn"
                  @click="showVideosByDoors([door])"
                />

                <div class="door-oprs-btns">
                  <el-button
                    type="text"
                    @click="openDoor(door, true)"
                  >
                    开门
                  </el-button>
                  <el-button
                    type="text"
                    @click="openDoorByAlways(door)"
                  >
                    常开
                  </el-button>
                  <el-button
                    type="text"
                    @click="showRecords([door])"
                  >
                    刷卡记录
                  </el-button>
                </div>
              </div>
            </div>

            <div class="card-info">
              <div class="card-title">
                {{ door.name }}
              </div>
              <div class="card-content">
                {{ door.controlName }}
              </div>
            </div>
          </el-card>
        </div>
      </main>
    </template>

    <videos-play-modal
      ref="videosPlayModal"
      force-replace-info
      empty-alert-text="所选门尚未关联摄像头，请先关联摄像头。"
    />
  </div>
</template>

<script>
import AdminLimitTooltips from 'feature/component/tedge-components/admin-limit-tooltips.vue';
import doorCloseSvg from './door-close.svg';
import doorOpenSvg from './door-open.svg';
import VideosPlayModal from 'module/tisspage/im-security-camera/videos-play-modal.vue';
import { DcosRtdWatcher } from 'services/tedge/data-watchers/dcos-rtd.ts';
import getEdgeRequest from 'feature/utils/request';

export default {
  components: {
    AdminLimitTooltips,
    VideosPlayModal,
  },
  props: {
    doors: {
      type: Array,
      default() {
        return null;
      },
    },
  },
  data() {
    return {
      doorsStatusMap: {},
      selectedDoorsMap: {},

      summaryMap: {
        totalDoors: {
          label: '门总数',
          value: 0,
        },
        doorsOpened: {
          label: '开的门',
          value: 0,
          status: 'warning',
        },
        doorsWithError: {
          label: '故障门',
          value: 0,
          status: 'error',
        },
      },

      statusMap: {
        open: {
          icon: doorOpenSvg,
        },
        close: {
          icon: doorCloseSvg,
        },
      },

      rtdWatcher: new DcosRtdWatcher(3000)
        .withDiffPlugin()
        .bindVueVm(this),
    };
  },
  computed: {
    checkedCount() {
      return _.chain(this.selectedDoorsMap)
        .filter(_.identity)
        .size()
        .value();
    },
    allChecked: {
      get() {
        return this.checkedCount > 0;
      },
      set(v) {
        this.toggleCheckedAll(v);
      },
    },
  },
  watch: {
    doors() {
      this.selectedDoorsMap = {};
      this.loadDoorsStatus();
      this.summaryMap.totalDoors.value = this.doors?.length || 0;
    },
  },
  created() {
    this.loadDoorsStatus = this.$intervalFunction(this.loadDoorsStatus, 1000, true);
  },
  methods: {
    async loadDoorsStatus() {
      const {
        doors,
      } = this;

      if (!doors?.length) return;

      const ids = _.map(doors, 'state_id');

      // const statusDataMap = await this.$axios.post('/api/dcos/tdac-cgi/rtd', {
      //   ids,
      // }, null, false);
      const statusDataMap = await this.rtdWatcher.mockRequest({
        ids,
      });

      this.doorsStatusMap = statusDataMap;
      this.updateSummary();
    },
    updateSummary() {
      const {
        doorsStatusMap,
        summaryMap,
      } = this;

      const statusCounts = _.countBy(doorsStatusMap, 'pv');
      summaryMap.doorsOpened.value = statusCounts['1'] || 0;

      // TODO: 补齐故障数量
      // summaryMap.doorsWithError.value = statusCounts['1'] || 0;
    },

    toggleCheckedAll(isChecked) {
      this.selectedDoorsMap = !isChecked ? {} : _.chain(this.doors)
        .map(door => ([door.id, true]))
        .fromPairs()
        .value();
    },
    getDoorStatusName(door, replacement = 'close') {
      const pv = this.doorsStatusMap[door.state_id]?.pv;
      return {
        0: 'close',
        1: 'open',
      }[pv] || replacement;
    },
    getDoorStatusImage(door) {
      const statusName = this.getDoorStatusName(door);

      const status = this.statusMap[statusName] || this.statusMap.close;
      return status.icon;
    },

    getDoorIds(doors) {
      return _.map(doors, 'id');
    },

    toggleDoorsStatusByIds(ids, isOpen, isAlways = false) {
      if (!ids?.length) return;

      return getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/door/state', {
        ids: ids.map(Number),
        // eslint-disable-next-line no-nested-ternary
        state: isAlways
          ? (isOpen ? 2 : 3)
          : (isOpen ? 1 : 0),
      });
    },

    toggleDoorsStatus(doors, isOpen, isAlways = false) {
      if (!doors?.length) return;
      return this.toggleDoorsStatusByIds(
        this.getDoorIds(doors),
        isOpen,
        isAlways,
      );
    },

    getSelectedDoorIds() {
      return _.chain(this.selectedDoorsMap)
        .toPairs()
        .filter(([, v]) => v)
        .map(item => item[0])
        .value();
    },

    getSelectedDoorIdsWithEmptyAlert() {
      const ids = this.getSelectedDoorIds();

      if (ids?.length) return ids;

      this.$message.info('请先勾选门后操作');

      throw new Error('No Ids');
    },

    // 操作：
    async setAllOpen() {
      await this.toggleDoorsStatusByIds(this.getSelectedDoorIdsWithEmptyAlert(), true);
      this.$message.success('全部开启成功');
    },
    async setAllClose() {
      await this.toggleDoorsStatusByIds(this.getSelectedDoorIdsWithEmptyAlert(), false);
      this.$message.success('全部关闭成功');
    },
    async setAllAlwaysOpen() {
      await this.toggleDoorsStatusByIds(this.getSelectedDoorIdsWithEmptyAlert(), true, true);
      this.$message.success('全部常开成功');
    },
    showVideos() {
      // TODO: 下一期实现
      console.log('TODO: showVideos');
    },
    startSetDoorParams() {
      const ids = this.getSelectedDoorIdsWithEmptyAlert();
      this.$emit('batchSetDoorParams', ids);
    },

    async openDoor(door) {
      await this.toggleDoorsStatus([door], true);
      this.$message.success('开门成功');
    },
    async openDoorByAlways(door) {
      await this.toggleDoorsStatus([door], true, true);
      this.$message.success('常开门成功');
    },
    showRecords(doors) {
      if (doors) {
        this.$emit('showRecords', doors);
        return;
      }

      const doorsMap = _.mapKeys(this.doors, 'id');
      const doorsToShow = doors || _.chain(this.selectedDoorsMap)
        .map((v, k) => v && doorsMap[k])
        .filter(Boolean)
        .value();

      this.$emit('showRecords', doorsToShow);
    },
    showVideosByDoors(doors) {
      const cameraList = _.chain(doors)
        .map(item => _.get(item, 'extend.relatedCameras', []))
        .flatten()
        .union()
        .filter(Boolean)
        .map(cameraNumber => ({
          device_number: cameraNumber,
          device_No: cameraNumber,
        }))
        .value();

      this.$refs.videosPlayModal.show(cameraList);
    },
    showCheckedVideos() {
      const doorsMap = _.mapKeys(this.doors, 'id');
      const doorsToShow = _.chain(this.selectedDoorsMap)
        .map((v, k) => v && doorsMap[k])
        .filter(Boolean)
        .value();

      this.showVideosByDoors(doorsToShow);
    },
  },
};
</script>

<style lang="scss" scoped>
.doors-view {
  padding: 16px 24px;
  height: 100%;
  overflow: hidden;

  display: flex;
  flex-direction: column;
}

.empty-text {
  color: #999;
  font-weight: 500;
  text-align: center;
  line-height: 48px;
  margin-top: 32px;
}

header {
  display: flex;
  align-items: end;
  min-width: 1000px;
}

.header-left {
  flex: 1;
}

.header-right {
  display: flex;
}

.summary-item {
  display: flex;
  align-items: flex-end;
  margin: 0 16px;
  position: relative;

  &:not(:first-child):before {
    content: '';
    display: block;

    position: absolute;
    top: 6px;
    left: -16px;
    height: 20px;
    border-left: 1px solid #e0e0e0;
  }
}

.summary-item-label {
  font-size: 16px;
  color: #999;
  position: relative;
  top: -4px;
  margin-right: 4px;
}

.summary-item-value {
  color: #333;
  font-size: 32px;
  font-weight: 600;

  &.error {
    color: var(--tn-color-danger);
  }
  &.warning {
    color: var(--tn-color-warning);
  }
}

.door-list-container {
  flex: 1;
  overflow: auto;
}

.door-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;

  margin-top: 24px;
  padding-bottom: 24px;
}

.door-card {
  width: 240px;
  height: 196px;
}

.door-oprs {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;

  & > *:not(.door-checkbox) {
    visibility: hidden;
  }
}

.door-checkbox {
  position: absolute;
  top: 14px;
  left: 14px;
}

.video-btn {
  position: absolute;
  top: 14px;
  right: 14px;
  color: #b088ff;
  cursor: pointer;
}

.disabled {
  color: #d0d0d0;
  cursor: not-allowed;
}

.door-oprs-btns {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  text-align: center;
}
</style>

<style lang="scss" scoped>
.card-container {
  display: inline-block;
  width: 312px;
  margin-right: 16px;
  overflow: hidden;
}

.card-container:hover .image {
  transform: scale(1.1);
}

.card-image {
  height: 136px;
  overflow: hidden;
  position: relative;
  background-repeat: no-repeat !important;
  background-position: center !important;
  background-size: 90% 60% !important;
  background-color: #f3f3f3 !important;

  &:hover {
    .door-oprs > * {
      visibility: visible;
    }
  }
}

.card-info {
  padding: 8px 16px 4px 16px;
}

.card-title {
  line-height: 24px;
  color: #333;
  font-weight: 600;
}

.card-content {
  font-size: 12px;
  color: #666;
  text-align: justify;
  line-height: 24px;
}

.card-bottom {
  height: 56px;
  line-height: 56px;
  text-align: right;
}
</style>
