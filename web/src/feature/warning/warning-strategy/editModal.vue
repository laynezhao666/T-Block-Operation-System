<template>
  <el-modal
    :visible.sync="modalVisible"
    :width="850"
    @closed="closed"
  >
    <template slot="title">
      编辑告警策略
    </template>
    <el-form
      id="editForm"
      ref="form"
      :model="form"
      label-width="180px"
      :rules="rule"
    >
      <el-form-item
        ref="protocolType"
        label="设备类型"
        prop="protocolType"
        class="must"
        :inline-message="true"
      >
        <el-select
          v-model="form.protocolType"
          clearable
          filterable
          collapse-tags
          reserve-keyword
          placeholder="请输入关键词"
          @change="changeProtocol"
        >
          <el-option
            v-for="item in protocolList"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <div
          slot="error"
          slot-scope="{ error }"
          class="el-form-item__error el-form-item__error--inline custom_error"
        >
          {{ error }}
        </div>
      </el-form-item>
      <el-form-item
        ref="deviceGidList"
        label="设备编号"
        class="must"
        :inline-message="true"
      >
        <div style="display:flex; height: 56px; align-items: center;">
          <el-select
            v-model="deviceNumberList"
            style="width: 430px"
            clearable
            filterable
            remote
            collapse-tags
            multiple
            reserve-keyword
            placeholder=""
            @change="deviceNumberListChange"
          >
            <!-- :remote-method="v => getList(v)" -->

            <el-option
              v-for="item in form.deviceNumberList"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-button
            type="text"
            @click="selectDevice"
          >
            {{ chooseName }}
          </el-button>
        </div>
        <div
          slot="error"
          slot-scope="{ error }"
          class="el-form-item__error el-form-item__error--inline custom_error"
        >
          {{ error }}
        </div>
      </el-form-item>

      <select-device
        :visible.sync="showSelectDevice"
        :listdata="form.deviceNumberList"
        :choosed-device="deviceNumberList"
        @confirm="choosedDevice"
      />
      <el-form-item
        prop="alarmType"
        class="alarm-tip-label"
        :inline-message="true"
      >
        <template slot="label">
          告警类型
          <el-help-tip
            width="400"
            popper-class="test"
          >
            <div class="help-wrap">
              <p>
                <span class="help-title">告警类型：</span>
                {{ textArray[4] }}
              </p>
              <br>
              <p>
                <span class="help-title"> 说明：</span>
                同一设备不同告警策略的告警类型名称不能重复。
              </p>
            </div>
          </el-help-tip>
        </template>
        <el-input
          v-model="form.alarmType"
          placeholder="请输入"
        />
        <div
          slot="error"
          slot-scope="{ error }"
          class="el-form-item__error el-form-item__error--inline custom_error"
        >
          {{ error }}
        </div>
      </el-form-item>
      <el-form-item
        label="告警等级"
        prop="alarmLevel"
        :inline-message="true"
      >
        <el-radio
          v-for="level in alarmLevel"
          :key="level"
          v-model="form.alarmLevel"
          :label="level"
        >
          {{ level | LevelMap }}
        </el-radio>
        <!-- <el-radio
          v-model="form.alarmLevel"
          label="2"
        >
          备选项
        </el-radio> -->
        <div
          slot="error"
          slot-scope="{ error }"
          class="el-form-item__error el-form-item__error--inline custom_error"
        >
          {{ error }}
        </div>
      </el-form-item>
      <el-form-item
        ref="occurExpression"
        class="alarm-tip-label"
        label="触发表达式"
        prop="occurExpression"
        :inline-message="true"
      >
        <template slot="label">
          触发表达式
          <el-help-tip
            width="400"
            popper-class="test"
          >
            <div class="help-wrap">
              <p>
                <span class="help-title">表达式：</span>
                是由数字、运算符、数字分组符号(括号)、测点等以能求得数值的有意义排列方法所得的组合。
              </p>
              <br>
              <div>
                <span class="help-title"> 运算符：</span>
                当前表达式中支持使用如下运算符
                <br>
                <span>{{ textArray[3] }}</span>
                <br>
                <span>{{ textArray[0] }}</span>
                <br>
                <span>{{ textArray[1] }}</span>
              </div>
              <br>
              <p>
                <span class="help-title">使用示例：</span>
                <br>
                1. 以"机架冷通道温度过高"告警为例，表达式输入内容为"冷通道温度 > 27 "。
                <br>
                {{ textArray[2] }}
              </p>
            </div>
          </el-help-tip>
        </template>
        <div style="display: flex; flex-wrap: wrap; align-items: center">
          <!-- <el-input
            v-model="form.occurExpression"
            placeholder="请输入"
            style="width:430px"
          /> -->
          <template v-for="item in occurExpressionList">
            {{ item.label }}
            <el-input
              v-if="item.value !== null"
              :key="item.label"
              v-model="item.value"
              placeholder="请输入"
              :style="{ width: `${item.value.length}0px` }"
              class="expression-input"
            />
            <template v-if="item.inputExtr">
              {{ item.inputExtr }}
            </template>
            <template v-if="item.unit">
              {{ item.unit }}
            </template>
          </template>
          <!-- <head-selector
            type="occurExpression"
            :url="queryPointUrl"
            :params="{deviceGidList: deviceParam}"
            :filter="filter"
            ajax-method="post"
            label-name="添加测点"
            @change="occurHeadChange"
          /> -->
        </div>
        <div
          slot="error"
          slot-scope="{ error }"
          class="el-form-item__error el-form-item__error--inline custom_error"
        >
          <!-- {{ error }} -->
          <p
            v-for="(item, index) in error.split('&&&')"
            :key="index"
          >
            {{ item }}
          </p>
        </div>
      </el-form-item>
      <!-- <el-form-item
        ref="occurPeriod"
        label="触发判断周期"
        prop="occurPeriod"
      >
        <el-input-number
          v-model="form.occurPeriod"
          :min="1"
          label="描述文字"
          @change="handleChange"
        />
      </el-form-item> -->
      <!-- <el-form-item
        ref="occurCount"
        label="触发周期内发生次数"
        prop="occurCount"
      >
        <el-input-number
          v-model="form.occurCount"
          :min="1"
          label="描述文字"
          @change="handleChange"
        />
      </el-form-item> -->
      <el-form-item
        ref="restoreExpression"
        label="恢复表达式"
        prop=""
        :inline-message="true"
      >
        <div style="display: flex; flex-wrap: wrap; align-items: center">
          <!-- <el-input
            v-model="form.restoreExpression"
            placeholder="请输入"
            style="width:430px"
          /> -->
          <template v-for="item in restoreExpressionList">
            {{ item.label }}
            <el-input
              v-if="item.value !== null"
              :key="item.label"
              v-model="item.value"
              placeholder="请输入"
              :style="{ width: `${item.value.length}0px` }"
              class="expression-input"
            />
            <template v-if="item.inputExtr">
              {{ item.inputExtr }}
            </template>
            <template v-if="item.unit">
              {{ item.unit }}
            </template>
          </template>
          <!-- <head-selector
            type="restoreExpression"
            :url="queryPointUrl"
            :params="{deviceGidList: deviceParam}"
            :filter="filter"
            ajax-method="post"
            label-name="添加测点"
            @change="restoreHeadChange"
          /> -->
        </div>
        <div
          slot="error"
          slot-scope="{ error }"
          class="el-form-item__error el-form-item__error--inline custom_error"
        >
          <!-- {{ error }} -->
          <p
            v-for="(item, index) in error.split('&&&')"
            :key="index"
          >
            {{ item }}
          </p>
        </div>
      </el-form-item>
      <!-- <el-form-item
        ref="restorePeriod"
        label="恢复判断周期"
        prop="restorePeriod"
      >
        <el-input-number
          v-model="form.restorePeriod"
          :min="1"
          label="描述文字"
          @change="handleChange"
        />
      </el-form-item> -->
      <!-- <el-form-item
        ref="restoreCount"
        label="恢复周期内发生次数"
        prop="restoreCount"
      >
        <el-input-number
          v-model="form.restoreCount"
          :min="1"
          label="描述文字"
          @change="handleChange"
        />
      </el-form-item> -->
      <el-form-item
        id="alarm-content-edit"
        ref="alarmContent"
        class="alarm-tip-label"
        label=""
        prop="alarmContent"
        :inline-message="true"
      >
        <template slot="label">
          告警内容
          <el-help-tip
            width="200"
            popper-class="test"
          >
            <p><span v-pre>当前单体电池温度为{{温度}},超过正常34</span></p>
          </el-help-tip>
        </template>
        <el-input
          v-model="form.alarmContent"
          placeholder="请输入"
        />
        <div
          slot="error"
          slot-scope="{ error }"
          class="el-form-item__error el-form-item__error--inline custom_error"
        >
          {{ error }}
        </div>
      </el-form-item>
      <el-form-item
        label="影响分析"
        prop=""
      >
        <el-input
          v-model="form.influence"
          placeholder="请输入"
        />
      </el-form-item>
      <el-form-item
        label="处理建议"
        prop=""
      >
        <el-input
          v-model="form.suggestion"
          placeholder="请输入"
        />
      </el-form-item>
    </el-form>
    <template slot="footer">
      <el-button
        type="primary"
        @click="submit"
      >
        确定
      </el-button>
      <!-- <el-button @click="closed">
        取消
      </el-button> -->
    </template>
  </el-modal>
