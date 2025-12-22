import * as _ from 'lodash';

export const pipeline = <T>(data: T):FpPipeline<T> => {
  return {
    _value: data,
    value: () => data,
    to: (fn) => pipeline(fn(data)),
  };
};

export interface FpPipeline<T> {
  _value: T;
  value: () => T;
  to: <F extends (...args: any[]) => any>(fn: F) => FpPipeline<ReturnType<F>>;
}

// test
// console.log(pipeline([1,2,3,4,5])
//   .to((arr) => arr.map(item => item + 1))
//   .to(arr => arr.reduce((count, item) => count + item, 0))
//   .value()
// );

export interface DoSomething<T, R> {
  doFn: (target: T) => R;
  whenFn?: (target: T) => boolean;
  on?: T;

  result: () => R;
}

class DoSomethingImp<T, DF extends (target: T) => any, EF extends ((target: T) => any) = ((target: T) => undefined)> {
  _doFn: DF;
  _whenFn?: (target: T) => boolean;
  _elseFn?: EF;
  target: T;

  do(fn: DF) {
    this._doFn = fn;
    return this;
  }

  when(fn: (target: T) => boolean) {
    this._whenFn = fn;
    return this;
  }

  else<EF extends (target: T) => any>(fn: EF) {
    this._elseFn = fn as any;
    return this as unknown as DoSomethingImp<T, DF, EF>;
  }

  on<TT>(target: TT): ReturnType<DF> | ReturnType<EF> {
    // this.target = target as unknown as T;
    // return this as unknown as DoSomethingImp<TT, (target: TT) => ReturnType<DF>, (target: TT) => ReturnType<EF>>;

    return !this._whenFn || this._whenFn(target as any)
      ? this._doFn(target as any)
      : this._elseFn(target as any);
  }

  onBindThis() {
    const t = <TT>(target: TT): ReturnType<DF> | ReturnType<EF> => {
      return this.on(target);
    };
    return t;
  }
}

export const doSomething = <T, DF extends (target: T) => any>(doFn: DF) => {
  return new DoSomethingImp<T, DF, ((target: T) => undefined)>().do(doFn);
};

export const doIf = <P, T extends (p: P) => any>(toDo: T, when: (boolean | (() => boolean)), p: P): ReturnType<T> => {
    return wrapIdentityIfNotFunction(when)() ? toDo(p) : p;
  };

export const isBooleanTrue = (data: any): data is true => {
  return data === true;
}

export const wrapIdentity = <T>(data: T) => (): T => data;

export const wrapIdentityIfNotFunction = <T>(data: T): (T extends Function ? T : (() => T)) => {
  return typeof data === 'function' ? (data as T) : (wrapIdentity(data) as any);
}

export const condition = <T>() => {
  let resultValue = undefined;
  let done = false;
  return {
    if(isTrue: boolean, value: T): T {
      if (!done && isTrue) {
        resultValue = value;
        done = true;
      }
      return this;
    },
    elseif(isTrue: boolean, value: T): T {
      return this.if(isTrue, value);
    },
    else(value: T): T {
      return done ? resultValue : value;
    },
  };
};
