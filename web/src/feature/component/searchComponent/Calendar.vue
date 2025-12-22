<template>
  <div class="change-calendar">
    <div class="change-calendar-toolbar">
      <div
        class="change-calendar-middle"
      >
        <el-button
          type="icon"
          icon="tn-icon-arrow-left"
          @click="goToMonth('prev')"
        />
        <span class="change-calendar-date">
          {{ currentDate.format('YYYY') }} 年 {{ currentDate.format('MM') }} 月
        </span>
        <el-button
          type="icon"
          icon="tn-icon-arrow-right"
          @click="goToMonth('next')"
        />
      </div>

      <div
        class="legend-list"
      >
        <slot name="legend" />
      </div>
    </div>
    <div
      class="change-calendar-calendar"
    >
      <FullCalendar
        ref="fullCalendar"
        :options="calendarOptions"
      />
    </div>
  </div>
</template>
<script>
import moment from 'moment';
import FullCalendar from '@fullcalendar/vue';
import dayGridPlugin from '@fullcalendar/daygrid';
import interactionPlugin from '@fullcalendar/interaction';
import tippy from 'tippy.js';
import 'tippy.js/dist/tippy.css';
import { throttle, cloneDeep } from 'lodash';
export default {
  inject: ['configCgi', 'tableConfig'],
  components: {
    FullCalendar,
  },
  props: {
    query: {
      type: Object,
      default() {
        return {};
      },
    },
    /**
     * @params {Object} config.options 直接传给fullCalendar
     * @params {String} config.currentDateKeys: 使用的日期key起始名称
     * @params {Function} bindEvents(list) 提供列表，返回fullcalendar事件对象
     */
    config: {
      type: Object,
      default() {
        return {};
      },
    },
  },
  data() {
    const DAY_MAP = {
      Sun: '星期日',
      Mon: '星期一',
      Tue: '星期二',
      Wed: '星期三',
      Thu: '星期四',
      Fri: '星期五',
      Sat: '星期六',
    };

    return {
      rst: [],
      calendarOptions: {
        plugins: [dayGridPlugin, interactionPlugin],
        initialView: 'dayGridMonth',

        eventDidMount(info) {
          tippy(info.el, {
            content: info.event.extendedProps.description,
          });
        },

        headerToolbar: false,
        contentHeight: '75vh',
        firstDay: 1,
        dayHeaderContent: ({ text }) => DAY_MAP[text],

        events: [],
        ...this.config.options,
      },
      currentDate: moment(),
      calendarApi: null,
      timeFormat: 'YYYY-MM-DD HH:mm:ss',
    };
  },
  computed: {
    list() {
      return this.rst;
    },
    currentDateKeys() {
      try {
        return this.config.currentDateKeys;
      } catch (e) {
        console.log('未配置日历时间对应的高级搜索字段key');
      }
      return '';
    },
    currentDateFilter() {
      const d = moment(this.currentDate);
      const GTE = moment(d.startOf('month'));
      const LTE = moment(d.endOf('month'));
      return {
        GTE,
        LTE,
      };
    },
    /**
     * 单独加一条日历月份的筛选，高级筛选按默认
     * conditions: {
     *   default: [], // 略
     *   time: [
     *     {
     *         "field":"plan_begin_time",
     *         "operator":"LIKE",
     *         "relation":"OR",
     *         "value":"2022-06"
     *     },
     *     {
     *         "field":"plan_end_time",
     *         "operator":"LIKE",
     *         "relation":"OR",
     *         "value":"2022-06"
     *     }
     *   ]
     * }
     */
    newQuery() {
      const query = cloneDeep(this.query);
      query.conditions.time = this.currentDateKeys.map(field => ({
        field,
        operator: 'LIKE',
        relation: 'OR',
        value: this.currentDate.format('YYYY-MM'),
      }));

      return query;
    },
    /**
     * 根据筛选和当前的月份，合并计算拉取数据的起始时间
     * conditions: {
     *   default: [{
     *     field: "plan_end_time"
     *     operator: "GTE"
     *     relation: "AND"
     *     value: "2021-07-01 00:00:00"
     *   },
     *   {
     *     field: "plan_end_time"
     *     operator: "LTE"
     *     relation: "AND"
     *     value: "2021-07-08 23:59:59"
     *   }]
     * }
     */
    // newQuery() {
    //   const opArr = ['GTE', 'LTE'];
    //   const query = cloneDeep(this.query);
    //   const defaultConditions = (query.conditions && query.conditions.default) || [];
    //   try {
    //     // 日历的起始
    //     const a = opArr.map(op => this.currentDateFilter[op]);
    //     // 筛选的起始
    //     const b = [];
    //     query.conditions = {
    //       ...(query.conditions || {}),
    //       default: defaultConditions.map((item) => {
    //         if (item.field !== this.currentDateKey) {
    //           return item;
    //         }
    //         b[opArr.indexOf(item.operator)] = moment(item.value);
    //         return null;
    //       }).filter(Boolean),
    //     };

    //     const result = opArr.map((op, i) => b[i] ? moment[['max', 'min'][i]](a[i], b[i]) : a[i]);
    //     const diff = result[0].diff(result[1]);

    //     // console.log([result, a, b].map(x => x.map(n => n.format(this.timeFormat))));

    //     if (diff >= 0) {
    //       // 说明没有符合日期要求的数据
    //       return false;
    //     }
    //     query.conditions.default = [
    //       ...query.conditions.default,
    //       ...result.map((d, i) => ({
    //         field: this.currentDateKey,
    //         operator: opArr[i],
    //         relation: 'AND',
    //         value: moment(d).format(this.timeFormat),
    //       })),
    //     ];
    //     return query;
    //   } catch (e) { console.log(e); }
    //   return query;
    // },
  },
  watch: {
    query: {
      handler() {
      // 筛选的时候，切换当前月份到对应日期key的第一个值
        try {
          if (!this.query.conditions || !this.query.conditions.default) return;
          this.query.conditions.default.forEach((item) => {
            if (item.field === this.currentDateKeys[0]) {
              this.currentDate = moment(item.value);
              return false;
            }
          });
        } catch (e) {
          console.log(e);
        }
      },
      deep: true,
    },
    newQuery: {
      handler() {
        this.refresh();
      },
      // immediate: true,
      deep: true,
    },
    list: {
      handler() {
        const events = this.config.buildEvents(this.list);
        this.$set(this.calendarOptions, 'events', events);
        this.calendarApi.refetchEvents();
      },
      deep: true,
    },
    currentDate() {
      this.calendarApi.gotoDate(this.currentDate.format());
    },
  },
  mounted() {
    this.calendarApi = this.$refs.fullCalendar && this.$refs.fullCalendar.getApi();
    this.calendarApi.updateSize();
  },
  methods: {
    /**
     * 初始化行事历界面
     */
    goToMonth(direction) {
      this.calendarApi[direction]();
      const d = this.calendarApi.getDate();
      this.currentDate = moment(d);
    },
    refresh() {
      this.filterHandler();
    },
    filterHandler() {
      // 说明没有符合日期要求的数据
      if (this.newQuery === false) {
        this.rst = [];
        return;
      }
      const params = {
        limit: -1,
        start: 0,
        ...this.newQuery,
      };
      if (params.conditions && params.conditions.default) {
        params.conditions.default = params.conditions.default.filter(v => v.field !== 'id');
      }
      this.getData(params);
    },
    getData: throttle(function (params) {
      this.$axios.post(this.configCgi.queryCgi, params).then((data) => {
        this.total = data.count;
        this.rst = data.list;
      })
        .catch(() => {
          this.total = 0;
          this.rst = [];
        });
    }, 1000, { leading: true, trailing: true }),
  },

};
</script>

<style lang="scss">
.change-calendar {
  &-toolbar {
    position: relative;
    display: flex;
    align-items: center;
    height: 56px;
    margin: 0 24px;
  }

  .legend-list {
    position: absolute;
    width: 350px;
    height: 56px;
    line-height: 56px;
    right: 0;
  }

  &-middle {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  &-date {
    font-size: 16px;
    margin: 0 8px;
  }

  .legend {
    &-list {
      margin-left: auto;
      display: flex;
      align-items: center;
      justify-content: flex-end;
    }

    display: flex;
    margin-left: 20px;
    align-items: center;

    &-icon {
      width: 12px;
      height: 12px;
      border-radius: 50%;

      margin-right: 8px;
    }
  }

  &-calendar {
    padding: 0px 24px 16px;
    // border-top: 1px solid #f0f0f0;
  }
}

// fullCalendar 样式覆盖
.fc {
  --fc-border-color: #f0f0f0;
  --fc-today-bg-color: rgba(208, 227, 245, .3);

  th {
    line-height: 48px;
    background: #fbfbfb;
  }

  &-daygrid-event {
    line-height: 1.5;
  }

  .fc-daygrid-day-top {
    flex-direction: row;
  }
}

</style>
