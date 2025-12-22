<template>
  <div>
    <el-page-holder
      type="not-found"
      :button-list="[
        { text: '返回首页', type: 'primary', onclick: back },
      ]"
    />
  </div>
</template>

<script>
export default {
  methods: {
    // 返回首页
    back() {
      location.href = '/';
    },
  },
};
</script>
