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
import mixin from 'component/script/mixin';

import configMixin from './mixin';

export default {
  mixins: [mixin, configMixin],
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
      default: '',
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
            if (column.type === 'bool' && data[column.name]) { // 后端可能会返回空
              value = data[column.name] === '1' ? '是' : '否';
            } else if (column.formatter) {
              value = column.formatter(column, data, data[column.name]);
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
