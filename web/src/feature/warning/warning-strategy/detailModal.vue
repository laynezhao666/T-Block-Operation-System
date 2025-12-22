<template>
  <el-modal
    :visible.sync="modalVisible"
    :width="960"
    @closed="closed"
  >
    <template slot="title">
      告警策略详情
    </template>
    <el-form
      id="detailForm"
      ref="form"
      :model="form"
      label-width="120px"
      style="padding: 0;"
    >
      <el-block
        v-for="group in groups"
        :key="group.label"
        class="form-group"
        inner
        :header="group.label"
      >
        <template v-for="item in group.options">
          <el-form-item
            v-if="item.props !== 'deviceNumberList'"
            :key="item.props"
            :label="item.label"
            label-width="200px"
            style="border-bottom:1px solid #f0f0f0;"
          >
            {{ item.value | LevelMap(item.label, item.value) }}{{ item.unit }}
          </el-form-item>
          <template v-else>
            <div
              :key="item.props"
              style="display:flex;flex-wrap:wrap;line-height:56px;padding: 0 72px;"
            >
              <div
                v-for="deviceItem in item.value"
                :key="deviceItem.deviceNumber"
                style="min-width:240px;text-align: center;"
              >
                <span style="margin-right:20px">{{ deviceItem.deviceNumber }}</span>
              </div>
            </div>
          </template>
        </template>
      </el-block>
    </el-form>
    <!-- <template slot="footer">
      <el-button @click="closed">
        取消
      </el-button>
    </template> -->
  </el-modal>
</template>
<script>
import { flatten, map } from 'lodash/fp';

export default {
  filters: {
    LevelMap(level, key) {
      if (key === '告警等级') {
        const levelMaps = {
          L0: '零级',
          L1: '一级',
          L2: '二级',
          L3: '三级',
          L4: '四级',
          L5: '五级',
        };
        return levelMaps[level];
      }
      return level;
    },
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    params: {
      type: Object,
      default() {
        return {};
      },
    },
    modalData: {
      type: Object,
      default() {
        return {};
      },
    },
  },
  data() {
    return {
      groups: [{
        label: '基本信息',
        options: [
          { label: '告警类型', value: '', props: 'alarmType' },
          { label: '设备类型', value: '', props: 'protocolTypeName' },
          { label: '标准策略', value: '', props: 'isStandard' },
          { label: '告警等级', value: '', props: 'alarmLevel' },
        ],
      }, {
        label: '设备信息',
        options: [
          { label: '设备编号', value: '', props: 'deviceNumberList' },
        ],
      }, {
        label: '触发表达式',
        options: [
          { label: '触发表达式', value: '', props: 'occurExpression' },
          // { label: '触发判断周期', value: '', props: 'occurPeriod' },
          // { label: '触发周期内发生次数', value: '', props: 'occurCount' },
        ],
      }, {
        label: '恢复表达式',
        options: [
          { label: '恢复表达式', value: '', props: 'restoreExpression' },
          // { label: '恢复判断周期', value: '', props: 'restorePeriod' },
          // { label: '恢复周期内发生次数', value: '', props: 'restoreCount' },
        ],
      }, {
        label: '其他',
        options: [
          { label: '告警内容', value: '', props: 'alarmContent' },
          { label: '影响分析', value: '', props: 'influence' },
          { label: '处理建议', value: '', props: 'suggestion' },
        ],
      }],
      // options: [
      //   { label: '设备编号', value: '', props: 'deviceNumberList' },
      //   { label: '设备协议类型', value: '', props: 'protocolTypeName' },
      //   { label: '标准策略', value: '', props: 'isStandard' },
      //   { label: '告警类型', value: '', props: 'alarmType' },
      //   { label: '告警级别', value: '', props: 'alarmLevel' },
      //   { label: '触发表达式', value: '', props: 'occurExpression' },
      //   { label: '触发判断周期', value: '', props: 'occurPeriod' },
      //   { label: '触发周期内发生次数', value: '', props: 'occurCount' },
      //   { label: '恢复表达式', value: '', props: 'restoreExpression' },
      //   { label: '恢复判断周期', value: '', props: 'restorePeriod' },
      //   { label: '恢复周期内发生次数', value: '', props: 'restoreCount' },
      //   { label: '告警内容', value: '', props: 'alarmContent' },
      //   { label: '影响分析', value: '', props: 'influence' },
      //   { label: '处理建议', value: '', props: 'suggestion' },
      // ],
      loading: false,
      form: {
        comment: '',
        uid: '',
      },
      modalVisible: false,
    };
  },
  computed: {
    options() {
      return this.groups |> map('options') |> flatten;
    },
  },
  watch: {
    visible(val) {
      this.modalVisible = val;
      this.options = this.options.map((item) => {
        if (item.props === 'deviceNumberList') {
          item.value = this.modalData[item.props];
          const newDeviceNumberList = [];
          item.value.forEach((item) => {
            newDeviceNumberList.push({ deviceNumber: item });
          });
          item.value = newDeviceNumberList;
          // item.value = this.modalData[item.props].join(',');
        } else {
          item.value = this.modalData[item.props];
        }
        if (item.props === 'isStandard') {
          item.value = this.modalData[item.props] === true ? '是' : '否';
        }
        return item;
      });
    },
  },
  mounted() {

  },
  methods: {
    closed() {
      this.form = {
        desc: '',
        uid: '',
      };
      //   this.modalVisible = false;
      this.$emit('update:visible', false);
    },
  },
};
</script>

<style lang="scss" scoped>
.el-form-item {
  border-bottom: none !important;
}
.form-group {
  border-bottom: 1px solid rgb(240, 240, 240);
  + .form-group {
    margin-top: 16px;
  }
}
.inner-title {
  color:#333;
  font-size: 16px;
  font-weight: 700;
}
</style>
