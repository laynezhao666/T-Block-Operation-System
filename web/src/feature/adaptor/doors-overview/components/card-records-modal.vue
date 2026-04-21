<template>
  <el-modal
    :visible.sync="visible"
    :width="1200"
  >
    <template slot="title">
      查看刷卡记录
    </template>

    <in-out-records
      v-if="doors && doors.length"
      :doors="doors"
    />
  </el-modal>
</template>

<script>
import InOutRecords from '../../security-records/in-out/index.vue';

export default {
  components: {
    InOutRecords,
  },
  data() {
    return {
      doors: null,
    };
  },
  computed: {
    visible: {
      get() {
        return Boolean(this.doors);
      },
      set(v) {
        if (!v) {
          this.doors = null;
        }
      },
    },
  },
  methods: {
    show(doors) {
      this.doors = doors;
    },
  },
};
</script>
