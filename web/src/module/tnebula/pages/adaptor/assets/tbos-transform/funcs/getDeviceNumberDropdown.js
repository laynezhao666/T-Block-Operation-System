module.exports = (function () {
  const result = {
    apiName: '（返回缺少字段）',
    sourcePath: '/cgi/alarm/getDeviceNumberDropdown',
    trueTargetPath: '/cgi/idc-tbos-cgi/Common/GetKvDict',
    targetPage: ["all"],
  };
  result.change = function () {
    const res = newResult => newResult.kvs;
    const req = newResult => ({
      dic_type: 'device_gid_kv',
      filter: newResult?.keyword,
    });
    return { res, req };
  };
  return result;
}());
