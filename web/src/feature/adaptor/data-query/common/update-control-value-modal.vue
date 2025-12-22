<template>
  <el-dialog
    :title="title"
    :visible.sync="visible"
    width="360px"
  >
    <el-form
      ref="form"
      v-loading="!pointDefine"
      :model="formData"
      :rules="rules"
      label-width="6em"
    >
      <el-form-item
        v-if="pointDefine"
        label="值:"
        prop="value"
        required
      >
        <template
          slot="label"
        >
          <el-help-tip
            v-if="!pointDefineIsEnum && pointDefine && pointDefine.MpTypeValueRange"
            width="300"
          >
            有效范围：{{ pointDefine.MpTypeValueRange }}
          </el-help-tip>
          值:
        </template>

        <el-select
          v-if="pointDefineIsEnum"
          v-model="formData.value"
          placeholder="请选择"
        >
          <el-option
            v-for="item in pointValueOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>

        <el-input
          v-else
          v-model.number="formData.value"
          :min="pointValueRange && pointValueRange[0]"
          :max="pointValueRange && pointValueRange[1]"
          placeholder="请输入新设定值"
          type="number"
        />
      </el-form-item>
    </el-form>

    <div slot="footer">
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
  props: {
    row: {
      type: Object,
      default() {
        return null;
      },
    },
  },
  data() {
    return {
      formData: {
        value: '',
      },
      pointTypeDefineMap: null,
    };
  },
  computed: {
    visible: {
      get() {
        return !!this.row;
      },
      set(v) {
        if (!v) {
          this.$emit('update:row', null);
        }
      },
    },
    title() {
      const { row } = this;

      if (!row) return '';

      return `设置确认【${row.attrName}】`;
    },
    pointDefine() {
      const {
        pointTypeDefineMap,
        row,
      } = this;

      const {
        attrId,
      } = row || {};

      if (!attrId || !pointTypeDefineMap) return null;

      return pointTypeDefineMap[attrId];
    },
    pointDefineIsEnum() {
      const { pointDefine } = this;
      return pointDefine && [
        '布尔型',
        '枚举型',
      ].includes(pointDefine.MpTypeValueType);
    },
    pointValueOptions() {
      return (this.pointDefine?.MpTypeValueStatus || '')
        .split(',')
        .map((str) => {
          const splited = str.split('=');
          return {
            value: splited[0],
            label: splited[1],
          };
        });
    },
    pointValueRange() {
      if (!this.pointDefine?.MpTypeValueRange) return null;
      const rangesStr = this.pointDefine?.MpTypeValueRange.split('~');

      return rangesStr.length ? [
        Number(rangesStr[0]),
        Number(rangesStr[1]),
      ] : null;
    },
    rules() {
      const required = {
        required: true,
        message: '不能为空',
      };

      const {
        pointValueRange,
      } = this;

      return {
        value: pointValueRange ? [
          required, {
            min: pointValueRange[0],
            max: pointValueRange[1],
            type: 'number',
            message: `有效范围为：${pointValueRange[0]}~${pointValueRange[1]}`,
          }] : [required],
      };
    },
  },
  watch: {
    row: {
      immediate: true,
      handler() {
        if (!this.row) return;

        this.formData = {
          value: this.row.value === '--'
            ? ''
            : Number(this.row.value),
        };

        if (this.row) {
          this.initByRow();
        }
      },
    },
  },
  methods: {
    async initByRow() {
      if (!this.pointTypeDefineMap) {
        const pointTypeDefineList = await this.$axios.get('/cgi/gidmapping/getAllMPType');
        this.pointTypeDefineMap = _.mapKeys(pointTypeDefineList, 'MpTypeIdentifier');
      }
    },
    close() {
      this.visible = false;
    },
    async submit() {
      if (!(await this.$refs.form.validate())) return;

      await this.$axios.post('/api/dcos/tboxmonitor-cgi/control/point', {
        point_id: `${this.row.gid}.${this.row.attrId}`,
        value: String(this.formData.value),
      });

      this.$emit('saved', String(this.formData.value));

      this.close();
    },
  },
};
</script>
