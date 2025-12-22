import dayjs from 'dayjs';

const colors = {
  L0: '#ff3e00',
  L1: '#ff3e00',
  L2: '#ff9200',
  L3: '#D7BA14',
  L4: '#008adc',
};

export default [
  {
    name: 'Level',
    label: '告警等级',
    size: 4,
    max: 10,
    type: 'enum',
    isFilter: true,
    seqNumber: 1,
    showInTable: true,
    showInSearch: true,
    modifyName: true,
    getColumnStyle: data => ({
      color: colors[data.row.Level],
    }),
    fieldEnum: [
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
      }],
  }, {
    name: 'OccurTime',
    label: '触发时间',
    isFilter: true,
    size: 12,
    max: 20,
    seqNumber: 2,
    showInTable: true,
    showInSearch: true,
    type: 'datetime',
    formatter(row) {
      return dayjs(row.OccurTime).format('YYYY-MM-DD HH:mm:ss');
    },
  }, {
    name: 'DeviceNumber',
    label: '设备编号',
    size: 25,
    max: 20,
    isFilter: true,
    type: 'string',
    seqNumber: 3,
    showInTable: true,
  }, {
    name: 'AlarmType',
    label: '告警类型',
    size: 12,
    max: 20,
    isFilter: true,
    type: 'select',
    seqNumber: 4,
    showInTable: true,
    showInSearch: true,
    dropdownMethod: 'post',
    dropdownPath: '/cgi/alarm/active/getAlarmType',
    dropdownQuery: () => ({ start: 0, limit: 0 }),
  }, {
    name: 'Content',
    label: '告警内容',
    size: 20,
    max: 30,
    isFilter: true,
    type: 'string',
    seqNumber: 5,
    showInTable: true,
  }, {
    name: 'MozuName',
    label: '模组',
    size: 12,
    max: 30,
    type: 'string',
    seqNumber: 5,
    showInTable: true,
  }, {
    name: 'RoomName',
    label: '房间',
    size: 12,
    max: 30,
    type: 'string',
    seqNumber: 5,
    showInTable: true,
  }, {
    name: 'operator',
    label: '操作',
    fixed: 'right',
    size: 3,
    max: 30,
    isFilter: false,
    seqNumber: 11,
    show: true,
    showInTable: true,
    operationUrl: [{ operation: '详情', url: 'c' }],
    operationMap: { 详情: '详情' },
    instantOperation: '详情',
  }];
