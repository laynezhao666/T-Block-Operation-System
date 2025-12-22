import * as qs from 'qs';
import * as _ from 'lodash';
import { wrapPromiseLoading } from './loading';
import { doSomething, pipeline } from './fp';
import { selectFile } from './fp-dom';

// tnfusion的webpack打包有bug，只能从全局获取引用
const Vue = window.Vue;

const showErrorToast = (msg) => {
  (Vue.prototype as any).$message.error(msg);
};

const getAxios = () => (Vue.prototype as any).$axios;

export interface CustomAxiosOptions {
  /** 是否显示loading，默认true */
  loading?: boolean;
  /** 自动解析{ code,message,data }，错误则弹窗提示，否则返回data，默认true */
  autoResolveData?: boolean;
  /** headers */
  headers: Record<string, string>,
}

const defaultResolveData = async <T>(respOrPromise: any): Promise<T> => {
  const resp = await respOrPromise;
  const { data: respData } = resp;

  if (respData.code !== 0) {
    showErrorToast(respData.message || '网络错误');
    throw new Error(respData.message);
  };

  return respData.data;
};

const toastOnError = <T>(promise: T): T => {
  if (promise instanceof Promise) {
    promise.catch(err => {
      showErrorToast(err.message);
    });
  }

  return promise;
};

const wrapPromiseLoadingOrIdentityWhen = doSomething(wrapPromiseLoading)
  .else(_.identity);

const defaultResolveDataOrIdentityWhen = doSomething(defaultResolveData)
  .else(_.identity);

const wrapByAxiosOptions = <T extends (...args: any[]) => any>(doRequest: T, axiosOptions?: CustomAxiosOptions): ReturnType<T> => {
  const opts = {
    loading: true,
    autoResolveData: true,
    ...(axiosOptions || {}),
  };

  const result = pipeline(doRequest())
    .to(toastOnError)
    .to(
      defaultResolveDataOrIdentityWhen
        .when(() => opts.autoResolveData)
        .onBindThis(),
    ).to(
      wrapPromiseLoadingOrIdentityWhen
        .when(() => opts.loading)
        .onBindThis(),
    ).value();

  return result as any;

  // return pipeline(doRequest())
  //   .to(doIf.bind(null, defaultResolveData, opts.autoResolveData))
  //   .to(doIf.bind(null, wrapPromiseLoading, opts.loading))
  //   .value();
};

export const axiosPut = async <T>(
  url: string,
  data: { [key: string]: any } = null,
  axiosOptions?: CustomAxiosOptions,
): Promise<T> => {
  return wrapByAxiosOptions(() => getAxios().ins({
    url,
    method: 'PUT',
    data,
    headers: axiosOptions?.headers,
  }), axiosOptions);
};

export const axiosDelete = async <T>(
  url: string,
  bodyData: { [key: string]: any } = null,
  axiosOptions?: CustomAxiosOptions,
): Promise<T> => {
  return wrapByAxiosOptions(() => getAxios().ins({
    url: url,
    data: bodyData,
    method: 'DELETE',
    headers: axiosOptions?.headers,
  }), axiosOptions);
};

export const fileSelectSymbol = Symbol('axiosUploadFile.fileSelect');

export const axiosUploadFile = Object.assign(async <T>(
  url: string,
  formData: { [key: string]: (typeof fileSelectSymbol | string | number | boolean | File) } | FormData,
  axiosOptions?: CustomAxiosOptions,
): Promise<T> => {
  let formDataToPost: FormData = formData instanceof FormData
    ? formData
    : new FormData();

  if (!(formData instanceof FormData)) {
    await Promise.all(_.map(formData, async (v, k) => {
      if (v === fileSelectSymbol) {
        formDataToPost.append(k, await selectFile());
      } else {
        const vTmp = v instanceof File ? v : v.toString();
        formDataToPost.append(k, vTmp);
      }
    }));
  }

  return wrapByAxiosOptions(async () => {
    return getAxios().ins({
      url,
      method: 'POST',
      data: formDataToPost,
      processData: false,
      contentType: false,
      headers: axiosOptions?.headers,
    })
  }, axiosOptions);
}, {
  fileSelectSymbol,
});
