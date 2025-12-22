<template>
  <!--
    feat-sanbox分支对这个文件修改过，但是没有合并到主分支，直接迁移/合并有什么影响未知。
    （需要搜索查阅各文件分析影响的页面、找对应的人了解背景、以确定是否可以合并为一个文件）
    本次迁移，直接拷贝一份完成任务，请见谅
  -->
  <div class="strategy-effective-verify">
    <div style="display:flex;margin-bottom:10px">
      <el-title style="width:200px;margin:16px 0 0 0">
        策略生效验证
      </el-title>
    </div>
    <el-space
      v-if="hasRight"
      direction="vertical"
      size="middle"
    >
      <el-block>
        <div style="display:flex;width:100%;padding:16px 0">
          <div style="display:flex;width:130px;align-items:center">
            <div
              class="data-patch-left"
            >
              汇总
            </div>
          </div>
          <div
            class="data-patch-right"
          >
            <el-data-patch
              title="总部署策略实例"
              :value="totalInfo.total"
            />
            <el-data-patch
              title="有效策略实例"
              :value="totalInfo.valid"
            />
            <el-data-patch
              title="无效策略实例"
              :value="totalInfo.invalid"
            />
            <el-data-patch
              title="生效率"
              :value="totalInfo.validPercent"
            />
          </div>
        </div>
        <!-- <transition name="bounce"> -->
        <div
          :class="showMoreDetail ? 'showMore': 'hideMore'"
        >
          <div class="data-patch-block">
            <div
              class="data-patch-wrap"
            >
              <div
                class="data-patch-left"
              >
                标准策略
              </div>
            </div>
            <div
              class="data-patch-right"
            >
              <el-data-patch
                title="总部署策略实例"
                :value="standardInfo.total"
              />
              <el-data-patch
                title="有效策略实例"
                :value="standardInfo.valid"
              />
              <el-data-patch
                title="无效策略实例"
                :value="standardInfo.invalid"
              />
              <el-data-patch
                title="生效率"
                :value="standardInfo.validPercent"
              />
            </div>
          </div>
          <div style="height:20px;background-color:#fbfbfb" />
          <div class="data-patch-block">
            <div
              class="data-patch-wrap"
            >
              <div
                class="data-patch-left"
              >
                自定义策略
              </div>
            </div>
            <div
              class="data-patch-right"
            >
              <el-data-patch
                title="总部署策略实例"
                :value="customInfo.total"
              />
              <el-data-patch
                title="有效策略实例"
                :value="customInfo.valid"
              />
              <el-data-patch
                title="无效策略实例"
                :value="customInfo.invalid"
              />
              <el-data-patch
                title="生效率"
                :value="customInfo.validPercent"
              />
            </div>
          </div>
        </div>
        <!-- </transition> -->
      </el-block>
      <el-block
        no-padding
      >
        <el-tabs
          v-model="activeName"
          @tab-click="handleClick"
        >
          <el-tab-pane
            :label="'无效策略实例（'+totalInfo.invalid+'）'"
            name="invalid"
            lazy
          >
            <strategy-table
              v-if="activeName==='invalid'"
              :mozu-id="mozuId"
              :config="invalidColumns"
              :config-cgi="invalidCgi"
              tab-name="无效策略实例"
            />
          </el-tab-pane>
          <el-tab-pane
            :label="'有效策略实例（'+totalInfo.valid+'）'"
            name="valid"
            lazy
          >
            <strategy-table
              v-if="activeName==='valid'"
              :extra-data="totalInfo"
              :mozu-id="mozuId"
              :config="validColumns"
              :config-cgi="validCgi"
              tab-name="有效策略实例"
              @updatedata="validTableClick"
            />
          </el-tab-pane>
        </el-tabs>
      </el-block>
    </el-space>

    <defaultPage
      v-else
      title="该模组暂无沙箱环境，请联系管理员配置"
    />
  </div>
</template>

