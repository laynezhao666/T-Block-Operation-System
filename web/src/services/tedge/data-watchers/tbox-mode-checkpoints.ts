import { RequestConfig } from "services/polling-request-proxy/request-config";
import { TedgeRemoteWatcher } from "../remote-data-watcher";

/**
 * 动环测点监听器
 * 用法：
 * const watcher = new TboxModeCheckpointsOfGidsAttrsListWatcher([{...}], (data) => console.log(data))
 * watcher.cancel();
 */
export class TboxModeCheckpointsOfGidsAttrsListWatcher extends TedgeRemoteWatcher<any, string[]> {
  resolveRequestConfig(ids: string[]): RequestConfig {
    return {
      url: '/cgi/standard/rtd',
      method: 'POST',
      data: {
        ids
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data;
  }
}

export const TboxModeCheckpointstWatcher = TboxModeCheckpointsOfGidsAttrsListWatcher;
