<template>
  <el-popover
    v-model="popoverVisible"
    trigger="manual"
    placement="bottom"
  >
    <div class="volume-oprs-container">
      <el-button
        type="primary"
        size="mini"
        @click="turnSound('off')"
      >
        静音60分钟
      </el-button>

      <el-button
        type="primary"
        size="mini"
        @click="confirmAllAlarms"
      >
        确认当前所有告警
      </el-button>

      <i
        class="el-icon-error close-btn"
        @click="popoverVisible = false"
      />
    </div>

    <el-button
      slot="reference"
      type="text"
      @click="handleClick"
      @touchend="handleClick"
    >
      <img
        :src="`/static/pad/images/volume-${enable ? 'on' : 'off'}.svg`"
      >
    </el-button>
  </el-popover>
</template>

<script>
export default {
  data() {
    return {
      enable: true,
      popoverVisible: false,
    };
  },
  watch: {
    popoverVisible: {
      immediate: true,
      handler() {
        setTimeout(() => {
          document.querySelector('#app_main')
            .classList[this.popoverVisible ? 'add' : 'remove']('dark-marks');
        }, 100);
      },
    },
  },
  mounted() {
    this.loadStatus();
    this.interval = setInterval(() => {
      this.loadStatus();
    }, 3000);
  },
  beforeDestroy() {
    clearInterval(this.interval);
  },
  methods: {
    async loadStatus() {
      const data = await this.$axios.get('/cgi/linkage/freeze', null, false);
      this.enable = !data.freeze;
    },
    async turnSound(onOff) {
      await this.$axios.post('/cgi/linkage/freeze', {
        freeze: onOff === 'off' ? 1 : 0,
        freeze_second: 60 * 60,
      });
      this.$message.success(onOff === 'off' ? '静音60分钟成功' : '解除静音60分钟成功');
      this.popoverVisible = false;
      this.loadStatus();
    },
    async confirmAllAlarms() {
      await this.$axios.post('/cgi/alarm/confirm/all', {
        is_confirm: 1,
      });
      this.$message.success('确认当前所有告警成功');
      this.popoverVisible = false;
    },
    handleClick() {
      if (this.enable) {
        this.popoverVisible = true;
      } else {
        this.turnSound('on');
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.volume-oprs-container {
}

.close-btn {
  margin-left: 16px;
}
</style>
