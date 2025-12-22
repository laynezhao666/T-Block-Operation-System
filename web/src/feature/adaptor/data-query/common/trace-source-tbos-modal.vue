<template>
    <el-modal
      :visible.sync="visible"
      :title="title"
    >
    <div style="padding-left: 20px;overflow:scroll;padding-top: 20px">
    <div>
        <span style="color: #000;font-weight: 600;font-size:16px;">公式：</span>
        <json-viewer :value="expression" :expand-depth="3"></json-viewer>
    </div>
    <span style="color: #000;font-weight: 600;font-size:16px;">溯源路径：<span style="font-size:14px;color:gray">表达式 | 测点名称 | 测点值 | 测点质量 | 测点类型</span></span>
      <el-tree
        v-if="showtree"
        style="color: #000;height:500px;overflow:scroll;margin-top: 10px"
        node-key="device_number"
        :default-checked-keys="defaultChecked"
        :data="treeData"
        ref="treeRef"
        :props="treeProps"
        :render-content="renderContent"
        default-expand-all
        :expand-on-click-node=false
        :highlight-current=true
        @node-click=nodeClick
      ></el-tree>
    <el-divider></el-divider>
    <span style="color: #000;font-weight: 600;font-size:16px;">当前节点数据：</span><span style="color: #1470CC"> {{ currentNodeName }} </span>
    <json-viewer :value="textarea" :expand-depth="3"></json-viewer>
   </div>
    </el-modal>
  </template>
  <script>
import { omit } from 'lodash';
import JsonViewer from 'vue-json-viewer';
import 'vue-json-viewer/style.css';

  export default {
    props: {
    point: {
      type: Object,
      default() {
        return null;
      },
    },
  },
    components: {
        JsonViewer
    },
    computed: {
      visible: {
        get() {
            return Boolean(this.point);
        },
        set(v) {
          if (!v) {
            this.treeData = [];
            this.traceInfo = {}
            this.title =  ''
            this.expression = ''
            this.textarea = ''
            this.currentNodeName = ''
            this.defaultChecked = []
            this.$emit('update:point', null);
          }
        },
      },
    },
    data() {
      return {
        title: '',
        expression: '',
        jsonData: {},
        treeData: [],
        traceInfo : {},
        treeProps: {
        //   label: 'point_name_zh',
          children: 'children'
        },
        textarea: '',
        currentNodeName: '',
        defaultChecked: [],
        showtree : true
      };
    },
    watch: {
    point(point) {
      if (point) {
        console.log(this.point,'this.point')
        this.loadData();
      }
    },
    },
    created() {
    },
    mounted(){
        if (this.point) {
            this.loadData();
        }
    },
    methods: {
    async loadData() {
      const tracingData = await this.$axios.post('/cgi/idc-tbos-cgi/Data/TracePoint', {
        mozu_id: this.$moduleInfo.mozuId,
        point_key: this.point.id,
      });
    this.treeData = [tracingData];
    this.traceInfo = omit(tracingData, 'children')
    this.title =  "溯源 【" + this.traceInfo?.point_name_zh + "】"
    this.expression = this.traceInfo?.std_info?.expression

    this.$nextTick(() => {
        this.defaultChecked.push(this.traceInfo?.device_number)
        this.$refs.treeRef.setCurrentKey(this.traceInfo?.device_number);  
        this.textarea = this.traceInfo
        this.currentNodeName = this.traceInfo?.point_name_zh      
    });
    },
      nodeClick(all,current,_){
        this.textarea = current.data
        this.currentNodeName = this.textarea?.point_name_zh
      },
      renderContent(h, { node, data }) {
        const { children, ...otherData } = data;
        const stdTag = node?.data?.std_info ? "标准" : '采集'
        const val = (node?.data?.value?.val === null || node?.data?.value?.val === undefined) ? '--' : node?.data?.value?.val
        return h('div', [
          h('span', (node?.data?.var_name || '--') + '\u00A0\u00A0\u00A0|\u00A0\u00A0\u00A0' + (node?.data?.point_name_zh || '--') + '\u00A0\u00A0\u00A0|\u00A0\u00A0\u00A0' + val + '\u00A0\u00A0\u00A0|\u00A0\u00A0\u00A0' + node?.data?.value?.quality + '\u00A0\u00A0\u00A0|\u00A0\u00A0\u00A0' + stdTag),
        ]);
      }
    }
  };
  </script>
  
  <style scoped>

  </style>