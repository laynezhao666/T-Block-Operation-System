module.exports = (function () {
  const result = {
    apiName: '（导出数据，后续统一测试）',
    sourcePath: '/cgi/dataQuery/edge/exportGidAndAttrListValueMapWithoutCache',
    trueTargetPath: '/cgi/idc-tbos-cgi/Common/ExportData',
    targetPage: ['all'],
  };
  result.change = function () {
    const res = newResult => newResult;
    const req = newResult => newResult;
    return { res, req };
  };
  return result;
}());
