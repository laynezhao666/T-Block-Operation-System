<template>
  <tedge-table-layout
    :context="tableLayoutContext"
  >
    <template #columns>
      <el-table-column
        prop="number"
        label="门编号"
        width="120"
      />

      <el-table-column
        prop="name"
        label="门名称"
        width="100"
      />

      <el-table-column
        prop="controlName"
        label="控制器"
        show-overflow-tooltip
      />

      <el-table-column
        prop="groupName"
        label="所属分组"
      />
    </template>
  </tedge-table-layout>
</template>

<script>
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import { loadControlAndGroupsTree } from '../../doors-overview/utils/fetch-group-control-door-trees';
import { createRemovableTableContext } from '../components/create-removable-table-context';

export default {
  components: {
    TedgeTableLayout,
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
    window.dt = this;
    return {
      tableLayoutContext: this.createTableLayoutContext(),
      allDoors: [],
    };
  },
  methods: {
    createTableLayoutContext() {
      return createRemovableTableContext(
        this.fetchData.bind(this),
        this.removeRow.bind(this),
      );
    },
    async fetchData() {
      if (!this.allDoors?.length) {
        await this.loadAllDoors();
      }

      const {
        allDoors,
        permissionGroup: {
          doors: relatedDoors,
        },
      } = this;

      const relatedDoorIdSet = new Set(_.map(relatedDoors, 'id'));

      return _.filter(allDoors, door => relatedDoorIdSet.has(door.id));
    },
    async loadAllDoors() {
      const {
        controlsTree,
        groupsTree,
      } = await loadControlAndGroupsTree();

      let doors = _.chain(controlsTree)
        .map('doors')
        .flatten()
        .filter(Boolean)
        .value();

      const controlMap = _.mapKeys(controlsTree, 'id');
      const groupMap = _.mapKeys(groupsTree, 'id');

      doors = _.map(doors, door => ({
        ...door,
        control: controlMap[door.controller_id],
        group: groupMap[door.group_id],
        groupName: groupMap[door.group_id]?.name,
      }));

      this.allDoors = doors;
    },
    async removeRow(row) {
      const permissionGroup = _.cloneDeep(this.permissionGroup);

      permissionGroup.doors = (permissionGroup.doors || []).filter(item => item.id !== row.id);

      await this.updatePermissionGroup(permissionGroup);

      this.tableLayoutContext.forceReloadData();
      this.$message.success('移除成功');
    },
  },
};
</script>
