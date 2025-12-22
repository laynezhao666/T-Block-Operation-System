<template>
  <admin-limit-tooltips>
    <span slot-scope="{ hasRight }">
      <el-button
        :disabled="!hasRight"
        type="text"
        icon="tn-icon-delete"
        @click="doRemove(props)"
      >
        删除所选
      </el-button>
    </span>
  </admin-limit-tooltips>
</template>

<script>
import AdminLimitTooltips from 'feature/component/tedge-components/admin-limit-tooltips.vue';

export default {
  components: {
    AdminLimitTooltips,
  },
  props: {
    tableContext: {
      type: Object,
      required: true,
    },
    selection: {
      type: Object,
      required: true,
    },
    getSelectedRows: {
      type: Function,
      required: true,
    },
  },
  methods: {
    doRemove() {
      const {
        getSelectedRows,
        tableContext,
      } = this;

      const rows = getSelectedRows();
      if (!tableContext.curd?.remove?.batchRemove) {
        console.error('请先注册curd及批量删除操作，或自定义删除组件');
        return;
      }
      tableContext.curd.remove.batchRemove(rows);
    },
  },
};
</script>
