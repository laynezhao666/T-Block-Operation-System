<template>
  <div class="advanced-search">
    <el-title>
      <span
        v-if="title"
      >
        【{{ title }}】历史数据高级查询
      </span>
      <span v-else>
        历史数据高级查询
      </span>
    </el-title>

    <el-block>
      <div class="advanced-search-container">
        <div class="advanced-search-aside">
          <div class="advanced-search-check-all">
            <el-checkbox
              v-if="pointList.length"
              v-model="checkAll"
              :indeterminate="isIndeterminate"
              @change="handleCheckAllChange"
            >
              测点列表
            </el-checkbox>
            <el-button
              type="primary"
              :disabled="pointList.length >= 50"
              @click="pointModalVisible = true"
            >
              添加测点
            </el-button>
            <el-button
              type="text"
              style="margin-left: auto"
              @click="yMainAxiosDialogVisible = true"
            >
              选项
            </el-button>
          </div>
          <div
            v-if="!pointList.length"
            class="advanced-search-empty-point-list"
          >
            暂无测点
          </div>
          <el-checkbox-group
            v-model="checkedPointList"
            class="advanced-search-checkbox-list"
            @change="handleCheckedPointListChange"
          >
            <el-checkbox
              v-for="(e, i) in pointList"
              :key="i"
              :label="e"
            >
              <el-tooltip
                effect="dark"
                :content="e"
                placement="top-start"
              >
                <span class="point-name">{{ e }}</span>
              </el-tooltip>

              <el-button
                type="text"
                size="mini"
                @click="openYAxisOptionDialog(e)"
              >
                选项
              </el-button>
            </el-checkbox>
          </el-checkbox-group>
        </div>

        <div class="advanced-search-main">
          <div class="advanced-search-toolbar">
            <el-space :size="24">
              <el-date-picker
                v-model="filter.timerange"
                class="advanced-search-toolbar-date-picker"
                type="datetimerange"
                range-separator="至"
                :format="{
                  date: 'yyyy-MM-dd',
                  time: 'HH:mm:ss'
                }"
                value-format="yyyy-MM-dd HH:mm:ss"
                start-placeholder="开始时间"
                end-placeholder="结束时间"
                :picker-options="pickerOptions"
                @change="timerangeChangeHandler"
              />

              <div class="advanced-search-toolbar-time">
                每
                <el-input
                  v-model="filter.duration"
                  class="advanced-search-toolbar-time-value"
                  clearable
                />
                <el-select
                  v-model="filter.unit"
                  class="advanced-search-toolbar-time-unit"
                  placeholder="时间"
                  no-input
                >
                  <el-option
                    v-for="item in unitOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </div>

              <el-select
                v-model="filter.stats"
                class="advanced-search-toolbar-value-type"
                placeholder="计算"
                multiple
                clearable
              >
                <el-option
                  v-for="item in statOption"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>

              <el-button
                type="primary"
                :disabled="!checkedPointList.length"
                @click="getData"
              >
                查询
              </el-button>
            </el-space>

            <div class="advanced-search-toolbar-template">
              <el-dropdown
                trigger="click"
                @command="handleCommand"
              >
                <span class="advanced-search-toolbar-template-btn">
                  模板管理<i class="tn-icon-drop" />
                </span>
                <el-dropdown-menu slot="dropdown">
                  <el-dropdown-item command="loadTemplate">
                    <tn-icon icon="import" />加载模板
                  </el-dropdown-item>
                  <el-dropdown-item command="manageTemplate">
                    <tn-icon icon="config" />管理模板
                  </el-dropdown-item>
                  <el-dropdown-item
                    :disabled="!checkedPointList.length"
                    command="saveTemplate"
                  >
                    <tn-icon icon="save" />保存为新模板
                  </el-dropdown-item>
                  <el-dropdown-item
                    v-if="currentTemplate.templateName"
                    command="editTemplate"
                  >
                    <tn-icon icon="save" />覆盖当前模板
                  </el-dropdown-item>
                </el-dropdown-menu>
              </el-dropdown>
            </div>
          </div>

          <div
            class="advanced-search-main-content"
          >
            <el-block
              inner
              border
            >
              <div
                ref="chart"
                :style="{height:autoHeight+'px'}"
                class="advanced-search-chart"
              />
            </el-block>

            <el-table-toolbar
              v-if="tableData.length"
              hide-search
            >
              <template #extra>
                <el-button
                  @click="exportHistory"
                >
                  导出
                </el-button>
              </template>
            </el-table-toolbar>

            <div
              v-if="tableData.length"
              class="advanced-search-table"
            >
              <el-table
                v-if="statTableData.length"
                :data="statTableData"
                style="width: 100%"
              >
                <el-table-column
                  prop="stat"
                  width="200"
                  fixed
                />
                <el-table-column
                  v-for="(column, index) in tableColumns"
                  :key="index"
                  v-bind="column"
                  min-width="360"
                >
                  <template
                    #header
                  >
                    <el-tooltip
                      class="th-point-name"
                      effect="dark"
                      :content="column.label"
                      placement="top"
                    >
                      <span>{{ column.label }}</span>
                    </el-tooltip>
                  </template>
                </el-table-column>
              </el-table>

              <el-table
                :data="tableData"
                style="width: 100%; margin-top: -1px;"
                height="546"
              >
                <el-table-column
                  prop="updateTime"
                  label="时间"
                  width="200"
                  fixed
                  sortable
                />
                <el-table-column
                  v-for="(column, index) in tableColumns"
                  :key="index"
                  v-bind="column"
                  min-width="360"
                >
                  <template
                    #header
                  >
                    <el-tooltip
                      class="th-point-name"
                      effect="dark"
                      :content="column.label"
                      placement="top"
                    >
                      <span>{{ column.label }}</span>
                    </el-tooltip>
                  </template>
                </el-table-column>
              </el-table>
              <el-pagination
                layout="total, prev, pager, next, sizes, jumper"
                styled
                background
                :pager-count="5"
                :total="totalItems"
                :current-page.sync="currentPage"
                :page-sizes="[10, 20, 30, 40, 50, 100]"
                :page-size="pageSize"
                @size-change="handleSizeChange"
                @current-change="handleCurrentChange"
              />
            </div>
          </div>
        </div>
      </div>
    </el-block>

    <!-- Y坐标轴选项 弹框 -->
    <el-dialog
      title="Y坐标轴选项"
      :visible.sync="yAxisOptionDialogVisible"
      width="500px"
    >
      <el-form
        ref="form"
        :model="form"
        label-width="120px"
      >
        <el-form-item>
          <template #label>
            添加独立Y轴
            <el-help-tip width="400">
              勾选添加独立Y轴，图表中增加对应测点颜色的独立Y轴，用户可根据需要选择合适的Y轴范围，如果用户没有输入Y轴最大值或最小值，该范围值根据测点值自动生成。
            </el-help-tip>
          </template>
          <el-switch
            v-model="form.ownYaxis"
          />
        </el-form-item>
        <el-form-item
          v-if="form.ownYaxis"
          label="最大值"
        >
          <el-input
            v-model="form.max"
            placeholder="自动值"
          />
        </el-form-item>
        <el-form-item
          v-if="form.ownYaxis"
          label="最小值"
        >
          <el-input
            v-model="form.min"
            placeholder="自动值"
          />
        </el-form-item>
      </el-form>
      <span
        slot="footer"
        class="dialog-footer"
      >
        <el-button
          type="text"
          style="color: #333;"
          @click="resetYAxisOption"
        >
          重置
        </el-button>
        <el-button
          type="text"
          @click="saveYAxisOption"
        >
          确定
        </el-button>
      </span>
    </el-dialog>
    <!-- Y坐标轴选项 弹框 end -->
    <!-- Y主坐标轴选项 弹框 -->
    <el-dialog
      title="Y坐标轴选项"
      :visible.sync="yMainAxiosDialogVisible"
      width="500px"
      @close="clearYMainAxios"
    >
      <el-form
        ref="yMainAxios"
        :model="yMainAxios"
        label-width="120px"
      >
        <el-form-item
          label="最大值"
        >
          <el-input
            v-model="yMainAxios.max"
            placeholder="自动值"
            @input="yMainAxios.max = yMainAxios.max.replace(/[^\-?\d.]/g,'')"
          />
        </el-form-item>
        <el-form-item
          label="最小值"
        >
          <el-input
            v-model="yMainAxios.min"
            placeholder="自动值"
            @input="yMainAxios.min = yMainAxios.min.replace(/[^\-?\d.]/g,'')"
          />
        </el-form-item>
      </el-form>
      <span
        slot="footer"
        class="dialog-footer"
      >
        <el-button
          type="text"
          style="color: #333;"
          @click="resetYMainAxios"
        >
          重置
        </el-button>
        <el-button
          type="text"
          @click="sureClick"
        >
          确定
        </el-button>
      </span>
    </el-dialog>
    <!-- Y坐标轴选项 弹框 end -->

    <!-- 加载模板 抽屉 -->
    <el-modal
      :visible.sync="loadTemplateModalVisible"
    >
      <template #title>
        {{ templateModalType === 'load' ? '加载模板' : '管理模板' }}
      </template>
      <template #actions>
        <el-checkbox
          v-if="templateModalType === 'load'"
          v-model="isMyTemp"
          @change="getTemplateList"
        >
          我的模板
        </el-checkbox>
      </template>

      <el-table
        :data="templateList"
        style="width: 100%"
        @sort-change="sortChange"
      >
        <el-table-column
          prop="templateName"
          label="模板名称"
          width="200"
          fixed
          sortable="template_name"
        >
          <template v-slot="{ row }">
            <template v-if="templateModalType === 'load'">
              <el-button
                type="text"
                @click="loadTemplate(row)"
              >
                {{ row.templateName }}
              </el-button>
            </template>
            <template v-else>
              {{ row.templateName }}
            </template>
          </template>
        </el-table-column>
        <el-table-column
          prop="templateDetail"
          label="模板详情"
          width="300"
          sortable="template_detail"
        />
        <el-table-column
          prop="author"
          label="模板创建人"
          width="150"
          sortable="author"
        />
        <el-table-column
          prop="updateTime"
          label="模板创建时间"
          width="180"
          sortable="updateTime"
        />
        <el-table-column
          v-if="templateModalType === 'manage'"
          label="操作"
          width="120"
          fixed="right"
        >
          <template v-slot="{ row }">
            <!-- <el-button
              type="text"
              @click="openEditTemplateModal(row)"
            >
              编辑
            </el-button> -->
            <el-button
              type="text"
              @click="deleteTemplate(row)"
            >
              删除
            </el-button>
            <el-dropdown @command="(com) => handleRowCommand(com, row)">
              <el-button type="text">
                更多
              </el-button>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item command="manageTempEdit">
                  编辑
                </el-dropdown-item>
                <el-dropdown-item command="manageTempEpt">
                  导出测点
                </el-dropdown-item>
                <el-dropdown-item command="manageTempImt">
                  <el-upload
                    :action="uploadUrl"
                    :data="{}"
                    :multiple="false"
                    :show-file-list="false"
                    :on-success="uploadSuccess"
                    :on-error="uploadFailed"
                    :before-upload="beforeUpload"
                    :http-request="httpRequest"
                    class="upload"
                  >
                    导入测点
                  </el-upload>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        layout="total, prev, pager, next, sizes, jumper"
        styled
        background
        :pager-count="5"
        :total="templatePagination.totalItems"
        :current-page.sync="templatePagination.currentPage"
        :page-sizes="[10, 20, 30, 40, 50, 100]"
        :page-size="templatePagination.pageSize"
        @size-change="handleTemplatePaginationSizeChange"
        @current-change="handleTemplatePaginationCurrentChange"
      />
    </el-modal>
    <!-- 加载模板 抽屉 end -->

    <!-- 保存模板 弹框 -->
    <el-dialog
      :visible.sync="saveTemplateModalVisible"
      width="600px"
      title="保存模板"
    >
      <el-form
        ref="templateForm"
        :model="templateForm"
        label-width="120px"
        :rules="templateFormRules"
      >
        <el-form-item
          label="模板名称"
          prop="templateName"
        >
          <el-input
            v-model="templateForm.templateName"
            placeholder="请输入模板名称"
          />
        </el-form-item>
        <el-form-item
          label="模板详情"
          prop="templateDetail"
        >
          <el-input
            v-model="templateForm.templateDetail"
            type="textarea"
            placeholder="请输入模板详情"
            maxlength="50"
            show-word-limit
            autosize
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button
          type="text"
          @click="saveTemplate()"
        >
          保存
        </el-button>
      </template>
    </el-dialog>
    <!-- 保存模板 弹框 end-->

    <!-- 添加测点 弹框 -->
    <el-modal :visible.sync="pointModalVisible">
      <template slot="title">
        添加测点
      </template>

      <el-form label-width="120px">
        <el-form-item label="设备编号">
          <el-select
            v-model="deviceNumber"
            filterable
            remote
            reserve-keyword
            placeholder="请输入关键词"
            :remote-method="remoteMethodNum"
            :loading="loading"
            @focus="remoteMethodNum"
          >
            <el-option
              v-for="item in devNumbers"
              :key="item"
              :label="item"
              :value="item"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="测点">
          <el-select
            v-model="pointVals"
            filterable
            remote
            reserve-keyword
            placeholder="请输入关键词"
            :remote-method="remoteMethodPoint"
            :loading="loading"
            multiple
            collapse-tags
            :multiple-limit="50 - pointList.length"
            @change="changePoint"
            @focus="remoteMethodPoint"
          >
            <el-option
              v-for="item in points"
              :key="item"
              :label="item"
              :value="item"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <template slot="footer">
        <el-button
          type="primary"
          @click="submitPoint"
        >
          提交
        </el-button>
      </template>
    </el-modal>
    <!-- 添加测点 弹框 end -->
  </div>
