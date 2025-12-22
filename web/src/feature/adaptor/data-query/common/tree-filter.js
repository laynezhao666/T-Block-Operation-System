
// 匹配关键字，记录并汇聚匹配个数
function mapNode(parent, node, regs, options) {
  if (parent && parent.fullMatched) {
    node.fullMatched = true;
    return;
  }
  const regLen = regs.length;
  const matchedIndexs = node.matchedIndexs || (node.matchedIndexs = new Array(regLen));
  regs.forEach((reg, index) => {
    if (matchedIndexs[index] === undefined) {
      matchedIndexs[index] = options.matchAttrs.some(v => reg.test(node[v]));
    }
  });
  node.fullMatched = matchedIndexs.every(v => v);
  node.matchedIndexs = matchedIndexs;
}

// 预处理，匹配,汇聚，标记
function markNodes(parent, nodes, regs, prevMatchedIndexs, options) {
  const prevMatchedLen = prevMatchedIndexs.length;
  nodes.forEach((item) => {
    let newMatchedIndexs;
    if (prevMatchedLen && item.matchedIndexs) {
      newMatchedIndexs = new Array(regs.length);
      prevMatchedIndexs.forEach((v, i) => {
        newMatchedIndexs[i] = item.matchedIndexs[v];
      });
    }
    if (parent && parent.matchedIndexs) {
      newMatchedIndexs = newMatchedIndexs || new Array(regs.length);
      parent.matchedIndexs.forEach((v, index) => {
        if (v) {
          newMatchedIndexs[index] = v;
        }
      });
    }
    item.matchedIndexs = newMatchedIndexs;
    item.fullMatched = false;
    mapNode(parent, item, regs, options);
    const children = item[options.children];
    if (children && children.length > 0) {
      markNodes(item, children, regs, prevMatchedIndexs, options);
    }
  });
}
// 过滤节点
function filterNodes(parent, nodes, options) {
  const childrenName = options.children;
  nodes.forEach((item) => {
    if (item.fullMatched) { // 直接放到树中，其下面的节点均不用处理
      parent[childrenName].push(item);
    } else {
      const children = item[childrenName];
      if (children && children.length > 0) {
        const nItem = {
          [childrenName]: [],
        };
        options.allAttrs.forEach((v) => {
          nItem[v] = item[v];
        });
        filterNodes(nItem, children, options);
        if (nItem[childrenName].length > 0) {
          parent[childrenName].push(nItem);
        }
      }
    }
  });
};

// 每个节点下新增属性：fullMatched（布尔值，true表示到这层级的该节点已经全部匹配），matchedIndexs
// 外部业务可依赖fullMatched进行节点是否显示
// matchedIndexs 作为搜索内部使用，具有不确定性
export default function getFilter(data, options) {
  let matchedWords = [];
  const _options = {
    // 关键字转换为对象实例,提供test方法
    getTester: word => new RegExp(word, 'i'),
    // 子节点数组名称
    children: 'children',
    // 需要匹配的属性值
    matchAttrs: ['name'],
    // 是否需要过滤节点，返回新的数组
    isFilterNode: true,
  };
  Object.assign(_options, options);
  if (!data || !Array.isArray(data) || !data.length) {
    throw (new Error('data must be Array, and must have at least 1 element.'));
  }
  if (_options.isFilterNode) {
    _options.allAttrs = [];
    const keys = Object.keys(data[0]);
    keys.forEach((v) => {
      if (v !== _options.children) {
        _options.allAttrs.push(v);
      }
    });
  }

  return (words) => {
    // 关键词转为正则
    const regs = [];
    const arrWords = words.toLowerCase().split(' ');// 转换为数组
    const distintWords = [];// 去除重复及空格
    const finalWords = [];// 与上一次关键字对比，重新调整顺序
    const prevMatchedIndexs = [];// 上一次匹配中，可用的索引
    // 非空，唯一
    arrWords.forEach((v) => {
      if (v && !distintWords.includes(v)) {
        distintWords.push(v);
      }
    });
    // 与上一次关键字做对比，找到可用索引，重新调整新关键字顺序
    if (matchedWords.length > 0) {
      for (let i = 0; i < matchedWords.length; i++) {
        if (distintWords.includes(matchedWords[i])) {
          prevMatchedIndexs.push(i);
          finalWords.push(matchedWords[i]);
        }
      }
    }
    distintWords.forEach((v) => {
      if (!finalWords.includes(v)) {
        finalWords.push(v);
      }
    });
    // 转换为匹配方法
    finalWords.forEach((v) => {
      regs.push(_options.getTester(v));
    });
    // 存储这次的关键词
    matchedWords = finalWords;

    if (regs.length === 0) {
      return data;
    }
    // 匹配并标记节点
    markNodes(null, data, regs, prevMatchedIndexs, _options);
    // 过滤节点
    if (_options.isFilterNode) {
      const ntree = { children: [] };
      filterNodes(ntree, data, _options);
      return ntree.children;
    }
    return data;
  };
}
