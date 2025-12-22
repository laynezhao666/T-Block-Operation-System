module.exports = (function () {
  const result = {
    apiName: '获取设备相关实时测点',
    sourcePath: '/cgi/dataQuery/edge/getCurrentBizGidAttrsWithValueByConditions',
    trueTargetPath: '/cgi/idc-tbos-cgi/Data/Query',
    targetPage: ['all'],
  };
  result.change = function () {
    const res = newResult => ({
      count: newResult.total !== undefined ? newResult.total : 0,
      list: newResult.list.map(item => ({
        gid: item.device_gid,
        attrId: item.point_name_en,
        attrName: item.point_name_zh,
        deviceNumber: item.device_number,
        categoryEn: item.device_type_en,
        categoryZh: item.device_type_zh,
        applicationTypeEn: item.application_type_en,
        applicationTypeZh: item.application_type_zh,
        id: item.point_key,
        templatePointId: item.point_id,
        updateTime: item.update_time,
        value: item.latest_value,
        enumValue: item.enum_value,
        unit: item.unit,
        q: item.q,
        signType: item.sign_type,
        status: item.status,
        readAndWrite: item.read_and_write,
        simulation: item.simulation,
      })),
    });
    const req = (newResult) => {
      const keyMap = {
        deviceGid: 'device_gid',
        applicationTypeZh: 'application_type_zh',
      };
      const originalArray = newResult?.conditions || [];
      const newCondition = [];

      const deviceGidObj = originalArray.find(obj => obj.name === 'deviceGid');
      const deviceTypesNameObj = originalArray.find(obj => obj.name === 'deviceTypesName') || originalArray.find(obj => obj.name === 'applicationTypeZh');

      if (deviceGidObj) {
        const combinedValues = deviceGidObj.value.map((gid, index) => `${gid}`);

        newCondition.push({
          name: 'device_gid',
          value: combinedValues,
        });
      }
      return {
        data_type: 0,
        // eslint-disable-next-line no-mixed-operators
        page: newResult.start / newResult.limit + 1,
        size: newResult.limit ? newResult.limit : 0,
        conditions: newCondition,
        keyword: newResult?.keyword || ''
      };
    };
    return { res, req };
  };
  return result;
}());
