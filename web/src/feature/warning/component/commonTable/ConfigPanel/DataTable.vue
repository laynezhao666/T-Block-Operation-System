<template>
  <custom-table
    ref="table"
    :columns="columns"
    :local-data="localData"
    :query="query"
    :play="play"
    :show-table-select="showTableSelect"
  >
    <template
      v-if="mainField"
      :slot="mainField"
      slot-scope="scope"
    >
      <a
        href="javascript:void(0)"
        @click="$emit('expand', scope.data.row, mainField)"
      >
        {{ scope.data.row[mainField] }}
      </a>
    </template>
    <template
      v-if="hasRights('canEdit')||hasRights('canDel') || tableConfig.hasDetail"
      slot="actions"
      slot-scope="scope"
    >
      <a
        v-if="tableConfig.hasDetail"
        href="javascript:void(0)"
        :disabled="unEdit(scope.data.row)"
        @click="detail(scope.data.row)"
      >
        详情
      </a>
      <a
        v-if="hasRights('canEdit')"
        v-appmatrixauth="roles.write"
        href="javascript:void(0)"
        :style="{marginLeft: '8px'}"
        :disabled="unEdit(scope.data.row)"
        @click="edit(scope.data.row)"
      >
        编辑
      </a>
      <a
        v-if="hasRights('canDel')"
        v-appmatrixauth="roles.write"
        href="javascript:void(0)"
        :disabled="unEdit(scope.data.row)"
        :style="{marginLeft: '8px'}"
        @click="del(scope.data.row)"
      >
        删除
      </a>
    </template>
    <template
      slot="buttons"
      slot-scope="scope"
    >
      <el-button
        v-if="hasRights('canExport')"
        type="text"
        @click="expBatch(scope.data)"
      >
        <i class="tn-icon-import" />
        <span>导出所选</span>
      </el-button>
      <el-button
        v-if="hasRights('canDel')"
        v-appmatrixauth="roles.write"
        type="text"
        @click="delBatch(scope.data)"
      >
        <i class="tn-icon-delete" />
        <span>删除所选</span>
      </el-button>
    </template>
  </custom-table>
</template>
<script>
import { map } from 'lodash';
import CustomTable from '../Table';
import mixin from '../script/mixin';
import getEdgeRequest from '../../../../utils/request';

export default {
  components: {
    CustomTable,
  },
  inject: ['configCgi', 'tableConfig'],
  mixins: [mixin],
  props: {
    codes: {
      type: Object,
      required: true,
    },
    roles: {
      type: Object,
      required: true,
    },
    columns: {
      type: Array,
      required: true,
    },
    localData: {
      type: Array,
      required: false,
      default: () => ([]),
    },
    query: {
      type: Object,
      default: () => ({}),
    },
    play: {
      type: Boolean,
      default: false,
    },
    showTableSelect: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      mainField: void 0,
      rights: this.tableConfig.rights,
      id: this.tableConfig.id || 'id',
    };
  },
  watch: {
    columns(v) {
      if (v.length) {
        const match = find(v, {
          fixed: true,
        });
        if (match) this.mainField = match.name;
      }
    },
  },
  methods: {
    expBatch(selection) {
      this.$emit('export', selection.map(item => item[this.id]));
    },
    delBatch(selection) {
      const ids = map(selection, this.id);
      this.$confirm('确认要删除吗？', '系统提示', { type: 'warning' }).then(() => {
        getEdgeRequest(this.$axios, this.tableConfig.searchParams.mozuId).post(this.tableConfig.deleteCgi, {
          ids: ids.map(item => parseInt(item)),
        })
          .then(this.cb);
      })
        .catch(() => {});
    },
    refresh(axiosLoading) {
      this.$refs.table.refresh(axiosLoading);
    },
    del(data) {
      this.$confirm('确认要删除吗？', '系统提示', { type: 'warning' }).then(() => {
        getEdgeRequest(this.$axios, this.tableConfig.searchParams.mozuId).post(this.configCgi.delCgi, {
          id: data[this.id],
        })
          .then(this.cb)
          .catch(() => {});
      })
        .catch(() => {});
    },
    edit(row) {
      this.$emit('edit', row);
    },
    detail(row) {
      location.href = `${this.configCgi.detailUrl}?id=${row.id}`;
    },
    cb() {
      this.$message({
        type: 'success',
        message: '删除成功',
      });
      this.refresh();
    },
    unEdit(row) {
      if (row.unedit) {
        return true;
      }
      return false;
    },
  },
};
</script>
