
// tnfusion的webpack打包有bug，只能从全局获取引用
const showLoading = (): { close: () => void } => {
  return (window.Vue.prototype as any).$loading();
}

export interface FWrapPromiseLoading<T> {
  async (promise: Promise<T>): Promise<T>
}

export const wrapPromiseLoading = async <T>(promise: Promise<T>): Promise<T> => {
  const loading = showLoading();

  return promise.finally(() => {
    loading.close();
  });
};
