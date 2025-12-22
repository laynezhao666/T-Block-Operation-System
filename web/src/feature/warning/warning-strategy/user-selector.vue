<template>
  <div>
    <span :style="{color: selectedName ? '#1470cc' : ''}">
      <slot />
    </span>
    <el-popover
      v-model="show"
      @show="toggleVisible(true)"
      @hide="toggleVisible(false)"
    >
      <el-select
        ref="select"
        v-model="selectedName"
        :value="value"
        clearable
        filterable
        remote
        reserve-keyword
        placeholder="请输入关键词"
        :remote-method="v => getList(v)"
        :loading="loading"
        @change="v => select(v)"
      >
        <el-option
          v-for="(item, index) in list"
          :key="index"
          :label="item.name"
          :value="item.uid"
        />
      </el-select>
      <i
        slot="reference"
        class="el-table__column-filter-trigger"
        :style="selectedName ? 'color:#1470CC' : ''"
        :class="!show ? 'el-icon-caret-bottom' : 'el-icon-caret-top'"
      />
    </el-popover>
  </div>
</template>

<script>

export default {
  props: {
    value: {
      type: String,
      default() {
        return '';
      },
    },
    limit: {
      type: Number,
      default: 10,
    },
  },
  // inject: ['cgi'],
  data() {
    return {
      selectedName: '',
      show: false,
      cache: [],
      list: [],
      loading: false,
      cgiUrl: '/cgi/dcom/common/account/getAllUser',
    };
  },
  mounted() {
  },
  methods: {
    getList(value) {
      if (this.cache.length) {
        if (value) {
          this.list = this.cache.filter(item => item.name.indexOf(value) > -1);
        } else {
          this.list = this.cache.slice(this.limit);
        }
        return;
      }
      this.loading = true;
      this.$axios.get(this.cgiUrl).then((data) => {
        this.cache = this.coverData(data);
        this.getList(value);
        this.loading = false;
      });
    },
    coverData(map) {
      const arr = [];
      for (const uid in map) {
        arr.push({
          uid,
          name: map[uid],
        });
      }
      return arr;
    },
    toggleVisible(visible) {
      this.show = visible;
      if (visible) {
        this.focus();
      }
    },
    focus() {
      this.$nextTick(() => {
        this.$refs.select.focus();
        this.getList('');
      });
    },
    select(v) {
      this.$emit('input', v);
      this.selectedName = v;
      this.toggleVisible(false);
    },
  },
};
</script>
