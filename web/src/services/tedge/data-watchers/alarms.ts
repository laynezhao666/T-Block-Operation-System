import { RequestConfig, RequestProxyConfig } from "services/polling-request-proxy/request-config";
import { TedgeRemoteWatcher } from "../remote-data-watcher";

export interface TedgeAlarmsWatchParams {
  deviceNumbers?: string[];
  /** 告警状态，1-活动告警，默认为1 */
  eventStatus?: 1;
  limit?: number;
};

export interface AlarmData {
  id: number;
  alarmId: string;
  alarm_id_string: string;
  occurTime: string;
  content: string;
  alarmType: string;
  level: string;
  fingerprint: string;
  mozuInfrastructureCompany: string;
  occurPointList: AlarmOccurPointList[];
  alarmStatus: number;
  eventStatus: number;
  hangupReason: string;
  hangupUpdateTime: string;
  hangupUserID: string;
  hangupUserName: string;
  mozuId: number;
  mozuName: string;
  roomName: string;
  boxName: string;
  deviceGid: string;
  deviceNumber: string;
  deviceType: string;
  position: string;
}

interface AlarmOccurPointList {
  zhName: string;
  enName: string;
  rw: string;
  unitType: string;
  unit: string;
  dataType: string;
  dataRange: string;
  dataDefine: string;
  desc: string;
  createTime: string;
  scale: string;
  point: string;
  value: number;
  isBusinessPoint: boolean;
  isVirtualPoint: boolean;
  gid: number;
}


/**
 * 告警列表监听器
 */
export class AlarmsWatcher extends TedgeRemoteWatcher<AlarmData[], TedgeAlarmsWatchParams> {
  resolveRequestConfig(params: TedgeAlarmsWatchParams): RequestConfig {
    return {
      url: '/cgi/alarm/active/getList',
      method: 'POST',
      data: {
        DeviceNumber: params.deviceNumbers,
        eventStatus: params.eventStatus || 1,
        ...params,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data.list;
  }
}

/**
 * 告警列表监听器，返回值包含总数
 */
export class AlarmsWithTotalWatcher extends AlarmsWatcher {
  resolveResponseData(data: any) {
    return data.data;
  }
}

/**
 * 告警总数监听器
 */
export class AlarmsCountWatcher extends TedgeRemoteWatcher<number, TedgeAlarmsWatchParams> {
  resolveRequestConfig(params: TedgeAlarmsWatchParams): RequestConfig {
    return {
      url: '/cgi/alarm/active/getList',
      method: 'POST',
      data: {
        DeviceNumber: params.deviceNumbers,
        eventStatus: params.eventStatus || 1,
        ...params,
        limit: 1,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data.count;
  }
}

/**
 * 告警按设备统计监听器
 */
export class AlarmsCountByDeviceWatcher extends TedgeRemoteWatcher<number, TedgeAlarmsWatchParams> {
  plugins?: RequestProxyConfig['plugins'] = [{
    id: 'countBy',
    config: {
      pickField: 'data.list',
      countByField: 'deviceGid',
    },
  }];

  resolveRequestConfig(params: TedgeAlarmsWatchParams): RequestConfig {
    return {
      url: '/cgi/alarm/active/getList',
      method: 'POST',
      data: {
        DeviceNumber: params.deviceNumbers,
        eventStatus: params.eventStatus || 1,
        ...params,
      },
    };
  }

  resolveResponseData(data: any) {
    return data;
  }
}

export interface TedgeHistoryAlarmsWatchParams {
  deviceGids?: string[];
  mozuId?: number;
  alarmTypes?: string[];

  limit?: number;
  offset?: number;

  occurTimeStart?: string;
  occurTimeEnd?: string;
}

/**
 * 历史告警列表监听器
 */
export class HistoryAlarmsWatcher extends TedgeRemoteWatcher<AlarmData[], TedgeAlarmsWatchParams> {
  resolveRequestConfig(params: TedgeHistoryAlarmsWatchParams): RequestConfig {
    return {
      url: '/cgi/alarm/history/getList',
      method: 'POST',
      data: {
        ...params,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data;
  }
}

export interface NewAlarmsTrendWatcherParams {
  mozuId?: number | string;
  occurTimeStart?: string;
  occurTimeEnd?: string;
}

/**
 * 新产生活动告警趋势
 */
export class NewAlarmsTrendWatcher extends TedgeRemoteWatcher<AlarmData[], NewAlarmsTrendWatcherParams> {
  resolveRequestConfig(params: NewAlarmsTrendWatcherParams): RequestConfig {
    return {
      url: '/cgi/alarm/ba/getHoursAlarmTrend',
      method: 'POST',
      data: {
        ...params,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data;
  }
}

export interface AlarmsCountByLevelWatcherParams {
  mozuId?: number | string;
  eventStatus?: -1 | 1 | 2;
}

/**
 * 告警按等级统计数量
 */
export class AlarmsCountByLevelWatcher extends TedgeRemoteWatcher<AlarmData[], AlarmsCountByLevelWatcherParams> {
  resolveRequestConfig(params: AlarmsCountByLevelWatcherParams): RequestConfig {
    return {
      url: '/cgi/alarm/active/getActiveOverview',
      method: 'POST',
      data: {
        ...params,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data;
  }
}

/**
 * 园区告警列表监听器
 */
export class ParkAlarmsWatcher extends TedgeRemoteWatcher<AlarmData[], TedgeAlarmsWatchParams> {
  resolveRequestConfig(params: TedgeAlarmsWatchParams): RequestConfig {
    return {
      // url: '/cgi/alarm/active/getList',
      url: '/cgi/alarm/active/park/getList',
      method: 'POST',
      data: {
        DeviceNumber: params.deviceNumbers,
        eventStatusList: params.eventStatus || 1,
        ...params,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data.list;
  }
}

/**
 * 园区告警列表监听器，返回值包含总数
 */
export class ParkAlarmsWithTotalWatcher extends ParkAlarmsWatcher {
  resolveResponseData(data: any) {
    return data.data;
  }
}

/**
 * 新产生园区活动告警趋势
 */
export class ParkNewAlarmsTrendWatcher extends TedgeRemoteWatcher<AlarmData[], NewAlarmsTrendWatcherParams> {
  resolveRequestConfig(params: NewAlarmsTrendWatcherParams): RequestConfig {
    return {
      url: '/cgi/alarm/active/park/getRecentInfo',
      method: 'POST',
      data: {
        ...params,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data;
  }
}

/**
 * 告警按等级统计数量
 */
export class ParkAlarmsCountByLevelWatcher extends TedgeRemoteWatcher<AlarmData[], AlarmsCountByLevelWatcherParams> {
  resolveRequestConfig(params: AlarmsCountByLevelWatcherParams): RequestConfig {
    return {
      url: '/cgi/alarm/active/park/getActiveOverview',
      method: 'POST',
      data: {
        ...params,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data;
  }
}

export interface ParkStatisticWatcherParams {
  mozuId?: number | string;
}

/**
 * 告警统计
 */
export class ParkStatisticWatcher extends TedgeRemoteWatcher<AlarmData[], ParkStatisticWatcherParams> {
  resolveRequestConfig(params: ParkStatisticWatcherParams): RequestConfig {
    return {
      url: '/cgi/alarm/active/park/getStatistic',
      method: 'POST',
      data: {
        ...params,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data;
  }
}