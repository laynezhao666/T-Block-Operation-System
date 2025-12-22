import * as _ from "lodash";
import { forEachTreeNode } from "../../../../utils/tree";

export interface ICustomMenuItem {
  name: string;
  href: string;
  menuCode: string;
  parentMenuCode: string;
  icon: string;
}

export type ICustomMenuItemTreeData = ICustomMenuItem & {
  level: number;
  parent?: ICustomMenuItemTreeData,
  children: ICustomMenuItemTreeData[];
};

export interface IMenuItemData {
  n_id:         number;
  a_code:       string;
  base_url:     string;
  n_opcode:     string;
  n_level:      number;
  n_pid:        number;
  n_name:       string;
  n_href:       string;
  n_target:     string;
  n_licls:      string;
  n_order:      number;
  n_scope:      number;
  n_showtype:   boolean;
  n_createtime: string;
}

export const customMenuToTree = (menuItems: ICustomMenuItem[]): ICustomMenuItemTreeData[] => {
  const menuTreeDataItems: ICustomMenuItemTreeData[] = _.map(menuItems, item => ({
    ...item,
    level: 1,
    children: [],
  }));
  const menuTreeDataItemsMap: { [key: string]: ICustomMenuItemTreeData } = _.mapKeys(menuTreeDataItems, 'menuCode');

  menuTreeDataItems.forEach(item => {
    if (!item.parentMenuCode) {
      return;
    }

    const parentMenuItem = menuTreeDataItemsMap[item.parentMenuCode];
    if (!parentMenuItem) {
      return;
    }

    parentMenuItem.children.push(item);
    item.level = parentMenuItem.level + 1;
    item.parent = parentMenuItem;
  });

  return menuTreeDataItems.filter(item => !item.parentMenuCode);
}

export const fillMenuTreeData = (menuItemTreeData: ICustomMenuItemTreeData[]): ICustomMenuItemTreeData[] => {
  const rootMenuItemTreeData = _.cloneDeep(menuItemTreeData);
  (window as any).rootMenuItemTreeData = rootMenuItemTreeData

  forEachTreeNode(rootMenuItemTreeData, (menuItem: ICustomMenuItemTreeData, parentMenuItem: ICustomMenuItemTreeData | null, indexUnderParent: number) => {
    if (menuItem.children.length || menuItem.level === 3) return;

    menuItem.children.push({
      ...menuItem,
      menuCode: `${menuItem.menuCode}.child`,
      parentMenuCode: menuItem.menuCode,
      level: menuItem.level + 1,
      children: [],
    });
  });
(window as any).rootMenuItemTreeData = rootMenuItemTreeData;
  return rootMenuItemTreeData;
}

export const formatCustomMenu = (menuItems: ICustomMenuItem[]): IMenuItemData[] => {
  const menuTreeData = customMenuToTree(menuItems);

  const resultMenuItems: IMenuItemData[] = [];

  const menuItemIdMap = new Map<ICustomMenuItem, {
    id: number,
    level: number,
  }>();

  const fullMenuTreeData = fillMenuTreeData(menuTreeData);

  forEachTreeNode(fullMenuTreeData, (menuItem: ICustomMenuItemTreeData, parentMenuItem: ICustomMenuItemTreeData, indexUnderParent: number) => {
    const {
      id: parentId,
    } = menuItemIdMap.get(parentMenuItem) || {
      id: 1,
    };

    const id = parentId * 100 + (indexUnderParent + 1);

    resultMenuItems.push({
      n_id: id,
      a_code: 'tnebula',
      base_url: '',
      n_opcode: '',
      n_level: menuItem.level,
      n_pid: parentId,
      n_name: menuItem.name,
      n_href: menuItem.href,
      n_target: '_self',
      n_licls: menuItem.icon,
      n_order: 1,
      n_scope: 0,
      n_showtype: true,
      n_createtime: '',
    });

    menuItemIdMap.set(menuItem, { id, level: menuItem.level });
  });

  return resultMenuItems;
}
