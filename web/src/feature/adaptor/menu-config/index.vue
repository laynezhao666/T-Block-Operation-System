<template>
  <admin-limit-content
    class="menu-config-page"
    @adminLoginedChange="handleAdminLoginedChange"
  >
    <el-block
      no-padding
      header-border
    >
      <template #header>
        菜单
      </template>

      <template slot="extra">
        <el-button
          size="mini"
          type="icon"
          icon="tn-icon-circle-add"
          @click.stop="addRoot()"
        />

        <el-dropdown
          @command="handleMoreAction"
        >
          <el-space :size="4">
            <i class="tn-icon-more tree-more-actions-btn" />
          </el-space>
          <el-dropdown-menu slot="dropdown">
            <el-dropdown-item command="export">
              导出
            </el-dropdown-item>
            <el-dropdown-item command="import">
              导入并替换
            </el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown>
      </template>

      <div class="filters-container">
        <el-select
          v-model="currentSystem"
          placeholder="请选择所属系统"
          default-first-option
          size="small"
          border-type="bordered"
          class="system-select"
        >
          <el-option
            label="PC端"
            value="default"
          />
          <el-option
            label="PAD端"
            value="pad"
          />
        </el-select>

        <el-input
          v-model="filters.keywords"
          placeholder="按关键字搜索"
          suffix-icon="tn-icon-search"
          border-type="bordered"
          size="small"
        />
      </div>

      <div class="tree-container">
        <el-tree
          ref="tree"
          :data="menuTreeData"
          :allow-drop="allowDrop"
          :allow-drag="allowDrag"
          :filter-node-method="filterMenuTreeNode"
          draggable
          @node-drop="handleDrop"
          @node-click="handleNodeClick"
        >
          <div
            slot-scope="{ node, data }"
            :class="{ active: edittingMenuItem && edittingMenuItem.id === data.id }"
            class="menu-item"
          >
            <div class="menu-item-name">
              {{ data.title }}
            </div>

            <div class="menu-item-oprs">
              <el-button
                v-if="data.level < 4"
                size="mini"
                type="icon"
                icon="tn-icon-circle-add"
                @click.stop="addChild(data, node)"
              />

              <el-popconfirm
                title="是否确认删除该菜单？"
                @onConfirm="remove(data, node)"
              >
                <el-button
                  slot="reference"
                  size="mini"
                  type="icon"
                  icon="tn-icon-delete"
                />
              </el-popconfirm>
            </div>
          </div>
        </el-tree>
      </div>
    </el-block>

    <div class="divider-vertical" />

    <el-block
      header="设置"
      padding
      header-border
      class="config-form-container"
    >
      <menu-item-config-form
        v-if="edittingMenuItem"
        :menu-item="edittingMenuItem"
        @submit="handleFormSubmit"
        @cancel="handleFormCancel"
      />
    </el-block>
  </admin-limit-content>
</template>

<script>
import { customMenuToTree } from './utils/custom-menu';
import MenuItemConfigForm from './components/menu-item-config-form';
import { computeDeep, forEachTreeNode } from '../../../utils/tree';
import { exportElTable } from 'utils/xlsx';
import { downloadByObject, downloadByUrl } from 'utils/download';
import { selectAndReadAsTextFile, selectAndReadStringFile, selectFile } from 'utils/fp-dom';
import AdminLimitContent from 'feature/component/tedge-components/admin-limit-content.vue';

