<template>
  <fragment>
    <el-table-column
      prop="group_name"
      label="时间组名称"
      width="120"
    />

    <el-table-column
      prop="timezone"
      label="时间段"
    >
      <template #default="{ row }">
        {{ formatTimeRanges(row.timezone) }}
      </template>
    </el-table-column>

    <el-table-column
      prop="week"
      label="生效日期"
    >
      <template #default="{ row }">
        {{ formatDays(row.week) }}
      </template>
    </el-table-column>
  </fragment>
</template>

<script>
import { weekDaysMap } from './const';

export default {
  methods: {
    formatTimeRange(range) {
      return `${range.begin}-${range.end}`;
    },
    formatTimeRanges(arr) {
      return _.chain(arr)
        .map(this.formatTimeRange)
        .join('，')
        .value();
    },
    formatDays(arr) {
      return _.chain(arr)
        .map(item => weekDaysMap[item])
        .join('，')
        .value();
    },
  },
};
</script>