<template>
  <div>
    <el-col :span="18">
      <el-row>
        <el-col :span="24">
          <el-card class="box-card">
            <div
              slot="header"
              class="card-header"
            >
              <span>收敛时间线</span>
              <el-date-picker
                v-model="value1"
                style="width:150px;margin-left:auto"
                placeholder="选择日期"
              />
            </div>
            <div
              ref="chart1"
              style="height:300px"
            />
          </el-card>
        </el-col>
      </el-row>
      <el-row>
        <el-col :span="24">
          <el-card class="box-card">
            <div
              slot="header"
              class="card-header"
            >
              <span>收敛列表</span>
              <el-button
                type="text"
                style="margin-left: auto;"
              >
                查看全部
              </el-button>
            </div>
            <el-table
              :data="tableData"
              border
              style="width: 100%"
            >
              <el-table-column
                prop="date"
                label="风暴时间"
                width="180"
              />
              <el-table-column
                prop="n1"
                label="收敛ID"
                width="180"
              />
              <el-table-column
                prop="n2"
                label="收敛根设备"
              />
              <el-table-column
                prop="n3"
                label="主告警描述"
              />
              <el-table-column
                prop="n4"
                label="收敛数量"
              />
            </el-table>
          </el-card>
        </el-col>
        <el-col :span="24">
          <el-card class="box-card">
            <div
              slot="header"
              class="card-header"
            >
              <span>收敛路径</span>
            </div>
            <el-col
              :span="6"
            >
              <el-button>
                告警风暴时间（ID）
              </el-button>
              <div style="display:flex; flex-direction: column ;max-height:500px;overflow-y:auto">
                <el-button
                  type="info"
                  style="margin-left:0px"
                >
                  2021-06-07 21:33:44.555
                </el-button>
                <el-button
                  type="warning"
                  style="margin-left:0px"
                >
                  2021-06-07 21:33:44.555
                </el-button>
                <el-button
                  v-for="i in 20"
                  :key="i"
                  type="info"
                  style="margin-left:0px"
                >
                  2021-06-07 21:33:44.555{{ i }}
                </el-button>
              </div>
            </el-col>
            <el-col :span="18">
              <div
                ref="chart"
                style="height:500px"
              />
            </el-col>
          </el-card>
        </el-col>
      </el-row>
    </el-col>
    <el-col :span="6">
      <el-row>
        <el-col :span="24">
          <el-card class="box-card">
            <div
              slot="header"
              class="card-header"
            >
              <span>实时告警</span>
              <el-button
                type="text"
                style="margin-left: auto;"
              >
                查看全部
              </el-button>
            </div>
            <div style="height:300px;overflow-y:auto;">
              <div
                v-for="o in 20"
                :key="o"
                class="text item"
              >
                2021-05-15 11:24:59   综保备自投充电完成...
                <el-divider />
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="24">
          <el-card class="box-card">
            <div
              slot="header"
              class="card-header"
            >
              <span>当前告警设备实时工况</span>
              <el-button
                type="text"
                style="margin-left: auto;"
              >
                定位设备
              </el-button>
            </div>
            <el-collapse
              accordion
              value="2"
            >
              <el-collapse-item name="2">
                <template slot="title">
                  高压进线柜1
                  <el-tag type="warning">
                    420条告警
                  </el-tag>
                </template>
                <el-form
                  ref="form"
                  label-width="120px"
                  size="small"
                >
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                </el-form>
              </el-collapse-item>
              <el-collapse-item name="2">
                <template slot="title">
                  高压进线柜2
                  <el-tag type="warning">
                    420条告警
                  </el-tag>
                </template>
                <el-form
                  ref="form"
                  label-width="120px"
                  size="small"
                >
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                </el-form>
              </el-collapse-item>
              <el-collapse-item name="2">
                <template slot="title">
                  高压进线柜2
                  <el-tag type="warning">
                    420条告警
                  </el-tag>
                </template>
                <el-form
                  ref="form"
                  label-width="120px"
                  size="small"
                >
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                  <el-form-item label="仪表A相电压">
                    5.4KV
                  </el-form-item>
                </el-form>
              </el-collapse-item>
            </el-collapse>
          </el-card>
        </el-col>
      </el-row>
    </el-col>
  </div>
</template>
<script>
import { initChart, initLine } from './echardata';

export default {
  props: {
  },
  data() {
    return {
      value1: '2021-06-07',
      tableData: [{
        date: '2021-06-07 20:14',
        n1: 'sl-234234234',
        n2: '高压进线柜（QY-BD-XX-EES-EFSEF-SEF-SEF1）',
        n3: '综保备自投充电完成',
        n4: 3941,
      }, {
        date: '2021-06-07 15:14',
        n1: 'sl-234234234',
        n2: '高压进线柜（QY-BD-XX-EES-EFSEF-SEF-SEF1）',
        n3: '综保备自投充电完成',
        n4: 3941,
      }],
      currentRow: null,

    };
  },
  mounted() {
    this.$nextTick().then(() => {
      initLine(this.$refs.chart1);
      initChart(this.$refs.chart);
    });
  },
  methods: {
    setCurrent() {
      this.$refs.singleTable.setCurrentRow(1);
    },
    handleCurrentChange(val) {
      this.currentRow = val;
    },
  },
};
</script>
<style lang="scss" scoped>
  .card-header {
    display: flex;
    align-items: center;
    font-weight: 900;
  }
</style>
