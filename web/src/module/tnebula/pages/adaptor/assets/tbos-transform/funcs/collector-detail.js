module.exports = (function () {
  const result = {
    apiName: '采集器设备详情',
    sourcePath: '/api/dcos/tboxmonitor-cgi/collector/detail',
    trueTargetPath: '/cgi/idc-tbos-cgi/Cmdb/GetCollectorInfo',
    targetPage: ['all'],
  };
  result.change = function () {
    const res = (newResult) => {
      function addPrefix(obj, prefix) {
        const result = {};
        for (const key in obj) {
          if (obj.hasOwnProperty(key)) {
            result[key] = obj[key].map(item => prefix + item);
          }
        }
        return result;
      }
      
      const state_id = addPrefix(newResult?.state, newResult?.device_gid);
      const rspData = {
            "desc": "",
            "extend": {
                "hardware_version": "",
                "kernel_version": "",
                "software_version": ""
            },
            "id": newResult?.device_gid,
            "link_channel": {
              ...newResult.channel_link
            },
            "name": newResult?.device_name,
            "position": {
                "mark": "",
                "room": "",
                "block": "",
                "no": "",
                "desc": ""
            },
            "profile": {
                "labels": null,
                "model": "",
                "sn": "",
                "thing_template": newResult?.template_info,
                "vendor": ""
            },
            "state_id": state_id,
            devices: newResult?.devices.map(i=>{
              return {
                "comm_state_id": i?.status_id,
                "desc": "",
                "id": i?.device_gid,
                "link_channel": i?.channel_link,
                "name": i?.device_name,
                "position": {
                  "mark": "",
                  "room": "",
                  "block": "",
                  "no": "",
                  "desc": ""
                }
              }
            }) || []
        };
      return rspData;
    };
    const req = newResult => ({
      device_gid: newResult?.id,
    });
    return { res, req };
  };
  return result;
}());
