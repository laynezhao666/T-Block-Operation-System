module.exports = (function () {
  const result = {
    apiName: '历史告警导出',
    sourcePath: '/cgi/alarm/history/exportList',
    trueTargetPath: '/cgi/idc-tbos-cgi/Common/ExportData',
    targetPage: ["all"],
  };
  result.change = function () {
    const req = (newResult) => {
      const getTime = () => {
        // 创建表示当前时间的Date对象
        const date = new Date();
        // 获取年
        const year = date.getFullYear().toString();
        // 获取月（注意要加1，因为月份从0开始计数），并格式化为两位数字
        const month = (`0${date.getMonth() + 1}`).slice(-2).toString();
        // 获取日，并格式化为两位数字
        const day = (`0${date.getDate()}`).slice(-2).toString();
        // 获取小时，并格式化为两位数字
        const hour = (`0${date.getHours()}`).slice(-2).toString();
        // 获取分钟，并格式化为两位数字
        const minute = (`0${date.getMinutes()}`).slice(-2).toString();
        // 获取秒，并格式化为两位数字
        const second = (`0${date.getSeconds()}`).slice(-2).toString();

        // 拼接成不带符号的时间字符串
        const timeStr = year + month + day + hour + minute + second;
        return timeStr;
      };

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
        newReq.alarm_id = newResult?.fingerprint.split(';')[1]
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

      return {
        export_type: 'alarm_list',
        fields: [
          {
            field_cn: '模组名称',
            field_en: 'mozu_name',
          },
          {
            field_cn: '告警Id',
            field_en: 'alarm_id',
          },
          {
            field_cn: '告警名称',
            field_en: 'alarm_name',
          },
          {
            field_cn: '告警级别',
            field_en: 'level',
          },
          {
            field_cn: '告警内容',
            field_en: 'alarm_content',
          },
          {
            field_cn: '告警设备',
            field_en: 'device_number',
          },
          {
            field_cn: '设备类型',
            field_en: 'device_type_zh',
          },
          {
            field_cn: '方仓名',
            field_en: 'box',
          },
          {
            field_cn: '房间名',
            field_en: 'room',
          },
          {
            field_cn: '测点列表',
            field_en: 'points',
          },
          {
            field_cn: '产生时间',
            field_en: 'occur_time',
          },
          {
            field_cn: '挂起原因',
            field_en: 'hangup_reason',
          },
          {
            field_cn: '恢复时间',
            field_en: 'restore_time',
          },
        ],
        file_name: `历史告警-${getTime()}.xlsx`,
        param: {
          ...newReq
        },
      };
    };
    const res = newResult => newResult;
    return { res, req };
  };
  return result;
}());
