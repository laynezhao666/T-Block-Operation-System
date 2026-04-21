<template>
  <tedge-table-layout
    :context="tableLayoutContext"
  >
    <template #columns>
      <card-common-columns />

      <el-table-column
        prop="valid_time"
        label="有效期"
        width="170"
      >
        <template #default="{ row }">
          {{ row.card_type === 0 ? '永久' : formatDatetime(row.valid_time) }}
        </template>
      </el-table-column>

      <el-table-column
        prop="staff.name"
        label="领卡人"
        width="100"
      />

      <el-table-column
        prop="staff.company"
        label="单位"
        width="100"
      />
    </template>
  </tedge-table-layout>
</template>

<script>
import _ from 'lodash';
import dayjs from 'dayjs';
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import CardCommonColumns from '../components/card-common-columns.vue';
import { createRemovableTableContext } from '../components/create-removable-table-context';

export default {
  components: {
    TedgeTableLayout,
    CardCommonColumns,
  },
  props: {
    permissionGroup: {
      type: Object,
      required: true,
    },
    updatePermissionGroup: {
      type: Function,
      required: true,
    },
  },
  data() {
    return {
      tableLayoutContext: this.createTableLayoutContext(),
    };
  },
  methods: {
    formatDatetime(time) {
      return dayjs(time * 1000).format('YYYY-MM-DD HH:mm:ss');
    },

    createTableLayoutContext() {
      return createRemovableTableContext(
        this.fetchData.bind(this),
        this.removeRow.bind(this),
      );
    },
    async fetchData() {
      const cardNoList = _.map(this.permissionGroup.cards || [], 'card_no');

      return this.fetchCardsByIds(cardNoList);
    },
    async fetchCardsByIds(cardNoList) {
      if (!cardNoList?.length) return [];

      const resp = await this.$axios.post('/api/dcos/tdac-cgi/cards', {
        offset: 0,
        limit: 100000,
        cards: cardNoList,
      });

      return resp.list;
    },
    async removeRow(row) {
      const permissionGroup = _.cloneDeep(this.permissionGroup);

      permissionGroup.cards = (permissionGroup.cards || []).filter(item => item.id !== row.id);

      await this.updatePermissionGroup(permissionGroup);

      this.tableLayoutContext.forceReloadData();
      this.$message.success('移除成功');
    },
  },
};
</script>
