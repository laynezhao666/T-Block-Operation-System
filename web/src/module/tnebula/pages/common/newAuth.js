import * as reqhel from './reqhel.js';
import { codeConf } from '@@/config/authCode';

let authCodePromise;

function getOpCodes() {
  if (!authCodePromise) {
    authCodePromise = new Promise((resolve) => {
      reqhel
        .getOpcodeList()
        .then((data) => {
          resolve(data);
        })
        .catch(() => {
          resolve({ opcodeList: [] });
        });
    });
  }
  return authCodePromise;
}
let allRoleListPromise;
function getRoleList() {
  if (!allRoleListPromise) {
    allRoleListPromise = new Promise((resolve) => {
      reqhel
        .getAllRoleList()
        .then((data) => {
          resolve(data);
        })
        .catch(() => {
          resolve([]);
        });
    });
  }
  return allRoleListPromise;
}

function getRoleAuthType() {
  if (window.currentRole > -1) {
    return window.currentRole;
  }
  if (!authCodePromise) {
    if (localStorage.getItem('getToleAuthTypesParams')) {
    // const params = {};
    // params.appID = 5038;
    // params.dimValueID = 1;
      const params = JSON.parse(localStorage.getItem('getToleAuthTypesParams'));
      authCodePromise = new Promise((resolve) => {
        reqhel.GetUserPrivilegByApp(params).then((result) => {
          if (localStorage.getItem('showAuthLog') === 'true') {
            console.log('**********auth-directive请求**********');
          }
          window.currentRole = result.role.roleType;
          resolve(result);
        })
          .catch(() => {
            resolve(-1);
          });
      });
    } else {
      authCodePromise = new Promise((resolve) => {
        window.currentRole = -1;
        resolve(-1);
      });
    }
  }
  return authCodePromise;
}

let gStateIsResetDOM = false;
async function setAuth(node) {
  // eslint-disable-next-line camelcase
  const auth_code = node.getAttribute && node.getAttribute('auth-roles');
  // eslint-disable-next-line camelcase
  if (!auth_code) {
    const subNodes = node.querySelectorAll && node.querySelectorAll('[auth-roles]');
    if (subNodes && subNodes.length > 0) {
      for (const n of subNodes) {
        setAuth(n);
      }
    }
    return;
  }
  const conf = codeConf;
  // eslint-disable-next-line camelcase
  let op_code = auth_code;
  let isRemove;
  const ops = conf[auth_code];
  if (ops) {
    // eslint-disable-next-line camelcase
    op_code = ops.op_code || auth_code || [];
    isRemove = ops.isRemove;
  }
  // op_code对比有权限的操作码，在找不到操作码的情况下进行设置
  let currentRole = null;

  currentRole = await getRoleAuthType();

  let isAuthed = true;
  // 多个权限时，有一个便说明有权限
  //   if (op_code.indexOf(',') > 0) {
  //     // eslint-disable-next-line camelcase
  //     const item_op_codes = op_code.split(',');
  //     let hadAuthed = false;
  //     // eslint-disable-next-line camelcase
  //     for (const c of item_op_codes) {
  //       hadAuthed = currentRole.opcodeList.includes(c);
  //       if (hadAuthed) {
  //         break;
  //       }
  //     }
  //     isAuthed = hadAuthed;
  //   } else {
  // console.log(op_code, 'op_codeop_codeop_code');
  // console.log(op_code, currentRole, op_code.includes(currentRole), 'op_code.includes(currentRole)');
  isAuthed = op_code.includes(currentRole);

  //   }
  if (isAuthed) {
    return;
  }
  if (isRemove) {
    node.remove();
    return;
  }
  // 有特殊的配置
  if (ops) {
    const { setDisplay } = ops;
    const { setVisibility } = ops;
    const { setClass } = ops;
    const { setDisabled } = ops;
    const { isResetDOM } = ops;
    if (setVisibility || setClass || setDisabled !== undefined) {
      if (setDisplay) {
        node.style.display = setDisplay;
      }
      if (setVisibility) {
        node.style.visibility = setVisibility;
      }
      if (setDisabled !== undefined) {
        node.style.disabled = setDisabled;
      }
      if (setClass) {
        node.classList.add(setClass);
      }
      if (isResetDOM) {
        // eslint-disable-next-line no-self-assign
        node.outerHTML = node.outerHTML;
        gStateIsResetDOM = true;
      }
    } else {
      node.style.display = setDisplay || 'none';
    }
  } else {
    // 默认操作
    node.style.display = 'none';
  }
}
// eslint-disable-next-line camelcase
function obserer_callback(records) {
  if (gStateIsResetDOM) {
    gStateIsResetDOM = false;
    return;
  }
  for (const r of records) {
    if (r.type === 'childList') {
      for (const node of r.addedNodes) {
        setAuth(node);
      }
    } else if (r.type === 'attributes') {
      setAuth(r.target);
    }
  }
}
export function observeNew(containerId) {
  const MutationObserver = window.MutationObserver
    || window.WebKitMutationObserver
    || window.MozMutationObserver;

  const supportMutationObserver = !!MutationObserver;
  if (supportMutationObserver) {
    const mo = new MutationObserver(obserer_callback);
    let container;
    if (containerId) {
      container = document.getElementById(containerId);
    }
    container = container || document.documentElement;
    const options = {
      childList: true,
      attributes: true,
      attributeFilter: ['auth-roles'],
      subtree: true,
    };
    mo.observe(container, options);
  }
}
export function checkAllNew() {
  const nodes = document.querySelectorAll('[auth-roles]');
  if (nodes) {
    for (const node of nodes) {
      setAuth(node);
    }
  }
}

export async function getAllOpCode() {
  return getOpCodes();
}

export async function getAllRoleList() {
  return getRoleList();
}
