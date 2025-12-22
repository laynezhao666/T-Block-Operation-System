import * as _ from 'lodash';
import Vue from 'vue';

type VmWithTimeout = Vue & {
  __timeoutMap__?: Map<string, ReturnType<typeof setTimeout>>;
  $timeout(key: string, func: () => void, ms: number): typeof setTimeout;
  $clearTimeout(key: string);
};

/** 因为框架有bug，不得不这样 */
const globalVue: typeof Vue = (window as any).Vue;

const timeoutUnmountCallback = function (this: VmWithTimeout) {
  const vm: VmWithTimeout = this;
  if (!vm.__timeoutMap__) return;

  vm.__timeoutMap__.forEach(t => {
    clearTimeout(t);
  });
};

/** 同setTimeout，但是会自动随组件卸载销毁定时，当key重复时，销毁之前的定时器 */
globalVue.prototype.$timeout = function(key: string, func: () => void, ms: number) {
  if (!key) throw new Error('定时器key不能为空');

  const vm: VmWithTimeout = this as any;

  if (!vm.__timeoutMap__) {
    vm.__timeoutMap__ = new Map();
    this.$once('hook:beforeDestroy', timeoutUnmountCallback);
  }

  const oldT = vm.__timeoutMap__.get(key);

  if (oldT) {
    clearTimeout(oldT);
  }

  const t = setTimeout(() => {
    func();
    this.$clearTimeout(key);
  }, ms);

  vm.__timeoutMap__.set(key, t);
  return t;
};

globalVue.prototype.$clearTimeout = function(key: string) {
  const vm: VmWithTimeout = this as any;
  const t = vm.__timeoutMap__?.get(key);
  if (!t) return;
  clearTimeout(t);
  vm.__timeoutMap__?.delete(key);
};

type VmWithInterval = Vue & {
  __intervalMap__?: Map<string, ReturnType<typeof setInterval>>;
  $interval(key: string, func: () => void, ms: number): ReturnType<typeof setTimeout>;
  $clearInterval(key: string): void;
};

// interval
const intervalUnmountCallback = function (this: VmWithInterval) {
  const vm: VmWithInterval = this;
  if (!vm.__intervalMap__) return;

  vm.__intervalMap__.forEach(t => {
    clearInterval(t);
  });
}

/** 同setInterval，但是会自动随组件卸载销毁定时，当key重复时，销毁之前的定时器 */
globalVue.prototype.$interval = function(this: VmWithInterval, key: string, func: () => void, ms: number) {
  if (!key) throw new Error('定时器key不能为空');

  const vm: VmWithInterval = this;

  if (!vm.__intervalMap__) {
    vm.__intervalMap__ = new Map();
    this.$once('hook:beforeDestroy', intervalUnmountCallback);
  }

  const oldT = vm.__intervalMap__.get(key);

  if (oldT) {
    clearInterval(oldT);
  }

  const t = setInterval(func, ms);

  vm.__intervalMap__.set(key, t);

  return t;
};

globalVue.prototype.$clearInterval = function(this: VmWithInterval, key: string) {
  const vm: VmWithInterval = this;
  const t = vm.__intervalMap__?.get(key);
  if (!t) return;
  clearInterval(t);
  vm.__intervalMap__?.delete(key);
}

// IntervalPromise
type VmWithIntervalPromise =  & {
  __intervalPromiseMap__?: Map<string, string>;
  $intervalPromise(key: string, func: () => any, ms: number): typeof setTimeout;
  $clearIntervalPromise(key: string): void;
};

globalVue.prototype.$intervalPromise = function(this: VmWithIntervalPromise & VmWithTimeout, key: string, func: () => any, ms: number) {
  if (!key) throw new Error('定时器key不能为空');

  const timeoutKey = `__intervalPromise__${key}`;

  const vm = this;

  if (!vm.__intervalPromiseMap__) {
    vm.__intervalPromiseMap__ = new Map();
    // 由于使用的$timeout，组件销毁时的清理任务由$timeout来自动清理
  }

  if (vm.__intervalPromiseMap__.get(key)) {
    (this as any).$clearTimeout(key);
  }

  vm.__intervalPromiseMap__.set(key, timeoutKey);

  const runTimeout = () => {
    return vm.$timeout(timeoutKey, async () => {
      await func();
      runTimeout();
    }, ms);
  };

  return runTimeout();
}

globalVue.prototype.$clearIntervalPromise = function(this: VmWithIntervalPromise & VmWithTimeout, key: string) {
  const vm = this;

  const timeoutKey = vm.__intervalPromiseMap__?.get(key);
  if (!timeoutKey) return;
  vm.$clearTimeout(timeoutKey);
  vm.__intervalPromiseMap__?.delete(key);
}

/**
 * 将函数封装为轮询任务，如果每次执行后都会重置计时，包括非定时触发的也会重置计时
 * @param isAwaitPromise boolean 默认false，为true时等待上一次轮询Promise完成才开始下一轮计时
 */
globalVue.prototype.$intervalFunction = function (func: (...args: any[]) => any, ms: number, isAwaitPromise: boolean = false) {
  const key = Math.random().toString(32);

  const resultFunc = (...args: any[]) => {
    return func.call(this, ...args);
  };

  const startTiming = () => {
    if (isAwaitPromise) {
      (this as unknown as VmWithIntervalPromise).$intervalPromise(key, resultFunc, ms);
    } else {
      (this as unknown as VmWithInterval).$interval(key, resultFunc, ms);
    }
  }

  const stopTiming = () => {
    if (isAwaitPromise) {
      (this as unknown as VmWithIntervalPromise).$clearIntervalPromise(key);
    } else {
      (this as unknown as VmWithInterval).$clearInterval(key);
    }
  }

  startTiming();

  return (...args: any[]) => {
    // 外部调用时重置计时
    stopTiming();

    const result = resultFunc.call(this, ...args);

    if (isAwaitPromise && result instanceof Promise) {
      result.then(() => {
        startTiming();
      });
    } else {
      startTiming();
    }

    return result;
  }
};
