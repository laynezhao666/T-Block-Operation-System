import _ = require("lodash");
import { IBaseTableContext } from "../table-layout-context";

export type ITableContextWithSelection = IBaseTableContext & {
  selection: ITableSelection;
};

export interface ITableSelection {
  identity: (row: any) => string;
  selectedLength: number;
  selectedRowsMap: Map<string, any>;
  status: 'idle' | 'selecting' | 'processing';
  oprs: Array<Vue.Component>;
  start(this: ITableSelection): void;
  cancel(this: ITableSelection, ctx: ITableContextWithSelection): void;
  getSelectedRows(this: ITableSelection): any[];
  handleSelectChange(this: ITableSelection, ctx: ITableContextWithSelection, selectedList: any[]): void;

  hideToolbar: boolean;
}

export type ITableSelectionOptions = {
  identity: string | ((row: any) => string);
  oprs?: Array<Vue.Component | 'delete' | 'export'>;
  hideToolbar?: boolean;
  // 当前只支持前者，即固定显示勾选列、勾选后自动展示选中操作、全部取消勾选后自动取消选中操作；后者由程序控制
  // toggleMode: 'select-auto-toggle' | 'manual-toggle';
};

export function tableSelection<T extends IBaseTableContext>(data: T, opts: ITableSelectionOptions): T & ITableContextWithSelection {
  const selection: ITableSelection = {
    identity: normalizeIdentity(opts.identity),
    selectedLength: 0,
    selectedRowsMap: new Map<string, any>(),
    status: 'idle',
    hideToolbar: _.isNil(opts.hideToolbar) ? false : opts.hideToolbar,
    oprs: normalizeOprs(opts.oprs || []),
    start() {
      this.status = 'selecting';
    },
    cancel(ctx: ITableContextWithSelection) {
      this.status = 'idle';
      this.selectedRowsMap = new Map();
      selection.selectedLength = 0;
      ctx.getTableRef()?.clearSelection()
      
    },
    getSelectedRows() {
      return Array.from(this.selectedRowsMap.values());
    },
    handleSelectChange,
  };

  return {
    ...data,
    tableProps: {
      ...data.tableProps,
      'row-key': selection.identity,
    },
    tableListeners: {
      ...data.tableListeners,
      'selection-change': handleSelectChange,
    },
    topBars: [
      ...data.topBars,
      () => import('./selection-toolbar.vue'),
    ],
    prefixColumns: [
      ...data.prefixColumns,
      () => import('./selection-column.vue'),
    ],
    selection,
  };
}

/** UI选中变化回调 */
const handleSelectChange = (ctx: ITableContextWithSelection, selectedRows: any[]) => {
  const {
    getTableRef,
    tableData,
    selection,
  } = ctx;

  const {
    identity,
  } = selection;

  if (selection.status === 'idle') {
    selection.start();
  }

  const selectedRowsMap = selectedRowsToMap(selectedRows, selection.identity);

  // 不分页情况下，直接替换
  if (!(ctx as any).pagination) {
    if (selectedRows.length === 0) {
      selection.cancel(ctx);
    } else {
      selection.selectedRowsMap = selectedRowsMap;
      selection.selectedLength = selectedRowsMap.size;
    }
    return;
  }

  // 考虑分页的情况，只更新当前显示的表格数据相关数据
  const oldSelectedRowsMap = selection.selectedRowsMap;
  const tableRowsIdSetToRemove = new Set(_.map(tableData, identity));

  selectedRows.forEach(newRow => {
    const id = identity(newRow);

    tableRowsIdSetToRemove.delete(id);

    // 已经包含，不用再新增
    if (oldSelectedRowsMap.has(id)) return;

    oldSelectedRowsMap.set(id, newRow)
  });

  tableRowsIdSetToRemove.forEach((id) => {
    oldSelectedRowsMap.delete(id);
  });

  selection.selectedLength = oldSelectedRowsMap.size;

  if (!selection.selectedLength) {
    selection.cancel(ctx);
  }
};

const selectedRowsToMap = (rows: any[], identity: ITableSelection['identity']): Map<string, any> => {
  const map = new Map<string, any>();

  rows.forEach(row => {
    map.set(identity(row), row);
  });

  return map;
}

export const normalizeIdentity = (identity: ITableSelectionOptions['identity']) => {
  if (typeof identity === 'string') {
    return (row: any) => {
      const id = row[identity];
      if (_.isNil(id)) warnNormalizeIdentityError(row, identity);
      return String(id);
    };
  }

  return (row: any) => {
    const id = identity(row);
    if (_.isNil(id)) warnNormalizeIdentityError(row, identity);
    return id;
  };
};

const warnNormalizeIdentityError = (row: any, identity: ITableSelectionOptions['identity']) => {
  console.error('表格布局--上下文对象--行选择器，获取行ID失败：', row, identity)
}

const normalizeOprs = (rawOprs: ITableSelectionOptions['oprs']): ITableSelection['oprs'] => {
  return rawOprs.map((opr) => {
    if (opr === 'delete') {
      return (() => import('./batch-delete.vue')) as any;
    }

    if (opr === 'export') {
      return (() => import('./batch-export.vue')) as any;
    }

    return opr;
  });
}
