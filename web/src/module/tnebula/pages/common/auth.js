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
let gStateIsResetDOM = false;
async function setAuth(node) {
  // eslint-disable-next-line camelcase
  const auth_code = node.getAttribute && node.getAttribute('auth-right-code');
  // eslint-disable-next-line camelcase
  if (!auth_code) {
    const subNodes = node.querySelectorAll && node.querySelectorAll('[auth-right-code]');
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
    op_code = ops.op_code || auth_code;
    isRemove = ops.isRemove;
  }
  // op_code对比有权限的操作码，在找不到操作码的情况下进行设置
  const authedCodes = await getOpCodes();
  // console.log('authedCodes', authedCodes);
  let isAuthed = true;
  // 多个权限时，有一个便说明有权限
  if (op_code.indexOf(',') > 0) {
    // eslint-disable-next-line camelcase
    const item_op_codes = op_code.split(',');
    let hadAuthed = false;
    // eslint-disable-next-line camelcase
    for (const c of item_op_codes) {
      hadAuthed = authedCodes.opcodeList.includes(c);
      if (hadAuthed) {
        break;
      }
    }
    isAuthed = hadAuthed;
  } else {
    isAuthed = authedCodes.opcodeList.includes(op_code);
  }
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
export function observe(containerId) {
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
      attributeFilter: ['auth-right-code'],
      subtree: true,
    };
    mo.observe(container, options);
  }
}
export function checkAll() {
  const nodes = document.querySelectorAll('[auth-right-code]');
  if (nodes) {
    for (const node of nodes) {
      setAuth(node);
    }
  }
}

export async function getAllOpCode() {
  return getOpCodes();
}
