module.exports = (function () {
  const result = {
    apiName: '告警等级接口统计(done)',
    sourcePath: '/cgi/alarm/active/getActiveOverview',
    trueTargetPath: '/cgi/idc-tbos-cgi/alarm/server/GetAlarmList',
    targetPage: ["all"],
  };
  result.change = function () {
    const res = (newResult) => {
      // 无法直接使用 key map，特殊化处理
      const rspData = {
        list: {
          L0: newResult.metrics.level.list.find(item => item.name == 'L0').count,
          L1: newResult.metrics.level.list.find(item => item.name == 'L1').count,
          L2: newResult.metrics.level.list.find(item => item.name == 'L2').count,
          L3: newResult.metrics.level.list.find(item => item.name == 'L3').count,
          L4: newResult.metrics.level.list.find(item => item.name == 'L4').count,
          L5: 0,
          all: newResult.total,
        },
      };
      return rspData;
    };
    const req = (newResult) => {
      const newReq = {
        alarm_type: 1, // alarm_type字段
        count_by_metric: true
      };
      const keyMap = {
        alarmStatus: 'alarm_type',
        eventStatus: 'event_status',
        mozuId: 'mozu_id',
      };
        // 遍历判断是否有字段
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
