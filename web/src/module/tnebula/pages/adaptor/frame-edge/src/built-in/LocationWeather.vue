<template>
  <div
    v-if="info"
    class="location-weather"
  >
    <span class="city">
      {{ mozuInfo.park }}
    </span>
    <span class="text-spacer" />
    <!-- <span>
      {{ getDataItem('forecast_24h[1].day_weather') }}
    </span>
    <span class="text-spacer" />
    <span
      class="temp"
    >{{ getDataItem('forecast_24h[1].min_degree') }}℃</span>
    ～
    <span
      class="temp"
    >{{ getDataItem('forecast_24h[1].max_degree') }}℃</span>
    <span class="text-spacer" />
    <span
      class="wind"
      style="padding:0 2px;"
    >{{
      getDataItem('forecast_24h[1].night_wind_direction')
    }} {{ getDataItem('forecast_24h[1].night_wind_power') }}级</span> -->
    <span>
      星期{{ weekDay }}
    </span>
  </div>
</template>

<script>
import _ from 'lodash';

export default {
  props: {
    mozuInfo: {
      type: Object,
      default() {
        return {};
      },
    },
  },
  data() {
    return {
      info: {
        forecast_24h: [{}, {
          day_weather: '',
          min_degree: '',
          max_degree: '',
          night_wind_direction: '',
          night_wind_power: '',
        }],
      },
      weekDay: ['天', '一', '二', '三', '四', '五', '六'][new Date().getDay()],
    };
  },
  mounted() {
    // this.loadData();
  },
  methods: {
    refresh() {
      const newWeekDay = ['天', '一', '二', '三', '四', '五', '六'][new Date().getDay()];
      if (this.weekDay === newWeekDay) return;
      this.weekDay = newWeekDay;
    },
    async loadData() {
      const data = JSON.parse(await this.$axios.post('/cgi/personaldesktop/dashboard/getWeatherInfo', {
        parkId: this.mozuInfo.parkId || 326,
      }));
      this.info = data.data;
      window.info = this.info;
    },
    getDataItem(key) {
      return _.get(this.info, key);
    },
  },
};
</script>

<style lang="scss" scoped>
.location-weather {
  font-size: 12px;
  line-height: 24px;
  color: #fff;
}

.city {
  display: inline-block;
}

.current-weatcher {
  display: inline-block;
}

.temp {
  font-weight: 700;
  color: #ffc107;
}

.text-spacer {
  padding-left: 4px;
}
</style>
