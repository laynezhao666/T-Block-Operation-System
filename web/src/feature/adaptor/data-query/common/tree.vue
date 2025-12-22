<template>
  <!-- 树状目录公用组件 -->
  <div class="c-tree">
    <div v-if="showDeviceSearch">
      <div class="c-tree-title">
        设备列表
      </div>
      <div class="c-tree-tools">
        <el-input
          v-model="filterText"
          class="c-tree-tools__input"
          placeholder="搜索设备"
          prefix-icon="tn-icon-search"
          clearable
          border-type="no-border"
        />
      </div>
    </div>
    <div
      v-loading="loading"
      class="c-tree-wrap"
    >
      <el-tree
        v-if="treeData"
        :key="treeVersion"
        ref="tree"
        highlight-current
        node-key="id"
        :data="treeData"
        :props="defaultProps"
        :filter-node-method="filterNode"
        :current-node-key="currentNodekey"
        :default-expanded-keys="expandedkeys"
        :render-content="renderContent"
        :expand-on-click-node="false"
        :height="749"
        @node-click="handleNodeClick"
      >
        <template v-slot="{ node }">
          <span class="c-tree-node">
            <span
              class="c-tree-node__label"
              :title="node.label"
            >
              {{ node.label }}
            </span>
          </span>
        </template>
      </el-tree>
    </div>
  </div>
</template>

<script>
import { dataQuery as cgi } from '@@/config/cgi';
import getFilter from './tree-filter';
import { debounce, cloneDeep, uniq, has } from 'lodash';
import { getQueryString } from 'common/script/utils.js';
import getEdgeRequest from '../../../utils/request';
import business from '@@/config/business';
import eventBus from '../eventBus';
import { forEachTreeNode } from '../../../../utils/tree';
import { DeviceTreeService } from 'services/tedge/device-tree.service.ts';

