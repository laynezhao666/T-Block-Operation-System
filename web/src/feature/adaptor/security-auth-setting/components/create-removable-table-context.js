import { chainTableLayout } from '../../../component/tedge-components/table-layout-context/table-layout-context';

/**
 * 创建可移除行的表格布局上下文
 * 供 cards-table 和 doors-table 等权限组子表格复用
 * @param {Function} fetchData - 数据获取函数
 * @param {Function} removeRow - 行移除函数
 * @returns {Object} tableLayoutContext
 */
export function createRemovableTableContext(fetchData, removeRow) {
  return chainTableLayout(fetchData)
    .tableStyle({
      stripe: true,
    })
    .pagination()
    .localFilterPagination()
    .search({
      placeholder: '请输入关键字搜索',
    })
    .indexColumn({
      label: '序号',
    })
    .baseCurd({
      rowEditColumnWidth: 74,
      add: false,
      edit: false,
      remove: {
        adminRight: true,
        label: '移除',
        confirm: row => removeRow(row),
      },
    })
    .done();
}