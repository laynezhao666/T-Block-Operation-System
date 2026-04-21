/* eslint-disable camelcase */

const BASE_URL = window.location.origin;

/**
 * @important !!!很重要,配菜单一定要注意
 * 就算是只有一个一级菜单也要写三层
 */

// 控制看板权限

window.setShowPanel = function () {
  if (localStorage.getItem('showPanel') && localStorage.getItem('showPanel') === 'true') {
    localStorage.setItem('showPanel', 'false');
  } else {
    localStorage.setItem('showPanel', 'true');
  }
  location.reload();
};

const menus = [
  {
    name: '监控告警',
    href: '/tedge/actived-warning',
    icon: 'alarm',
    showtype: true,
    children: [
      {
        name: '当前告警',

        href: '/tedge/actived-warning',
        icon: 'notice',
        showtype: true,
        children: [
          {
            name: '活动告警',
            href: '/tedge/actived-warning',
            showtype: true,
          },
          {
            name: '告警详情',
            href: '/tedge/clt-dev-qa',
            showtype: false,
          },
          {
            name: '告警详情',
            href: '/tedge/warning-detail',
            showtype: false,
          },
          {
            name: '导入工具',
            href: '/tedge/import-tools',
            showtype: false,
          },
        ],
      },
      {
        name: '已挂起告警',
        href: '/tedge/hangup-warning',
        showtype: true,
        icon: 'label',
        children: [
          {
            name: '已挂起告警',
            href: '/tedge/hangup-warning',
            showtype: true,
          },
        ],
      },

      {
        name: '历史告警',
        icon: 'history',
        href: '/tedge/warning-history',
        showtype: true,
        children: [
          {
            name: '历史告警',
            href: '/tedge/warning-history',
            showtype: true,
          },
        ],
      },

      {
        name: '告警策略',
        icon: 'tips',
        href: '/tedge/warning-strategy',
        showtype: true,
        children: [
          {
            name: '告警策略',
            href: '/tedge/warning-strategy',
            showtype: true,
          },
          {
            name: '策略生效验证',
            href: '/tedge/warning-effective-verify',
            showtype: true,
          },
          {
            name: '策略生效验证详情',
            href: '/tedge/warning-strategy-detail',
            showtype: false,
          },
        ],
      },
    ],
  },
  {
    name: '安防系统',
    href: '/tisspage/im-security-camera',
    icon: 'shield',
    showtype: true,
    children: [
      {
        name: '门禁管理',
        icon: 'door',
        href: '/tedge/door-system',
        showtype: true,
        children: [
          {
            name: '控制器及门状态',
            href: '/tedge/doors-overview',
            showtype: true,
          },
          {
            name: '门禁事件',
            href: '/tedge/security-records',
            showtype: true,
          },
          {
            name: '授权发卡',
            href: '/tedge/security-auth-setting',
            showtype: true,
          },
          {
            name: '时间组',
            href: '/tedge/security-time-period-setting',
            showtype: true,
          },
          {
            name: '异步消息',
            href: '/tedge/security-requests',
            showtype: true,
          },
        ],
      },
    ],
  },
  {
    name: '数据查询',
    icon: 'monitor',
    href: '/tedge/data-query-index',
    showtype: true,
    children: [
      {
        name: '数据查询',
        href: '/tedge/data-query-index',
        showtype: true,
        children: [
          {
            name: '数据查询',
            href: '/tedge/data-query-index',
            showtype: true,
          },
        ],
      },
    ],
  },
  {
    name: '系统配置',
    icon: 'system-config',
    href: '/tedge/device-manage',
    showtype: true,
    children: [
      {
        showtype: true,
        name: '设备管理',
        href: '/tedge/device-manage',
        icon: 'device-manage',
        children: [
          {
            name: '设备管理',
            href: '/tedge/device-manage',
            showtype: true,
          },
        ],
      },
      {
        showtype: true,
        name: '模组配置管理',
        href: '/tedge/mozu-config-manage',
        icon: 'view-module',
        children: [
          {
            name: '模组配置管理',
            href: '/tedge/mozu-config-manage',
            showtype: true,
          },
        ],
      },
    ],
  },

];

// 一级菜单是3位，从100开始
// 二级菜单一共6位，前3位是一级菜单，后面3位是二级
// 三级菜单一共9位，前3位是一级菜单，中间3位是二级，后面3位是3级
function getId(pid, index) {
  pid = pid || '';
  return `${pid}${100 + index}`;
}

const defaultOptions = {
  base_url: BASE_URL,
  a_code: 'tnebula',
  n_createtime: '2020-01-01 00:00:00',
  n_opcode: '',
  n_target: '',
  n_licls: 'assets',
  n_containercls: '',
  n_containerhtml: '',
  n_scope: 0,
};

function generateMenu(m, datas, pid) {
  datas.forEach((item, index) => {
    const n_haschild = !!((item.children && item.children.length > 0));
    const n_pid = pid || '0';
    const n_level = pid ? (pid.length === 3 ? 2 : 3) : 1;
    const n_order = index + 1;
    const n_licls = item.icon || '';
    const n_showtype = item.showtype;
    const n_name = item.name;
    const n_href = item.href;
    const n_id = getId(pid, n_order);
    let menu = {
      n_haschild,
      n_pid,
      n_level,
      n_order,
      n_licls,
      n_showtype,
      n_name,
      n_href,
      n_id,
    };
    menu = Object.assign({}, defaultOptions, menu);
    m.push(menu);
    if (n_haschild) {
      generateMenu(m, item.children, n_id);
    }
  });
}
const results = [];
generateMenu(results, menus);

export default {
  code: 0,
  message: 'successful',
  data: {
    menus: results,
  },
};
