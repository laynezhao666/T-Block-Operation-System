<template>
  <div class="strategy-detail-block">
    <el-title>
      <el-breadcrumb separator-class="el-icon-arrow-right">
        <el-breadcrumb-item
          :to="{ path: '/timpage/strategy-effective-verify' }"
        >
          <span
            style="color:#1470CC"
            @click="jumpUrl('/timpage/strategy-effective-verify')"
          >策略生效验证</span>
        </el-breadcrumb-item>
        <el-breadcrumb-item>详情</el-breadcrumb-item>
      </el-breadcrumb>
    </el-title>
    <el-space
      direction="vertical"
      size="middle"
    >
      <el-block padding>
        <data-block :info="deviceInfo" />
      </el-block>
      <el-block padding>
        <data-block :info="strategyInfo" />
      </el-block>
      <el-block>
        <div
          style="font-size:18px;font-weight:550;margin:16px 24px;padding-top:16px"
        >
          判断信息
        </div>
        <el-table
          :data="tableData"
          style="width: 100%"
        >
          <el-table-column
            prop="pointName"
            label="策略所用测点"
            width="180"
          />
          <!-- <el-table-column
            prop="thresholdValue"
            label="判断值"
            width="180"
          /> -->
          <el-table-column
            prop="currentValue"
            label="当前值"
          />
          <el-table-column
            prop="address"
            label="历史值"
          >
            <template slot-scope="scope">
              <el-button
                type="text"
                @click="jumpHistory(scope.row)"
              >
                查看
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-block>
      <el-block
        v-if="showErrorLog"
        padding
      >
        <data-block
          :info="errorLog"
          :showtitle="false"
        />
      </el-block>
    </el-space>
  </div>
</template>
<script>
import { warning as cgi } from '@@/config/cgi';
import dataBlock from '../component/datablock';
import { getQueryString } from 'common/script/utils.js';
import getEdgeRequest from '../../utils/request';

export default {
  components: {
    dataBlock,
  },
  data() {
    return {
      mozuId: 326,
      devName: '',
      devNum: '',
      url: '',
      tableData: [],
      showErrorLog: true,
      deviceInfo: {
        title: '设备信息',
        data: [
          { label: '设备编号', prop: 'deviceNumber', value: '' },
          { label: '设备名称', prop: 'deviceName', value: '' },
          { label: '设备协议类型', prop: 'protocolType', value: '' },
        ],
      },
      strategyInfo: {
        title: '策略信息',
        data: [
          { label: '告警类型', prop: 'alarmType', value: '' },
          { label: '告警等级', prop: 'alarmLevel', value: '' },
          { label: '告警内容', prop: 'alarmContent', value: '' },
          { label: '触发表达式', prop: 'alarmExpress', value: '' },
          { label: '恢复表达式', prop: 'restoreExpress', value: '' },
        ],
      },
      errorLog: {
        title: '错误日志',
        data: [
          { label: '错误码', prop: 'code', value: '' },
          { label: '无效原因', prop: 'msg', value: '' },

        ],
      },
      levelMap: { L0: '零级',
        L1: '一级',
        L2: '二级',
        L3: '三级',
        L4: '四级',
        L5: '五级' },
      params: {

      },
    };
  },
  mounted() {
    if (getQueryString('deviceGid')) {
      this.url = cgi.getValidateDetail;
      this.params = {
        deviceGid: getQueryString('deviceGid'),
        ruleId: parseInt(getQueryString('ruleId')),
        validateType: parseInt(getQueryString('validateType')),
        mozuId: parseInt(getQueryString('mozuId')),
      };
    } else {
      this.url = cgi.getHistoryDetail;
      this.params = {
        validateId: parseInt(getQueryString('validateId')),
        mozuId: parseInt(getQueryString('mozuId')),

      };
    }
    this.mozuId = parseInt(getQueryString('mozuId'));
    if (getQueryString('valid') === '1') {
      this.showErrorLog = false;
    }
    this.getValidateDetail();
  },
  methods: {
    jumpUrl(url) {
      window.open(url);
    },
    jumpHistory(row) {
      window.open(`/timpage/advanced-search?pointlist=${this.devNum}.${row.pointName}&mozuId=${this.mozuId}`);
    },
    getValidateDetail() {
      getEdgeRequest(this.$axios, this.mozuId).post(this.url, { ...this.params }, undefined, { isJson: true })
        .then((data) => {
          this.errorLog.data.map((item) => {
            item.value = data[item.prop];
            return item;
          });
          this.strategyInfo.data.map((item) => {
            item.value = data.ruleInfo[item.prop];
            if (item.prop === 'alarmLevel' && this.levelMap[data.ruleInfo[item.prop]]) {
              item.value = this.levelMap[data.ruleInfo[item.prop]];
            }
            return item;
          });
          this.deviceInfo.data.map((item) => {
            item.value = data.deviceInfo[item.prop];
            if (item.prop === 'deviceName') {
              this.devName = data.deviceInfo[item.prop];
            }
            if (item.prop === 'protocolType') {
              if (!data.deviceInfo[item.prop]) {
                item.value = '--';
              }
            }
            if (item.prop === 'deviceNumber') {
              this.devNum = data.deviceInfo[item.prop];
            }
            return item;
          });
          this.tableData = data.pointDetail;
        });
    },
  },
};
</script>

<style lang="scss">
.strategy-detail-block{
  .el-icon-arrow-right {
    // content: "\e604";
    transform: translate(0px,5px);
  }
}
</style>
