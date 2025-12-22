declare module '*.vue'

export { };

declare global {
  interface Window {
    tnwebServices: {
      loginStatusService: import('services/login-status.service').LoginStatusService;
      customConfigService: import('services/custom-config.service').CustomConfigService;
      v2DeviceNumberTransformerService: import('services/v2-device-number-transformer.service').V2DeviceNumberTransformerService;
      checkPointRealtimeDataService: import('services/tedge/check-point-realtime-data.service').CheckPointRealtimeDataService;
      pollingProxyAgentService: import('services/polling-request-proxy/polling-proxy.service').PollingProxyAgentService;
      changeApiMap: {};
      isTbos: true
    },
    changeApiMap: {};
  }
}
