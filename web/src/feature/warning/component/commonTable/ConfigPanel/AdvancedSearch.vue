<template>
  <div class="advanced-search-wrap">
    <div class="advanced-search-title">
      高级筛选
      <el-button
        type="text"
        class="pull-right"
        @click="toggle"
      >
        {{ expand ? '收起': '展开' }}
        <i
          class="tn-icon-arrow-up"
          :class="{'is-rotated': !expand}"
        />
      </el-button>
    </div>
    <div
      v-show="expand"
      class="advanced-search-body"
      :class="{ 'expand' : expand }"
    >
      <el-row id="advanceRow">
        <el-form
          ref="form"
          :model="formData"
          label-width="114px"
        >
          <template
            v-for="column in columns"
          >
            <el-col
              v-if="isSearch(column) && showItem(column)"
              :key="column.name"
              :span=" clientWidth < 800 ? 12: getFormItemType(column) === 'datetime' ? 8 : 8"
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
                  :placeholder="column.placeholder || placeholder"
                />
                <el-input
                  v-if="getFormItemType(column) === 'pureStringNotArray'"
                  v-model="formData[column.name]"
                  :placeholder="column.placeholder || placeholder"
                />
                <el-input
                  v-if="getFormItemType(column) === 'num'"
                  v-model="formData[column.name]"
                  :placeholder="column.placeholder || placeholder"
                  type="number"
                />
                <el-cascader
                  v-if="getFormItemType(column) === 'cascader'"
                  :ref="column.name"
                  v-model="formData[column.name]"
                  :options="opts[column.name]"
                  :show-all-levels="false"
                  :props="{ multiple: column.isCascaderMulti || false,
                            label: column.cascaderLabel || 'label' ,
                            value: column.cascaderValue || 'value',
                            lazy:column.lazy,
                            lazyLoad($event,commonAxios){
                              return column.lazyLoad($event,commonAxios)
                            }}"
                  collapse-tags
                  filterable
                  style="width:100%"
                  popper-class="custom-cascader"
                  @visible-change="initOpts($event,column)"
                  @change="onChange($event, column.onChange)"
                />
                <select-tree
                  v-if="getFormItemType(column) === 'treeSelect'"
                  ref="treeSelect"
                  :width="`100%`"
                  height="500px"
                  :max-height="`400px`"
                  node-key="id"
                  size=""
                  multiple
                  clearable
                  :default-props="{
                    children: 'children',
                    label: 'name',
                  }"
                  :default-expanded-keys="[]"
                  :checked-keys="[]"
                  :get-group-sequence="column.getTree"
                  @change="changeTreeItem($event,column)"
                />
                <el-select
                  v-if="getFormItemType(column) === 'select'"
                  v-model="formData[column.name]"
                  filterable
                  remote
                  multiple
                  reserve-keyword
                  collapse-tags
                  :loading="loading[column.name]"
                  :remote-method="remoteMethod(column)"
                  @focus="initOpts($event,column)"
                  @change="onChange($event, column.onChange)"
                >
                  <el-option
                    v-for="(label, value) in opts[column.name]"
                    :key="value"
                    :label="label"
                    :value="value"
                  />
                </el-select>
                <el-select
                  v-if="getFormItemType(column) === 'selectNotRemote'"
                  v-model="formData[column.name]"
                  filterable
                  remote
                  multiple
                  reserve-keyword
                  collapse-tags
                  :loading="loading[column.name]"
                  @focus="initOpts($event,column)"
                  @change="onChange($event, column.onChange)"
                >
                  <el-option
                    v-for="(label, value) in opts[column.name]"
                    :key="value"
                    :label="label"
                    :value="value"
                  />
                </el-select>
                <el-select
                  v-if="getFormItemType(column) === 'singleSelectRemote'"
                  v-model="formData[column.name]"
                  filterable
                  remote
                  reserve-keyword
                  collapse-tags
                  :loading="loading[column.name]"
                  :remote-method="remoteMethod(column)"
                  @focus="initOpts($event,column)"
                  @change="onChange($event, column.onChange)"
                >
                  <el-option
                    v-for="(label, value) in opts[column.name]"
                    :key="value"
                    :label="label"
                    :value="value"
                  />
                </el-select>
                <el-select
                  v-if="getFormItemType(column) === 'singleSelect'"
                  v-model="formData[column.name]"
                  filterable
                  clearable
                  :loading="loading[column.name]"
                  @focus="initOpts($event,column)"
                  @change="onChange($event, column.onChange)"
                >
                  <el-option
                    v-for="(label, value) in opts[column.name]"
                    :key="value"
                    :label="label"
                    :value="value"
                  />
                </el-select>
                <template v-if="getFormItemType(column) === 'compareSearch'">
                  <div style="display:flex;">
                    <el-select
                      v-model="opType"
                      style="width:150px"
                    >
                      <el-option
                        label="大于"
                        value=">"
                      >
                        大于
                      </el-option>
                      <el-option
                        label="小于"
                        value="<"
                      >
                        小于
                      </el-option>
                    </el-select>
                    <el-input
                      v-model="formData[column.name]"
                      :placeholder="column.placeholder || placeholder"
                      style="padding-top:12.5px;margin: 0px 10px"
                    />
                    <el-select
                      v-model="timeType"
                      style="width:150px"
                    >
                      <el-option value="秒">
                        秒
                      </el-option>
                      <el-option value="分钟">
                        分钟
                      </el-option>
                      <el-option value="小时">
                        小时
                      </el-option>
                      <el-option value="天">
                        天
                      </el-option>
                    </el-select>
                  </div>
                </template>
                <el-checkbox-group
                  v-if="getFormItemType(column) === 'bool'"
                  v-model="formData[column.name]"
                >
                  <el-checkbox
                    value="是"
                    label="是"
                  />
                  <el-checkbox
                    value="否"
                    label="否"
                  />
                </el-checkbox-group>
                <date-picker
                  v-if="getFormItemType(column) === 'date'"
                  v-model="formData[column.name]"
                  type="daterange"
                />
                <el-select
                  v-if="getFormItemType(column) === 'enum'"
                  v-model="formData[column.name]"
                  multiple
                  collapse-tags
                >
                  <el-option
                    v-for="item in opts[column.name]"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
                <user-selector
                  v-if="getFormItemType(column) === 'user'"
                  v-model="formData[column.name]"
                />
                <user-selector
                  v-if="getFormItemType(column) === 'mutiUser'"
                  v-model="formData[column.name]"
                  multiple
                />
                <date-picker
                  v-if="getFormItemType(column) === 'date'"
                  v-model="formData[column.name]"
                  type="daterange"
                />
                <el-date-picker
                  v-if="getFormItemType(column) === 'datetime'"
                  v-model="formData[column.name]"
                  value-format="yyyy-MM-dd HH:mm:ss"
                  type="datetimerange"
                  :clearable="datePickerClearable"
                />
                <numberRange
                  v-if="getFormItemType(column) === 'numberRange'"
                  v-model="formData[column.name]"
                />
              </el-form-item>
            </el-col>
          </template>
        </el-form>
      </el-row>
      <div
        class="advanced-search-footer"
      >
        <!-- <el-button
          v-if="showMore"
          type="text"
          @click="toggle"
        >
          {{ expand ? '收起' : '更多' }}
          <i :class="expand ? 'tn-icon-arrow-up' : 'tn-icon-arrow-down'" />
        </el-button> -->
        <el-button
          type="text"
          style="color:#333333"
          @click="reset"
        >
          重置
        </el-button>
        <el-button
          type="text"
          @click="search"
        >
          查询
        </el-button>
      </div>
    </div>
  </div>
