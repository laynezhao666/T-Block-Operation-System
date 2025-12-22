module.exports = (function () {
  const result = {
    apiName: '告警详情',
    sourcePath: '/cgi/alarm/active/detail',
    trueTargetPath: '/cgi/idc-tbos-cgi/alarm/server/GetAlarmDetail',
    targetPage: ['all'],
  };
  result.change = function () {
    const res = (newResult) => {
      if (!newResult) {
        return {};
      }
      const sourceRes = newResult
      return {
        detail: {
          alarm: {
            AlarmId: sourceRes.alarm_id,
            DeviceGid: sourceRes.device_gid,
            OccurTime: sourceRes.occur_time,
            Fingerprint: `${sourceRes.rid};${sourceRes.device_gid}`,
            rid: sourceRes.rid,
            Content: sourceRes.alarm_content,
            DealSuggestion: '',
            InfluenceAnalyze: '',
            alarmLevelZh: sourceRes?.level,
          },
          deviceInfo: {
            DeviceNumber: sourceRes?.device_number,
          },
          ruleInfo: {},
          points: sourceRes.points,
        },
      };
    };
    const req = newResult => ({
      alarm_id: newResult?.AlarmId,
      mozu_id: newResult?.MozuId,
    });
    return { res, req };
  };
  return result;
}());
