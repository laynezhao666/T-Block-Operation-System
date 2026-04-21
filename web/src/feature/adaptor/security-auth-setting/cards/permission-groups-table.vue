<template>
  <div>
    <split-header-bar
      title="卡授权"
    >
      <el-button
        type="text"
        @click="startSelectGroups"
      >
        选择权限组
      </el-button>
    </split-header-bar>

    <tedge-table-layout
      :context="tableLayoutContext"
    >
      <template #columns>
        <el-table-column
          label="权限组"
          prop="name"
          width="120px"
        />

        <el-table-column
          label="时间组"
          prop="time_group.group_name"
          width="120px"
        />

        <el-table-column
          label="门范围"
          prop="doors"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ formatDoorInfo(row.doors) }}
          </template>
        </el-table-column>

        <el-table-column
          label="类型"
          prop="label"
          width="120px"
        />

        <el-table-column
          label="说明"
          prop="comment"
          width="120px"
          show-overflow-tooltip
        />
      </template>
    </tedge-table-layout>

    <el-dialog
      :visible.sync="selectPermissionDialogVisible"
      title="选择权限组"
      width="80%"
      append-to-body
    >
      <transform-table-to-table
        ref="TransformTableToTable"
        :all-options="allGroups"
        :selected-keys="selectedGroupIdList"
        :row-identity="getGroupId"
      >
        <template
          #table-columns
        >
          <el-table-column
            label="权限组"
            prop="name"
            width="120px"
          />

          <el-table-column
            label="门范围"
            prop="doors"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              {{ formatDoorInfo(row.doors) }}
            </template>
          </el-table-column>

          <el-table-column
            label="类型"
            prop="label"
            width="120px"
          />

          <el-table-column
            label="说明"
            prop="comment"
            width="120px"
            show-overflow-tooltip
          />
        </template>
      </transform-table-to-table>

      <span
        slot="footer"
        class="dialog-footer"
      >
        <el-button @click="selectPermissionDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="submitPermissionGroups"
        >
          确定
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import SplitHeaderBar from '../../../component/tedge-components/split-header-bar.vue';
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';
import TransformTableToTable from '../../../component/tedge-components/transform-table-to-table.vue';

export default {
  components: {
    SplitHeaderBar,
    TedgeTableLayout,
    TransformTableToTable,
  },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    /** 选中的分组id数组 */
    value: {
      type: Array,
      required: true,
    },
  },
  data() {
    window.pgt = this;
    return {
      tableLayoutContext: this.createTableLayoutContext(),
      selectPermissionDialogVisible: false,
      allGroups: [],
    };
  },
  computed: {
    selectedGroupIdList() {
      return this.value;
    },
  },
  watch: {
    groups: {
      handler() {
        this.tableLayoutContext.forceReloadData();
      },
    },
    allGroups: {
      handler() {
        this.tableLayoutContext.forceReloadData();
      },
    },
  },
  created() {
    this.loadAllGroups();
  },
  methods: {
    formatDoorInfo(doorList) {
      return _.chain(doorList)
        .map('name')
        .join('，')
        .value();
    },
    getGroupId(group) {
      return Number(group.id);
    },

    createTableLayoutContext() {
      return chainTableLayout(this.getTableData)
        .hideToolbar()
        .indexColumn()
        .pagination()
        .localFilterPagination()
        .done();
    },
    getTableData() {
      const map = _.mapKeys(this.allGroups, 'id');
      return this.selectedGroupIdList.map(id => map[id]).filter(Boolean);
    },
    async loadAllGroups() {
      const url = '/api/dcos/tdac-cgi/access-groups';
      const resp = await this.$axios.get(url, {
        offset: 0,
        limit: 100000,
      });
      this.allGroups = resp.list;
    },
    startSelectGroups() {
      this.selectPermissionDialogVisible = true;
    },
    submitPermissionGroups() {
      this.$emit('change', this.selectedGroupIdList);
      this.tableLayoutContext.forceReloadData();
      this.selectPermissionDialogVisible = false;
    },
  },
};
</script>

<style lang="scss" scoped>

</style>
