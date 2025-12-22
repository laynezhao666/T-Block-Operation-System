module.exports = (function () {
  const result = {
    apiName: '采集设备详情',
    sourcePath: '/api/dcos/tboxmonitor-cgi/collector/device/detail',
    trueTargetPath: '/cgi/idc-tbos-cgi/Cmdb/GetCollectorInfo',
    targetPage: ['all'],
  };
  result.change = function () {
    const res = (newResult) => {
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
            "state": newResult?.state,
            devices: newResult?.devices || []
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
