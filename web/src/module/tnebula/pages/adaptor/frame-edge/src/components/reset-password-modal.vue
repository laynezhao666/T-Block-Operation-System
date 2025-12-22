<template>
  <el-dialog
    title="修改密码"
    :visible.sync="visible"
  >
    <el-form
      ref="form"
      label-width="8em"
      :model="formData"
      :rules="rules"
    >
      <el-form-item
        label="原密码"
        prop="oldPassword"
      >
        <el-input
          v-model.trim="formData.oldPassword"
          type="password"
          autocomplete="off"
        />
      </el-form-item>
      <el-form-item
        label="新密码"
        prop="newPassword"
      >
        <el-input
          v-model.trim="formData.newPassword"
          type="password"
          autocomplete="off"
        />
      </el-form-item>
      <el-form-item
        label="重复密码"
        prop="repeatPassword"
      >
        <el-input
          v-model.trim="formData.repeatPassword"
          type="password"
          autocomplete="off"
        />
      </el-form-item>
    </el-form>
    <div
      slot="footer"
      class="dialog-footer"
    >
      <el-button
        type="text"
        @click="close"
      >
        取消
      </el-button>
      <el-button
        type="text"
        @click="submit"
      >
        提交
      </el-button>
    </div>
  </el-dialog>
</template>

<script>
export default {
  data() {
    return {
      visible: false,
      formData: {
        oldPassword: '',
        newPassword: '',
        repeatPassword: '',
      },
      currentPassword: '',

      rules: {
        oldPassword: [
          { required: true, message: '请输入旧密码' },
          {
            validator: (rule, value, cbFunc) => {
              if (value !== this.currentPassword) {
                cbFunc('旧密码不匹配，请重新输入');
              } else {
                cbFunc();
              }
            },
          },
        ],
        newPassword: [
          { required: true, message: '请输入新密码' },
          {
            validator: (rule, value, cbFunc) => {
              this.checkNewPasswords(cbFunc, value, false);
            },
          },
        ],
        repeatPassword: [
          { required: true, message: '请重复新密码' },
          {
            validator: (rule, value, cbFunc) => {
              this.checkNewPasswords(cbFunc, value, true);
            },
          },
        ],
      },
    };
  },
  methods: {
    show() {
      this.visible = true;
      this.loadCurrentPassword();
    },
    close() {
      this.visible = false;
    },
    checkNewPasswords(cbFunc, value, isRepeatPwd) {
      const {
        newPassword,
        repeatPassword,
      } = this.formData;

      if (value.length < 6) {
        cbFunc('长度不能小于6');
        return;
      }

      if (!isRepeatPwd) return cbFunc();

      if (newPassword !== repeatPassword) {
        cbFunc('两次输入的密码不一致');
      } else {
        cbFunc();
      }
    },
    async loadCurrentPassword() {
      this.currentPassword = await window.tnwebServices.loginStatusService.fetchRightPassword();
    },
    async submit() {
      if (!(await this.$refs.form.validate())) return;

      const isSuccess = await window.tnwebServices.loginStatusService.resetPassword(this.formData.newPassword);

      if (isSuccess) {
        this.$message.success('修改密码成功');
        this.close();
      } else {
        this.$message.success('修改密码失败，可能网络异常，请稍后重试');
      }
    },
  },
};
</script>
