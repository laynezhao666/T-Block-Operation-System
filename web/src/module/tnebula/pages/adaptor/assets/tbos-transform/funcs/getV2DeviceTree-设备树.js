module.exports = (function () {
  const result = {
    apiName: '设备树(done)',
    sourcePath: '/cgi/gidmapping/getV2DeviceTree',
    trueTargetPath: '/cgi/idc-tbos-cgi/Cmdb/GetDeviceTree',
    targetPage: ["all"],
    targetPath: '/Cmdb/GetDeviceTree',
  };
  result.change = function () {
    const res = (newResult) => {
      // 字段映射
      const keyMap = {
        no: 'device_gid',
        id: 'device_gid',
        name: 'device_number',
        deviceTypeEn: 'old_device_type_en',
        deviceTypeName: 'old_device_type_zh',
        deviceCategoryEn: 'device_type_en',
        deviceCategoryName: 'device_type_zh',
        applicationTypeEn: 'application_type_en',
        applicationTypeZh: 'application_type_zh',
        deviceCount: 'device_count',
      };
      // 定义递归处理函数
      function transformTree(node) {
        // 处理当前节点
        const transformedNode = {};
        for (const oldKey in keyMap) {
          if (Object.prototype.hasOwnProperty.call(node, keyMap[oldKey])) {
            if (oldKey === 'deviceTypeEn') {
              transformedNode[oldKey] = (node[keyMap[oldKey]] || node['device_type_en']) === 'MG' ? 'mozu' : (node[keyMap[oldKey]] || node['device_type_en'])
            } else if (oldKey === 'deviceTypeName') {
              transformedNode[oldKey] = (node[keyMap[oldKey]] || node['device_type_zh']) === 'TB模组' ? '模组' : (node[keyMap[oldKey]] || node['device_type_zh']);
            } else {
              transformedNode[oldKey] = node[keyMap[oldKey]];
            }
          }
        }
        // 递归处理子节点
        if (node.children && node.children.length > 0) {
          transformedNode.children = node.children.map(item => transformTree(item));
        }

        return transformedNode;
      }
      // 处理异常情况
      if (newResult?.list && newResult.list.length === 0) {
        return [];
      }
      // 处理结果
      const rspData = newResult.list.map(item => transformTree(item));
      return rspData;
    };
    const req = (newResult) => {
      // 字段应色号
      const keyMap = {
        mozuId: 'mozu_id',
      };
      const reqData = {};
      for (const oldKey in keyMap) {
        if (Object.prototype.hasOwnProperty.call(newResult, keyMap[oldKey])) {
          reqData[oldKey] = newResult[keyMap[oldKey]];
        }
      }
      // tree_type 兼容性处理
      if (newResult?.tree_type !== undefined) {
        reqData.tree_type = newResult.tree_type;
      }
      return reqData;
    };
    return { res, req };
  };
  return result;
}());
