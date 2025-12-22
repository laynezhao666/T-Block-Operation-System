<template>
  <div>
    <el-table
      :key="mode"
      :data="showingTableData"
      :height="height ? (height - 12) : null"
      stripe
    >
      <el-table-column
        label="等级"
        width="90px"
      >
        <template #default="{ row }">
          <el-tag :type="getLevelTagType(mode === 'active' ? row.level : row.alarmLevel)">
            {{ getLevelName(mode === 'active' ? row.level : row.alarmLevel) }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column
        label="告警类型"
      >
        <template #default="{ row }">
          <el-button
            type="text"
            class="el-button__text-ellipsis"
            @click="openAlarmDetailsModal(row)"
          >
            {{ row.alarmType }}
          </el-button>
        </template>
      </el-table-column>

      <el-table-column
        label="告警原因"
        :prop="mode === 'active' ? 'content' : 'alarmContent'"
      />

      <el-table-column
        label="产生时间"
        prop="occurTime"
        width="180px"
      />

      <el-table-column
        v-if="mode === 'history'"
        label="恢复时间"
        prop="restoreTime"
        width="180px"
      />

      <el-table-column
        label="持续时间"
      >
        <template #default="{ row }">
          {{ computeDuration(row) }}
        </template>
      </el-table-column>

      <el-table-column
        v-if="mode === 'active'"
        label="状态"
      >
        <template #default="{ row }">
          <el-tag
            :type="getStatusTagType(row.eventStatus)"
          >
            {{ getStatusName(row.eventStatus) }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      :current-page.sync="pagination.current"
      :page-size.sync="pagination.size"
      :page-sizes="pagination.sizes"
      :total="total"
      layout="total, prev, pager, next, sizes, jumper"
      styled
      background
    />

    <detail-modal
      v-if="detailsVisible"
      :visible.sync="detailsVisible"
      :data="showingDetailsAlarm"
    />
  </div>
</template>

<script>
import dayjs from 'dayjs';
import DetailModal from 'feature/warning/actived-warning/detail-modal.vue';

const LEVEL_CONFIG_MAP = {
  L0: {
    label: '零级',
    tagType: 'danger',
  },
  L1: {
    label: '一级',
    tagType: 'danger',
  },
  L2: {
    label: '二级',
    tagType: 'warning',
  },
  L3: {
    label: '三级',
    tagType: 'warning',
  },
  L4: {
    label: '四级',
    tagType: 'primary',
  },
  L5: {
    label: '五级',
    tagType: 'primary',
  },
};

const EVENT_STATUS_CONFIG_MAP = {
  1: {
    label: '未处理',
    tagType: 'danger',
  },
  2: {
    label: '已转单',
    tagType: 'primary',
  },
};

export default {
  components: {
    DetailModal,
  },
  props: {
    alarms: {
      type: Array,
      default() {
        return [];
      },
    },
    mode: {
      type: String,
      default() {
        return 'active';
      },
    },
    height: {
      type: Number,
      default() {
        return null;
      },
    },
    fetchHistoryAlarms: {
      type: Function,
      required: true,
    },
    nodeData: {
      type: Object,
      default() {
        return null;
      },
    },
    historyTotal: {
      type: Number,
      default() {
        return null;
      },
    },
  },
  data() {
    return {
      now: dayjs(),

      pagination: {
        current: 1,
        size: 10,
        sizes: [
          10,
          15,
          20,
          25,
          30,
          50,
          100,
        ],
      },

      showingDetailsAlarm: null,
    };
  },
  computed: {
    showingTableData() {
      const {
        pagination: {
          current,
          size,
        },
        alarms,
      } = this;

      return alarms?.slice((current - 1) * size, size * current);
    },
    total() {
      const {
        mode,
        alarms,
        historyTotal,
      } = this;

      if (mode === 'active') return alarms ? alarms.length : 0;

      return historyTotal || 0;
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
  watch: {
    pagination() {
      if (this.mode === 'history') {
        this.fetchHistoryAlarms(this.pagination);
      }
    },
    mode() {
      if (this.mode === 'history') {
        this.fetchHistoryAlarms(this.pagination);
      }
    },
    nodeData() {
      if (this.mode === 'history') {
        this.fetchHistoryAlarms(this.pagination);
      }
    },
  },
  mounted() {
    this.updateNowInterval = setInterval(() => {
      this.now = dayjs();
    }, 1000);
  },
  beforeDestroy() {
    if (this.updateNowInterval) {
      clearInterval(this.updateNowInterval);
    }
  },
  methods: {
    getLevelName(level) {
      return LEVEL_CONFIG_MAP[level]?.label;
    },
    getLevelTagType(level) {
      return LEVEL_CONFIG_MAP[level]?.tagType;
    },
    openAlarmDetailsModal(alarm) {
      this.showingDetailsAlarm = alarm;
    },
    getStatusName(status) {
      return EVENT_STATUS_CONFIG_MAP[status]?.label;
    },
    getStatusTagType(status) {
      return EVENT_STATUS_CONFIG_MAP[status]?.tagType;
    },
    computeDuration(row) {
      const occurTime = dayjs(row.occurTime);
      let endTime;
      if (row.restoreTime) {
        endTime = dayjs(row.restoreTime);
      } else if (row.closeTime) {
        endTime = dayjs(row.closeTime);
      } else {
        endTime = this.now;
      }
      let seconds = endTime.diff(occurTime) / 1000;
      const days = Math.floor(seconds / (24 * 3600));
      seconds = seconds % (24 * 3600);
      const hours = Math.floor(seconds / 3600);
      seconds = seconds % 3600;
      const mins = Math.floor(seconds / 60);
      seconds = Math.floor(seconds % 60);
      return `${days ? `${days}天` : ''}${hours ? `${hours}小时` : ''}${mins ? `${mins}分钟` : ''}${seconds ? `${seconds}秒` : ''}`;
    },
  },
};
</script>

<style lang="scss" scoped>
.el-button__text-ellipsis {
  width: 100%;
  /deep/ span {
    display: block;
    width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}
</style>
