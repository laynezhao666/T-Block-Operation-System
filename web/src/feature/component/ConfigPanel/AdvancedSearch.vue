<template>
  <div class="advanced-search-wrap">
    <div
      class="advanced-search-body"
    >
      <el-row>
        <el-form
          ref="form"
          :model="formData"
          label-width="114px"
        >
          <template
            v-for="column in columns"
          >
            <template
              v-for="field in column.fields"
            >
              <el-col
                v-if="isSearch(field) && showItem(field)"
                :key="field.name"
                :span="8"
              >
                <el-form-item
                  :prop="field.name"
                >
                  <span
                    slot="label"
                    :title="field.label"
                  >
                    {{ field.label }}
                  </span>
                  <el-input
                    v-if="getFormItemType(field) === 'text'"
                    v-model="formData[field.name]"
                  />
                  <el-input
                    v-if="getFormItemType(field) === 'num'"
                    v-model="formData[field.name]"
                    type="number"
                  />
                  <el-cascader
                    v-if="getFormItemType(field) === 'cascader'"
                    v-model="formData[field.name]"
                    :options="opts[field.name]"
                    :props="{ multiple: field.isMulti }"
                    collapse-tags
                    style="width:100%"
                    popper-class="custom-cascader"
                    filterable
                    @visible-change="initOpts(field)"
                    @change="onChange($event, field.onChange)"
                  />
                  <el-select
                    v-if="getFormItemType(field) === 'select'"
                    v-model="formData[field.name]"
                    filterable
                    remote
                    multiple
                    reserve-keyword
                    collapse-tags
                    :remote-method="remoteMethod(field)"
                    :loading="loading[field.name]"
                    @focus="initOpts(field)"
                    @change="onChange($event, field.onChange)"
                  >
                    <el-option
                      v-for="(label, value) in opts[field.name]"
                      :key="value"
                      :label="label"
                      :value="field.isIdField?value:label"
                    />
                  </el-select>
                  <el-select
                    v-if="getFormItemType(field) === 'enum'"
                    v-model="formData[field.name]"
                    multiple
                    filterable
                    collapse-tags
                  >
                    <el-option
                      v-for="item in opts[field.name]"
                      :key="item.value"
                      :label="item.label"
                      :value="item.value"
                    />
                  </el-select>
                  <el-checkbox-group
                    v-if="getFormItemType(field) === 'bool'"
                    v-model="formData[field.name]"
                  >
                    <el-checkbox label="1">
                      是
                    </el-checkbox>
                    <el-checkbox label="0">
                      否
                    </el-checkbox>
                  </el-checkbox-group>
                  <user-selector
                    v-if="getFormItemType(field) === 'user'"
                    v-model="formData[field.name]"
                  />
                  <user-selector
                    v-if="getFormItemType(field) === 'mutiUser'"
                    v-model="formData[field.name]"
                    :collapse-tags="true"
                    multiple
                  />
                  <date-picker
                    v-if="getFormItemType(field) === 'date'"
                    v-model="formData[field.name]"
                    type="daterange"
                  />
                  <date-picker
                    v-if="getFormItemType(field) === 'datetime'"
                    v-model="formData[field.name]"
                    type="datetimerange"
                  />
                  <date-picker
                    v-if="getFormItemType(field) === 'month'"
                    v-model="formData[field.name]"
                    type="monthrange"
                  />
                  <el-input
                    v-if="getFormItemType(field) === 'textarea'"
                    v-model="formData[field.name]"
                    type="textarea"
                    :autosize="{ minRows: 1, maxRows: 1 }"
                    placeholder="支持拷贝Excel多单元格批量查询"
                  />
                </el-form-item>
              </el-col>
            </template>
          </template>
        </el-form>
      </el-row>
    </div>
    <div
      class="advanced-search-footer"
    >
      <el-button
        type="text"
        class="text-dark"
        @click="reset"
      >
        重置
      </el-button>
      <el-button
        type="text"
        :auth-right-code="codes.cx"
        @click="search"
      >
        查询
      </el-button>
    </div>
  </div>
