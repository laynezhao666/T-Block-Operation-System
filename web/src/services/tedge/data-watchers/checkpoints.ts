import { RequestConfig } from "services/polling-request-proxy/request-config";
import { TedgeRemoteWatcher } from "../remote-data-watcher";

export interface TedgeGidsAttrs {
  gids: string[];
  attrs: string[];
};

export interface CheckPointData {
  applicationTypeEn: string;
  applicationTypeZh: string;
  attrId:            string;
  attrName:          string;
  categoryEn:        string;
  categoryZh:        string;
  deviceNumber:      string;
  deviceTypesName:   string;
  enumValue:         string;
  gid:               string;
  id:                string;
  readAndWrite:      boolean;
  simulation:        boolean;
  status:            boolean;
  templatePointId:   string;
  unit:              string;
  updateTime:        string;
  value:             string;
}


/**
 * 动环测点监听器
 * 用法：
 * const watcher = new CheckpointsOfGidsAttrsListWatcher([{...}], (data) => console.log(data))
 * watcher.cancel();
 */
export class CheckpointsOfGidsAttrsListWatcher extends TedgeRemoteWatcher<CheckPointData[], TedgeGidsAttrs[]> {
  resolveRequestConfig(params: TedgeGidsAttrs[]): RequestConfig {
    return {
      url: '/cgi/dataQuery/edge/getGidAndAttrListValueMapWithoutCache',
      method: 'POST',
      data: {
        gidWithAttrListMap: params,
        limit: 999999,
      },
    };
  }

  resolveResponseData(data: any) {
    return data.data.list;
  }
}
