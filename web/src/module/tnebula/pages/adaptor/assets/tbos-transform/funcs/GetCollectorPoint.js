module.exports = (function () {
  const result = {
    apiName: '采集测点列表',
    sourcePath: '/api/dcos/tboxmonitor-cgi/collector/points',
    trueTargetPath: '/cgi/idc-tbos-cgi/Cmdb/GetCollectorPoint',
    targetPage: ['all'],
  };
  result.change = function () {
    const res = (newResult) => {
      const rspData = newResult?.points.map(i => ({
        access: i?.point_rw,
        almdef: {},
        deltadef: i?.delta_def,
        id: i?.point_key,
        is_standard: i?.point_standard,
        name: i?.point_name_zh,
        no: i?.point_name_en,
        protdef: i?.prot_def,
        simulator: i?.simulator,
        valdef: i?.val_def,
        valtype: i?.point_type,
      }));
      return rspData;
    };
    const req = newResult => ({
      device_gid: newResult?.id,
    });
    return { res, req };
  };
  return result;
}());
