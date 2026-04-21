/**
 * 安防系统 - 门禁请求页面入口
 * 加载 security-requests 页面组件并初始化单页应用
 */
import simpleEntry from '@@/script/spa';
import page from 'feature/adaptor/security-requests/index.vue';
import '../../../../utils/vue-timer.ts';

export default simpleEntry(page);

export * from '@@/script/spa';