</template>
<script>
// import headSelector from './head-selector';
import { warning as cgi } from '@@/config/cgi';
import getEdgeRequest from '../../utils/request';
import selectDevice from './selectDevice';

export default {
  filters: {
    LevelMap(level) {
      const levelMaps = {
        L0: '零级',
        L1: '一级',
        L2: '二级',
        L3: '三级',
        L4: '四级',
        L5: '五级',
      };
      return levelMaps[level];
    },
  },
  components: {
    // headSelector,
    selectDevice,
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
    mozuId: {
      type: Number,
      default: 326,
    },
  },
  inject: ['tableConfig'],
  data() {
    return {
      showSelectDevice: false,
      deviceParam: [],
      protocolList: [],
      queryPointUrl: cgi.queryPointTypeList,
      alarmLevel: ['L0', 'L1', 'L2', 'L3', 'L4', 'L5'],
      filter: {},
      num1: 0,
      abc: '',
      radio: '1',
      form: {
        deviceNumberList: [],
        protocolType: [],
        alarmType: '',
        alarmLevel: '',
        occurExpression: '',
        occurPeriod: '',
        occurCount: 0,
        restoreExpression: '',
        restorePeriod: '',
        restoreCount: '',
        alarmContent: '',
        influence: '',
        suggestion: '',
      },
      rule: {
        alarmType: [
          { required: true, message: '请输入', trigger: 'change' },
        ],
        alarmLevel: [
          { required: true, message: '请输入', trigger: 'change' },
        ],
        occurExpression: [
          { required: true, message: '请输入', trigger: 'change' },
        ],
        occurPeriod: [
          { required: true, message: '请输入', trigger: 'change' },
        ],
        occurCount: [
          { required: true, message: '请输入', trigger: 'change' },
        ],
        // alarmContent: [
        //   { required: true, message: '请输入', trigger: 'change' },
        // ],
      },
      options: [
        { label: '设备编号', value: '', props: 'deviceNumberList' },
        { label: '设备协议类型', value: '', props: 'protocolType' },
        { label: '标准策略', value: '', props: 'isStandard' },
        { label: '告警类型', value: '', props: 'alarmType' },
        { label: '告警等级', value: '', props: 'alarmLevel' },
        { label: '触发表达式', value: '', props: 'occurExpression' },
        { label: '触发判断周期', value: '', props: 'occurPeriod' },
        { label: '触发周期内发生次数', value: '', props: 'occurCount' },
        { label: '恢复表达式', value: '', props: 'restoreExpression' },
        { label: '恢复判断周期', value: '', props: 'restorePeriod' },
        { label: '恢复周期内发生次数', value: '', props: 'restoreCount' },
        { label: '告警内容', value: '', props: 'alarmContent' },
        { label: '影响分析', value: '', props: 'influence' },
        { label: '处理建议', value: '', props: 'suggestion' },
      ],
      validateMap: {
        alarmType: 0,
        alarmLevel: 1,
        occurExpression: 2,
        occurPeriod: 3,
        occurCount: 4,
        alarmContent: 5,
      },
      loading: false,
      checkedCities1: [],
      deviceNumberList: [],
      value: '',
      modalVisible: false,
      protocolTypeParam: '',
      textArray: [
        '2）关系运算符："<"(小于)、">"(大于)、"小于等于"(<=)、 ">="(大于等于)、"=="(等于)、"!="(不等于)。',
        '3）逻辑运算符："&&"(逻辑与)、"||"(逻辑或)。',
        '2. 以"列头柜A路直流输入电压异常"告警为例，表达式输入内容为 "(A路直流输入电压<230)||(A路直流输入电压>285)"。',
        '1）数学运算符："+"(加)、"-"(减)、"*"(乘)、"/"(除)。',
        '由用户定义的一串文本，以尽量精简的文字说明当前发生的异常。告警类型和设备编号一起定义了某个设备所发生的一个告警，如 “SZ-TX-BD1-M201-MDC-A 市电A路A相电压异常”。告警类型会在告警发生时展示在告警相关的视图中，也会在发送给用户的通知中看到该信息，如“冷通道温度过高告警”，“机架PDU跳闸告警”。',
      ],

      occurExpressionList: [],
      restoreExpressionList: [],
      computeSymbol: ['>=', '==', '<=', '<', '>', '!='],
    };
  },
  computed: {
    chooseName() {
      return this.showSelectDevice ? '取消选择' : '选择设备';
    },

  },
  watch: {
    visible(val) {
      if (val === false) {
        this.occurExpressionList = [];
        this.restoreExpressionList = [];
      } else {
        this.modalVisible = val;
        this.getProtocol();
        Object.keys(this.modalData).forEach((item) => {
          if (item === 'deviceGidList') {
            this.deviceParam = this.modalData[item];
          }
        });
        Object.keys(this.form).forEach((item) => {
          if (item === 'deviceNumberList') {
            this.protocolTypeParam = this.modalData.protocolType;
            this.getList();
          } else {
            this.form[item] = this.modalData[item];
            if (item === 'occurExpression') {
              this.occurExpressionList = this.getExpressionList(this.modalData[item]);
            } else if (item === 'restoreExpression') {
              this.restoreExpressionList = this.getExpressionList(this.modalData[item]);
            }
          }
        });
      }
    },
  },
  mounted() {
  },
  methods: {
    getExpressionList(val) {
      let unit = null;
      let list = val.split();
      if (val.includes('&&')) {
        unit = '&&';
        list = val.split('&&');
      } else if (val.includes('||')) {
        unit = '||';
        list = val.split('||');
      }
      const { computeSymbol } = this;
      const newList = list?.map((e, index) => {
        const symbol = computeSymbol.find(item => e.includes(item));
        let value = null;
        let inputExtr = null;
        if (e.match(/\d/g)) {
          // eslint-disable-next-line prefer-destructuring
          value = e.split(symbol)[1];
          if (isNaN(value)) {
            const inputSymbol = ['+', '-', '*', '/', ')'].find(item => e.includes(item));
            inputExtr = inputSymbol + value.split(inputSymbol)[1];
            [value] = value.split(inputSymbol);
          }
        }
        return {
          unit: index === list.length - 1 ? null : unit || null,
          label: value ? e.split(symbol)[0] + symbol || '' : e,
          value,
          inputExtr,
        };
      });
      return newList;
    },
    getSubmitExpression(val) {
      const value = val.reduce((prev, cur) => prev + cur.label + cur.value + cur.inputExtr + cur.unit, '');
      return value.replace(/null/g, '');
    },
    selectDevice() {
      this.showSelectDevice = !this.showSelectDevice;
    },
    choosedDevice(val) {
      this.deviceParam = val;
      this.deviceNumberList = val;
    },
    changeProtocol(data) {
      this.protocolTypeParam = data;
      this.getList();
      this.deviceNumberList = [];
    },
    submit() {
      this.$refs.form.validate((valid) => {
        if (valid) {
          const { occurExpressionList, restoreExpressionList } = this;
          const params = { ...this.form };
          params.deviceGidList = this.deviceNumberList;
          delete params.deviceNumberList;
          params.occurExpression = this.getSubmitExpression(occurExpressionList);
          params.restoreExpression = this.getSubmitExpression(restoreExpressionList);
          getEdgeRequest(this.$axios, this.tableConfig.searchParams.mozuId).post(
            cgi.editCustom,
            { ...params, mozuId: this.mozuId || Number(window.__GetFrameDataByKey('curMozuData')?.id), id: parseInt(this.modalData.id) }
          )
            .then((data) => {
              if (!data.ok) {
                Object.keys(data.invalidations).forEach((item) => {
                  if (Object.prototype.hasOwnProperty.call(this.modalData, item) && this.$refs[item]) {
                    if (item === 'occurExpression' && data.invalidations.occurPointValue) {
                      data.invalidations.occurExpression += `&&&${data.invalidations.occurPointValue}`;
                    } else if (item === 'restoreExpression' && data.invalidations.restorePointValue) {
                      data.invalidations.restoreExpression += `&&&${data.invalidations.restorePointValue}`;
                    }
                    this.$refs[item].validateMessage = data.invalidations[item];
                    this.$refs[item].validateState = 'error';
                  }
                });
                return false;
              }
              this.$message.success('更新告警策略成功');
              this.modalVisible = false;
              this.tableConfig.refreshNow = !this.tableConfig.refreshNow;
              this.$emit('successchange');
            });
          console.log(this.form);
        }
      });
    },
    getProtocol() {
      getEdgeRequest(this.$axios, this.tableConfig.searchParams.mozuId).post(
        cgi.getProtocolTypeDropdown,
        { mozuId: this.mozuId }
        // { mozuId: 326, protocolType: form.protocolType, keyword: v }
      )
        .then((data) => {
          const list = [];
          Object.keys(data).forEach((item) => {
            const temp = {};
            temp.label = data[item];
            temp.value = item;
            list.push(temp);
          });
          this.protocolList = list;
        // this.$nextTick(() => {
        //   this.deviceNumberList = this.modalData.deviceGidList;
        // });
        });
    },
    deviceNumberListChange() {
      this.deviceParam = this.deviceNumberList;
    },
    handleChange() {

    },
    occurHeadChange(data) {
      this.form.occurExpression = this.form.occurExpression + data;
    },
    restoreHeadChange(data) {
      this.form.restoreExpression = this.form.restoreExpression + data;
    },
    headChange(data) {
      this.abc = this.abc + data;
    },
    getList(v) {
      getEdgeRequest(this.$axios, this.tableConfig.searchParams.mozuId).post(
        cgi.getDeviceListByProtocolType,
        { mozuId: this.mozuId, protocolType: this.protocolTypeParam || this.form.protocolType, keyword: v }
      )
        .then((data) => {
          const list = [];
          Object.keys(data).forEach((item) => {
            const temp = {};
            temp.label = data[item];
            temp.value = item;
            list.push(temp);
          });
          this.form.deviceNumberList = list;
          this.$nextTick(() => {
            this.deviceNumberList = this.modalData.deviceGidList;
          });
        });
    },
    closed() {
      this.modalVisible = false;
      this.$emit('update:visible', false);
      this.showSelectDevice = false;
      // this.form = {
      //   deviceNumberList: [],
      //   protocolType: [],
      //   alarmType: [],
      //   alarmLevel: '',
      //   occurExpression: '',
      //   occurPeriod: '',
      //   occurCount: '',
      //   restoreExpression: '',
      //   restorePeriod: '',
      //   restoreCount: '',
      //   alarmContent: '',
      //   influence: '',
      //   suggestion: '',
      // };
    },
  },
};
</script>
<style lang="scss">
#editForm {
  .must .el-form-item__label:before {
    content: '*';
    color: #ff3e00;
    margin-right: 4px;
  }
  .alarm-tip-label {
    .el-form-item__label {
      width: 160px !important;
    }
    .el-help-tip {
      transform: translate(0px, -1px);
    }
  }
}
.help-wrap {
  padding: 8px;
  .help-title {
    font-weight: bold;
  }
}
.expression-input {
  min-width: 50px;
  >input {
    text-align: center;
  }
}
.el-form-item.is-error {
  margin-bottom: -24px;
}
.el-form-item__error--inline.custom_error {
  margin-left: 0 !important;
  top: -12px !important;
}
</style>
