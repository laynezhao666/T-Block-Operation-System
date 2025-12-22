import * as dayjs from 'dayjs';

// 检测间隔时间
const TICK_INTERVAL_GAP_MS = 1000;

let tickingInterval: ReturnType<typeof setInterval> | null = null;

let day: number | null = null;

const handlesSet = new Set<Function>();

const getDay = () => {
  // 以周一到周六来记录，正常时间，过一天必然变化；主动修改系统时间会失效，这里不考虑这种特殊情况（如果要更精准可以用格式化，但是更耗费性能）
  return dayjs().day();
}

const tick = () => {
  const newDay = getDay();
  if (newDay === day) return;

  handlesSet.forEach(fn => {
    fn();
  });
};

const startTicking = () => {
  day = getDay();
  tickingInterval = setInterval(tick, TICK_INTERVAL_GAP_MS);
};

const stopTicking = () => {
  clearInterval(tickingInterval);
  tickingInterval = null;
};

export const listen = <T extends Function>(fn: T) => {
  handlesSet.add(fn);

  if (!tickingInterval) {
    startTicking();
  }
};

export const unlisten = <T extends Function>(fn: T) => {
  handlesSet.delete(fn);
  if (tickingInterval) {
    stopTicking();
  }
};

export const listenInVue = <T extends Function>(vm: Vue, fn: T) => {
  const callVmFn = () => fn.call(vm);

  vm.$once('hook:beforeDestroy', function() {
    unlisten(callVmFn);
  })

  handlesSet.add(callVmFn);

  if (!tickingInterval) {
    startTicking();
  }
};
