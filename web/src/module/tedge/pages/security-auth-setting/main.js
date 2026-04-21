/**
 * 安防系统 - 授权发卡页面入口
 * 加载 security-auth-setting 页面组件并初始化单页应用
 */
import simpleEntry from '@@/script/spa';
import page from 'feature/adaptor/security-auth-setting/index.vue';
import '../../../../utils/vue-timer.ts';

export default simpleEntry(page);

export * from '@@/script/spa';
