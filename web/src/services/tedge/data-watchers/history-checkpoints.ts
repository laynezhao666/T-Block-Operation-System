import { RequestConfig } from "services/polling-request-proxy/request-config";
import { TedgeRemoteWatcher } from "../remote-data-watcher";

export interface TedgeGidsAttrs {
  gids: string[];
  attrs: string[];
};

export interface CheckpointsHistoryWatcherParams {
  gidWithAttrList: TedgeGidsAttrs[],
  /** 秒 */
  startTime: string,
  /** 秒 */
  endTime: string,
  /** 秒 */
  interval: number,
};

/**
 * 如：{ "31xxxxxx": { 1699848720: "3767", 1699848722: "3867" } }
 */
export interface CheckPointHistoryData {
  [gidAndPointId: string]: {
    [ts: string]: string;
  };
}


/**
 * 动环测点监听器
 * 用法：
 * const watcher = new CheckpointsHistoryWatcher([{...}], (data) => console.log(data))
 * watcher.cancel();
 */
export class CheckpointsHistoryWatcher extends TedgeRemoteWatcher<CheckPointHistoryData, CheckpointsHistoryWatcherParams> {
  resolveRequestConfig(params: CheckpointsHistoryWatcherParams): RequestConfig {
    return {
      url: '/cgi/dataQuery/edge/getHistoryBizGidAttrValues',
      method: 'POST',
      data: params,
    };
  }

  resolveResponseData(data: any) {
    return data.data;
  }
}
