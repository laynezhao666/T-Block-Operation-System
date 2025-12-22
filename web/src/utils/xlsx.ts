import * as _ from 'lodash';
import * as XlsxPopulate from 'xlsx-populate';

export interface IExportElTableOptions {
  sheetName: string;
  fileName: string;
  includeIndexColumn: boolean;
  excludeColumnTitles: string[];
}

export const exportElTable = async (elTableRef: any, opts: Partial<IExportElTableOptions> = {}) => {
  const options = {
    sheetName: 'Sheet1',
    fileName: '数据.xlsx',
    includeIndexColumn: false,
    excludeColumnTitles: [],
    ...opts,
  };

  const workbook = await XlsxPopulate.fromBlankAsync();
  const sheet = workbook.sheet(options.sheetName);

  interface IColumn {
    title: string;
    field: string;
    width: number | undefined;
    type: 'index' | 'selection' | 'default',
  };

  const columns: Array<IColumn> = _.map(elTableRef.columns, item => ({
    title: item.label,
    field: item.property,
    width: item.realWidth || item.width,
    type: item.type
  }));

  const columnRetainArray = _.map(columns, col => {
    return col.type !== 'selection' // 排除selection列
      && !(!options.includeIndexColumn && col.type === 'index') // 排除不保留index时的index列
      && !options.excludeColumnTitles.includes(col.title);
  });

  const checkIsColumnIndexRetain = (colIndex: number) => {
    return columnRetainArray[colIndex];
  }

  const dataRows = _.map((elTableRef.$el as HTMLDivElement).querySelectorAll('.el-table__row'), trElt => {
    return _.map(trElt.querySelectorAll('td'), tdElt => tdElt.innerText.trim());
  });

  let skipColumnCount = 0;
  _.forEach(columns, (row: IColumn, i: number) => {
    if (!checkIsColumnIndexRetain(i)) {
      skipColumnCount += 1;
      return;
    }

    const colNumber = String.fromCharCode('A'.charCodeAt(0) + i - skipColumnCount);

    sheet.column(colNumber).width(row.width / 8);

    return;
  });

  const rowArr = _.map([
    columns.map(item => item.title), // 标题行
    ...dataRows
  ], row => _.filter(row, (row, i) => checkIsColumnIndexRetain(i)) // 排除不保留的列
  );

  sheet.cell('A1').value(rowArr);

  const blob = await workbook.outputAsync();
  downloadBlob(blob, options.fileName);
}

// export async function exportXlsx(data, name, cols) {
//   /* convert state to workbook */
//   // const ws = XLSX.utils.aoa_to_sheet(data);
//   const ws = XLSX.utils.json_to_sheet(data);
//   ws['!cols'] = cols;
//   const wb = XLSX.utils.book_new();
//   XLSX.utils.book_append_sheet(wb, ws, 'SheetJS');
//   /* generate file and send to client */
//   XLSX.writeFile(wb, `${name}.xlsx`);
// }

function downloadBlob(blob: Blob, filename: string) {
  const a = document.createElement('a');
  document.body.appendChild(a);
  const url = window.URL.createObjectURL(blob);
  a.href = url;
  a.download = filename;
  a.click();
  setTimeout(() => {
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  }, 0)
}
