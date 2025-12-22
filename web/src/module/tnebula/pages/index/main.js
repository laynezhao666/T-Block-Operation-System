
import Vue from 'vue';
import init from 'common/script/bootstrap2';
import passportPlugin from 'common/script/passport_plugin';
import frameNew from '@@/pages/index/frameNew';
import { cgi, url } from '../../config/api';
import 'common/style/ui/_reset.scss';
import 'common/style/reset.css';
// import * as reqhel from '../common/reqhel.js';
import { getAllOpCode } from '../common/auth';
import { getAllRoleList } from '../common/newAuth.js';
import * as reqhel from '../common/reqhel.js';
import authDirectives from './auth-directives.js';
Vue.directive('appmatrixauth', authDirectives.appMatrixAuth);

const psPlugin = passportPlugin();

$(document).ajaxSuccess((event, request, settings, data) => {
  try {
    if (typeof data === 'string') {
      // eslint-disable-next-line no-param-reassign
      data = JSON.parse(data);
    }
    if (data && data.code) {
      psPlugin.args.codeHandler.call(vue, data.code);
    }
  } catch (e) { }
});

Vue.config.errorHandler = function (err /* , vm, info */) {
  if (err.code) {
    vue.$message.error(err.message || '发生错误了');
  }
  throw err;
};

// 全局捕获promise的reject
window.addEventListener('unhandledrejection', (event) => {
  if (event.reason?.code) {
    vue.$message.error(event.reason.message || '发生错误了');

    event.preventDefault();
  }
}, false);

// 全局获取opcodeList
if (!window.TNBL) {
  window.TNBL = {};
}
window.TNBL.getAllOpCode = getAllOpCode;
window.TNBL.getScopeModules = reqhel.getScopeModules;
window.TNBL.getAllRoleList = getAllRoleList;

init({
  plugins: [psPlugin],
  customHttp: true,
});

const vue = new Vue({
  el: '#__TNBL__',
  render: h => h(frameNew, { props: { cgi: cgi.Index, url } }),
});

/*
按钮级权限开始
*/
// if (window.hasAppAuth) {
//   console.log('********开始新auth check');
//   observeNew();
//   checkAllNew();
// } else {
//   console.log('********开始老auth check');
//   observe();
//   checkAll();
// }

/*
按钮级权限结束
*/

export default vue;
