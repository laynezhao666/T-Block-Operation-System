<template>
  <tedge-table-layout
    :context="tableLayoutContext"
  >
    <template #columns>
      <el-table-column
        prop="id"
        label="人员编号"
        width="120"
      />

      <el-table-column
        prop="name"
        label="姓名"
        width="100"
      />

      <el-table-column
        prop="card_no"
        label="领卡数量"
        width="100"
      >
        <template #default="{ row }">
          <el-tooltip
            :disabled="!row.card_no || !row.card_no.length"
          >
            <el-button type="text">
              {{ row.card_no ? row.card_no.length : 0 }}
            </el-button>

            <div
              slot="content"
              class="card-id-list"
            >
              <div
                v-for="(cardId, i) in row.card_no"
                :key="i"
                class="card-id"
              >
                {{ cardId }}
              </div>
            </div>
          </el-tooltip>
        </template>
      </el-table-column>

      <el-table-column
        prop="paper"
        label="证件号"
        width="172"
        show-overflow-tooltip
      />

      <el-table-column
        prop="phone"
        label="手机号"
        width="130"
      />

      <el-table-column
        prop="email"
        label="邮箱"
        width="120"
      />

      <el-table-column
        prop="company"
        label="人员组"
        width="120"
      />

      <el-table-column
        prop="comment"
        label="备注"
      />
    </template>

    <template
      slot="outer-modals"
    >
      <card-list-belong-user
        ref="cardListBelongUser"
      />
    </template>
  </tedge-table-layout>
</template>

<script>
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';
import { curryingRenderElTextButton } from '../../../component/tedge-components/table-layout-context/render-el-text-button';
import CardListBelongUser from './card-list-belong-user.vue';
import { defaultResolveOffsetLimitOfPagination } from '../../../../utils/pagination';
import { axiosDelete, axiosPut, axiosUploadFile } from '../../../../utils/axios-methods';
import AdminLimitTooltips from '../../../component/tedge-components/admin-limit-tooltips.vue';


export default {
  components: {
    TedgeTableLayout,
    CardListBelongUser,
    AdminLimitTooltips,
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
          placeholder: '按姓名、手机、邮箱、备注搜索',
        })
        .filters({
          company: '',
        })
        .remoteFilterPagination()
        .indexColumn({
          label: '序号',
        })
        .baseCurd({
          add: {
            adminRight: true,
            action() {
              return {
                paper_type: '身份证',
              };
            },
          },
          edit: {
            adminRight: true,
          },
          remove: {
            adminRight: true,
            label: '移除',
            confirm: (row) => {
              this.removeRow(row);
            },
            batchRemove: (rows) => {
              this.batchRemoveRows(rows);
            },
          },
          rowOprsComponents: [
            curryingRenderElTextButton({
              label: '领卡信息',
              onClick: ({ row }) => {
                this.showCardInfoOfUser(row);
              },
            }),
          ],
          rowEditColumnWidth: 180,
        })
        .curdFormModal({
          title: '人员',
          formComp: () => import('./form.vue'),
          beforeEdit: (editting, replace) => replace({
            ...editting,
            oldPaper: editting.paper,
            password: '******',
            passwordConfirm: '******',
          }),
          onSubmit: this.submitEditForm.bind(this),
        })
        .selection({
          identity: 'id',
          oprs: ['export', 'delete'],
        })
        .toolbarActions({
          icon: 'tn-icon-import',
          text: '导入人员信息',
          action: () => {
            this.handleImportStaffs();
          },
        })
        .toolbarActions({
          icon: 'tn-icon-export',
          text: '导出人员信息',
          action: () => {
            this.handleExportStaffs();
          },
        })
        .done();
    },
    async fetchData(filters, search, pagination) {
      const url = '/api/dcos/tdac-cgi/staffs';
      return this.$axios.get(url, {
        ...defaultResolveOffsetLimitOfPagination(pagination),
        query: search,
        company: filters.company,
      });
    },
    joinArray(arr) {
      return arr.join('，');
    },
    async removeRow(row) {
      await axiosDelete(`/api/dcos/tdac-cgi/staff/${row.id}`);
    },
    batchRemoveRows(rows) {
      console.log('batchRemoveRows', rows);
    },
    async submitEditForm(row, isCreate) {
      const rowToPost = {
        ...row,
        paper: row.paper === row.oldPaper ? undefined : row.paper,
        oldPaper: undefined,
        password: row.password === '******' ? undefined : row.password,
        passwordConfirm: undefined,
      };

      if (isCreate) {
        await this.$axios.post('/api/dcos/tdac-cgi/staffs', rowToPost);
      } else {
        await axiosPut(`/api/dcos/tdac-cgi/staff/${rowToPost.id}`, rowToPost);
      }

      return true;
    },
    showCardInfoOfUser(user) {
      this.$refs.cardListBelongUser.show(user, () => {
        this.tableLayoutContext.loadData();
      });
    },
    async handleImportStaffs() {
      try {
        await axiosUploadFile('/api/dcos/tdac-cgi/staffs/import', {
          file: axiosUploadFile.fileSelectSymbol,
        });
        this.tableLayoutContext.loadData();
        this.$message.success('导入人员信息完成');
      } catch (error) {
        console.error('导入人员信息失败:', error);
      }
    },
    handleExportStaffs() {
      this.$axios.download('/api/dcos/tdac-cgi/staffs/export', {}, true, {
        fileName: '人员信息.xlsx',
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.card-id {
  &:not(:first-child) {
    margin-top: 4px;
  }
}
</style>