</template>
<script>
import { each, isEmpty, find, forEach, isNil, isFunction } from 'lodash';
import DatePicker from '../DateTimePicker';
import userSelector from '../user/user-selector';
import numberRange from '../numberRange';
import mixin from '../script/mixin';
import $ from 'jquery';
import getEdgeRequest from '../../../../utils/request';
import { eventBus } from '../script/eventBus';
import selectTree from './selectTree';
import moment from 'moment';

const timeBeforeOneMonth = moment().add('year', 0)
  .month(moment().month() - 1)
  .format('YYYY-MM-DD HH:mm:ss');
const currentTime = moment(Date.now()).format('YYYY-MM-DD HH:mm:ss');

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
    numberRange,
    selectTree,
  },
  mixins: [mixin],
  inject: ['tableConfig', 'commonCgi'],
  props: {
    advancedColumns: {
      type: Array,
      required: true,
      default: () => ([]),
    },
    play: {
      type: Boolean,
      default: false,
    },
    showCustomizeSearch: {
      type: Boolean,
      required: false,
      default: true,
    },
  },
  data() {
    return {
      treeValue: null,
      formData: {

      },
      opts: {},
      loading: {},
      pureStringNotArray: {},
      readonly: [],
      expand: true,
      clientWidth: 0,
      placeholder: '支持多个查询，用分号分隔',
      // showMore: true,
      opType: '>',
      timeType: '秒',
      options: [{
        id: 'a',
        label: 'a',
        children: [{
          id: 'aa',
          label: 'aa',
        }, {
          id: 'ab',
          label: 'ab',
        }],
      }, {
        id: 'b',
        label: 'b',
      }, {
        id: 'c',
        label: 'c',
      }],
    };
  },
  computed: {
    columns() {
      return this.advancedColumns;
    },
    datePickerClearable() {
      return this.tableConfig?.datePickerClearable;
    },
  },
  watch: {
    'tableConfig.searchParams.mozuId': {
      handler(val) {
        if (val && this.$refs.form) {
          this.$refs.form.resetFields();
          this.search(1);
        }
      },
      deep: true,
      immediate: true,
    },
    'tableConfig.searchParams.fired': {
      handler() {
        if (this.$refs.form) {
          this.$refs.form.resetFields();
          eventBus.$emit('clearKeyword');
          this.search(1);
        }
      },
      deep: true,
      immediate: true,
    },
  },
  mounted() {
    this.init();
    eventBus.$on('toggleSearch', () => {
      this.search();
      this.$emit('doelse', { type: 'analysis' });
    });
    this.$nextTick(() => {
      this.clientWidth = $('.app-container').length && $('.app-container')[0].clientWidth;
      // this.showMore = $('#advanceRow')[0].clientHeight > 240;
    });
  },
  beforeDestroy() {
    eventBus.$off('toggleSearch');
  },
  methods: {
    changeTreeItem(data, column) {
      this.formData[column.name] = data;
    },
    toggle() {
      this.expand = !this.expand;
    },
    init() {
      const formData = {};
      this.columns.forEach((field) => {
        if (field.fieldEnum && field.fieldEnum.length) {
          this.$set(this.opts, field.name, field.fieldEnum);
        }
        if (this.getFormItemType(field) === 'pureStringNotArray') {
          this.pureStringNotArray[field.name] = '';
        }
        if (this.getFormItemType(field) === 'singleSelect') {
          formData[field.name] = '';
        }

        if (this.isSearch(field)) {
          const searchParams = this.tableConfig?.searchParams || {};
          const types = ['bool', 'enum', 'select', 'text', 'cascader'];
          if (types.includes(this.getFormItemType(field))) {
            // eslint-disable-next-line no-param-reassign
            if (this.getFormItemType(field) !== 'bool' && searchParams[field.name]) {
              const tempdata = searchParams[field.name][0].split(',');
              formData[field.name] = searchParams[field.name][0].split(',');
              this.tableConfig.searchParams[field.name] = tempdata;
            } else {
              formData[field.name] = [];
            }
          }
          if (this.getFormItemType(field) === 'datetime') {
            if (searchParams[field.name]) {
              formData[field.name] = searchParams[field.name][0].split(',');
              this.tableConfig.searchParams[`${field.name}Start`] = formData[field.name][0];
              this.tableConfig.searchParams[`${field.name}End`] = formData[field.name][1];
            } else {
              formData[field.name] = [];
            }
          }
        }
      });
      if (Object.hasOwnProperty.call(this.tableConfig, 'occurTime')) {
        formData.occurTime = [timeBeforeOneMonth, currentTime];
        this.tableConfig.searchParams.occur_begin = timeBeforeOneMonth;
        this.tableConfig.searchParams.occur_end = currentTime;
      }
      this.formData = { ...formData };
      this.refreshState();
    },
    refreshState() {
      const fields = this.columns;
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
          const match = find(this.columns, {
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
        const match = find(this.columns, {
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

            if (this.isSearch(field)) {
              const tag = this.getFormItemType(field) === 'bool' || this.getFormItemType(field) === 'enum' || this.getFormItemType(field) === 'select';
              if (tag) {
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
          isEqual = v === '0';
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
      // this.opts = {};
      // this.$refs.form.resetFields();
      this.$nextTick(() => {
        this.$refs.form.resetFields();
        if (this.$refs.treeSelect) this.$refs.treeSelect[0].clearSelectedNodes();
        eventBus.$emit('clearKeyword');
        this.search();
      });
    },
    search(status) {
      const conditions = {};
      each(this.formData, (value, key) => {
        if (Array.isArray(value)) {
          // eslint-disable-next-line no-param-reassign
          value = value.filter(Boolean);
        }
        if (isEmpty(value)) return;
        const field = find(this.columns, {
          name: key,
        });
        conditions[key] = value;
        const type = this.getFormItemType(field);
        if (type === 'date' || type === 'datetime' || type === 'numberRange') {
          // eslint-disable-next-line prefer-destructuring
          conditions[`${key}Start`] = value[0];
          // eslint-disable-next-line prefer-destructuring
          conditions[`${key}End`] = value[1];
          delete conditions[key];
        } else if (type === 'text' || type === 'num') {
          conditions[key] = value.replace('；', ';').split(';');
        } else if (type === 'cascader') {
          conditions[key] = this.$refs[key][0].checkedNodes.map(v => v.label);
        } else if (type === 'mutiUser' || type === 'user') {
          conditions[key] = value.map(v => v.userUid);
        } else if (type === 'pureStringNotArray') {
          conditions[key] = value;
        } else if (type === 'compareSearch') {
          const timeMap = {
            秒: 1,
            分钟: 60,
            小时: 3600,
            天: 86400,
          };
          conditions[key] = { duration: parseInt(value) * timeMap[this.timeType], sign: this.opType };
        } else if (type === 'singleSelect') {
          conditions[key] = value;
        }
      });
      const searchNameMap = this.tableConfig?.searchNameMap || {};
      if (Object.keys(searchNameMap).length !== 0) {
        Object.keys(searchNameMap).forEach((item) => {
          if (Object.hasOwnProperty.call(conditions, item)) {
            conditions[searchNameMap[item]] = conditions[item];
            delete conditions[item];
          }
        });
      }
      if (status === 1) {
        this.$emit('search');
      } else {
        this.$emit('search', { ...conditions });
      }
    },
    isSearch(column) {
      return column.isFilter;
    },
    showItem(field) {
      return !this.readonly.includes(field.name) && field.showInSearch;
    },
    getFormItemType(column) {
      if (!column) return '';
      if (column.fieldEnum && column.fieldEnum.length) {
        return 'enum';
      } if (column.type === 'string') {
        return 'text';
      } if (column.type === 'mutiSelect') {
        return 'select';
      } if (column.fieldSingleEnum && column.fieldSingleEnum.length) {
        return 'singleSelect';
      }
      return column.type;
    },
    remoteMethod(field) {
      return (keyword) => {
        if (!this.play) {
          if (field.fieldMultiEnum) {
            let newItem = {};
            if (field.fieldMultiEnum instanceof Array) {
              field.fieldMultiEnum = field.fieldMultiEnum.forEach((item) => {
                newItem[item] = item;
              });
            } else {
              newItem = field.fieldMultiEnum;
            }
            this.$set(this.opts, field.name, newItem);
          } else {
            const method = field.dropdownMethod || 'get';
            const path = field.dropdownPath;
            const query = (field.dropdownQuery && field.dropdownQuery(this.formData)) || {};
            if (field.needMozu) {
              query.mozuId = this.tableConfig.searchParams.mozuId;
            }
            this.$set(this.loading, field.name, true);
            // this.$axios[method](path, {
            getEdgeRequest(this.$axios, this.tableConfig.searchParams.mozuId)[method](path, {
              ...query,
              keyword,
            }, false)
              .then((dataResp) => {
                let data = field.dropdownFormatter ? field.dropdownFormatter(dataResp) : dataResp;
                if (Array.isArray(data)) { // 数组
                  this.$set(this.opts, field.name, data);
                } else { // 对象
                  if (field.needReverse) {
                    data = this.reverseData(data);
                  }
                  this.$set(this.opts, field.name, data.list || data);
                }
              })
              .finally(() => {
                this.loading[field.name] = false;
              });
          }
        } else {
          setTimeout(() => {
            this.$set(this.opts, field.name, {
              1: '选项1',
              2: '黄金糕',
            });
          });
        }
      };
    },
    initOpts(visible, field) {
      if (field.type === 'treeSelect' && visible === false) {
        this.$refs.selectTree.filterText = '';
        return;
      }
      if (!field.localData) {
        this.remoteMethod(field)();
      } else {
        if (!this.opts[field.name]) {
          let newItem = {};
          if (field.fieldSingleEnum) {
            if (field.fieldSingleEnum instanceof Array) {
              field.fieldSingleEnum = field.fieldSingleEnum.forEach((item) => {
                newItem[item] = item;
              });
            } else {
              newItem = field.fieldSingleEnum;
            }
          }

          if (field.fieldMultiEnum) {
            if (field.fieldMultiEnum instanceof Array) {
              field.fieldMultiEnum = field.fieldMultiEnum.forEach((item) => {
                newItem[item] = item;
              });
            } else {
              newItem = field.fieldMultiEnum;
            }
          }

          this.$set(this.opts, field.name, newItem);
        }
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
@import "../style/common";

.advanced-search-wrap {
  // border-top: 1px solid #f0f0f0;
  .is-rotated {
    -webkit-transform:rotate(180deg);
    -ms-transform:rotate(180deg);
    transform:rotate(180deg)
  }
  background-color: white;

  .advanced-search-title {
    line-height: 64px;
    padding: 0 24px;
    font-size: 16px;
    color: #333333;
    font-weight: 700;

    .el-button {
      margin-top: 20px;
    }
  }

  .advanced-search-body {
    // border-bottom: 1px solid #f0f0f0;
    padding: 0 $space-s;
    max-height: 170px;
    overflow: hidden;
  }
  .expand {
    overflow: auto;
    max-height: unset;
  }

  .advanced-search-footer {
    padding: 15px 24px;
    text-align: right;
    border-bottom: 1px solid #f0f0f0;

    button + button {
      margin-left: 32px;
    }
  }

  .el-form {
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
.el-select__tags-text {
  max-width: 160px;
}
</style>
