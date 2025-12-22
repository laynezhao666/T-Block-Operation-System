import * as _ from "lodash";

export interface ICachedItemContent {
  args: any[];
  result: any;
  ts: Number;
  clearTimeout?: ReturnType<typeof setTimeout>;
}

export interface ICachedItem<T extends Function> {
  func: T;
  duration: number;
  content: Array<ICachedItemContent>;
}

const cachedList: Array<ICachedItem<any>> = [];

/**
 * 缓存方法返回，可设置缓存时间，如果缓存被命中则重新计时，默认缓存10分钟
 * @param func 被缓存函数
 * @param duration 缓存时间，默认600000，即10分钟
 * @returns 被缓存的函数
 */
export const memoriedFunction = <T extends Function>(func: T, duration: number = 600000): T => {
  const cachedItem: ICachedItem<T> = {
    func,
    duration,
    content: [],
  };
  cachedList.push(cachedItem);

  return ((...args) => {
    const cachedContent = findContentItem(cachedItem, args);

    if (cachedContent) {
      setCachedTimeout(cachedItem, cachedContent);
      return cachedContent.result;
    }

    const result = func(...args);
    const content: ICachedItemContent = {
      args,
      result,
      ts: new Date().getTime(),
    }
    setCachedTimeout(cachedItem, content);
    cachedItem.content.push(content);

    return result;
  }) as any;
}

const setCachedTimeout = (cachedItem: ICachedItem<any>, content: ICachedItemContent) => {
  if (content.clearTimeout) {
    clearTimeout(content.clearTimeout);
  }

  content.clearTimeout = setTimeout(() => {
    const index = _.indexOf(cachedItem.content, content);
    if (index < 0) return;

    cachedItem.content.splice(index, 1);
    content.clearTimeout = undefined;
  }, cachedItem.duration);
};

const findContentItemIndex = (cachedItem: ICachedItem<any>, args: any[]) => {
  return _.findIndex(cachedItem.content, (contentItem) => {
    return contentItem.args.length === args.length && _.every(args, (arg, i) => arg === contentItem.args[i]);
  });
}

const findContentItem = (cachedItem: ICachedItem<any>, args: any[]): ICachedItemContent | undefined => {
  const index =  findContentItemIndex(cachedItem, args);
  return cachedItem.content[index];
}
