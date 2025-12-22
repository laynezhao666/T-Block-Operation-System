<template>
  <el-popover
    v-model="showFilter"
    popper-class="list-popover filter-popover"
    placement="bottom-end"
    width="800"
    :offset="16"
  >
    <el-tabs
      v-model="activeName"
    >
      <el-tab-pane
        v-for="tab in tabs"
        :key="tab.name"
        :label="`自定义${tab.label}内容`"
        :name="tab.name"
        lazy
      >
        <div>
          <div class="filter-popover-content">
            <el-row>
              <template
                v-for="field in showColumns"
              >
                <template v-if="tab.name === 'table'">
                  <el-col
                    v-if="!field.hide && !field.notShowItem"
                    :key="field.label"
                    :span="6"
                  >
                    <el-checkbox
                      v-model="field.showInTable"
                      :disabled="field.fixed === true || field.fixed === 'right'"
                    />
                    <span :title="field.label">{{ field.label }}</span>
                  </el-col>
                </template>
                <template v-else-if="tab.name === 'search'">
                  <el-col
                    v-if="field.isFilter"
                    :key="field.label"
                    :span="6"
                  >
                    <el-checkbox
                      v-model="field.showInSearch"
                      :disabled="field.fixed === true || field.fixed === 'right'"
                    />
                    <span :title="field.label">{{ field.label }}</span>
                  </el-col>
                </template>
              </template>
            </el-row>
          </div>
          <div class="popover-footer">
            <el-button
              type="text"
              class="text-dark"
              @click="cancel"
            >
              取消
            </el-button>
            <el-button
              type="text"
              class="text-dark"
              @click="apply"
            >
              应用
            </el-button>
            <el-button
              type="text"
              @click="save"
            >
              保存
            </el-button>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
    <i
      slot="reference"
      class="tn-icon-filter"
    />
  </el-popover>
</template>
<script>
import { cloneDeep } from 'lodash';
import mixin from '../script/mixin';

export default {
  mixins: [mixin],
  props: {
    columns: {
      type: Array,
      required: true,
    },
  },
  data() {
    return {
      showFilter: false,
      activeName: 'table',
      showColumns: cloneDeep(this.columns),
      showSearch: cloneDeep(this.searchcolumns),
      tabs: [{
        label: '展示',
        name: 'table',
      }, {
        label: '查询',
        name: 'search',
      }],
    };
  },
  watch: {
    columns() {
      this.reset();
    },
  },
  methods: {
    reset() {
      this.showColumns = cloneDeep(this.columns);
    },
    apply() {
      this.$emit('change', this.showColumns);
      this.showFilter = false;
      this.reset();
    },
    save() {
      this.$emit('save', this.showColumns);
      this.apply();
    },
    cancel() {
      this.showFilter = false;
      this.reset();
    },
  },
};
</script>
<style lang="scss">
@import "../style/common";

.filter-popover {
  padding: 0 !important;
  margin: 0 !important;

  .el-popover__title {
    padding: 24px;
    margin: 0;
  }

  .filter-popover-content {
    max-height: 600px;
    overflow-y: auto;
  }

  .el-row {
    padding: $space-xs 0;
    margin: 0;
    border-top: 1px solid #f0f0f0;

    .el-col {
      padding: 0 24px;
      height: 40px;
      line-height: 40px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .el-checkbox {
    margin-right: $space-xs;
  }

  .popover-footer {
    padding: $space-m;
    text-align: right;
    border-top: 1px solid #f0f0f0;
  }
}

/deep/.el-tabs__nav-scroll .is-top{
  width: 138px!important;
}
</style>
