module.exports = (function () {
  const result = {
    apiName: '数据查询导出',
    sourcePath: '/cgi/dataQuery/edge/exportHistoryBizGidAttrValuesByTemplate',
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

      const solveFunc = (newResult) => {
        const newCondition = [];
        newCondition.push({
          name: 'point_id',
          value: newResult?.templatePointList || [],
        });
        const { startTime, endTime, interval } = newResult;
        return {
          dataType: 1,
          page: 1,
          size: newResult.limit !== undefined ? newResult.limit : 0,
          conditions: newCondition,
          startTime,
          endTime,
          interval
        };
      };
      return {
        export_type: 'point_data_history',
        file_name: `业务测点${getTime()}.xlsx`,
        param: {
          ...solveFunc(newResult),
        },
      };
    };
    const res = newResult => newResult;
    return { res, req };
  };
  return result;
}());
