/**
 * 安防记录模块 - 数据请求工具
 * 提供门禁控制器和门的数据获取方法
 */
import _ = require("lodash");
import { memoriedFunction } from "../../../../utils/memoried-function";
import getEdgeRequest from 'feature/utils/request';

/** 获取 axios 请求实例 */
const getAxios = () => getEdgeRequest((window.Vue.prototype as any).$axios);

/**
 * 获取所有控制器及其关联的门信息
 * @returns 包含 controls（控制器列表）和 doors（门列表）的对象
 */
export const fetchControlsAndDoors = async (): Promise<{
  controls: { [key: string]: any }[],
  doors: { [key: string]: any }[],
}> => {
  const controls = await getAxios().get('/api/dcos/tdac-cgi/controllers');

  // 从所有控制器中提取并扁平化门列表
  const doors = _.chain(controls)
    .map('doors')
    .flatten()
    .filter(Boolean)
    .value();

  return {
    controls,
    doors,
  };
}

/** 带缓存的 fetchControlsAndDoors，缓存时间 1000ms，避免短时间内重复请求 */
export const memoriedFetchControlsAndDoors = memoriedFunction(fetchControlsAndDoors, 1000);
