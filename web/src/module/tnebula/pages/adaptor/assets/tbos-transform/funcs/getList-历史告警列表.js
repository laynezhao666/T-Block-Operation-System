module.exports = (function () {
  const result = {
    "apiName": "历史告警列表done",
    "sourcePath": "/cgi/alarm/history/getList",
    "trueTargetPath": "/cgi/idc-tbos-cgi/alarm/server/GetAlarmList",
    "targetPage": ['all'],
    "targetPath": "/alarm/server/GetAlarmList"
  };
  result.change = function () {
    const res = (newResult) => {
      const keyMap = {
        "alarmId": "alarm_id",
        "occurTime": "occur_time",
        "restoreTime": "restore_time",
        "content": "alarm_content",
        "alarmContent": "alarm_content",
        "alarmType": "alarm_name",
        "level": "level",
        "alarmLevel": "level",
        "alarmStatus": 'alarm_status',
        "eventStatus": 'event_status',
        "mozuId": 'mozu_id',
        "mozuName": 'mozu_name',
        "roomName": 'room',
        "boxName": 'box',
        "deviceGid": "device_gid",
        "deviceNumber": "device_number",
        "deviceType": "device_type_zh",
        "hangupReason": "hangup_reason",
      };
      const rspData = {
        count: newResult?.total !== undefined ? newResult.total : 0,
        list: [],
      };
      rspData.list = newResult?.list?.map((item) => {
        const oldItem = {};
        for (const oldKey in keyMap) {
          if (Object.prototype.hasOwnProperty.call(item, keyMap[oldKey])) {
            oldItem[oldKey] = item[keyMap[oldKey]];
          }
        }
        // occurPointList 需要特殊处理

        // position 字段需要特殊处理
        oldItem.position = item[keyMap["roomName"]] + "/" + item[keyMap["boxName"]];

        // 返回数据
        return oldItem;
      });
      return rspData;
    };
    const req = (newResult) => {
      let newReq = {
        "alarm_type": 3, // alarm_type字段
      };
      let levelMap = {
        "零级": "L0",
        "一级": "L1",
        "二级": "L2",
        "三级": "L3",
        "四级": "L4",
        "五级": "L5"
      }
      const keyMap = {
        "alarmStatus": "alarm_type",
        "eventStatus": "event_status",
        "limit": "size",
        "level": "level",
        "mozuId": "mozu_id",
        "sortedType": "sort_type",
        "occurTimeStart": "occur_begin",
        "occurTimeEnd": "occur_end",
        "restoreTimeEnd": "restore_end",
        "restoreTimeStart": "restore_begin",
        alarmTypes: "alarm_name"
      };
      // 遍历判断是否有字段
      for (const oldKey in keyMap) {
        if (Object.prototype.hasOwnProperty.call(newResult, oldKey)) {
          newReq[keyMap[oldKey]] = newResult[oldKey];
        }
      }
      // page 字段特殊处理
      if (newResult?.offset !== undefined && newResult?.limit !== undefined && newResult?.limit !== 0) {
        newReq.page = newResult.offset / newResult.limit + 1;
      }

      // sort_type 字段特殊处理
      if (newResult?.sortedType !== undefined) {
        newReq.sort_type = newResult.sortedType == 0 ? 1 : 2
      }
      if (newResult?.deviceGid) {
        newReq.device_gid = typeof newResult.deviceGid === 'string' ? [newResult.deviceGid] : newResult.deviceGid
      }

      if (newResult?.deviceGids) {
        newReq.device_gid = typeof newResult.deviceGids === 'string' ? [newResult.deviceGids] : newResult.deviceGids
      }
      if (newResult?.fingerprint) {
        newReq.rid = newResult?.fingerprint.split(';')[0]
        newReq.deviceGid = [newResult?.fingerprint.split(';')[1]]
      }
      if (newResult?.alarmLevels) {
        newReq.level = newResult?.alarmLevels.map(i => levelMap[i])
      }

      if (newResult?.duration) {
        if (newResult?.duration?.sign === '>') {
          newReq.min_duration = newResult?.duration?.duration
        } else {
          newReq.max_duration = newResult?.duration?.duration
        }
      }

      return newReq;
    }
    return { res, req }
  };
  return result;
})();