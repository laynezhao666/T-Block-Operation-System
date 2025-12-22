<template>
  <el-form
    ref="form"
    :model="edittingMenuItem"
    label-width="100px"
  >
    <el-form-item
      label="菜单标题"
      prop="title"
      required
    >
      <el-input
        v-model="edittingMenuItem.title"
        :disabled="disabledFieldsMap.title"
        placeholder="请输入菜单标题"
      />
    </el-form-item>

    <el-form-item
      label="菜单编号"
      prop="menuCode"
      required
    >
      <el-input
        v-model="edittingMenuItem.menuCode"
        :disabled="disabledFieldsMap.menuCode"
        placeholder="请输入菜单编号"
      />
    </el-form-item>

    <el-form-item
      label="菜单路径"
      prop="href"
      required
    >
      <el-input
        v-model="edittingMenuItem.href"
        :disabled="disabledFieldsMap.href"
        placeholder="请输入菜单路径"
      />
    </el-form-item>

    <el-form-item
      label="图标"
      prop="icon"
    >
      <el-select
        v-model="edittingMenuItem.icon"
        :disabled="disabledFieldsMap.icon"
        placeholder="请选择图标"
        filterable
        allow-create
      >
        <el-option-group
          v-for="group in iconsGroups"
          :key="group.group"
          :label="group.group"
        >
          <el-option
            v-for="icon in group.icons"
            :key="icon"
            :label="icon"
            :value="icon"
          >
            <i
              :class="icon"
              style="float: left"
            />
            <span style="float: right">{{ icon }}</span>
          </el-option>
        </el-option-group>

        <i
          slot="prefix"
          :class="edittingMenuItem.icon"
        />
      </el-select>
    </el-form-item>

    <el-form-item
      label="是否隐藏"
      prop="isHide"
    >
      <el-checkbox
        v-model="edittingMenuItem.isHide"
        :disabled="disabledFieldsMap.isHide"
      />
    </el-form-item>

    <el-form-item
      label="用户限制"
      prop="limitUserNames"
    >
      <el-input
        v-model="edittingMenuItem.limitUserNames"
        :disabled="disabledFieldsMap.limitUserNames"
        placeholder="如果需要显示访问用户，请填写改字段，多个用户请用英文分号（;）分割"
        clearable
      />
    </el-form-item>

    <el-form-item
      v-if="!noFooter"
      label=""
    >
      <el-button
        type="primary"
        size="small"
        @click="submit"
      >
        保存
      </el-button>

      <el-button
        size="small"
        @click="cancel"
      >
        取消
      </el-button>
    </el-form-item>
  </el-form>
</template>

<script>
import { iconsGroups } from '../icons';

export default {
  props: {
    menuItem: {
      type: Object,
      required: true,
    },
    noFooter: {
      type: Boolean,
      default() {
        return false;
      },
    },
    disabledFieldsMap: {
      type: Object,
      default() {
        return {};
      },
    },
    forceEdittingAsModel: {
      type: Boolean,
      default() {
        return false;
      },
    },
  },
  data() {
    return {
      edittingMenuItem: null,
      iconsGroups,
    };
  },
  watch: {
    menuItem: {
      immediate: true,
      deep: true,
      handler() {
        this.edittingMenuItem = this.forceEdittingAsModel ? this.menuItem : {
          ...this.menuItem,
        };
      },
    },
  },
  methods: {
    validate() {
      return this.$refs.form.validate();
    },
    async submit() {
      if (!(await this.validate())) {
        return;
      }
      this.$emit('submit', this.edittingMenuItem);
    },
    cancel() {
      this.$emit('cancel');
    },
  },
};
</script>

<style lang="scss" scoped>

</style>
