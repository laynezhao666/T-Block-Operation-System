<template>
  <div>
    <div
      v-if="isShowSelect"
      class="mask"
      @click="isShowSelect = !isShowSelect"
    />
    <el-popover
      v-model="isShowSelect"
      placement="bottom-start"
      :width="popoverWidth"
      trigger="manual"
      @hide="popoverHide"
    >
      <!-- <el-tabs
        v-if="isShowSelect"
        v-model="activeSearchType"
        class="treeSelect-search-tabs"
        @tab-click="tabClick"
      >
        <el-tab-pane
          label="选择设备"
          name="device"
        />
        <el-tab-pane
          label="选择房间"
          name="room"
        />
      </el-tabs> -->
      <el-input
        v-if="activeSearchType === 'device'"
        v-model="filterText"
        class="c-tree-tools__input"
        placeholder="搜索设备"
        prefix-icon="tn-icon-search"
        clearable
        style="margin: 16px 0 "
      />
      <el-tree
        ref="tree"
        :class="treeClass"
        style="overflow: auto;"
        :data="treeData"
        :load="loadNode"
        :style="style"
        :props="defaultProps"
        :show-checkbox="multiple"
        :node-key="nodeKey"
        :check-strictly="checkStrictly"
        :expand-on-click-node="expandOnClick"
        :default-expanded-keys="defaultExpand"
        :default-checked-keys="defaultDeptAll"
        :filter-node-method="filterNode"
        :current-node-key="currentNodekey"
        :highlight-current="true"
        accordion
        :render-after-expand="true"
        @node-click="handleNodeClick"
        @check-change="handleCheckChange"
      />
      <el-select
        slot="reference"
        ref="select"
        v-model="defaultDeptAll"
        :style="selectStyle"
        :size="size"
        :multiple="multiple"
        :clearable="clearable"
        :collapse-tags="collapseTags"
        class="tree-select"
        @click.native="isShowSelect = !isShowSelect"
        @remove-tag="removeSelectedNodes"
        @clear="removeSelectedNode"
        @change="changeSelectedNodes"
      >
        <el-option
          v-for="item in options"
          :key="item.id"
          :label="item.name"
          :value="item.id"
        />
      </el-select>
    </el-popover>
  </div>
</template>

<script>
import getFilter from './tree-filter';
import { debounce, cloneDeep } from 'lodash';

