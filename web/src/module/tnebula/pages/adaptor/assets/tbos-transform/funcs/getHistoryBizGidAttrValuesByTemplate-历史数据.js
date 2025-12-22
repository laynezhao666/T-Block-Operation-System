module.exports = (function () {
  const result = {
    apiName: '历史数据（done）',
    sourcePath: '/cgi/dataQuery/edge/getHistoryBizGidAttrValuesByTemplate',
    trueTargetPath: '/cgi/idc-tbos-cgi/Data/Query',
    targetPage: ["all"],
  };
  result.change = function () {
    const res = newResult => newResult.list.map(item => ({
      id: item.point_id,
      deviceNumber: item.device_number,
      attrName: item.attr_name,
      unit: item.unit,
      data: item.point_data.map(dataItem => ({
        updateTime: dataItem.update_time,
        value: dataItem.value,
      })),
      stats: item.stats !== undefined ? item.stats : [],
    }));
    const req = (newResult) => {
      const newReq = {
        data_type: 1, // 历史告警类型
        conditions: [
          {
            name: 'point_id',
            value: newResult.templatePointList !== undefined ? newResult.templatePointList : [],
          },
        ],
        interval: newResult.interval !== undefined ? newResult.interval : 60,
        start_time: newResult.startTime,
        end_time: newResult.endTime,
        stats: newResult.stats !== undefined ? newResult.stats : ['avg', 'min', 'max'],
        keyword: newResult?.keyword || ''
      };
      // 组装设备编号测点
      return newReq;
    };
    return { res, req };
  };
  return result;
}());
