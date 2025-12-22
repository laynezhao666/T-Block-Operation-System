
<template>
  <div class="port-log">
    <div class="operate-container">
      <el-row style="line-height: 32px; ">
        <el-col :span="2">
          <div class="label">
            过滤选项:
          </div>
        </el-col>
        <el-col
          :span="22"
          style="display: flex;"
        >
          <div
            class="filter"
            style="flex: 1;"
          >
            <el-select
              v-model="portType"
              placeholder="端口类型"
              border-type="bordered"
            >
              <el-option
                v-for="item in portTypes"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-select
              v-model="portNumber"
              placeholder="端口号"
              border-type="bordered"
            >
              <el-option
                v-for="item in portNumbers"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-select
              v-model="devicePosition"
              placeholder="设备地址"
              border-type="bordered"
            >
              <el-option
                v-for="item in devicePositions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-select
              v-model="logType"
              placeholder="全部"
              border-type="bordered"
            >
              <el-option
                v-for="item in logTypes"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </div>
          <div
            class="operate"
            style="text-align: right;"
          >
            <el-button type="text">
              启动/暂停
            </el-button>
            <el-button type="text">
              另存为
            </el-button>
            <el-button type="text">
              清空
            </el-button>
          </div>
        </el-col>
      </el-row>
    </div>
    <div
      ref="log"
      class="log"
    >
      <el-row style="line-height: 32px; ">
        <el-col :span="2">
          <div class="label">
            侦听日志:
          </div>
        </el-col>
        <el-col :span="22">
          <div
            class="log-container"
            :style="{ height: `${logHeight}px` }"
          >
            <div>2022-10-10 09:34:46【TX】：</div>
            <div>01 03 00 00 00 02 【CRC】</div>
            <div>2022-10-10 09:34:46【RX】：</div>
            <div>01 03 00 02 00 64 【CRC】</div>
          </div>
          <div>成功：xx次，失败：xx次</div>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      portType: '',
      portTypes: [
        {
          value: '串口',
          label: '串口',
        },
        {
          value: '网口',
          label: '网口',
        },
      ],
      portNumber: '',
      portNumbers: [
        {
          value: 'COM1',
          label: 'COM1',
        },
        {
          value: 'COM2',
          label: 'COM2',
        },
        {
          value: 'COM3',
          label: 'COM3',
        },
      ],
      devicePosition: '',
      devicePositions: [
        {
          value: '1',
          label: '1',
        },
        {
          value: '2',
          label: '2',
        },
      ],
      logType: 'all',
      logTypes: [
        {
          value: 'all',
          label: '全部',
        },
        {
          label: '仅看请求',
          value: 'request',
        },
        {
          label: '仅看返回',
          value: 'response',
        },
        {
          label: '查看失败',
          value: 'fail',
        },
      ],
      logHeight: 0,
    };
  },
  mounted() {
    this.logHeight = window.innerHeight - this.$refs.log.offsetTop - 32 - 32 - 32;
  },
};
</script>

<style lang="scss" scoped>
.port-log {
  padding: 32px 32px 0 0;
  color: #333;
  display: flex;
  flex-direction: column;
  .filter {
    display: flex;
    .el-select {
      width: 130px;
      margin-right: 10px;
    }
    /deep/ .el-select .el-input__inner {
      border-radius: 2px;
    }
  }
  .label {
    font-family: TencentSansW3;
    color: #333;
    font-weight: 800;
    text-align: right;
    padding-right: 16px;
  }
  .log {
    flex: 1;
    padding: 16px 0;
    .log-container {
      background: #f5f5f5;
      height: 100%;
      border-radius: 5px;
      border: 1px solid #e4e4e4;
      color: #56BD06;
      font-size: 14px;
      overflow: scroll;
      padding: 10px;
      font-family: 'ArialMT', 'Arial', sans-serif;
    }
  }
}
</style>
