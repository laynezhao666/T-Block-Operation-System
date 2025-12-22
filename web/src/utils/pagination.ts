
/** 将分页解析为后端等使用的offset+limit */
export const curryingResolveOffsetLimitOfPagination = <T extends string, K extends string>(paginationFields: {
  currentPageField: T;
  pageSizeField: K;
}) => {
  return function <U extends string, V extends string>(offsetLimitFields: {
    offsetField: U;
    limitField: V;
  }) {
    type Pagination = {
      [key in (T | K)]: number;
    };
    type Result = {
      [key in (U | V)]: number;
    };

    return function (pagination: Pagination): Result {
      const currentPage = pagination[paginationFields.currentPageField];
      const pageSize = pagination[paginationFields.pageSizeField];

      return {
        [offsetLimitFields.offsetField]: (currentPage - 1) * pageSize,
        [offsetLimitFields.limitField]: pageSize,
      } as any;
    };
  }
}

/** 分页字段为：current+size */
export const curryingResolveOffsetLimitOfSimplePagination = curryingResolveOffsetLimitOfPagination({
  currentPageField: 'current',
  pageSizeField: 'size',
});

/** 分页字段为：current+size，解析为offset+limit */
export const defaultResolveOffsetLimitOfPagination = curryingResolveOffsetLimitOfSimplePagination({
  offsetField: 'offset',
  limitField: 'limit',
});
