<template>
  <div>
    <span :style="{color: (v && (v[0] || v[1])) ? '#1470cc' : ''}">
      <slot />
    </span>
    <el-popover
      v-model="show"
      placement="bottom"
      width="400"
      trigger="click"
      @show="toggleVisible(true)"
    >
      <el-date-picker
        ref="datetime"
        v-model="v"
        type="datetimerange"
        value-format="yyyy-MM-dd HH:mm:ss"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        align="right"
        :default-time="['00:00:00', '23:59:59']"
        @change="change"
      />
      <i
        slot="reference"
        :style="{color: v && v.length !== 0 ? '#1470cc' : ''}"
        class="el-table__column-filter-trigger"
        :class=" !show ? 'el-icon-caret-bottom' : 'el-icon-caret-top' "
      />
    </el-popover>
  </div>
</template>

<script>
export default {
  props: {
    value: {
      type: Array,
      default() {
        return [];
      },
    },
  },
  data() {
    return {
      show: false,
      v: this.value,
    };
  },
  watch: {
    value: {
      handler(val) {
        this.v = this.value;
        console.log(val, '???');
      },
      deep: true,
    },
  },
  methods: {
    change(v) {
      this.v = v;
      this.$emit('input', v);
      this.toggleVisible(false);
    },
    toggleVisible(visible) {
      if (visible) {
        this.$nextTick(() => {
          this.$refs.datetime.focus();
        });
      } else {
        this.show = false;
      }
    },
  },
};
</script>