export default {
  props: {
    // TODO 链接从外部传入
    urls: Object,
    mozuloaded: Boolean,
    mozuId: Number,
    treeOption: {
      type: String,
      default: 'all',
    },
    dataSourceType: {
      type: String,
      default: 'byBiz',
    },
    treeFilterText: {
      type: String,
      default: '',
    },
    showDeviceSearch: {
      type: Boolean,
      default: true,
    },
    treeDataProp: {
      type: Array,
      default: () => [],
    },
    warningDeviceGidList: {
      type: Array,
      default: () => [],
    },
    treeType: {
      type: String,
      default: '',
    },
    enableDeviceNumberV2: {
      type: String,
      default() {
        return '0';
      },
    },
  },
  data() {
    const defaultDevId = getQueryString('devId');
    const defaultDevName = getQueryString('devName');

    return {
      defaultDevId,
      defaultDevName,
      loading: false,
      cgiMap: {
        byBiz: 'getBizDeviceLevelTree',
        byCollect: 'getCollectDeviceTree',
      },
      currentNode: {},
      filterText: '',
      filterTextRegs: null,
      treeData: [],
      orgTreeData: [],
      defaultProps: {
        children: 'children',
        label: 'name',
      },
      currentNodekey: '',
      expandedkeys: [],
      nodePathArray: [],
      // allAlarmGids: [],

      treeVersion: 0,
    };
  },
  computed: {
    wordsList() {
      return this.filterText
        .toLowerCase()
        .trim()
        .split(/\s+/);
    },
    wordsRegexp() {
      return this.wordsList.join('');
    },
  },
  watch: {
    treeData: {
      immediate: true,
      handler() {
        this.treeVersion += 1;
      },
    },
    treeOption() {
      this.getData();
    },
    treeFilterText(val) {
      if (this.filterText?.trim() === val?.trim()) return;
      this.filterText = val;
      // this.$refs.tree.filter(val);
      // this.filterData();
    },
    filterText() {
      this.filterData();
    },
    currentNode(data) {
      this.$refs.tree.setCurrentKey(data.id);
      this.$emit('node-click', data, this.nodePathArray);
    },
    mozuId() {
      this.getData();
    },
    treeDataProp: {
      handler(val) {
        if (!val.length) return;
        this.treeData = val;
        const data = this.treeData[0].children[0];

        this.$nextTick(() => {
          this.expandedkeys = [data.id];
          this.currentNodekey = data.id;
          this.currentNode = data;
        });
      },
      deep: true,
      immediate: false,
    },
    // warningDeviceGidList() {
    //   if (this.treeData && !this.allAlarmGids.length) {
    //     // TODO getGidPath
    //     this.allAlarmGids = this.getGidPath(this.treeData, this.warningDeviceGidList);
    //   }
    // },
  },
  async created() {
    this.filterData = debounce(this.filterData, 200, {
      leading: false,
      trailing: true,
    });
    this.getData();
    // setTimeout(() => {
    //   this.treeForeach(this.treeData);
    //   console.log(this.treeData);
    // }, 2000);
  },
  methods: {
    treeForeach(tree) {
      let node = {};
      const list = tree;
      // eslint-disable-next-line no-cond-assign
      while (node = list.shift()) {
        node.children && list.push(...node.children);
      }
    },
    getData() {
      if (business.showModuleSelected && !this.mozuloaded) {
        return;
      }

      this.loading = true;
      const cgiUrl = this.cgiMap[this.dataSourceType];

      let actualUrl = cgi[cgiUrl];
      if (this.treeType === 'type') {
        actualUrl = `${cgi[cgiUrl]}${actualUrl.includes('?') ? '&' : '?'}type=true`;
      }

      // return getEdgeRequest(this.$axios, this.mozuId)
      //   .get(actualUrl, {
      //   })

      return (this.dataSourceType === 'byBiz'
        ? DeviceTreeService.instance.fetchTreeData(this.treeType === 'type')
        : DeviceTreeService.instance.fetchCollectDeviceTree(this.treeType === 'type')
      ).then((data) => {
        const matchAttrs = ['name', 'no', 'deviceTypeName', 'applicationTypeZh', 'categoryZh'];
        const matchAttrsMap = _.mapKeys(matchAttrs, _.identity);

        // 这里有bug，如果选中告警树，warningDeviceGidList变化，这里不会立即请求？
        const {
          warningDeviceGidList,
          treeOption,
        } = this;

        const isFilterAlarm = treeOption === 'alarm';

        const treeData = isFilterAlarm ? [] : data;
        const allTreeData = data;

        forEachTreeNode(data, (node, parent, indexInParent, deep) => {
          const keyMap = {
            applicationTypeEn: 'application_type_en',
            applicationTypeZh: 'application_type_zh',
            children: 'children',
            id: 'device_gid',
            name: 'device_number',
            no: 'device_no',
            deviceCategoryEn: 'device_type_en',
            deviceCategoryName: 'device_type_zh',
            deviceTypeEn: 'old_device_type_en',
            deviceTypeName: 'old_device_type_zh',
          };

          for (const oldKey in keyMap) {
            if (has(node, keyMap[oldKey])) {
              if (oldKey === 'deviceTypeEn') {
                node[oldKey] = node[keyMap[oldKey]] || node['device_type_en']
              } else if (oldKey === 'deviceTypeName') {
                node[oldKey] = node[keyMap[oldKey]] || node['device_type_zh']
              } else {
                node[oldKey] = node[keyMap[oldKey]];
              }
            }
          }

          node.children = _.orderBy(node.children, child => child.name);

          // eslint-disable-next-line no-param-reassign
          node.level = deep;
          node.alarming = false;
          // eslint-disable-next-line no-param-reassign
          node.pathMap = _.mapValues(matchAttrsMap, (att) => {
            const value = node[att]?.toLowerCase();
            if (!parent) return value ? [value] : [];
            return [
              ...parent.pathMap[att],
              value,
            ];
          });

          // eslint-disable-next-line no-param-reassign
          node.pathMapString = _.chain(node.pathMap)
            .map(values => values.join('/'))
            .join(';')
            .value();
          // eslint-disable-next-line no-param-reassign
          node.pathId = _.chain(node.pathMap.no)
            .join('/')
            .value();
          if (isFilterAlarm) {
            const isNodeAlarm = warningDeviceGidList.includes(node.id)
                && node.deviceTypeName !== '房间' && node.deviceTypeName !== '模组';

            if (isNodeAlarm) {
              treeData.push(node);
            }
          }
        });

        this.treeData = treeData;
        // this.allAlarmGids = this.getGidPath(data, this.warningDeviceGidList);

        this.handleNodeClick(this.treeData[0]);
        this.$emit('getTreeData', this.treeData);
        this.$emit('getAllTreeData', allTreeData);
        this.loading = false;

        if (this.defaultDevId || this.defaultDevName) {
          this.$nextTick(() => {
            this.findSelectNode(this.defaultDevId, this.defaultDevName);
          });
        } else {
          this.$nextTick(() => {
            this.selectFirstTemplate();
          });
        }

        
      })
        .catch((e) => {
          console.log(e);
          this.loading = false;
        });
    },

    /**
     * 自动选择第一个模板
     */
    selectFirstTemplate() {
      const level1 = this.treeData.find(e => e.children && e.children.length);
      if (!level1) {
        return;
      }
      const level2 = level1.children.find(e => e.children && e.children.length);
      if (!level2) {
        return;
      }

      const [data] = level2.children;
      const [rootNode] = this.treeData;
      this.currentNodekey = rootNode.id;
      this.expandedkeys = [data.id];
      this.currentNode = rootNode;
    },

    /**
     * 自动选择
     */
    findSelectNode(devId, devName) {
      const $this = this;
      const traverse = function (datas) {
        datas.forEach((i) => {
          if (i.id === devId || i.name === devName) {
            $this.currentNodekey = i.id;
            $this.expandedkeys = [i.id];
            $this.currentNode = i;
            return true;
          }
          if (i.children && i.children.length > 0) {
            traverse(i.children);
          }
        });
      };

      traverse(this.treeData);
      eventBus.$emit('notCascade', true);
    },

    testFilterNodeCountPerformance() {
      const id = Math.random().toString(32)
        .substring(2);
      console.time(id);
      let execTimes = 0;
      forEachTreeNode(this.treeData, (node) => {
        this.filterNode('', node);
        execTimes += 1;
      });
      console.timeEnd(id);
      console.log('exec times:', execTimes);
    },

    filterData() {
      // this.testFilterNodeCountPerformance();
      const treeStore = this.$refs.tree.store;
      const sourceLazy = treeStore.lazy;
      treeStore.lazy = true;
      this.$refs.tree.filter(this.filterText);
      // 恢复
      treeStore.lazy = sourceLazy;

      const { filterText } = this;
      if (filterText.replace(/\s/g, '') === '') {
        this.filterTextRegs = null;
      } else {
        this.filterTextRegs = new RegExp(`(${filterText.trim().replace(/\s+/g, ')|(')})`, 'ig');
      }
    },

    filterNode(value, data, node) {
      const { wordsList } = this;
      if (!wordsList.length) {
        return true;
      }

      // 几种字符串匹配算法性能比较：includes与indexOf相差无几
      // http://jsben.ch/5qRcU
      // https://www.measurethat.net/Benchmarks/Show/14772/0/includes-vs-test-vs-match-vs-indexof

      // 整体检索，更快
      return _.every( // 有任何一个关键字匹配
        wordsList,
        // word => data.pathMapString?.indexOf(word) > -1,
        word => data.pathMapString.includes(word),
        // word => checkStringIncludes(word, data.pathMapString),
      );
    },

    renderContent(h, { node }) {
      const { filterTextRegs } = this;
      let { label } = node;
      // const originLabel = label;

      if (this.enableDeviceNumberV2 === '1') {
        if (node.level > 1) {
          const [firstPart, ...otherParts] = label.split(/(?=[@#])/);
          label = [_.last(firstPart.split('-')), ...otherParts].join('');
        }
      } else {
        // 三级后的设备名称隐藏房间部分
        if (node.level >= 3) {
          let parentNode = node.parent;
          while (parentNode.level > 2) {
            parentNode = parentNode.parent;
          }

          const roomName = parentNode.data.name;
          if (this.dataSourceType === 'byBiz') {
            // label = label.includes(roomName)
            //   ? label.split(roomName)[1].slice(1)
            //   : '';
            if (label.includes(roomName)) {
              try {
                const [devName, devNum] = label.split(' ');
                label = `${devName} ${devNum.split(roomName)[1].slice(1)}`;
              } catch (error) {
                label = label.includes(roomName)
                  ? label.split(roomName)[1].slice(1)
                  : '';
              }
            }
          }
        }
      }
      const typeName = this.enableDeviceNumberV2 === '1'
        ? node.data.applicationTypeZh
        : node.data.deviceTypeName;

      label = `${typeName} | ${label}`;

      if (this.treeType === 'type' && node.data.deviceCount) {
        label = `${label} (${node.data.deviceCount})`;
      }

      if (filterTextRegs) {
        label = label.replace(filterTextRegs, '<span style="color:#ff9200;font-weight: 600;">$&</span>');
      }

      // if (this.allAlarmGids.includes(node.data.id)) {
      //   label = `<span style="color:red">${label}</span>`;
      // }
      if (node.data.alarming) {
        label = `<span style="color:red">${label}</span>`;
      }

      return (
        <span
          class="c-tree-node"
        >
          <span
            class="c-tree-node__label"
            title={node.data.no}
            domPropsInnerHTML={label}
          />
        </span>
      );
    },
    getGidPath(tree, gids) {
      const gidsSet = new Set(gids);
      const resultGidSet = new Set();

      forEachTreeNode(tree, (node) => {
        if (gidsSet.has(node.id)) {
          node.pathMap.no.forEach((gid) => {
            resultGidSet.add(gid);
          });
        }
      });

      return Array.from(resultGidSet);
    },
    getNodeRoute(tree, targetId) {
      for (let index = 0; index < tree.length; index++) {
        if (tree[index].children) {
          const endRecursiveLoop = this.getNodeRoute(tree[index].children, targetId);
          if (endRecursiveLoop) {
            this.nodePathArray.push(tree[index].deviceTypeName);
            return true;
          }
        }
        if (tree[index].id === targetId) {
          this.nodePathArray.push(tree[index].deviceTypeName);
          return true;
        }
      }
    },
    /**
     * 点击节点后，向父组件传递节点数据
     * @param {Object} data -- 传递给 data 属性的数组中该节点所对应的对象
     */
    handleNodeClick(data) {
      // 根据树类型使用不同的路径构建策略
      if (this.treeType === 'type') {
        // 按类型树：直接使用节点的 pathMap 信息，避免重复ID问题
        this.nodePathArray = data.pathMap?.deviceTypeName || [];
      } else {
        // 按位置树：使用原有的递归查找逻辑
        this.nodePathArray = [];
        if (this.treeOption === 'all') {
          this.getNodeRoute(this.treeData, data.id);
        }
      }
      this.currentNode = data;
      
      eventBus.$emit('notCascade', false);
    },
  },
};
</script>

<style lang="scss" scoped>
/deep/ .el-tree > .el-tree-node {
  min-width: 100%;
  display: inline-block;
}

.c-tree {
  &-title {
    font-size: 16px;
    font-weight: bold;
    line-height: 55px;
    padding-left: 24px;
    border-bottom: 1px solid #f0f0f0;
    color: #333;
  }

  &-tools {
    display: flex;
    padding: 0 16px;
    height: 63px;
    border-bottom: 1px solid #f0f0f0;
    align-items: center;

    &__input {
      margin-right: 8px;
    }

    &__more {
      color: #999;
    }
  }

  /deep/ &-node {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 14px;
    padding-right: 16px;

    $this: &;

    &-icon {
      color: #999;
      font-size: 20px;
      vertical-align: middle;
      margin-right: 4px;

      &.tn-icon-more {
        display: none;
      }
      &.tn-icon-more--show {
        display: inline-block;
      }
    }

    &__label {
      line-height: 24px;

      #{$this}-icon {
        margin-right: 4px;
        margin-top: -2px;
      }
    }

    &:hover {
      .tn-icon-more {
        display: inline-block;
      }
    }
  }
}
</style>
