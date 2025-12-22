<template>
  <div class="status-bar">
    <span class="status-bar-time">{{ time }}</span>

    <!-- <img
      :src="statuMap['net1-status'] === '0'?net1on:net1off"
    >
    <img
      :src="statuMap['net2-status'] === '0'?net2on:net2off"
    > -->
  </div>
</template>

<script>
import { getNowTime } from './utils/business.js';
import net1on from './images/windows-network.png';
import net1off from './images/windows-no-network.png';
import net2on from './images/xinhao.svg';
import net2off from './images/wuxinhao2.svg';

export default {
  data() {
    return {
      statuMap: {
        'net1-status': '0',
        'net2-status': '0',
      },
      time: '2020-03-05 : asdas',
      safeDays: 999,
      net1on,
      net1off,
      net2on,
      net2off,

      // runningStatusWatcher: new RunningStatusWatcher(
      //   10000,
      // ).withDiffPlugin(),
    };
  },

  mounted() {
    // this.watchStatsus();
    this.initTime();
  },

  beforeDestroy() {
    // this.runningStatusWatcher.cancel();
  },

  methods: {
    initTime() {
      this.time = getNowTime();
      setTimeout(() => {
        this.initTime();
      }, 1000);
    },
    // 链接状态

    // watchStatsus() {
    //   this.runningStatusWatcher.watch(null, (data) => {
    //     this.safeDays = Math.floor(Number(data['AlarmHA-runTime-runTime'].value / (3600 * 24)) + 1);
    //     // eslint-disable-next-line no-restricted-syntax
    //     for (const key in data) {
    //       if (Object.hasOwnProperty.call(data, key)) {
    //         if (key.indexOf('net-1') !== -1) {
    //           // 内网
    //           this.statuMap['net1-status'] = data[key].status;
    //         }
    //         if (key.indexOf('net-2') !== -1) {
    //           // 4G
    //           this.statuMap['net2-status'] = data[key].status;
    //         }
    //         if (key.indexOf('runTime') !== -1) {
    //           // 4G
    //           if (data[key].value) {
    //             this.statuMap.runTime = Math.ceil(Number(data[key].value) / (24 * 60 * 60));
    //           } else {
    //             this.statuMap.runTime = 235;
    //           }
    //         }
    //       }
    //     }
    //   });
    // },
  },
};
</script>

<style lang="scss" scoped>
  @mixin lib-vw($prop, $value) {
    #{$prop}: $value / (1920px * 0.01) + vw;
  }

  .status-bar {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    & > img {
        @include lib-vw(width, 24px);
        margin-right: 5px;
    }
    &-time{
      position: relative;
      top: -2px;
      margin-right: 10px;
      @include lib-vw(font-size, 16px);
      vertical-align: middle;
  }
  }
</style>