<script>
import invalidConfig from './invalid-strategy-config';
import validConfig from './valid-strategy-config';
import { warning as cgi } from '@@/config/cgi';
import strategyTable from './strategy-table';
import { eventBus } from '../component/commonTable/script/eventBus';
import getEdgeRequest from '../../utils/request';
import { getMozuId } from 'feature/utils/business';
import business from '@@/config/business';
import defaultPage from 'feature/component/defaultPage';

export default {
  components: {
    // commonTable,
    strategyTable,
    defaultPage,
  },
  provide() {
    return {
      configCgi: cgi,
    };
  },
  data() {
    return {
      business,
      hasRight: false,
      mozuloaded: false,
      upArrowSvg: 'M29.381772 120.037733l453.660918 453.666336c15.992267 15.994073 41.934997 15.994073 57.927264 0L994.632678 120.037733c15.000653-15.002459 15.000653-39.315979 0-54.303988-14.988009-15.000653-39.301529-15.000653-54.303988 0-142.774424 142.778036-285.55246 285.543429-428.31424 428.321465-142.778036-142.778036-285.55246-285.543429-428.330496-428.321465-15.000653-15.000653-39.301529-15.000653-54.302182 0-15.000653 14.988009-15.000653 39.303336 0 54.303988z',
      triggerWarning: true,
      showMoreDetail: false,
      activeName: 'invalid',
      validColumns: validConfig,
      invalidColumns: invalidConfig,
      validCgi: {
        queryCgi: cgi.getValidList,
        exportCgi: cgi.exportValidList,
      },
      invalidCgi: {
        queryCgi: cgi.getInvalidList,
        exportCgi: cgi.exportInValidList,
      },
      cgi,
      customInfo: {
        total: '',
        valid: '',
        invalid: '',
        validPercent: '',
      },
      standardInfo: {
        total: '',
        valid: '',
        invalid: '',
        validPercent: '',
      },
      totalInfo: {
        total: '',
        valid: '',
        invalid: '',
        validPercent: '',
      },
      form: {},
      mozuId: 326,
      mozuName: [],
      options: [],
      valid_type: 0
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
          this.hasRight = !!getMozuId();
          this.initEdgePage();
        }
      },
      immediate: true,
    },
  },
  mounted() {
    this.hasRight = !!getMozuId();
    if (!this.hasRight) return;
    if (!this.business.showModuleSelected) {
      // this.initEdgePage(); // 会造成两次刷新
      this.mozuId = getMozuId();
      this.initPage();
    } else {
      this.mozuId = parseInt(TNBL.getCurModuleId());
      this.initPage();
    }
  },

  beforeDestroy() {
    eventBus.$off('showModal');
  },
  methods: {
    validTableClick(v) {
      this.valid_type = v
      this.getOverview();
    },
    initPage() {
      eventBus.$on('showModal', ({ type, data }) => {
        if (type === '详情') {
          const valid = this.activeName === 'valid' ? '1' : '0';

          window.open(`/tedge/warning-strategy-detail?valid=${valid}&deviceGid=${data.deviceGid}&ruleId=${data.ruleId}&validateType=${data.validateType}&mozuId=${this.mozuId}&alarmType=${data.alarmType}&valid_type=${this.valid_type}`);
        }
      });
      this.getOverview();
    },
    initEdgePage() {
      eventBus.$on('showModal', ({ type, data }) => {
        if (type === '详情') {
          const valid = this.activeName === 'valid' ? '1' : '0';

          window.open(`/tedge/warning-strategy-detail?valid=${valid}&deviceGid=${data.deviceGid}&ruleId=${data.ruleId}&validateType=${data.validateType}&mozuId=${this.mozuId}&alarmType=${data.alarmType}&valid_type=${this.valid_type}`);
        }
      });
      this.mozuId = getMozuId();
      if (!this.curMozuData || !this.curMozuData.alarmapi) {
        return;
      }
      this.getOverview();
    },
    // scopeloaded(val) {
    //   this.mozuloaded = true;
    //   this.mozuId = val.id;
    //   this.initPage();
    // },
    // changeMozu(val) {
    //   // this.$set(this.tableConfig.searchParams, 'mozuId', val.id);
    //   this.mozuId = val.id;
    //   // this.tableConfig.refreshNow = !this.tableConfig.refreshNow;
    //   this.getOverview();
    //   console.log(val);
    // },
    jumpUrl() {
      window.open('/timpage/strategy-invalid-history');
    },
    getOverview() {
      window.cgi = cgi;
      getEdgeRequest(this.$axios, this.mozuId).post(cgi.getValidOverview, { mozuId: this.mozuId })
        .then((data) => {
          console.log(data, 'strategy');
          Object.keys(data).forEach((item) => {
            data[item].validPercent = `${parseFloat(data[item].validPercent * 100).toFixed(3)}%`;
          });
          this.totalInfo = data.total;
          this.customInfo = data.custom;
          this.standardInfo = data.standard;
        });
    },
    hasWarning() {
      this.triggerWarning = true;
    },
    handleClick() {
      this.valid_type = this.activeName === 'invalid' ? 0 : 1
    },
    handleChange() {

    },
    haventWarning() {
      this.triggerWarning = false;
    },
  },
};
</script>
<style lang="scss" scoped>
  .data-patch-block {
    display:flex;
    width:100%;
    background-color:#f6f6f6;
    padding: 16px 0;
  }
  .invalid-history-button {
    float:right;
    margin-top:16px;
    color:#1470CC;
    border-color:#1470CC
  }
  .arrow-style {
    height:24px;
    background-color: #f6f6f6;
    text-align:center;
    margin:5px 0px 5px 0;
    padding-top:5px
  }
  .data-patch-left {
      text-align:left;
      width:100%;
      height:24px;
      font-size:16px;
      font-weight:700;
      padding-left:40px;
      line-height: 24px;
  }
  .data-patch-right {
      flex:1;
      display: flex;
      justify-content: space-between;
      margin-right: 40px;
  }
  .data-patch-wrap {
    display:flex;
    width:130px;
    align-items:center
  }
  .showMore {
    max-height: 400px;
    transition: max-height .3s ease-in;
    transform-origin: 50% 0;
    animation: slide-down 0.3s ease-in;
  }

  .hideMore {
    max-height: 0px;
    overflow: hidden;
    transition: max-height .3s ease-out;
  }
  /deep/ .el-data-patch__title {
    margin-bottom: 0;
  }

  @keyframes slide-down{
    0%{transform:scale(1,0);}
    100%{transform:scale(1,1);}
  }

</style>

<style lang="scss">
  #mozu-cascader {
    margin-top: -10px;
    width:400px;
  //  .el-input::before{
  //     content:none;
  //   }
  //   .el-input::after{
  //     content:none;
  //   }
  }
  .strategy-effective-verify{
    .el-data-patch__value {
      font-size: 20px;
      font-weight: 400;
      font-family: 'MicrosoftYaHei', 'Microsoft YaHei', sans-serif;
      font-style: normal;
    }
    .el-block__body {
      padding: 0;
    }
  }
  .socpeselector-valid{
    .el-input::before{
      content: "";
      position: absolute;
      left: 0;
      right: 0;
      bottom: 0;
      -webkit-transition: border-bottom-color 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
      -o-transition: border-bottom-color 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
      transition: border-bottom-color 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
      border-bottom: 1px solid #999;
      pointer-events: none;
      width:96%
    }
    .el-input::after{
      content: "";
      position: absolute;
      left: 0;
      right: 0;
      bottom: 0;
      -webkit-transition: border-bottom-color 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
      -o-transition: border-bottom-color 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
      transition: border-bottom-color 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
      border-bottom: 1px solid #999;
      pointer-events: none;
    }
  }
</style>
