import Cookies from 'js-cookie';
import http from 'common/script/http2';
// import { create } from 'common/script/http2';
// import { unloggedTip } from 'common/script/tips';
// import {
//   redirectToLogin,
// } from 'common/script/passport_login';
import { cgi } from '../../config/api';

// const http = create({
//   codeHandler: (code) => {
//     if (code === -99) {
//       unloggedTip().then(() => {
//         redirectToLogin({ newTarget: true });
//       });
//     }
//   },
// });

/**
 * 查询&存储 cookie 值
 * expires 天数
 */
export function cs(key, val, expires = 1) {
  console.log('init:cookie', key, val);

  if (val === undefined) {
    return Cookies.get(key);
  }

  const domainArr = location.hostname.split('.');
  const domain = domainArr.filter((item, i) => i > (domainArr.length - 4)).join('.');

  return Cookies.set(key, val, { expires, domain, path: '/' });
}

//  删除cookie
export function deleteCookie(key) {
  if (Cookies.get(key)) {
    const domainArr = location.hostname.split('.');
    const domain = domainArr.filter((item, i) => i > (domainArr.length - 4)).join('.');
    Cookies.remove(key, { domain, path: '/' });
  }
}

/**
 * 查询&存储 localstorage 值
 */
export function ls(key, val, delOper = false) {
  if (delOper) {
    return localStorage.removeItem(key);
  }
  if (val === undefined) {
    return localStorage.getItem(key);
  }
  return localStorage.setItem(key, val);
}
/**
 * 查询&存储 sessionstorage 值
 */
export function ss(key, val, delOper = false) {
  if (delOper) {
    return sessionStorage.removeItem(key);
  }

  if (val === undefined) {
    return sessionStorage.getItem(key);
  }
  return sessionStorage.setItem(key, val);
}

/**
 * @param params {scope_type:"module_group",scope_value:"1"}
 */
export async function getMenu(params = {}, username,) {
  const lsPrefixData = 'tnebula_ls_menu_data_';
  const lsPrefixTime = 'tnebula_ls_menu_time_';
  const lsDays = 3;
  try {
    const url = params.hasAppAuth ? cgi.appMatrix.getAppMenu : cgi.Index.getMenu;
    const menusData = await http.get(url, params, true);

    if (username) {
      localStorage.setItem(lsPrefixData + username, JSON.stringify(menusData.menus));
      localStorage.setItem(lsPrefixTime + username, (+new Date()));
    }
    return menusData.menus;
  } catch (ex) {
    console.log(ex);
    if (username) {
      const strTime = localStorage.getItem(lsPrefixTime + username);
      if (strTime && (+strTime) > ((+new Date()) - (3600 * 1000 * 24 * lsDays))) {
        const strMenu = localStorage.getItem(lsPrefixData + username);
        if (strMenu) {
          return JSON.parse(strMenu);
        }
      }
    }
    throw ex;
  }
}

export function getScopeGroup(params = {}) {
  return http.get(cgi.Index.getScopeGroup, params, true)
    .then(r => r.scopes)
    .catch(r => console.log(r));
}
export function getEdgeLocation(params = {}) {
  return http.get(cgi.Index.getEdgeLocation, params, true)
    .catch(r => console.log(r));
}

let scopeModulesPromise;
export function getScopeModules(params = {}) {
  if (!scopeModulesPromise) {
    scopeModulesPromise = new Promise((resolve) => {
      http.get(cgi.Index.getModules, params, true)
        .then((r) => {
          if (r && r.module_groups) {
            const f = (d) => {
              d.forEach((ele) => {
                if (ele.type === 'mozu') {
                  // eslint-disable-next-line no-param-reassign
                  ele.children = undefined;
                  return;
                }
                if (ele.children && ele.children.length > 0) {
                  f(ele.children);
                }
              });
            };
            f(r.module_groups);
          }
          resolve(r);
        })
        .catch((r) => {
          // resolve({ module_groups: [] });
          resolve({
            module_groups: [],
          });
          console.log(r);
        });
    });
  }
  return scopeModulesPromise;
}

// 获取切换用户权限
export function hasPrivSwitch() {
  return http.get(cgi.Index.hasPrivSwitch, {}, false)
    .catch(r => console.log(r));
}

export function clearDownloadedData(ids) {
  return http.post(cgi.Index.clearDownloadedData, { ids }, true, { isJson: true });
}

// 拉取用户
export async function getSimpleUserList(keywords, vm) {
  return http.post(cgi.Index.getSimpleUserList, {
    start: 0,
    limit: 5,
    keywords: keywords || '',
    fields: [
      'userUid',
      'userName',
      'userRealName',
    ],
  }, false, {
    isJson: true,
  })
    .catch((r) => {
      console.error(r);
      vm.$message.error('拉取用户失败');
    });
}

