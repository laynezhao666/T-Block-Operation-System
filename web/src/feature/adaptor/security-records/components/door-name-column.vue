<template>
  <el-table-column
    :width="width"
    prop="door_name"
    label="门名称"
  >
    <template
      #header
    >
      <span>门名称</span>

      <!-- 接口该参数延期 -->
      <drop-down-checkboxes
        v-if="false"
        v-model="filters.door_name"
        :options="doorNameOptions"
      />
    </template>
  </el-table-column>
</template>

<script>
import DropDownCheckboxes from '../../../component/tedge-components/drop-down-checkboxes.vue';
import { memoriedFetchControlsAndDoors } from '../utils/fetch-data';

export default {
  components: {
    DropDownCheckboxes,
  },
  props: {
    width: {
      type: Number,
      default() {
        return 100;
      },
    },
    filters: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {
      doorNameOptions: [],
    };
  },
  created() {
    this.loadDoorNameOptions();
  },
  methods: {
    async loadDoorNameOptions() {
      const {
        doors,
      } = await memoriedFetchControlsAndDoors();

      this.doorNameOptions = _.map(doors, door => ({
        label: door.name,
        // TODO: 需要等后端接口看按什么查询
        value: door.id,
      }));
    },
  },
};
</script>

<style lang="scss" scoped>

</style>
