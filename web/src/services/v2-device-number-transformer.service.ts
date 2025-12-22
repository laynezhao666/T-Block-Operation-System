import * as _ from 'lodash';
import * as xlsx from 'xlsx';

export const FILE_CONTENT_SPLITTER = '$%__file_content_input__%$';

export class V2DeviceNumberTransformerService {
  mapping: {
    [key: string]: string;
  } = {};

  get(oldDeviceNumber: string, isFallback = true) {
    const newDeviceNumber = this.mapping[oldDeviceNumber];

    return !newDeviceNumber && isFallback ? oldDeviceNumber : newDeviceNumber;
  }

  parseCustomConfigContent(content: string) {
    if (!content) return;
    const [, base64] = content.split(FILE_CONTENT_SPLITTER);
    const base64DataPart = base64.replace(/^.+;base64,/, '').replace(/_/g, "/").replace(/-/g, "+");
    const workbook = xlsx.read(base64DataPart, { type: 'base64' });

    const mapping = this.readFromXlsxWorkbook(workbook);
    this.mapping = mapping;
  }

  private readFromXlsxWorkbook(workbook: xlsx.WorkBook) {
    return _.reduce(workbook.SheetNames, (resultMap, sheetName) => {
      return {
        ...resultMap,
        ...this.readFromXlsxSheet(workbook.Sheets[sheetName]),
      };
    }, {});
  }

  private readFromXlsxSheet(sheet: xlsx.WorkSheet) {
    const {
      start: startPos,
      end: endPos,
    } = resolveXlsxRefRanges(sheet['!ref']);

    const colNoList = _.range(startPos.colNo.charCodeAt(0), endPos.colNo.charCodeAt(0) + 1)
      .map(charCode => String.fromCharCode(charCode));

    let oldDeviceNumberColNo: string | undefined;
    let newDeviceNumberColNo: string | undefined;

    _.forEach(colNoList, colNo => {
      const cellValue = sheet[`${colNo}1`]?.h;
      if (cellValue === '设备编号') {
        oldDeviceNumberColNo = colNo;
      } else if (cellValue === '设备编号（新）') {
        newDeviceNumberColNo = colNo;
      }
    });

    if (!oldDeviceNumberColNo || !newDeviceNumberColNo) return {};

    return _.chain(_.range(startPos.rowIndex + 1, endPos.rowIndex + 1))
      .map((rowIndex) => ([sheet[`${oldDeviceNumberColNo}${rowIndex}`]?.h, sheet[`${newDeviceNumberColNo}${rowIndex}`]?.h]))
      .fromPairs()
      .value();
  }
}

const resolveXlsxRefRanges = (ref: string) => {
  const [start, end] = ref.split(':')
    .map(item => {
      const [colNo, ...rowIndexArr] = item.split(/(?=\d)/);
      return { colNo, rowIndex: Number(rowIndexArr.join('')) };
    });

  return {
    start,
    end,
  };
};
