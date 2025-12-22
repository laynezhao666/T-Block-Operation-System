import _ = require("lodash");
import { IBaseTableContext } from "../table-layout-context";

export type ITableStyleOptions = {
  /** 是否带有纵向边框，开启后内容行的列显示边框，且有可拖动改变大小特性，默认false */
  border?: boolean;
  /** 是否为斑马纹 table，默认false */
  stripe?: boolean;
  /** Table 的尺寸，默认空 */
  size?: 'medium' | 'small' | 'mini';
  /** 列的宽度是否自撑开，默认true */
  fit?: boolean;
  /** 是否显示表头，默认true */
  showHeader?: boolean;
  /** 是否要高亮当前行，默认false */
  highlightCurrentRow?: boolean;
  /** 高度 */
  height?: number;
  /** 行的 className 的回调方法，也可以使用字符串为所有行设置一个固定的 className。 */
  rowClassName?: string | ((row: any, rowIndex: number) => string);
  /** 合并行或列的计算方法 */
  spanMethod?: (params: {row: any, rowIndex: number, column: any, columnIndex: number}) => string;
};

export function tableStyle<T extends IBaseTableContext>(data: T, opts: ITableStyleOptions): T {
  return {
    ...data,
    tableProps: {
      ...data.tableProps,
      ...opts,
    },
  };
}
