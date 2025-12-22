<template>
  <div class="table-toolbar">
    <slot name="buttons" />

    <slot name="statistics" />

    <div class="table-toolbar__right">
      <div
        v-if="showSearch"
        class="table-toolbar__search"
      >
        <el-input
          ref="input"
          v-model="keyword"
          align="right"
          :placeholder="placeholder ? placeholder : '请输入您需要搜索的内容'"
          @input.native="search"
        />
        <i
          class="tn-icon tn-icon-search"
        />
      </div>

      <slot name="handlers" />

      <div
        v-if="hasRights('canImport')||hasRights('canExport')"
        v-clickoutside="hideMoreDropdown"
        class="zero-height"
        :auth-right-code="authRightCode"
      >
        <el-button
          type="icon"
          icon="tn-icon tn-icon-more"
          @click="toggleMoreDropdown"
        />
        <config-popover
          :table="table"
          :text="text"
          :codes="codes"
          :roles="roles"
          :more-dropdown-visible="moreDropdownVisible"
          @uploadSuccess="refresh"
          @export="download"
        />
      </div>
    </div>
  </div>
</template>

<script>
import configMixin from 'component/ConfigPanel/mixin.js';
import ConfigPopover from './ConfigPopover.vue';
import Clickoutside from 'element-ui/src/utils/clickoutside';
import { debounce } from 'lodash';

export default {
  components: { ConfigPopover },
  directives: { Clickoutside },
  mixins: [configMixin],
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
      default: () => ({}),
    },
    roles: {
      type: Object,
      default: () => ({}),
    },
    table: {
      type: String,
      default: '',
    },
    text: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      keyword: '',
      showPopover: false,
      showMore: false,
      moreDropdownVisible: false,
      authRightCode: (this.codes.dr && this.codes.dc) ? `${this.codes.dr},${this.codes.dc}` : '',
      authRoles: (this.roles.write) ? `${this.roles.write}` : '',
    };
  },
  mounted() {
    this.search = debounce(this.search, 500, {
      leading: false,
      trailing: true,
    });
  },
  methods: {
    hideMoreDropdown() {
      this.moreDropdownVisible = false;
    },
    toggleMoreDropdown() {
      this.moreDropdownVisible = !this.moreDropdownVisible;
    },
    search() {
      this.$emit('search', this.keyword);
    },
    refresh() {
      this.$emit('refresh');
    },
    download() {
      this.$emit('export');
    },
  },
};
</script>

<style lang="scss" scoped>
@import './toolbar'
</style>
