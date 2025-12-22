module.exports = (function () {
  const result = {
    apiName: '历史告警统计(done)',
    sourcePath: '/cgi/alarm/history/getStat',
    trueTargetPath: '/cgi/idc-tbos-cgi/alarm/server/GetAlarmList',
    targetPage: ["all"],
  };
  result.change = function () {
    const res = (newResult) => {
      // 名称映射
      const levelMap = {
        L0: '零级',
        L1: '一级',
        L2: '二级',
        L3: '三级',
        L4: '四级',
        L5: '五级',
      };
      // 键名映射关系
      const keyMap = {
        levelList: 'level',
        topTenByDeviceType: 'device_type_zh',
        topTenByDeviceNumber: 'device_gid',
        topTenByAlarmType: 'alarm_name',
        topTenByAlarmFp: 'fingerprint',
      };
        // 最终结果
      const rspData = {};
      // 新接口数据
      const data = newResult.metrics;
      // 遍历映射关系，处理每个对应的列表数据
      for (const oldKey in keyMap) {
        const newKey = keyMap[oldKey];
        if (!Object.prototype.hasOwnProperty.call(data, newKey)) {
          continue;
        }
        // 告警总数
        const totalCount = data[newKey].list.reduce((sum, item) => sum + parseInt(item.count), 0);
        rspData[oldKey] = data[newKey].list.map((item, index) => ({
          seqId: index + 1,
          name: oldKey === 'levelList' ? levelMap[item.name] : item.name,
          count: parseInt(item.count),
          percent: `${((parseInt(item.count) / totalCount) * 100).toFixed(2)}%`,
        }));
      }
      rspData.total = newResult.total;
      return rspData;
    };
    const req = (newResult) => {
      const newReq = {
        alarm_type: 3, // alarm_type字段
        alarm_level: ['L0', 'L1', 'L2', 'L3', 'L4'],
        count_by_metric: true
      };
      const keyMap = {
        mozuId: 'mozu_id',
        occurTimeStart: 'occur_begin',
        occurTimeEnd: 'occur_end',
        limit: 'size',
      };
        // 遍历判断是否有字段并做映射
      for (const oldKey in keyMap) {
        if (Object.prototype.hasOwnProperty.call(newResult, oldKey)) {
          newReq[keyMap[oldKey]] = newResult[oldKey];
        }
      }
      // page字段特殊处理
      const start = newResult.offset !== undefined ? newResult.offset : 0;
      const limit = newResult.limit !== undefined ? newResult.limit : 10;
      newReq.page = start / limit + 1;

      return newReq;
    };
    return { res, req };
  };
  return result;
}());
