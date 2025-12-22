<template>
  <el-dialog
    :visible.sync="logVisible"
    width="500px"
    @close="close"
  >
    <template
      slot="title"
    >
      位置筛选
    </template>
    <el-form
      ref="form"
      label-width="100px"
      :model="form"
    >
      <el-form-item
        label="显示房间"
        prop="serverAssetId"
      >
        <el-checkbox
          v-model="checkAll"
          :indeterminate="isIndeterminate"
          @change="handleCheckAllChange"
        >
          全选
        </el-checkbox>
        <div style="margin: 15px 0;" />
        <el-checkbox-group
          v-model="checkedRooms"
          @change="handlecheckedRoomsChange"
        >
          <el-checkbox
            v-for="room in rooms"
            :key="room.value"
            :label="room.label"
          >
            {{ room.value }}
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
import getEdgeRequest from '../../../utils/request';
import { dataQuery as cgi } from '@@/config/cgi';
import { cloneDeep } from 'lodash';

export default {
  components: {
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    treeData: {
      type: Array,
      default: () => ([]),
    },
    checkedRoomsProp: {
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
      checkedRooms: this.checkedRoomsProp,
      isIndeterminate: false,
      rooms: [],
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
    // this.getRoomList();

    this.rooms = cloneDeep(this.treeData[0].children.map(i => ({ label: i.name, value: i.name })));
    this.checkAll = this.rooms.length === this.checkedRooms.length;
    this.isIndeterminate = this.rooms.length !== this.checkedRooms.length;
  },
  methods: {
    handleCheckAllChange(val) {
      this.checkedRooms = val ? this.rooms.map(i => i.value) : [];
      this.isIndeterminate = false;
    },
    handlecheckedRoomsChange(value) {
      const checkedCount = value.length;
      this.checkAll = checkedCount === this.rooms.length;
      this.isIndeterminate = checkedCount > 0 && checkedCount < this.rooms.length;
    },
    getRoomList() {
      getEdgeRequest(this.$axios, this.mozuId)
        .post(cgi.getDistinctByFieldName, { fieldName: 'roomCode' }, true)
        .then((data) => {
          this.rooms = data.map(e => ({
            value: e,
            label: e,
          }));
        });
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
          this.$emit('confirm', { checkedRooms: this.checkedRooms, seletedAll: this.seletedAll });
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
