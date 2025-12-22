module.exports = (function () {
  const result = {
    apiName: '数据查询导出',
    sourcePath: '/cgi/dataQuery/edge/exportCurrentBizGidAttrsWithValueByConditions',
    trueTargetPath: '/cgi/idc-tbos-cgi/Common/ExportData',
    targetPage: ['all'],
  };
  result.change = function () {
    const req = (newResult) => {
      const getTime = () => {
        // 创建表示当前时间的Date对象
        const date = new Date();
        // 获取年
        const year = date.getFullYear().toString();
        // 获取月（注意要加1，因为月份从0开始计数），并格式化为两位数字
        const month = (`0${date.getMonth() + 1}`).slice(-2).toString();
        // 获取日，并格式化为两位数字
        const day = (`0${date.getDate()}`).slice(-2).toString();
        // 获取小时，并格式化为两位数字
        const hour = (`0${date.getHours()}`).slice(-2).toString();
        // 获取分钟，并格式化为两位数字
        const minute = (`0${date.getMinutes()}`).slice(-2).toString();
        // 获取秒，并格式化为两位数字
        const second = (`0${date.getSeconds()}`).slice(-2).toString();

        // 拼接成不带符号的时间字符串
        const timeStr = year + month + day + hour + minute + second;
        return timeStr;
      };

      const solveFunc = (newResult) => {
        const keyMap = {
          deviceGid: 'device_gid',
          applicationTypeZh: 'application_type_zh',
        };
        const originalArray = newResult?.conditions || [];
        const newCondition = [];

        const deviceGidObj = originalArray.find(obj => obj.name === 'deviceGid');
        const deviceTypesNameObj = originalArray.find(obj => obj.name === 'deviceTypesName') 
        const applicationTypeZhObj =originalArray.find(obj => obj.name === 'applicationTypeZh');

        if(deviceTypesNameObj) {
          newCondition.push({
            name: 'device_type_zh',
            value: deviceTypesNameObj.value,
          });
        }

        if(applicationTypeZhObj) {
          newCondition.push({
            name: 'application_type_zh',
            value: applicationTypeZhObj.value,
          });
        }
        if (deviceGidObj) {
          const combinedValues = deviceGidObj.value.map((gid, index) => `${gid}`);

          newCondition.push({
            name: 'device_gid',
            value: combinedValues,
          });
        }
        return {
          data_type: 0,
          page: 1,
          size: newResult.limit !== undefined ? newResult.limit : 0,
          conditions: newCondition,
        };
      };
      return {
        export_type: 'point_data',
        fields: [
          {
            field_cn: '设备GID',
            field_en: 'device_gid',
          },
          {
            field_cn: '设备编号',
            field_en: 'device_number',
          },
          {
            field_cn: '设备种类',
            field_en: 'device_type_zh',
          },
          {
            field_cn: '应用类型',
            field_en: 'application_type_zh',
          },
          {
            field_cn: '测点标识符',
            field_en: 'point_name_en',
          },
          {
            field_cn: '测点名称',
            field_en: 'point_name_zh',
          },
          {
            field_cn: '更新时间',
            field_en: 'update_time',
          },
          {
            field_cn: '测点值',
            field_en: 'latest_value',
          },
          {
            field_cn: '单位',
            field_en: 'unit',
          },
        ],
        file_name: `业务测点${getTime()}.xlsx`,
        param: {
          ...solveFunc(newResult),
        },
      };
    };
    const res = newResult => newResult;
    return { res, req };
  };
  return result;
}());
