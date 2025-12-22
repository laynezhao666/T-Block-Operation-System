<template>
  <el-modal :visible.sync="visible">
    <template slot="title">
      报表导出
    </template>

    <el-form
      ref="form"
      :model="form"
      label-width="120px"
    >
      <el-form-item label="选择模板">
        <el-select
          v-model="form.id"
        >
          <el-option
            v-for="(item, index) in templates"
            :key="index"
            :value="item.id"
            :label="item.t_name"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="时间段">
        <el-date-picker
          v-model="form.time"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="yyyy-MM-dd HH:mm:ss"
          :picker-options="pickerOptions"
        />
      </el-form-item>
    </el-form>

    <template slot="footer">
      <el-button
        type="primary"
        :disabled="btnDisabled"
        @click="handleExport"
      >
        导出
      </el-button>
    </template>
  </el-modal>
</template>
<script>
import http from 'common/script/http';
export default {
  props: ['show', 'type'],
  data() {
    return {
      form: {
        time: [],
        id: '',
      },
      templates: [],
      pickerOptions: {
        disabledDate(time) {
          return time.getTime() > Date.now() - 8.64e6;
        },
      },
      visible: false,
    };
  },
  computed: {
    btnDisabled() {
      return this.form.time.length === 0 || !this.form.id;
    },
  },
  watch: {
    show(val) {
      this.visible = val;
      this.form = {
        time: [],
        id: '',
      };
      if (val) {
        http.post('/cgi/dcom/report/data/getTemplates', {
          type: this.type,
        }).then((ret) => {
          this.templates = ret;
        });
      }
    },
    visible(val) {
      if (!val) {
        this.$emit('update:show', val);
      }
    },
  },
  methods: {
    handleExport() {
      const { time, id } = this.form;
      window.open(`/cgi/dcom/report/data/getExcel?id=${id}&startTime=${time[0]}&endTime=${time[1]}&type=${this.type}`);
      this.$emit('update:show', false);
    },
  },
};
</script>
