import {
  throttle,
} from 'lodash';

export type resizeHandler = (rect: DOMRect) => void;

export const bindResize = (el: Element, func: resizeHandler, throttleInterval: number = 100) => {
  const handle = throttle(() => {
    const rect = el.getClientRects()[0];
    func(rect);
  }, throttleInterval);

  const observer = new ResizeObserver(handle);
  observer.observe(el);

  const cancel = () => {
    observer.disconnect();
    observer.unobserve(el);
  };

  return cancel;
};

/**
 * 尽量放在 mounted 回调里，例如：
 * { mounted() {
 *  bindVueResize(this, this.handleResize.bind(this));
 * } }
*/
export const bindVueResize = (vm: Vue, func: resizeHandler, el?: Element, throttleInterval?: number) => {
  let cancelObserver: ReturnType<typeof bindResize>;

  vm.$once('hook:mounted', () => {
    cancelObserver = bindResize(el || vm.$el, func, throttleInterval || 100);
  });

  vm.$once('hook:destroyed', () => {
    cancelObserver();
  });
};
