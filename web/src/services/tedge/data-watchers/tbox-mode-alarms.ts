import { RequestConfig } from "services/polling-request-proxy/request-config";
import { TedgeRemoteWatcher } from "../remote-data-watcher";
import * as _ from 'lodash';
import dayjs = require('dayjs');

export interface TboxAlarm {
  uuid:                string;
  collect_point_id:    string;
  collect_point_name:  string;
  collect_device_id:   string;
  collect_device_name: string;
  point_id:            string;
  point_name:          string;
  device_id:           string;
  device_name:         string;
  desc:                string;
  state:               string;
  type:                string;
  value:               string;
  trigger_time:        string;
  level:               number;
  begin_time:          number;
  update_time:         number;
  trigger_expression:  string;
  current_state:       string;
  resume_expression:   string;
  is_valid:            number;
  is_confirm:          number;
  confirm_time:        number;
  device_type:         string;
  query_types:         any[];
};

export interface TedgeAlarmsWatchParams {
  is_active?: 1,
};

/**
 * 告警监听器
 */
export class TboxModeAlarmsWatcher extends TedgeRemoteWatcher<any, TedgeAlarmsWatchParams> {
  resolveRequestConfig(params: TedgeAlarmsWatchParams): RequestConfig {
    return {
      url: '/cgi/alarm/list',
      method: 'POST',
      data: {
        ...params,
        is_active: _.isNil(params.is_active) ? 1 : params.is_active,
        limit: 9999999,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data.list.map(tboxAlarm => {
      const resultAlarm: any = {
        id: 0,
        alarmId: Math.random().toString(32).substring(2),
        alarm_id_string: Math.random().toString(32).substring(2),
        occurTime: dayjs(tboxAlarm.trigger_time).format('YYYY-MM-DD HH:mm:ss'),
        content: tboxAlarm.desc,
        alarmType: tboxAlarm.type,
        level: `L${tboxAlarm.level}`,
        fingerprint: '',
        mozuInfrastructureCompany: '',
        occurPointList: [],
        alarmStatus: 0,
        eventStatus: 0,
        hangupReason: '',
        hangupUpdateTime: '',
        hangupUserID: '',
        hangupUserName: '',
        mozuId: 0,
        mozuName: '',
        roomName: '',
        boxName: '',
        deviceGid: tboxAlarm.device_id,
        deviceNumber: tboxAlarm.device_id,
        deviceType: tboxAlarm.device_type,
        position: ''
      };

      return resultAlarm;
    });;
  }
}

/**
 * 告警监听器
 */
export class TboxModeAlarmsCountWatcher extends TedgeRemoteWatcher<number, { device_id: string[] }> {
  resolveRequestConfig(params: { device_id: string[] }): RequestConfig {
    return {
      url: '/cgi/alarm/count',
      method: 'POST',
      data: {
        device_id: params.device_id,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data;
  }
}
