module.exports = (function () {
  const result = {
    apiName: '（后台未实现）',
    sourcePath: '/cgi/alarm/validate/realtime/getDropdown',
    targetPage: ["all"],
  };
  result.change = function () {
    const res = newResult => newResult;
    const req = newResult => newResult;
    return { res, req };
  };
  return result;
}());
