<template>
  <el-block>
    <div
      v-if="showExportBtn"
      class="right-btns"
    >
      <el-button
        type="primary"
        :auth-right-code="codes.dc"
        @click.native="download"
      >
        导出全部
      </el-button>
      <slot name="extraArea" />
    </div>
    <el-block
      collapsible
      :collapsed="searchCollapsed"
      inner
    >
      <template slot="header">
        高级筛选
      </template>
      <div class="advanced-search-wrap">
        <div
          class="advanced-search-body"
        >
          <el-form
            ref="form"
            :model="formData"
            label-width="114px"
          >
            <el-row>
              <template
                v-for="column in curColumns"
              >
                <el-col
                  v-if="column.isFilter"
                  :key="column.name"
                  :span="8"
                >
                  <el-form-item
                    :prop="column.name"
                  >
                    <span
                      slot="label"
                      :title="column.label"
                    >
                      {{ column.label }}
                    </span>
                    <el-input
                      v-if="getFormItemType(column) === 'text'"
                      v-model="formData[column.name]"
                      clearable
                      :placeholder="column.placeholder || placeholder"
                      @keyup.enter.native="search"
                    />
                    <el-input
                      v-if="getFormItemType(column) === 'mutiString'"
                      v-model="formData[column.name]"
                      clearable
                      :placeholder="column.placeholder || multiStringPlaceholder"
                      @keyup.enter.native="search"
                    />
                    <el-cascader
                      v-else-if="getFormItemType(column) === 'cascader'"
                      :ref="column.name"
                      v-model="formData[column.name]"
                      :options="opts[column.name]"
                      :props="{ multiple: column.isCascaderMulti || false }"
                      collapse-tags
                      filterable
                      clearable
                      style="width:100%"
                      popper-class="custom-cascader"
                      @visible-change="initOpts(column)"
                      @change="onChange($event, column.onChange)"
                    />
                    <el-select
                      v-else-if="getFormItemType(column) === 'select'"
                      v-model="formData[column.name]"
                      filterable
                      remote
                      clearable
                      :multiple="column.type==='mutiselect'"
                      reserve-keyword
                      collapse-tags
                      :remote-method="remoteMethod(column)"
                      @focus="initOpts(column)"
                      @change="onChange($event, column.onChange)"
                    >
                      <el-option
                        v-for="(label, value) in opts[column.name]"
                        :key="value"
                        :label="label"
                        :value="label"
                      />
                    </el-select>
                    <el-select
                      v-else-if="getFormItemType(column) === 'enum'"
                      v-model="formData[column.name]"
                      filterable
                      clearable
                      :multiple="true"
                      collapse-tags
                    >
                      <el-option
                        v-for="opt in opts[column.name]"
                        :key="opt.value"
                        :label="opt.label"
                        :value="opt.value"
                      />
                    </el-select>
                    <user-selector
                      v-else-if="getFormItemType(column) === 'userselect'"
                      v-model="formData[column.name]"
                      :multiple="column.type === 'mutiuserselect'"
                    />
                    <date-picker
                      v-else-if="getFormItemType(column) === 'datetimerange'"
                      v-model="formData[column.name]"
                      :clearable="true"
                      type="datetimerange"
                    />
                    <date-picker
                      v-else-if="getFormItemType(column) === 'daterange'"
                      v-model="formData[column.name]"
                      :clearable="true"
                      type="daterange"
                    />
                    <el-input
                      v-if="getFormItemType(column) === 'mutilinestext'"
                      v-model="formData[column.name]"
                      type="textarea"
                      :autosize="{ minRows: 1, maxRows: 1 }"
                      placeholder="支持拷贝Excel多单元格批量查询"
                    />
                  </el-form-item>
                </el-col>
              </template>
            </el-row>
          </el-form>
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
            :auth-right-code="codes.cx"
            type="text"
            @click="search"
          >
            查询
          </el-button>
        </div>
      </div>
    </el-block>
  </el-block>
</template>
<script>
import { each, isEmpty, find, forEach, isNil, isFunction } from 'lodash';
import DatePicker from '../DateTimePicker';
import userSelector from '../user/user-selector';

function readonlyify(fields, field) {
  const match = find(fields, {
    name: field,
  });
  match.readonly = true;
}

