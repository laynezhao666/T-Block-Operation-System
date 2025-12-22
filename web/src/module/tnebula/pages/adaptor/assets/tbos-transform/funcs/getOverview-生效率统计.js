module.exports = (function () {
  const result = {
    apiName: '生效率统计done',
    sourcePath: '/cgi/alarm/validate/realtime/getOverview',
    trueTargetPath: '/cgi/idc-tbos-cgi/alarm/server/GetValidate',
    targetPage: ["all"],
  };
  result.change = function () {
    const res = (newBackResult) => {
      const newResult = JSON.parse(JSON.stringify(newBackResult));
      const keyMap = {
        all: 'total',
        valid: 'valid',
        invalid: 'invalid',
        fired: 'fired',
        unfired: 'unfired',
        efficiency: 'validPercent',
      };
      const newObj1 = {
      };
      for (const key in keyMap) {
        if (Object.prototype.hasOwnProperty.call(newResult?.metrics, key)) {
          if (key === 'efficiency') {
            newObj1[keyMap[key]] = newResult?.metrics[key];
          } else {
            newObj1[keyMap[key]] = newResult?.metrics[key];
          }
        }
      }
      const rspData1 = {
        total: { ...newObj1 },
        custom: newObj1,
        standard: newObj1,
      };
      return rspData1;
    };
    const req = (newResult) => {
      const reqTransformMap = {
        limit: 'size',
        alarmLevel: 'level',
        alarmType: 'alarm_name',
        occurTimeStart: 'begin',
        occurTimeEnd: 'end',
        deviceGid: 'gid',
        validateType: 'error_code',
        mozuId: 'mozu_id',
        isStandard: 'rule_type',
        fired: 'valid_type',
      };
      function convertObjectKeys(obj, keyMapping) {
        const convertedObj = {
        };
        for (const key in keyMapping) {
          if (Object.prototype.hasOwnProperty.call(obj, key)) {
            convertedObj[keyMapping[key]] = obj[key];
          }
        }
        return convertedObj;
      }
      const reqData = convertObjectKeys(newResult, reqTransformMap);
      const extra = {
        page: newResult.offset / newResult.limit + 1,
      };
      return { ...reqData, ...extra };
    };
    return { res, req };
  };
  return result;
}());
