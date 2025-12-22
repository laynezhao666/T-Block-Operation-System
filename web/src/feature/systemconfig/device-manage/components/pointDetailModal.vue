
<template>
  <el-modal :visible.sync="modalVisible">
    <template slot="title">
      点位查看
    </template>
    <el-block padding>
      <base-title title="基本信息" />
      <div class="form">
        <div
          v-for="(info, index) in baseInfoMap[pointData.valtype]"
          :key="index"
          class="info-item"
        >
          <span class="label">{{ info.label }}: </span>
          <span
            v-if="info.type"
            class="value"
            :class="info.style"
          >
            {{ typeEnum[pointData[info.key]] }}
          </span>
          <span
            v-else
            class="value"
            :class="info.style"
          >{{
            Array.isArray(info.key) ?
              pointData[info.key[0]] + '-' + pointData[info.key[1]] :
              (pointData[info.key] || '--') }}</span>
        </div>
      </div>
      <base-title title="采集信息" />
      <div class="form">
        <div
          v-for="(info, index) in collectInfoMap[pointData.valtype]"
          :key="index"
          class="info-item"
        >
          <span class="label">{{ info.label }}: </span>
          <span class="value">{{ pointData[info.key] || "--" }}</span>
        </div>
      </div>
    </el-block>
  </el-modal>
</template>

<script>
import BaseTitle from './baseTitle';
export default {
  components: {
    BaseTitle,
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    pointInfo: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      baseInfo: [{
        label: '测点标识',
        key: 'no',
      }, {
        label: '测点名称',
        key: 'name',
      }, {
        label: '数据类型',
        key: 'valtype',
        type: 'enum',
      }, {
        label: '读写权限',
        key: 'access',
        style: 'button',
        type: 'enum',
      }, {
        label: '单位',
        key: 'unit',
        show: 'A',
      }, {
        label: '值描述',
        key: 'valdesc',
        show: 'E',
      }, {
        label: '有效值范围',
        key: ['minVal', 'maxVal'],
        show: 'A',
      }, {
        label: '变化死区',
        key: 'scale',
        show: 'A',
      }, {
        label: '取值方式',
        key: 'source',
      }],
      collectInfo: [
        {
          label: '功能码',
          key: 'cmd',
        },
        {
          label: '寄存器',
          key: 'reg',
        },
        {
          label: '解析公式',
          key: 'expression',
        },
        {
          label: '字节序',
          key: 'byteorder',
        },
        {
          label: '缩放因子',
          key: 'scale',
          show: 'A',
        },
        {
          label: '偏移量',
          key: 'offset',
          show: 'A',
        },
      ],
      pointData: {},
      typeEnum: {
        A: '浮点型',
        E: '枚举型',
        D: '状态量',
        R: '只读',
        W: '只写',
        RW: '读写',
      },
      baseInfoMap: {},
      collectInfoMap: {},
    };
  },
  computed: {
    modalVisible: {
      set(v) {
        this.$emit('update:visible', v);
      },
      get() {
        return this.visible;
      },
    },
  },
  watch: {
    pointInfo(v) {
      const info = {};
      _.forEach(v, (value, key) => {
        if (key === 'simulator') return;
        if (v[key] instanceof Object) {
          _.forEach(v[key], (value, key) => {
            info[key] = value;
          });
        } else {
          info[key] = value;
        }
      });
      this.pointData = info;
    },
  },
  mounted() {
    this.baseInfoMap = {
      A: this.baseInfo.filter(item => !item.show || item.show === 'A'),
      E: this.baseInfo.filter(item => !item.show || item.show === 'E'),
      D: this.baseInfo.filter(item => !item.show || item.show === 'E'),
    };
    this.collectInfoMap = {
      A: this.collectInfo.filter(item => !item.show || item.show === 'A'),
      E: this.collectInfo.filter(item => !item.show || item.show === 'E'),
      D: this.collectInfo.filter(item => !item.show || item.show === 'E'),
    };
  },
  methods: {
  },
};
</script>

<style lang="scss" scoped>
.form {
  padding-left: 20px;
  .info-item {
      padding: 5px 0 20px 0;
      .label {
        width: 80px;
        text-align: right;
        display: inline-block;
        margin-right: 12px;
        color: #bfbcbc;
        font-weight: 800;
      }
      .value {
        color: #333;
      }
      .button {
        display: inline-block;
        padding: 4px 5px;
        width: 40px;
        border: 1px solid;
        border-radius: 5px;
        text-align: center;
        border-color: #3ecc46;
        color: #3ecc46;
      }
  }
}
</style>
