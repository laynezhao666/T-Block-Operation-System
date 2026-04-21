<template>
  <tedge-table-layout
    :context="tableLayoutContext"
  >
    <template #toolbar-extra="{ filters }">
      <day-range-radio-group
        v-model="filters.days"
      />
    </template>

    <template #columns>
      <el-table-column
        prop="time"
        label="时间"
        width="180"
      />

      <el-table-column
        prop="state_desc"
        label="状态"
        width="180"
      />

      <el-table-column
        prop="type"
        label="类型"
        width="180"
      >
        <template slot-scope="{ row }">
          {{ getTypeName(row) }}
        </template>
      </el-table-column>

      <el-table-column
        prop="desc"
        label="告警描述"
      />

      <el-table-column
        prop="door_name"
        label="门名称"
        width="180"
      />

      <el-table-column
        prop="controller_name"
        label="门禁控制器"
        width="180"
      />
    </template>
  </tedge-table-layout>
</template>

<script>
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import DayRangeRadioGroup from '../../../component/tedge-components/day-range-radio-group.vue';
import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';
import { defaultResolveOffsetLimitOfPagination } from '../../../../utils/pagination';

const typeMap = {
  0: '门开超时',
  1: '异常开门',
  2: '发现火警',
  3: '机箱被打开',
};

export default {
  components: {
    TedgeTableLayout,
    DayRangeRadioGroup,
  },
  data() {
    window.io = this;
    return {
      tableLayoutContext: this.createTableLayoutContext(),
      directionOptions: [{
        label: '进门',
        value: '进门',
      }, {
        label: '出门',
        value: '出门',
      }],
      doorNameOptions: [{
        label: '门1',
        value: '门1',
      }, {
        label: '门2',
        value: '门2',
      }],
      controls: [{
        label: '控制器1',
        value: '控制器1',
      }, {
        label: '控制器2',
        value: '控制器2',
      }],
      personNameOptions: [{
        label: '人1',
        value: '人1',
      }, {
        label: '人2',
        value: '人2',
      }],

      lastFilters: null,
    };
  },
  methods: {
    createTableLayoutContext() {
      return chainTableLayout(this.fetchData.bind(this))
        .pagination()
        .search({
          placeholder: '请输入您要检索的告警内容',
          isHide: true,
        })
        .toolbarActions({
          text: '导出',
          icon: 'tn-icon-download',
          action: () => {
            this.exportData();
          },
        })
        .filters({
          days: null,
          doorName: [],
          directions: [],
          controls: [],
          personName: [],
        })
        .remoteFilterPagination()
        .indexColumn({
          label: '序号',
        })
        .done();
    },
    getTypeName(row) {
      return typeMap[row.type] || '未知';
    },
    async fetchData(filters, search, pagination) {
      if (!filters.days) {
        return {
          total: 0,
          list: [],
        };
      }

      this.lastFilters = _.cloneDeep(filters);

      const payload = {
        // doorIds: this.doorIds,
        keyword: search.trim(),
        // ...filters,
        begin_time: Math.round(filters.days[0].getTime() / 1000),
        end_time: Math.round(filters.days[1].getTime() / 1000),
        controller_ids: [],
        ...defaultResolveOffsetLimitOfPagination(pagination),
      };
      return this.$axios.post('/api/dcos/tdac-cgi/alarms', payload);
    },
    exportData() {
      this.$axios.download('/api/dcos/tdac-cgi/alarms/export', {
        begin_time: Math.round(this.lastFilters.days[0].getTime() / 1000),
        end_time: Math.round(this.lastFilters.days[1].getTime() / 1000),
      }, true, {
        fileName: '门禁告警记录.xlsx',
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.tn-icon-filter {
  font-size: 16px;
  position: relative;
  top: 2px;
  color: #a0a0a0;
  transition: 0.3s color;
  cursor: pointer;

  &.active {
    color: var(--tn-color-primary);
  }
}
</style>
