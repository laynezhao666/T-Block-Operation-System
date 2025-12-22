// import { PollingProxy } from "socket/entities/polling-proxy.entity";
import { PollingProxyPlugin } from "../plugin";
import * as _ from "lodash";

export interface PollingPorxyPluginCountByConfig {
  pickField?: string;
  countByField: string;
}

export class PollingPorxyPluginCountBy extends PollingProxyPlugin {
  postRequest(config: PollingPorxyPluginCountByConfig, data: any, proxy: any) {
    const list = config.pickField
      ? _.get(data, config.pickField)
      : data;

    return _.countBy(list, config.countByField);
  }
}
