<template>
  <el-modal
    :visible.sync="visible"
    type="message"
    custom-layout
    @close="handleClose"
    @opened="handleOpen"
  >
    <template slot="title">
      消息
    </template>

    <el-menu
      mode="horizontal"
      :default-active="menuIndex"
      @select="selectMenu"
    >
      <el-menu-item index="1">
        任务中心
      </el-menu-item>
      <el-menu-item index="2">
        消息中心
      </el-menu-item>
    </el-menu>

    <div
      v-if="menuIndex === '1'"
      id="taskDiv"
    >
      <el-block
        no-padding
        style="position: relative;"
      >
        <el-dropdown
          trigger="click"
          :hide-on-click="false"
          style="
              position: absolute;
              top: 16px;
              right: 16px;
              z-index: 1;
            "
        >
          <i
            class="tn-icon-filter"
            style="color: #999; cursor: pointer;"
          />
          <el-dropdown-menu
            slot="dropdown"
            width="170"
          >
            <el-dropdown-item>
              <el-checkbox
                v-model="checkAll"
                :indeterminate="isIndeterminate"
                @change="handleCheckAllChange"
              >
                全部任务
              </el-checkbox>
            </el-dropdown-item>
            <el-checkbox-group
              v-model="checkedItems"
              @change="handleCheckedItemsChange"
            >
              <el-dropdown-item
                v-for="(t,i) in todoType"
                :key="i"
              >
                <el-checkbox
                  :label="t"
                >
                  {{ t }}任务
                  <el-tag
                    size="mini"
                    :type="todoTypeTag[t]"
                    effect="dark"
                  >
                    {{ t }}
                  </el-tag>
                </el-checkbox>
              </el-dropdown-item>
            </el-checkbox-group>
          </el-dropdown-menu>
        </el-dropdown>
        <el-tabs
          v-model="activeName"
          @tab-click="handleTabClick"
        >
          <el-tab-pane
            :label="todoTitle"
            name="todo"
          />

          <el-tab-pane
            label="我的已办"
            name="ok"
          />
          <div style="padding: 16px 24px;border-bottom: 1px solid #f0f0f0;">
            <el-input
              v-model="key"
              border-type="plain"
              prefix-icon="tn-icon-search"
              placeholder="输入任务名/单号/派单人进行搜索"
              @input="searchHandler"
            />
          </div>
          <div
            v-infinite-scroll="scrollLoad"
            infinite-scroll-delay="300"
            infinite-scroll-immediate="false"
            style="overflow-y: auto;max-height: calc(100vh - 300px);"
          >
            <el-collapse
              v-for="(task,ti) in taskList"
              :key="ti"
            >
              <el-collapse-item arrow-position="left">
                <el-tag
                  slot="tag"
                  size="mini"
                  :type="todoTypeTag[task.t_type]"
                  effect="dark"
                >
                  {{ task.t_type }}
                </el-tag>
                <template slot="title">
                  【{{ task.t_instanceid }}】{{ task.t_instancename }}
                </template>
                <template slot="subtitle">
                  <span style="margin-right: 32px;">
                    当前流程：{{ task.t_taskname }}
                  </span>
                  <span style="margin-right: 32px;">派单人：{{ task.t_createuname }}</span>
                  <span>派单时间：{{ (task.t_createtime).substr(0,16) }}</span>
                </template>

                <template
                  v-if="checkjsontext(task)"
                >
                  <template
                    v-for="(td,tdi) in task.t_jsontext.show_column"
                  >
                    <el-row
                      :key="tdi"
                      type="flex"
                    >
                      <el-col :span="24">
                        {{ tdi }}：{{ td }}
                      </el-col>
                    </el-row>
                  </template>

                  <!-- <el-form
                  v-show="showOper"
                  ref="form"
                  :model="form"
                  label-width="100px"
                >
                  <el-form-item label="审批意见">
                    <el-input v-model="form.remark" />
                  </el-form-item>
                  <el-form-item style="text-align: right;">
                    <el-button
                      type="text"
                      @click="handleTodo('reject',task.instance_id,task.task_id)"
                    >
                      驳回
                    </el-button>
                    <el-button
                      type="text"
                      @click="handleTodo('pass',task.instance_id,task.task_id)"
                    >
                      通过
                    </el-button>
                  </el-form-item>
                </el-form> -->
                  <el-form
                    v-show="!showOper"
                  >
                    <el-form-item style="text-align: right;">
                      <el-button
                        v-if="task.t_jsontext.todo_url"
                        :path="task.t_jsontext.todo_url"
                        type="text"
                        target="_blank"
                        :data-event="frameflag"
                      >
                        去处理
                      </el-button>
                      <el-button
                        v-if="task.t_jsontext.detail_url"
                        :path="task.t_jsontext.detail_url"
                        type="text"
                        target="_blank"
                        :data-event="frameflag"
                      >
                        查看详情
                      </el-button>
                    </el-form-item>
                  </el-form>
                </template>
              </el-collapse-item>
            </el-collapse>
          </div>
        </el-tabs>
      </el-block>

      <el-button
        type="text"
        style="display: flex; margin: 0 auto;"
        @click.native="scrollLoad"
      >
        加载更多
      </el-button>
    </div>
    <div
      v-if="menuIndex === '2'"
      id="msgDiv"
    >
      <el-block
        no-padding
        style="position: relative;"
      >
        <!-- 消息类型筛选后面做 -->
        <el-dropdown
          v-if="false"
          trigger="click"
          :hide-on-click="false"
          style="
              position: absolute;
              top: 16px;
              right: 16px;
              z-index: 1;
            "
        >
          <i
            class="tn-icon-filter"
            style="color: #999; cursor: pointer;"
          />
          <el-dropdown-menu
            slot="dropdown"
            width="170"
          >
            <el-dropdown-item>
              <el-checkbox
                v-model="checkAll"
                :indeterminate="isIndeterminate"
                @change="handleCheckAllChange"
              >
                全部
              </el-checkbox>
            </el-dropdown-item>
            <el-checkbox-group
              v-model="checkedItems"
              @change="handleCheckedItemsChange"
            >
              <el-dropdown-item
                v-for="(t,i) in msgType"
                :key="i"
              >
                <el-checkbox
                  :label="t"
                >
                  {{ t }}
                  <el-tag
                    size="mini"
                    :type="msgTypeTag[t]"
                    effect="dark"
                  >
                    {{ t }}
                  </el-tag>
                </el-checkbox>
              </el-dropdown-item>
            </el-checkbox-group>
          </el-dropdown-menu>
        </el-dropdown>

        <el-tabs
          v-model="msgActiveName"
          @tab-click="handleTabClick"
        >
          <el-tab-pane
            :label="todoTitle"
            name="unread"
          />

          <el-tab-pane
            label="已读消息"
            name="read"
          />

          <div style="padding: 16px 24px;border-bottom: 1px solid #f0f0f0;display:inline-block;width:100%">
            <el-input
              v-model="msgKey"
              border-type="plain"
              prefix-icon="tn-icon-search"
              placeholder="输入标题/内容进行搜索"
              style="width:500px!important"
              @input="searchHandler"
            />
            <!-- <div
              v-show="showReadButton"
              class="have-read "
              @click="popoverVisible = true"
            >
              全部已读
            </div> -->
            <div
              v-show="showReadButton"
              class="have-read"
            >
              <el-popover
                v-model="popoverVisible"
                placement="top"
                width="160"
              >
                <p style="padding-bottom:6px">
                  确定全部标记为已读？
                </p>
                <div style="text-align: right; margin: 0">
                  <el-button
                    size="mini"
                    type="text"
                    @click="popoverVisible = false"
                  >
                    取消
                  </el-button>
                  <el-button
                    type="primary"
                    size="mini"
                    @click="handleReadAll()"
                  >
                    确定
                  </el-button>
                </div>
                <a slot="reference">
                  全部已读
                </a>
              </el-popover>
            </div>
          </div>

          <div
            v-infinite-scroll="scrollLoad"
            infinite-scroll-delay="300"
            :infinite-scroll-disabled="busy"
            infinite-scroll-immediate="false"
            style="overflow-y: auto;max-height: calc(100vh - 300px);"
          >
            <el-collapse
              v-for="(msg,ti) in msgList"
              :key="ti"
            >
              <el-collapse-item arrow-position="left">
                <el-tag
                  slot="tag"
                  size="mini"
                  :type="msgTypeTag[msg.msType]"
                  effect="dark"
                >
                  {{ msg.msType }}
                </el-tag>
                <template slot="title">
                  <!-- 【{{ msg.muId }}】{{ msg.msTitle }} -->
                  &nbsp; {{ msg.msTitle }}
                </template>
                <template slot="subtitle">
                  <span style="margin-right: 32px;">
                    &nbsp; 创建时间：{{ msg.muCreateTime }}
                  </span>
                  <span>更新时间：{{ msg.muUpdateTime }}</span>
                </template>

                <template>
                  <el-row
                    type="flex"
                  >
                    &nbsp; {{ msg.msData }}
                  </el-row>

                  <el-form
                    v-show="msgActiveName === 'unread'"
                  >
                    <el-form-item style="text-align: right;">
                      <el-button

                        type="text"
                        target="_blank"
                        @click="handleRead(msg)"
                      >
                        查看
                      </el-button>
                    </el-form-item>
                  </el-form>
                </template>
              </el-collapse-item>
            </el-collapse>
          </div>
        </el-tabs>
      </el-block>
      <el-button
        v-if="haveMore"
        type="text"
        style="display: flex; margin: 0 auto;"
        @click.native="scrollLoad"
      >
        加载更多
      </el-button>
      <el-button
        v-if="!haveMore"
        type="text"
        style="display: flex; margin: 0 auto;"
      >
        无更多数据
      </el-button>
    </div>
  </el-modal>
