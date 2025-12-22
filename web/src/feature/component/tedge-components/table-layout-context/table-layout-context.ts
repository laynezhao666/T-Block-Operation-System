import * as _ from "lodash";
import { WatchOptions } from "vue";
import { baseCurd, IBaseCurd } from "./base-curd";
import { curdFormModal, ICurdFormModal, ICurdFormModalOptions } from "./curd-form-modal";
import { IPagination, pagination } from "./pagination";
import { localFilterPagination } from "./pagination/local-pagination";
import { IRemoteFilterPaginationOptions, remoteFilterPagination } from "./pagination/remote-pagination";
import { ITableRadioRowSelection, radioRowSelect } from "./radio-row-selection";
import { ITableSelectionOptions, normalizeIdentity, tableSelection } from "./selection";
import { ITableStyleOptions, tableStyle } from "./table-style";

interface ITableLayoutChainCall<T> {
  func: Function;
  args: any[];
}

type OmitFirstArg<F> = F extends (x: any, ...args: infer P) => infer R ? P : never;

export type ITableLayoutChain<T extends IBaseTableContext> = {
  _data: T,
  _calls: Array<ITableLayoutChainCall<any>>;
  done: () => T;
} & {
  [key in keyof ChainTableMethodsMap<T>]: (...args: OmitFirstArg<ChainTableMethodsMap<T>[key]>) => ITableLayoutChain<ReturnType<ChainTableMethodsMap<T>[key]>>;
}

export interface IBaseTableContext {
  fetch: TableFetcher<any>;
  forceFetch: TableFetcher<any>;
  loadData: () => Promise<any>;
  forceReloadData: () => Promise<any>;
  tableData: any[];

  getTableRef: null | (() => any);

  hideToolbar: boolean,

  tableProps: {
    [key: string]: any;
  };
  tableListeners: {
    [key: string]: Function;
  };

  prefixColumns: any[];
  suffixColumns: any[];
  extras: Array<Vue.Component>;
  oprsColumnOprs: Array<Vue.Component>;

  topBars: Array<Vue.Component>;
  footerBars: Array<Vue.Component>;

  modals: Array<Vue.Component>;
  watches: Array<{
    expOrFn: string,
    callback: (this: Vue, context: IBaseTableContext, n: any, o: any) => void,
    options?: WatchOptions
  }>;
}

export interface ITableSearch {
  value: string;
  placeholder: string;
  isHide: boolean;
  doSearch: (context: IBaseTableContext) => any;
}

export interface ITableColumn {
  label?: string;
  width?: number;
  fixed?: 'left' | 'right';
}

export type TableFetcher<T> = (
  filters,
  search,
  pagination,
  tableLayoutContext,
) => T | Promise<T>;

// 为了类型推导，借用class解析泛型
class ChainTableMethodsMap<T extends IBaseTableContext> {
  hideToolbar(data: T): T {
    return {
      ...data,
      hideToolbar: true,
    }
  }

  pagination(data: T, opts?: Partial<IPagination>): T & { pagination: IPagination } {
    return pagination(data, opts);
  }

  /** 远程过滤及分页 */
  remoteFilterPagination(data: T, opts?: IRemoteFilterPaginationOptions): T {
    return remoteFilterPagination(data as any, opts);
  }

  /** 本地过滤及分页 */
  localFilterPagination(data: T): T {
    return localFilterPagination(data as any);
  }

  search(data: T, opts?: Partial<ITableSearch>): T & { search: ITableSearch } {
    const doSearch = function(context: IBaseTableContext) {
      const currentPage = (context as any).pagination?.current || 1;

      if (currentPage > 1) {
        (context as any).pagination.current = 1;
      } else {
        context.forceReloadData();
      }
    }

    return {
      ...data,
      search: {
        value: '',
        placeholder: '',
        isHide: false,
        doSearch,
        ...opts,
      },
    };
  }

