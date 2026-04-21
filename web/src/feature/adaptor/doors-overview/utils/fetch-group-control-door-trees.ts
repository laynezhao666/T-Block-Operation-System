/**
 * 门禁总览模块 - 树形数据构建工具
 * 负责获取控制器和分组数据，并构建树形结构供页面展示
 */
import _ = require("lodash");
import getEdgeRequest from 'feature/utils/request';

/** 获取 axios 请求实例 */
const getAxios = () => getEdgeRequest((window.Vue.prototype as any).$axios);

/**
 * 加载控制器和分组的树形数据
 * @returns controlsTree - 按控制器组织的树（控制器 -> 门）
 * @returns groupsTree - 按分组组织的树（分组 -> 门）
 */
export const loadControlAndGroupsTree = async () => {
  // 并行请求控制器列表和分组列表
  const [controls, groups] = await Promise.all([
    getAxios().get('/api/dcos/tdac-cgi/controllers'),
    getAxios().get('/api/dcos/tdac-cgi/groups'),
  ]);

  // 构建控制器树，为每个控制器和门添加类型标识和唯一 nodeKey
  const controlsTree = _.map(controls as any[], control => ({
    ...control,
    type: 'control',
    nodeKey: `control-${control.id}`,
    doors: _.map(control.doors, door => ({
      ...door,
      nodeKey: `door-${door.id}`,
      type: 'door',
      controlId: control.id,
      controlName: control.name,
    })),
  }));

  // 将所有门按 group_id 分组，用于构建分组树
  const doorsGroupsMap = _.chain(controlsTree)
    .map('doors')
    .flatten()
    .groupBy('group_id')
    .value();

  // 构建分组树，将对应的门挂载到各分组下
  const groupsTree = _.map(groups, group => ({
    nodeKey: `group-${group.id}`,
    type: 'group',
    ...group,
    name: group.name,
    doors: doorsGroupsMap[group.id] || [],
  }));

  return {
    controlsTree,
    groupsTree,
  };
}
