<template>
  <el-form
    ref="form"
    label-position="top"
    :model="editting"
    :rules="rules"
  >
    <el-form-item
      prop="time_group_no"
      required
      label-width="100%"
      class="header-like-form-item"
    >
      <split-header-bar
        slot="label"
        title="设置时间组（单选）"
        no-padding
      />

      <tedge-table-layout
        :context="tableLayoutContext"
        class="time-group-select-table"
      >
        <template #columns>
          <time-group-columns />
        </template>
      </tedge-table-layout>
    </el-form-item>
  </el-form>
</template>

<script>
import SplitHeaderBar from '../../../component/tedge-components/split-header-bar.vue';
import TedgeTableLayout from '../../../component/tedge-components/tedge-table-layout.vue';
import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';
import TimeGroupColumns from '../../security-time-period-setting/time-period/time-group-columns.vue';

export default {
  components: {
    SplitHeaderBar,
    TedgeTableLayout,
    TimeGroupColumns,
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
    window.tgs = this;
    return {
      tableLayoutContext: this.createTableLayoutContext(),

      rules: {
        time_group_no: [
          {
            validator(rule, value, cb) {
              cb(!value && value !== 0 ? '时间组不能为空' : undefined);
            },
          },
        ],
      },
    };
  },
  methods: {
    async validate() {
      return this.$refs.form.validate();
    },
    fetchTimeGroups() {
      return this.$axios.get('/api/dcos/tdac-cgi/time-groups');
    },
    createTableLayoutContext() {
      return chainTableLayout(this.fetchTimeGroups.bind(this))
        .tableStyle({
          stripe: true,
          size: 'small',
          height: window.innerHeight - 296,
        })
        .hideToolbar()
        .indexColumn({
          label: '序号',
        })
        .radioRowSelect({
          title: '选择',
          identify: row => row.group_no,
          value: this.editting.time_group_no,
          onChange: (value) => {
            this.$set(this.editting, 'time_group_no', value);
          },
        })
        .done();
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
      position: absolute;
      top: -24px;
      left: 50%;
      transform: translateX(-50%);
    }
  }
}

.time-group-select-table {
  margin-top: 8px;

  /deep/  {
    th {
      height: 24px !important;
      line-height: 24px !important;
    }
  }
}
</style>