export default {
  props: {
    getGroupSequence: {
      type: Function,
      default() {},
    },
    defaultProps: {
      type: Object,
      default() {
        return {};
      },
    },
    // 配置是否可多选
    multiple: {
      type: Boolean,
      default() {
        return false;
      },
    },
    clear: {
      type: Boolean,
      default() {
        return false;
      },
    },
    // 配置是否可清空选择
    clearable: {
      type: Boolean,
      default() {
        return false;
      },
    },

    // 配置多选时是否将选中值按文字的形式展示
    collapseTags: {
      type: Boolean,
      default() {
        return true;
      },
    },
    nodeKey: {
      type: String,
      default() {
        return 'id';
      },
    },
    // 显示复选框情况下，是否严格遵循父子不互相关联
    checkStrictly: {
      type: Boolean,
      default() {
        return false;
      },
    },
    // 默认选中的节点key数组
    checkedKeys: {
      type: Array,
      default() {
        return [];
      },
    },
    size: {
      type: String,
      default() {
        return 'medium';
      },
    },
    width: {
      type: String,
      default() {
        return `250px`;
      },
    },
    height: {
      type: String,
      default() {
        return `300px`;
      },
    },
    defaultExpandedKeys: {
      type: Array,
      default() {
        return [];
      },
    },
    maxHeight: {
      type: String,
      default() {
        return `400px`;
      },
    },
  },
  data() {
    return {
      currentNodekey: '',
      filterText: '',
      treeData: [],
      isShowSelect: false, // 是否显示树状选择器
      options: [], // select选中项创建
      selectedData: [], // 选中的节点
      defaultDeptAll: [], // 默认选中的节点数组
      defaultExpand: [],
      popoverWidth: 0,
      firstLoad: false,
      style:
        `width:${
          this.width
        };`
        + `height:${
          this.height
        };max-height:${
          this.maxHeight
        };`,
      selectStyle: `width:${this.width};`,
      checkedIds: [],
      checkedData: [],
      treeDataRoom: [],
      activeSearchType: 'device',
      expandOnClick: true,
      treeClass: '',
    };
  },
  watch: {
    filterText() {
      this.filterData();
    },
    isShowSelect(val) {
      // 隐藏select自带的下拉框
      this.$refs.select.blur();
      if (!val) {
        // this.filterText = '';
      }
      if (val) {
        if (this.treeData.length > 0) {
          this.$nextTick(() => {
            this.selectFirstTemplate();
          });
        }
      }
    },
    defaultExpandedKeys(val) {
      if (!val && this.clear) return;
      this.defaultExpand = val;
    },
    checkedKeys(val) {
      if (!val && this.clear) {}
    },
  },
  mounted() {
    this.getTreeData(() => {
      this.initCheckedData();
    });
  },
  methods: {
    tabClick() {
      this.defaultDeptAll = [];
      this.getTreeData();
      if (this.activeSearchType === 'room') {
        this.treeClass = 'warning-tree-select';
        this.expandOnClick = false;
      } else {
        this.treeClass = '';
        this.expandOnClick = true;
      }
      // this.getTreeData();
    },
    filterNode: debounce((value, data) => {
      if (!value) {
        return true;
      }
      const sv = value.toLowerCase().split(' ');
      let isFound = true;
      const name = data.name.toLowerCase();
      sv.forEach((v) => {
        if (v && isFound) {
          if (name.indexOf(v) === -1) {
            isFound = false;
          }
        }
      });
      return isFound;
    }, 500),
    filterData() {
      const { filterText } = this;

      this.filterMethod = getFilter(this.oriTreeData, {
        children: 'children',
        // 需要匹配的属性值
        matchAttrs: ['name', 'no'],
      });
      this.treeData = this.filterMethod(filterText);
      if (this.treeData.length > 0) {
        this.$nextTick(() => {
          this.selectFirstTemplate();
        });
      }
    },
    // 单选时点击tree节点，设置select选项
    setSelectOption(node) {
      const tmpMap = {};
      tmpMap.value = node.key;
      tmpMap.label = node.label;
      this.options = [];
      this.options.push(tmpMap);
      this.selectedData = node.key;
    },
    // 单选，清空选中
    clearSelectedNode() {
      this.selectedData = [];
      this.$refs.tree.setCurrentKey(null);
    },
    // 多选，清空所有勾选
    clearSelectedNodes() {
      const checkedKeys = this.$refs.tree.getCheckedKeys(); // 所有被选中的节点的 key 所组成的数组数据
      for (let i = 0; i < checkedKeys.length; i++) {
        this.$refs.tree.setChecked(checkedKeys[i], false);
      }
    },
    selectFirstTemplate() {
      const level1 = this.treeData.find(e => e.children && e.children.length);
      if (!level1) {
        return;
      }
      const level2 = level1.children.find(e => e.children && e.children.length);
      if (!level2) {
        return;
      }
      // const data = this.treeData[0];
      const data = level2.children[0];

      this.currentNodekey = data.id;
      this.defaultExpand = [data.id];
      // this.currentNode = data;
    },
    treeTraverse(datas, level) {
      const d = [];
      datas.forEach((i) => {
        let children;
        if (i.children && i.children.length > 0) {
          if (level < 2) {
            children = this.treeTraverse(i.children, level + 1);
          }
        }
        const { id, no, name } = i;
        d.push({
          id,
          name,
          children,
          level,
          no,
        });
      });
      return d;
    },
    getTreeData(cb) {
      const that = this;
      this.getGroupSequence().then((res) => {
        const traverse = function (datas, level) {
          const d = [];
          datas.forEach((i) => {
            let children;
            if (i.children && i.children.length > 0) {
              children = traverse(i.children, level + 1);
            }
            const { id, no, name } = i;
            const disabled = level < 3 && that.activeSearchType === 'device';
            d.push({
              id,
              name,
              children,
              level,
              no,
              disabled,
            });
          });
          return d;
        };
        const treeData = traverse(res, 1);
        // if (this.activeSearchType === 'device') {
        //   treeData = traverse(res, 1);
        // } else {
        //   treeData = this.treeTraverse(res, 1);
        // }
        this.treeData = treeData;
        this.oriTreeData = cloneDeep(treeData);
        this.selectFirstTemplate();
        if (cb) cb();
      });
    },
    initCheckedData() {
      if (this.multiple) {
        // 多选
        this.defaultExpand = this.defaultExpandedKeys;
        this.selectedData = this.checkedKeys;
        this.options = this.checkedKeys;
        this.checkedKeys.map((item) => {
          if (item) {
            this.defaultDeptAll.push(item.id);
          }
        });
      } else {
        // 单选
        if (this.selectedData.length > 0) {
          this.checkSelectedNode(this.selectedData);
        }
      }
      this.$nextTick(() => {
        this.popoverWidth = this.$refs.select.$el.clientWidth - 24;
      });
    },
    popoverHide() {
      this.$emit('popoverHide', this.checkedIds, this.checkedData);
    },
    // 单选，节点被点击时的回调,返回被点击的节点数据
    handleNodeClick(data, node) {
      if (!this.multiple) {
        this.setSelectOption(node);
        this.isShowSelect = !this.isShowSelect;
        this.$emit('change', this.selectedData);
      }
    },
    // 多选，节点勾选状态发生变化时的回调
    handleCheckChange(data, checked) {
      if (checked) {
        const array = this.$refs.tree.getNode(data.id);
        if (array.childNodes.length == 0) {
          this.checkedData.push(data);
          this.defaultDeptAll.push(data.id);
        }
      } else {
        if (this.checkedData.length > 0) {
          this.checkedData.forEach((item, index) => {
            if (item.id == data.id) {
              this.defaultDeptAll.splice(index, 1);
              this.checkedData.splice(index, 1);
            }
          });
        }
      }
      if (this.firstLoad || this.checkedKeys.length == 0) {
        this.setCheckedKey();
      } else {
        this.firstLoad = true;
      }
      this.$emit('change', this.defaultDeptAll, this.defaultExpand);
    },
    // 多选,删除任一select选项的回调
    removeSelectedNodes(val) {
      this.$refs.tree.setChecked(val, false);
      const node = this.$refs.tree.getNode(val);
      if (!this.checkStrictly && node.childNodes.length > 0) {
        this.treeToList(node).map((item) => {
          if (item.childNodes.length <= 0) {
            this.$refs.tree.setChecked(item, false);
          }
        });
      }
      this.$emit('change', this.selectedData);
    },
    treeToList(tree) {
      let queen = [];
      const out = [];
      queen = queen.concat(tree);
      while (queen.length) {
        const first = queen.shift();
        if (first.childNodes) {
          queen = queen.concat(first.childNodes);
        }
        out.push(first);
      }
      return out;
    },
    // // 单选,清空select输入框的回调
    removeSelectedNode() {
      this.clearSelectedNode();
      this.$emit('change', this.selectedData);
    },
    // 选中的select选项改变的回调
    changeSelectedNodes(selectedData) {
      // // 多选,清空select输入框时，清除树勾选
      if (this.multiple && selectedData.length <= 0) {
        this.clearSelectedNodes();
      }
      this.$emit('change', this.selectedData);
    },
    setCheckedKey() {
      const treeData = this.$refs.tree.getCheckedNodes();
      // 过滤全选的bug
      const temp = JSON.parse(JSON.stringify(treeData));
      const result = temp.filter((arr) => {
        const parentId = arr.parent_id;
        const curIdArr = treeData.filter(ele => ele.id == parentId);
        return curIdArr.length == 0;
      });
      const array = this.$refs.tree.getHalfCheckedKeys();
      result.forEach((item) => {
        array.concat(item.parent_id);
      });
      // eslint-disable-next-line no-multi-assign
      this.options = this.selectedData = result;
      this.defaultDeptAll = result.map((item) => {
        if (item.level >= 3) return item.id;
      }).filter(i => i);
      this.defaultExpand = array;
    },
    loadNode(node, resolve) {
      if (node.level === 0) {
        return;
      }
      this.getGroupSequence({ id: node.data.id }).then((res) => {
        if (res.data && res.data.length > 0) {
          resolve(res.data);
          this.$refs.tree.setCheckedKeys(this.defaultDeptAll);
        } else {
          resolve([]);
        }
      });
    },
  },
};
</script>

<style scoped>
.mask {
  width: 100%;
  height: 100%;
  position: fixed;
  top: 0;
  left: 0;
  opacity: 0;
  z-index: 11;
}

.tree-select {
  z-index: 111;
}
</style>

<style lang="scss">
.treeSelect-search-tabs {
  .el-tabs__item {
    padding: 0 10px;
    min-width: 60px;
    height: 32px;
    line-height: 32px
  }
}
.warning-tree-select {
  padding: 10px 0 0 10px;
  .el-tree-node__expand-icon {
    display: none;
  }
}
</style>
