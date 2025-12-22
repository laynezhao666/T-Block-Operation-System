<template>
  <div class="data-query">
    <el-title v-if="showTitle">
      采集数据查询
    </el-title>
    <el-block>
      <div class="data-query-container">
        <tree-component
          ref="treeContainer"
          data-source-type="byCollect"
          class="data-query-tree"
          :mozuloaded="mozuloaded"
          :mozu-id="mozuId"
          @node-click="handleClick"
          @refresh="refresh"
        />
        <realtime-component
          ref="realtime"
          class="data-query-realtime"
          :mozuloaded="mozuloaded"
          :mozu-id="mozuId"
        />
      </div>
    </el-block>
  </div>
</template>

<script>
import business from '@@/config/business';
import treeComponent from './common/tree.vue';
import realtimeComponent from './common/by-collect.vue';
import mixin from 'feature/utils/mixin';

export default {
  components: {
    treeComponent,
    realtimeComponent,
  },
  mixins: [mixin],
  provide() {
    const that = this;
    return {
      getSelNodeData() {
        return that.nodeData;
      },
    };
  },
  props: {
    showTitle: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      business,
      nodeData: null,
      mozuloaded: false,
      mozuId: 0,
    };
  },
  mounted() {
    this.changeMozu();
  },
  methods: {
    changeMozu() {
      this.mozuloaded = true;
      this.mozuId = TNBL.getCurrModule().id;
      this.nodeData = null;
    },
    handleClick(data) {
      this.nodeData = data;

      this.refresh();
    },
    // 切换到“配置管理”tab并刷新
    refresh() {
      const $child = this.$refs.realtime;
      $child.refresh();
    },
  },
};

</script>

<style lang="scss" scoped>
/deep/ .el-tabs--card > .el-tabs__header .el-tabs__nav{
  border: none;
}
.data-query {
  &-container {
    display: flex;
    height: calc(100vh - 120px - 64px - 16px);
  }

  &-tree {
    width: 20%;
    border-right: 1px solid #f0f0f0;
    height: 100%;

    /deep/ .c-tree-wrap {
      height: calc(100% - 56px - 64px);
      overflow-y: overlay;
    }
  }

  &-realtime {
    flex: 1;
  }
}
</style>
