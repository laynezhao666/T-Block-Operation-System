import _ = require("lodash");
import { IBaseTableContext } from "../table-layout-context";

export type ITableRadioRowSelection = {
  title: string;
  value: any;
  identify: (row: any) => string;
  onChange: (row) => void;
};

export function radioRowSelect<T extends IBaseTableContext>(data: T, opts: ITableRadioRowSelection): T & {
  radioRowSelection: ITableRadioRowSelection,
} {
  return {
    ...data,
    prefixColumns: [
      ...data.prefixColumns,
      () => import('./radio-column.vue'),
    ],
    radioRowSelection: {
      ...opts,
    },
  };
}
