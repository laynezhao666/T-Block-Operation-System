<template>
  <!--
    feat-sanbox分支对这个文件修改过，但是没有合并到主分支，直接迁移/合并有什么影响未知。
    （需要搜索查阅各文件分析影响的页面、找对应的人了解背景、以确定是否可以合并为一个文件）
    本次迁移，直接拷贝一份完成任务，请见谅
  -->
  <div class="strategy-detail-block">
    <el-title>
      <el-breadcrumb separator-class="el-icon-arrow-right">
        <el-breadcrumb-item
          :to="{ path: '/tedge/warning-effective-verify' }"
        >
          <span
            style="color:#1470CC"
            @click="jumpUrl('/tedge/warning-effective-verify')"
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
import { warning as cgi, tbosWarning as tbosCgi } from '@@/config/cgi';
import dataBlock from '../component/datablock';
import { getQueryString } from 'common/script/utils.js';
import getEdgeRequest from '../../utils/request';
import { getMozuId } from 'feature/utils/business';
import business from '@@/config/business';

export default {
  components: {
    dataBlock,
  },
  data() {
    return {
      business,
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
      mounted: false,
    };
  },
  computed: {
    curMozuData() {
      return window.__GetFrameDataByKey('curMozuData');
    },
  },
  watch: {
    curMozuData: {
      handler(v) {
        if (v && v.alarmapi) {
          if (!this.mounted) {
            this.getValidateDetail();
          }
        }
      },
      immediate: true,
    },
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
        mozuId: this.business.showModuleSelected ? getMozuId() : parseInt(getQueryString('mozuId')),

      };
    }
    if (!this.business.showModuleSelected) {
      this.mozuId = getMozuId();
    } else {
      this.mozuId = parseInt(getQueryString('mozuId'));
    }

    if (getQueryString('valid') === '1') {
      this.showErrorLog = false;
    }
    if (!this.curMozuData) {
      return;
    }
    this.mounted = true;
    if (this.$moduleInfo?.isTbos) {
      this.getTbosValidateDetail();
    } else {
      this.getValidateDetail();
    }
  },
  methods: {
    jumpUrl(url) {
      window.open(url);
    },
    jumpHistory(row) {
      window.open(`/tedge/advanced-search?pointlist=${this.devNum || row?.deviceNumber}.${row.pointName}&mozuId=${this.mozuId}`);
    },
    async getTbosValidateDetail() {
      const result = await this.getTbosValidDetail();

      await this.getRuleInfo(
        {
          alarm_name: [getQueryString('alarmType')],
          page: 1,
          size: 10,
        },
        {
          deviceNumber: result?.device_number,
        }
      );
    },
    async getTbosValidDetail() {
      const result = await getEdgeRequest(this.$axios, this.mozuId)
        .post(tbosCgi.GetValidate, {
          alarm_name: [getQueryString('alarmType')],
          device_gid: [getQueryString('deviceGid')],
          valid_type: getQueryString('valid_type')
        }, undefined, { isJson: true });
      const { points } = result?.list[0];
      const logItem = {
        code: result?.list[0]?.error_code,
        msg: result?.list[0]?.error_name
      }
      this.errorLog.data.forEach((item) => {
          item.value = logItem[item.prop];
          return item;
      });
      const pointDateList = (await getEdgeRequest(this.$axios, '').post(tbosCgi.pointQuery, {
        conditions: [{
          name: 'point_key',
          value: points,
        }],
        data_type: 0,
      }))?.list;
      this.tableData = pointDateList.map(i => ({
        pointName: i?.point_name_zh,
        point: i?.point_key,
        currentValue: i?.latest_value || '--',
        deviceNumber: i?.device_number
      }));
      return result?.list[0];
    },
    async getRuleInfo(data, otherParams = {}) {
      const ruleInfo = await getEdgeRequest(this.$axios, this.mozuId)
        .post(tbosCgi.GetStrategy, { ...data }, undefined, { isJson: true });
      const { list } = ruleInfo;
      const [ruleItemInfo] = list;
      this.strategyInfo = {
        title: '策略信息',
        data: [
          { label: '告警类型', prop: 'alarm_name', value: '' },
          { label: '告警等级', prop: 'level', value: '' },
          { label: '告警内容', prop: 'content', value: '' },
          { label: '触发表达式', prop: 'alarm_exp', value: '' },
          { label: '恢复表达式', prop: 'restore_exp', value: '' },
        ],
      };
      console.log(ruleItemInfo, 'ruleItemInfo');
      this.strategyInfo.data.map((item) => {
        item.value = ruleItemInfo[item.prop];
        return item;
      });
      const deviceInfo = {
        deviceNumber: otherParams?.deviceNumber,
        protocolType: ruleItemInfo?.device_type,
        deviceName: '',
      };
      this.deviceInfo.data.map((item) => {
        item.value = deviceInfo[item.prop];
        return item;
      });

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
