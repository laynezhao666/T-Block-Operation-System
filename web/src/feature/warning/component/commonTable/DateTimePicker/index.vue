<template>
  <el-date-picker
    v-model="inputValue"
    :size="size"
    :type="type"
    :show-seconds="false"
    range-separator="-"
    start-placeholder="开始日期"
    end-placeholder="结束日期"
    :format="format"
    :value-format="valueFormat"
    prefix-icon="el-icon-caret-bottom"
    :picker-options="dateOpts"
    :clearable="false"
    :default-time="defaultTime"
  />
</template>
<script>
const mapping = {
  datetimerange: {
    format: {
      date: 'yyyy年MM月dd日',
      time: 'HH:mm',
    },
    valueFormat: 'yyyy-MM-dd HH:mm:ss',
  },
  datetime: {
    format: {
      date: 'yyyy年MM月dd日',
      time: 'HH:mm',
    },
    valueFormat: 'yyyy-MM-dd HH:mm:ss',
  },
  daterange: {
    format: 'yyyy年MM月dd日',
    valueFormat: 'yyyy-MM-dd',
  },
  monthrange: {
    format: 'yyyy年MM月',
    valueFormat: 'yyyy-MM',
  },
  date: {
    format: 'yyyy年MM月dd日',
    valueFormat: 'yyyy-MM-dd',
  },
  month: {
    format: 'yyyy年MM月',
    valueFormat: 'yyyy-MM',
  },
};

export default {
  props: {
    type: {
      type: String,
      default: 'date',
    },
    history: {
      type: Boolean,
      default: false,
    },
    value: {},
    size: {
      type: String,
      default: 'large',
    },
    defaultTime: {
      type: Array,
      default: () => ['00:00:00', '23:59:59'],
      validator(value) {
        return value.filter(Boolean).length === 2;
      },
    },
  },
  data() {
    const props = mapping[this.type] || {};
    let dateOpts;

    if (this.history) {
      dateOpts = {
        disabledDate: time => time.getTime() > new Date(),
      };
    }
    return {
      ...props,
      dateOpts,
      inputValue: this.value,
    };
  },
  watch: {
    inputValue(val) {
      this.$emit('input', val);
    },
    value(v) {
      this.inputValue = v;
    },
  },
};
</script>