</template>

<script>
import { dataQuery as cgi } from '@@/config/cgi';
import { getQueryString } from 'common/script/utils.js';
import getEdgeRequest from '../../utils/request';
import 'feature/utils/business';

import * as echarts from 'echarts';
import moment from 'moment';
import qs from 'qs';
import Cookies from 'js-cookie';
import axios from 'axios';

const color = [
  '#1470CC',
  '#FFB20D',
  '#0ACCCC',
  '#FF7A0D',
  '#5939D6',
  '#0ACC78',
  '#DA5FAC',
  '#96C0EB',
  '#FFE09C',
  '#9BE8E8',
  '#FFCA9E',
  '#AE9EEB',
  '#9BEDCA',
  '#E4A5D4',
];

/**
 * 找到一个大约等于 x 的“好”数。
 */
function nice(span, round) {
  let val = Math.abs(span);
  const exponent = Math.floor(Math.log(val) / Math.LN10);
  const exp10 = 10 ** exponent;
  const f = val / exp10;

  let nf;

  if (round) {
    if (f < 1.5) {
      nf = 1;
    } else if (f < 2.5) {
      nf = 2;
    } else if (f < 4) {
      nf = 3;
    } else if (f < 7) {
      nf = 5;
    } else {
      nf = 10;
    }
  } else {
    if (f < 1) {
      nf = 1;
    } else if (f < 2) {
      nf = 2;
    } else if (f < 3) {
      nf = 3;
    } else if (f < 5) {
      nf = 5;
    } else {
      nf = 10;
    }
  }

  val = nf * exp10;

  return exponent >= -20 ? +val.toFixed(exponent < 0 ? -exponent : 0) : val;
}

