<template>
  <span>
    <admin-limit-tooltips>
      <span slot-scope="{ hasRight }">
        <el-button
          :disabled="!hasRight || !hasTemporaryCards"
          type="text"
          icon="tn-icon-edit"
          @click="showDialog"
        >
          更新有效期
        </el-button>
      </span>
    </admin-limit-tooltips>

    <el-dialog
      :visible.sync="dialogVisible"
      width="500px"
      :append-to-body="true"
      @open="handleOpen"
      @close="handleClose"
    >
      <template slot="title">
        <span style="font-weight: bold;">更新有效期</span>
      </template>
      <el-form
        ref="form"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <div style="margin-bottom: 16px; margin-left: 8px; font-size: 14px; font-weight: bold;">
          本次更新选中的临时卡有
          <span style="color: #f56c6c; font-weight: bold;">
            {{ temporaryCardsCount }}
          </span>
          张
        </div>
        <el-form-item
          label="新的有效期"
          prop="validTime"
        >
          <el-date-picker
            ref="datePicker"
            v-model="form.validTime"
            type="datetime"
            placeholder="选择时间"
            style="width: 100%;"
          />
        </el-form-item>
      </el-form>
      <template slot="footer">
        <el-button
          type="primary"
          :loading="loading"
          @click="handleConfirm"
        >
          确定
        </el-button>
        <el-button @click="handleClose">
          取消
        </el-button>
      </template>
    </el-dialog>
  </span>
</template>

<script>
import AdminLimitTooltips from 'feature/component/tedge-components/admin-limit-tooltips.vue';
import { axiosPut } from 'utils/axios-methods';

export default {
  components: {
    AdminLimitTooltips,
  },
  props: {
    tableContext: {
      type: Object,
      required: true,
    },
    selection: {
      type: Object,
      required: true,
    },
    getSelectedRows: {
      type: Function,
      required: true,
    },
  },
  data() {
    return {
      dialogVisible: false,
      loading: false,
      temporaryCardsCount: 0,
      temporaryCards: [],
      form: {
        validTime: null,
      },
      rules: {
        validTime: [
          { required: true, message: '请选择有效期', trigger: 'change' },
        ],
      },
    };
  },
  computed: {
    // 检查选中的卡片中是否有临时卡
    hasTemporaryCards() {
      const rows = this.getSelectedRows();
      return rows.length > 0 && rows.some(row => row.card_type === 1);
    },
  },
  methods: {
    showDialog() {
      const { getSelectedRows } = this;
      const rows = getSelectedRows();
      
      // 过滤出临时卡
      this.temporaryCards = rows.filter(row => row.card_type === 1);
      
      if (this.temporaryCards.length === 0) {
        this.$message.warning('请选择至少一张临时卡');
        return;
      }

      // 重置表单
      this.temporaryCardsCount = this.temporaryCards.length;
      this.form.validTime = null;
      this.dialogVisible = true;
    },
    handleOpen() {
      // 使用 $nextTick 确保 DOM 完全渲染后再聚焦
      this.$nextTick(() => {
        if (this.$refs.form) {
          this.$refs.form.clearValidate();
        }
      });
    },
    handleClose() {
      this.dialogVisible = false;
      if (this.$refs.form) {
        this.$refs.form.resetFields();
      }
    },
    handleConfirm() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) {
          return;
        }

        this.loading = true;
        try {
          const cardIds = this.temporaryCards.map(card => card.card_no);
          await this.batchUpdateValidTime(cardIds, this.form.validTime);
          this.$message.success(`成功更新 ${this.temporaryCards.length} 张卡片的有效期`);
          this.dialogVisible = false;
          // 刷新表格
          this.tableContext.loadData();
          // 取消选中
          this.selection.cancel(this.tableContext);
        } catch (error) {
          this.$message.error('批量更新失败：' + error.message);
        } finally {
          this.loading = false;
        }
      });
    },
    async batchUpdateValidTime(cardIds, validTime) {
      // validTime 是 Date 对象，需要转换为秒级时间戳
      const validTimeInSeconds = Math.floor(validTime.getTime() / 1000);
      
      const url = '/api/dcos/tdac-cgi/card/valid_time';
      await axiosPut(url, {
        cards: cardIds,
        valid_time: validTimeInSeconds,
      });
    },
  },
};
</script>
