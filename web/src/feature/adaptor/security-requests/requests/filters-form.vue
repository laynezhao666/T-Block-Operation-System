<template>
  <el-form
    class="filters-form"
    size="small"
    label-width="8em"
  >
    <el-row :gutter="16">
      <el-col :span="8">
        <el-form-item
          label="消息类型"
        >
          <el-select
            v-model="filters.method"
            placeholder="请选择消息类型"
            clearable
            :fix-width="true"
          >
            <el-option
              :value="'add_card'"
              label="新增卡信息"
            />
            <el-option
              :value="'set_door_parameter'"
              label="设置门参数"
            />
            <el-option
              :value="'set_time_group'"
              label="设置时间组"
            />
            <el-option
              :value="'update_card_stuff'"
              label="更新卡成员信息"
            />
            <el-option
              :value="'delete_card'"
              label="删除卡信息"
            />
          </el-select>
        </el-form-item>
      </el-col>

      <el-col :span="8">
        <el-form-item
          label="消息状态"
        >
          <el-select
            v-model="filters.state"
            placeholder="请选择消息状态"
            clearable
            :fix-width="true"
          >
            <el-option
              :value="'待执行'"
              label="待执行"
            />
            <el-option
              :value="'成功'"
              label="成功"
            />
            <el-option
              :value="'失败'"
              label="失败"
            />
            <el-option
              :value="'过期'"
              label="过期"
            />
          </el-select>
        </el-form-item>
      </el-col>

      <el-col :span="8">
        <el-form-item
          label="创建时间"
        >
          <el-date-picker
            v-model="filters.create_time"
            type="datetimerange"
            :picker-options="pickerOptions"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            text-align="right">
          </el-date-picker>
        </el-form-item>
      </el-col>
    </el-row>
  </el-form>
</template>

<script>
export default {
  props: {
    tableContext: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {
      allGroups: [],

      typeOptions: [{
        label: '简单组态页',
        value: '简单组态页',
      }, {
        label: '树形组态页',
        value: '树形组态页',
      }, {
        label: '模板页',
        value: '模板页',
      }],
      systemOptions: [{
        label: '电脑端',
        value: '电脑端',
      }, {
        label: '平板端',
        value: '平板端',
      }],

      pickerOptions: {
        shortcuts: [{
          text: '最近一小时',
          onClick(picker) {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000);
            picker.$emit('pick', [start, end]);
          }
        }, {
          text: '最近一天',
          onClick(picker) {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000 * 24);
            picker.$emit('pick', [start, end]);
          }
        }, {
          text: '最近一周',
          onClick(picker) {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
            picker.$emit('pick', [start, end]);
          }
        }]
      },
    };

  },
  computed: {
    filters() {
      return this.tableContext.filters;
    },
  },
};
</script>

<style lang="scss" scoped>
.filters-form {
  border-top: 1px solid #efefef;
}
</style>
