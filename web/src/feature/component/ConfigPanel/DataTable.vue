<template>
  <custom-table
    ref="table"
    :columns="columns"
    :cgi="url"
    :img-url="getImgUrl"
    :query="query"
    :cur-table="table"
    paging="remote"
    :manual-init="true"
    column-key="name"
    method="post"
    :row-key="id"
    :border-prop="borderProp"
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
      v-if="hasRights('canEdit')||hasRights('canDel')||hasRights('customizeBtn')"
      slot="actions"
      slot-scope="scope"
    >
      <el-button
        v-if="hasRights('canEdit')"
        type="text"
        size="small"
        :auth-roles="roles.write"
        :auth-right-code="codes.bj"
        :disabled="unEdit(scope.data.row)"
        @click="edit(scope.data.row)"
      >
        编辑
      </el-button>
      <el-button
        v-if="hasRights('canDel')"
        v-confirm="() => del(scope.data.row)"
        type="text"
        size="small"
        :auth-roles="roles.write"
        :auth-right-code="codes.sc"
        :disabled="unEdit(scope.data.row)"
        :style="{marginLeft: '8px'}"
      >
        删除
      </el-button>
      <div v-if="hasRights('customizeBtn')">
        <slot
          name="customizeBtn"
          :data="scope.data.row"
        />
      </div>
    </template>
    <template
      slot="buttons"
      slot-scope="scope"
    >
      <el-button
        v-if="hasRights('canExport')"
        :auth-right-code="codes.dc"
        type="text"
        @click="expBatch(scope.data)"
      >
        <i class="tn-icon-import" />
        <span>导出所选</span>
      </el-button>
      <el-button
        v-if="hasRights('canDel')"
        :auth-roles="roles.write"
        :auth-right-code="codes.sc"
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
import { map, find } from 'lodash';
import CustomTable from 'component/Table';
import mixin from 'component/script/mixin';
import configMixin from './mixin';

export default {
  components: {
    CustomTable,
  },
  inject: ['configCgi', 'commonCgi'],
  mixins: [mixin, configMixin],
  // eslint-disable-next-line vue/require-prop-types
  props: ['columns', 'query', 'table', 'id', 'codes', 'borderProp', 'roles'],
  data() {
    return {
      url: this.configCgi.getMgrList,
      getImgUrl: this.commonCgi.downloadImage,
      mainField: void 0,
    };
  },
  watch: {
    columns(v) {
      if (v.length) {
        const match = find(v, {
          fixed: true,
        });
        if (match && !match.jump) this.mainField = match.name;
      }
    },
  },
  methods: {
    expBatch(selection) {
      this.$emit('export', selection.map(item => item[this.id]));
    },
    delBatch(selection) {
      this.$confirm('确认要删除吗？', '系统提示').then(() => {
        this.$axios.post(this.configCgi.delMgrData, {
          table: this.table,
          id: map(selection, this.id),
        }, true, { restAxios: { timeout: 60000 } }).then(this.cb);
      });
    },
    refresh() {
      this.$emit('success');
      this.$refs.table.refresh();
    },
    del(data) {
      this.$axios.post(this.configCgi.delMgrData, {
        table: this.table,
        id: [data[this.id]],
      }).then(this.cb);
    },
    edit(row) {
      const path = this.getPath();
      const query = this.getQuery();
      if (path) {
        this.jump(path, query, row);
      } else {
        this.$emit('edit', row);
      }
    },
    cb() {
      this.$message({
        type: 'success',
        message: '删除成功',
      });
      this.refresh();
    },
    unEdit(row) {
      const unEditTable = ['assettype', 'assetmodel'];
      const key = `${this.table}_unedit`;
      try {
        return unEditTable.includes(this.table) && row[key] === '1';
      } catch {
        return false;
      }
    },
  },
};
</script>
