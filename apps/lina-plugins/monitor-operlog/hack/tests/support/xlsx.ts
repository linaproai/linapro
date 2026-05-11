import * as XLSX from 'xlsx';

const xlsxModule = XLSX as typeof XLSX & { default?: typeof XLSX };

function requireXlsxRead(): typeof XLSX.read {
  const read = xlsxModule.read ?? xlsxModule.default?.read;
  if (!read) {
    throw new Error('xlsx read API is unavailable');
  }
  return read;
}

function requireXlsxUtils(): typeof XLSX.utils {
  const utils = xlsxModule.utils ?? xlsxModule.default?.utils;
  if (!utils) {
    throw new Error('xlsx utils API is unavailable');
  }
  return utils;
}

export const xlsxRead = requireXlsxRead();
export const xlsxUtils = requireXlsxUtils();
