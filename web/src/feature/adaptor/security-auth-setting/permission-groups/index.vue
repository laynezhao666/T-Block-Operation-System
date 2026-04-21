<template>
  <tedge-table-layout
    :context="tableLayoutContext"
  >
    <template #columns>
      <el-table-column
        prop="name"
        label="权限组名称"
        width="120"
      />

      <el-table-column
        prop="label"
        label="类型"
        width="100"
      />

      <el-table-column
        prop="doors"
        label="授权门范围"
        show-overflow-tooltip
      >
        <template #default="{ row }">
          {{ joinArray(mapArr(row.doors, 'name')) }}
        </template>
      </el-table-column>

      <el-table-column
        prop="time_group"
        label="时间组"
        show-overflow-tooltip
        width="120"
      >
        <template #default="{ row }">
          {{ row.time_group.group_name }}
        </template>
      </el-table-column>

      <el-table-column
        prop="cards"
        label="人员/卡号"
        show-overflow-tooltip
      >
        <template #default="{ row }">
          {{ formatCardCell(row.cards) }}
        </template>
      </el-table-column>

      <el-table-column
        prop="comment"
        label="备注"
        show-overflow-tooltip
      />
    </template>
  </tedge-table-layout>
</template>

<script>
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';
import BasicInfoForm from './basic-info-form.vue';
import DoorsSelectForm from './doors-select-form.vue';
import TimeGroupSelectForm from './time-group-select-form.vue';
import PersonListForm from './person-list-form.vue';
import { defaultResolveOffsetLimitOfPagination } from '../../../../utils/pagination';
import { axiosDelete, axiosPut } from '../../../../utils/axios-methods';

export default {
  components: {
    TedgeTableLayout,
  },
  data() {
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
        .pagination()
        .search({
          placeholder: '请输入关键字搜索',
          isHide: true,
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
        .baseCurd({
          rowEditColumnWidth: 200,
          add: {
            adminRight: true,
            action() {
              return {
                groupType: '公共区域',
                doors: [],
                cards: [],
              };
            },
          },
          remove: {
            adminRight: true,
            confirm: row => this.removeRow(row),
            batchRemove: rows => this.batchRemoveRows(rows),
          },
          edit: {
            adminRight: true,
          },
          rowOprsComponents: [
            () => import('./relative-info-modal.vue'),
          ],
        })
        .curdFormModal({
          title: '权限组',
          width: 900,
          steps: [{
            title: '基本信息',
            comp: BasicInfoForm,
          }, {
            title: '授权门范围',
            comp: DoorsSelectForm,
          }, {
            title: '设置时间组',
            comp: TimeGroupSelectForm,
          }, {
            title: '授权人员',
            comp: PersonListForm,
          }],
          beforeEdit: (editting, replace) => {
            const newEditting = {
              ...editting,
              doors: _.map(editting.doors, door => `door-${door.id}`),
              cards: _.map(editting.cards, card => card.card_no),
            };
            delete newEditting.door;
            replace(newEditting);
          },
          onSubmit: this.submitEditForm.bind(this),
        })
        .selection({
          identity: 'id',
          oprs: ['export', 'delete'],
        })
        .done();
    },
    async fetchData(filters, search, pagination) {
      return this.$axios.get('/api/dcos/tdac-cgi/access-groups', {
        ...defaultResolveOffsetLimitOfPagination(pagination),
      });
    },
    joinArray(arr) {
      return arr.join('，');
    },
    mapArr(arr, iterator) {
      return _.map(arr, iterator);
    },
    formatCardCell(cards) {
      return _.chain(cards)
        .map(item => item.staff?.name || item.name || item.card_no)
        .union()
        .join('，')
        .value();
    },
    removeRow(row) {
      return axiosDelete(`/api/dcos/tdac-cgi/access-group/${row.id}`);
    },
    batchRemoveRows(rows) {
      console.log('batchRemoveRows', rows);
    },
    normalizeFormDataForPost(data) {
      return {
        ...data,
        // cards: undefined,
        card: undefined,
        time_group: undefined,
        // doors: undefined,
        doors: _.chain(data.doors)
          .filter(item => item.indexOf('door-') === 0)
          .map(item => Number(item.replace('door-', '')))
          .value(),
        cards: data.cards,
        time_group_no: data.time_group_no,
      };
    },
    async submitEditForm(row, isCreate) {
      const normalizeForPost = this.normalizeFormDataForPost(row);
      if (isCreate) {
        await this.$axios.post('/api/dcos/tdac-cgi/access-groups', normalizeForPost);
      } else {
        await axiosPut(`/api/dcos/tdac-cgi/access-group/${normalizeForPost.id}`, normalizeForPost);
      }
      return true;
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
