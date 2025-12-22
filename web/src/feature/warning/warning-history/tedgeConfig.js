import dayjs from 'dayjs';
import { getMozuId } from '../../utils/business';
import business from '@@/config/business';
import getEdgeRequest from '../../utils/request';
import { DeviceTreeService } from 'services/tedge/device-tree.service';

// import http from 'common/script/http2';
const duration = require('dayjs/plugin/duration');
dayjs.extend(duration);

function floor(v) {
  return Math.floor(v.toFixed(2));
}
const colors = {
  L0: '#ff3e00',
  L1: '#ff3e00',
  L2: '#ff9200',
  L3: '#fbd743',
  L4: '#008adc',
  L5: '#8acbf2',
};
export default [
  {
    name: 'deviceGids',
    label: '设备树',
    size: 25,
    max: 30,
    isFilter: true,
    type: 'treeSelect',
    isCascaderMulti: true,
    cascaderLabel: 'name',
    cascaderValue: 'id',
    seqNumber: 4,
    showInTable: false,
    showInSearch: true,
    dropdownMethod: 'get',
    dropdownPath: '/cgi/dataQuery/edge/getBizDeviceLevelTree',
    dropdownQuery: () => ({ field: 'deviceNumber', mozuId: getMozuId() }),
    getTree() {
      return DeviceTreeService.instance.fetchTreeData();
    },
    onChange(data) {
      console.log(data, '??');
      // this.formData.deviceNumber = data[0][4];
    },
  },

  {
    name: 'alarmLevels',
    label: '告警等级',
    size: 6,
    max: 20,
    isFilter: true,
    type: 'select',
    seqNumber: 1,
    showInTable: false,
    showInSearch: true,
    modifyName: true,
    getColumnStyle: data => ({
      color: colors[data.row.alarmLevel],
      border: `1px solid${colors[data.row.alarmLevel]}`,
      padding: '0 8px',
      'border-radius': '6px',
    }),
    modifyMap: [
      {
        label: '零级',
        value: 'L0',
      }, {
        label: '一级',
        value: 'L1',
      }, {
        label: '二级',
        value: 'L2',
      }, {
        label: '三级',
        value: 'L3',
      }, {
        label: '四级',
        value: 'L4',
      }, {
        label: '五级',
        value: 'L5',
      },
    ],
    localData: true,
    fieldMultiEnum: {
      零级: '零级',
      一级: '一级',
      二级: '二级',
      三级: '三级',
      四级: '四级',
      五级: '五级',
    },
  },
  {
    name: 'alarmLevel',
    label: '告警等级',
    size: 6,
    max: 20,
    isFilter: false,
    type: 'select',
    seqNumber: 1,
    showInTable: true,
    showInSearch: true,
    modifyName: true,
    getColumnStyle: data => ({
      color: colors[data.row.alarmLevel],
      border: `1px solid${colors[data.row.alarmLevel]}`,
      padding: '0 8px',
      'border-radius': '6px',
    }),
    modifyMap: [
      {
        label: '零级',
        value: 'L0',
      }, {
        label: '一级',
        value: 'L1',
      }, {
        label: '二级',
        value: 'L2',
      }, {
        label: '三级',
        value: 'L3',
      }, {
        label: '四级',
        value: 'L4',
      }, {
        label: '五级',
        value: 'L5',
      },
    ],
    localData: true,
    fieldMultiEnum: {
      零级: '零级',
      一级: '一级',
      二级: '二级',
      三级: '三级',
      四级: '四级',
      五级: '五级',
    },
  },

  {
    name: 'alarmTypes',
    label: '告警类型',
    size: 15,
    // fixed: true,
    max: 30,
    isFilter: true,
    type: 'selectNotRemote',
    seqNumber: 1,
    showInTable: false,
    showInSearch: true,
    modal: 'alarmType',
    dropdownMethod: 'post',
    dropdownPath: '/cgi/alarm/history/getDropdown',
    dropdownQuery: () => ({ field: 'alarmType', mozuId: getMozuId() }),
  },
  {
    name: 'alarmType',
    label: '告警类型',
    size: 15,
    // fixed: true,
    max: 30,
    isFilter: false,
    type: 'selectNotRemote',
    seqNumber: 1,
    showInTable: true,
    showInSearch: true,
    modal: 'alarmType',
    dropdownMethod: 'post',
    dropdownPath: '/cgi/alarm/history/getDropdown',
    dropdownQuery: () => ({ field: 'alarmType', mozuId: getMozuId() }),
  },

  {
    name: 'alarmOrigin',
    label: '告警源',
    size: 30,
    max: 30,
    isFilter: false,
    type: 'string',
    seqNumber: 1,
    showInTable: true,
    showInSearch: true,
    formatter(row) {
      const deviceNumber = window.tnwebServices.v2DeviceNumberTransformerService.get(row.deviceNumber);
      return `${row.deviceType}【${deviceNumber}】`;
    },
  },
  {
    name: 'alarmContent',
    label: '告警原因',
    size: 20,
    max: 30,
    isFilter: false,
    type: 'pureStringNotArray',
    seqNumber: 6,
    showInTable: true,
    showInSearch: true,
  },
  {
    name: 'mozuName',
    label: '模组',
    isFilter: false,
    size: 8,
    max: 20,
    seqNumber: 2,
    showInTable: false,
    showInSearch: false,
  },
  //    {
  //     name: 'roomName',
  //     label: '房间',
  //     size: 10,
  //     max: 10,
  //     isFilter: false,
  //     seqNumber: 3,
  //     showInTable: true,
  //     showInSearch: true,
  //   },

  {
    name: 'closeOperatorName',
    label: '关闭人',
    size: 8,
    max: 30,
    isFilter: !business.isTedge,
    type: 'singleSelect',
    seqNumber: 10,
    showInTable: !business.isTedge,
    showInSearch: !business.isTedge,
    dropdownMethod: 'post',
    dropdownPath: '/cgi/alarm/history/getDropdown',
    dropdownQuery: () => ({ field: 'closeOperatorName', mozuId: getMozuId() }),
  },
  {
    name: 'occurTime',
    label: '产生时间',
    size: 12,
    max: 30,
    isFilter: true,
    type: 'datetime',
    seqNumber: 7,
    showInTable: true,
    showInSearch: true,
  },
  {
    name: 'restoreTime',
    label: '恢复时间',
    size: 12,
    max: 30,
    isFilter: true,
    type: 'datetime',
    seqNumber: 8,
    showInTable: true,
    showInSearch: true,
  },
  {
    name: 'closeReason',
    label: '恢复方式',
    size: 12,
    max: 30,
    isFilter: false,
    seqNumber: 8,
    showInTable: true,
    showInSearch: false,
    formatter(row) {
      return row.closeTime ? '手工恢复' : '自动恢复';
    },
  },
  {
    name: 'duration',
    label: '持续时间',
    size: 12,
    max: 30,
    isFilter: true,
    type: 'compareSearch',
    seqNumber: 8,
    showInTable: true,
    showInSearch: true,
    placeholder: '请输入',
    formatter(row) {
      const occurTime = dayjs(row.occurTime);
      let endTime;
      if (row.restoreTime) {
        endTime = dayjs(row.restoreTime);
      } else if (row.closeTime) {
        endTime = dayjs(row.closeTime);
      } else {
        endTime = dayjs();
      }
      let seconds = endTime.diff(occurTime) / 1000;
      const days = floor(seconds / (24 * 3600));
      seconds = seconds % (24 * 3600);
      const hours = floor(seconds / 3600);
      seconds = seconds % 3600;
      const mins = floor(seconds / 60);
      seconds = seconds % 60;
      return `${days ? `${days}天` : ''}${hours ? `${hours}小时` : ''}${mins ? `${mins}分钟` : ''}${seconds ? `${seconds}秒` : ''}`;
    },
  },

  //   {
  //     name: 'closeTime',
  //     label: '关闭时间',
  //     size: 12,
  //     max: 30,
  //     isFilter: true,
  //     type: 'datetime',
  //     seqNumber: 9,
  //     showInTable: true,
  //     showInSearch: true,
  //   },
  //   {
  //     name: 'closeReason',
  //     label: '关闭原因',
  //     size: 12,
  //     max: 30,
  //     isFilter: false,
  //     seqNumber: 10,
  //     showInTable: true,
  //     showInSearch: false,
  //   },
  // {
  //   name: 'operator',
  //   label: '操作',
  //   fixed: 'right',
  //   size: 3,
  //   max: 30,
  //   isFilter: false,
  //   seqNumber: 11,
  //   show: true,
  //   showInTable: true,
  //   // operationUrl: [{ 转责任人: '/tassets/statistics-asset-detail' }, { 确认工单: '确认工单' }, { 认领: 'c' }],
  //   operationUrl: [{ operation: '详情', url: 'c' }],
  //   operationMap: { 详情: '详情' },
  //   instantOperation: '详情',
  //   // jumpQuery: ['asset_id'],
  // },
];
