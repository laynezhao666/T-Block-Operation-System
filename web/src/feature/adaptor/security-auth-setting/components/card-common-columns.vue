<template>
  <div>
    <el-table-column
      prop="card_no"
      label="卡号"
      width="130"
    />

    <el-table-column
      prop="card_type"
      label="卡类型"
      width="100"
    >
      <template #default="{ row }">
        {{ row.card_type === 0 ? '长期卡' : '临时卡' }}
      </template>
    </el-table-column>

    <el-table-column
      prop="card_flag"
      label="状态"
      width="80"
      class-name="column-cell-center"
    >
      <template #default="{ row }">
        <el-button
          v-if="row.card_flag === 0"
          type="text"
          icon="tn-icon-checkbox-checked"
          class="status-success"
        />
        <el-button
          v-else
          type="text"
          class="status-disabled"
        >
          禁
        </el-button>
      </template>
    </el-table-column>

    <el-table-column
      prop="access_groups"
      label="权限组"
      show-overflow-tooltip
    >
      <template #default="{ row }">
        {{ joinArray(row.access_groups, 'name') }}
      </template>
    </el-table-column>

    <el-table-column
      prop="access_groups"
      label="授权门范围"
      show-overflow-tooltip
    >
      <template #default="{ row }">
        {{ formatDoors(row) }}
      </template>
    </el-table-column>

    <el-table-column
      prop="card_info"
      label="有效期"
      width="180"
    >
      <template #default="{ row }">
        {{ row.card_type === 0 ? '永久' : formatTimeSeconds(row.valid_time) }}
      </template>
    </el-table-column>
  </div>
</template>

<script>
import _ from 'lodash';
import dayjs from 'dayjs';

export default {
  methods: {
    joinArray(arr, fieldsPath) {
      return (fieldsPath ? _.map(arr, fieldsPath) : arr).join('，');
    },
    formatDoors(row) {
      return _.chain(row.access_groups)
        .map('doors')
        .flatten()
        .map('name')
        .union()
        .join('，')
        .value();
    },
    formatTimeSeconds(validTime) {
      return dayjs(validTime * 1000).format('YYYY-MM-DD HH:mm:ss');
    },
  },
};
</script>

<style lang="scss" scoped>
.status-success {
  color: var(--tn-color-success);

  /deep/ {
    & > .tn-icon-checkbox-checked {
      font-size: 32px;
    }
  }
}

.status-disabled {
  color: #ffffff;
  background-color: var(--tn-color-danger);
  width: 24px;
  height: 24px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: 500;
}
</style>
