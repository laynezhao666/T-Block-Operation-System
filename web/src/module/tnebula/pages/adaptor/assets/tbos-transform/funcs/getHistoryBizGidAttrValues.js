module.exports = (function () {
  const result = {
    apiName: '测点历史数据查询getHistoryBizGidAttrValues',
    sourcePath: '/cgi/dataQuery/edge/getHistoryBizGidAttrValues',
    trueTargetPath: '/cgi/idc-tbos-cgi/Data/Query',
    targetPage: ["all"],
  };
  result.change = function () {
    const res = (newResult) => {
      const result = {}
      newResult.list.forEach(item => {
        const kvMap = {}
        item.point_data.forEach(i => {
          const timestamp = Math.floor(new Date(i.update_time).getTime() / 1000)
          kvMap[timestamp] = i.value;
        })
        result[item.point_key] = kvMap
      });
      return result;
    }

    const req = (newResult) => {
      const newReq = {
        data_type: 1, // 历史告警类型
        conditions: [
          {
            name: 'point_key',
            value: newResult?.gidWithAttrList || [],
          },
        ],
        interval: newResult?.interval || 60,
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