export default {
  data() {
    const defaultTimeRange = [
      moment().subtract(1, 'hour')
        .format('YYYY-MM-DD HH:mm:ss'),
      moment().format('YYYY-MM-DD HH:mm:ss'),
    ];

    return {
      title: getQueryString('title'),

      autoHeight: 360,
      uploadUrl: cgi.importTemplateByCondition,
      deviceNumber: '',
      loading: false,
      pointModalVisible: false,
      devNumbers: [],
      points: [],
      pointVals: [],
      isMyTemp: true,
      mozuId: null,

      /**
       * 测点列表
       */
      isIndeterminate: false,
      checkAll: false,
      pointList: [], // 从 url 获取的测点列表
      checkedPointList: [], // checkbox group绑定用的测点列表
      fullPointList: [], // 记录了Y轴设置的完整测点列表

      /**
       * 查询条件
       */
      filter: {
        timerange: defaultTimeRange,
        duration: 1,
        unit: 60,
        stats: [],
      },
      pickerOptions: {
        shortcuts: [
          {
            text: '最近半小时',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (1800 * 1000));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '最近一小时',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (3600 * 1000));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '最近两小时',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (3600 * 1000 * 2));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '最近四小时',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (3600 * 1000 * 4));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '最近八小时',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (3600 * 1000 * 8));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '最近一天',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (3600 * 1000 * 24));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '最近一周',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (3600 * 1000 * 24 * 7));
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: '最近一年',
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - (3600 * 1000 * 24 * 365));
              picker.$emit('pick', [start, end]);
            },
          },
        ],
        disabledDate: time => time.getTime() > Date.now(),
      },
      unitOptions: [
        {
          value: 1,
          label: '秒',
        },
        {
          value: 60,
          label: '分钟',
        },
        {
          value: 60 * 60,
          label: '小时',
        },
        {
          value: 60 * 60 * 24,
          label: '天',
        },
      ],
      statOption: [
        {
          value: 'avg',
          label: '平均值',
        },
        {
          value: 'max',
          label: '最大值',
        },
        {
          value: 'min',
          label: '最小值',
        },
      ],

      /**
       * table
       */
      tableColumns: [],
      tableData: [],
      originTableData: [],
      statTableData: [],
      currentPage: 1,
      totalItems: 0,
      pageSize: 10,

      /**
       * y轴选项
       */
      currentPoint: {},
      yAxisOptionDialogVisible: false,
      form: {
        max: '',
        min: '',
        ownYaxis: false,
      },
      yMainAxios: {
        max: '',
        min: '',
      },
      yMainAxiosDialogVisible: false,

      /**
       * 加载模板
       */
      loadTemplateModalVisible: false,
      templateList: [],
      currentTemplate: {},
      templateModalType: 'load',
      templatePagination: {
        totalItems: 0,
        currentPage: 1,
        pageSize: 10,
      },
      orderByMap: {},
      author: Cookies.get('tnebula_username'),
      /**
       * 保存模板
       */
      saveTemplateModalVisible: false,
      templateForm: {
        templateName: '',
        templateDetail: '',
      },
      templateFormRules: {
        templateName: [
          { required: true, message: '请输入模板名称', trigger: 'blur' },
        ],
      },
    };
  },
  mounted() {
    const query = qs.parse(location.search.slice(1));
    const mozuId = query.mozuId ? query.mozuId : 0;
    this.mozuId = mozuId || TNBL.getCurrModule().id;

    if (getQueryString('pointlist')) {
      this.getPointList();
    }
  },
  methods: {
    clearYMainAxios() {
      if (Number(this.yMainAxios.max) < Number(this.yMainAxios.min)) {
        this.yMainAxios.min = '';
        this.yMainAxios.max = '';
      }
    },
    sureClick() {
      const { min, max } = this.yMainAxios;
      // if (!min || !max) {
      //   this.$message.error('请输入！');
      //   return;
      // }
      if (min && max) {
        if (Number(max) > Number(min)) {
          this.yMainAxiosDialogVisible = false;
        } else {
          this.$message.error('最大值不能小于最小值！');
        }
      }
      this.yMainAxiosDialogVisible = false;
    },
    /**
     * 模板管理
     */
    handleTestCommand(command) {
      switch (command) {
        case 'seled': {
          this.yMainAxiosDialogVisible = true;
          break;
        }
        case 'addTest': {
          this.pointModalVisible = true;
          break;
        }
      }
    },
    // 重置y轴->查询
    resetYMainAxios() {
      this.yMainAxios = {
        max: '',
        min: '',
      };
    },
    async beforeUpload(file) {
      const formData = new FormData();
      formData.append('file', file);
      try {
        const { data } = await axios.post(cgi.importTemplateByCondition, formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
            mozuId: this.mozuId,
            templateId: this.importInfo.id,
          },
        });
        if (data.code === 0) {
          this.$message.success('导入成功');
          this.getTemplateList();
        } else {
          this.$message.error('只能导入自己的模板测点');
        }
      } catch (error) {
        this.$message.error('导入失败');
      }
    },
    httpRequest() {

    },
    uploadSuccess(res) {
      if (res.code < 0) {
        this.$message({
          showClose: true,
          message: res.message,
          type: 'error',
        });
        this.filterHandler();
      } else {
        this.$message.success('数据导入成功');
      }
    },
    uploadFailed() {
      this.$message.error('导入数据失败');
    },
    submitPoint() {
      const { pointVals, deviceNumber, pointList } = this;
      this.pointModalVisible = false;
      if (pointVals.length === 0) return;
      const devNum = deviceNumber;
      pointVals.forEach((item) => {
        const pointName = `${devNum}.${item}`;
        if (pointList.includes(pointName)) return;
        pointList.push(pointName);
      });
      this.getPointList(pointList);
      this.deviceNumber = '';
      this.pointVals = [];
      this.devNumbers = [];
      this.points = [];
    },
    changePoint(v) {
      const arr = [...v, ...this.pointList];
      if (arr.length >= 50) {
        this.$message.warning('测点最大容量50个！');
        return false;
      }
    },
    async remoteMethodNum(v) {
      const flag = typeof v === 'object';
      if (flag && this.devNumbers.length !== 0) return; // 获取焦点
      const { list } = await this.getDevNumPoint({ condition: flag ? '' : v });
      this.devNumbers = (list || []).slice(0, 100);
    },
    async remoteMethodPoint(v) {
      const flag = typeof v === 'object';
      if (!this.deviceNumber) {
        this.$message.warning('请先选择设备编号！');
        return;
      }
      if (flag && this.points.length !== 0) return; // 获取焦点
      const { list } = await this.getDevNumPoint({ deviceNumber: this.deviceNumber, condition: flag ? '' : v });
      this.points = (list || []).slice(0, 100);
    },
    async getDevNumPoint(params) {
      const data = await getEdgeRequest(this.$axios, this.mozuId).post('/cgi/dataQuery/edge/getDeviceNumberListAndAttrs ', params);
      return data;
    },
    // 排序
    sortChange({ order, column }) {
      this.orderByMap = {};
      if (order === 'ascending') {
        this.orderByMap[column.sortable] = 'asc';
      } else if (order === 'descending') {
        this.orderByMap[column.sortable] = 'desc';
      }
      this.getTemplateList();
    },
    /**
     * 从 url 参数中获取测点
     */
    getPointList(newPointList = []) {
      this.pointList = newPointList.length === 0 ? getQueryString('pointlist').split(',') : newPointList;
      this.checkedPointList = this.pointList;
      this.checkAll = true;

      this.fullPointList = this.pointList.map(e => ({
        id: e,
        deviceNo: e.split('.')[0],
        attrName: e.split('.')[1],
        yaxisMax: '',
        yaxisMin: '',
        ownYaxis: false,
      }));

      this.getData();
    },

    /**
     * 全选处理
     */
    handleCheckAllChange(val) {
      if (val) {
        this.checkedPointList = this.pointList;
      } else {
        this.checkedPointList = [];
      }

      this.isIndeterminate = false;
    },

    /**
     * 单选处理
     */
    handleCheckedPointListChange(val) {
      const checkedCount = val.length;
      this.checkAll = checkedCount === 10;
      this.isIndeterminate = checkedCount > 0 && checkedCount < 10;
    },

    /**
     * 根据时间跨度，设置时间单位
     * 时间跨度小于10分钟时，默认1秒
     * 大于等于10分钟，小于1天，默认1分钟
     * 大于等于1天，小于30天，默认1小时
     * 大于等于30天，默认1天
     * @param {Array} val - 时间跨度
     */
    timerangeChangeHandler(val) {
      const [startTime, endTime] = val;
      const diff = moment(endTime).diff(moment(startTime));

      if (diff >= 30 * 24 * 60 * 60 * 1000) {
        this.filter.unit = 60 * 60 * 24;
      } else if (diff >= 24 * 60 * 60 * 1000) {
        this.filter.unit = 60 * 60;
      } else if (diff >= 10 * 60 * 1000) {
        this.filter.unit = 60;
      } else {
        this.filter.unit = 1;
      }

      this.filter.duration = 1;
    },

    /**
     * 根据模板测点列表查询历史数据
     */
    getData() {
      const [startTime, endTime] = this.filter.timerange;
      const params = {
        templatePointList: this.checkedPointList,
        startTime,
        endTime,
        interval: this.filter.duration * this.filter.unit,
        stats: ['avg', 'min', 'max'],
      };

      getEdgeRequest(this.$axios, this.mozuId)
        .post(cgi.getHistoryBizGidAttrValuesByTemplate, params)
        .then((data) => {
          const sortedData = _.sortBy(
            data,
            o => this.checkedPointList.findIndex(e => e === o.id)
          );

          this.createTable(sortedData);
          this.createChart(sortedData);
        });
    },

    /**
     * 根据模板测定列表导出历史数据
     */
    exportHistory() {
      const [startTime, endTime] = this.filter.timerange;
      const params = {
        templatePointList: this.checkedPointList,
        startTime,
        endTime,
        interval: this.filter.duration * this.filter.unit,
        stats: this.filter.stats,
      };

      const { title } = this;

      const getTsString = () => moment().format('YYYYMMDDHHmmss');

      getEdgeRequest(this.$axios, this.mozuId)
        .download(cgi.exportHistoryBizGidAttrValuesByTemplate, params, true, title ? {
          fileName: `${title}-${getTsString()}.xlsx`,
        } : {});
    },

    /**
     * 将获取到的历史数据转换为表格支持的数据
     * @param {Array} data - 历史数据
     */
    createTable(data) {
      this.currentPage = 1;
      this.pageSize = 10;

      const statMap = {
        avg: '平均值',
        max: '最大值',
        min: '最小值',
      };

      this.tableColumns = data.map(e => ({
        prop: e.deviceNumber + e.attrName,
        label: e.id,
        sortable: true,
      }));

      this.originTableData = data[0].data.map((e, i) => {
        const columnData = {
          updateTime: e.updateTime,
        };

        data.forEach((item) => {
          const id = item.deviceNumber + item.attrName;
          columnData[id] = item.data[i]?.value;
        });

        return columnData;
      });
      this.totalItems = this.originTableData.length;
      this.sliceTableData();

      this.statTableData = this.filter.stats.map((e) => {
        const columnData = {
          stat: statMap[e],
        };

        data.forEach((item) => {
          const id = item.deviceNumber + item.attrName;
          columnData[id] = item.stats.find(s => s.name === e).value;
        });

        return columnData;
      });
    },

    sliceTableData() {
      const start = (this.currentPage - 1) * this.pageSize;
      const limit = this.pageSize;

      this.tableData = this.originTableData.slice(start, start + limit);
    },

    handleSizeChange(val) {
      this.pageSize = val;
      this.currentPage = 1;

      this.sliceTableData();
    },

    handleCurrentChange(val) {
      this.currentPage = val;

      this.sliceTableData();
    },

    /**
     * 创建折线图
     * @param {Array} data - 历史数据
     */
    createChart(data) {
      this.chart = echarts.init(this.$refs.chart);
      this.chart.clear();
      const number = ((data || []).length / 4);
      this.autoHeight = 360;
      const justCount = number > 4 ? 120 + (number * 15) : 120;
      this.autoHeight = this.autoHeight + justCount;
      this.chart.resize({
        height: this.autoHeight,
      });

      const noYAxisPoints = data.filter(e => this.fullPointList.find(a => a.id === e.id && !a.ownYaxis));
      const yAxisPoints = data.filter(e => this.fullPointList.find(a => a.id === e.id && a.ownYaxis));
      const yAxisCount = (noYAxisPoints.length ? 1 : 0) + yAxisPoints.length;

      const option = {
        color,
        toolbox: {
          show: true,
          feature: {
            saveAsImage: {},
          },
          top: 4,
          right: 24,
        },
        tooltip: {
          trigger: 'axis',
          backgroundColor: '#fff',
          padding: 16,
          extraCssText: 'border-radius: 0; min-width: 200px; box-shadow: 0 3px 5px 0 rgba(203,203,203,0.50);',
          // position: ['50%', '50%'],
          // position(point, params, dom, rect, size) {
          //   console.log(point[0]);
          //   // 固定在顶部
          //   return [point[0], '-30%'];
          // },
          textStyle: {
            color: '#666',
            fontSize: 12,
          },
          formatter(params) {
            const arr = params.map((param) => {
              const { unit } = data.find(e => e.id === param.seriesName);
              return `${param.marker}${param.seriesName}：${param.data} ${unit}`;
            });

            return `
              ${params[0].name}<br>
              ${arr.join('<br>')}
            `;
          },
        },
        legend: {
          data: data.map(e => e.id),
          top: 30,
          left: 24,
          selectedMode: false,
          textStyle: {
            fontSize: 11,
          },
        },
        grid: {
          top: justCount,
          bottom: 40,
          left: 90,
          right: yAxisCount <= 2
            ? 80
            : ((yAxisCount - 1) * 80),
        },
        xAxis: {
          type: 'category',
          data: data[0].data.map(e => e.updateTime),
          axisTick: { show: false },
          axisLine: {
            lineStyle: {
              color: '#f0f0f0',
            },
          },
          axisLabel: {
            color: '#999',
            fontSize: 10,
            formatter(value) {
              return value.split(' ').join('\n');
            },
          },
        },
        yAxis: this.createYAxis(data),
        series: this.createSeries(data),
      };

      this.chart.setOption(option);
    },

    createYAxis(data) {
      const yAxisList = [];

      // 最左侧的主轴
      const noYAxisPoints = data.filter(e => this.fullPointList.find(a => a.id === e.id && !a.ownYaxis));
      const yAxisPoints = data.filter(e => this.fullPointList.find(a => a.id === e.id && a.ownYaxis));
      let max;
      let min;
      if (data.length === 1) {
        // eslint-disable-next-line no-inner-declarations
        function integer(num, key) {
          const decimals = num.toString().split('.')[1];
          if (!decimals) return num;

          if (decimals.length === 1) return Math[key](num);

          if (decimals.length === 2) return Math[key](num * 10) / 10;

          return num;
        }
        max = integer(Math.max(...data[0].data.filter(e => e.value !== '--').map(e => e.value)), 'ceil');
        min = integer(Math.min(...data[0].data.filter(e => e.value !== '--').map(e => e.value)), 'floor');
      } else {
        max = this.yMainAxios.max || nice(
          Math.max(...noYAxisPoints.map(e => e.stats.find(item => item.name === 'max').value)),
          false
        );
        // const min = noYAxisPoints.map(e => e.stats.find(item => item.name === 'min').value);
        min = this.yMainAxios.min || Math.min(...noYAxisPoints.map(e => e.stats.find(item => item.name === 'min').value));
        if (max === 0) {
          max = 1;
          if (min === 0) {
            min = -1;
          }
        }
      }

      let mainAxisCount = 0;
      if (noYAxisPoints.length) {
        yAxisList.push({
          type: 'value',
          axisTick: { show: false },
          axisLine: { show: false },
          axisLabel: {
            color: '#999',
          },
          splitLine: {
            lineStyle: {
              color: '#f0f0f0',
            },
          },
          splitNumber: 5,
          // interval: Math.round(max / 5 * 100) / 100,
          interval: Math.round((Math.abs(max) + Math.abs(min)) / 5 * 100) / 100,
          // min: 0,
          min,
          max,
        });

        mainAxisCount = 1;
      }

      yAxisPoints.forEach((e, i) => {
        const index = data.findIndex(item => item.id === e.id);
        const { yaxisMax, yaxisMin } = this.fullPointList.find(p => p.id === e.id);
        const maxValue = e.stats.find(item => item.name === 'max').value;

        let min = yaxisMin === '' ? 0 : yaxisMin;
        let max = yaxisMax === '' ? nice(maxValue, false) : yaxisMax;

        if (max === 0) {
          max = 1;
        }
        if (min === 1) {
          min = 0;
        }

        let offset = 0;
        if (mainAxisCount === 1) {
          offset = 80 * i;
        } else {
          if (i === 0) {
            offset = 0;
          } else {
            offset = 80 * (i - 1);
          }
        }

        yAxisList.push({
          type: 'value',
          axisTick: { show: false },
          axisLine: { show: false },
          axisLabel: {
            color: color[index],
          },
          splitLine: {
            lineStyle: {
              color: '#f0f0f0',
            },
          },
          splitNumber: 5,
          interval: Math.round((max - min) / 5 * 100) / 100,
          min,
          max,
          offset,
        });
      });

      return yAxisList;
    },

    createSeries(data) {
      const noYAxisPoints = data.filter(e => this.fullPointList.find(a => a.id === e.id && !a.ownYaxis));

      let yAxisCount = 0;
      if (noYAxisPoints.length) {
        yAxisCount = 1;
      }
      console.log(this.fullPointList);
      return data.map((e) => {
        const fullPoint = this.fullPointList.find(p => p.id === e.id);
        const { ownYaxis } = fullPoint;

        let yAxisIndex = 0;

        if (ownYaxis) {
          yAxisIndex = yAxisCount;
          yAxisCount += 1;
        }

        return {
          name: e.id,
          data: e.data.map(e => e.value),
          yAxisIndex: ownYaxis ? yAxisIndex : 0,
          type: 'line',
          markPoint: {
            symbolSize: 40,
            data: [
              this.filter.stats.includes('max')
                ? { type: 'max', name: '最大值' } : {},
              this.filter.stats.includes('min')
                ? { type: 'min', name: '最小值' } : {},
            ],
          },
          markLine: this.filter.stats.includes('avg') ? {
            symbol: 'none',
            data: [
              { type: 'average', name: '平均值' },
            ],
            label: {
              show: false,
            },
          } : {},
          symbol: 'none',
        };
      });
    },

    /**
     * 打开Y轴设置编辑弹框
     * @param {string} pointId - 测点的id
     */
    openYAxisOptionDialog(pointId) {
      this.yAxisOptionDialogVisible = true;

      this.currentPoint = this.fullPointList.find(e => e.id === pointId);

      const { yaxisMax, yaxisMin, ownYaxis } = this.currentPoint;
      this.form = {
        max: yaxisMax,
        min: yaxisMin,
        ownYaxis: Number(ownYaxis) === 1,
      };
    },

    /**
     * Y轴选项恢复为默认值
     */
    resetYAxisOption() {
      // 恢复为空值，在图表设置中会自动计算
      this.form = {
        max: '',
        min: '',
        ownYaxis: false,
      };
    },

    /**
     * 保存测点的Y轴选项
     */
    saveYAxisOption() {
      const { form, currentPoint } = this;
      const { max, min, ownYaxis } = form;

      currentPoint.yaxisMax = max;
      currentPoint.yaxisMin = min;
      currentPoint.ownYaxis = ownYaxis;

      this.yAxisOptionDialogVisible = false;
    },

    /**
     * 模板管理
     */
    handleCommand(command) {
      switch (command) {
        case 'loadTemplate': {
          this.openTemplateListModal({ type: 'load' });
          break;
        }
        case 'manageTemplate': {
          this.openTemplateListModal({ type: 'manage' });
          break;
        }
        case 'saveTemplate': {
          this.openSaveTemplateModal();
          break;
        }
        case 'editTemplate': {
          this.overrideTemplate();
          break;
        }
      }
    },

    /**
     * 打开加载模板弹框
     */
    openTemplateListModal(options) {
      this.templateModalType = options.type;
      this.loadTemplateModalVisible = true;
      this.orderByMap = {};
      this.isMyTemp = true;
      this.getTemplateList();
    },
    /**
     * 模板管理
     */
    handleRowCommand(command, row) {
      switch (command) {
        case 'manageTempEdit': {
          this.openEditTemplateModal(row);
          break;
        }
        case 'manageTempEpt': {
          this.manageTempEpt(row);
          break;
        }
        case 'manageTempImt': {
          this.importInfo = row;
          // this.manageTempImt(row);
          break;
        }
      }
    },
    /**
     * 管理模板导出测点
     */
    manageTempEpt({ id }) {
      const { currentPage, pageSize, orderByMap } = this.templatePagination;
      const params = {
        req: {
          fieldWithValueMap: {
            id,
          },
          start: (currentPage - 1) * pageSize,
          limit: pageSize,
          orderByMap,
        },
        headNames: {
          checked: '检查',
          deviceNo: '编号',
          attrName: '指标',
          yaxisMax: '最大',
          yaxisMin: '最小',
          ownYaxis: 'own',
        },
      };

      getEdgeRequest(this.$axios, this.mozuId)
        .download(cgi.exportTemplateByCondition, params);
    },

    /**
     * 加载模板列表
     */
    getTemplateList() {
      const { currentPage, pageSize } = this.templatePagination;
      const params = {
        fieldWithValueMap: {},
        orderByMap: this.orderByMap,
        start: (currentPage - 1) * pageSize,
        limit: pageSize,
      };
      // if (this.isMyTemp && this.templateModalType === 'load') {
      params.fieldWithValueMap.author = this.author;
      // }

      getEdgeRequest(this.$axios, this.mozuId)
        .post(cgi.selectTemplateByCondition, params, '', false)
        .then((data) => {
          this.templateList = data.list;
          this.templatePagination.totalItems = data.count;
        });
    },

    /**
     * 点击模板名称加载模板
     */
    loadTemplate(row) {
      this.currentTemplate = { ...row };

      this.fullPointList = row.data.map(e => ({
        ...e,
        id: `${e.deviceNo}.${e.attrName}`,
      }));
      this.pointList = this.fullPointList.map(e => e.id);

      this.checkedPointList = this.fullPointList
        .filter(e => Number(e.checked) === 1)
        .map(e => e.id);

      this.checkAll = false;
      if (this.checkedPointList.length) {
        this.isIndeterminate = true;
      }
      if (this.checkedPointList.length === this.fullPointList) {
        this.checkAll = true;
        this.isIndeterminate = false;
      }

      const { startTime, endTime, step, unit, stats } = row;

      this.filter = {
        timerange: [startTime, endTime],
        duration: step,
        unit,
        stats,
      };

      this.loadTemplateModalVisible = false;
      this.getData();
    },

    /**
     * 删除模板
     * @param {Object} row - 需要删除的模板列
     */
    deleteTemplate(row) {
      this
        .$confirm(`确定删除模板“${row.templateName}”吗？`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        })
        .then(() => {
          getEdgeRequest(this.$axios, this.mozuId)
            .post(`${cgi.deleteTemplateById}/${row.id}`, {}, '', false)
            .then(() => {
              this.$message.success('已删除模板');

              this.getTemplateList();
            });
        })
        .catch(() => {});
    },

    /**
     * 打开保存模板弹框
     */
    openSaveTemplateModal() {
      this.templateForm = {
        templateName: '',
        templateDetail: '',
      };

      this.saveTemplateModalVisible = true;
    },

    /**
     * 打开编辑模板弹框
     * @param {Object} row - 需要编辑的模板列
     */
    openEditTemplateModal(row) {
      const { id, templateName, templateDetail } = row;

      this.templateForm = {
        id,
        templateName,
        templateDetail,
      };
      this.saveTemplateModalVisible = true;
    },

    createParams() {
      const data = this.fullPointList.map(e => ({
        ...e,
        ownYaxis: e.ownYaxis ? 1 : 0,
        checked: this.checkedPointList.includes(e.id) ? 1 : 0,
      }));
      const { duration, unit, timerange, stats } = this.filter;
      const step = duration;
      const [startTime, endTime] = timerange;

      return {
        data,
        step,
        unit,
        stats,
        startTime,
        endTime,
      };
    },

    /**
     * 发送保存模板请求
     */
    saveTemplate() {
      this.$refs.templateForm.validate((valid) => {
        if (valid) {
          const { templateName, templateDetail } = this.templateForm;

          const params = {
            ...this.createParams(),
            templateName,
            templateDetail,
          };

          // 如果带有 id 属性，为编辑模板
          if ('id' in this.templateForm) {
            params.id = this.templateForm.id;
          }

          getEdgeRequest(this.$axios, this.mozuId)
            .post(cgi.insertOrUpdateTemplate, params, '', false)
            .then(() => {
              this.$message.success('保存模板成功');
              this.saveTemplateModalVisible = false;
            });
        }
      });
    },

    /**
     * （修改测点的Y轴设置后）覆盖模板
     */
    overrideTemplate() {
      const { templateName } = this.currentTemplate;
      this
        .$confirm(`确定覆盖模板“${templateName}”吗？`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        })
        .then(() => {
          const { id, templateDetail } = this.currentTemplate;

          const params = {
            ...this.createParams(),
            id,
            templateName,
            templateDetail,
          };

          getEdgeRequest(this.$axios, this.mozuId)
            .post(cgi.insertOrUpdateTemplate, params, '', false)
            .then(() => {
              this.$message.success('编辑模板成功');
              this.saveTemplateModalVisible = false;
            });
        })
        .catch(() => {});
    },

    handleTemplatePaginationSizeChange(val) {
      this.templatePagination.pageSize = val;
      this.templatePagination.currentPage = 1;

      this.getTemplateList();
    },

    handleTemplatePaginationCurrentChange(val) {
      this.templatePagination.currentPage = val;

      this.getTemplateList();
    },
  },
};
</script>

