<template>
  <tedge-table-layout
    :context="tableLayoutContext"
    class="table"
  >
    <template #columns>
      <card-common-columns />

      <el-table-column
        prop="staff.name"
        label="姓名"
        width="120"
      >
        <template #default="{ row }">
          <span v-show="row.staff && row.staff.name">
            {{ row.staff && row.staff.name }}
          </span>
          <span
            v-show="!row.staff || !row.staff.name"
            class="danger-text"
          >
            未分配
          </span>
        </template>
      </el-table-column>
    </template>

    <template
      slot="outer-modals"
    >
      <allocation-dialog
        ref="allocationDialog"
      />
    </template>
  </tedge-table-layout>
</template>

<script>
import _ from 'lodash';
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';
import { curryingRenderElTextButton } from '../../../component/tedge-components/table-layout-context/render-el-text-button';
import allocationDialog from './allocation-dialog.vue';
import CardCommonColumns from '../components/card-common-columns.vue';
import { axiosDelete, axiosPut, axiosUploadFile } from '../../../../utils/axios-methods';
import FiltersForm from './filters-form.vue';

export default {
  components: {
    TedgeTableLayout,
    allocationDialog,
    CardCommonColumns,
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
          placeholder: '请输入关键字按卡号、姓名搜索',
          // isHide: true,
        })
        .filters({
          card_type: '',
          card_flag: '',
          access_group: '',
        }, {
          filtersForm: FiltersForm,
        })
        .remoteFilterPagination()
        .indexColumn({
          label: '序号',
        })
        .baseCurd({
          rowEditColumnWidth: 220,
          add: {
            adminRight: true,
            action: () => this.nomalizeForForm({}),
          },
          remove: {
            confirm: this.removeRow.bind(this),
            batchRemove: this.batchRemoveRows.bind(this),
            adminRight: true,
          },
          edit: false,
          rowOprsComponents: [
            curryingRenderElTextButton({
              adminRight: true,
              label: ({ row }) => {
                const text = row.card_flag === 0 ? '停用' : '启用';
                return text;
              },
              onClick: (props) => {
                this.toggleRowEnable(props.row, props.index, !(props.row.card_flag === 0));
              },
            }),
            curryingRenderElTextButton({
              adminRight: true,
              label: '重新授权',
              onClick: ({ row, tableContext }) => {
                // eslint-disable-next-line no-param-reassign
                tableContext.curd.editting = _.cloneDeep(row);
                // eslint-disable-next-line no-param-reassign
                tableContext.curd.isCreate = false;
              },
            }),
            curryingRenderElTextButton({
              adminRight: true,
              label: '分配',
              disabled: ({ row }) => !!row.userName,
              // btnProps: ({ row }) => ({
              //   disabled: !!row.userName,
              // }),
              onClick: ({ row }) => {
                this.allocationRow = row;
                this.$refs.allocationDialog.show(row, () => {
                  this.tableLayoutContext.loadData();
                });
              },
            }),
          ],
        })
        .extraBtn(curryingRenderElTextButton({
          label: '发卡',
          adminRight: true,
          btnProps: {
            type: 'primary',
          },
          onClick: () => {
            this.handleIssueCard();
          },
        }))
        .curdFormModal({
          title: (isCreate) => {
            const text = isCreate ? '添加门禁卡' : '重新授权';
            return text;
          },
          formComp: () => import('./form.vue'),
          beforeEdit: (editting, replace) => replace(this.nomalizeForForm(editting)),
          onSubmit: this.submitEditForm.bind(this),
        })
        .selection({
          identity: row => row.card_no,
          oprs: [
            'export',
            'delete',
            () => import('./batch-update-valid-time.vue'),
          ],
        })
        .toolbarActions({
          icon: 'tn-icon-import',
          text: '导入门禁卡信息',
          action: () => {
            this.handleImportCards();
          },
        })
        .toolbarActions({
          icon: 'tn-icon-export',
          text: '导出门禁卡信息',
          action: () => {
            this.handleExportCards();
          },
        })
        .done();
    },
    async fetchData(filters, search, pagination) {
      const url = '/api/dcos/tdac-cgi/cards';
      return this.$axios.post(url, {
        offset: (pagination.current - 1) * pagination.size,
        limit: pagination.size,
        query: search,
        card_type: _.isNil(filters.card_type) || filters.card_type === '' ? undefined : filters.card_type,
        query_card_type: !_.isNil(filters.card_type) && filters.card_type !== '',

        card_flag: _.isNil(filters.card_flag) || filters.card_flag === '' ? undefined : filters.card_flag,
        query_card_flag: !_.isNil(filters.card_flag) && filters.card_flag !== '',

        access_group: _.isNil(filters.access_group) || filters.access_group === '' ? undefined : filters.access_group,
        query_access_group: !_.isNil(filters.access_group) && filters.access_group !== '',
      });
    },
    joinArray(arr, fieldsPath) {
      return (fieldsPath ? _.map(arr, fieldsPath) : arr).join('，');
    },
    async removeRow(row) {
      await axiosDelete('/api/dcos/tdac-cgi/card', {
        card: row.card_no,
      });
    },
    batchRemoveRows(rows) {
      // TODO: 批量删除
      console.log('//TODO: batchRemoveRows', rows);
    },
    async toggleRowEnable(row, index, isEnable) {
      await this.toggleCardEnabelStatus(row.card_no, isEnable);

      // eslint-disable-next-line no-param-reassign
      row.card_flag = isEnable ? 0 : 1;
      this.$message.success(`${isEnable ? '启用' : '停用'}成功`);
    },
    async submitEditForm(row, isCreate) {
      const dataToPost = this.parseRowFormDataToPost(row);

      try {
        if (isCreate) {
          await this.$axios.post('/api/dcos/tdac-cgi/card', dataToPost);
        } else {
          // 重新授权时更新卡类型
          await this.updateCardType(row.card_no, row.card_type);

          // 重新授权时更新访问组
          await this.updateCardAccess(row.card_no, row.access_groups, false);

          // 更新卡片有效期
          if (row.card_type === 1) {
            // 临时卡：更新有效期
            if (row.valid_time) {
              await this.updateCardValidTime(row.card_no, dataToPost.valid_time);
            }
          } else {
            // 永久卡：清除有效期（设置为0）
            await this.updateCardValidTime(row.card_no, 0);
          }

          // 更新员工绑定
          if (row.staff?.id) {
            this.grantToStaff(row.card_no, row.staff?.id, false);
          } else {
            this.unbindCard(row);
          }
        }

        return true;
      } catch (error) {
        // 处理卡号重复错误
        if (error.code === 10001 && error.message?.includes('1062')) {
          this.$message.error(`卡号 ${row.card_no} 已存在，请使用其他卡号`);
          // 抛出错误以阻止表单关闭，但不显示默认错误提示
          throw new Error('CARD_DUPLICATE');
        }

        throw error;
      }
    },

    async unbindCard(row) {
      await axiosPut('/api/dcos/tdac-cgi/card/unbind', {
        card: row.card_no,
      });
    },
    async grantToStaff(cardId, staffId) {
      await axiosPut('/api/dcos/tdac-cgi/card/staff', {
        card: cardId,
        staff: staffId,
      });
    },
    async updateCardAccess(cardId, permissionGroupsIds) {
      await axiosPut('/api/dcos/tdac-cgi/card/access', {
        access_groups: permissionGroupsIds,
        cards: [cardId],
      });
    },
    async updateCardType(cardId, cardType) {
      const url = '/api/dcos/tdac-cgi/card/type';
      await axiosPut(url, {
        cards: [cardId],
        type: cardType,
      });
    },
    async updateCardValidTime(cardId, validTime) {
      const url = '/api/dcos/tdac-cgi/card/valid_time';
      await axiosPut(url, {
        cards: [cardId],
        valid_time: validTime,
      });
    },
    async toggleCardEnabelStatus(cardId, isEnable) {
      const url = '/api/dcos/tdac-cgi/card/flag';
      await axiosPut(url, {
        cards: [cardId],
        flag: isEnable ? 0 : 1,
      });
    },

    nomalizeForForm(card) {
      return {
        card_no: '',
        card_flag: 0,
        card_type: 0,
        // valid_date: [],
        ...card,
        valid_time: card.valid_time * 1000,
        staff: (!card.staff || card.staff.id === 0) ? {
          id: null,
          enable: 1,
        } : card.staff,
        access_groups: _.map(card.access_groups || [], 'id'),
      };
    },
    parseRowFormDataToPost(rowFormData) {
      return {
        ...rowFormData,
        valid_time: (rowFormData.valid_time?.getTime
          ? rowFormData.valid_time.getTime()
          : rowFormData.valid_time
        ) / 1000,
        staff: rowFormData.staff,
        access_groups: _.map(rowFormData.access_groups),
      };
    },
    async handleImportCards() {
      try {
        await axiosUploadFile('/api/dcos/tdac-cgi/cards/import', {
          file: axiosUploadFile.fileSelectSymbol,
        });
        this.tableLayoutContext.loadData();
        this.$message.success('导入门禁卡信息完成');
      } catch (error) {
        console.error('导入门禁卡信息失败:', error);
      }
    },
    handleExportCards() {
      this.$axios.download('/api/dcos/tdac-cgi/cards/export', {}, true, {
        fileName: '门禁卡信息.xlsx',
      });
    },
    // 发卡：调用NFC本地服务接口
    async handleIssueCard() {
      try {
        const { value: cardNo } = await this.$prompt('请输入卡号（纯数字，最多16位）', '发卡', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          inputPattern: /^[0-9]{1,16}$/,
          inputErrorMessage: '卡号只能包含数字（0-9），最多16位',
          inputPlaceholder: '请输入数字卡号',
        });

        if (!cardNo) return;

        await this.issueNfcCard(cardNo.trim());
      } catch {
        // 用户取消输入，不做处理
      }
    },
    async issueNfcCard(cardNo) {
      // 将数字字符串的每个字符转为2位16进制
      const paraBin = Array.from(cardNo)
        .map(ch => ch.charCodeAt(0).toString(16).padStart(2, '0'))
        .join('');

      // NFC 读卡器密钥从环境变量获取，避免硬编码
      const mfOldKey = process.env.NFC_MF_OLD_KEY;
      const mfNewKey = process.env.NFC_MF_NEW_KEY;
      const dfKey = process.env.NFC_DF_KEY;

      const requestData = {
        cmdID: String(Date.now()),
        para: {
          paraMFOldKey: mfOldKey,
          paraMFNewKey: mfNewKey,
          paraDFID: '02',
          paraDFKey: dfKey,
          paraBin,
        },
      };

      const loading = this.$loading({
        lock: true,
        text: '正在发卡，请将卡片放置在读卡器上...',
        spinner: 'el-icon-loading',
        background: 'rgba(0, 0, 0, 0.7)',
      });

      try {
        const response = await fetch('http://localhost:8088/api/nfc/initial', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(requestData),
        });
        const result = await response.json();

        if (result.para?.result === true) {
          this.$message.success(`发卡成功，卡号为${result.para.uid}`);
        } else {
          const errMsg = result.para?.errMessage || '未知错误';
          this.$message.error(`发卡失败，${errMsg}`);
        }
      } catch (error) {
        console.error('发卡请求失败:', error);
        this.$message.error('发卡请求失败，请检查读卡器服务是否启动（端口8088）');
      } finally {
        loading.close();
      }
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

.danger-text {
  color: var(--tn-color-danger);
}

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

.table /deep/ {
  .column-cell-center {
    & > .cell {
      text-align: center;
    }
  }
}
</style>
