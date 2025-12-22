import * as _ from "lodash";
import { parseXlsx } from './xlsx-populate.js';

export const getSheetRowTexts = (sheet: any, rowIndex: number): Array<string | undefined> => {
  return sheet.row(rowIndex)._cells.slice(1).map((cell: any) => {
    const v = cell.value();
    const text = (v?.text ? v.text() : v);

    return _.isNil(text)
      ? ''
      : text
  });
};

export const rowsToJson = (sheet: any, titleRowIndex: number | number[], rowIndexs: number[]) => {
  const titleRowIndexes = titleRowIndex instanceof Array ? titleRowIndex : [titleRowIndex];
  const titles = _.chain(titleRowIndexes)
    .map(i => getSheetRowTexts(sheet, i))
    .thru(rows => {
      const cells: any[] = [];
      rows.forEach(row => {
        row.forEach((cell, i) => {
          if (cell) {
            cells[i] = cell;
          }
        });
      });

      return cells;
    })
    .value();

  _.chain(titles)
    .map((title, i) => [title, i])
    .groupBy(0)
    .filter(items => items.length > 1)
    .forEach((items) => {
      items.forEach((item, i) => {
        titles[item[1]] = `${item[0]}-${i}`;
      });
    })
    .value();

  const rows = rowIndexs.map(i => getSheetRowTexts(sheet, i));

  const objList = rows.map(row => {
    const obj = _.zipObject(titles, row);
    delete obj.undefined;
    return obj;
  });

  return objList;
};

export const findRowIndexUntil = (sheet: any, func: (row: any) => boolean) => {
  return _.findLastIndex(sheet._rows, r => func(r));
};

export const findRowIndexUntilHasContent = (sheet: any, colIndexTest: number = 1) => {
  return findRowIndexUntil(sheet, (row) => {
    const value = row._cells[colIndexTest]?.value();
    return Boolean(value?.text ? value.text() : value);
  });
}

export const getCellTextValue = (cell: any, defaultValue?: string) => {
  const value = cell?.value();

  if (_.isNil(value) || value === '') {
    return _.isNil(defaultValue)
      ? value
      : defaultValue;
  }

  if (typeof value === 'number') {
    return String(value);
  }

  if (value._error) {
    return '';
  }

  return typeof value !== 'string'
    ? value.text()
    : value;
};

export const parseXlsxFileToJson = async <T extends Record<string, string>>(file: File, forceCellText: boolean = true): Promise<T[]> => {
  // @ts-ignore
  const XlsxPopulate: any = (await import('xlsx-populate/browser/xlsx-populate-no-encryption')).default;

  const workbook = await XlsxPopulate.fromDataAsync(file);

  const sheet = workbook.sheet(0);
  return parseXlsxSheetToJson(sheet, forceCellText);
};

export const parseXlsxSheetToJson = <T extends Record<string, string>>(sheet: any, forceCellText: boolean = true, headerRowIndex: number = 1): T[] => {
  const rows = sheet._rows;

  const headerRow = rows[headerRowIndex];

  if (!headerRow) return [];

  const list: T[] = [];

  const headers: string[] = headerRow._cells.filter((cell: any) => !!cell).map((cell: any) => {
    return getCellTextValue(cell);
  });

  for (let i = headerRowIndex + 1; i < rows.length; i++) {
    const row = rows[i];
    if (!row) continue;
    list.push(
      _.fromPairs(
        headers.map((header, colIndex) => {
          const cell = row.cell(colIndex + 1);
          const resultValue = forceCellText
            ? getCellTextValue(cell)
            : cell?.value()
          return [header, resultValue];
        })
      ) as any,
    );
  }

  return list;
};

export interface PairsToXlsxOption {
  sheetName?: string;
  headers: string[];
  data: any[][];
  colWidths?: number[];
}
export const pairsToXlsx = async (opts: PairsToXlsxOption): Promise<Blob> => {
  // @ts-ignore
  const XlsxPopulate: any = (await import('xlsx-populate/browser/xlsx-populate-no-encryption')).default;
  const workbook = await XlsxPopulate.fromBlankAsync();
  const sheet = workbook.sheet('Sheet1');

  if (opts.sheetName) {
    sheet.name(opts.sheetName);
  }

  if (opts.colWidths) {
    opts.colWidths.forEach((colWidth, i) => {
      sheet.column(i + 1).width(colWidth);
    });
  }

  const headerRow = sheet.row(1);
  opts.headers.forEach((col, i) => {
    headerRow.cell(i + 1).value(col);
  });

  opts.data.forEach((pair, rowIndex) => {
    const row = sheet.row(rowIndex + 2);
    pair.forEach((item, colIndex) => {
      row.cell(colIndex + 1).value(item);
    });
  });

  const blob = await workbook.outputAsync();
  return blob;
};

export interface ListToXlsxOptionFieldConfig {
  label: string;
  format?: (value: any) => string;
  revertFormat?: (text: string) => any;
  width?: number;
  colIndex?: number;
}

export interface ListToXlsxOption {
  sheetName?: string;
  list: Record<string, any>[];
  ignoreFieldsNotInInMap?: boolean;
  headerIncludeFields?: boolean;
  defaultWidth?: number;
  fieldsMap?: {
    [field: string]: ListToXlsxOptionFieldConfig;
  };
}

export const listToXlsx = async (opts: ListToXlsxOption): Promise<Blob> => {
  // @ts-ignore
  const XlsxPopulate: any = (await import('xlsx-populate/browser/xlsx-populate-no-encryption')).default;
  const workbook = await XlsxPopulate.fromBlankAsync();
  const sheet = workbook.sheet('Sheet1');

  listToXlsxSheet(sheet, opts);

  const blob = await workbook.outputAsync();
  return blob;
};

