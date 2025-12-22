module.exports = (function () {
  const result = {
    apiName: '告警策略：①告警类型下拉 ②设备类型下拉',
    sourcePath: '/cgi/alarm/rule/getDropdown',
    trueTargetPath: '/cgi/idc-tbos-cgi/alarm/server/GetAlarmName',
    targetPage: ["all"],
    targetPath: '/alarm/server/GetAlarmName',
  };
  result.change = function () {
    const res = (newResult) => {
      const rspData = {};
      newResult.list.reduce((acc, item) => {
        acc[item] = item;
        return acc;
      }, rspData);
      return rspData;
    };
    const req = (newResult) => {
      const newReq = {
        alarm_type: 0,
        page: 0,
        size: 0,
      };

      // 字段映射
      const keyMap = {
        alarmStatus: 'alarm_type',
        start: 'page',
        limit: 'size',
      };
      // 遍历判断是否有字段并做映射
      for (const oldKey in keyMap) {
        if (Object.prototype.hasOwnProperty.call(newResult, oldKey)) {
          newReq[keyMap[oldKey]] = newResult[oldKey];
        }
      }
      return newReq;
    };
    return { res, req };
  };
  return result;
}());
