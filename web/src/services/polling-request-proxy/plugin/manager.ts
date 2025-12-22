// import { PollingProxy } from "socket/entities/polling-proxy.entity";
import { PollingProxyPlugin } from "./plugin";
import { PollingProxyPluginData } from "./types";
import * as _ from "lodash";
import { PollingPorxyPluginCountBy } from "./plugins/count-by.plugin";
import { AxiosRequestConfig, AxiosResponse } from "axios";
import { PollingPorxyPluginTimeParams } from "./plugins/time-params.plugin";
import { PollingPorxyPluginSummaryList } from './plugins/summary-list.plugin';
import { PollingPorxyPluginTransform } from './plugins/transform.plugin';

const buildBuildInPluginsMap = () => {
  const pluginsMap: Map<string, PollingProxyPlugin> = new Map();

  _.forEach({
    countBy: new PollingPorxyPluginCountBy(),
    timeParams: new PollingPorxyPluginTimeParams(),
    summaryList: new PollingPorxyPluginSummaryList(),
    transform: new PollingPorxyPluginTransform(),
  }, (plugin, id) => {
    pluginsMap.set(id, plugin);
  });

  return pluginsMap;
}

export class PollingPorxyPluginManager {
  pluginsMap: Map<string, PollingProxyPlugin> = buildBuildInPluginsMap();

  postRequest(
    pluginDataList: PollingProxyPluginData[],
    resp: AxiosResponse<any>,
    proxy: any,
  ) {
    if (!pluginDataList?.length) return resp;

    const { pluginsMap } = this;

    const resultData = pluginDataList.reduce((currentData, pluginData) => {
      const plugin = pluginsMap.get(pluginData.id);
      return plugin
        ? plugin.postRequest(pluginData.config, currentData, proxy)
        : currentData;
    }, resp.data);

    return {
      ...resp,
      data: resultData,
    };
  }

  prepareRequestConfig(requestConfig: AxiosRequestConfig, pluginsDataList: PollingProxyPluginData[]): AxiosRequestConfig {
    return !pluginsDataList?.length ? requestConfig : pluginsDataList.reduce((result, pluginData) => {
      return this.pluginsMap.get(pluginData.id)?.prepareRequestConfig(result, pluginData.config) || result;
    }, requestConfig);
  }
}
