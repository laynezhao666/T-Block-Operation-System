import * as dayjs from 'dayjs';
import * as Cookies from 'js-cookie';

const ADMIN_LOGINED_COOKIE_FIELD = 'tedge_admin_logined';

/**
 * 简易的管理员操作密码验证
 */
export class LoginStatusService {
  private _adminLogined: boolean = false;

  constructor() {
    this.refreshLoginStatusByCookie();
    this.startKeepRefreshInterval();
  }

  get adminLogined() {
    return this._adminLogined;
  }

  async fetchRightPassword() {
    const axios = (window.Vue as any).prototype.$axios;

    const config = await axios.get('/cgi/tedge-bff/user-custom-config/get', {
      id: 'simpleAdminPwd',
    }, false);

    const pwd = config?.content?.content?.trim() || '';

    return pwd || '';
  }

  startKeepRefreshInterval() {
    setInterval(() => {
      this.refreshLoginStatusByCookie();
    }, 5000);
  }

  refreshLoginStatusByCookie() {
    if (this._adminLogined === Cookies.get(ADMIN_LOGINED_COOKIE_FIELD)) return;

    this._adminLogined = Cookies.get(ADMIN_LOGINED_COOKIE_FIELD);
  }

  async login() {
    const rightPwd = await this.fetchRightPassword();

    let isSuccess = false;

    while (!isSuccess) {
      let isCancel = false;
      let pwd = '';

      await (window.Vue as any).prototype.$prompt('请输入管理员密码', '管理员登录', {
        confirmButtonText: '登录',
        cancelButtonText: '取消',
        inputPattern: /^\w{6,}$/,
        inputErrorMessage: '请输入至少6位长度的密码',
      }).then(({ value }) => {
        pwd = value;
      })
        .catch(() => {
          pwd = null;
          isCancel = true;
        });

      if (isCancel) return;

      if (pwd === rightPwd) {
        isSuccess = true;
      } else {
        (window.Vue as any).prototype.$message('密码错误，充重新输入');
      }
    }

    const loginTs = new Date().getTime();

    const adminLoginExpireMinutes = (await window.tnwebServices.customConfigService.get('adminLoginExpireMinutes'))
      || 60;
    const expires = dayjs().add(adminLoginExpireMinutes, 'minutes').toDate();

    Cookies.set(ADMIN_LOGINED_COOKIE_FIELD, loginTs, {
      path: '',
      expires,
    });
    (window.Vue as any).prototype.$message.success(`登录成功，有效期${adminLoginExpireMinutes}分钟`);

    this.refreshLoginStatusByCookie();
  }

  logout() {
    Cookies.remove(ADMIN_LOGINED_COOKIE_FIELD, {
      path: '',
    });

    (window.Vue as any).prototype.$message.success('已退出管理员账号');

    this.refreshLoginStatusByCookie();
  }

  hasRight(shouldConfirmIfNoRight: boolean = false) {
    if (this.adminLogined) return true;

    if (shouldConfirmIfNoRight) {
      (window.Vue as any).prototype.$confirm('该操作需要登录管理员身份，请先登录后再次操作。', '无权限', {
        inputType: 'password',
        confirmButtonText: '登录',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        this.login();
      }).catch(() => {});
    }

    return false;
  }

  async resetPassword(password: string) {
    if (!password?.trim()) return false;

    const axios = (window.Vue as any).prototype.$axios;

    const config = await axios.get('/cgi/tedge-bff/user-custom-config/get', {
      id: 'simpleAdminPwd',
    });

    if (config) {
      config.content.content = password;
      await axios.post('/cgi/tedge-bff/user-custom-config/update', config);
      return true;
    }

    const newConfig = {
      id: 'simpleAdminPwd',
      key: 'simpleAdminPwd',
      label: '管理员密码',
      desc: null,
      category: null,
      url: null,
      moduleId: (window as any).__GetFrameDataByKey("curMozuData")?.id,
      preload: true,
      enable: true,
      contentType: 'Text',
      createdAt: '2023-08-01T07:37:17.143Z',
      updatedAt: '2023-08-01T07:37:17.143Z',
      content: {
          configId: 'simpleAdminPwd',
          type: 'Text',
          content: password,
          version: 2,
      },
    };

    await axios.post('/cgi/tedge-bff/user-custom-config/create', newConfig);
    return true;
  }
};
