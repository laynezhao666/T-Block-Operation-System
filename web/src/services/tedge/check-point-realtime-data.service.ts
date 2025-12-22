import * as _ from 'lodash';
import * as dayjs from 'dayjs';

export interface ICacheData {
  gids: string[];
  attrs: string[]
};

export interface ICheckPointRealtimeDataServiceConfig {
  /** 默认存活时间，如果在存活时间里没有再次命中，则会被抛弃 */
  cacheSurvivalTimeSeconds: number,
  /** 最大存活时间，如果总存活时间超过该值，则会被抛弃 */
  cacheMaxSurvivalTimeSeconds: number,
  /** 存活检测的检测间隔 */
  clearCheckIntervalSeconds: number;
}

const defaultCheckPointRealtimeDataServiceConfig = {
  cacheMaxSurvivalTimeSeconds: 1800, // 30分钟，30 * 60
  cacheSurvivalTimeSeconds: 180, // 3分钟，3 * 60
  clearCheckIntervalSeconds: 10000, // 不需要太精确的时间点清理缓存，所以这里设置默认为10秒
}

/**
 * 获取数据，并支持多次获取时，从第二次开始用缓存模式接口
 * 背景：后端getCurrentBizGidAttrsWithValueByConditions接口为模糊查询，要求前端在后续获取阶段用另外的接口获取。
 * 在第一次获取到测点值后，缓存下具体查询到的gid和attrId，后续相同参数请求以这两者信息调用另外的接口获取。
 * 注意/使用限制：不支持分页，每次查询最多返回10000条
 * 核心方法：fetchDataCachedMode(payload: Record<string, string>, cacheGroupKey: string, isShowLoading = true)
 */
export class CheckPointRealtimeDataService {
  constructor(config: Partial<ICheckPointRealtimeDataServiceConfig>) {
    this.config = {
      ...defaultCheckPointRealtimeDataServiceConfig,
      ...config,
    };
    this.startClearLoop();
  }

  defaultUrl = '/cgi/dataQuery/edge/getCurrentBizGidAttrsWithValueByConditions';
  /** 这个cache是从前端角度的cache，即前端缓存模式时调用的接口 */
  cachedUrl = '/cgi/dataQuery/edge/getGidAndAttrListValueMapWithoutCache';

  config: ICheckPointRealtimeDataServiceConfig;

  /** 格式：{ [缓存组KEY]: { [缓存payload的JSON]: ICacheData } } */
  cacheMap: Map<string, Map<string, ICacheData>> = new Map();

  timingClearFuncMap: Map<string, {
    /** 触发清理的时间戳，当前场景下，第一个为临时时间戳，第二个为长期时间戳 */
    triggerTsList: Array<number>;
    clear: () => void;
  }> = new Map();
  clearLoopInterval: ReturnType<typeof setInterval>;

  get axios() {
    return (window as any).Vue.prototype.$axios;
  }

  /** 销毁方法 */
  despose() {
    clearInterval(this.clearLoopInterval);
    this.cacheMap.clear();
    this.timingClearFuncMap.clear();
    this.clearLoopInterval = null;
  }

  /** 基础的获取方法 */
  fetchData(payload: Record<string, string>, isShowLoading = true) {
    return this.axios.post(this.defaultUrl, payload, isShowLoading);
  }

  /**
   * 支持缓存测点的缓存式获取方法
   * cacheGroupKey可以使用当前url路径window.location.href活pathname，是否考虑search等因素要由业务决定
   * 如果是需要同一个页面内不同部分执行不同的缓存分组，可自定义
   * */
  async fetchDataCachedMode(payload: Record<string, string>, cacheGroupKey: string, isShowLoading = true) {
    const payloadJson = JSON.stringify(payload);

    let cacheGroupDataMap = this.cacheMap.get(cacheGroupKey);

    if (!cacheGroupDataMap) {
      cacheGroupDataMap = new Map();
      this.cacheMap.set(cacheGroupKey, cacheGroupDataMap);
    }

    const cachedData = cacheGroupDataMap.get(payloadJson);

    const cacheKey = `${cacheGroupKey}.${payloadJson}`;

    if (!cachedData) {
      const data = await this.fetchData(payload, isShowLoading);
      const cacheData = this.resolveCacheData(data);
      cacheGroupDataMap.set(payloadJson, cacheData);
      this.pushClearFunc(cacheKey, () => {
        cacheGroupDataMap.delete(payloadJson);
      });
      return data;
    }

    const payloadOfCacheRequest = {
      gidWithAttrListMap: [cachedData],
      limit: 10000,
    };

    this.refreshCacheSurvivalTime(cacheKey);

    return this.axios.post(this.cachedUrl, payloadOfCacheRequest, isShowLoading);
  }

  /** 清理轮询 */
  startClearLoop() {
    if (this.clearLoopInterval) return;

    this.clearLoopInterval = setInterval(() => {
      const { timingClearFuncMap } = this;
      const nowTs = new Date().getTime();

      const keysDeleted: string[] = [];

      timingClearFuncMap.forEach((v, k) => {
        const shouldClear = v.triggerTsList.some(item => item < nowTs);
        if (shouldClear) {
          v.clear();
          keysDeleted.push(k);
        }
      });

      keysDeleted.forEach(k => {
        timingClearFuncMap.delete(k);
      });
    }, this.config.clearCheckIntervalSeconds * 1000); // 不需要太精确的时间点清理缓存，所以这里设置为10秒
  }

  private resolveCacheData(data: any): ICacheData {
    const { list } = data;

    const attrs: string[] = [];
    const gids: string[] = [];

    _.forEach(list, item => {
      gids.push(item.gid);
      attrs.push(item.attrId);
    });

    return {
      gids: _.union(gids),
      attrs: _.union(attrs),
    };
  }

  private pushClearFunc(key: string, clearFunc: () => void) {
    const {
      cacheSurvivalTimeSeconds,
      cacheMaxSurvivalTimeSeconds,
    } = this.config;

    const tmpTs = dayjs().add(cacheSurvivalTimeSeconds, 'seconds').toDate().getTime();
    const maxTs = dayjs().add(cacheMaxSurvivalTimeSeconds, 'seconds').toDate().getTime();

    this.timingClearFuncMap.set(key, {
      triggerTsList: [tmpTs, maxTs],
      clear: clearFunc,
    });
  }

  private refreshCacheSurvivalTime(cacheKey: string) {
    const clearFunData = this.timingClearFuncMap.get(cacheKey);
    if (!clearFunData) return;

    const tmpTs = dayjs().add(this.config.cacheSurvivalTimeSeconds, 'seconds').toDate().getTime();
    clearFunData.triggerTsList[0] = tmpTs;
  }
}
