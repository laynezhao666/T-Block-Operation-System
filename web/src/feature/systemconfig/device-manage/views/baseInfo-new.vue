
<template>
  <div>
    <collector-base-info
      v-show="collector && collector.type === 'collector'"
      :device-status="deviceStatus"
      :collector="collectorInfo"
    />
    <device-base-info
      v-show="collector && collector.type === 'device'"
      :collector="deviceInfo"
    />
  </div>
</template>

<script>
import DeviceBaseInfo from '../components/deviceBaseInfo';
// import BaseInfo3D from './baseInfo-3D';
import CollectorBaseInfo from '../components/collectorBaseInfo';
export default {
  components: {
    DeviceBaseInfo,
    CollectorBaseInfo,
    // BaseInfo3D,
  },
  props: {
    collector: {
      type: Object,
      default: () => null,
    },
    deviceStatus: {
      type: Object,
      default: () => null,
    },
  },

  data() {
    return {};
  },
  computed: {
    deviceInfo() {
      if (this.collector && this.collector.type === 'device') {
        return this.collector;
      }
      return null;
    },
    collectorInfo() {
      if (this.collector && this.collector.type === 'collector') {
        return this.collector;
      }
      return null;
    },
  },
  mounted() {},
  methods: {},
};
</script>
