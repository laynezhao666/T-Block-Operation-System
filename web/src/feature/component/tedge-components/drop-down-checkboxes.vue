<template>
  <el-dropdown :hide-on-click="false">
    <span class="el-dropdown-link">
      <i
        class="tn-icon-filter"
        :class="{
          active: value && value.length
        }"
      />
    </span>
    <el-dropdown-menu
      slot="dropdown"
      width="160"
    >
      <el-dropdown-item
        v-if="filterable"
      >
        <el-input
          v-model="filterKeywords"
          border-type="bordered"
          size="small"
          placeholder="输入关键字搜索"
          class="search-input"
        />
      </el-dropdown-item>

      <el-dropdown-item
        v-for="(opt, i) in filteredOptions"
        :key="i"
      >
        <el-checkbox
          :value="testIsChecked(opt)"
          @change="handleCheckedChange(opt, $event)"
        />
        &nbsp;
        {{ opt.label }}
      </el-dropdown-item>
    </el-dropdown-menu>
  </el-dropdown>
</template>

<script>
export default {
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Array,
      default() {
        return null;
      },
    },
    options: {
      type: Array,
      required: true,
    },
    filterable: {
      type: Boolean,
      default() {
        return true;
      },
    },
  },
  data() {
    return {
      filterKeywords: '',
    };
  },
  computed: {
    filteredOptions() {
      const {
        filterKeywords,
        options,
      } = this;

      return _.filter(options, opt => opt.label?.includes?.(filterKeywords));
    },
  },
  methods: {
    testIsChecked(opt) {
      return _.some(this.value || [], item => item === opt.value);
    },
    handleCheckedChange(opt, checked) {
      const value = this.value || [];
      const newValue = [...value];
      const indexOfOpt = _.findIndex(value, item => item === opt.value);

      if (checked) {
        newValue.push(opt.value);
      } else {
        newValue.splice(indexOfOpt, 1);
      }
      console.log(value, newValue, indexOfOpt);
      this.$emit('change', newValue);
    },
  },
};
</script>

<style lang="scss" scoped>
.search-input {
  line-height: 24px;
  height: 24px;
  margin: 8px 0;

  /deep/ .el-input__inner {
    line-height: 22px;
    height: 24px;
  }
}

.tn-icon-filter {
  font-size: 16px;
  position: relative;
  top: 2px;
  color: #a0a0a0;
  transition: 0.3s color;
  cursor: pointer;

  &.active {
    color: var(--tn-color-primary);
  }
}
</style>
