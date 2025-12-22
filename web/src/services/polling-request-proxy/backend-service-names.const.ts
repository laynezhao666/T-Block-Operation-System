
export const BACKEND_SERVICE_NAMES_MAP = {
  '/cgi/alarm': 'tnhttp-alarmcgi:11116',
  '/cgi/dataQuery/edge': 'tadaptor-point-query',
  '/api/dcos/tdac-cgi': 'dcos-tdac:31234/api/dcos/tdac-cgi',
  '/cgi/idc-tbos-cgi': 'idc-public-gateway:8080',
  '/cgi/alarm/active/getList': 'idc-public-gateway:8080/cgi/idc-tbos-cgi/alarm/server/GetAlarmList',
  '/cgi/dataQuery/edge/getGidAndAttrListValueMapWithoutCache': 'idc-public-gateway:8080/cgi/idc-tbos-cgi/Data/Query',
};

export const POLORIS_BACKEND_SERVICE_NAMES = ['/cgi/idc-tbos-cgi']

// 路径接口都换
export const BACKEND_SVC_CGI_NAMES = ['http://tnhttp-alarmcgi:11116/active/getList"','/cgi/alarm/active/getList','/cgi/dataQuery/edge/getGidAndAttrListValueMapWithoutCache', 'http://tadaptor-point-query/getGidAndAttrListValueMapWithoutCache']