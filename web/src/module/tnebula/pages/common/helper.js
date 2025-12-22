import * as reqhel from './reqhel';
import qs from 'qs';

// 2021-05-13 后台无人投入，暂时前端配置，后期可以在boss系统维护此字段
const mozuUrl = {
  showMozu: [
    '/timpage/im-security',
    '/timpage/complex-view',
    // 监控管理
    '/timpage/actived-warning',
    '/timpage/warning-history',
    '/timpage/warning-strategy',
    '/timpage/warning-notify-config',
    '/timpage/warning-notify-add',
    '/timpage/warning-conv',

    '/timpage/electric-hvac-system',
    '/timpage/electric-hvac-system-second',

    '/timpage/electric-relation',
    '/timpage/electric-sld',
    '/timpage/electric-plan',
    '/timpage/electric-system',

    '/timpage/data-query-index',
    '/timpage/data-query-by-collect',
    '/timpage/advanced-search',

    '/timpage/hangup-warning',
    '/timpage/strategy-invalid-history',
    '/timpage/strategy-effective-verify',
    '/timpage/warning-strategy-detail',
    '/timpage/leak-system',

    // 设施运维
    '/tompage/patrol-overview-index',
    '/tompage/patrol-overview-detail',
    '/tompage/patrol-task-list',
    '/tompage/patrol-plan-index',
    '/tompage/patrol-template-list',
    '/tompage/patrol-template-add',
    '/tompage/patrol-point-index',
    '/tompage/patrol-logicarea-addlogic',
    '/tompage/patrol-logicarea-index',
    '/tompage/patrol-config-index',

    '/tompage/event-order-list',
    '/tompage/event-follow-list',
    '/tompage/event-repair-list',

    '/tompage/overview-maintain-index',
    '/tompage/maintain-template-index',
    '/tompage/maintain-template-create',

    '/tompage/maintain-task-index',
    '/tompage/maintain-logicarea-index',
    '/tompage/maintain-logicarea-addlogic',
    '/tompage/maintain-item-index',
    '/tompage/maintain-config-index',
    '/tompage/maintain-plan-manage',

    '/tompage/duty-group',
    '/tompage/duty-userduty',
    '/timpage/capacity-dashboard',
    '/timpage/capacity-capacity',
    '/timpage/rack-capacity',

    '/tompage/tools-maintainconfig-planconfig',

    '/tompage/lmc-dashboard-list',
    '/equipment/battery-overview',
  ],
  showAllMozu: [
    '/tompage/event-order-list',
    '/tompage/event-follow-list',

    '/tompage/event-repair-list',
  ],

};

// 全部模组选项
const allMozuItem = {
  id: '全部',
  name: '全部区域',
  type: 'region',
  children: [{
    id: '全部',
    name: '全部园区',
    type: 'campus',
    children: [
      {
        id: '全部',
        name: '全部模组',
        alias: '全部',
        type: 'mozu',
      },
    ],
  }],
};

export function initMozuByUrl(keyPath, vm) {
  const that = vm;

  if (!checkMozuSelectVisible(keyPath)) {
    that.mozuSelectVisible = false;
    return false;
  }

  that.mozu_data = refreshMozuData(that.curUserMozuData, keyPath);

  that.curMozuId = initCurMozuId(that.mozu_data);
  if (that.curMozuId) {
    // eslint-disable-next-line no-underscore-dangle
    setCurMozuCookie(checkMozuIdIsExist(that.mozu_data, that.curMozuId, true));

    // 暂时不修改url
    // refreshUrlMozuId(that.curMozuId);

    that.mozuSelectVisible = true;
  }

  return true;
};

// 是否展示模组显示框
export function checkMozuSelectVisible(pathname) {
  const keyPath = pathname.replace('.html', '');
  let showFlag = false;
  mozuUrl.showMozu.map((url) => {
    if (keyPath === url) {
      showFlag = true;
    }
    return true;
  });
  return showFlag;
}

// 是否带全部选项
export function refreshMozuData(mozuData, pathname) {
  const keyPath = pathname.replace('.html', '');
  return mozuUrl.showAllMozu.includes(keyPath) && mozuData[0].id !== '全部' ? [allMozuItem, ...mozuData] : mozuData;
}

// 初始化mozuId，优先级 url ${mozuId} > cookie > 默认第一个
export function initCurMozuId(mozuData) {
  // url
  const urlMozuId = qs.parse(location.search, { ignoreQueryPrefix: true })?.mozuId;
  if (urlMozuId && checkMozuIdIsExist(mozuData, urlMozuId)) {
    return urlMozuId;
  }
  // cookie
  const cookieMozuId = reqhel.cs('tnebula_cu_moduleid');
  if (cookieMozuId && checkMozuIdIsExist(mozuData, cookieMozuId)) {
    return cookieMozuId;
  }

  // 默认第一个
  return mozuData[0].children[0].children[0].id;
}

export function checkMozuIdIsExist(mozuData, id, rtnCurMozu = false) {
  let flag = false;
  let curMozuData = {};
  mozuData.map((region) => {
    region.children.map((campus) => {
      campus.children.map((mozu) => {
        if (`${mozu.id}` === `${id}`) {
          flag = true;
          curMozuData = mozu;
        }
      });
    });
  });
  if (rtnCurMozu === true) {
    return curMozuData;
  }
  return flag;
}

// 设置模组
export function setCurMozuCookie(mozuData) {
  reqhel.cs('tnebula_cu_moduleid', mozuData.id);
  reqhel.cs('tnebula_cu_modulename', mozuData.name);
  reqhel.cs('tnebula_cu_modulealias', mozuData.alias);
  reqhel.cs('tnebula_cu_modulesource', mozuData.source);
}

// 检测当前路径是否有mozuId，如果存在切换前需要修改
export async function refreshUrlMozuId(curMozuId, isGo) {
  const urlMozuId = qs.parse(location.search, { ignoreQueryPrefix: true })?.mozuId;

  // url存在模组需要修改
  if (urlMozuId) {
    const paramObj = { mozuId: curMozuId };
    const search = qs.parse(location.search, { ignoreQueryPrefix: true });
    Object.assign(search, paramObj);
    const href = `${location.pathname}?${qs.stringify(search)}${location.hash}`;
    if (isGo) {
      location.href = href;
    } else {
      history.pushState(paramObj, '', href);
    }
  } else {
    location.reload();
  }
}

export function dataToTree(list) {
  const treeList = list.reduce((prev, cur) => {
    prev[cur.n_id] = cur;
    return prev;
  }, {});

  return list.reduce((prev, cur) => {
    const pid = cur.n_pid;
    const parent = treeList[pid];
    if (parent) {
      parent.children ? parent.children.push(cur) : parent.children = [cur];
    } else if (pid === 0) {
      prev.push(cur);
    }
    return prev;
  }, []);
}
export function flatten(list) {
  return list.reduce((prev, cur) => {
    const { children, ...i } = cur;
    return prev.concat(i, children && children.length ? flatten(children) : []);
  }, []);
}

export default {};
