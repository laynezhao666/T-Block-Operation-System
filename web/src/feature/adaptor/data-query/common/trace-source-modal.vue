<template>
  <el-modal
    :visible.sync="visible"
    :title="title"
  >
    <el-form
      v-if="tracingData"
      label-width="120px"
    >
      <el-form-item
        label="表达式"
      >
        <el-input
          :value="tracingData.standard_expression || tracingData.collect_expression"
          :autosize="false"
          :rows="5"
          type="textarea"
          disabled
          class="expression-input"
        />
      </el-form-item>

      <el-form-item
        label="引用测点"
      >
        <el-tree
          :data="treeData"
          class="tree"
        >
          <template #default="{ data }">
            <div class="tree-node">
              {{ data.variable }} | {{ data.name }} | --

              <el-button
                type="text"
                class="tree-node-btn"
                @click="openLocationModal(data)"
              >
                定位
              </el-button>
            </div>
          </template>
        </el-tree>
      </el-form-item>
    </el-form>
    <pre>{{ tracingData }}</pre>
  </el-modal>
</template>

<script>
export default {
  props: {
    point: {
      type: Object,
      default() {
        return null;
      },
    },
  },
  data() {
    return {
      tracingData: null,
    };
  },
  computed: {
    visible: {
      get() {
        return Boolean(this.point);
      },
      set(v) {
        if (!v) {
          this.$emit('update:point', null);
        }
      },
    },
    title() {
      const { point } = this;
      return `溯源【${point?.attrId || '--'}】`;
    },
    treeData() {
      const { tracingData } = this;
      if (!tracingData) return [];
      return [
        ...tracingData.standard_points,
        ...tracingData.collect_points,
      ];
    },
  },
  watch: {
    point(point) {
      if (point) {
        this.loadData();
      }
    },
  },
  mounted() {
    if (this.point) {
      this.loadData();
    }
  },
  methods: {
    async loadData() {
      const tracingData = await this.$axios.post('/api/dcos/tboxmonitor-cgi/trace/point', {
        id: this.point.id,
      });
      this.tracingData = tracingData;
    },
    openLocationModal() {
      // TODO: 等接口改好、ready了继续做
    },
  },
};
</script>

<style lang="scss" scoped>
.expression-input {
  /deep/ textarea {
    background-color: #efefef;
    color: #000 !important;
    padding: 4px 8px;
  }
}

.tree {
  max-height: calc(100vh - 280px);
  overflow: scroll;
}

.tree-node {
  position: relative;
  width: 100%;

  &:hover {
    .tree-node-btn {
      display: block;
    }
  }
}

.tree-node-btn {
  position: absolute;
  right: 16px;
  top: 16px;

  display: none;
}
</style>
