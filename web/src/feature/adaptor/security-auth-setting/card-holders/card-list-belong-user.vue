<template>
  <el-dialog
    title="领卡信息"
    width="80%"
    :visible.sync="visible"
  >
    <div class="alert-text">
      该用户名下有{{ cardIds.length }}张卡
    </div>

    <tedge-table-layout
      :context="tableLayoutContext"
    >
      <template #columns>
        <card-common-columns />
      </template>
    </tedge-table-layout>
  </el-dialog>
</template>

<script>
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';
import { curryingRenderElTextButton } from '../../../component/tedge-components/table-layout-context/render-el-text-button';
import CardCommonColumns from '../components/card-common-columns.vue';
import { axiosDelete, axiosPut } from '../../../../utils/axios-methods';

export default {
  components: {
    TedgeTableLayout,
    CardCommonColumns,
  },
  data() {
    return {
      owner: null,
      cardIds: [],
      callback: null,
      tableLayoutContext: this.createTableLayoutContext(),
    };
  },
  computed: {
    visible: {
      get() {
        return !!this.owner;
      },
      set(visible) {
        if (!visible) {
          // eslint-disable-next-line no-unused-expressions, babel/no-unused-expressions
          this.callback?.();

          this.owner = null;
          this.callback = null;
        }
      },
    },
  },
  methods: {
    show(owner, callback) {
      this.owner = owner;
      this.cardIds = owner.card_no ? [...owner.card_no] : [];
      this.callback = callback;
      // this.reloadTable();
    },
    reloadTable() {
      // this.tableLayoutContext.forceReloadData();
      this.tableLayoutContext.loadData();
    },
    async fetchData() {
      if (!this.cardIds?.length) return [];

      // TODO: 对接接口
      const resp = await this.$axios.post('/api/dcos/tdac-cgi/cards', {
        offset: 0,
        limit: 10000,
        cards: this.cardIds,
      });
      return resp.list;
    },
    createTableLayoutContext() {
      return chainTableLayout(this.fetchData.bind(this))
        .tableStyle({
          stripe: true,
        })
        .hideToolbar()
        .indexColumn({
          label: '序号',
        })
        .baseCurd({
          edit: false,
          remove: {
            adminRight: true,
            confirm: this.removeCard.bind(this),
          },
          rowOprsComponents: [
            curryingRenderElTextButton({
              label: '解绑',
              adminRight: true,
              confirm: {
                title: '确认解绑吗？',
              },
              onClick: ({ row }) => this.unbindCard(row),
            }),
          ],
        })
        .done();
    },
    async removeCard(row) {
      await axiosDelete(`/api/dcos/tdac-cgi/card/${row.card_no}`);
      this.$message.success('删除成功');
    },
    async unbindCard(row) {
      await axiosPut('/api/dcos/tdac-cgi/card/unbind', {
        card: row.card_no,
      });
      this.cardIds = _.filter(this.cardIds, id => row.card_no !== id);
      this.reloadTable();
    },
  },
};
</script>

<style lang="scss" scoped>
.alert-text {
  font-size: 12px;
  color: #999;
}
</style>
