module.exports = (function () {
  const result = {
    apiName: '历史下拉(done)',
    sourcePath: '/cgi/alarm/history/getDropdown',
    trueTargetPath: '/cgi/idc-tbos-cgi/alarm/server/GetAlarmName',
    targetPage: ["all"],
  };
  result.change = function () {
    const res = (newResult) => {
      const tempResult = {};
      newResult.list.reduce((acc, item) => {
        acc[item] = item;
        return acc;
      }, tempResult);
      return tempResult;
    };
    const req = (newResult) => {
      let tmpAns = {};
      const mozu_id = newResult.mozuId !== undefined ? newResult.mozuId : 0;
      switch (newResult.eventStatus) {
        // 活动告警（未挂起）
        case 1:
          tmpAns = {
            mozu_id,
            page: 0,
            size: 0,
            alarm_type: 1,
          };
          return {
            mozu_id,
            page: 0,
            size: 0,
            alarm_type: 1,
          };
        // 活动告警（已挂起）
        case -1:
          return {
            mozu_id,
            page: 0,
            size: 0,
            alarm_type: 2,
          };
        // 默认：活动告警（未挂起）
        default:
          tmpAns = {
            mozu_id,
            page: 0,
            size: 0,
            alarm_type: 1,
          };
          return {
            mozu_id: newResult.mozuId,
            page: 0,
            size: 0,
            alarm_type: 1,
          };
      }
    };
    return { res, req };
  };
  return result;
}());
