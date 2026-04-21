<template>
  <el-form
    ref="form"
    label-position="top"
    :model="editting"
    :rules="rules"
  >
    <el-form-item
      prop="cards"
      label-width="100%"
      class="header-like-form-item"
    >
      <split-header-bar
        slot="label"
        title="授权人员/门禁卡"
        no-padding
      >
        <!-- 暂时隐藏 -->
        <el-radio-group
          v-if="false"
          v-model="activeMode"
          size="small"
        >
          <el-radio-button
            v-for="(mode, i) in modeOptions"
            :key="i"
            :label="mode.value"
          >
            {{ mode.label }}
          </el-radio-button>
        </el-radio-group>
      </split-header-bar>

      <el-transfer
        v-model="editting.cards"
        :data="userCards"
        :filter-method="filterMethod"
        :props="transferProps"
        :titles="transferTitles"
        filterable
        filter-placeholder="请输入关键字"
        class="transfer-select"
      >
        <template
          #default="{ option }"
        >
          <span>

            <span v-show="activeMode === 'users'">
              {{ option.staff && option.staff.company || '无' }} |
            </span>
            <span>
              {{ (option.staff && option.staff.name) || '未分配' }}
            </span>
            <span v-show="activeMode === 'users'">
              | {{ option.card_no }}
            </span>
          </span>
        </template>
      </el-transfer>
    </el-form-item>
  </el-form>
</template>

<script>
import SplitHeaderBar from '../../../component/tedge-components/split-header-bar.vue';

export default {
  components: {
    SplitHeaderBar,
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
    window.plf = this;
    return {
      activeMode: 'users',
      modeOptions: [{
        value: 'users',
        label: '人员视角',
      }, {
        value: 'cards',
        label: '门卡视角',
      }],

      transferProps: {
        key: 'card_no',
        label: 'card_no',
      },
      transferTitles: [
        '待选列表',
        '已选列表',
      ],

      userCards: [],

      rules: {
        // cards: [
        //   {
        //     validator(rule, value, cb) {
        //       cb(value?.length ? undefined : '授权人员不能为空');
        //     },
        //   },
        // ],
      },
    };
  },
  created() {
    this.loadUserCards();
  },
  methods: {
    async validate() {
      return this.$refs.form.validate();
    },
    async loadUserCards() {
      const { list: cards } = await this.$axios.post('/api/dcos/tdac-cgi/cards', {
        offset: 0,
        limit: 100000,
      });

      this.userCards = cards;
    },
    filterMethod(query, item) {
      const keywords = query?.trim();

      if (!keywords) return true;

      return ['staff.name', 'staff.company', 'card_no']
        .some(key => _.get(item, key)?.includes(keywords));
    },
  },
};
</script>

<style lang="scss" scoped>
.header-like-form-item {
  /deep/ {
    .el-form-item__label {
      padding-right: 0;
      width: 100%;

      &:before {
        display: none !important;
      }
    }

    .el-form-item__error {
      position: absolute;
      top: -24px;
      left: 50%;
      transform: translateX(-50%);
      height: 1em;
    }
  }
}

.transfer-select {
  margin-top: 16px;

  /deep/ .el-transfer-panel {
    width: calc(50% - 48px);
  }
}
</style>
