import { PollingProxyPlugin } from '../plugin';
import * as _ from 'lodash';

export interface PollingPorxyPluginSummaryListConfig {
  [summaryItemId: string]: any[];
}

export class PollingPorxyPluginSummaryList extends PollingProxyPlugin {
  beforeEmit(): boolean {
    return true;
  }

  postRequest(config: PollingPorxyPluginSummaryListConfig, data: any) {
    const summaryResult = _.mapValues(config, configItem => computeSummary(data, configItem));

    return summaryResult;
  }
}

interface ICondition {
  opr: string,

  config: {
    // fieldPath: string,
    // value: any,
    [key: string]: any;
  },

  cachedData?: {
    [key: string]: any;
  },
}

const summaryRulesMap: {
  [id: string]: (data: any, config: any) => any;
} = {
  countBy: (data: any, config: { listField?: string, condition: ICondition }) => {
    const list: any[] = config.listField ? _.get(data, config.listField) : data;
    const { condition } = config;
    const filtered =  _.filter(list, item => computeCondition(item, condition));
    return filtered.length;
  },
  sumBy: (data: any, config: { listField?: string, condition: ICondition, sumField: string }) => {
    const list: any[] = config.listField ? _.get(data, config.listField) : data;
    const { condition } = config;
    const filtered =  _.filter(list, item => computeCondition(item, condition));
    return _.sumBy(filtered, config.sumField);
  },

  countByAndConditions: (data: any, config: { listField?: string, conditions: ICondition[] }) => {
    const list: any[] = config.listField ? _.get(data, config.listField) : data;
    const filtered =  _.filter(list, item => config.conditions.every(cond => computeCondition(item, cond)));

    return filtered.length;
  },
  countByOrConditions: (data: any, config: { listField?: string, conditions: ICondition[] }) => {
    const list: any[] = config.listField ? _.get(data, config.listField) : data;
    const filtered =  _.filter(list, item => config.conditions.some(cond => computeCondition(item, cond)));

    return filtered.length;
  },
  sumByAndConditions: (data: any, config: { conditions: ICondition[], sumField: string, listField?: string }) => {
    const list: any[] = config.listField ? _.get(data, config.listField) : data;
    const filtered =  _.filter(list, item => config.conditions.every(cond => computeCondition(item, cond)));

    return _.sumBy(filtered, config.sumField);
  },
  sumByOrConditions: (data: any, config: { conditions: ICondition[], sumField: string, listField?: string }) => {
    const list: any[] = config.listField ? _.get(data, config.listField) : data;
    const filtered =  _.filter(list, item => config.conditions.some(cond => computeCondition(item, cond)));

    return _.sumBy(filtered, config.sumField);
  },
};

const conditionComputersMap = {
  in(data: any, condition: ICondition) {
    if (!condition.cachedData?.inConditionFunc) {
      if (!condition.cachedData) {
        // eslint-disable-next-line no-param-reassign
        condition.cachedData = {};
      }

      // eslint-disable-next-line no-param-reassign
      condition.cachedData.inConditionFunc = ((valueSet: Set<any>) => {
        // 用Set.has替代Array.includes，根据7000+告警数据实测，性能提升4~5倍
        const func = value => valueSet.has(value);
        return func;
      })(new Set(condition.config.list));
    }

    const value = condition.config.fieldPath ? _.get(data, condition.config.fieldPath) : data;
    return condition.cachedData.inConditionFunc(value);
  },
  compare(data: any, condition: ICondition & {
    config: {
      type: keyof typeof _,
      fieldPath?: string, matchValue: any,
    },
  }) {
    const value = condition.config.fieldPath ? _.get(data, condition.config.fieldPath) : data;
    const func = _[condition.config.type || 'eq'] as any;
    return func(value, condition.config.matchValue);
  },
  or(data: any, condition: ICondition & {
    config: {
      conditions: ICondition[];
    },
  }) {
    return condition.config.conditions.some(subCondition => computeCondition(data, subCondition));
  },
  and(data: any, condition: ICondition & {
    config: {
      conditions: ICondition[];
    },
  }) {
    return condition.config.conditions.every(subCondition => computeCondition(data, subCondition));
  },
};

const computeSummary = (list: any, configData: any) => summaryRulesMap[configData.opr]?.(list, configData.config);

const computeCondition = (data: any, condition: ICondition) => {
  const result = conditionComputersMap[condition.opr]?.(data, condition);
  return result;
};
