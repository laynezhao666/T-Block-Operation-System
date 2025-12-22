import * as _ from "lodash";

type TreeNode<T extends Record<any, any>, ChildrenKey extends string = 'children'> = T & { [key in ChildrenKey]: TreeNode<T, ChildrenKey>[] };

export interface IIteratorOptions<T> {
  /** 节点字段，默认chldren */
  childrenField?: string;
  /** 获取子节点方法，优先级高于childrenField */
  getChildrenFunc?: (node: T) => T[];
  /** 获取子节点方法，优先级高于childrenField */
  setChildrenFunc?: <R>(node: R, children?: R[]) => void;

  /** 是否深度优先，默认为false */
  isLeafFirst?: boolean;
}

export type FIterator<T, R> = (node: T, parentNode: T | undefined | null, indexUnderParent: number, deep: number) => R;
export type FMapIterator<T, R> = (node: T, parentNode: T | undefined | null, children: R[] | null, indexUnderParent: number, deep: number) => R;
export type FFilterIterator<T> = (node: T, parentNode: T | undefined | null, children: T[] | null, indexUnderParent: number, deep: number) => boolean;

export const forEachTreeNode = <T>(rootNodeList: T[], func: FIterator<T, void>, opts: IIteratorOptions<T> = {}) => {
  const getChildren = opts.getChildrenFunc || (node => node[opts.childrenField || 'children']);
  const isLeafFirst = opts.isLeafFirst || false;

  const loop = (nodes: T[], parentNode: T | undefined | null, deep: number) => {
    (nodes || []).forEach((node, indexUnderParent) => {
      if (!isLeafFirst) {
        func(node, parentNode, indexUnderParent, deep);
      }

      const children = getChildren(node);

      if (children) {
        loop(children, node, deep + 1);
      }

      if (isLeafFirst) {
        func(node, parentNode, indexUnderParent, deep);
      }
    });
  };

  loop(rootNodeList, null, 1);
};

export const mapTreeNode = <T, R>(rootNodeList: T[], func: FMapIterator<T, R>, opts: IIteratorOptions<T> = {}): R[] => {
  const getChildren = opts.getChildrenFunc || (node => node[opts.childrenField || 'children']);
  const setChildren = opts.setChildrenFunc || ((node, children) => {
    node[opts.childrenField || 'children'] = children;
  });

  const loop = (nodes: T[], parentNode: T | undefined | null, deep: number) => {
    return (nodes || []).map((node, indexUnderParent): R => {
      const sourceChildren = getChildren(node);
      const children = sourceChildren ? loop(sourceChildren, node, deep + 1) : null;

      const newNode = func(node, parentNode, children, indexUnderParent, deep);
      setChildren(newNode, children);
      return newNode;
    })
  };

  return loop(rootNodeList, null, 1);
};

export const filterTreeNode = <T>(rootNodeList: T[], func: FFilterIterator<T>, opts: IIteratorOptions<T> = {}): T[] => {
  const getChildren = opts.getChildrenFunc || (node => node[opts.childrenField || 'children']);
  const setChildren = opts.setChildrenFunc || ((node, children) => {
    node[opts.childrenField || 'children'] = children;
  });

  const loop = (nodes: T[], parentNode: T | undefined | null, deep: number) => {
    const newNodes = [];

    nodes.forEach((node, indexUnderParent): boolean => {
      const sourceChildren = getChildren(node);
      const children = sourceChildren ? loop(sourceChildren, node, deep + 1) : sourceChildren;

      const isReserve = func(node, parentNode, children, indexUnderParent, deep);

      if (!isReserve) return;

      const newNode = {
        ...node,
      };

      setChildren(newNode, children);
      newNodes.push(newNode);
    });

    return newNodes;
  };

  return loop(rootNodeList, null, 1);
};

/** 计算深度，层级从1开始算，并且rootNodeList算第一层 */
export const computeDeep = <T>(rootNodeList: T[], opts: IIteratorOptions<T> = {}) => {
  let resultDeep = 0;
  forEachTreeNode(rootNodeList, (node, parentNode, indexUnderParent, deep) => {
    if (deep > resultDeep) {
      resultDeep = deep;
    }
  }, opts);

  return resultDeep;
};

export const arrayToTree = <T, K extends string = 'children'>(arr: T[], idField: string, parentField: string, newChildrenField: K) => {
  type TreeNode = T & { [key in K]: TreeNode[] };

  const nodes: TreeNode[] = _.map(_.cloneDeep(arr), item => {
    return {
      ...item,
      ...({
        [newChildrenField]: ([] as TreeNode[])
      } as { [key in K]: TreeNode[] }),
    };
  });

  const nodesMap = _.mapKeys(nodes, idField);

  nodes.forEach(item => {
    const parentId = item[parentField];
    const parentNode = nodesMap[parentId];

    if (_.isNil(parentId) || !parentNode) return;

    parentNode[newChildrenField].push(item);
  });

  return _.filter(nodes, n => _.isNil(n[parentField]));
};

export const getTreeNodes = <T>(rootNodeList: T[], opts?: IIteratorOptions<T>): T[] => {
  const nodes: T[] = [];

  forEachTreeNode(rootNodeList, (node) => {
    nodes.push(node);
  }, opts);

  return nodes;
};

export const flattenTreeNodes = getTreeNodes;

export default {};
