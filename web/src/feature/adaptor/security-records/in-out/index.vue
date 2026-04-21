<template>
  <tedge-table-layout
    :context="tableLayoutContext"
  >
    <template #toolbar-extra="{ filters }">
      <div style="display: flex; align-items: center; gap: 10px;">
        <day-range-radio-group
          v-model="filters.days"
        />
        <el-input
          v-if="shouldShowDoorInput"
          v-model="filters.doorName"
          placeholder="输入门名称进行查询"
          style="width: 200px;"
          clearable
        />
      </div>
    </template>

    <template #columns>
      <el-table-column
        prop="time"
        label="时间"
        width="180"
      />

      <el-table-column
        prop="desc"
        label="事件描述"
      />

      <el-table-column
        prop="direction"
        label="进出方向"
        width="100"
      >
        <template
          #header
        >
          <span>进出方向</span>
          <!-- 接口该参数延期 -->
          <drop-down-checkboxes
            v-if="false"
            v-model="tableLayoutContext.filters.directions"
            :options="directionOptions"
            :filterable="false"
          />
        </template>

        <template #default="{ row }">
          <el-tag
            :type="row.direction === '进门' ? 'success' : 'warning'"
            effect="light"
          >
            {{ row.direction }}
          </el-tag>
        </template>
      </el-table-column>

      <door-name-column
        :filters="tableLayoutContext.filters"
        width="auto"
      />

      <control-column
        :filters="tableLayoutContext.filters"
        width="auto"
      />

      <el-table-column
        prop="card_number"
        label="卡号"
        width="180"
      />

      <el-table-column
        prop="username"
        label="名称"
        width="120"
      >
        <template
          #header
        >
          <span>名称</span>
          <!-- 接口该参数延期 -->
          <drop-down-checkboxes
            v-if="false"
            v-model="tableLayoutContext.filters.username"
            :options="personNameOptions"
          />
        </template>
      </el-table-column>

      <el-table-column
        prop="company"
        label="单位"
        width="120"
      />
    </template>
  </tedge-table-layout>
</template>

<script>
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import DayRangeRadioGroup from '../../../component/tedge-components/day-range-radio-group.vue';
import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';
import DropDownCheckboxes from '../../../component/tedge-components/drop-down-checkboxes.vue';
import DoorNameColumn from '../components/door-name-column.vue';
import ControlColumn from '../components/control-column.vue';
import { defaultResolveOffsetLimitOfPagination } from '../../../../utils/pagination';

export default {
  computed: {
    shouldShowDoorInput() {
      // 显示条件：controllers未初始化或空数组
      return !this.controllers || this.controllers.length === 0
    },
    
    controllers() {
      return _.chain(this.doors)
        .groupBy('controller_id')
        .map((doors, controllerId) => ({
          id: Number(controllerId),
          doors: _.map(doors, 'number'),
        }))
        .value()
    }
  },
  components: {
    TedgeTableLayout,
    DayRangeRadioGroup,
    DropDownCheckboxes,
    DoorNameColumn,
    ControlColumn,
  },
  props: {
    doors: {
      type: Array,
      default() {
        return null;
      },
    },
  },
  data() {
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
      // personNameOptions: [{
      //   label: '人1',
      //   value: '人1',
      // }, {
      //   label: '人2',
      //   value: '人2',
      // }],

      lastFilters: null,
    };
  },
  // watch: {
  //   filters: {
  //     deep: true,
  //     handler() {
  //       // this.tableLayoutContext.forceReloadData();
  //       this.tableLayoutContext.loadData();
  //     },
  //   },
  // },
  methods: {
    createTableLayoutContext() {
      return chainTableLayout(this.fetchData.bind(this))
        .pagination()
        .tableStyle({
          height: this.doors?.length
            ? (window.innerHeight - 224)
            : undefined,
        })
        .toolbarActions({
          text: '导出',
          icon: 'tn-icon-download',
          action: () => {
            this.exportData();
          },
        })
        .search({
          placeholder: '请输入用户名或卡号进行搜索',
        })
        .filters({
          days: null,
          doorName: '',
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
    async fetchData(filters, search, pagination) {
      if (!filters.days) {
        return {
          total: 0,
          list: [],
        };
      }

      this.lastFilters = _.cloneDeep(filters);

      const payload = {
        door_name: filters.doorName,
        // keyword: search.trim(),
        // ...filters,
        begin_time: Math.round(filters.days[0].getTime() / 1000),
        end_time: Math.round(filters.days[1].getTime() / 1000),

        query: search,

        controllers: _.chain(this.doors)
          .groupBy('controller_id')
          .map((doors, controllerId) => ({
            id: Number(controllerId),
            doors: _.map(doors, 'number'),
          }))
          .value(),

        ...defaultResolveOffsetLimitOfPagination(pagination),
      };
      const url = this.doors?.length ? '/api/dcos/tdac-cgi/doors/events' : '/api/dcos/tdac-cgi/events';
      return this.$axios.post(url, payload);
    },
    exportData() {
      this.$axios.download('/api/dcos/tdac-cgi/events/export', {
        begin_time: Math.round(this.lastFilters.days[0].getTime() / 1000),
        end_time: Math.round(this.lastFilters.days[1].getTime() / 1000),

        controllers: _.chain(this.doors)
          .groupBy('controller_id')
          .map((doors, controllerId) => ({
            id: Number(controllerId),
            doors: _.map(doors, 'number'),
          }))
          .value(),
        
        door_name: this.lastFilters.doorName || '',

        query: this.tableLayoutContext.search.value || '',

      }, true, {
        fileName: '进出记录.xlsx',
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
