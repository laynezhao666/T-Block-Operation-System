<template>
  <tedge-table-layout
    :context="tableLayoutContext"
  >
    <template #columns>
      <time-group-columns />
    </template>

    <template
      #outer-modals
    >
      <sync-time-groups
        ref="syncTimeGroups"
      />
    </template>
  </tedge-table-layout>
</template>

<script>
import dayjs from 'dayjs';
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';
import { curryingRenderElTextButton } from '../../../component/tedge-components/table-layout-context/render-el-text-button';
import { axiosPut } from '../../../../utils/axios-methods';
import SyncTimeGroups from './sync-time-groups.vue';
import { defaultResolveOffsetLimitOfPagination } from '../../../../utils/pagination';
import getEdgeRequest from 'feature/utils/request';
import TimeGroupColumns from './time-group-columns.vue';

export default {
  components: {
    TedgeTableLayout,
    SyncTimeGroups,
    TimeGroupColumns,
  },
  data() {
    window.tp = this;
    return {
      tableLayoutContext: this.createTableLayoutContext(),
    };
  },
  methods: {
    createTableLayoutContext() {
      return chainTableLayout(this.fetchData.bind(this))
        .tableStyle({
          stripe: true,
        })
        .pagination({
          size: 20,
        })
        .localFilterPagination()
        .search({
          placeholder: '输入关键词搜索',
          isHide: true,
        })
        .indexColumn({
          label: '序号',
        })
        .baseCurd({
          add: {
            adminRight: true,
            disabled(rows) {
              return rows.length >= 12 ? '最多支持12个时间组' : false;
            },
          },
          edit: {
            adminRight: true,
          },
          remove: null,
        })
        .curdFormModal({
          title: '时间组',
          formComp: () => import('./form.vue'),
          onSubmit: this.submitEditForm.bind(this),
        })
        .selection({
          identity: 'group_no',
          oprs: ['export', 'delete'],
        })
        .extraBtn(curryingRenderElTextButton({
          adminRight: true,
          label: '同步时间组',
          btnProps: {
            type: 'default',
          },
          confirm: {
            title: '请确认要同步时间组吗?',
          },
          onClick: this.syncTimeGroups.bind(this),
        }))
        .done();
    },
    async fetchData(filters, search, pagination) {
      const url = '/api/dcos/tdac-cgi/time-groups';
      const list = await getEdgeRequest(this.$axios).get(url, {
        ...defaultResolveOffsetLimitOfPagination(pagination),
      });

      return list;
    },
    joinArray(arr) {
      return arr.join('，');
    },
    async submitEditForm(row) {
      const url = `/api/dcos/tdac-cgi/time-group/${row.group_no}`;

      await axiosPut(url, {
        enable: 1,
        ...row,
        week: _.sortBy(row.week, _.identity),
        timezone: _.map(row.timezone, touple => ({
          begin: touple[0] instanceof Date ? dayjs(touple[0]).format('HH:mm') : touple[0],
          end: touple[1] instanceof Date ? dayjs(touple[1]).format('HH:mm') : touple[1],
        })),
      });
      return true;
    },
    async syncTimeGroups() {
      this.$refs.syncTimeGroups.startSync();
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
