module.exports = (function () {
  const result = {
    apiName: '告警策略：①应用类型下拉 ②设备种类下拉',
    sourcePath: '/cgi/gidmapping/getDeviceTypeDicInfo',
    trueTargetPath: '/cgi/idc-tbos-cgi/Common/GetKvDict',
    targetPage: ['all'],
  };
  result.change = function () {
    const res = newResult => ({
      dataList: Object.keys(newResult.kvs).map(key => ({
        typeEn: newResult.kvs[key],
        typeZh: newResult.kvs[key],
      })),
      count: newResult?.count,
    });
    const req = (newResult) => {
      switch (newResult.typeCode) {
        // 原请求 typeCode 为 1 说明是设备类型
        case 1:
          return {
            dic_type: 'device_type_kv',
          };
          // 原请求 typeCode 为 2 说明是应用类型
        case 2:
          return {
            dic_type: 'application_type_kv',
          };
          // 默认设备类型
        default:
          return {
            dic_type: 'device_type_kv',
          };
      }
    };
    return { res, req };
  };
  return result;
}());
