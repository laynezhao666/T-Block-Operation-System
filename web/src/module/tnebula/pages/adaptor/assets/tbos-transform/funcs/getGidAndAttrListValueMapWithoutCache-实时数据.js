module.exports = (function () {
  const result = {
    apiName: '实时数据',
    sourcePath: '/cgi/dataQuery/edge/getGidAndAttrListValueMapWithoutCache',
    trueTargetPath: '/cgi/idc-tbos-cgi/Data/Query',
    targetPage: ['all'],
    targetPath: '/Data/Query',
  };
  result.change = function () {
    const res = newResult => ({
      hasTransformed: true,
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
      if (newResult?.conditions) {
        return newResult;
      }
      const pointList = [];
      const data = newResult?.gidWithAttrListMap || [];
      data.forEach((obj) => {
        const gids = obj.gids || [];
        const attrs = obj.attrs || [];
        gids.forEach((gid) => {
          attrs.forEach((attr) => {
            const combined = `${gid}.${attr}`;
            pointList.push(combined);
          });
        });
      });
      const newReq = {
        page: 1,
        size: newResult.limit !== undefined ? newResult.limit : 99999,
        conditions: [{
          name: 'point_key',
          value: pointList,
        }],
        keyword: newResult?.keyword || ''
      };
      return newReq;
    };
    return { res, req };
  };
  return result;
}());
