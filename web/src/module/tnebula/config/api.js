import { ENV_NAME } from 'common/script/passport_login';

const pathPre = '/cgi/dcom';

const wsHOST = {
  local: '',
  dev: '',
  test: ``,
  pre: '',
  publish: '',
};

// 页面api
export const cgi = {
  Index: {
    // getMenu: `${pathPre}/common/Menu/GetGrantedMenus`,
    getMenu: '/cgi/menu/getGrantedMenus',
    getScopeGroup: `${pathPre}/common/Privilege/GetGrantedScopes`,
    getModules: '/cgi/account/getGrantedScopes',
    getEdgeLocation: '/cgi/dataQuery/edge/getEdgeLocation',
    getTaskList: `${pathPre}/common/todo/gettask`,
    getMsgList: `${pathPre}/siteMessages/message/getList`, // 消息中心获取消息
    changeStatus: `${pathPre}/siteMessages/message/changeStatus`, // 消息中心修改消息状态
    // getOpcodeList: `${origin}/cgi/go/cmdb/getUserOpcode?systemId=${systemId}`, // 获取权限码
    getOpcodeList: '/cgi/userinfo/privilege/opcode/list', // 获取权限码 统一接口
    getTaskType: `${pathPre}/common/todo/gettasktype`, // 获取待办类型
    getFairyTaskList: '/cgi/taskCenter/getTask', // 获取待办
    getFairyTaskType: '/cgi/taskCenter/getTypeCategory', // 获取fairy待办类型
    getTodoTotal: '/cgi/taskCenter/getTodoTotal', // 获取待办数量修改铃铛状态

    // 获取只有name、uid、phone的接口
    getSimpleUserList: '/cgi/userinfo/account/all',
    // 获取切换用户权限
    hasPrivSwitch: '/cgi/tnpassport/hasPrivSwitch',
    // 切换用户
    idcswitchUser: '/cgi/switch',
    tnidcswitchUser: '/cgi/tnpassport/switch',
    getUpdateContent: '/cgi/appmatrix/version/getversions',
    readUpdate: '/cgi/appmatrix/version/update',

    clearDownloadedData: '/cgi/bigfile/clear', // 清除已下载的文件
    downloadFile: '/cgi/bigfile/download', // 下载文件
    getDownloadFileList: `${wsHOST[ENV_NAME]}/cgi/ws/bigfile/connect`, // 获取下载文件列表
    getCfgByReferer: '/cgi/frontDict/getCfgByReferer',
    getCfg: '/cgi/frontDict/getCfg',
  },
  appMatrix: {
    getAppMenu: '/cgi/appmatrix/userApp/list',
    getAppOpcodes: '/cgi/appmatrix/userApp/getOpcodes',
    // 获取数据域维度
    getDim: '/cgi/appmatrix/userDim/get',
    moduleList: '/cgi/appmatrix/userDim/moduleList',
    getAllOpcodes: '/cgi/appmatrix/userApp/getAllOpcodes',
    GetRoleByApp: '/cgi/appmatrix/userApp/GetRoleByApp',
    GetUserPrivilegByApp: '/cgi/appmatrix/userApp/GetUserPrivilegByApp',
    getByUrl: '/cgi/appmatrix/app/getByUrl',
    useRoleAll: '/cgi/appmatrix/useRole/all',
  },
};

// 页面跳转url
export const url = {
};

const websocketProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';

let wshost = `${websocketProtocol}://${location.host}`;
let alarmHost = wshost;
let haHost = wshost;

export const wsScreen = {
  screenBizIndicator: `${wshost}/ws/screenBizIndicator`,
  quaScreenIndicator: `${wshost}/ws/quaScreenIndicator`,
  alarm: `${alarmHost}/ws/alarm`,
  tbosAlarm: `ws://${location.hostname}:8081/ws`,
  ha: `${haHost}/ws/HA`,
};
