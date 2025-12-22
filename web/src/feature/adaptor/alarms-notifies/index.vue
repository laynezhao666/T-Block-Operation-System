<template>
  <div>
    <transition-group
      v-show="alarms.length > 0 && !ignoreNotifiesEnd"
      :css="enableTransition"
      class="alarms-notifies"
      name="alarm-card"
      tag="div"
    >
      <div
        v-for="(alarm, i) in showingAlarms"
        :key="alarm.AlarmId"
        class="alarm-card"
      >
        <header :style="{ color: alarmsDir[alarm.Level] && alarmsDir[alarm.Level].color }">
          【
          {{ alarmsDir[alarm.Level] && alarmsDir[alarm.Level].label }}
          】
          {{ alarm.AlarmType }}
          <el-button
            type="text"
            class="close-btn"
            @click="close(alarms.length - i - 1)"
          >
            <i class="tn-icon-close small-icon" />
          </el-button>
        </header>

        <main>
          <p>
            <em>【{{
              freezedData.v2DeviceNumberTransformerService.get(alarm.DeviceNumber)
            }}】</em>
            {{ alarm.Content }}
          </p>

          <p class="flex-between">
            <el-button
              type="text"
              @click="showDetails(alarm)"
            >
              详情
            </el-button>

            {{ alarm.OccurTime }}
          </p>
        </main>
      </div>

      <div
        key="oprs"
        class="alarm-card alrams-oprs"
      >
        <el-button
          :disabled="alarms.length <= maxShowingCount"
          type="text"
          @click="next"
        >
          下一批
        </el-button>

        <el-button
          type="text"
          @click="clear"
        >
          关闭所有({{ alarms.length }})
        </el-button>

        <div class="space" />

        <el-button
          type="text"
          @click="ignoreNotifiesInTime(30, 'minutes')"
        >
          30分钟不再弹出
        </el-button>
      </div>
    </transition-group>

    <!-- 这个组件的实现很奇葩，不用v-if会报错，其他地方也是这么用v-if的 -->
    <detail-modal
      v-if="detailsVisible"
      key="detailModal"
      :visible.sync="detailsVisible"
      :data="showingDetailsAlarm"
    />
  </div>
</template>

<script>
import Cookie from 'js-cookie';
import dayjs from 'dayjs';
import getEdgeRequest from '../../utils/request';

const COMFIRMED_STORAGE_KEY = `totalConfirmedAlarmIds__${Cookie.get('tnebula_username') || 'common'}`;
const IGNORE_NOTIFIES_END_STORAGE_KEY = 'ignore_alarms_notifies_end';

