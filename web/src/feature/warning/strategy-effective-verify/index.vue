<template>
  <div class="strategy-effective-verify">
    <div style="display:flex;margin-bottom:10px">
      <el-title style="width:200px;margin:16px 0 0 0">
        策略生效验证
      </el-title>
      <div style="flex:1; ">
        <el-button
          type="plain"
          class="invalid-history-button"
          @click="jumpUrl"
        >
          查看无效策略历史数据
        </el-button>
      </div>
    </div>
    <el-space
      direction="vertical"
      size="middle"
    >
      <!-- <el-block>
        <el-form
          ref="form"
          :model="form"
          label-width="75px"
          style="width:500px;display:flex"
        >
          <el-form-item
            label="模组"
          >
            <datascope-selector
              ref="socpeselector"
              class="socpeselector-valid"
              style="margin-top: 5px;"
              :with-all-scope="false"
              @change="changeMozu"
              @mozuloaded="scopeloaded"
            />
          </el-form-item>
        </el-form>
      </el-block> -->
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
        <div
          class="arrow-style"
        >
          <svg
            v-if="!showMoreDetail"
            t="1613810164495"
            style="width:18px;height:15px;cursor:pointer"
            class="icon"
            viewBox="0 0 1024 1024"
            version="1.1"
            xmlns="http://www.w3.org/2000/svg"
            p-id="4787"
            @click="showMoreDetail = true"
          ><path
            :d="upArrowSvg"
            fill="#231815"
            p-id="4788"
          /><path
            d="M29.381772 503.857563c151.214887 151.2185 302.431581 302.449643 453.660918 453.666336 15.992267
          16.008523 41.934997 16.008523 57.927264 0l453.662724-453.666336c14.988009-15.002459
          14.988009-39.315979 0-54.303988-15.000653-15.000653-39.315979-15.000653-54.303988
          0a13680362.528566 13680362.528566 0 0 1-428.326884 428.321464C369.22377 735.109647 226.46199
          592.331611 83.683954 449.553575c-15.000653-15.000653-39.315979-15.000653-54.302182 0-15.000653
          14.988009-15.000653 39.301529 0 54.303988z"
            fill="#231815"
            p-id="4789"
          /></svg>
          <svg
            v-else
            t="1613811506022"
            class="icon"
            viewBox="0 0 1024 1024"
            version="1.1"
            xmlns="http://www.w3.org/2000/svg"
            p-id="5199"
            style="width:18px;height:15px;cursor:pointer"
            @click="showMoreDetail = false"
          ><path
            d="M29.381772 903.962267l453.660918-453.666336c15.994073-15.994073 41.934997-15.994073
          57.927264 0l453.662724 453.666336c15.000653 15.002459 15.000653 39.315979 0 54.305794-14.988009
          15.000653-39.303336 15.000653-54.303988
          0-142.77623-142.767199-285.55246-285.545235-428.316046-428.323271-142.77623 142.778036-285.55246
          285.556072-428.32869 428.323271-15.002459 15.000653-39.303336
          15.000653-54.303988 0-15.000653-14.989815-15.000653-39.305142 0.001806-54.305794z"
            fill="#231815"
            p-id="5200"
          /><path
            d="M29.381772 520.142437c151.214887-151.2185 302.431581-302.449643 453.660918-453.668142
          15.994073-15.994073 41.922354-15.994073 57.927264 0 151.216693 151.216693 302.433387 302.449643
          453.662724 453.668142 14.986203 15.000653 14.986203 39.315979 0 54.303988-15.000653
          15.000653-39.315979 15.000653-54.303988
          0-142.77623-142.778036-285.55246-285.543429-428.32869-428.321464L83.683954 574.446425c-15.002459
           15.000653-39.315979 15.000653-54.303988 0-15.000653-14.988009-15.000653-39.303336 0.001806-54.303988z"
            fill="#231815"
            p-id="5201"
          /></svg>
        </div>
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
              @updatedata="getOverview()"
            />
          </el-tab-pane>
        </el-tabs>
      </el-block>
    </el-space>
  </div>
</template>

<script>
import invalidConfig from './invalid-strategy-config';
import validConfig from './valid-strategy-config';
import { warning as cgi } from '@@/config/cgi';
import strategyTable from './strategy-table';
import { eventBus } from '../component/commonTable/script/eventBus';
import getEdgeRequest from '../../utils/request';

export default {
  components: {
    // commonTable,
    strategyTable,
  },
  provide() {
    return {
      configCgi: cgi,
    };
  },
  data() {
    return {
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
    };
  },
  mounted() {
    this.mozuId = parseInt(TNBL.getCurModuleId());
    this.initPage();
  },

  beforeDestroy() {
    eventBus.$off('showModal');
  },
  methods: {
    initPage() {
      eventBus.$on('showModal', ({ type, data }) => {
        if (type === '详情') {
          const valid = this.activeName === 'valid' ? '1' : '0';

          window.open(`/timpage/warning-strategy-detail?valid=${valid}&deviceGid=${data.deviceGid}&ruleId=${data.ruleId}&validateType=${data.validateType}&mozuId=${this.mozuId}`);
        }
      });
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
      padding-left:40px
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
