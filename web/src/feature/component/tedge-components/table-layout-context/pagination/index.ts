import _ = require("lodash");
import { IBaseTableContext } from "../table-layout-context";

export interface IPagination {
  current: number;
  size: number;
  total: number;
  pageSizes: number[];
}

export type ITableContextWithPagination = IBaseTableContext & {
  pagination: IPagination;
};

export function pagination<T extends IBaseTableContext>(data: T, opts?: Partial<IPagination>): T & ITableContextWithPagination {
  const reload = function (context) {
    context.loadData();
  };

  data.watches.push({
    expOrFn: 'pagination.current',
    callback: reload,
  }, {
    expOrFn: 'pagination.size',
    callback: reload,
  });

  return {
    ...data,
    pagination: {
      current: 1,
      size: 10,
      total: 0,
      pageSizes: [10, 20, 30, 40, 50, 100],
      ...(opts || {}),
    },
    footerBars: [
      ...data.footerBars,
      () => import('./pagination.vue'),
    ],
  };
}