<style lang="scss" scoped>
.advanced-search {
  &-container {
    display: grid;
    grid-template-columns: 20% minmax(0, 1fr);
  }

  &-aside {
    border-right: 1px solid #f0f0f0;
    position: relative;
  }

  &-empty-point-list {
    line-height: 60px;
    text-align: center;
    color: #999;
  }

  &-check-all {
    height: 64px;
    display: flex;
    align-items: center;
    padding: 0 16px;
    border-bottom: 1px solid #f0f0f0;
    // justify-content: space-between;
  }

  &-checkbox-list {
    .el-checkbox {
      box-sizing: border-box;
      width: 100%;
      display: flex;
      align-items: center;
      height: 48px;
      padding: 0 16px;

      /deep/ &__label {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: space-between;
        overflow: hidden;
      }

      .point-name {
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        margin-right: 8px;
      }
    }
  }

  &-toolbar {
    height: 64px;
    display: flex;
    align-items: center;
    padding: 0 24px;
    border-bottom: 1px solid #f0f0f0;

    &-date-picker {
      margin-right: 24px;
      flex-shrink: 0;
      flex-basis: 400px;
    }

    &-time {
      display: flex;
      align-items: center;

      &-value {
        width: 100px;

        /deep/ input {
          text-align: center;
        }
      }

      &-unit {
        width: 60px;
      }
    }

    &-value-type {
      width: 240px;
    }

    &-template {
      margin-left: auto;

      &-btn {
        display: flex;
        align-items: center;
        cursor: pointer;
      }
    }
  }

  &-chart {
    width: 100%;
    height: 400px;
  }
}

/deep/ .th-point-name {
  display: inline-block;
  vertical-align: top;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  max-width: 300px;
}
/deep/ .el-modal-body__actions {
  font-size: 16px;
}
.advanced-search-checkbox-list {
  max-height: 1240px;
  overflow: auto;
}
</style>