  filters<K extends { [key: string]: any }>(data: T, filtersMap: K, opts?: {
    isResetPagination?: boolean,
    filtersForm?: Vue,
  }): T & { filters: { [key: string]: any } } {
    const watches = data.watches.concat({
      expOrFn: 'filters',
      callback: function(context) {
        let shouldTriggerReload = true;
        if (opts?.isResetPagination !== false) {
          const currentPage = (context as any).pagination?.current || 1;
          if (currentPage > 1) {
            (context as any).pagination.current = 1;
            // 由分页变化来触发重新加载数据
            shouldTriggerReload = false;
          }
        }

        if (shouldTriggerReload) {
          context.loadData();
        }
      },
      options: {
        deep: true,
      },
    });

    return {
      ...data,
      watches,
      filters: filtersMap,
      topBars: opts?.filtersForm ? [
        ...data.topBars,
        opts.filtersForm,
      ] : data.topBars,
    };
  }

  indexColumn(data: T, opts?: ITableColumn): T & { indexColumn: ITableColumn } {
    return {
      ...data,
      indexColumn: {
        ...(opts || {}),
      },
    };
  }

  baseCurd(data: T, opts?: Partial<IBaseCurd>): T & {
    curd: IBaseCurd;
  } {
    return baseCurd<T>(data, opts);
  }

  curdFormModal(data: T, opts: ICurdFormModalOptions) {
    return curdFormModal<T>(data, opts);
  }

  selection(data: T, opts: ITableSelectionOptions) {
    return tableSelection(data, opts);
  }

  tableStyle(data: T, opts: ITableStyleOptions) {
    return tableStyle(data, opts);
  }

  radioRowSelect(data: T, opts: ITableRadioRowSelection) {
    return radioRowSelect(data, opts);
  }

  extraBtn(data: T, comp: Vue.Component): T {
    return {
      ...data,
      extras: [
        ...data.extras,
        comp,
      ],
    };
  }

  toolbarActions(data: T, action: {
    text: string;
    icon: string;
    action: Function;
  }): T {
    return {
      ...data,
      toolbarActions: [
        ...((data as any).toolbarActions || []),
        action,
      ],
    };
  }
}

const wrapMethod = <T extends Function>(func: T, getChain: () => ITableLayoutChain<any>) => {
  return (...args: any[]) => {
    const chain = getChain();
    chain._calls.push({
      func,
      args,
    });

    return chain;
  };
};

const done = function <T extends IBaseTableContext>(this: ITableLayoutChain<T>): T {
  return this._calls.reduce((data: any, callData) => {
    return callData.func(data, ...callData.args);
  }, this._data);
}

const baseLoadData = async function(this: IBaseTableContext) {
  const {
    filters,
    search,
    pagination,
  } = this as any;
  this.tableData = await this.fetch(filters, search?.value, pagination, this);
}

const forceReloadData = async function(this: IBaseTableContext) {
  const {
    filters,
    search,
    pagination,
  } = this as any;
  this.tableData = await this.forceFetch(filters, search?.value, pagination, this);
}

const chainTableMethodsMap = {
  ...ChainTableMethodsMap.prototype,
  ...new ChainTableMethodsMap(),
} as unknown as ChainTableMethodsMap<any>;

export const chainTableLayout = (fetch: TableFetcher<any>): ITableLayoutChain<IBaseTableContext> => {
  const chainData: ITableLayoutChain<IBaseTableContext> = {
    _data: {
      fetch,
      forceFetch: fetch,
      loadData: baseLoadData,
      forceReloadData: forceReloadData,
      tableData: [],
      getTableRef: null,
      hideToolbar: false,
      tableProps: {},
      tableListeners: {},
      prefixColumns: [],
      suffixColumns: [],
      extras: [],
      oprsColumnOprs: [],
      topBars: [],
      footerBars: [],
      modals: [],
      watches: [],
    },
    _calls: [],
    done,
    ..._.mapValues((chainTableMethodsMap), (func) => wrapMethod(func, () => chainData)),
  };

  return chainData;
}

export {
  normalizeIdentity,
}
