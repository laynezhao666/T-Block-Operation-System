import { map, isPlainObject, isFunction, merge, find } from 'lodash';
import config from '@module/script/config';
import { tables } from '@module/script/tables';

function length(text) {
  const reg = /[\x21-\x7E]/g;
  const match = text.match(reg);
  if (match) {
    // eslint-disable-next-line no-mixed-operators
    return text.length - match.length / 2;
  }
  return text.length;
}

function calcWidth(label, { count = 0, type = 'text' }) {
  const padding = 24 + 24;
  const border = 1;
  let body;
  if (type === 'date') {
    body = 72;
  } else if (type === 'num') {
    body = count * 12 / 2;
  } else if (type === 'char') {
    body = count * 12 / 1.5;
  } else {
    body = count * 12;
  }
  const header = length(label) * 14;
  const max = Math.max(header, body);
  return max + padding + border;
}

/**
 * @author 
 * @param {} config/*.js的内容
 * @param {*} config对应的表
 * @param {*} 当前tab的表
 * @returns
 */
export function handleConfig(config, curTable = '', table = '') {
  return config.map((item) => {
    let { label } = item;
    if (isFunction(label)) {
      label = label(table);
    }
    if (isFunction(item.jumpQuery)) {
      // eslint-disable-next-line no-param-reassign
      item.jumpQuery = item.jumpQuery(table);
    }
    return {
      ...item,
      label,
      fixed: curTable === table ? item.fixed : false,
      show: Array.isArray(item.show) ? item.show.indexOf(table) > -1 : item.show,
      placeholder: item.placeholder || '',
      showInTabelSetting: item.showInTabelSetting ? item.showInTabelSetting.includes(table) : true,
      width: calcWidth(label, isPlainObject(item.size) ? item.size : { count: item.size }),
      showContent: Array.isArray(item.show) ? item.show.indexOf(table) > -1 : item.show,
      showIndex: Array.isArray(item.isFilter) ? item.isFilter.includes(table) : item.isFilter,
    };
  });
}

function handleExtraConfig(table, remoteConfig) {
  return remoteConfig.filter(item => item.name.startsWith(`${table}@`)).map(({ name, memo: label }) => {
    const match = find(config.privates, {
      name,
    });
    const field = {
      name,
      show: true,
      label,
    };
    if (match) {
      merge(field, match);
      field.width = calcWidth(label, isPlainObject(match.size) ? match.size : { count: match.size });
    } else {
      field.width = 200;
    }
    return field;
  });
}

const FNMap = {
  // columns_deviceCode: 'device_number', // 用不到，先注释了
  // uPos_svrAssetId: 'svr_assetId',
};

export function getSelectNamespace(field, self = true) {
  const names = map(tables, table => table.name);
  if (names.includes(field) && self) {
    return field;
  } if (field.endsWith('_id')) {
    return getNameById(field);
  }
  return FNMap[field];
}

export function getNameById(id) {
  if (tables.tableName === 'resourceChat') {
    if (id === 'bizArea_id') {
      return id;
    }
  }
  const table = id.replace('_id', '');
  return tables[table] ?.name || `${table}_name`;
}

export function getTextByTable(table) {
  return tables[table] ?.text;
}

export function isExtraTable(table) {
  return table.startsWith('device_');
}

export function getExtraTable() {
  return 'device';
}

export function getRemoteTable(table, remoteTables) {
  if (isExtraTable(table)) {
    if (remoteTables.indexOf(table) > -1) {
      return table;
    }
    return getExtraTable();
  }
  return table;
}

export function createTableConfig(table, remoteTables, remoteTableConfig) {
  let supTable = table;
  if (supTable.startsWith('device_')) {
    supTable = 'device';
  }
  const tableOrder = Object.keys(tables);
  const index = getWeight(supTable);

  let i = 0;
  const rst = {};
  while (i <= index) {
    const curTable = tableOrder[i];
    if (remoteTables.indexOf(curTable) > -1 && curTable !== supTable) {
      rst[curTable] = handleConfig(config[curTable], curTable, supTable);
    }
    i = i + 1;
  }
  // 表是具备私有字段的设备表
  if (isExtraTable(table) && remoteTables.indexOf(table) > -1) {
    // eslint-disable-next-line max-len
    rst[table] = [...handleConfig(config[supTable], supTable, supTable), ...handleExtraConfig(table, remoteTableConfig)];
  } else {
    rst[supTable] = handleConfig(config[supTable], supTable, supTable);
  }
  return rst;
}

function getWeight(name) {
  const curTable = tables[name];
  return curTable.weight || Object.keys(tables).indexOf(name);
}
