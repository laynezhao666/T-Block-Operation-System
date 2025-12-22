export const defaultOpts = {
  post: {
    restAxios: {},
  },
  get: {
    restAxios: {},
  },
  download: {
    fileName: '',
    method: 'post',
  },
}

export const defaultConfig = {
  needStrip: true,
  hasCode: true,
  isJson: false,
  extraParams: {},
  config: {
    timeout: 30000,
    headers: {
      common: {
        'X-Requested-With': 'XMLHttpRequest',
      },
    },
  },
}
