<template>
  <div class="common-user-selector">
    <el-select
      v-model="v"
      v-bind="$attrs"
      filterable
      remote
      collapse-tags
      placeholder="请输入名字进行搜索"
      value-key="userName"
      :remote-method="search"
      @focus="search()"
    >
      <el-option
        v-for="item in result"
        :key="item.userName"
        :label="getLabel(item)"
        :value="item"
      >
        <span
          v-if="isSlot"
          style="float: left"
        >{{ getSlotConfig(item) }}</span>
      </el-option>
    </el-select>
  </div>
</template>

<script>
/**
 * 用户选择器:select组件
 * @param {Array} value 数组
 * @param {Array} fields 要求返回的key列表，必须包含userName
 * @param {Array} userType 要求返回的用户类型列表，整数数组
 * 其他多选，数量限制等，和el-select相同
 * TODO: select的option可以为插槽模式
 * value: [{
 *  userName: <string>, // 必须
 * }]
 */
import userSelectorMixin from './user-selector.mixin';
export default {
  mixins: [userSelectorMixin],
  inject: ['commonCgi'],
  props: {
    value: {
      type: Array,
      required: true,
      default: () => ([]),
      validator(value) {
        return value.every(v => v.userName);
      },
    },
    // 待选项个数
    popperNum: {
      type: Number,
      default: 5,
    },
    // 获取的数据字段信息，也可获取用户的邮箱、电话等
    // userName始终为key
    fields: {
      type: Array,
      default() {
        return [
          'userUid',
          'userName',
          'userRealName',
        ];
      },
      validate(list) {
        return list.indexOf('userName') > -1;
      },
    },
    // 筛选用户类型
    // 必须为整数数字数组
    userType: {
      type: [Array, undefined],
      default: undefined,
      validate(list) {
        return list === undefined || list.every(item => (typeof item === 'number' && item % 1 === 0));
      },
    },
    isSlot: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      // 展示列表
      result: [],
    };
  },
  computed: {
    multiple() {
      // 直接只写属性名时，值的空的，认为是true
      return this.$attrs.multiple !== undefined && this.$attrs.multiple !== false;
    },
    // 原始select的value是单个为任意，多个为数组
    // 组件这里强制都为数组了，所以需要转换一次
    v: {
      set(v) {
        if (!(v instanceof Array)) {
          v = [v];
        }
        this.$emit('input', v);
      },
      get() {
        if (!this.multiple) {
          return this.value[0];
        }
        return this.value || [];
      },
    },
  },
  watch: {
    value() {
      // 保证外部切换值的时候也可以正确显示value tags
      this.initResult();
    },
  },
  mounted() {
    // 初始化的时候就要计算值先
    this.initResult();
  },
  methods: {
    getLabel(item) {
      if (item.userRealName) {
        return `${item.userName}(${item.userRealName})`;
      }
      console.warn('拉取数据缺少userRealName，请尽快修改');
      return item.userName;
    },
    getSlotConfig(item) {
      if (item.userName && item.userRealName && item.userEmailAddress) {
        return `${item.userName}(${item.userRealName})<${item.userEmailAddress}>`;
      }
      console.warn('拉取数据缺少userName或userRealName或userEmailAddress，请尽快修改');
      return item.userName;
    },
    search(key) {
      // 这个逻辑是表示不同的交互：
      // key为空的情况：初始焦点、确认一个下拉，删除已输入词语
      // - result为空表示无输入无下拉，这时空下拉dom位置可能会从左上角跳到下拉处（演练页面，样式问题，未定位）
      // - result为当前表示已搜索的下拉不因key为空而切换
      // 现在暂时改为：key为空表示全量搜索

      // if (!key) {
      //   // 把初始进result的都还原
      //   this.result = this.result || []
      //   this.result = []
      //   return
      // }
      this.getList({
        keywords: key || '',
        limit: this.popperNum,
        fields: this.fields,
        userType: this.userType,
      }).then((list) => {
        this.result = list;
      });
    },
    initResult() {
      // 远程的对象值
      // 必须出现在列表，才能正常被显示在框内tag上
      if (this.value && this.value.length) {
        this.result = this.value;
      }
      this.$nextTick(() => {
        // 清空initResult后等于value的result，不然下拉列表会闪一次value列表
        this.result = [];
      });
    },
  },
};
</script>