export const listToXlsxSheet = (sheet: any, opts: ListToXlsxOption) => {
  if (opts.sheetName) {
    sheet.name(opts.sheetName);
  }

  const fieldsMap = new Map(Object.entries(opts.fieldsMap || {}));

  let lastColIndex = 0;

  fieldsMap.forEach(item => {
    if (!_.isNil(item.colIndex) || (item.colIndex || 0) <= lastColIndex) return;

    lastColIndex = item.colIndex!;
  });

  fieldsMap.forEach(item => {
    if (!_.isNil(item.colIndex)) return;

    item.colIndex = lastColIndex + 1;
    lastColIndex = item.colIndex;
  });

  opts.list.forEach((item, i) => {
    const row = sheet.row(i + 2);

    _.forEach(item, (v, k) => {
      let fieldConfig = fieldsMap.get(k);
      if (!fieldConfig && !opts.ignoreFieldsNotInInMap) {
        fieldsMap.set(k, {
          label: k,
          colIndex: lastColIndex + 1,
        });

        fieldConfig = fieldsMap.get(k);
        lastColIndex = fieldConfig!.colIndex!;
      }

      if (!fieldConfig) return;

      const cell = row.cell(fieldConfig.colIndex);

      cell.value(fieldConfig.format ? fieldConfig.format(v) : v);
    });
  });

  const buildHeaderText = opts.headerIncludeFields ? ((fieldConfig: ListToXlsxOptionFieldConfig, field: string) => {
    return fieldConfig.label === field ? field : `${fieldConfig.label}/${field}`;
  }) : ((fieldConfig: ListToXlsxOptionFieldConfig) => {
    return fieldConfig.label;
  });

  fieldsMap.forEach((fieldConfig, field) => {
    const headerRow = sheet.row(1);
    headerRow.cell(fieldConfig.colIndex).value(buildHeaderText(fieldConfig, field));
    sheet.column(fieldConfig.colIndex).width(fieldConfig.width || opts.defaultWidth);
  });
};

export const parseXlsxFromUrl = async (url: string) => {
  const blob = await fetch(url).then(resp => resp.blob());

  const workbook = await parseXlsx(blob);

  return workbook;
};

export const deleteColumnsOfXlsx = async (sheet: any, columnsToDelete: number[]) => {
  columnsToDelete = _.orderBy(columnsToDelete, _.identity);

  // 删除列
  for (const columnIndex of columnsToDelete) {
    (sheet as any)._columns = (sheet as any)._columns.filter(
      (_: any, index: number) => !index || index !== columnIndex
    );

    (sheet as any)._colsNode.children = (sheet as any)._colsNode.children.filter(
      (_: any, index: number) => !index || index !== columnIndex
    );

    (sheet as any)._colNodes = (sheet as any)._colNodes.filter(
      (_: any, index: number) => !index || index !== columnIndex
    );
  }

  // 删除行里的单元格
  sheet._rows = sheet._rows.map((row: any) => {
    (row as any)._cells = (row as any)._cells
      .filter((cell: any) => !columnsToDelete.includes(cell._columnNumber))
      .map((cell: any, index: number) => {
        cell._columnNumber = index;
        return cell;
      });
    (row as any)._node.children = (row as any)._node.children
      .filter((cell: any) => !columnsToDelete.includes(cell._columnNumber))
      .map((cell: any, index: number) => {
        cell._columnNumber = index;
        return cell;
      });
    return row;
  });

  // 合并单元格，删除列
  const lastColIndexToDelete = _.last(columnsToDelete)!;
  const columnsOffsets = _.chain(_.range(0, lastColIndexToDelete + 1, 1))
    .map(item => {
      if (item < columnsToDelete[0]) {
        return 0;
      }

      if (item > columnsToDelete[columnsToDelete.length - 1]) {
        return columnsToDelete.length;
      }

      return _.filter(columnsToDelete, colToDelete => colToDelete <= item).length;
    }).value()

  const mergeCells = sheet._mergeCells as Record<string, any>;
  const newMergeCells: Record<string, any> = {};

  for (const key of Object.keys(mergeCells)) {
    const [startCell, endCell] = key.split(':');

    const startCol = startCell.replace(/\d+/g, '')
    const endCol = endCell.replace(/\d+/g, '')

    let [
      newStartColIndex,
      newEndColIndex,
    ] = [
      startCol,
      endCol,
    ].map(colName => {
      let colIndex = colName.charCodeAt(0) - 64;

      if (colIndex > lastColIndexToDelete) {
        return colIndex - columnsToDelete.length;
      }

      return colIndex - columnsOffsets[colIndex]
    });

    if (newStartColIndex < 1) {
      newStartColIndex = 1;
    }

    // 开始小于结束，或者非法结束index，删除合并单元格，跳过
    if (newStartColIndex > newEndColIndex || newEndColIndex < 1) {
      continue;
    }

    const newKey = [
      startCell.replace(startCol, String.fromCharCode(newStartColIndex + 64)),
      endCell.replace(endCol, String.fromCharCode(newEndColIndex + 64)),
    ].join(':')

    newMergeCells[newKey] = mergeCells[key];
  }

  sheet._mergeCells = newMergeCells;
  _.forEach(sheet._mergeCells, (item, key) => item.attributes.ref = key);
  sheet._mergeCellsNode.attributes.count = _.keys(sheet._mergeCells).length;

  sheet._dataValidations = {}
};
