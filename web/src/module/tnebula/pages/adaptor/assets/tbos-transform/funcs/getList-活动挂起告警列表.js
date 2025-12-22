module.exports = (function () {
    const result = {
        "apiName": "活动/挂起告警列表done",
        "sourcePath": "/cgi/alarm/active/getList",
        "trueTargetPath": "/cgi/idc-tbos-cgi/alarm/server/GetAlarmList",
        "targetPage": ['all'],
        "targetPath": "/alarm/server/GetAlarmList"
    };

    result.change = function () {
        const res = (newResult) => {
            const keyMap = {
                "alarmId": "alarm_id",
                "occurTime": "occur_time",
                "content": "alarm_content",
                "alarmContent": "alarm_content",
                "alarmType": "alarm_name",
                "level": "level",
                "alarmStatus": 'alarm_status',
                "eventStatus": 'event_status',
                "mozuId": 'mozu_id',
                "mozuName": 'mozu_name',
                "roomName": 'room',
                "boxName": 'box',
                "deviceGid": "device_gid",
                "deviceNumber": "device_number",
                "deviceType": "device_type_zh",
                "hangupReason": "hangup_reason",
            };
            const rspData = {
                count: newResult?.total !== undefined ? newResult.total : 0,
                list: [],
            };
            rspData.list = newResult?.list?.map((item) => {
                const oldItem = {};
                for (const oldKey in keyMap) {
                    if (Object.prototype.hasOwnProperty.call(item, keyMap[oldKey])) {
                        oldItem[oldKey] = item[keyMap[oldKey]];
                    }
                }
                // occurPointList 需要特殊处理
                oldItem.occurPointList = [
                    {
                        zhName: item.alarm_content,
                        value: "",
                        unit: ""
                    }
                ]

                // position 字段需要特殊处理
                oldItem.position = item[keyMap["roomName"]] + "/" + item[keyMap["boxName"]];
                // 返回数据
                return oldItem;
            });
            return rspData;
        };
        const req = (newResult) => {
            let newReq = {
                "alarm_type": 1, // alarm_type字段
                "page": 1
            };
            const keyMap = {
                "alarmStatus": "alarm_type",
                "eventStatus": "event_status",
                "limit": "size",
                "level": "level",
                "mozuId": "mozu_id",
                "sortedType": "sort_type",
                "alarmType": "alarm_name"
            };
            // 遍历判断是否有字段
            for (const oldKey in keyMap) {
                if (Object.prototype.hasOwnProperty.call(newResult, oldKey)) {
                    newReq[keyMap[oldKey]] = newResult[oldKey];
                }
            }
            //alarm_id
            if (newResult?.AlarmId) {
                newReq.alarm_id = typeof newResult?.AlarmId === 'int' ? newResult?.AlarmId : newResult?.AlarmId[0]
            }
            // page 字段特殊处理
            if (newResult?.start !== undefined && newResult?.limit !== undefined && newResult?.limit !== 0) {
                newReq.page = newResult.start / newResult.limit + 1;
            }
            // sort_type 字段特殊处理
            if (newResult?.sortedType !== undefined) {
                newReq.sort_type = newResult.sortedType == 0 ? 1 : 2
            }
            // content字段特殊处理-新模组
            if (newResult.Content && newResult.Content.length > 0) {
                newReq.content = newResult.Content[0];
            }
            // content字段特殊处理-旧模组
            if (newResult.DeviceNumber && newResult.DeviceNumber.length > 0) {
                newReq.device_number = newResult.DeviceNumber;
            }
            // content字段特殊处理-旧模组
            if (newResult?.deviceNumber && newResult?.deviceNumber.length > 0) {
                newReq.device_number = newResult.deviceNumber;
            }
            return newReq;
        }
        return { res, req }
    };

    return result;
})();