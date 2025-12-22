export class Timming {
  constructor(duration: number, onTrigger: Timming['onTrigger']) {
    this.duration = duration;
    this.onTrigger = onTrigger;
  }

  duration: number;
  times: number = 0;
  intervalHandle: ReturnType<typeof setInterval> | null = null;

  onTrigger: (times: number) => void;

  start() {
    if (this.intervalHandle) {
      this.stop();
    }

    this.times = 0;
    this.intervalHandle = setInterval(() => {
      this.onTrigger(++this.times);
    }, this.duration);
  }

  resume() {
    this.intervalHandle = setInterval(() => {
      this.onTrigger(++this.times);
    }, this.duration);
  }

  stop() {
    if (!this.intervalHandle) return;
    clearInterval(this.intervalHandle);
    this.intervalHandle = null;
  }
}