export default {
  components: {
    DetailModal: () => import('feature/warning/actived-warning/detail-modal.vue'),
  },
  data() {
    return {
      alarms: [],
      maxShowingCount: 3,
      alarmsDir: {
        L0: {
          label: '零级',
          color: 'var(--tn-color-danger)',
        },
        L1: {
          label: '一级',
          color: 'var(--tn-color-danger)',
        },
        L2: {
          label: '二级',
          color: 'var(--tn-color-warning)',
        },
        L3: {
          label: '三级',
          color: 'var(--tn-color-warning)',
        },
        L4: {
          label: '四级',
          color: 'var(--tn-color-warning)',
        },
        L5: {
          label: '五级',
          color: 'var(--tn-color-warning)',
        },
      },
      confirmedAlarmIdsSet: this.getInitConfirmedAlarmIdsSet(),

      freezedData: Object.freeze({
        v2DeviceNumberTransformerService: window.tnwebServices.v2DeviceNumberTransformerService,
        // confirmed || notified
        // alarmIds
      }),

      ignoreNotifiesEnd: this.getInitIgnoreNotifiesEnd(),

      showingDetailsAlarm: null,

      enableTransition: !document.hidden,
    };
  },
  computed: {
    showingAlarms() {
      const {
        alarms,
        maxShowingCount,
      } = this;

      return _.range(alarms.length - 1, alarms.length - maxShowingCount - 1)
        .map(i => alarms[i])
        .filter(Boolean);
    },
    detailsVisible: {
      get() {
        return !!this.showingDetailsAlarm;
      },
      set(v) {
        if (!v) {
          this.showingDetailsAlarm = null;
        }
      },
    },
  },
  created() {
    document.addEventListener('visibilitychange', this.handleDocumentVisibilityChange);
  },
  methods: {
    getInitConfirmedAlarmIdsSet() {
      const ids = (sessionStorage.getItem(COMFIRMED_STORAGE_KEY) || '').split(';');
      const set = new Set(ids);
      return Object.freeze(set);
    },
    getInitIgnoreNotifiesEnd() {
      const endTs = sessionStorage.getItem(IGNORE_NOTIFIES_END_STORAGE_KEY) || null;
      setTimeout(() => {
        this.ignoreNotifiesUntil(endTs);
      }, 0);
      return endTs;
    },

    push(alarm) {
      if (
        this.confirmedAlarmIdsSet.has(alarm.AlarmId.toString())
        || this.alarms.some(item => (item.AlarmId === alarm.AlarmId))
      ) return;

      this.alarms.push(alarm);
    },
    pushAll(alarms) {
      if (!this.$$notifiedAlarmIdSet) {
        this.$$notifiedAlarmIdSet = new Set([...this.confirmedAlarmIdsSet.values()]);
      }

      // 缓存已确认和已通知过的id，提高性能，[].some性能很差，在7、8千条数据情况下，一次对比耗时近3s，缓存为set用has后提升到毫秒级
      const { $$notifiedAlarmIdSet } = this;

      for (let i = 0; i < alarms.length; i++) {
        const alarm = alarms[i];
        if (
          !$$notifiedAlarmIdSet.has(alarm.AlarmId)
        ) {
          this.alarms.push(alarm);
          $$notifiedAlarmIdSet.add(alarm.AlarmId);
        }
      }
    },
    clear() {
      this.triggerCloseAlarms(this.alarms);
      this.alarms = [];
    },
    close(index) {
      const closedAlarms = this.alarms.splice(index, 1);
      this.triggerCloseAlarms(closedAlarms);
    },
    next() {
      const closedAlarms = this.alarms.splice(this.alarms.length - this.maxShowingCount, this.maxShowingCount);
      this.triggerCloseAlarms(closedAlarms);
    },
    ignoreNotifiesInTime(duration, unit) {
      const endDayjs = dayjs().add(duration, unit);
      this.ignoreNotifiesUntil(endDayjs);

      this.ignoreNotifiesEnd = endDayjs.format('YYYY-MM-DD HH:mm:ss');
      sessionStorage.setItem(IGNORE_NOTIFIES_END_STORAGE_KEY, this.ignoreNotifiesEnd);
    },
    ignoreNotifiesUntil(endTime) {
      const diffTs = dayjs(endTime).diff(Date.now(), 'millseconds');

      setTimeout(() => {
        this.ignoreNotifiesEnd = null;
        sessionStorage.removeItem(IGNORE_NOTIFIES_END_STORAGE_KEY);
      }, Math.max(diffTs, 0));
    },
    triggerCloseAlarms(alarms) {
      const { confirmedAlarmIdsSet } = this;
      const newConfirmedAlarmIds = alarms.map(alarm => alarm.AlarmId.toString()).join(';');

      const confirmedString = sessionStorage.getItem(COMFIRMED_STORAGE_KEY) || '';
      sessionStorage.setItem(COMFIRMED_STORAGE_KEY, `${newConfirmedAlarmIds};${confirmedString}`);

      alarms.forEach((alarm) => {
        confirmedAlarmIdsSet.add(alarm.AlarmId.toString());
      });
    },

    async showDetails(alarm) {
      const { list: matchedAlarms } = await getEdgeRequest(this.$axios, '').post('/cgi/alarm/active/getList', {
        eventStatus: 1,
        limit: 100,
        AlarmId: [alarm.AlarmId],
        DeviceNumber: [alarm.DeviceNumber],
      }, false, false);

      const matchedAlarm = matchedAlarms.find(item => item.id === alarm.Id);

      this.showingDetailsAlarm = {
        ...matchedAlarm,
      };
    },

    handleDocumentVisibilityChange() {
      this.enableTransition = !document.hidden;
    },
  },
};
</script>

<style lang="scss" scoped>
.alarms-notifies {
  position: fixed;
  top: 48px;
  right: 8px;

  width: 340px;
  z-index: 9999;
}

.alarm-card {
  position: relative;

  background-color: #FFFFFF;
  color: #333;
  padding: 14px 26px 14px 13px;

  border: 1px solid #ebebeb;
  border-radius: 4px;
  box-shadow: 0 3px 5px 0 hsla(0,0%,79.6%,.5);

  margin-bottom: 8px;

  header {
    font-weight: 700;
    font-size: 14px;
    color: #333;
    margin: 0;
  }

  main {
    font-size: 12px;
    margin-top: 12px;
  }

  p {
    margin-top: 8px;
  }

  em {
    font-weight: 600;
  }
}

.close-btn {
  position: absolute;
  top: 10px;
  right: 14px;
}

.text-right {
  text-align: right;
}

.indent {
  text-indent: 2em;
}

.alrams-oprs {
  display: flex;
}

.space {
  flex: 1;
}

.small-icon {
  font-size: 12px;
}

.flex-between {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.alarm-card-enter-active, .alarm-card-leave-active {
  transition: all 0.5s;
}
.alarm-card-enter, .alarm-card-leave-to /* .alarm-card-leave-active below version 2.1.8 */ {
  transform: translateX(100%) scaleX(0);
  transform-origin: bottom right;
  height: 0;
  overflow: hidden;
  padding: 0;
}
.alarm-card-enter-active, .alarm-card-enter {
  transition-delay: 0.4s;
}
</style>
