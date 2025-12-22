import { PollingProxyPlugin } from "../plugin";
import * as _ from "lodash";
import qs from "qs";
import * as dayjs from "dayjs";
import { AxiosRequestConfig } from "axios";

export type PollingPorxyPluginTimeParamsConfig = {
  [paramPath: string]: {
    paramType: 'body' | 'query';
    diff: number;
    format?: string;
  },
};

export class PollingPorxyPluginTimeParams extends PollingProxyPlugin {
  pluginSignature(config: Record<string, any>): any {
    return config;
  }

  prepareRequestConfig(requestConfig: AxiosRequestConfig, pluginConfig: PollingPorxyPluginTimeParamsConfig): AxiosRequestConfig {
    const params = pluginConfig;

    const body: Record<string, any> = {};
    const query: Record<string, any> = {};

    _.forEach(params, (param, paramPath) => {
      const value = dayjs().add(param.diff, 'milliseconds').format(param.format || 'YYYY-MM-DD HH:mm:ss');
      if (param.paramType === 'body') {
        _.set(body, paramPath, value);
      } else if (param.paramType === 'query') {
        _.set(query, paramPath, value);
      }
    });

    if (!_.isEmpty(body)) {
      requestConfig.data = {
        ...(requestConfig.data || {}),
        ...body,
      };
    }

    if (!_.isEmpty(query) && requestConfig.url) {
      if (requestConfig.url.includes('?')) {
        requestConfig.url = `${requestConfig.url}&${qs.stringify(query)}`;
      } else {
        requestConfig.url = `${requestConfig.url}?${qs.stringify(query)}`;
      }
    }

    return requestConfig;
  }
}
