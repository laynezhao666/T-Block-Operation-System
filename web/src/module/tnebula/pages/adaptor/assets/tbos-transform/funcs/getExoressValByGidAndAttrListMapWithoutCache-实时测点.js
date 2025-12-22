module.exports = (function() {
    // 接口转换基础信息
    const result = {
        "apiName": "获取实时测点",
        "sourcePath": "/cgi/expCompute/edge/getExoressValByGidAndAttrListMapWithoutCache",
        "trueTargetPath": "/cgi/idc-tbos-cgi/Data/Query",
        "targetPage": ["/tedge/data-query-index","/tedge/data-compare-analysis",'/tedge/electric-view','/tedge/electric-gen-view'],
        "targetPath": "/Data/Query"
    };

    // 接口转换具体逻辑
    result.change = function() {
        // 响应转换
        const res = (newResult) => {
            // 初始化响应
            const rspData = {}

            // 字段映射
            const keyMap = {
                "mozuId": "",
                "gid": "device_gid",
                "attr": "point_name_en",
                "expValue": "latest_value",
                "q": "q",
                "valueType": "value_type",
                "unit": "unit",
                // "err": "err",
                // "ts": "ts",  // 需特殊处理
            }
            // 映射逻辑
            newResult?.list?.map((item) => {
                const oldItem = {};
                // 生成普通key映射后的数据
                for (const oldKey in keyMap) {
                    if (Object.prototype.hasOwnProperty.call(item, keyMap[oldKey])) {
                        oldItem[oldKey] = item[keyMap[oldKey]];
                    };
                }
                // 特殊逻辑
                oldItem.ts = new Date(item.update_time.replace(' ', 'T')).getTime() / 1000;
                // 生成到数据中
                if (!rspData[item.device_gid]) {
                    rspData[item.device_gid] = [];
                }
                rspData[item.device_gid].push(oldItem);
            });

            return rspData;
        }

        // 请求转换
        const req = (newResult) => {
            // 初始化
            let newReq = {
                "data_type": 0,
                "cascade": false,
                "conditions": []
            }
            // 补齐模组ID
            if (newResult?.mozuId || newResult?.mozu_id || newResult?.mozuID) {
                newReq.mozu_id = newResult.mozuId || newResult?.mozu_id || newResult?.mozuID;
            }
            // 变量转换
            let pointKeys = [];
            const gidWithAttrListMap = newResult?.gidWithAttrListMap || [];
            gidWithAttrListMap.forEach((item) => {
                item?.gids?.forEach((gid) => {
                    item?.attrs?.forEach((attr) => {
                        pointKeys.push(`${gid}.${attr}`);
                    })
                })
            });
            newReq.conditions.push({
                "name": "point_key",
                "value": pointKeys
            })
            return newReq;
        }

        return {res, req}
    }
    return result
})();
