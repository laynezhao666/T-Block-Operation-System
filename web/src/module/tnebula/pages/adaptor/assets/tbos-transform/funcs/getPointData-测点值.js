module.exports = (function () {
  const result = {
    apiName: '测点值',
    sourcePath: '/cgi/alarm/active/getPointData',
    targetPage: ["all"],
  };
  result.change = function () {
    const res = newResult => newResult;
    const req = newResult => newResult;
    return { res, req };
  };
  return result;
}());
