module.exports = (function () {
  const result = {
    apiName: '告警策略列表',
    sourcePath: '/cgi/alarm/rule/getList',
    trueTargetPath: '/cgi/idc-tbos-cgi/alarm/server/GetStrategy',
    targetPage: ['/tedge/warning-strategy', 'all'],
  };
  result.change = function () {
    const res = (newResult) => {
      const keyMap = {
        protocolType: 'device_type',
        protocolTypeName: 'device_type',
        deviceGidList: 'gid_list',
        deviceNumberList: 'device_list',
        deviceCount: 0,
        alarmType: 'alarm_name',
        alarmLevel: 'level',
        alarmContent: 'content',
        isStandard: 'standard',
        alarmExpressionStr: 'alarm_exp',
        occurExpression: 'alarm_exp',
        restoreExpressionStr: 'restore_exp',
        restoreExpression: 'restore_exp',
        createdByName: 'owner',
        createdAt: 'create_at',
        updatedByName: 'owner',
        updatedAt: 'update_at',
        contentTemplate: 'content',
        applicationTypeZh: 'apply_type',
        applicationTypeEn: 'apply_type',
        deviceCategoryZh: 'device_type',
        deviceCategoryEn: 'device_type',
      };
      const rspData = {
        count: newResult?.total,
        list: [],
      };
      rspData.list = newResult?.list?.map((newItem) => {
        const oldItem = {};
        for (const oldKey in keyMap) {
          if (oldKey === 'deviceCount') oldItem[oldKey] = newItem.gid_list.length;
          if (Object.prototype.hasOwnProperty.call(newItem, keyMap[oldKey])) {
            oldItem[oldKey] = newItem[keyMap[oldKey]];
          }
        }
        return oldItem;
      });
      return rspData;
    };
    const req = (newResult) => {
      // 告警等级映射
      const alarmLevelMap = {
        零级: 'L0',
        一级: 'L1',
        二级: 'L2',
        三级: 'L3',
        四级: 'L4',
        五级: 'L5',
      };
      // 模组ID
      const mozu_id = newResult?.mozuId ? newResult.mozuId : 0;
      // 分页数据
      const start = newResult.offset !== undefined ? newResult.offset : 0;
      const limit = newResult?.limit !== undefined ? newResult.limit : 10;
      // 结果数据
      const newReq = {
        mozu_id,
        page: start / limit + 1,
        size: limit,
        level: [
          'L0', 'L1', 'L2', 'L3', 'L4',
        ],
      };
      // 应用类型
      if (newResult.applicationTypeEn !== undefined) {
        newReq.apply_type = [newResult.applicationTypeEn];
      }
      // 设备种类
      if (newResult.deviceCategoryEn !== undefined || newResult.protocolType !== undefined) {
        newReq.device_type = [newResult.deviceCategoryEn !== undefined ? newResult.deviceCategoryEn : newResult.protocolType];
      }
      // 告警类型
      if (newResult.alarmType !== undefined) {
        newReq.alarm_name = [newResult.alarmType];
      }
      // 告警等级
      if (newResult.alarmLevel !== undefined) {
        newReq.level = [alarmLevelMap[newResult.alarmLevel]];
      }

      return newReq;
    };
    return { res, req };
  };
  return result;
}());
