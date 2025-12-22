module.exports = (function () {
  const result = {
    apiName: '（请求无法转换，需要判断不同版本设置不同传参）',
    sourcePath: '/cgi/dataQuery/edge/getDistinctByFieldNameByConditions',
    trueTargetPath: '/cgi/idc-tbos-cgi/Cmdb/GetSubTreeFieldDic',
    targetPage: ["all"],
    targetPath: '/Cmdb/GetApplicationTypeDict',
  };
  result.change = function () {
    const res = newResult => newResult.list.filter(item => item !== '');
    // 请求不用转，这里伪造假的，通过校验
    const req = (newResult) => {
      const device_gid = newResult.conditions.find(i => i.name === 'deviceGid')?.value;
      return {
        device_gid: device_gid?.length ? device_gid[0] : '',
        filed_type: 'application_type_zh',
      };
    };
    return { res, req };
  };
  return result;
}());
