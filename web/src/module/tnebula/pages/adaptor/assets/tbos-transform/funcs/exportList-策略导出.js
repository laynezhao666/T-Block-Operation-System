module.exports = (function () {
  const result = {
    apiName: '策略导出',
    sourcePath: '/cgi/alarm/rule/exportList',
    trueTargetPath: '/cgi/idc-tbos-cgi/Common/ExportData',
    targetPage: ['all'],
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
        export_type: 'alarm_strategy',
        fields: [
          {
              "field_cn": "策略Rid",
              "field_en": "rid"
          },
          {
              "field_cn": "告警名称",
              "field_en": "alarm_name"
          },
          {
              "field_cn": "告警级别",
              "field_en": "level"
          },
          {
              "field_cn": "告警表达式",
              "field_en": "alarm_exp"
          },
          {
              "field_cn": "恢复表达式",
              "field_en": "restore_exp"
          },
          {
              "field_cn": "应用类型",
              "field_en": "apply_type"
          },
          {
              "field_cn": "设备类型",
              "field_en": "device_type"
          },
          {
              "field_cn": "告警内容",
              "field_en": "content"
          },
          {
              "field_cn": "设备列表",
              "field_en": "device_list"
          },
          {
              "field_cn": "创建时间",
              "field_en": "create_at"
          },
          {
              "field_cn": "更新时间",
              "field_en": "update_at"
          },
          {
              "field_cn": "责任人",
              "field_en": "owner"
          }
        ],
        file_name: `告警策略-${getTime()}.xlsx`,
        param: {
          alarm_type: 1,
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
