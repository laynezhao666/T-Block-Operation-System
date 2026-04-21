<template>
  <el-dialog
    :visible.sync="visible"
    title="分配门禁"
    append-to-body
  >
    <el-form
      ref="form"
      :model="formModel"
      :rules="rules"
      label-width="120px"
    >
      <el-form-item
        label="领卡人员"
        prop="staff"
      >
        <el-select
          v-model="formModel.staff"
          value-key="id"
          filterable
        >
          <el-option
            v-for="(staff, i) in staffs"
            :key="i"
            :label="staff.name"
            :value="staff"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        label="单位"
      >
        {{ getStaffInfo('company') }}
      </el-form-item>

      <el-form-item
        label="证件"
      >
        {{ getStaffInfo('paper_type') }} / {{ getStaffInfo('paper') }}
      </el-form-item>

      <el-form-item
        label="手机"
      >
        {{ getStaffInfo('phone') }}
      </el-form-item>
    </el-form>

    <span
      slot="footer"
      class="dialog-footer"
    >
      <el-button @click="close">取消</el-button>
      <el-button
        type="primary"
        @click="submit"
      >
        确定
      </el-button>
    </span>
  </el-dialog>
</template>

<script>
import _ from 'lodash';
import { axiosPut } from '../../../../utils/axios-methods';

export default {
  data() {
    return {
      doorCard: null,
      callback: null,
      formModel: {
        staff: null,
      },
      staffs: null,

      rules: {
        staff: [
          { required: true, message: '领卡人员不能为空' },
        ],
      },
    };
  },
  computed: {
    visible: {
      get() {
        return !!this.doorCard;
      },
      set(visible) {
        if (!visible) {
          this.doorCard = null;
        }
      },
    },
  },
  methods: {
    getStaffInfo(path) {
      const data = _.get(this.formModel.staff, path);
      return _.isNil(data) || data === '' ? '--' : data;
    },

    show(doorCard, callback) {
      this.doorCard = doorCard;
      this.formModel.staff = doorCard.staff;
      this.callback = callback;

      if (!this.staffs) {
        this.loadStaffs();
      }
    },
    async loadStaffs() {
      const url = '/api/dcos/tdac-cgi/staffs';
      const resp = await this.$axios.get(url, {
        offset: 0,
        limit: 100000,
      });
      this.staffs = resp.list;
    },
    async grantToStaff() {
      const cardId = this.doorCard.card_no;
      const staffId = this.formModel.staff.id;
      await axiosPut('/api/dcos/tdac-cgi/card/staff', {
        card: cardId,
        staff: staffId,
      });
    },
    async submit() {
      if (!(await this.$refs.form.validate())) return;

      await this.grantToStaff();

      this.$message.success('分配成功');
      this.callback();
      this.close();
    },
    close() {
      this.doorCard = null;
      this.callback = null;
      this.form = {
        staff: null,
      };
      this.$refs.form.clearValidate();
    },
  },
};
</script>
