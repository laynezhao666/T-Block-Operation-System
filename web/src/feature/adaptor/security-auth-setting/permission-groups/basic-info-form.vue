<template>
  <el-form
    ref="form"
    :model="editting"
    :rules="rules"
    label-width="120px"
  >
    <el-form-item
      label="权限组名称"
      prop="name"
      required
    >
      <el-input
        v-model.trim="editting.name"
        placeholder="请输入权限组名称"
      />
    </el-form-item>

    <el-form-item
      label="类型"
      prop="label"
      required
    >
      <el-radio-group
        v-model="editting.label"
      >
        <el-radio
          v-for="(opt, i) in typeOptions"
          :key="i"
          :label="opt.value"
        >
          {{ opt.label }}
        </el-radio>
      </el-radio-group>
    </el-form-item>

    <el-form-item
      label="备注"
      prop="comment"
    >
      <el-input
        v-model="editting.comment"
        :autosize="{ minRows: 2 }"
        placeholder="输入内容"
        type="textarea"
      />
    </el-form-item>
  </el-form>
</template>

<script>
import { ruleCantEmpty } from '../../../component/tedge-components/table-layout-context/form-rules';

export default {
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
    return {
      typeOptions: [{
        label: '公共区域',
        value: '公共区域',
      }, {
        label: 'IT方仓',
        value: 'IT方仓',
      }, {
        label: '仓库',
        value: '仓库',
      }, {
        label: '其他',
        value: '其他',
      }],

      rules: {
        name: [
          ruleCantEmpty('权限组名称'),
        ],
        label: [
          ruleCantEmpty('类型'),
        ],
      },
    };
  },
  methods: {
    async validate() {
      return this.$refs.form.validate();
    },
  },
};
</script>
