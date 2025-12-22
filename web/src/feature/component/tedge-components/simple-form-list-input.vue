<template functional>
  <div class="simple-form-list-input">
    <div
      v-for="(item, i) in props.array"
      :key="i"
      class="item"
    >
      <slot
        :item="item"
        :setItem="$options.curryingSetItem(props.array, i)"
      />

      <el-button
        :disabled="props.min === props.array.length"
        type="text"
        icon="tn-icon-circle-remove"
        @click="$options.removeItem(props.array, i)"
      />
    </div>

    <el-button
      v-show="props.max > props.array.length"
      type="text"
      icon="tn-icon-add"
      class="add-btn"
      @click="$options.createNew(props)"
    >
      {{ props.addLabel }}
    </el-button>
  </div>
</template>

<script>
export default {
  props: {
    array: {
      type: Array,
      required: true,
    },
    newItem: {
      type: Function,
      required: true,
    },
    addLabel: {
      type: String,
      default() {
        return '新增';
      },
    },
    max: {
      type: Number,
      default() {
        return Number.MAX_SAFE_INTEGER;
      },
    },
    min: {
      type: Number,
      default() {
        return 0;
      },
    },
  },
  createNew(props) {
    props.array.push(props.newItem());
  },
  removeItem(arr, index) {
    arr.splice(index, 1);
  },
  curryingSetItem(arr, index) {
    return (newItem) => {
      arr.splice(index, 1, newItem);
    };
  },
};
</script>

<style lang="scss" scoped>
.simple-form-list-input {
  /deep/ {
    & + .el-form-item__error {
      position: static;
    }
  }
}

.item {
  display: flex;

  /deep/ {
    & > *:first-child {
      flex: 1;
    }
  }
}

.add-btn {
  margin-top: -16px;
}
</style>
