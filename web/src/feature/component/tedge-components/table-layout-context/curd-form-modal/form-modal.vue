<template>
  <el-modal
    :visible.sync="modalVisible"
    :width="formModal.width"
  >
    <template slot="title">
      {{ getTitle() }}
    </template>

    <compnoent
      :is="formModal.formComp"
      v-if="isInited && formModal.formComp"
      ref="form"
      :editting="curd.editting"
      :is-create="isCreate"
    />

    <el-block
      v-if="steps"
      padding
    >
      <el-steps
        :active="activeStepIndex"
      >
        <el-step
          v-for="(step, i) in steps"
          :key="i"
          :title="step.title"
        />
      </el-steps>

      <component
        :is="activeStep && activeStep.comp"
        v-if="activeStep && activeStep.comp"
        ref="form"
        :editting="curd.editting"
        :is-create="isCreate"
      />
    </el-block>

    <template slot="footer">
      <el-button
        v-if="steps && activeStepIndex > 0"
        type="primary"
        @click="prevStep"
      >
        上一步
      </el-button>

      <el-button
        v-if="steps && activeStepIndex < (steps.length - 1)"
        type="primary"
        @click="nextStep"
      >
        下一步
      </el-button>

      <el-button
        v-if="formModal.formComp || activeStepIndex === (steps.length - 1)"
        type="primary"
        @click="submit"
      >
        提交
      </el-button>

      <el-button
        @click="cancel"
      >
        取消
      </el-button>
    </template>
  </el-modal>
</template>

<script>
export default {
  props: {
    tableContext: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {
      activeStepIndex: 0,
      isInited: false,
    };
  },
  computed: {
    curd() {
      return this.tableContext.curd;
    },
    formModal() {
      return this.curd.formModal;
    },
    modalVisible: {
      get() {
        return Boolean(this.curd.editting);
      },
      set(visible) {
        if (!visible) {
          this.activeStepIndex = 0;
          this.formModal.cancel(this.tableContext);
          this.isInited = false;
        }
      },
    },
    isCreate() {
      return this.curd.isCreate;
    },
    steps() {
      return this.formModal.steps;
    },
    activeStep() {
      return this.steps[this.activeStepIndex];
    },
  },
  watch: {
    modalVisible(modalVisible) {
      if (!modalVisible) {
        this.activeStepIndex = 0;
        this.isInited = false;
      }
    },
    'curd.editting': {
      handler() {
        const {
          curd: {
            editting,
          },
          formModal: {
            beforeEdit,
          },
        } = this;

        if (this.isInited || !editting) return;

        if (beforeEdit) {
          beforeEdit(editting, (newEditting) => {
            this.curd.editting = newEditting;
          });
        }

        this.isInited = true;
      },
    },
  },
  methods: {
    getTitle() {
      const { formModal } = this;
      if (typeof formModal.title === 'function') return formModal.title(this.isCreate, this.editting);
      return `${this.isCreate ? '新增' : '编辑'} ${formModal.title}`;
    },
    async prevStep() {
      this.activeStepIndex -= 1;
    },
    async nextStep() {
      const isValid = await this.$refs.form.validate?.();
      if (isValid === false) return;

      this.activeStepIndex += 1;
    },
    async submit() {
      const isValid = await this.$refs.form.validate?.();
      if (isValid === false) return;

      await this.formModal.submit(this.tableContext);

      this.tableContext.forceReloadData();
      this.isInited = false;
    },
    cancel() {
      this.formModal.cancel(this.tableContext);
      this.isInited = false;
    },
  },
};
</script>

<style>

</style>
