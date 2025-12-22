<template>
  <el-modal
    custom-layout
    :visible.sync="visibleData"
  >
    <template slot="title">
      {{ data[main] }}
    </template>

    <el-table
      :data="list"
      style="width: 100%;"
    >
      <el-table-column
        prop="name"
        label="字段"
        width="200"
      />
      <el-table-column
        prop="value"
        label="内容"
      />
    </el-table>
  </el-modal>
</template>
<script>
import mixin from '../script/mixin';

export default {
  mixins: [mixin],
  props: {
    visible: Boolean,
    data: {
      type: Object,
      default: () => ({}),
    },
    columns: {
      type: Array,
      required: true,
    },
    main: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      visibleData: this.visible,
      list: [],
    };
  },
  watch: {
    visible(v) {
      if (this.visibleData !== v) {
        this.visibleData = v;
      }
    },
    visibleData(v) {
      if (v) {
        const { data } = this;
        // eslint-disable-next-line array-callback-return
        this.list = this.columns.map((column) => {
          if (column.label && column.name !== 'actions') {
            let value;
            if (column.type === 'bool') {
              value = data[column.name] === '1' ? '是' : '否';
            } else {
              value = data[column.name];
            }

            return {
              name: column.label,
              value,
            };
          }
        }).filter(Boolean);
      }
      this.$emit('update:visible', v);
    },
  },
};
</script>
