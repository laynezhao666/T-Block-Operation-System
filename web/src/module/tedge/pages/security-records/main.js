/**
 * 安防系统 - 门禁记录页面入口
 * 加载 security-records 页面组件并初始化单页应用
 */
import simpleEntry from '@@/script/spa';
import page from 'feature/adaptor/security-records/index.vue';
import '../../../../utils/vue-timer.ts';

export default simpleEntry(page);

export * from '@@/script/spa';
