import { AxiosRequestConfig } from "axios";
import { PollingProxyPluginData } from "./plugin/types";

export type RequestConfig = AxiosRequestConfig;

export interface RequestProxyConfig {
  /** 代理的客户端ID，由客户端生成，标识对于客户端来说是哪个请求 */
  clientProxyId: string;
  /** unit: ms */
  interval: number;
  /** 由于池化改造，一个后端可能有多个模组，需要通过域名区分和路由，所以必须携带当前模组域名 */
  moduleDomain: string;
  request: RequestConfig;
  plugins?: PollingProxyPluginData[];
}
