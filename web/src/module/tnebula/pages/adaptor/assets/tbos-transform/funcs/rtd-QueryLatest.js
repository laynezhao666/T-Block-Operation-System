module.exports = (function () {
  const result = {
    apiName: '采集测点数据详情',
    sourcePath: '/api/dcos/tboxmonitor-cgi/collector/points/rtd',
    trueTargetPath: '/cgi/idc-tbos-cgi/Data/QueryLatest',
    targetPage: ['all'],
  };
  result.change = function () {
    const res = (newResult) => {
      const rspData = {};
      const datamaps = newResult?.maps || {};
      Object.keys(datamaps).forEach((key) => {
        const item = datamaps[key];
        if (typeof item === 'object' && item !== null) {
          rspData[key] = {
            id: key,
            pv: `${item?.val}`,
            tms: item?.ts,
            des: '',
            qua: `${item?.quality}`,
          };
        }
      });
      return rspData;
    };
    const req = newResult => ({
      point_keys: newResult?.ids,
    });
    return { res, req };
  };
  return result;
}());
