<template>
  <div>
    <span :style="{color: filter[keyName] ? '#1470cc' : ''}">
      <slot />
    </span>
    <el-popover
      :value="show"
      placement="bottom"
      @show="toggleVisible(true)"
      @hide="toggleVisible(false)"
    >
      <el-select
        ref="select"
        v-model="filter[keyName]"
        clearable
        filterable
        reserve-keyword
        placeholder="请输入关键词"
        :remote-method="v => getList(v)"
        :loading="loading"
        @change="v => select(v)"
      >
        <el-option
          v-for="(item, index) in list"
          :key="index"
          :label="item"
          :value="item"
        />
      </el-select>
      <span
        v-if="['occurExpression', 'restoreExpression'].indexOf(type) > -1"
        slot="reference"
        style="color:#1470CC;cursor:pointer"
      >{{ labelName }}<i
        style="color:#1470CC"
        class="el-table__column-filter-trigger"
        :class="
          !show ? 'el-icon-caret-bottom' : 'el-icon-caret-top'
        "
      /></span>
      <span
        v-else
        slot="reference"
        :style="filter[keyName] ? 'color:#1470CC;cursor:pointer' : ''"
      >{{ labelName }}<i
        :style="filter[keyName] ? 'color:#1470CC' : ''"
        class="el-table__column-filter-trigger"
        :class="
          !show ? 'el-icon-caret-bottom' : 'el-icon-caret-top'
        "
      /></span>
    </el-popover>
  </div>
</template>

<script>

/**
 * 表头的可关联下拉组件
 * 多个该组件可配合使用，传入各个model的合并对象，cgi每次一起发给后台，以计算当前可选项
 * TODO:cgi地址暂写到组件内，后续公用组件化可分离出来
 * 接口1. 区域-城市-园区-楼宇-房间-资产架-资产位
 * 字段名范围：{area, city, park, building, room, rack, position}
 * 接口2. 其他单列
 * 字段名范围：['company', 'assetStatus', 'logicName', 'operateType', 'instanceSystem', 'projectPm', 'project']
 * 接口3. 资产类型，资产型号
 * 字段名范围：{assetType, assetModel}
 * @props {string} type，当前所属字段key(以接口/库内为准)
 * @props {object} filter={}，原样展开发给cgi
 * @props {string} keyName=this.type， filter里对应key的名称，会同步更新
 * @props {boolean|string} useCache=undefined(根据type设置默认值), 所有取值：
 * true(一次拉取一直缓存,针对无关联且短时间内不会变化的数据)；
 * 'muti'(每次下拉拉取一次，仅用于每次拉取时，能返回全部可选项的情况)；
 * false(始终拉取,针对可进行关键字搜索的字段)
 * @props {string} columnNameKey='columnName'，cgi里“当前要拉取的列”的key名
 * @event change(value)
 */
import getEdgeRequest from '../../utils/request';

export default {
  props: {
    type: {
      type: String,
      required: true,
    },
    labelName: {
      type: String,
      default: '',
    },
    url: {
      type: String,
      default: '',
    },
    keyName: {
      type: String,
      default() {
        return this.type;
      },
    },
    filter: {
      type: Object,
      default() {
        return {};
      },
    },
    params: {
      type: Object,
      default() {
        return {};
      },
    },
    useCache: {
      type: [Boolean, String],
      default: undefined,
    },
    columnNameKey: {
      type: String,
      default: 'columnName',
    },
    ajaxMethod: {
      type: String,
      default: 'get',
    },
  },
  inject: ['tableConfig'],
  data() {
    return {
      LIMIT: 10,
      cache: [],
      list: [],
      show: false,
      loading: false,
      cgiUrl: '',
      cacheType: this.useCache,
    };
  },
  watch: {
    params: {
      handler(val) {
        this.deviceParam = val;
      },
      deep: true,
    },
  },
  mounted() {
    if (['area', 'city', 'park', 'building', 'room', 'rack', 'position'].indexOf(this.type) > -1) {
      this.cgiUrl = this.cgi.commonCgi.getPosOption;
      if (this.cacheType === undefined) {
        if (['rack', 'position'].indexOf(this.type) > -1) {
          this.cacheType = false;
        } else {
          this.cacheType = 'muti';
        }
      }
    } else if (['company', 'assetStatus', 'logicName', 'operateType',
      'instanceSystem', 'projectPm', 'project', 'assetOutType', 'assetInType', 'transferType',
      'department', 'orderStatus'].indexOf(this.type) > -1) {
      this.cgiUrl = this.cgi.commonCgi.getDictOption;
      if (this.cacheType === undefined) {
        this.cacheType = 'muti';
      }
    } else if (['type', 'model'].indexOf(this.type) > -1) {
      this.cgiUrl = this.cgi.commonCgi.getTypeOption;
      if (this.cacheType === undefined) {
        this.cacheType = false;
      }
    } else if (['operator'].indexOf(this.type) > -1) {
      this.cgiUrl = this.url;
      if (this.cacheType === undefined) {
        this.cacheType = false;
      }
    } else if (['occurExpression', 'restoreExpression'].indexOf(this.type) > -1) {
      this.cgiUrl = this.url;
      if (this.cacheType === undefined) {
        this.cacheType = false;
      }
    } else {
      // console.error('没有对应的url', this.type);
      return '';
    }
    this.coverData();
  },
  methods: {
    getList(value) {
      if (this.cacheType && this.cache.length) {
        this.list = this.cache.filter(item => item.toLowerCase().indexOf(value.toLowerCase()) > -1);
        return;
      }
      let params = {};
      if (['occurExpression', 'restoreExpression'].indexOf(this.type) > -1) {
        params = { mozuId: this.tableConfig.searchParams.mozuId };
      }
      this.loading = true;
      getEdgeRequest(
        this.$axios,
        this.tableConfig.searchParams.mozuId
      )[this.ajaxMethod](this.cgiUrl, {
        // 资产架和资产位使用模糊搜索，限制条件加在前端
        // 当前列关键字
        keyword: value,
        start: 0,
        limit: this.LIMIT,
        ...params,
        // 所有关联列已选项
        ...this.filter,
        ...this.deviceParam,
        // 当前列名
        [this.columnNameKey]: this.type,
      })
        .then((data) => {
          if (['occurExpression', 'restoreExpression'].indexOf(this.type) > -1) {
          // debugger;
            data = data.pointTypeList;
          }
          let items = [];
          if (data instanceof Array) {
            items = data;
          } else {
            for (const key in data) {
              if (data[key] instanceof Array) {
                items = data[key];
                break;
              }
            }
          }
          if (this.cacheType) {
            this.cache = items;
          }
          if (this.type === 'logOperation') {
            this.list = ['新增', '编辑', '删除', '批量删除'];
          } else {
            this.list = items;
          }
          this.loading = false;
        });
    },
    toggleVisible(show) {
      if (show) {
        this.getList('');
        this.show = true;
        this.$nextTick(() => {
          this.$refs.select.focus();
        });
      } else {
        if (this.cacheType !== true) {
          this.list = [];
          this.cache = [];
        }
        this.show = false;
      }
    },
    select(v) {
      this.toggleVisible(false);
      this.coverData();
      this.$emit('change', v);
    },
    coverData() {
      if (this.keyName !== this.type) {
        this.filter[this.type] = this.filter[this.keyName];
      }
    },
  },
};
</script>
