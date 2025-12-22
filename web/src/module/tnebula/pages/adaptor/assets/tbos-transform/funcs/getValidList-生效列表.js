module.exports = (function () {
  const result = {
    apiName: '生效列表done',
    sourcePath: '/cgi/alarm/validate/realtime/getValidList',
    trueTargetPath: '/cgi/idc-tbos-cgi/alarm/server/GetValidate',
    targetPage: ['all'],
  };
  result.change = function () {
    const res = (newResult) => {
      const keyMap = {
        deviceGid: 'device_gid',
        rid: 'rid',
        validateType: 'error_code',
        alarmLevel: 'level',
        deviceNumber: 'device_number',
        errorMsg: 'error_detail',
        alarmType: 'alarm_name',
        validateTypeStr: 'error_name',
        lastRuntime: 'eval_time',
        alarmContent: 'content',
      };

      const rspData = {
        count: newResult?.total,
        list: [],
      };
      rspData.list = newResult?.list?.map((newItem) => {
        const oldItem = {};
        for (const oldKey in keyMap) {
          if (Object.prototype.hasOwnProperty.call(newItem, keyMap[oldKey])) {
            oldItem[oldKey] = newItem[keyMap[oldKey]];
          }
        }
        return oldItem;
      });
      return rspData;
    };
    const req = (newResult) => {
      const reqTransformMap = {
        limit: 'size',
        offset: 'page',
        alarmLevel: 'level',
        alarmType: 'alarm_name',
        occurTimeStart: 'begin',
        occurTimeEnd: 'end',
        deviceGid: 'device_gid',
        validateType: 'error_code',
        mozuId: 'mozu_id',
        isStandard: 'rule_type',
        fired: 'valid_type',
      };
      if (!Object.prototype.hasOwnProperty.call(newResult, 'fired')) {
        newResult.fired = true;
      }
      function convertObjectKeys(obj, keyMapping) {
        const convertedObj = {
        };
        for (const key in keyMapping) {
          if (Object.prototype.hasOwnProperty.call(obj, key)) {
            if (key === 'offset') {
              convertedObj[keyMapping[key]] = Math.ceil(obj[key] / obj.limit) + 1;
            } else if (key === 'isStandard') {
              convertedObj[keyMapping[key]] = obj[key] === '0' ? 0 : 0;
            } else if (key === 'fired') {
              convertedObj[keyMapping[key]] = obj[key] === true ? 1 : 2;
            } else if (key === 'alarmType' || key === 'deviceGid' || key === 'validateType' || key === 'alarmLevel') {
              convertedObj[keyMapping[key]] = [obj[key]];
            } else {
              convertedObj[keyMapping[key]] = obj[key];
            }
          }
        }
        return convertedObj;
      }
      const reqData = convertObjectKeys(newResult, reqTransformMap);
      return reqData;
    };
    return { res, req };
  };
  return result;
}());
