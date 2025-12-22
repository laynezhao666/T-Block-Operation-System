<template>
  <div class="tab shadow-base">
    <ul>
      <template v-for="(item, i) in tabList">
        <li
          v-if="item.n_showtype"
          :key="i"
          :class="activeIndex === item.n_href ? 'active' : ''"
          @click="switchTabHandler(item)"
        >
          <div class="tab-item-warp">
            <!-- <i
              v-if="item.n_licls"
              :class="item.n_licls"
            /> -->
            <tn-icon
              v-if="item.n_licls"
              :icon="item.n_licls"
            />
            <template v-if="item.n_name === '告警'">
              <el-badge
                :value="totalAlarm"
                :max="99"
                class="item"
              >
                <span class="tab-item-text tab-item-text-alarm">{{ item.n_name }}</span>
              </el-badge>
            </template>
            <span
              v-else
              class="tab-item-text"
            >
              {{ item.n_name }}
            </span>
          </div>
        </li>
      </template>
    </ul>
  </div>
</template>
<script>

export default {
  name: 'SwitchTabs',
  props: {
    tabList: {
      type: Array,
      default() {
        return [];
      },
    },
    activeIndex: {
      type: String,
      default: '',
    },
    alarmTotal: {
      type: Number,
      default: 0,
    },
  },
  data() {
    return {
      totalAlarm: 0,
    };
  },
  watch: {
    alarmTotal(val) {
      this.totalAlarm = val;
    },
  },
  methods: {
    switchTabHandler(item) {
      this.$emit('onTabSwitched', item.n_href);
    },
  },
};
</script>

<style lang="scss" scoped>
@import './../style/switch_tabs.scss';
</style>