// 切换用户
export async function switchUser(username, vm) {
  const env = ['local', 'dev', 'test'].includes(vm.env) ? 'test' : 'pub';
  const idcswitchUrl = {
    pub: ``,
    test: ``,
  }[env];

  const r = [idcswitchUrl, cgi.Index.tnidcswitchUser].map(url => http.get(
    url,
    { username },
    true,
    { restAxios: { withCredentials: true } }
  ));
  await Promise.all(r).then(() => {
    location.reload();
  })
    .catch((e) => {
      console.error(e);
      vm.$message.error('切换用户失败');
    });
}

// @todo 后台依赖uid，前端统一模拟
// let mockParam = { UID: 'CQDX11811191540092379376', dn: 'styd' }
export function getTaskList(params = {}, status) {
  // params = { ...params, ...mockParam }
  if (status === 'nebula_fairy') {
    return http.get(cgi.Index.getFairyTaskList, params, true)
      .catch(r => console.log(r));
  }
  return http.post(cgi.Index.getTaskList, params, true)
    .catch(r => console.log(r));
}
export function taskStatusHandle(params = {}) {
  // params = { ...params, ...mockParam }
  return http.get(cgi.Index.getTodoTotal, params, false, true)
    .catch(r => console.log(r));
}

// 获取待办类型
export function getTaskType(params = {}, status) {
  if (status === 'nebula_fairy') {
    return http.get(cgi.Index.getFairyTaskType, params, true)
      .catch(r => console.log(r));
  }
  return http.get(cgi.Index.getTaskType, params, true)
    .catch(r => console.log(r));
}

export function getMsgList(params = {}) {
  // params = { ...params, ...mockParam }
  return http.post(cgi.Index.getMsgList, params, true)
    .catch(r => console.log(r));
}

export function getOpcodeList(/* params = {} */) {
  // params = { ...params, ...mockParam }
  return http.get(cgi.Index.getOpcodeList)
    .catch(r => console.log(r));
}

export function changeStatus(params = {}) {
  // params = { ...params, ...mockParam }
  return http.post(cgi.Index.changeStatus, params, true)
    .catch(r => console.log(r));
}
// 先前端写死，后台接口为开发完成
// export function getTodoType (params = {}) {
//   return http.get(cgi.Index.getTodoType, params, true)
//     .then(r => r.todo_type).catch(r => console.log(r))
// }
// export function doTodo (params = {}) {
//   return http.post(cgi.Index.doTodo, params, true).catch(r => console.log(r))
// }

export function checkBrower(Obj) {
  const sesCheckBrower = ss('_TNBL_CHKBROW_');
  if (!sesCheckBrower) {
    const testChrome = window.navigator.userAgent.match(/Chrome\/(\d*)\./);
    const chromeVer = testChrome ? parseInt(testChrome[1], 10) : 0;
    if (chromeVer && parseInt(testChrome[1], 10) < 63) {
      const alertMes = [
        '您的浏览器需要更新。',
        '推荐您使用最新的<a style="text-decoration: underline;" target="_blank" href="https://pc.qq.com/search.html#!keyword=chrome">Chrome</a> 浏览器，',
        '以获得安全快速的最佳体验。',
      ].join('');
      console.log(`Chrome:${chromeVer}`);
      Obj.$alert(alertMes, '', { dangerouslyUseHTMLString: true });
      ss('_TNBL_CHKBROW_', chromeVer);
      return false;
    }
  }
  return true;
};

// 获取版本更新内容
export function getUpdateContent({ params = {}, loading = true }) {
  return http.post(cgi.Index.getUpdateContent, params, loading, { isJson: true })
    .catch(r => console.log(r));
}

// 已读版本更新
export function readUpdate(params = {}) {
  return http.post(cgi.Index.readUpdate, params, false, { isJson: true })
    .catch(r => console.log(r));
}

// 获取TT待办
export function getTTTodo() {
  return http.post('/cgi/troubleTicket/todo/getList', { type: 'myTodo' }, false, { isJson: true })
    .catch(r => console.log(r));
}

// 获取模组维度的模组列表
export function getUserDmiMozu(params) {
  return http.post(cgi.appMatrix.moduleList, { ...params }, undefined, { isJson: true });
}

// 基于二级菜单\数据域值获取功能列表
export function getAppmatrixOpcodes(params) {
  return http.post(cgi.appMatrix.getAppOpcodes, { ...params }, undefined, { isJson: true });
}
// 基于三级应用和数据域获取当前角色
export function GetRoleByApp(params) {
  return http.post(cgi.appMatrix.GetRoleByApp, { ...params }, undefined, { isJson: true });
}

export function getMenuWhiteList() {
  return http.get(cgi.Index.getCfg, { key: 'appmatrix-whiteMenuList' });
}

export function GetUserPrivilegByApp(params) {
  return http.post(cgi.appMatrix.GetUserPrivilegByApp, { ...params }, undefined, { isJson: true });
}
export function getByUrl(params) {
  return http.post(cgi.appMatrix.getByUrl, { ...params }, undefined, { isJson: true });
}
export function getAllRoleList(params) {
  return http.post(cgi.appMatrix.useRoleAll, { ...params }, undefined, { isJson: true });
}
