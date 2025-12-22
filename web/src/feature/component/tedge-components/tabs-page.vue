<template>
  <el-card
    class="tabs-page"
  >
    <el-tabs
      slot="header"
      v-model="actualActiveTabIndex"
    >
      <el-tab-pane
        v-for="(tab, i) in tabs"
        :key="i"
        :label="tab.label"
        :name="i.toString()"
      />
    </el-tabs>

    <component
      :is="actualActiveTab.component"
    />
  </el-card>
</template>

<script>
export default {
  props: {
    tabs: {
      type: Array,
      required: true,
    },
    activeTab: {
      type: Object,
      default() {
        return null;
      },
    },
  },
  data() {
    return {
      uncontrolledActiveTab: this.tabs[0],
    };
  },
  computed: {
    actualActiveTab() {
      return this.activeTab || this.uncontrolledActiveTab;
    },
    actualActiveTabIndex: {
      get() {
        return this.tabs?.indexOf(this.actualActiveTab).toString();
      },
      set(index) {
        const tab = this.tabs[index];
        if (!this.activeTab) {
          this.uncontrolledActiveTab = tab;
        }
        this.$emit('updated:activeTab', tab);
      },
    },
  },
};
</script>

<style lang="scss" scoped>
.tabs-page {
  /deep/ {
    .el-card__header {
      padding: 0;
    }

    .el-card__body {
      padding: 0;
    }

    .el-block {
      box-shadow: none;
    }
  }
}
</style>
