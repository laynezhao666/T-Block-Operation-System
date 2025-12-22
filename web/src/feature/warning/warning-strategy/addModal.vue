<template>
  <el-modal
    :visible.sync="modalVisible"
    :width="850"
    @closed="closed"
  >
    <template slot="title">
      新增告警策略
    </template>
    <el-form
      id="addForm"
      ref="form"
      :model="form"
      label-width="180px"
      :rules="rule"
    >
      <el-form-item
        ref="protocolType"
        label="设备类型"
        class="must"
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
          <!-- :remote-method="v => getList(v)" -->

          <el-option
            v-for="item in protocolList"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item
        ref="deviceGidList"
        label="设备编号"
        class="must"
      >
        <div style="display:flex; height: 56px; align-items: center;">
          <el-select
            v-model="deviceNumberList"
            disabled
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
      </el-form-item>
      <el-form-item
        label="告警等级"
        prop="alarmLevel"
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
      </el-form-item>
      <el-form-item
        ref="occurExpression"
        class="alarm-tip-label"
        label="触发表达式"
        prop="occurExpression"
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
        <div style="display:flex; height: 56px; align-items: center;">
          <el-input
            v-model="form.occurExpression"
            placeholder="请输入"
            style="width:430px"
          />
          <head-selector
            type="occurExpression"
            :url="queryPointUrl"
            :params="{deviceGidList: deviceParam}"
            :filter="filter"
            ajax-method="post"
            label-name="添加测点"
            @change="occurHeadChange"
          />
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
      >
        <div style="display:flex; height: 56px; align-items: center;">
          <el-input
            v-model="form.restoreExpression"
            placeholder="请输入"
            style="width:430px"
          />
          <head-selector
            type="restoreExpression"
            :url="queryPointUrl"
            :params="{deviceGidList: deviceParam}"
            :filter="filter"
            ajax-method="post"
            label-name="添加测点"
            @change="restoreHeadChange"
          />
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
        id="alarm-content"
        ref="alarmContent"
        class="alarm-tip-label"
        prop="alarmContent"
      >
        <template
          slot="label"
        >
          告警内容
          <el-help-tip
            popper-class="test"
          >
            <p><span v-pre>当前单体电池温度为{{温度}},超过正常34</span></p>
          </el-help-tip>
        </template>
        <el-input
          v-model="form.alarmContent"
          placeholder="请输入"
        />
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
    </template>
  </el-modal>
</template>
<script>
import headSelector from './head-selector';
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
    headSelector,
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
      queryPointUrl: cgi.queryPointTypeList,
      alarmLevel: ['L0', 'L1', 'L2', 'L3', 'L4', 'L5'],
      filter: {},
      num1: 0,
      abc: '',
      radio: '1',
      protocolList: [],
      form: {
        deviceNumberList: [],
        protocolType: [],
        alarmType: '',
        alarmLevel: '',
        occurExpression: '',
        occurPeriod: 1,
        occurCount: 1,
        restoreExpression: '',
        restorePeriod: 1,
        restoreCount: 1,
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
        alarmContent: [
          { required: true, message: '请输入', trigger: 'change' },
        ],
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
      textArray: [
        '2）关系运算符："<"(小于)、">"(大于)、"小于等于"(<=)、 ">="(大于等于)、"=="(等于)、"!="(不等于)。',
        '3）逻辑运算符："&&"(逻辑与)、"||"(逻辑或)。',
        '2. 以"列头柜A路直流输入电压异常"告警为例，表达式输入内容为 "(A路直流输入电压<230)||(A路直流输入电压>285)"。',
        '1）数学运算符："+"(加)、"-"(减)、"*"(乘)、"/"(除)。',
        '由用户定义的一串文本，以尽量精简的文字说明当前发生的异常。告警类型和设备编号一起定义了某个设备所发生的一个告警，如 “SZ-TX-BD1-M201-MDC-A 市电A路A相电压异常”。告警类型会在告警发生时展示在告警相关的视图中，也会在发送给用户的通知中看到该信息，如“冷通道温度过高告警”，“机架PDU跳闸告警”。',
      ],
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
        this.$nextTick(() => {
          this.form = {
            deviceNumberList: [],
            protocolType: [],
            alarmType: '',
            alarmLevel: '',
            occurExpression: '',
            occurPeriod: 1,
            occurCount: 1,
            restoreExpression: '',
            restorePeriod: 1,
            restoreCount: 1,
            alarmContent: '',
            influence: '',
            suggestion: '',
          };
          this.deviceNumberList = [];
        });
      } else {
        this.modalVisible = val;
        // Object.keys(this.form).forEach((item) => {
        //   if (item === 'deviceNumberList') {
        //   // const tempDeviceNumberList = [];
        //   // this.modalData[item].forEach((item) => {
        //   //   const tempJson = {};
        //   //   tempJson.label = item;
        //   //   tempJson.value = 1;
        //   //   tempDeviceNumberList.push(tempJson);
        //   // });
        //   // this.deviceNumberList = this.modalData.deviceGidList.map(item => item.toString());
        //   // this.form.deviceNumberList = tempDeviceNumberList;
        // this.getList();
        this.$nextTick(() => {
          this.getProtocol();
        });
        //   } else {
        //     this.form[item] = this.modalData[item];
        //   }
        // });
      }
    },
  },
  mounted() {

  },
  methods: {
    selectDevice() {
      this.showSelectDevice = !this.showSelectDevice;
    },
    choosedDevice(val) {
      this.deviceParam = val;
      this.deviceNumberList = val;
    },
    changeProtocol() {
      this.$nextTick(() => {
        this.getList();
        this.deviceNumberList = [];
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
    submit() {
      this.$refs.form.validate((valid) => {
        if (valid) {
          const params = this.form;
          params.deviceGidList = this.deviceNumberList;
          // params.protocolType = 'SPM';
          getEdgeRequest(
            this.$axios,
            this.tableConfig.searchParams.mozuId
          ).post(cgi.addCustom, { ...params, mozuId: this.mozuId })
            .then((data) => {
              if (!data.ok) {
                Object.keys(data.invalidations).forEach((item) => {
                  if (this.$refs[item]) {
                    this.$refs[item].validateMessage = data.invalidations[item];
                    this.$refs[item].validateState = 'error';
                  }
                });
                return;
              }

              this.$message.success('新增告警策略成功');
              this.modalVisible = false;
              this.tableConfig.refreshNow = !this.tableConfig.refreshNow;
              this.$emit('successchange');
            });
          console.log(this.form);
        }
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
    getList(v) {
      getEdgeRequest(this.$axios, this.tableConfig.searchParams.mozuId).post(
        cgi.getDeviceListByProtocolType,
        { mozuId: this.mozuId, protocolType: this.form.protocolType || 'SPM', keyword: v }
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
          this.form.deviceNumberList = list;
        // this.$nextTick(() => {
        //   this.deviceNumberList = this.modalData.deviceGidList;
        // });
        });
    },
    closed() {
      this.modalVisible = false;
      this.$emit('update:visible', false);
      this.showSelectDevice = false;
    },
  },
};
</script>
<style lang="scss">
#addForm {
  .must .el-form-item__label:before{
    content: '*';
    color: #ff3e00;
    margin-right: 4px;
  }
  .alarm-tip-label {
    .el-form-item__label {
      width: 160px !important;
    }
    .el-help-tip{
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
</style>
