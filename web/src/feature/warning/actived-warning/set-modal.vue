<template>
  <el-dialog
    :visible.sync="logVisible"
    width="500px"
    @close="close"
  >
    <template
      slot="title"
    >
      消息提醒设置
    </template>
    <el-form
      ref="form"
      label-width="100px"
      :model="form"
    >
      <el-form-item
        label="等级过滤"
        prop="serverAssetId"
      >
        <!-- <el-checkbox
          v-model="checkAll"
          :indeterminate="isIndeterminate"
          @change="handleCheckAllChange"
        >
          全选
        </el-checkbox> -->
        <div style="margin: 15px 0;" />
        <el-checkbox-group
          v-model="checkedRooms"
          @change="handlecheckedRoomsChange"
        >
          <el-checkbox
            v-for="room in levels"
            :key="room.value"
            :label="room.value"
          >
            {{ room.label }}
          </el-checkbox>
        </el-checkbox-group>
      </el-form-item>
    </el-form>
    <template slot="footer">
      <el-button
        type="primary"
        @click="confirm()"
      >
        确定
      </el-button>
    </template>
  </el-dialog>
</template>

<script>

export default {
  components: {
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    checkedLevels: {
      type: String,
      default: '',
    },
    levels: {
      type: Array,
      default: () => ([]),
    },

  },
  data() {
    return {
      form: {
        serverAssetId: '',
      },
      checkAll: true,
      checkedRooms: [],
      isIndeterminate: false,
    };
  },
  computed: {
    logVisible: {
      set(v) {
        this.$emit('update:visible', v);
      },
      get() {
        return this.visible;
      },
    },
    seletedAll() {
      if (this.isIndeterminate === false && this.checkAll === true) {
        return true;
      }
      return false;
    },
  },

  mounted() {
    this.checkedRooms = this.checkedLevels.split(';');
    this.checkAll = this.levels.length === this.checkedRooms.length;
    this.isIndeterminate = this.levels.length !== this.checkedRooms.length;
  },
  methods: {
    handleCheckAllChange(val) {
      this.checkedRooms = val ? this.levels.map(i => i.value) : [];
      this.isIndeterminate = false;
    },
    handlecheckedRoomsChange(value) {
      console.log(value);
      const checkedCount = value.length;
      this.checkAll = checkedCount === this.levels.length;
      this.isIndeterminate = checkedCount > 0 && checkedCount < this.levels.length;
    },
    close() {
      this.logVisible = false;
      this.form = {};
      this.$emit('close');
    },
    logout() {
      console.log(this.tag);
    },
    confirm() {
      this.$refs.form.validate((valid) => {
        if (valid) {
          // this.treeData[0].children = this.treeData[0].children.filter(i => this.checkedRooms.includes(i.name));
          this.$emit('confirm', { notifyLevels: this.checkedRooms.join(';') });
          this.$message.success('设置成功');
          this.close();
        }
      });
    },
  },
};
</script>
<style>

</style>
