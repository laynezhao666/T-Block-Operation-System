<template>
  <!-- <div
    v-if="visible"
    style="text-align: center;width:100%"
  > -->
  <el-card
    v-if="selectVisible"
    style="position:fixed;z-index: 9999;width: 850px;"
  >
    模组设备数量总共<span style="color: #1470CC"> {{ deviceCount }}</span>
    <el-transfer
      v-model="transferData"
      class="strategy-transfer"
      :filter-method="filterMethod"
      filterable
      :render-content="renderFunc"
      filter-placeholder="多个设备编号模糊查询请用英文分号分隔"
      :titles="['待选择设备列表', '已选择设备列表']"
      :format="{
        noChecked: '${total}',
        hasChecked: '${checked}/${total}'
      }"
      :data="data"
      @change="handleChange"
    >
      <el-button
        slot="left-footer"
        type="text"
        class="transfer-footer"
        size="small"
      />
      <el-button
        slot="right-footer"
        type="plain"
        class="transfer-footer rightfooter"
        size="small"
        @click="cancel"
      >
        取消
      </el-button>
      <el-button
        slot="right-footer"
        type="primary"
        class="transfer-footer rightfooter"
        size="small"
        @click="confirm"
      >
        确认
      </el-button>
    </el-transfer>
  </el-card>
</template>

<style>
  .transfer-footer {
    margin-left: 16px;
  }
</style>

<script>
export default {
  props: {
    listdata: {
      type: Array,
      default() {
        return [];
      },
    },
    choosedDevice: {
      type: Array,
      default() {
        return [];
      },
    },

    visible: {
      type: Boolean,
      default() {
        return false;
      },
    },
  },
  data() {
    return {
      transferData: [],
      renderFunc(h, option) {
        return <span>{ option.label }</span>;
      },
    };
  },
  computed: {
    data() {
      this.listdata.forEach((item) => {
        item.key = item.value;
      });

      return this.listdata;
    },
    deviceCount() {
      return this.listdata.length;
    },
    selectVisible: {
      set(v) {
        this.$emit('update:visible', v);
      },
      get() {
        // eslint-disable-next-line vue/no-side-effects-in-computed-properties
        this.transferData = this.choosedDevice;
        return this.visible;
      },
    },
  },
  methods: {
    confirm() {
      this.$emit('confirm', this.transferData);
      console.log(this.choosedDevice);
      console.log(this.transferData);
      this.selectVisible = false;
    },
    cancel() {
      this.selectVisible = false;
    },
    handleChange() {
    },
    filterMethod(query, item) {
      const deviceList = query.split(';').filter(item => item);
      console.log(deviceList);
      if (!deviceList.length) return item;
      return deviceList.some(queryItem => item.label.indexOf(queryItem) > -1);
    },
  },
};
</script>
<style lang="scss" scoped>
.rightfooter {
   float: right;
   display:flex;
   justify-content:center;
   margin: 8px 10px 0 0
}
</style>
<style lang="scss">
.strategy-transfer {
    text-align: left;
    display: inline-block;
    margin-top: 15px;
    width:100%;
    .el-transfer-panel {
      width:350px;
    }
}

</style>
