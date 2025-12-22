// import { PollingProxy } from "socket/entities/polling-proxy.entity";

import { AxiosRequestConfig } from "axios";

export abstract class PollingProxyPlugin {
  postRequest(config: Record<string, any>, data: any, proxy: any): any {
    return data;
  };
  prepareRequestConfig(requestConfig: AxiosRequestConfig, pluginConfig: Record<string, any>): any {
    return requestConfig
  }
}
