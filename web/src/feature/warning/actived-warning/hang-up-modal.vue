<template>
  <el-modal :visible.sync="modalVisible">
    <template slot="title">
      告警挂起
    </template>

    <el-form
      ref="form"
      :model="formData"
      label-width="120px"
    >
      <el-form-item
        label="挂起原因"
        prop="hangupReason"
      >
        <el-select
          v-model="formData.hangupReason"
          placeholder="请选择挂起原因"
        >
          <el-option
            v-for="(reason, i) in reasonList"
            :key="i"
            :label="reason.label"
            :value="reason.value"
          />
        </el-select>
      </el-form-item>
    </el-form>

    <template slot="footer">
      <el-button
        type="primary"
        @click="handleSubmit"
      >
        提交
      </el-button>
    </template>
  </el-modal>
</template>

<script>
export default {
  props: {
    visible: {
      type: Boolean,
      required: true,
    },
  },
  data() {
    return {
      formData: {
        hangupReason: '',
      },
      reasonList: [{
        label: '设备故障',
        value: '设备故障',
      }, {
        label: '设备维护',
        value: '设备维护',
      }, {
        label: '应急演练',
        value: '应急演练',
      }, {
        label: '无效告警',
        value: '无效告警',
      }, {
        label: '其他原因',
        value: '其他原因',
      }],
    };
  },
  computed: {
    modalVisible: {
      get() {
        return !!this.visible;
      },
      set(visible) {
        if (!visible) {
          this.$emit('close');
        }
      },
    },
  },
  methods: {
    async handleSubmit() {
      const isValid = await this.$refs.form.validate();
      if (!isValid) return;

      this.$emit('submit', this.formData);
    },
  },
};
</script>

<style>

</style>
