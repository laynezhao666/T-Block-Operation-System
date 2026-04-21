<template>
  <el-form
    ref="form"
    :model="editting"
    :rules="rules"
    label-width="120px"
  >
    <el-form-item
      label="时间组名称"
      prop="group_name"
      required
    >
      <el-input
        v-model="editting.group_name"
        placeholder="请输入时间组名称"
      />
    </el-form-item>

    <el-form-item
      label="生效日期"
      prop="week"
      required
    >
      <el-checkbox-group v-model="editting.week">
        <el-checkbox
          v-for="(opt, i) in dayOptions"
          :key="i"
          :label="opt.value"
        >
          {{ opt.label }}
        </el-checkbox>
      </el-checkbox-group>
    </el-form-item>

    <el-form-item
      label="时段范围"
      prop="timezone"
      required
    >
      <simple-form-list-input
        :array="editting.timezone"
        :new-item="createNewTimeRange"
        :max="3"
        :min="1"
        add-label="添加时段"
      >
        <template
          #default="{ item, setItem }"
        >
          <el-time-picker
            v-if="checkIsTimeRangeNormalized"
            :value="item"
            is-range
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            placeholder="选择时间范围"
            value-format="HH:mm"
            format="HH:mm"
            @input="setItem"
          />
        </template>
      </simple-form-list-input>
    </el-form-item>
  </el-form>
</template>

<script>
import dayjs from 'dayjs';
import SimpleFormListInput from '../../../component/tedge-components/simple-form-list-input.vue';

const ruleCantEmpty = name => ({
  required: true, message: `${name}不能为空`,
});

const getDayStartTime = (time = '00:00') => dayjs(`2023-01-01 ${time}:00`).toDate();
const getDayEndTime = (time = '23:59') => dayjs(`2023-01-01 ${time}:59`).toDate();

export default {
  components: {
    SimpleFormListInput,
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
      dayOptions: [{
        value: 1,
        label: '周一',
      }, {
        value: 2,
        label: '周二',
      }, {
        value: 3,
        label: '周三',
      }, {
        value: 4,
        label: '周四',
      }, {
        value: 5,
        label: '周五',
      }, {
        value: 6,
        label: '周六',
      }, {
        value: 7,
        label: '周日',
      }],

      rules: {
        group_name: [ruleCantEmpty('时间组名称')],
        week: [ruleCantEmpty('生效日期')],
        timezone: [
          ruleCantEmpty('时段范围'),
          { validator: this.validateRange.bind(this) },
        ],
      },
    };
  },
  watch: {
    editting: {
      immediate: true,
      handler() {
        if (!this.editting) return;

        if (!this.editting.week) {
          this.$set(this.editting, 'week', []);
        }

        if (!this.editting.timezone) {
          this.$set(this.editting, 'timezone', [
            [getDayStartTime(), getDayEndTime()],
          ]);
        } else {
          this.editting.timezone = this.editting.timezone.map(item => this.normalizeTimeRange(item));
        }
      },
    },
  },
  methods: {
    async validate() {
      return this.$refs.form.validate();
    },
    validateRange(rule, value, callback) {
      const errors = [];

      value.forEach((tuple, i) => {
        if (tuple.length !== 2 || !tuple[0] || !tuple[1]) {
          errors.push(`第${i + 1}个时段为空`);
        }
      });

      const dayjsBetween = (target, range) => target.isAfter(range[0]) && target.isBefore(range[1]);

      const timeToDayjs = (strOrDate) => {
        const result = strOrDate instanceof Date
          ? dayjs(strOrDate)
          : dayjs(getDayStartTime(strOrDate));

        return result;
      };

      for (let i = value.length - 1; i > 0; i--) {
        const tuple1 = value[i];
        const t1Start = timeToDayjs(tuple1[0]);
        const t1End = timeToDayjs(tuple1[1]);
        for (let j = i - 1; j >= 0; j--) {
          const tuple2 = value[j];
          const t2Start = timeToDayjs(tuple2[0]);
          const t2End = timeToDayjs(tuple2[1]);

          const isIntersect = (dayjsBetween(t1Start, [t2Start, t2End]) || dayjsBetween(t1End, [t2Start, t2End]))
            || (dayjsBetween(t2Start, [t1Start, t1End]) || dayjsBetween(t2End, [t1Start, t1End]));

          if (isIntersect) {
            errors.push(`第${j + 1}个与第${i + 1}个时段有交叉`);
          }
        }
      }

      callback(errors.length ? errors.join('; ') : undefined);
    },
    createNewTimeRange() {
      return [
        getDayStartTime(),
        getDayEndTime(),
      ];
    },
    checkIsTimeRangeNormalized() {
      return this.editting.timezone instanceof Array;
    },
    normalizeTimeRange(range) {
      if ((range instanceof Array) && !range?.length) return range;

      if (typeof range === 'object') {
        return [
          getDayStartTime(range.begin),
          getDayStartTime(range.end),
        ];
      }

      if (typeof range === 'string') {
        return range.split('-').map(getDayStartTime);
      }

      if (range.every(item => item instanceof Date)) return range;

      return [
        getDayStartTime(range.begin),
        getDayStartTime(range.end),
      ];
    },
  },
};
</script>

<style>

</style>
