<template>
    <tedge-table-layout
      :context="tableLayoutContext"
      class="table"
      :default-sort = "{prop: 'date', order: 'descending'}"
    >
      <template #columns>
        <el-table-column
          prop="id"
          label="消息序号"
          width="100"
          header-align="center"
        />

        <el-table-column
          prop="method"
          label="消息类型"
          width="200"
          header-align="center"
        />

        <el-table-column
          prop="payload"
          label="消息内容"
          width="320"
          header-align="center"
          show-overflow-tooltip
        />

        <el-table-column
          prop="create_time"
          label="创建时间"
          width="180"
          sortable
          header-align="center"
        />

        <el-table-column
          prop="controller_name"
          label="门禁控制器"
          width="370"
          header-align="center"
        />

        <el-table-column
          prop="state"
          label="状态"
          width="105"
          header-align="center"
        >
        <template #default="{ row }">
          <el-tag v-if="row.state === '过期'" type="warning" effect="light">过期</el-tag>
          <el-tag v-if="row.state === '成功'" type="success" effect="light">成功</el-tag>
          <el-tag v-if="row.state === '待执行'" type="primary" effect="light">待执行</el-tag>
          <el-tag v-if="row.state === '失败'" type="danger" effect="light">失败</el-tag>
        </template>
        </el-table-column>

        <el-table-column
          prop="access_time"
          label="实际执行时间"
          sortable
          :min-width="180"
          header-align="center"
        />
      </template>
    </tedge-table-layout>
  </template>

<script>
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';
import FiltersForm from './filters-form.vue';
import { curryingRenderElTextButton } from '../../../component/tedge-components/table-layout-context/render-el-text-button';
import { downloadByUrl } from '../../../../utils/download';

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
          border: true,
          height: "600",
        })
        .pagination()
        .search({
          placeholder: '输入门禁控制器编号进行搜索',
        })     
        .filters({
          method: '',
          state: '',
          create_time: [],
        }, {
          filtersForm: FiltersForm,
        })
        .remoteFilterPagination()
        .indexColumn({
          label: '序号',
        })
        .baseCurd({
          rowEditColumnWidth: 210,
          add: false,
          edit: false,
          remove: false,
          rowOprsComponents: [
            curryingRenderElTextButton({
              label: '重新执行',
              confirm: {
                title: '是否确认重新执行命令'
              }, 
              onClick: (props) => {
                this.reExecuteRow(props.row);
                this.tableLayoutContext.loadData();              
              },
              adminRight: true,
            }),
            curryingRenderElTextButton({
              label: '取消请求',
              adminRight: true,
              confirm: {
                title: '是否确认取消请求'
              },
              disabled: ({ row }) => {
                if (row.state === '过期') {
                  return '已经过期的请求无法取消';
                }
                return false;
              },
              onClick: (props) => {
                this.outdateRow(props.row);
                this.$message.success('请求已成功取消');
                this.tableLayoutContext.loadData();
              }
            }),
          ],
        })
        .selection({
          identity: row => row.id,
          oprs: [
            // () => import('./batch-re-execute.vue'),
            curryingRenderElTextButton({
              label: '重新执行',
              btnProps: {
                icon: 'tn-icon-refresh'
              },
              adminRight: true,
              confirm: {
                title: '是否确认批量重新执行选中记录?'
              },
              onClick: (props) => {
                this.batchReExecuteRow(props);
              },
              adminRight: true,
            }),
            'export'
          ],
        })
        .extraBtn(curryingRenderElTextButton({
          adminRight: true,
          label: '全部导出',
          btnProps: {
            type: 'primary',
            icon: 'tn-icon-download',
          },
          onClick: () => {
            this.exportData().bind(this);
          } 
        }))
        .extraBtn(curryingRenderElTextButton({
          adminRight: true,
          label: '设置自动过期策略',
          btnProps: {
            type: 'default',
            icon: 'el-icon-edit'
          },
          onClick: () => {
            this.open();
          } 
        }))
        .done();
    },
    open() {
      this.$prompt('请设置超时过期时间', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        // inputPattern: /[\w!#$%&'*+/=?^_`{|}~-]+(?:\.[\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\w](?:[\w-]*[\w])?\.)+[\w](?:[\w-]*[\w])?/,
        // inputErrorMessage: '邮箱格式不正确'
      }).then(({ value }) => {
        this.$message({
          type: 'success',
          message: '已设定超时过期时间为: ' + value + '天'
        });
      }).catch(() => {
        this.$message({
          type: 'info',
          message: '取消输入'
        });       
      });
    },
    async fetchData(filters, search, pagination) {
      const url = '/api/dcos/tdac-cgi/requests/info';
      if (filters.create_time === null) {
        filters.create_time = [];
      }

      return this.$axios.post(url, {
        offset: (pagination.current - 1) * pagination.size,
        limit: pagination.size,
        query: search,
        method: _.isNil(filters.method) || filters.method === '' ? undefined : filters.method,
        query_method: !_.isNil(filters.method) && filters.method !== '',

        state: _.isNil(filters.state) || filters.state === '' ? undefined : filters.state,
        query_state: !_.isNil(filters.state) && filters.state !== '',

        begin_time: _.isNil(filters.create_time[0]) ? undefined : filters.create_time[0].getTime(),
        end_time: _.isNil(filters.create_time[1]) ? undefined : filters.create_time[1].getTime(),
        query_create_time: !_.isNil(filters.create_time[0]) && !_.isNil(filters.create_time[1]),
      });
    },
    exportData() {
      const url = '/api/dcos/tdac-cgi/requests/export/all';
      downloadByUrl(url);
    },
    async reExecuteRow(row) {
      const url = '/api/dcos/tdac-cgi/requests/re-execute';
      await this.$axios.post(url, {
        ids: [row.id],
        method: row.method,
        payload: row.payload,
      })
    },  
    batchReExecuteRow: async function(props) {
      try {
        const selectedRows = props.getSelectedRows();
        const idList = selectedRows.map(row => row.id);
        
        // 使用组件实例的axios
        await this.$axios.post('/api/dcos/tdac-cgi/requests/batch-re-execute', {
          ids: idList
        });
        
        this.$message.success(`已触发 ${idList.length} 条记录重新执行`);
        props.tableContext.loadData();
      } catch (error) {
        this.$message.error('操作失败: ' + error.message);
      }
    },
    async outdateRow(row) {
      const url = '/api/dcos/tdac-cgi/requests/outdate';
      await this.$axios.post(url, {
        ids: [row.id],
      })
    },
    joinArray(arr) {
      return arr.join('，');
    },
    showCardInfoOfUser(user) {
      this.$refs.cardListBelongUser.show(user, () => {
        this.tableLayoutContext.loadData();
      });
    },
  },
};
</script>