import * as _ from "lodash";
import { ITableContextWithPagination } from ".";
import { IBaseTableContext, ITableSearch } from "../table-layout-context";

  /** 本地过滤及分页 */
  export function localFilterPagination<T extends ITableContextWithPagination>(data: T): T {
    const sourceFetch = data.fetch;

    let localData: any[] | null = null;

    data.forceFetch = async function(...args: any[]) {
      localData = null;
      return (data.fetch as any)(...args);
    };

    data.fetch = async (...args: any[]) => {
      const ctx: IBaseTableContext = args[args.length - 1];

      if (!localData) {
        localData = await (sourceFetch as any)(...args);
      }

      const filters: { [key: string]: any } | undefined = (ctx as any).filters;
      const search: ITableSearch | undefined = (ctx as any).search;

      const isFiltersEmpty = !filters || _.size(filters) === 0;
      const isSearchEmpty = !search?.value?.trim();

      const filteredData = isFiltersEmpty && isSearchEmpty
      ? localData
      : localData?.filter(item => {
        return (isFiltersEmpty || _.every(filters, (v, k) => (v === '' || item[k] === v)))
          && (isSearchEmpty || _.some(item, (v) => {
            if (typeof v === 'object') {
              try {
                const json = JSON.stringify(v);
                return json.includes(search!.value);
              } catch(err) {
                console.log(err);
                return _.some(v, (v, k) => v.toString().includes(search!.value));
              }
            }

            return v?.includes?.(search!.value);
          }))
      });

      const {
        current,
        size,
      } = data.pagination;

      data.pagination.total = filteredData?.length || 0;

      return _.slice(filteredData, (current - 1) * size, current * size);
    };

    return data;
  }