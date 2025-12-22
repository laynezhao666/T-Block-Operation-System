module.exports = (function () {
  const result = {
    apiName: '告警时间线done',
    sourcePath: '/cgi/alarm/ba/getHoursAlarmTrend',
    trueTargetPath: '/cgi/idc-tbos-cgi/alarm/server/GetAlarmCntTrend',
    targetPage: ["all"],
    targetPath: '/Alarm/GetAlarmStatus',
  };
  result.change = function () {
    const res = (newResult) => {
      let count = 0;
      const result = {};
        // 遍历数组，将数据提取出来按照要求存储到新对象中
        newResult?.list.forEach((item) => {
          result[item.u_time] = item.count;
          count += item.count;
        });
        return {
          count,
          map: result,
        };
    };
    const req = (newResult) => {
      // 获取当前时间
      const endTime = new Date();
      // 克隆当前时间对象，并将时间往前推23小时
      const beginTime = new Date(endTime.setMinutes(0, 0, 0)).setTime(endTime.getTime() - 23 * 60 * 60 * 1000);

      return {
        // alarm_type: 0,
        // begin: beginTime / 1000,
        // end: endTime / 1000 + 3600,
        // event_status: -1,
        // interval: 3600,
        mozu_id: newResult.mozuId,
      };
    };
    return { res, req };
  };
  return result;
}());
