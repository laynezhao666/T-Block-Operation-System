<template>
  <custom-config-value
    class="header"
    name="safe_operating_time_start"
  >
    <template #default="{ value: startTimeText }">
      <div
        class="header-bg"
      >
        <img :src="backgroundUrl">
      </div>
      <div class="header-bg-color" />
      <div class="location-weather-wrapper">
        <LocationWeather :mozu-info="mozuInfo" />
        <slot name="footer-button" />
      </div>
      <div
        class="header-content"
      >
        <div class="logo-container">
          <img
            :src="logoUrl"
            alt=""
          >
          <slot name="status-bar">
            <StatusBars />
          </slot>
        </div>

        <div
          class="header-main"
        >
          <slot name="header-main">
            <div class="header-main-title">
              {{ mainTitle }}
            </div>
            <div class="header-main-mozu-title">
              {{ mozuInfo.mozu }}
            </div>
            <div class="header-main-subtext">
              已安全运营 <span>{{ formatDurationToNow(startTimeText) }}</span> 天
            </div>
          </slot>
        </div>
      </div>
    </template>
  </custom-config-value>
</template>

<script>
import dayjs from 'dayjs';
import LocationWeather from '../built-in/LocationWeather';
import StatusBars from '../built-in/StatusBar';

export default {
  components: {
    LocationWeather,
    StatusBars,
  },

  props: {
    mozuInfo: {
      type: Object,
      default() {
        return {};
      },
    },
    backgroundUrl: {
      type: String,
      default: '',
    },
    logoUrl: {
      type: String,
      default: '',
    },
    mainTitle: {
      type: String,
      default: '',
    },

  },

  data() {
    return {
      now: Date.now(),
    };
  },
  created() {
    this.$$updateNowInterval = setInterval(() => {
      this.now = Date.now();
    }, 1000);
  },
  beforeDestroy() {
    if (this.$$updateNowInterval) {
      clearTimeout(this.$$updateNowInterval);
    }
  },
  methods: {
    formatDurationToNow(startTimeText) {
      return dayjs(this.now).diff(startTimeText, 'day');
    },
  },

};
</script>
<style lang="scss" scoped>
@import './../style/header.scss';
</style>
