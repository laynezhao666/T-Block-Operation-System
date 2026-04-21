<template>
  <el-table-column
    :width="width"
    prop="controller_name"
    label="门禁控制器"
  >
    <template
      #header
    >
      <span>门禁控制器</span>
      <!-- 接口该参数延期 -->
      <drop-down-checkboxes
        v-if="false"
        v-model="filters.controls"
        :options="controlsOptions"
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
        return 130;
      },
    },
    filters: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {
      controlsOptions: [],
    };
  },
  created() {
    this.loadControlsOptions();
  },
  methods: {
    async loadControlsOptions() {
      const {
        controls,
      } = await memoriedFetchControlsAndDoors();

      this.controlsOptions = _.map(controls, door => ({
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
