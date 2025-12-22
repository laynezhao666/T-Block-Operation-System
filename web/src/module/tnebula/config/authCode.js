
// 特殊权限码配置
// 配置中是否有特殊的编码：
// auth_code 权限系统中的配置操作码
// op_code 页面中硬编码的操作码
// 1. is_remove 设为true表示如果没权限会删除该节点,优化级最高
// 2. 其他set属性根据配置的值来设置
// 3.如果没有is_remove和set属性，默认是display:none
// {auth_code:{op_code:"",set_display:"none"//默认,
// set_visibility:"hidden",set_class:"",set_disabled:false,is_remove:true//优先级最高}}
// eslint-disable-next-line import/prefer-default-export
export const codeConf = {};
