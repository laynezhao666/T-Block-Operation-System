<template>
  <div class="statistics">
    <div
      v-for="(item, key, index) in areaCount"
      :key="index"
    >
      <span class="key">{{ key }}</span>
      <span class="value">{{ item }}</span>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    table: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      areaCount: {},
    };
  },
  inject: ['configCgi', 'commonCgi'],
  created() {
    this.$axios.post(this.configCgi.getTotal, { table: this.table }).then((res) => {
      this.areaCount = res;
    });
  },
  methods: {},
};
</script>

<style lang="scss">
.statistics {
  display: flex;

  div {
    margin-right: 20px;

    span {
      font-size: 16px;
      font-family: 'Microsoft YaHei';

      &.key {
        color: #333;
        margin-right: 10px;
        font-weight: Bold;
      }

      &.value {
        color: #1470cc;
        font-weight: 500;
        font-family: 'DINMittelschriftStd';
      }
    }
  }
}
</style>
