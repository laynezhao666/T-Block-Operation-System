<template>
  <two-columns-outline
    :drawer="false"
    class="doors-overview"
  >
    <div
      slot="left"
      class="left"
    >
      <el-tabs
        v-model="activeTreeTab"
      >
        <el-tab-pane
          label="控制器"
          name="control"
        />
        <el-tab-pane
          label="分组"
          name="group"
        />
      </el-tabs>

      <div
        class="tree-view-container"
      >
        <tree-view
          ref="treeView"
          :mode="activeTreeTab"
          @doorsChange="handleDoorsChange"
          @showRecords="showRecords"
          @setDoorParams="setDoorParams"
        />
      </div>
    </div>

    <div
      slot="right"
      class="full-width-only"
    >
      <doors-view
        :doors="showingDoors"
        @reloadTree="reloadTree"
        @showRecords="showRecords"
        @batchSetDoorParams="batchSetDoorParams"
      />

      <card-records-modal
        ref="cardRecordsModal"
      />

      <door-form-modal
        ref="doorFormModal"
      />
    </div>
  </two-columns-outline>
</template>

<script>
import TwoColumnsOutline from '../../component/tedge-components/two-columns-outline.vue';
import TreeView from './components/tree-view.vue';
import DoorsView from './components/doors-view.vue';
import CardRecordsModal from './components/card-records-modal.vue';
import DoorFormModal from './components/door-form-modal.vue';
import { axiosPut } from '../../../utils/axios-methods';

export default {
  components: {
    TwoColumnsOutline,
    TreeView,
    DoorsView,
    CardRecordsModal,
    DoorFormModal,
  },
  data() {
    return {
      activeTreeTab: 'control',
      showingDoors: [],
    };
  },
  methods: {
    reloadTree() {
      this.$refs.treeView.loadData();
    },
    showRecords(doors) {
      this.$refs.cardRecordsModal.show(doors);
    },
    setDoorParams(door) {
      this.$refs.doorFormModal.edit(door, (params) => {
        return this.saveDoorsParams([door.id], params, false);
      }, false);
    },
    batchSetDoorParams(ids) {
      this.$refs.doorFormModal.edit({}, params => this.saveDoorsParams(ids, params, true), true);
    },
    async saveDoorsParams(ids, params, isBatch) {
      const pureParams = {
        ...params,
        password: params.password === '******' ? undefined : params.password,
        passwordConfirm: undefined,
        controlId: undefined,
        controlName: undefined,
        type: undefined,
        parameters: undefined,
        idcdbCode: undefined,
        relatedCameras: undefined,
      };

      const dataToPost = isBatch ? {
        ids: _.map(ids, Number),
        params: pureParams,
      } : {
        id: ids[0],
        params: pureParams,
        code: params.idcdbCode,
        extend: {
          relatedCameras: params.relatedCameras,
        },
      };

      const url = isBatch
        ? '/api/dcos/tdac-cgi/doors'
        : '/api/dcos/tdac-cgi/door';

      await axiosPut(url, dataToPost);
      this.$message.success('设置门参数成功');

      setTimeout(() => {
        this.$refs.treeView.loadData();
      }, 300);

      return 'close';
    },
    handleDoorsChange(doors) {
      this.showingDoors = doors;
    },
  },
};
</script>

<style lang="scss" scoped>
.doors-overview {
  height: calc(100vh - 92px);
}

.left {
  min-width: 300px;
  height: 100%;

  display: flex;
  flex-direction: column;
}

.tree-view-container {
  flex: 1;
  overflow: auto;
}

.full-width-only {
  height: 100%;
  overflow: hidden;
}
</style>
