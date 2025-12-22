<template>
  <el-dialog
    :visible.sync="logVisible"
    width="500px"
    @close="close"
  >
    <template
      slot="title"
    >
      新建关注组
    </template>
    <el-form
      ref="form"
      label-width="130px"
      :rules="rules"
      :model="form"
    >
      <el-form-item
        label="关注标题"
        prop="title"
      >
        <el-input v-model="form.title" />
      </el-form-item>
      <el-form-item
        label="模板描述"
        prop="describe"
      >
        <el-input
          v-model="form.describe"
          type="textarea"
        />
      </el-form-item>
      <el-form-item
        label="规则类型"
        prop="hasRule"
      >
        <el-radio
          v-model="form.hasRule"
          :label="false"
        >
          静态规则
        </el-radio>
        <el-radio
          v-model="form.hasRule"
          :label="true"
        >
          动态规则
        </el-radio>
      </el-form-item>
      <el-form-item
        v-if="form.hasRule"
        label="规则定义"
        prop="value"
      >
        <div style="display:flex">
          <div style="width:100px">
            采集值
          </div>
          <el-select
            v-model="form.operator"
            style="width: 150px;text-align:center"
          >
            <el-option value=">" />
            <el-option value="<" />
            <el-option value="=" />
            <el-option value="!=" />
          </el-select>

          <el-input
            v-model="form.value"
            style="margin: 12px 0 12px 10px"
          />
        </div>
      </el-form-item>
    </el-form>
    <template slot="footer">
      <el-button
        type="primary"
        @click="confirm()"
      >
        确定
      </el-button>
    </template>
  </el-dialog>
</template>

<script>
import { required } from 'common/script/form_rules';

export default {
  components: {
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    cgi: {
      type: Object,
      default: () => ({}),
    },
    data: {
      type: Object,
      default: () => {},
    },
    view: {
      type: String,
      default: '',
    },
    status: {
      type: String,
      default: 'add',
    },
  },
  data() {
    return {
      serverAssetIdOptions: [],
      form: {
        title: '',
        operator: '>',
        hasRule: '',
        value: 0,
        describe: '',
      },
      rules: {
        title: [required()],
      },
    };
  },
  computed: {
    logVisible: {
      set(v) {
        this.$emit('update:visible', v);
      },
      get() {
        return this.visible;
      },
    },
  },

  mounted() {
    if (this.data.id) {
      this.form = this.data;
    }
  },
  methods: {
    close() {
      this.logVisible = false;
      this.form = {};
      this.$emit('close');
    },
    logout() {
      console.log(this.tag);
    },
    confirm() {
      this.$refs.form.validate((valid) => {
        if (valid) {
          if (this.data.id) {
            this.form.value = parseFloat(this.form.value);
            this.$axios.post('/cgi/dashboardaux/updateInterestGroup', { id: this.data.id, ...this.form }).then(() => {
              this.$message.success('更新关注组成功');
              this.$emit('confirm');
              this.close();
            });
          } else {
            this.$axios.post('/cgi/dashboardaux/addInterestGroup', {
              ...this.form,
            }, undefined, { isJson: true }).then(() => {
              this.$message.success('新建关注组成功');
              this.$emit('confirm');
              this.close();
            });
          }
        }
      });
    },
  },
};
</script>
<style>

</style>
