<template>
  <el-modal
    :visible.sync="visibleData"
  >
    <template slot="title">
      <span v-if="isEdit">
        编辑{{ text }}
      </span>
      <span v-else>
        添加{{ text }}
      </span>
    </template>
    <el-form
      ref="form"
      :rules="rules"
      :model="formData"
      label-width="160px"
      :validate-on-rule-change="false"
    >
      <template v-for="field in fields">
        <el-form-item
          v-if="showItem(field)"
          :key="field.name"
          :prop="field.name"
        >
          <span
            slot="label"
            :title="field.label"
          >
            {{ field.label }}
            <el-help-tip
              v-if="field.tips"
              width="400"
              popper-class="test"
            >
              {{ field.tips }}
            </el-help-tip>
          </span>
          <div
            v-if="getFormItemType(field) === 'readonly' || getFormItemType(field) === 'addInit'"
          >
            {{ formData[field.name] }}
          </div>
          <div
            v-if="getFormItemType(field) === 'pic'"
          >
            <uploadImage
              :pic="{ 'fieldName': field.name, 'picList': formData[field.name] || '' }"
              @updatePic="updatePic"
            />
          </div>
          <el-input
            v-if="getFormItemType(field) === 'textarea'"
            v-model="formData[field.name]"
            type="textarea"
            resize="none"
            :disabled="field.disabled || isDisabled(field)"
            :rows="2"
            :maxlength="field.max"
          />
          <el-input
            v-if="getFormItemType(field) === 'text'"
            v-model="formData[field.name]"
            :maxlength="field.max"
            :disabled="field.disabled || isDisabled(field)"
            :placeholder="field.placeholder"
            @input="onChange($event, field)"
          />
          <el-input
            v-if="getFormItemType(field) === 'num'"
            v-model="formData[field.name]"
            :maxlength="field.max"
            :disabled="field.disabled || isDisabled(field)"
            type="number"
          />
          <el-cascader
            v-if="getFormItemType(field) === 'cascader'"
            v-model="formData[field.name]"
            :options="opts[field.name]"
            collapse-tags
            style="width:100%"
            popper-class="custom-cascader"
            filterable
            clearable
            @visible-change="initOpts(field)"
            @change="onChange($event, field)"
          />
          <el-select
            v-if="getFormItemType(field) === 'idselect'"
            v-model="formData[field.name]"
            clearable
            filterable
            remote
            :disabled="isDisabled(field)"
            :multiple="isMultiple(field)"
            reserve-keyword
            :remote-method="remoteMethod(field)"
            :loading="loading[field.name]"
            @focus="initOpts(field)"
            @change="changeOpts($event, field)"
            @visible-change="handleChangeFlag"
          >
            <el-option
              v-for="(label, value) in opts[field.name]"
              :key="value"
              :label="label"
              :value="value"
            />
          </el-select>
          <el-select
            v-if="getFormItemType(field) === 'nameselect'"
            v-model="formData[field.name]"
            clearable
            filterable
            remote
            :multiple="isMultiple(field)"
            reserve-keyword
            :remote-method="remoteMethod(field)"
            :loading="loading[field.name]"
            @focus="initOpts(field)"
            @change="changeOpts($event, field)"
            @visible-change="handleChangeFlag"
          >
            <el-option
              v-for="label in opts[field.name]"
              :key="label"
              :label="label"
              :value="label"
            />
          </el-select>
          <el-select
            v-if="getFormItemType(field) === 'enum'"
            v-model="formData[field.name]"
            clearable
            @change="onChange($event, field)"
            @visible-change="handleChangeFlag"
          >
            <el-option
              v-for="item in opts[field.name]"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-radio-group
            v-if="getFormItemType(field) === 'bool'"
            v-model="formData[field.name]"
            @change="onChange($event, field)"
            @visible-change="handleChangeFlag"
          >
            <el-radio label="1">
              是
            </el-radio>
            <el-radio label="0">
              否
            </el-radio>
          </el-radio-group>
          <date-picker
            v-if="getFormItemType(field)=== 'date'"
            v-model="formData[field.name]"
            type="date"
          />
          <date-picker
            v-if="getFormItemType(field)=== 'month'"
            v-model="formData[field.name]"
            type="month"
          />
          <date-picker
            v-if="getFormItemType(field)=== 'datetime'"
            v-model="formData[field.name]"
            type="datetime"
          />
        </el-form-item>
      </template>
    </el-form>

    <template slot="footer">
      <el-button
        type="primary"
        @click="save"
      >
        确定
      </el-button>
    </template>
  </el-modal>
