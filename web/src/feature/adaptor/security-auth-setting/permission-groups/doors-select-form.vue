<template>
  <el-form
    ref="form"
    label-position="top"
    :model="editting"
    :rules="rules"
  >
    <el-form-item
      prop="doors"
      label-width="100%"
      class="header-like-form-item"
    >
      <split-header-bar
        slot="label"
        title="授权门范围"
        no-padding
      />

      <el-tabs
        v-model="treeMode"
        type="border-card"
        size="small"
        class="tabs"
      >
        <el-tab-pane
          label="按分组选择"
          name="byGroup"
          lazy
        >
          <div class="tree-toolbar">
            <el-checkbox
              v-model="allCheckedByGroup"
              @change="onSelectAllGroup"
            >
              全选
            </el-checkbox>
            <el-input
              v-model="filterText"
              placeholder="请输入关键词"
              size="small"
              clearable
              suffix-icon="el-icon-search"
              class="tree-filter-input"
            />
          </div>
          <el-tree
            ref="byGroup"
            :data="groupTree"
            :props="treeProps"
            :default-checked-keys="editting.doors"
            :filter-node-method="filterNode"
            node-key="nodeKey"
            show-checkbox
            @check-change="handleCheckChange"
          />
        </el-tab-pane>
        <el-tab-pane
          label="按控制器选择"
          name="byControl"
          lazy
        >
          <div class="tree-toolbar">
            <el-checkbox
              v-model="allCheckedByControl"
              @change="onSelectAllControl"
            >
              全选
            </el-checkbox>
            <el-input
              v-model="filterTextControl"
              placeholder="请输入关键词"
              size="small"
              clearable
              suffix-icon="el-icon-search"
              class="tree-filter-input"
            />
          </div>
          <el-tree
            ref="byControl"
            :data="controlTree"
            :props="treeProps"
            :default-checked-keys="editting.doors"
            :filter-node-method="filterNode"
            node-key="nodeKey"
            show-checkbox
            @check-change="handleCheckChange"
          />
        </el-tab-pane>
      </el-tabs>
    </el-form-item>
  </el-form>
</template>

<script>
import SplitHeaderBar from '../../../component/tedge-components/split-header-bar.vue';
import { loadControlAndGroupsTree } from '../../doors-overview/utils/fetch-group-control-door-trees';

export default {
  components: {
    SplitHeaderBar,
  },
  props: {
    editting: {
      type: Object,
      required: true,
    },
    isCreate: {
      type: Boolean,
      required: true,
    },
  },
  data() {
    window.dsf = this;
    return {
      groupTree: [],
      controlTree: [],
      treeMode: 'byGroup',
      filterText: '',
      filterTextControl: '',
      allCheckedByGroup: false,
      allCheckedByControl: false,
      treeProps: {
        label: data => data.name,
        children: 'doors',
      },

      rules: {
      },
    };
  },
  computed: {
    // 获取分组树所有叶子节点的 nodeKey
    allGroupLeafKeys() {
      const keys = [];
      const traverse = (nodes) => {
        nodes.forEach(node => {
          if (node.doors && node.doors.length) {
            traverse(node.doors);
          } else {
            keys.push(node.nodeKey);
          }
        });
      };
      traverse(this.groupTree);
      return keys;
    },
    // 获取控制器树所有叶子节点的 nodeKey
    allControlLeafKeys() {
      const keys = [];
      const traverse = (nodes) => {
        nodes.forEach(node => {
          if (node.doors && node.doors.length) {
            traverse(node.doors);
          } else {
            keys.push(node.nodeKey);
          }
        });
      };
      traverse(this.controlTree);
      return keys;
    },
  },
  watch: {
    filterText(val) {
      this.$refs.byGroup && this.$refs.byGroup.filter(val);
    },
    filterTextControl(val) {
      this.$refs.byControl && this.$refs.byControl.filter(val);
    },
  },
  created() {
    this.loadOptions();
  },
  methods: {
    async validate() {
      return this.$refs.form.validate();
    },
    async loadOptions() {
      await this.loadTrees();
      this.initSetTreeChecked();
    },
    async loadTrees() {
      const {
        controlsTree,
        groupsTree,
      } = await loadControlAndGroupsTree();

      this.controlTree = controlsTree;
      this.groupTree = groupsTree;
    },
    initSetTreeChecked() {
      const {
        byGroup,
        byControl,
      } = this.$refs;
      const {
        doors,
      } = this.editting;

      if (!doors) return;

      // eslint-disable-next-line no-unused-expressions, babel/no-unused-expressions
      byGroup?.setCheckedKeys(doors);
      // eslint-disable-next-line no-unused-expressions, babel/no-unused-expressions
      byControl?.setCheckedKeys(doors);
    },
    filterNode(value, data) {
      if (!value) return true;
      return data.name.toLowerCase().includes(value.toLowerCase());
    },
    onSelectAllGroup(checked) {
      const tree = this.$refs.byGroup;
      if (!tree) return;
      if (checked) {
        tree.setCheckedKeys(this.allGroupLeafKeys);
      } else {
        tree.setCheckedKeys([]);
      }
      this.handleCheckChange();
    },
    onSelectAllControl(checked) {
      const tree = this.$refs.byControl;
      if (!tree) return;
      if (checked) {
        tree.setCheckedKeys(this.allControlLeafKeys);
      } else {
        tree.setCheckedKeys([]);
      }
      this.handleCheckChange();
    },
    handleCheckChange() {
      const currentTree = this.$refs[this.treeMode === 'byGroup' ? 'byGroup' : 'byControl'];
      const anotherTree = this.$refs[this.treeMode !== 'byGroup' ? 'byGroup' : 'byControl'];
      this.editting.doors = currentTree.getCheckedKeys();
      // eslint-disable-next-line no-unused-expressions, babel/no-unused-expressions
      anotherTree?.setCheckedKeys(this.editting.doors, true);

      // 更新全选状态
      this.allCheckedByGroup = this.$refs.byGroup
        ? this.$refs.byGroup.getCheckedKeys().length >= this.allGroupLeafKeys.length && this.allGroupLeafKeys.length > 0
        : false;
      this.allCheckedByControl = this.$refs.byControl
        ? this.$refs.byControl.getCheckedKeys().length >= this.allControlLeafKeys.length && this.allControlLeafKeys.length > 0
        : false;
    },
  },
};
</script>

<style lang="scss" scoped>
.header-like-form-item {
  /deep/ {
    .el-form-item__label {
      padding-right: 0;
      width: 100%;

      &:before {
        display: none !important;
      }
    }

    .el-form-item__error {
      position: relative;
    }
  }
}

.row {
  display: flex;
  gap: 8px;
  margin: 0;
}

.control-select {
  width: 180px;
}

.door-select {
  flex: 1
}

.time-group-select {
  width: 180px;
}

.control-select, .door-select, .time-group-select {
  &:not(.error) /deep/ .el-input:before {
    border-bottom: 1px solid #999 !important;
  }
}

.tabs {
  margin-top: 16px;

  /deep/ {
    .el-tabs__content {
      height: calc(100vh - 400px);
      overflow: auto;
    }
  }
}

.tree-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 2px 0;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 8px;
}

.tree-filter-input {
  flex: 1;
  max-width: 300px;
}
</style>
