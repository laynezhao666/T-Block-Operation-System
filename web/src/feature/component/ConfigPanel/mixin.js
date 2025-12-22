import { isExtraTable, getExtraTable } from 'component/script/configHelper';
import { tables } from '@module/script/tables';

export default {
  methods: {
    isMK(field) {
      if (isExtraTable(this.table)) {
        return field === `${getExtraTable()}_id`;
      }
      return field === `${this.table}_id`;
    },
    isFK(field) {
      let { table } = this;
      if (isExtraTable(table)) {
        table = getExtraTable();
      }
      return field.endsWith('_id') && field !== `${table}_id`;
    },
    isKey(field) {
      if (this.table === 'itpartsasset' && field === 'uplayertype_id') return false;
      return field.endsWith('_id');
    },
    hasRights(opr) {
      const rights = tables[this.table]?.rights || 0b1000000;
      // eslint-disable-next-line max-len
      const oprName = {
        canExport: 0b0010000,
        canImport: 0b0001000,
        canDel: 0b0000100,
        canEdit: 0b0000010,
        canAdd: 0b0000001,
        showStatistics: 0b0100000,
        customizeBtn: 0b1000000,
      };
      return rights & oprName[opr];
    },
    getPath() {
      return tables[this.table]?.addPath;
    },
    getQuery() {
      return tables[this.table]?.addQuery || [];
    },
    parseEnum(enums) {
      return enums.map((enu) => {
        const arr = enu.split('|');
        if (arr.length === 1) {
          return {
            value: arr[0],
            label: arr[0],
          };
        }
        return {
          value: arr[0],
          label: arr[1],
        };
      });
    },
    // tableName() {
    //   return tables.tableName;
    // },
  },
};
