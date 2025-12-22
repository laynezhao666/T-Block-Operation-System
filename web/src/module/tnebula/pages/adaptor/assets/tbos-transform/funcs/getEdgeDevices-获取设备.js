module.exports = (function() {
    // 接口转换基础信息
    const result = {
        "apiName": "获取设备信息",
        "sourcePath": "/cgi/dataQuery/edge/getEdgeDevices",
        "trueTargetPath": "/cgi/idc-tbos-cgi/Cmdb/GetDeviceEntity",
        "targetPage": ["all"],
        "targetPath": "/Cmdb/GetDeviceEntity"
    };

    // 接口转换具体逻辑
    result.change = function() {
        // 响应转换
        const res = (newResult) => {
            // 初始化响应
            const rspData = {
                count: newResult?.total || 0,
                list: [],
            };
            // 字段映射
            const keyMap = {
                "applicationTypeEn": "application_type_en",
                "applicationTypeZh": "application_type_zh",
                "device_number": "device_number",
                "deviceNumberRoute": "device_number_route",
                "device_belongDeviceNumber": "parent_device_number",
                "device_uid": "device_gid",
                "mozu_id": "mozu_id",
                "mozu_name": "mozu_name",
                "room_name": "func_room",
                // TODO 未知字段
                "room_id": "room_id",
                "categoryEn": "device_type_en",
                "categoryZh": "device_type_zh",
                "devicetypes_enAbbreviation": "device_type_en",
                "devicetypes_name": "device_type_zh"
            };
            // 映射逻辑
            rspData.list = newResult?.list?.map((item)=> {
                const oldItem = {};
                for (const oldKey in keyMap) {
                    if (Object.prototype.hasOwnProperty.call(item, keyMap[oldKey])) {
                        oldItem[oldKey] = item[keyMap[oldKey]];
                    };
                }
                return oldItem;
            });

            return rspData;
        };

        // 请求转换
        const req = (newResult) => {
            let newReq = {
                "device_number": []
            };
            // 补齐模组ID
            if (newResult?.mozuId || newResult?.mozu_id) {
                newReq.mozu_id = newResult.mozuId || newResult?.mozu_id;
            }
            // 变量转换
            const conditions = newResult?.conditions || [];
            conditions.forEach((item) => {
                newReq[item.name] = item.value;
            });
            return newReq;
        };

        return {res, req}
    }

    // 返回结果
    return result;
})();