<template>
  <div class="day-range-radio-group">
    <el-radio-group
      v-model="selectedDaysKey"
      size="small"
      variant="default-filled"
    >
      <el-radio-button
        v-for="(opt, i) in options"
        :key="i"
        :label="opt.value"
      >
        {{ opt.label }}
      </el-radio-button>

      <el-radio-button label="custom">
        自定义
      </el-radio-button>
    </el-radio-group>

    <el-date-picker
      v-show="selectedDaysKey === 'custom'"
      :value="value"
      type="datetimerange"
      range-separator="至"
      start-placeholder="开始日期"
      end-placeholder="结束日期"
      border-type="bordered"
      size="small"
      class="datetime-range-picker"
      @input="chanegValue"
    />
  </div>
</template>

<script>
import dayjs from 'dayjs';
import { listenInVue } from './date-change-callback.ts';

const generateOptions = () => [{
  label: '今日',
  value: 'today',
  dateRange: [
    dayjs().startOf('days')
      .toDate(),
    dayjs().add(1, 'd')
      .startOf('days')
      .toDate(),
  ],
}, ...[
  3, 7, 30,
].map(days => ({
  label: `近${days}日`,
  value: `${days}D`,
  dateRange: [
    dayjs().add(-days, 'd')
      .toDate(),
    dayjs().toDate(),
  ],
}))];

export default {
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Array,
      default() {
        return null;
      },
    },
  },
  data() {
    window.dr = this;
    return {
      options: generateOptions(),
      isForceCustom: false,
    };
  },
  computed: {
    selectedDaysKey: {
      get() {
        if (this.isForceCustom) return 'custom';

        const { value } = this;
        if (value?.length !== 2) return null;

        const matchOpt = this.getOptionByDates(value);

        return matchOpt?.value || 'custom';
      },
      set(optValue) {
        if (optValue === 'custom') {
          this.isForceCustom = true;
          return;
        }

        this.isForceCustom = false;
        this.$emit('change', this.getOptionByOptValue(optValue).dateRange);
      },
    },
  },
  watch: {
    value: {
      immediate: true,
      handler(value) {
        if (value?.length) return;

        this.$emit('change', this.options[0].dateRange);
      },
    },
  },
  created() {
    listenInVue(this, () => {
      this.options = generateOptions();
    });
  },
  methods: {
    getOptionByDates(dates) {
      return this.options.find(opt => opt.dateRange[0].getTime() === dates[0].getTime()
          && opt.dateRange[0].getTime() === dates[0].getTime());
    },
    getOptionByOptValue(optValue) {
      return this.options.find(opt => opt.value === optValue);
    },
    chanegValue(times) {
      this.$emit('change', times);
    },
  },
};
</script>

<style lang="scss" scoped>
.day-range-radio-group {
  display: inline-block;

  & > * {
    display: inline-block;
  }
}

.datetime-range-picker {
  width: 358px;
  display: inline-block;
  border: 1px solid silver;
  padding: 0 4px;
  line-height: 24px;
  height: 28px;
  box-sizing: border-box;

  position: relative;
  top: 4px;
}
</style>
