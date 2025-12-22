module.exports = (function () {
  const result = {
    apiName: '非AI模组的树',
    sourcePath: '/cgi/dataQuery/edge/getBizDeviceLevelTree',
    trueTargetPath: '/cgi/idc-tbos-cgi/Cmdb/GetDeviceTree',
    targetPage: ["all"],
  };
  result.change = function () {
    const res = (newResult) => {
      // 字段映射
      const keyMap = {
        no: 'device_gid',
        id: 'device_gid',
        name: 'device_number',
        deviceTypeName: 'device_type_zh',
        deviceCount: 'device_count',
      };
      // 定义递归处理函数
      function transformTree(node) {
        // 处理当前节点
        const transformedNode = {};
        for (const oldKey in keyMap) {
          if (Object.prototype.hasOwnProperty.call(node, keyMap[oldKey])) {
            if (node[keyMap[oldKey]] === 'TB模组') {
              transformedNode[oldKey] = '模组';
            } else if (node[keyMap[oldKey]] === '旧房间') {
              transformedNode[oldKey] = '房间';
            } else {
              transformedNode[oldKey] = node[keyMap[oldKey]];
            }
          }
        }
        if ((transformedNode?.id || '').includes('room')) {
          transformedNode.deviceTypeName = '房间'
          transformedNode.name = (transformedNode?.id || ':').split(':')[1] || transformedNode.name
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
    const req = newResult => newResult;
    return { res, req };
  };
  return result;
}());
