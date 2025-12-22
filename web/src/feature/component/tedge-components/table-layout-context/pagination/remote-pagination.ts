import _ = require("lodash");
import { IPagination, ITableContextWithPagination } from ".";

// 本地筛选、本地分页=>替换fetch
// 远程筛选、远程分页=>替换fetch
// 数据获取阶段，远程数据处理阶段，
// fetch拉接口=>分离总数及list=>设置tableData及pagination=>过滤或分页，重复执行
// fetch拉接口=>统计总数及list=>设置tableData及pagination=>过滤或分页=>计算总计及list=>设置tableData及Pagination
export interface IRemoteFilterPaginationOptions {
  totalFields: number;
  listFields: number;
}

/** 本地过滤及分页 */
export function remoteFilterPagination<T extends ITableContextWithPagination>(data: T, opts?: IRemoteFilterPaginationOptions): T {
  const sourceFetch = data.fetch;
  const pagedData = data as T & { pagination: IPagination };

  pagedData.fetch = pagedData.forceFetch = async (...args: any[]) => {
    const result = await (sourceFetch as any)(...args);

    const {
      totalFields = 'total',
      listFields = 'list',
    } = opts || {};

    pagedData.pagination.total = result[totalFields];
    return result[listFields];
  };

  return pagedData;
}