export default {
  components: {
    DatePicker,
    userSelector,
  },
  inject: ['tableConfig', 'configCgi', 'commonCgi', 'codes', 'roles'],
  props: {
    columns: {
      type: Array,
      required: true,
    },
    flowKey: {
      type: String,
      required: true,
    },
    conditions: {
      type: Array,
      default: () => ([]),
    },
    showExportBtn: {
      type: Boolean,
      required: true,
    },
    searchCollapsed: {
      type: Boolean,
      default: false,
    },
    initFrom: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      formData: {},
      opts: {},
      readonly: [],
      clientWidth: 0,
      placeholder: '请输入',
      multiStringPlaceholder: '支持多个查询，用分号分隔',
    };
  },
  computed: {
    curColumns() {
      return [...this.columns].sort((item1, item2) => item2.filterOrder - item1.filterOrder);
    },
    showMore() {
      const total = this.columns.filter(v => v.isFilter).reduce((acc, cur) => {
        const num = cur.type === 'datetimerange' ? 2 : 1;
        return acc + num;
      }, 0);
      return total > 9;
    },
  },
  watch: {
    columns() {
      this.init();
      this.initForm();
    },
  },
  mounted() {
    this.$nextTick(() => {
      this.clientWidth = this.$refs.form.$el.offsetWidth;
      this.init();
      this.initForm();
    });
  },
  methods: {
    // 解决初始化url带过来form参数
    initForm() {
      if (Object.keys(this.initFrom).length === 0) return;
      Object.keys(this.initFrom).forEach((key) => {
        this.$set(this.formData, key, this.initFrom[key]);
      });
      if (this.curColumns.length !== 0) {
        this.search();
      }
    },
    parseEnum(enumSet) {
      if (typeof (enumSet) === 'string') {
        return enumSet.slice(1, enumSet.length - 1).split(',')
          .map((item) => {
            const arr = item.split('|');
            return {
              value: arr[0],
              label: arr[arr.length - 1],
            };
          });
      }
      return enumSet.map(item => ({
        value: item,
        label: item,
      }));
    },
    init() {
      const formData = {};
      const opts = {};
      this.curColumns.forEach((field) => {
        const { name, isFilter, type, enumSet } = field;
        if (isFilter) {
          if (['muti', 'enum', 'mutiselect'].includes(type)) {
            formData[name] = [];
          }
          if (field.type === 'enum') {
            const options = this.parseEnum(enumSet);
            this.$set(opts, name, options);
          }
        }
      });
      this.opts = opts;
      this.formData = formData;
      this.refreshState();
    },
    refreshState() {
      const fields = this.curColumns;
      if (fields.length) {
        this.readonly = [];
        fields.forEach((field) => {
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
        });
      }
    },
    triggerInit() {
      forEach(this.formData, (v, key) => {
        if (v) {
          const match = find(this.curColumns, {
            name: key,
          });
          if (match && match.onInit) {
            match.onInit.call(this, v);
          }
        }
      });
    },
    genListener(field) {
      const fn = (v, subField) => {
        const match = find(this.curColumns, {
          name: subField,
        });

        let isEqual;
        if (isNil(field.subValue)) {
          isEqual = true;
        } else if (isFunction(field.subValue)) {
          isEqual = field.subValue(v);
        } else if (Array.isArray(field.subValue)) {
          isEqual = field.subValue.includes(v);
        } else {
          isEqual = v === field.subValue;
        }
        if (isEqual && v.length) {
          if (this.readonly.indexOf(match.name) > -1) {
            this.readonly.splice(this.readonly.indexOf(match.name), 1);
          }
        } else {
          if (this.readonly.indexOf(match.name) === -1) {
            this.readonly.push(match.name);

            if (field.isFilter) {
              if (this.getFormItemType(field).includes('muti')) {
                this.$set(this.formData, match.name, []);
                delete this.opts[match.name];
              }
            }
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
          isEqual = !!v.length;
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
      this.opts = {};
      this.init();
    },
    search() {
      const conditions = {};
      // operator: 多选为IN、单选EQUALS、其他LIKE
      each(this.formData, (value, key) => {
        if (Array.isArray(value)) {
          value = value.filter(Boolean);
        }
        if (isEmpty(value)) return;
        const field = find(this.curColumns, {
          name: key,
        });
        const { relationGroup, name, relation } = field;
        if (!conditions[relationGroup]) {
          conditions[relationGroup] = [];
        }
        const base = { field: name, relation };
        const type = this.getFormItemType(field);
        if (['datetimerange', 'daterange'].includes(type)) {
          conditions[relationGroup].push({
            ...base,
            operator: 'GTE',
            value: value[0],
          });
          conditions[relationGroup].push({
            ...base,
            operator: 'LTE',
            value: value[1],
          });
        } else if (type === 'text') {
          conditions[relationGroup].push({
            ...base,
            operator: 'LIKE',
            value: value.replace('；', ';'),
          });
        } else if (type === 'mutiString') {
          conditions[relationGroup].push({
            ...base,
            operator: 'IN',
            value: value.replace('；', ';').split(';'),
          });
        } else if (type === 'cascader') {
          conditions[relationGroup].push({
            ...base,
            operator: 'EQUALS',
            value: this.$refs[key][0].checkedNodes.map(v => v.label).join(';'),
          });
        } else if (type === 'mutiuserselect' || type === 'userselect') {
          conditions[relationGroup].push({
            ...base,
            operator: type === 'mutiuserselect' ? 'IN' : 'LIKE',
            value: type === 'mutiuserselect' ? value.userUid : value.map(v => v.userUid).join(';'),
          });
        } else if (type === 'select') {
          conditions[relationGroup].push({
            ...base,
            operator: field.type === 'mutiselect' ? 'IN' : 'EQUALS',
            value,
          });
        } else if (type === 'enum') {
          conditions[relationGroup].push({
            ...base,
            operator: 'IN',
            value,
          });
        } else if (type === 'mutilinestext') {
          conditions[relationGroup].push({
            ...base,
            operator: 'IN',
            value: value.trim().replace(/\n/g, ';')
              .split(';')
              .filter(Boolean)
              .map(v => v.trim()) });
        }
      });
      this.$emit('search', conditions);
    },
    getFormItemType(column) {
      if (column.type === 'string') {
        return 'text';
      } if (column.type === 'mutiselect') {
        return 'select';
      } if (column.type === 'mutiuserselect') {
        return 'userselect';
      }
      return column.type;
    },
    remoteMethod(field) {
      return (keyword) => {
        const method = field.dropdownMethod || 'post';
        const path = field.dropdownPath || this.configCgi.getOptions;
        const query = (field.dropdownQuery && field.dropdownQuery(this.formData)) || {
          flowKey: this.flowKey,
          field: field.name,
        };
        query.conditions = {
          [field.relationGroup]: [],
        };

        if (keyword && keyword.length) {
          query.conditions[field.relationGroup].push({
            field: field.name,
            operator: 'LIKE',
            value: keyword,
          });
        }

        if (field.subFieldSet.length) {
          field.subFieldSet.forEach((v) => {
            const value = this.formData[v];
            if (value.length) {
              query.conditions[field.relationGroup].push({
                field: v,
                operator: 'EQUALS',
                value,
                relation: 'AND',
              });
            }
          });
        }

        this.$axios[method](path, {
          ...query,
        }, false).then((data) => {
          this.$set(this.opts, field.name, data);
        });
      };
    },
    initOpts(field) {
      // 远程搜索后没有选择数据也需要重新搜索
      const value = this.formData[field.name];
      const initFromKeys = Object.keys(this.initFrom);
      if (!value || !value.length || initFromKeys.includes(field.name)) {
        this.remoteMethod(field)();
      }
    },
    onChange(v, fn) {
      if (fn) {
        return fn.call(this, v);
      }
    },
    download() {
      this.$emit('export');
    },
  },
};
</script>
<style lang="scss" scoped>
.advanced-search-wrap {
  background-color: white;

  .advanced-search-title {
    line-height: 64px;
    padding: 0 24px;
    font-size: 16px;;
    color: rgb(51,51,51);

    .el-button {
      margin-top: 20px;
    }
  }

  .advanced-search-body {
    padding: 0 8px;
  }

  .advanced-search-footer {
    padding: 15px 24px;
    text-align: right;

    button + button {
      margin-left: 32px;
    }

    .dropdown-button {
      margin-left: 0;
      .el-dropdown-link {
          cursor: pointer;
        }
    }

    .text-dark {
      color: #333 !important;
    }
  }

  .el-form {
    padding-top: 0;

    &-item {
      padding-right: 16px;
      padding-left: 16px;

      /deep/ .el-form-item__label {
        padding-right: 16px;
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
.right-btns{
  padding: 0 16px;
  height: 64px;
  line-height: 64px;
  border-bottom: 1px solid #f0f0f0;
  text-align: left;
}
</style>
<style lang="scss">
.custom-cascader .el-cascader-node {
  max-width: 200px;
}
</style>
