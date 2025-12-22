<template>
    <div>
      <el-page-holder
        type="no-permission"
        title="无模组访问权限，请联系管理员添加"
        :button-list="[{ text: '返回首页', type: 'primary', onclick: back }]"
      />
    </div>
  </template>
  
  <script type="text/babel">
  export default {
    methods: {
      back() {
        location.href = '/';
      },
    },
  };
  </script>
  