export default {
  components: {
    MenuItemConfigForm,
    AdminLimitContent,
  },
  data() {
    return {
      currentSystem: 'default',
      filters: {
        keywords: '',
      },

      menuTreeData: [],
      edittingMenuItem: null,
      edittingMenuItemParent: null,
    };
  },
  watch: {
    currentSystem() {
      this.triggerFilter();
    },
    filters: {
      deep: true,
      handler() {
        this.triggerFilter();
      },
    },
  },
  created() {
    this.loadMenuItems();
  },
  methods: {
    fetchMenuItems() {
      return this.$axios.get('/cgi/tedge-bff/menu/list');
    },
    triggerFilter() {
      if (!this.$refs.tree) return;

      this.$refs.tree.filter(this.currentSystem);
    },
    async loadMenuItems() {
      const menuItems = await this.fetchMenuItems();
      this.menuTreeData = customMenuToTree(_.orderBy(menuItems, item => (item.order || 0)));

      setTimeout(() => {
        this.triggerFilter();
      }, 100);
    },
    async remove(menuItem, node) {
      await this.$axios.delete('/cgi/tedge-bff/menu/delete', {
        id: menuItem.id,
      });
      this.edittingMenuItem = node.parent.data instanceof Array ? node.parent.data[0] : node.parent.data;
      node.remove();
      this.$message.success('删除成功');
    },
    addRoot() {
      this.handleFormCancel();

      const menuItem = {
        belongSystem: this.currentSystem,
        title: '--',
        href: '',
        menuCode: '',
        parentMenuCode: null,
      };
      this.menuTreeData.push(menuItem);
      this.edittingMenuItem = menuItem;
    },
    addChild(menuItem, node) {
      const child = {
        belongSystem: this.currentSystem,
        title: '--',
        href: '',
        menuCode: '',
        parentMenuCode: menuItem.menuCode,
      };
      menuItem.children.push(child);
      this.edittingMenuItem = child;
      this.edittingMenuItemParent = menuItem;
      node.expand();
    },
    async saveMenuItem(menuItem) {
      return this.$axios.post(`/cgi/tedge-bff/menu/${menuItem.id ? 'update' : 'create'}`, {
        ...menuItem,
        parent: undefined,
        children: undefined,
        level: undefined,
      });
    },
    allowDrop(draggingNode, dropNode) {
      const draggingNodeDeep = computeDeep([draggingNode.data]);
      const dropNodeLevel = dropNode.data.level;

      return (draggingNodeDeep + dropNodeLevel) <= 3;
    },
    allowDrag(draggingNode) {
      return computeDeep([draggingNode.data]) < 3;
    },
    filterMenuTreeNode(v, data) {
      const {
        currentSystem,
        filters: {
          keywords,
        },
      } = this;

      if (!currentSystem && !keywords.trim()) return true;

      return data.belongSystem === currentSystem
        && (!data.title || data.title?.includes(keywords.trim()));
    },
    handleDrop(draggingNode, dropNode, position) {
      const draggingData = draggingNode.data;
      const dropData = dropNode.data;

      if (['before', 'after'].includes(position)) {
        draggingData.parentMenuCode = dropData.parentMenuCode;
        draggingData.parent = dropData.parent;
      } else {
        draggingData.parentMenuCode = dropData.menuCode;
        draggingData.parent = dropData;
      }
      this.saveMenuItem(draggingNode.data)
        .then(() => {
          this.reorderAndSave();
        });
      // eslint-disable-next-line no-param-reassign
      // draggingNode.data.parentMenuCode = dropNode.data.menuCode;
      // this.saveMenuItem(draggingNode.data);
    },
    async reorderAndSave() {
      const { menuTreeData } = this;

      const ordersChangeMap = {};

      forEachTreeNode(menuTreeData, (node, parent, indexInParent) => {
        if (node.order === (indexInParent + 1)) return;

        ordersChangeMap[node.id] = indexInParent + 1;
      });

      await this.$axios.post('/cgi/tedge-bff/menu/batchSetOrders', ordersChangeMap);
    },
    async exportMenus() {
      const menuItems = await this.fetchMenuItems();
      downloadByObject(menuItems, '边端动环菜单.json');
    },
    async importMenus() {
      const jsonText = await selectAndReadAsTextFile('.json');

      let json;

      try {
        json = JSON.parse(jsonText);
        if (!(json instanceof Array)) {
          throw new Error('文件内容不是数组');
        }
      } catch (err) {
        console.error(err);
        this.$message.error('文件格式不正确');
      }

      await this.$axios.post('/cgi/tedge-bff/menu/importByJson', {
        menuItems: json,
        mode: 'replace-all',
      });

      this.$message.success('上传成功');
    },
    handleNodeClick(menuItem, node) {
      if (!this.edittingMenuItem?._id) {
        this.handleFormCancel();
      }
      this.edittingMenuItem = menuItem;
      this.edittingMenuItemParent = menuItem.data;
    },
    handleFormCancel() {
      if (!this.edittingMenuItem) return;
      if (!this.edittingMenuItem?._id && this.edittingMenuItemParent) {
        const index = _.findIndex(this.edittingMenuItemParent.children, item => item === this.edittingMenuItem);
        this.edittingMenuItemParent.children.splice(index, 1);
      }
      this.edittingMenuItem = null;
    },
    async handleFormSubmit(newMenuItemData) {
      const isCreate = !newMenuItemData.id;
      const savedMenuItem = await this.saveMenuItem(newMenuItemData);
      Object.assign(this.edittingMenuItem, savedMenuItem);

      if (isCreate) {
        await this.reorderAndSave();
      }

      this.$message.success('保存成功');
    },
    handleMoreAction(command) {
      ({
        export: this.exportMenus.bind(this),
        import: this.importMenus.bind(this),
      })[command]();
    },
    handleAdminLoginedChange() {
      setTimeout(() => {
        this.triggerFilter();
      }, 200);
    },
  },
};
</script>

<style lang="scss" scoped>
.filters-container {
  display: flex;
}

.system-select {
  width: 120px;
}

.menu-config-page {
  display: flex;
  width: 100%;
  height: 100%;

  /deep/ {
    .el-block__body, .el-block__body-inner {
      height: auto;
    }
  }
}

.tree-container {
  width: 360px;
  padding-right: 8px;
  height: calc(100vh - 196px);
  overflow: auto;

  /deep/ {
    .el-tree-node__label {
      flex: 1;
    }
  }
}

.config-form-container {
  flex: 1;
}

.menu-item {
  display: flex;
  width: 100%;
  padding: 0 4px;

  &.active {
    background-color: #1470cc11;
  }

  &-name {
    flex: 1;
    line-height: 24px;
  }
}

.white-space {
  width: 8px;
}

.divider-vertical {
  border-right: 1px solid var(--td-border-level-1-color);
}

.tree-more-actions-btn {
  position: relative;
  top: 6px;
}
</style>
