module.exports = (function () {
  const result = {
    apiName: '导出活动告警',
    sourcePath: '/cgi/alarm/active/exportList',
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
        file_name: newResult?.alarmStatus == '2' ? `挂起告警-${getTime()}.xlsx` : `活动告警-${getTime()}.xlsx`,
        param: {
          alarm_type: newResult?.alarmStatus == '2' ? 2 : 1,
          mozu_id: newResult?.mozuId,
          page: 0,
          size: 0,
        },
      };
    };
    const res = newResult => newResult;
    return { res, req };
  };
  return result;
}());
