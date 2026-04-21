<template>
  <el-button
    type="text"
    @click="show"
  >
    查看权限信息

    <el-modal
      v-if="modalVisible"
      :visible.sync="modalVisible"
    >
      <template slot="title">
        查看权限信息
      </template>

      <el-tabs
        v-model="activeTab"
      >
        <el-tab-pane
          label="授权门范围"
          name="doors"
        >
          <doors-table
            :permission-group="row"
            :update-permission-group="updatePermissionGroup"
          />
        </el-tab-pane>
        <el-tab-pane
          label="关联人员"
          name="staffs"
        >
          <cards-table
            :permission-group="row"
            :update-permission-group="updatePermissionGroup"
          />
        </el-tab-pane>
      </el-tabs>
    </el-modal>
  </el-button>
</template>

<script>
import DoorsTable from './doors-table.vue';
import CardsTable from './cards-table.vue';
import { axiosPut } from '../../../../utils/axios-methods';

export default {
  components: {
    DoorsTable,
    CardsTable,
  },
  props: {
    row: {
      type: Object,
      required: true,
    },
    tableContext: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {
      modalVisible: false,
      activeTab: 'doors',
    };
  },
  watch: {
    modalVisible(visible) {
      if (!visible) {
        this.tableContext.loadData();
      }
    },
  },
  methods: {
    show() {
      this.modalVisible = true;
    },
    async updatePermissionGroup(newPermissionGroup) {
      const dataToPost = {
        ...newPermissionGroup,
        door: undefined,
        card: undefined,
        door_id: undefined,
        doors: _.map(newPermissionGroup.doors, 'id'),
        cards: _.map(newPermissionGroup.cards, 'card_no'),
      };

      await axiosPut(`/api/dcos/tdac-cgi/access-group/${newPermissionGroup.id}`, dataToPost);
      Object.assign(this.row, newPermissionGroup);
    },
  },
};
</script>
