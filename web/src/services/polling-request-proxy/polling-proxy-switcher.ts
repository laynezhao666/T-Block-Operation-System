import { ClientProxy, PollingProxyAgentService } from "./polling-proxy.service"
import { RequestConfig } from "./request-config";

export class PollingProxySwitcher {
  constructor(opts: { interval: number }) {
    this.interval = opts.interval;
  }

  interval: number;
  agent: PollingProxyAgentService = window.tnwebServices.pollingProxyAgentService;

  lastRequestConfigJson: string = '';
  lastClientProxy: ClientProxy<any> | null = null;

  onData: (data: any) => void;

  proxy(requestConfig: RequestConfig, onData: typeof this.onData) {
    if (!requestConfig) return;
    const { agent } = this;

    const requestConfigJson = JSON.stringify(requestConfig);
    if (requestConfigJson === this.lastRequestConfigJson) return;

    this.cancel();

    this.lastClientProxy = agent.proxy({
      interval: this.interval,
      request: requestConfig,
    }, onData);
  }

  cancel() {
    if (!this.lastClientProxy) return;
    this.agent.exit([this.lastClientProxy]);
  }
}
