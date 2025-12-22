// import { PollingProxy } from "socket/entities/polling-proxy.entity";
import { PollingProxyPlugin } from "../plugin";
import * as _ from "lodash";

export interface PollingPorxyPluginTransformConfig {
  func?: (data: any) => any;
}

export class PollingPorxyPluginTransform extends PollingProxyPlugin {
  postRequest(config: PollingPorxyPluginTransformConfig, data: any, proxy: any) {
    return config.func(data);
  }
}
