<template>
  <div>
    <el-page-holder
      type="no-permission"
      title="无权访问，请联系管理员添加"
      :button-list="[{ text: '返回首页', type: 'primary', onclick: back }]"
    />
  </div>
</template>

<script >
export default {
  methods: {
    back() {
      location.href = '/';
    },
  },
};
</script>
