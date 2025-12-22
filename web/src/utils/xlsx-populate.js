
/**
 *
 * @param {File | Blob} file
 */
export const parseXlsx = async (file) => {
  // const fileReader = new FileReader();
  // fileReader.readAsArrayBuffer()
  const XlsxPopulate = (await import('xlsx-populate/browser/xlsx-populate-no-encryption')).default;

  return XlsxPopulate.fromDataAsync(file)
    .then(workbook => workbook);
};

export const blankXlsx = async () => {
  // const fileReader = new FileReader();
  // fileReader.readAsArrayBuffer()
  const XlsxPopulate = (await import('xlsx-populate/browser/xlsx-populate-no-encryption')).default;

  return XlsxPopulate.fromBlankAsync()
    .then(workbook => workbook);
};