</template>
<script>
import { each, isEmpty, find, findLast, isNil, isFunction, cloneDeep } from 'lodash';
import { map, flatten } from 'lodash/fp';
import DatePicker from 'component/DateTimePicker';
import userSelector from 'component/user/user-selector';
import { getSelectNamespace } from 'component/script/configHelper';
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
    userSelector,
  },
  mixins: [configMixin],
  inject: ['configCgi'],
  props: {
    columns: {
      type: Array,
      required: true,
    },
    hasCondition: {
      type: Boolean,
      required: true,
    },
    codes: {
      type: Object,
      required: true,
    },
    roles: {
      type: Object,
      default: () => ({}),
    },
    table: {
      type: String,
      required: true,
    },
    showSearchSetting: {
      type: Boolean,
      required: true,
    },
    isCustomizeLocal: {
      type: Boolean,
      default: false,
    },
    /**
     * 预设的 formData 数据
     */
    presetFormData: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      formData: {},
      opts: {},
      loading: {},
      enumNames: [],
    };
  },
  computed: {
    fields() {
      return this.columns |> map('fields') |> flatten;
    },
  },
  watch: {
    fields: {
      handler(v) {
        if (!this.init) {
          this.initForm();
          this.initializeFormData();
          this.init = true;
          v.forEach((field) => {
            if (field?.fieldEnum?.length) {
              this.$set(this.opts, field.name, this.parseEnum(field.fieldEnum));
            }
          });
        } else if (this.isCustomizeLocal) {
          this.initForm();
        }
      },
      deep: true,
    },
  },
  methods: {
    initForm() {
      const formData = {};
      this.enumNames = [];
      this.fields.forEach((field) => {
        if (this.isSearch(field)) {
          if (['bool', 'enum', 'select', 'cascader'].includes(this.getFormItemType(field))) {
            formData[field.name] = [];
          }
          if (this.getFormItemType(field) === 'enum') {
            this.enumNames.push(field.name);
          }
        }
      });
      this.formData = formData;
      this.refreshState();
    },
    /**
     * 预设 formData 初始值
     * 如果 presetFormData 非空，则合并到 formData，并发起搜索
     */
    initializeFormData() {
      if (!isEmpty(this.presetFormData)) {
        this.formData = Object.assign(this.formData, this.presetFormData);
        this.search();
      }
    },
    refreshState() {
      const { fields } = this;
      if (fields.length) {
        this.readonly = [];
        fields.forEach((field) => {
          if (field.subField) {
            if (!field.showSubField) { // 隐藏被级联项
              if (Array.isArray(field.subField)) {
                field.subField.forEach(field => readonlyify(fields, field));
              } else {
                readonlyify(fields, field.subField);
              }
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
          if (field.readonly && !field.readonlyInModal) {
            this.readonly.push(field.name);
          }
        });
      }
    },
    genListener(field) {
      const fn = (v, subField, showSubField) => {
        this.$set(this.formData, subField, []);
        if (this.opts[subField]) {
          this.$set(this.opts, subField, void 0);
        }
        if (showSubField) return;

        const match = find(this.fields, {
          name: subField,
        });
        if (!match) return;
        let isEqual;
        if (isNil(field.subValue)) {
          isEqual = !!v.length;
        } else if (isFunction(field.subValue)) {
          isEqual = field.subValue(v);
        } else if (Array.isArray(field.subValue)) {
          isEqual = field.subValue.includes(v);
        } else {
          isEqual = v === field.subValue;
        }
        if (isEqual) {
          if (this.readonly.indexOf(match.name) > -1) {
            this.readonly.splice(this.readonly.indexOf(match.name), 1);
          }
        } else {
          if (this.readonly.indexOf(match.name) === -1) {
            this.readonly.push(match.name);

            if (match.subField) { // 多级级联的筛选
              if (Array.isArray(match.subField)) {
                match.subField.forEach(subField => fn(v, subField));
              } else {
                fn(v, match.subField);
              }
            }
          }
        }
      };
      return (v) => {
        const { showSubField } = field;
        if (Array.isArray(field.subField)) {
          field.subField.forEach(subField => fn(v, subField, showSubField));
        } else {
          fn(v, field.subField, showSubField);
        }
      };
    },
    genEmitter(field) {
      return (data, v) => {
        let isEqual;
        if (isNil(field.subValue)) {
          if (field.type === 'bool') { // bool类型只有选“是”才能出现级联项
            isEqual = v === '1';
          } else { // 非bool类型默认直接展示级联项
            isEqual = !!v.length;
          }
        } else if (isFunction(field.subValue)) {
          isEqual = !field.subValue(v);
        } else {
          isEqual = v !== field.subValue;
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
    reset() {
      // 清空下拉选项
      const opts = {};
      Object.keys(this.opts).forEach((name) => { // 下拉枚举值不能被清空
        if (this.enumNames.includes(name)) {
          opts[name] = this.opts[name];
        }
      });
      this.opts = opts;

      // 清空formData
      this.initForm(); // 直接初始化formData
    },
    search() {
      const conditions = [];
      const formData = cloneDeep(this.formData);
      each(formData, (value, key) => {
        if (Array.isArray(value)) {
          // eslint-disable-next-line no-param-reassign
          value = value.filter(Boolean);
        }
        if (isEmpty(value)) return;
        const field = findLast(this.fields, {
          name: key,
        });
        const condition = {
          name: key,
          value,
        };
        const type = this.getFormItemType(field);
        if (type === 'date' || type === 'datetime' || type === 'month') {
          condition.operator = 'between';
        } else if (type === 'text') {
          condition.value = value.split(';');
          condition.operator = field.isIdField ? 'in' : (field.operator || 'like');
        } else if (type === 'num') {
          condition.value = value.split(';');
        } else if (type === 'cascader') {
          if (Array.isArray(value[0])) {
            condition.value = value.map(v => v.slice(-1)[0]);
          } else {
            condition.value = [value.slice(-1)[0]];
          }
        } else if (type === 'mutiUser' || type === 'user') {
          condition.value = value.map(v => v.userUid);
        } else if (type === 'textarea') {
          condition.value = value.trim().replace(/\n/g, ';')
            .split(';')
            .filter(Boolean)
            .map(v => v.trim());
        }
        conditions.push(condition);
      });
      this.$emit('search', conditions);
    },
    isSearch(field) {
      if (field.name === 'devicetypes_name') {
        return !this.hasCondition;
      }
      const isFilter = Array.isArray(field.isFilter) ? field.isFilter.includes(this.table) : field.isFilter;
      return field.isIndex && (!field.name.endsWith('_id') || field.isIdField) && isFilter;
    },
    showItem(field) {
      return !this.readonly.includes(field.name) && (this.showSearchSetting ? field.showIndex : true);
    },
    getFormItemType(field) {
      if (field?.fieldEnum?.length) {
        return 'enum';
      } if (field.type === 'cascader') {
        return 'cascader';
      } if (field.type === 'mutiUser' || field.type === 'user') {
        return field.type;
      } if (field.type === 'string' && field.isIdField) {
        return 'text';
      } if (field.type === 'string' && getSelectNamespace(field.name)) {
        return 'select';
      } if (field.type === 'string' || field.type === 'int64') {
        return 'text';
      } if (field.type === 'int' || field.type === 'float') {
        return 'num';
      }
      return field.type;
    },
    remoteMethod(field) {
      return (keyword) => {
        const method = field.dropdownMethod || 'post';
        const path = (isFunction(field.dropdownPath) ? field.dropdownPath(this.formData, this.configCgi)
          : field.dropdownPath) || this.configCgi.getMgrSelect;
        const query = (field.dropdownQuery && field.dropdownQuery({ ...this.formData, keyword })) || {
          fieldName: getSelectNamespace(field.name),
        };

        this.$set(this.loading, field.name, true);
        this.$axios[method](path, {
          ...query,
          keyword,
        }, false).then((data) => {
          if (data.count >= 0) { // 兼容使用/cgi/asset/get接口的下拉
            const list = data.list.map(item => Object.values(item)[0]);
            this.$set(this.opts, field.name, list);
          } else {
            if (['{}', '[]'].includes(JSON.stringify(data))) {
              this.$message.info('数据为空');
            }
            this.$set(this.opts, field.name, data);
          }
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
    onChange(v, fn) {
      if (fn) {
        return fn.call(this, v);
      }
    },
  },
};
</script>
<style lang="scss" scoped>
@import "~common/style/mixin";
.advanced-search-wrap {
  background-color: white;

  .advanced-search-title {
    line-height: 64px;
    padding: 0 $space-l;
    font-size: $font-size-m;
    font-weight: bold;

    .el-button {
      margin-top: 20px;
    }
  }

  .advanced-search-footer {
    padding: 15px $space-l;
    text-align: right;

    button + button {
      margin-left: 32px;
    }
  }

  .el-form {
    padding: 0 24px;
    &-item {
      padding-right: $space-m;
      padding-left: $space-m;

      /deep/ .el-form-item__label {
        padding-right: $space-m;
        text-overflow: ellipsis;
        white-space: nowrap;
        overflow: hidden;
      }

      /deep/ .el-form-item__content {
        height: 32px;
      }

      .el-date-editor {
        padding: 0;
        height: 24px !important;

        /deep/ .el-range__icon {
          right: 0;
        }
      }
    }
  }
}
</style>

<style lang="scss">
.custom-cascader .el-cascader-node {
  max-width: 200px;
}
input.el-cascader__search-input::placeholder{
  color: transparent !important;
}
</style>