</template>

<script>
import * as reqhel from '../common/reqhel.js';
import { debounce } from 'lodash';

export default {

  props: {
    childData: {
      type: Number,
      required: true,
    },
    frameflag: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      popoverVisible: false,
      countForShowAllMsg: 0,
      haveMore: false,
      showId: false,
      busy: false,
      showOper: 0,
      total: 0,
      todoCnt: 0,
      unreadCnt: 0,
      visible: false,
      menuIndex: '1', // 任务中心1  消息中心 2
      checkAll: false,
      isIndeterminate: false,
      checkedItems: [], // task item
      checkedMsgItems: [], // msg item
      todoType: ['演练', '维保', '事件', '巡检', '变更', '维修'],
      todoTypeTag: { 巡检: 'success', 维保: 'warning', 演练: 'info', 变更: 'danger', 事件: '', 维修: 'danger' },
      msgType: ['通知', '消息'],
      msgTypeTag: { 通知: '', 消息: '' },
      activeName: 'todo',
      msgActiveName: 'unread',
      taskList: [],
      msgList: [],
      limit: 10,
      start: 0,
      key: '',
      msgKey: '',
      form: {
        remark: '',
      },
    };
  },
  computed: {
    todoTitle() {
      if (this.menuIndex === '1') {
        return `我的待办（${this.todoCnt}）`;
      }
      return `未读消息（${this.unreadCnt}）`;
    },
    showReadButton() {
      if (this.msgActiveName === 'unread' && this.countForShowAllMsg !== 0) {
        return true;
      }
      return false;
    },
    params() {
      if (this.menuIndex === '1') {
        return {
          taskStatus: this.activeName,
          start: this.start,
          limit: this.limit,
          type: (this.checkedItems).join(','),
          keyWord: this.key,
        };
      }
      return {
        status: this.msgActiveName === 'read' ? '1' : '0', // 已读1   未读0
        start: this.start,
        length: this.limit,
        type: (this.checkedItems).join(','),
        keyword: this.msgKey,
      };
    },
  },
  watch: {
    childData(val) {
      this.visible = (val > 0 && true) || false;
    },
    menuIndex(val) {
      this.key = '';
      this.msgKey = '';
      this.checkAll = false;
      this.isIndeterminate = false;
      this.checkedItems = [];
      // this.start = 0
      // this.limit = 10
      if (val === '1') {
        this.getTaskList(0);
      } else if (val === '2') {
        this.getMsgList(0);
      }
    },
  },
  methods: {
    handleReadAll() {
      this.popoverVisible = false;

      let ids = '';

      // eslint-disable-next-line no-restricted-syntax
      for (const i in this.msgList) {
        if (ids) {
          ids = `${ids},${this.msgList[i].muId}`;
        } else {
          ids = this.msgList[i].muId;
        }
      }
      reqhel.changeStatus({ ids }).then((r) => {
        if (r.succ) {
          this.getMsgList(0);
        }
      });
    },
    handleRead(data) {
      reqhel.changeStatus({ ids: data.muId }).then((r) => {
        if (r.succ) {
          this.getMsgList(0);
        }
      });
    },
    checkjsontext(task) {
      return task.t_jsontext && Object.keys(task.t_jsontext).length > 0;
    },
    searchHandler: debounce(function () {
      if (this.menuIndex === '1') {
        this.key = this.key.trim();
        this.getTaskList(0);
      } else if (this.menuIndex === '2') {
        this.msgKey = this.msgKey.trim();
        this.getMsgList(0);
      }
    }, 300),
    selectMenu(menuIndex) {
      this.menuIndex = menuIndex; // 切换是
    },
    // task handle
    handleCheckAllChange(val) {
      if (this.menuIndex === '1') {
        this.checkedItems = val ? this.todoType : [];
        this.isIndeterminate = false;
        this.getTaskList(0);
      } else if (this.menuIndex === '2') {
        this.checkedItems = val ? this.msgType : [];
        this.isIndeterminate = false;
        this.getMsgList(0);
      }
    },
    handleCheckedItemsChange(value) {
      if (this.menuIndex === '1') {
        const checkedCount = value.length;
        this.checkAll = checkedCount === this.todoType.length;
        this.isIndeterminate = checkedCount > 0 && checkedCount < this.todoType.length;
        this.getTaskList(0);
      } else if (this.menuIndex === '2') {
        const checkedCount = value.length;
        this.checkAll = checkedCount === this.msgType.length;
        this.isIndeterminate = checkedCount > 0 && checkedCount < this.msgType.length;
        this.getMsgList(0);
      }
    },
    handleTabClick() {
      if (this.menuIndex === '1') {
        this.key = '';
        this.getTaskList(0);
      } else if (this.menuIndex === '2') {
        this.msgKey = '';
        this.getMsgList(0);
      }
    },
    // handleTodo (oper, instanceid, taskid) {
    //   reqhel.doTodo({
    //     remark: this.form.remark,
    //     oper: oper,
    //     instance_id: instanceid,
    //     task_id: taskid,
    //   }).then(r => console.log(r))
    // },
    getTaskList(start) {
      this.start = start;

      reqhel.getTaskList(this.params).then((r) => {
        const data = this.formatData(r);
        if (this.activeName === 'todo') {
          this.todoCnt = data.count;
        }
        this.taskList = data.list;
      });
    },
    getMsgList(start) {
      this.start = start;
      reqhel.getMsgList(this.params).then((r) => {
        if (this.msgActiveName === 'unread') {
          this.unreadCnt = r.count;
          this.countForShowAllMsg = r.count;
        }

        if (r.count <= 10) {
          this.haveMore = false;
        } else {
          this.haveMore = true;
        }
        this.total = r.count;
        this.msgList = r.list;
      });
    },
    handleClose() {
      this.$emit('closeDialog');
    },
    handleOpen() {
      // 获取待办
      if (this.menuIndex === '1') {
        this.getTaskList(0);
      } else if (this.menuIndex === '2') {
        this.getMsgList(0);
      }

      this.initTodoType();
      // initMsgType ()
    },
    initTodoType() {
      this.key = '';
      this.checkAll = false;
      this.isIndeterminate = false;
      this.checkedItems = [];
      this.todoType = ['巡检', '维保', '演练', '变更', '事件', '维修'];
      this.msgType = ['通知', '消息'];
      // 获取待办类别
      // reqhel.getTodoType().then(r => {
      //   this.todoType = r
      // })
    },
    initMsgType() {
      this.checkAll = false;
      this.isIndeterminate = false;
      this.checkedItems = [];
      this.msgType = ['告警', '故障'];
    },
    scrollLoad() {
      this.start = this.start + this.limit;
      if (this.menuIndex === '1') {
        reqhel.getTaskList(this.params).then((r) => {
          // eslint-disable-next-line no-param-reassign
          r = this.formatData(r);
          if (this.activeName === 'todo') {
            this.todoCnt = r.count;
          }
          this.taskList = this.taskList.concat(r.list);
        });
      } else if (this.menuIndex === '2') {
        this.busy = true;
        reqhel.getMsgList(this.params).then((r) => {
          if (this.msgActiveName === 'unread') {
            this.unreadCnt = r.count;
          }
          if (r.list.length === 0) {
            this.haveMore = false;
          } else {
            this.haveMore = true;
          }
          this.total = r.count;
          this.msgList = this.msgList.concat(r.list);
          setTimeout(() => {
            this.busy = false;
          }, 1000);
        });
      }
    },
    formatData(taskData) {
      // eslint-disable-next-line no-param-reassign
      taskData.count = parseInt(taskData.count, 10);
      if (taskData.count > 0) {
        taskData.list.map((task) => {
          // eslint-disable-next-line no-param-reassign
          task.t_jsontext = JSON.parse(task.t_jsontext);
          return task;
        });
      }
      return taskData;
    },
  },
};
</script>
<style scoped>
  .have-read {
      /* margin-left: 700px; */
      float: right;
      margin-right: 40px;
      margin-top: 6px;
      /* margin-bottom: 17px; */
      font-size: 14px;
      color: #1470CC;
      white-space:nowrap;
      cursor: pointer;
  }
</style>

<style>
  #msgDiv .el-collapse-item__header{
      box-sizing: content-box !important;
  }
  #taskDiv .el-collapse-item__header{
      box-sizing: content-box !important;
  }
</style>
