<template>
  <div class="table-toolbar">
    <slot name="buttons" />
    <div
      class="table-toolbar__right"
    >
      <div
        v-if="showSearch"
        class="table-toolbar__search"
      >
        <i
          class="tn-icon tn-icon-search"
        />
        <el-input
          ref="input"
          v-model="keyword"
          align="right"
          :placeholder="placeholder ? placeholder : '请输入您需要搜索的内容'"
          @input.native="search"
        />

        <!-- <el-table-toolbar
          id="common-toolbar"
          v-model="keyword"
          style="text-align:right;"
          placeholder="输入关键字"
          @search="search"
        />
      </div> -->

        <slot name="handlers" />

        <!-- 更多 -->
        <div
          v-if="$slots.more"
          class="zero-height"
        >
          <el-dropdown
            trigger="hover"
          >
            <span class="el-dropdown-link">
              <i
                slot="reference"
                class="tn-icon tn-icon-more"
              />
            </span>
            <el-dropdown-menu
              slot="dropdown"
              class="more-dropdown"
              width="160"
            >
              <slot name="more" />
            </el-dropdown-menu>
          </el-dropdown>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { debounce } from 'lodash';

export default {
  model: {
    prop: 'value',
    event: 'input',
  },
  props: {
    showSearch: Boolean,
    value: {
      type: String,
      default: '',
    },
    placeholder: {
      type: String,
      default: '',
    },
    codes: {
      type: Object,
      required: true,
    },
    roles: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {
      keyword: '',
      showPopover: false,
    };
  },
  mounted() {
    this.search = debounce(this.search, 500, {
      leading: false,
      trailing: true,
    });
  },
  methods: {
    togglePopover() {
      this.showPopover = !this.showPopover;
    },
    search() {
      this.$emit('search', this.keyword);
    },
  },
};
</script>
<style lang="scss" >
 #common-toolbar{
   .el-input {
    //  position: static;
    padding-top: 13px;
    // padding-right: 10px;
   }
    .el-input__inner::placeholder {
      font-size: 16px;
    }
 }
</style>
<style lang="scss" scoped>
@import './toolbar'
</style>
