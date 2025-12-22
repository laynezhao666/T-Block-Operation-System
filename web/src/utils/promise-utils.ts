import * as _ from 'lodash';

export const delayPromise = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export type FPromiseIterator<T = any, K = number | string> = (value: T, key: K) => Promise<any>;

export const forEachPromise = (listOrObject, func: FPromiseIterator) => {
  return Promise.all(_.map(listOrObject, func));
};

export const filterPromise = async (listOrObject, func: FPromiseIterator) => {
  const isArray = listOrObject instanceof Array;
  const result = isArray ? [] : {};
  const resultKeys: Array<number | string> = [];

  await forEachPromise(listOrObject, async (value, keyOrIndex) => {
    if (!await (func(value, keyOrIndex))) return;

    resultKeys.push(keyOrIndex);
  });

  const resultObj = _.pick(listOrObject, resultKeys);

  return isArray ? _.values(resultObj) : result;
};

export const mapPromise = async (listOrObject, func: FPromiseIterator) => {
  const result: Array<number | string> = [];

  await forEachPromise(listOrObject, async (value, keyOrIndex) => {
    result[keyOrIndex] = await func(value, keyOrIndex);
  });

  return result;
};

export function seriesPromise<T>(list: Array<T>, func: FPromiseIterator<T, number>): Promise<any>;
export function seriesPromise<T>(obj: { [key: string]: T }, func: FPromiseIterator<T, string>): Promise<any>;
export async function seriesPromise<T>(listOrObject: Array<T> | { [key: string]: T }, func: FPromiseIterator<T, any>) {
  const keys = _.keys(listOrObject);
  const results: any[] | { [key: string]: any } = _.isArrayLike(listOrObject) ? [] : {};

  for (let i = 0; i < keys.length; i++) {
    const key = keys[i];
    const value = listOrObject[key];

    results[key] = await func(value, key);
  }

  return results;
};

export const chunkSeriesPromise = async (listOrObject, chunkSize: number, func: FPromiseIterator) => {
  const keys = _.keys(listOrObject);
  const results: any[] | { [key: string]: any } = _.isArrayLike(listOrObject) ? [] : {};

  const keyChunks = _.chunk(keys, chunkSize);

  for (let chunkIndex = 0; chunkIndex < keyChunks.length; chunkIndex++) {
    const chunk = keyChunks[chunkIndex];

    await Promise.all(_.map(chunk, async (key) => {
      const value = listOrObject[key];
      results[key] = await func(value, key);
    }));
  }

  return results;
};

export default {
  delayPromise,
  forEachPromise,
  filterPromise,
  mapPromise,
  seriesPromise,
};