</template>
<script>
import { cloneDeep, omit, forEach, omitBy, find, isFunction, isNil, each } from 'lodash';
import { required, common } from 'common/script/form_rules';
import DatePicker from 'component/DateTimePicker';
import uploadImage from 'component/uploadImage';
import mixin from 'component/script/mixin';
import { getSelectNamespace, getNameById } from 'component/script/configHelper';

import configMixin from './mixin';

function readonlyify(fields, field) {
  const match = find(fields, {
    name: field,
  });
  if (match) { match.readonly = true; }
}

export default {
  components: {
    DatePicker,
    uploadImage,
  },
  inject: ['configCgi', 'commonCgi'],
  mixins: [mixin, configMixin],
  props: {
    id: {
      type: String,
      default: '',
    },
    visible: Boolean,
    data: {
      type: Object,
      default: () => ({}),
    },
    columns: {
      type: Array,
      required: true,
    },
    text: {
      type: String,
      required: true,
    },
    table: {
      type: String,
      required: true,
    },
    options: {
      type: Object,
      required: true,
    },
    conditions: {
      type: [Array, undefined],
      default: () => undefined,
    },
    isEdit: Boolean,
    preData: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      changeFlag: false,
      visibleData: this.visible,
      formData: {},
      // formData里只有外键id值，把外键name值也算出来
      relativeFormData: {},
      rules: {},
      opts: {},
      loading: {},
      readonly: [],
      disabled: [],
    };
  },
  computed: {
    fields() {
      return cloneDeep(this.columns[this.columns.length - 1].fields);
    },
    casSonItems() {
      return this.fields.filter(v => v.isCasSon).map(k => k.name);
    },
  },
  watch: {
    preData(v) {
      this.formData = {
        ...this.formData,
        ...v,
      };
      Object.keys(v).forEach((key) => {
        this.disabled.push(key);
      });
      this.triggerInit();
    },
    formData: {
      handler() {
        each(this.formData, (value, key) => {
          if (this.isFK(key)) {
            const nameKey = getNameById(key);
            if (value instanceof Array) {
              this.relativeFormData[nameKey] = value.map(v => this.opts[key][v]);
            } else if (this.opts[key]) {
              this.relativeFormData[nameKey] = this.opts[key][value];
            }
          }
        });
      },
      deep: true,
    },
    columns(v1, v2) {
      const nextTable = v1[v1.length - 1];
      const preTable = v2[v2.length - 1];
      if (nextTable.table !== preTable.table) {
        this.fields = nextTable.fields;
      }
    },
    visible(v) {
      if (this.visibleData !== v) {
        this.visibleData = v;
      }
    },
    visibleData(v) {
      if (v) {
        this.opts = {
          ...this.opts,
          ...this.options,
        };
      } else {
        // 隐藏时重置 form 的内容
        this.opts = omit(this.opts, Object.keys(this.options));
        this.formData = {
          ...this.preData,
        };
        this.$refs.form.clearValidate();
        this.readonly = this.fields.filter(field => field.readonly).map(field => field.name);
      }
      this.$emit('update:visible', v);
    },
    data(v) {
      this.formData = Object.keys(v).length ? cloneDeep(v) : {
        ...this.preData,
      };
      this.fields.forEach((field) => {
        const { notAllowEdit, defaultValue, name } = field;
        if (notAllowEdit && this.isEdit && this.isFK(name)) {
          this.formData[name] = this.opts[name][this.formData[name]];
        }
        if (defaultValue.length && name.endsWith('_id')) {
          this.remoteMethod(field)();
        }
      });
      this.triggerInit();
    },
    fields: {
      handler(v) {
        this.refreshState(v);
      },
      deep: true,
    },
  },
  mounted() {
    this.refreshState(this.fields);
  },
  methods: {
    handleChangeFlag(v) { // 判断当前下拉框是否展示，解决idcdb编辑初始化数据被清空的问题
      this.changeFlag = v;
    },
    refreshState(v) {
      const fields = v || this.fields;
      const rules = {};
      if (fields.length) {
        this.readonly = [];
        fields.forEach((field) => {
          if (field.isRequire) {
            rules[field.name] = [required()];
          }
          if (field.pattern) {
            const pattern = common(field.pattern);
            rules[field.name] = rules[field.name] ? [...rules[field.name], pattern] : [pattern];
          }
          if (field.subField) {
            if (Array.isArray(field.subField)) {
              field.subField.forEach(field => readonlyify(fields, field));
            } else {
              readonlyify(fields, field.subField);
            }
            const listener = this.genListener(field);
            const emitter = this.genEmitter(field);
            // eslint-disable-next-line no-param-reassign
            field.onInit = listener;
            // eslint-disable-next-line no-param-reassign
            field.onChange = listener;
            // eslint-disable-next-line no-param-reassign
            field.onSubmit = emitter;
          }
          if (field.readonly) {
            this.readonly.push(field.name);
          }
          if (field?.fieldEnum?.length) {
            this.$set(this.opts, field.name, this.parseEnum(field.fieldEnum));
          }
        });
      }
      this.rules = rules;
    },
    triggerInit() {
      forEach(this.formData, (v, key) => {
        if (v) {
          const match = find(this.fields, {
            name: key,
          });
          setTimeout(() => {
            if (match?.onInit) {
              match.onInit.call(this, v);
            }
          });
        }
      });
    },
    genListener(field) {
      const fn = (v, subField) => {
        const match = find(this.fields, {
          name: subField,
        });
        if (!match) return;
        let isEqual;
        if (isNil(field.subValue)) { // 如果没有指定subValue
          if (field.type === 'bool') { // bool类型只有选“是”才能出现级联项
            isEqual = v === '1';
          } else { // 非bool类型默认直接展示级联项
            isEqual = !!v.length;
          }
        } else if (isFunction(field.subValue)) {
          isEqual = field.subValue(v);
        } else if (Array.isArray(field.subValue)) {
          isEqual = field.subValue.includes(v);
        } else {
          isEqual = v === field.subValue;
        }

        if (this.changeFlag) { // 下拉触发genListener时，需要清空subField的数据
          this.$set(this.formData, subField, '');
          if (this.opts[subField] && this.getFormItemType(field) !== 'enum') {
            this.$set(this.opts, subField, void 0);
          }
        }

        if (isEqual) {
          if (this.readonly.indexOf(match.name) > -1) {
            this.readonly.splice(this.readonly.indexOf(match.name), 1);
          }
        } else {
          if (this.readonly.indexOf(match.name) === -1) {
            this.readonly.push(match.name);
          }
          if (match.subField) { // 多级级联的筛选
            if (Array.isArray(match.subField)) {
              match.subField.forEach(subField => fn(v, subField));
            } else {
              fn(v, match.subField);
            }
          }
        }
      };
      return (v) => {
        if (Array.isArray(field.subField)) {
          field.subField.forEach(subField => fn(v, subField));
        } else {
          fn(v, field.subField);
        }
      };
    },
    genEmitter(field) {
      return (data, v) => {
        let isEqual;
        if (isNil(field.subValue)) {
          isEqual = v === '0';
        } else if (isFunction(field.subValue)) {
          isEqual = !field.subValue(v);
        } else {
          isEqual = Array.isArray(field.subValue) ? !field.subValue.includes(v) : v !== field.subValue;
        }
        if (isEqual) {
          if (Array.isArray(field.subField)) {
            field.subField.forEach((item) => {
              // eslint-disable-next-line no-param-reassign
              data[item] = '';
            });
          } else {
            // eslint-disable-next-line no-param-reassign
            data[field.subField] = '';
          }
        }
      };
    },
    showItem(field) {
      if (field.notShowItem || field.notShowInForm) {
        return false;
      }
      if (this.getFormItemType(field) === 'addInit') {
        return true;
      }
      if (this.conditions && field.name === 'devicetypes_id') {
        return false;
      }
      if (this.readonly.includes(field.name) && !field.readonlyInSearch) {
        return false;
      }
      if (!this.isEdit) {
        return this.getFormItemType(field) !== 'readonly';
      }
      return !this.isMK(field.name);
    },
    updatePic(v1, v2) {
      this.formData[v2] = v1;
    },
    save() {
      this.$refs.form.validate((valid) => {
        if (valid) {
          let method; let data;
          const primaryFormData = cloneDeep(this.formData);
          const formData = {};
          each(primaryFormData, (value, key) => {
            if (typeof (value) !== 'string') {
              if (value) {
                formData[key] = value.join(';');
              } else {
                formData[key] = '';
              }
            } else {
              formData[key] = value.trim();
            }
          });

          if (this.isEdit) {
            method = this.configCgi.updMgrData;
            each(formData, (value, key) => {
              if (this.isFK(key) && this.opts[key]) {
                if (!parseInt(value, 10)) {
                  if (Object.keys(this.opts[key]).length === 1) {
                    // eslint-disable-next-line prefer-destructuring
                    formData[key] = Object.keys(this.opts[key])[0];
                  } else if (Object.keys(this.opts[key]).length > 20) {
                    formData[key] = '';
                  } else {
                    formData[key] = Object.keys(this.opts[key]).join(';');
                  }
                }
              }
            });

            data = {
              table: this.table,
              data: formData,
              id: formData[this.id],
            };
          } else {
            method = this.configCgi.addMgrData;
            data = {
              table: this.table,
            };
            if (this.conditions) {
              data.data = {
                ...formData,
                [this.conditions[0].name]: this.conditions[0].value[0],
              };
            } else {
              data.data = {
                ...formData,
              };
            }
          }
          // 后端返回了不合法的 ''
          data.data = omitBy(data.data, (v, key) => {
            if (this.isKey(key) && v === '') {
              const match = find(this.fields, {
                name: key,
              });
              if (match?.isRequire) {
                return true;
              }
            }
          });
          this.fields.forEach((field) => {
            if (field.type === 'cascader') {
              data.data[field.name] = (field.notAllowAdd || field.notAllowEdit) ? data.data[field.name]
                : data.data[field.name].substring(data.data[field.name].lastIndexOf(';') + 1);
            }
            if (field.isFrontEndField) {
              delete data.data[field.name];
            }
            if (field.onSubmit) {
              field.onSubmit.call(this, data.data, data.data[field.name]);
            }
          });
          this.$axios.post(method, data).then(() => {
            this.visibleData = false;
            this.$emit('success');
          })
            .catch((err) => {
              this.$message.error(err);
            });
        } else {
          return false;
        }
      });
    },
    isMultiple(field) {
      return field.type === 'mutiInt';
    },
    getFormItemType(field) {
      const namespace = getSelectNamespace(field.name, false);
      if (field.addInit) {
        return 'addInit';
      } if (field.notAllowAdd && !this.isEdit) { // 是否允许新增
        return 'readonly';
      } if (field.notAllowEdit && this.isEdit) { // 是否允许编辑
        return 'readonly';
      } if (field?.fieldEnum?.length) {
        return 'enum';
      } if (field.type === 'cascader') {
        return 'cascader';
      } if (namespace) {
        if (this.isKey(field.name)) {
          return 'idselect';
        }
        return 'nameselect';
      } if (field.type === 'string' || field.type === 'mutistring' || field.type === 'select') {
        return 'text';
      } if (field.type === 'int' || field.type === 'float') {
        return 'num';
      }
      return field.type;
    },
    remoteMethod(field) {
      return (keyword) => {
        const method = field.dropdownMethod || 'post';
        const path = field.dropdownPath || this.configCgi.getMgrSelect;
        const query = (field.dropdownQuery && field.dropdownQuery({ ...this.formData, ...this.relativeFormData })) || {
          fieldName: getSelectNamespace(field.name),
        };
        this.$axios[method](path, {
          ...query,
          keyword,
        }, false).then((data) => {
          this.$set(this.opts, field.name, data);
        })
          .finally(() => {
            this.loading[field.name] = false;
          });
      };
    },
    initOpts(field) {
      if (!this.opts[field.name]) {
        this.remoteMethod(field)();
      }
    },
    changeOpts($event, field) {
      if (!$event.length) { // 选项被清空时，需要重置opts
        this.$set(this.opts, [field.name], void 0);
      }
      if (field.isCasParent) {
        this.getCascader(field);
      }
      if (field.onChange) {
        return field.onChange.call(this, $event);
      }
    },
    getCascader(field) {
      this.$axios.post(this.configCgi.getMgrList, {
        conditions: [{ name: field.name, value: [this.formData[field.name]] }],
        fields: this.casSonItems,
        table: this.table,
      }).then((res) => {
        this.casSonItems.forEach((k) => {
          this.$set(this.formData, k, res.list[0][k]);
        });
      });
    },
    onChange(v, field) {
      if (!v.length) { // 选项被清空时，需要重置opts
        this.$set(this.opts, [field.name], void 0);
      }
      if (field.onChange) {
        field.onChange.call(this, v);
      }
    },
    isDisabled(field) {
      return this.disabled.includes(field.name);
    },
  },
};
</script>

<style lang="scss">
.custom-cascader .el-cascader-node {
  max-width: 200px;
}
</style>
