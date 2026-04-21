<template>
  <div>
    <el-form
      v-if="editting"
      ref="form"
      :model="editting"
      :rules="rules"
      label-width="120px"
    >
      <el-form-item
        label="门卡号"
        prop="card_no"
      >
        <div
          v-if="isCreate"
          class="card-id-type-container"
        >
          <el-input
            v-model="editting.card_no"
            placeholder="请输入卡号"
            class="card-id-input"
          />

          <el-select
            v-model="editting.card_type"
            placeholder="卡类型"
            class="card-type-select"
            @change="handleCardTypeChange"
          >
            <el-option
              :value="0"
              label="长期卡"
            />
            <el-option
              :value="1"
              label="临时卡"
            />
          </el-select>
        </div>

        <div v-else>
          {{ editting.card_no }}
        </div>
      </el-form-item>

      <!-- 重新授权时也允许修改卡类型 -->
      <el-form-item
        v-if="!isCreate"
        label="卡类型"
        prop="card_type"
      >
        <el-select
          v-model="editting.card_type"
          placeholder="卡类型"
          @change="handleCardTypeChange"
        >
          <el-option
            :value="0"
            label="长期卡"
          />
          <el-option
            :value="1"
            label="临时卡"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        label="卡状态"
        prop="card_flag"
      >
        <el-radio-group
          v-if="isCreate"
          v-model="editting.card_flag"
        >
          <el-radio
            v-for="(opt, i) in validOptions"
            :key="i"
            :label="opt.value"
          >
            {{ opt.label }}
          </el-radio>
        </el-radio-group>

        <span v-else>
          {{ editting.card_flag === 0 ? '正常' : '停用' }}
        </span>
      </el-form-item>

      <el-form-item
        v-if="editting.card_type === 1"
        label="有效期"
        prop="valid_time"
      >
        <el-date-picker
          v-model="editting.valid_time"
          type="datetime"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          placeholder="选择时间"
        />
      </el-form-item>

      <el-form-item
        label="领卡人"
        prop="staff"
      >
        <div class="user-info-input-container">
          <el-input
            :value="editting.staff.company"
            disabled
            placeholder="单位"
            class="company-input"
          />
          <el-select
            v-model="editting.staff.id"
            placeholder="领卡人姓名"
            class="user-input"
            filterable
            clearable
            @change="handleStaffChange"
          >
            <el-option
              v-for="staff in staffs"
              :key="staff.id"
              :value="staff.id"
              :label="staff.name"
            />
          </el-select>
        </div>
      </el-form-item>
    </el-form>

    <permission-groups-table
      v-if="editting"
      v-model="editting.access_groups"
    />
  </div>
</template>

<script>
import _ from 'lodash';
import PermissionGroupsTable from './permission-groups-table.vue';

const ruleCantEmpty = name => ({
  required: true, message: `${name}不能为空`,
});

export default {
  components: {
    PermissionGroupsTable,
  },
  props: {
    editting: {
      type: Object,
      required: true,
    },
    isCreate: {
      type: Boolean,
      required: true,
    },
  },
  data() {
    return {
      item: [],
      validOptions: [{
        value: 0,
        label: '正常',
      }, {
        value: 1,
        label: '停用',
      }],
      staffs: [],

      rules: {
        card_no: [
          ruleCantEmpty('门卡号'),
          {
            validator: (rule, value, callback) => {
              callback(_.isNil(this.editting.card_type) ? '卡类型不能为空' : undefined);
            },
          },
        ],
        valid_time: [
          ruleCantEmpty('有效期'),
        ],
      },
    };
  },
  created() {
    this.loadStaffOptions();
  },
  watch: {
    editting: {
      handler(val) {
        if (val) {
          this.loadStaffOptions();
        }
      },
    },
  },
  methods: {
    async validate() {
      return this.$refs.form.validate();
    },
    async loadStaffOptions() {
      const url = '/api/dcos/tdac-cgi/staffs';
      const resp = await this.$axios.get(url, {
        offset: 0,
        limit: 100000,
      });
      this.staffs = resp.list;
    },
    handleStaffChange(staffId) {
      const {
        staffs,
        editting,
      } = this;

      const selectedStaff = _.find(staffs, { id: staffId });

      editting.staff.company = selectedStaff.company;
    },
    handlePermissionGroupsChange(groups) {
      this.$set(this.editting, 'permissionGroups', groups);
      this.editting.permissionGroups = groups;
    },
    // 处理卡类型变化
    handleCardTypeChange(newType) {
      // 从永久卡(0)切换到临时卡(1)时，如果没有有效期，设置默认值
      if (newType === 1 && !this.editting.valid_time) {
        // 默认设置为30天后
        const defaultValidTime = new Date();
        defaultValidTime.setDate(defaultValidTime.getDate() + 30);
        this.$set(this.editting, 'valid_time', defaultValidTime);
      }
      // 从临时卡(1)切换到永久卡(0)时，清除有效期（提交时会处理）
      // 这里不清除是为了让用户可以看到之前的值，如果需要再次切换回临时卡
    },
  },
};
</script>

<style lang="scss" scoped>
.card-id-type-container {
  display: flex;
}

.card-id-input {
  margin: 12px 8px 12px 0;
}

.card-type-select {
  width: 120px;
}

.user-info-input-container {
  display: flex;
}

.company-input {
  width: 120px;
  margin-right: 8px;
}

.company-input, .user-input {
  margin-top: 12px;
  margin-bottom: 12px;
}
</style>
