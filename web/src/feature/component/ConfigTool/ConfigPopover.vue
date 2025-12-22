<template>
  <div class="config-popover">
    <import-modal
      :visible.sync="showModal"
      :text="text"
      :table="table"
      @success="uploadSuccess"
    />
    <ul
      v-if="moreDropdownVisible"
      class="dropdown"
    >
      <li
        v-if="hasRights('canImport')"
        class="dropdown-item"
        :auth-roles="roles.write"
        :auth-right-code="codes.dr"
      >
        <el-button
          type="icon"
          icon="tn-icon tn-icon-export"
          @click.native="show"
        >
          导入
        </el-button>
      </li>
      <li
        v-if="hasRights('canExport')"
        class="dropdown-item"
        :auth-right-code="codes.dc"
      >
        <el-button
          type="icon"
          icon="tn-icon tn-icon-import"
          @click.native="download"
        >
          导出全部
        </el-button>
      </li>
    </ul>
  </div>
</template>
<script>
import ImportModal from './ImportModal';
import configMixin from '../ConfigPanel/mixin.js';

export default {
  components: { ImportModal },
  mixins: [configMixin],
  props: {
    table: {
      type: String,
      default: '',
    },
    text: {
      type: String,
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
    moreDropdownVisible: {
      type: Boolean,
      required: true,
    },
  },
  data() {
    return {
      showModal: false,
    };
  },
  methods: {
    uploadSuccess() {
      this.$emit('uploadSuccess');
    },
    show() {
      this.showModal = true;
    },
    download() {
      this.$emit('export');
    },
  },
};
</script>
<style lang="scss" scoped>
.dropdown {
  position: absolute;
  box-sizing: border-box;
  width: 120px;
  right: 0;
  z-index: 99;
  box-shadow: 0 15px 30px 3px rgba(62,62,62,0.20);
  background: #fff;

  li.dropdown-item {
    display: flex;
    align-items: center;
    height: 48px;
    color: rgb(153, 153, 153);
    padding: 0px 16px;
    font-size: 14px;
    cursor: pointer;
    .el-button--icon {
      color: #666;
    }
    &:hover {
      background-color: rgb(246, 246, 246);
      color: rgb(102, 102, 102);
    }
    span {
      color: #666;
    }
    /deep/i.tn-icon-export, /deep/i.tn-icon-import{
      margin-right: 0px;
    }
  }
}

</style>
