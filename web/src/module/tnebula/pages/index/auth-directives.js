import * as reqhel from '../common/reqhel.js';

let authCodePromise;

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

// 全局获取opcodeList
if (!window.TNBL) {
  window.TNBL = {};
}
window.TNBL.getRoleAuthType = getRoleAuthType;

const checkAuth = async (el, binding) => {
  if (window.hasAppAuth) {
    if (binding.value) {
      const auths = binding.value;
      const roleType = await getRoleAuthType();
      let hadAuthed = true;
      if (localStorage.getItem('showAuthLog') === 'true') {
        console.log(binding, 'binding');
        console.log(auths, 'auths角色列表');
        console.log(roleType, 'roleType当前角色');
        console.log(auths.includes(roleType), '是否有权限');
      }

      hadAuthed = auths.includes(roleType);

      if (!hadAuthed) {
        // el.parentNode && el.parentNode.removeChild(el);
        el.style.display = 'none';
      } else {
        el.style.display = '';
      }
    }
  }
};

export default {
  appMatrixAuth: {
    inserted(el, binding, v) {
      checkAuth(el, binding, v);
    },
    componentUpdated(el, binding, v) {
      if (binding.oldValue !== binding.value) {
        checkAuth(el, binding, v);
      }
    },
  },
};
